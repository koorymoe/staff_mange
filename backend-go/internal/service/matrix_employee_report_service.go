package service

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ═══ تقرير الموظف اليومي — باللهجة العراقية، بأرقام حقيقية ═══
//
// قرار (ع): تقرير وافي عن كل موظف: دوامه مقابل جدوله، وتأخيره، وتسلسل
// كل حجز (تجهيز ← طلعة ← وصول ← بدء ← إنجاز) ومنو أخّر منو، وإجازاته
// ونمطها، و«شخصيته بالشغل» من أرقام ٣٠ يوم.
// ⚠️ بلا توقعات شخصية (مرض، سفر…) — قراره الصريح. أرقام وأنماط بس.
// ⚠️ للمدير والمالك بس. ولو هايكو يصيغ الخلاصة: بلا اسم.

const lateGraceMin = 10

type ReportLine struct {
	Text  string   `json:"text"`
	Tone  string   `json:"tone"`            // OK | WARN | BAD | INFO
	Link  string   `json:"link,omitempty"`  // «ودّيني» — مسار بالواجهة للمشكلة
	Fixes []string `json:"fixes,omitempty"` // حلول مقترحة (قواعد ثابتة، بلا اسم)
}

func rl(text, tone string) ReportLine { return ReportLine{Text: text, Tone: tone} }

// withFix يضيف رابط وحلول لسطر — نستعمله بس للسطور الي بيها مشكلة.
func withFix(l ReportLine, link string, fixes ...string) ReportLine {
	l.Link, l.Fixes = link, fixes
	return l
}

type ReportStage struct {
	Label string `json:"label"`
	At    string `json:"at"`
	Gap   int    `json:"gap"` // دقائق من المرحلة الي قبلها
}

type ReportJob struct {
	Code    string        `json:"code"`
	Service string        `json:"service"`
	Status  string        `json:"status"`
	Stages  []ReportStage `json:"stages"`
	Lines   []ReportLine  `json:"lines"`
}

type EmployeeReport struct {
	EmployeeID  string         `json:"employeeId"`
	Name        string         `json:"name"`
	Group       string         `json:"group"`
	Day         string         `json:"day"`
	Attendance  []ReportLine   `json:"attendance"`
	Jobs        []ReportJob    `json:"jobs"`
	Leaves      []ReportLine   `json:"leaves"`
	Behavior    []ReportLine   `json:"behavior"`
	Reminders   []ReportLine   `json:"reminders"`
	Performance []ReportLine   `json:"performance"`
	SlowJobs    []SlowJob      `json:"slowJobs"`
	Summary     string         `json:"summary"`
	SummaryBy   string         `json:"summaryBy"` // RULES | MODEL
	Workload    []WorkloadItem `json:"workload"`
}

type MatrixEmployeeReportService struct {
	repo   *repository.MatrixReportRepository
	watch  *MatrixAutopilotService
	client *anthropic.Client
	model  string
}

func NewMatrixEmployeeReportService(repo *repository.MatrixReportRepository, watch *MatrixAutopilotService) *MatrixEmployeeReportService {
	return &MatrixEmployeeReportService{repo: repo, watch: watch}
}

func (s *MatrixEmployeeReportService) EnableModel(apiKey, modelName string) {
	if apiKey == "" {
		return
	}
	if modelName == "" {
		modelName = "claude-haiku-4-5"
	}
	c := anthropic.NewClient(option.WithAPIKey(apiKey))
	s.client, s.model = &c, modelName
}

func hm(t time.Time) string { return t.In(debriefLoc).Format("15:04") }

// shiftWindow بداية ونهاية الجدول — من الموظف، وإلا افتراضي حسب الوجبة.
func shiftWindow(e *repository.ReportEmployee) (start, end string, assumed bool) {
	if e.ShiftStart.Valid && e.ShiftStart.String != "" && e.ShiftEnd.Valid && e.ShiftEnd.String != "" {
		return e.ShiftStart.String, e.ShiftEnd.String, false
	}
	if e.Shift.String == "EVENING" {
		return "16:00", "23:59", true
	}
	return "08:00", "16:00", true
}

func atClock(day time.Time, clock string) time.Time {
	t, err := time.ParseInLocation("15:04", clock, debriefLoc)
	if err != nil {
		return day
	}
	return time.Date(day.Year(), day.Month(), day.Day(), t.Hour(), t.Minute(), 0, 0, debriefLoc)
}

func mins(d time.Duration) int { return int(math.Round(d.Minutes())) }

func durText(m int) string {
	if m < 60 {
		return fmt.Sprintf("%d دقيقة", m)
	}
	h, r := m/60, m%60
	if r == 0 {
		return fmt.Sprintf("%d ساعة", h)
	}
	return fmt.Sprintf("%d ساعة و%d دقيقة", h, r)
}

func (s *MatrixEmployeeReportService) Report(employeeID, day string) (*EmployeeReport, error) {
	if _, err := time.Parse("2006-01-02", day); err != nil {
		day = time.Now().In(debriefLoc).Format("2006-01-02")
	}
	dayT, _ := time.ParseInLocation("2006-01-02", day, debriefLoc)
	e, err := s.repo.Employee(employeeID)
	if err != nil {
		return nil, err
	}
	subj, _ := s.watch.actions.Subject(employeeID)
	rep := &EmployeeReport{EmployeeID: e.ID, Name: e.Name, Day: day,
		Attendance: []ReportLine{}, Jobs: []ReportJob{}, Leaves: []ReportLine{}, Behavior: []ReportLine{}, Reminders: []ReportLine{}, Performance: []ReportLine{}, SlowJobs: []SlowJob{}}
	if subj != nil {
		rep.Group = WatchGroup(*subj)
		if day == time.Now().In(debriefLoc).Format("2006-01-02") {
			rep.Workload = s.watch.workload(*subj)
		}
	}

	// ── الدوام ──
	start, end, assumed := shiftWindow(e)
	note := ""
	if assumed {
		note = " (الجدول افتراضي — ماكو جدول مسجّل إله)"
	}
	att, _ := s.repo.Attendance(employeeID, day)
	shiftMin := mins(atClock(dayT, end).Sub(atClock(dayT, start)))
	if len(att) == 0 {
		rep.Attendance = append(rep.Attendance, rl(Fmt("ما سجّل حضور هاليوم. جدوله من %s لـ%s%s.", start, end, note), "BAD"))
		if assumed {
			rep.Attendance[len(rep.Attendance)-1].Link = "/work-schedule"
		}
	} else {
		first := att[0].CheckIn
		late := mins(first.Sub(atClock(dayT, start)))
		line := rl(Fmt("حضر الساعة %s، وجدوله يبدي %s%s.", hm(first), start, note), "OK")
		if late > lateGraceMin {
			line = rl(Fmt("حضر الساعة %s وجدوله يبدي %s — تأخّر %s%s.", hm(first), start, durText(late), note), "WARN")
			line = withFix(line, "/work-schedule", "ذكّره قبل دوامه بساعة.", "إذا وقت بدايته ما يناسب شغله، عدّله من جدول الدوام.")
		}
		rep.Attendance = append(rep.Attendance, line)
		worked := 0
		lastOut := ""
		for _, a := range att {
			if a.CheckOut.Valid {
				worked += mins(a.CheckOut.Time.Sub(a.CheckIn))
				lastOut = hm(a.CheckOut.Time)
			}
		}
		if lastOut == "" {
			rep.Attendance = append(rep.Attendance, rl("ما سجّل انصراف لحد هسه.", "WARN"))
		} else {
			tone := "OK"
			if worked < shiftMin-30 {
				tone = "WARN"
			}
			rep.Attendance = append(rep.Attendance, rl(Fmt("انصرف الساعة %s. اشتغل %s من أصل %s بالجدول.", lastOut, durText(worked), durText(shiftMin)), tone))
		}
	}

	// ── الحجوزات ومراحلها ──
	jobs, _ := s.repo.Jobs(employeeID, day)
	for _, j := range jobs {
		rj := ReportJob{Code: j.Code, Service: j.Service.String, Status: j.Status, Stages: []ReportStage{}, Lines: []ReportLine{}}
		type st struct {
			label string
			t     *time.Time
		}
		all := []st{}
		add := func(label string, ok bool, t time.Time) {
			if ok {
				tt := t
				all = append(all, st{label, &tt})
			}
		}
		add("الموعد", j.ScheduledAt.Valid, j.ScheduledAt.Time)
		add("التكليف", j.AssignedAt.Valid, j.AssignedAt.Time)
		add("تجهيز المواد", j.MaterialsReadyAt.Valid, j.MaterialsReadyAt.Time)
		add("الطلعة", j.DepartedAt.Valid, j.DepartedAt.Time)
		add("الوصول", j.ArrivedAt.Valid, j.ArrivedAt.Time)
		add("بدء العمل", j.StartedAt.Valid, j.StartedAt.Time)
		add("الإنجاز", j.CompletedAt.Valid, j.CompletedAt.Time)
		var prev *time.Time
		for _, x := range all {
			g := 0
			if prev != nil && x.label != "الموعد" {
				g = mins(x.t.Sub(*prev))
			}
			rj.Stages = append(rj.Stages, ReportStage{Label: x.label, At: hm(*x.t), Gap: g})
			if x.label != "الموعد" {
				prev = x.t
			}
		}
		leader := j.LeaderName.String
		if leader == "" {
			leader = "الليدر"
		}
		// منو أخّر: الوصول بعد الموعد، وسببه أطول مرحلة قبل الوصول.
		if j.ScheduledAt.Valid && j.ArrivedAt.Valid {
			late := mins(j.ArrivedAt.Time.Sub(j.ScheduledAt.Time))
			if late > 15 {
				reason := ""
				if j.MaterialsReadyAt.Valid && j.MaterialsReadyAt.Time.After(j.ScheduledAt.Time.Add(-30*time.Minute)) {
					by := j.MaterialsBy.String
					if by == "" {
						by = leader
					}
					reason = Fmt(" — المواد تجهّزت الساعة %s (جهّزها %s)، فالكادر طلع متأخر", hm(j.MaterialsReadyAt.Time), by)
				}
				rj.Lines = append(rj.Lines, rl(Fmt("الكادر وصل متأخر %s عن الموعد%s.", durText(late), reason), "WARN"))
			} else {
				rj.Lines = append(rj.Lines, rl("الكادر وصل بالوقت.", "OK"))
			}
		}
		if j.StartedAt.Valid && j.CompletedAt.Valid {
			work := mins(j.CompletedAt.Time.Sub(j.StartedAt.Time))
			if work > 14*60 || work < 0 {
				// رقم ما ينصدّق: الأرجح ضغطة «إنجاز» تأخرت أيام — ما نعرضه كوقت شغل.
				rj.Lines = append(rj.Lines, rl(Fmt("وقت الشغل ما ينحسب: بين ضغطة البدء (%s) وضغطة الإنجاز %s — الأرجح «إنجاز» انضغط متأخر.", hm(j.StartedAt.Time), durText(work)), "WARN"))
			} else {
				rj.Lines = append(rj.Lines, rl(Fmt("وقت الشغل الفعلي (من ضغطة البدء للإنجاز): %s.", durText(work)), "INFO"))
			}
		} else if j.StartedAt.Valid {
			rj.Lines = append(rj.Lines, rl(Fmt("بدا العمل الساعة %s وبعده ما انضغط «إنجاز».", hm(j.StartedAt.Time)), "WARN"))
		} else if j.Status == "CONFIRMED" || j.Status == "IN_PROGRESS" {
			rj.Lines = append(rj.Lines, rl("بعد ما بدا العمل بهالحجز.", "INFO"))
		}
		rep.Jobs = append(rep.Jobs, rj)
	}

	// ── الإجازات ونمطها (هالشهر + ٩٠ يوم للنمط) ──
	monthStart := time.Date(dayT.Year(), dayT.Month(), 1, 0, 0, 0, 0, debriefLoc)
	lv, _ := s.repo.Leaves(employeeID, dayT.AddDate(0, 0, -90))
	mCount, mDays, longest := 0, 0, 0
	weekdays := map[time.Weekday]int{}
	for _, l := range lv {
		d := int(l.EndDate.Sub(l.StartDate).Hours()/24) + 1
		if d > longest {
			longest = d
		}
		weekdays[l.StartDate.Weekday()]++
		if !l.StartDate.Before(monthStart) {
			mCount++
			mDays += d
		}
	}
	tone := "OK"
	if mDays > e.MonthlyLeaves {
		tone = "WARN"
	}
	rep.Leaves = append(rep.Leaves, rl(Fmt("هالشهر طلب %d إجازة بمجموع %d يوم، والمسموح %d يوم.", mCount, mDays, e.MonthlyLeaves), tone))
	if len(lv) >= 3 {
		var topDay time.Weekday
		top := 0
		for wd, n := range weekdays {
			if n > top {
				top, topDay = n, wd
			}
		}
		if top*2 > len(lv) {
			rep.Leaves = append(rep.Leaves, rl(Fmt("نمط: %d من %d إجازاته (آخر ٩٠ يوم) تبدي يوم %s.", top, len(lv), weekdayAr(topDay)), "WARN"))
		}
	}
	if longest >= 4 {
		rep.Leaves = append(rep.Leaves, rl(Fmt("أطول إجازة بآخر ٩٠ يوم: %d أيام.", longest), "INFO"))
	}

	// ── شخصيته بالشغل (٣٠ يوم، أرقام بس) ──
	rng, _ := s.repo.AttendanceRange(employeeID, 30)
	if len(rng) > 0 {
		onTime, lateSum, lateN, noOut := 0, 0, 0, 0
		lateDays := map[time.Weekday]int{}
		seen := map[string]bool{}
		for _, a := range rng {
			d := a.CheckIn.In(debriefLoc)
			key := d.Format("2006-01-02")
			if seen[key] {
				continue
			}
			seen[key] = true
			l := mins(a.CheckIn.Sub(atClock(d, start)))
			if l <= lateGraceMin {
				onTime++
			} else {
				lateSum += l
				lateN++
				lateDays[d.Weekday()]++
			}
			if !a.CheckOut.Valid {
				noOut++
			}
		}
		days := len(seen)
		pct := onTime * 100 / days
		t := "OK"
		if pct < 70 {
			t = "BAD"
		} else if pct < 90 {
			t = "WARN"
		}
		line := Fmt("الالتزام: حضر بوقته %d من %d يوم (%d٪).", onTime, days, pct)
		if lateN > 0 {
			line += Fmt(" متوسط تأخيره %s.", durText(lateSum/lateN))
		}
		rep.Behavior = append(rep.Behavior, rl(line, t))
		if noOut > 0 {
			rep.Behavior = append(rep.Behavior, rl(Fmt("ما سجّل انصراف %d يوم من %d.", noOut, days), "WARN"))
		}
		if lateN >= 3 {
			var wd time.Weekday
			top := 0
			for k, v := range lateDays {
				if v > top {
					top, wd = v, k
				}
			}
			if top >= 2 {
				rep.Behavior = append(rep.Behavior, rl(Fmt("روتين التأخير: أكثر يوم يتأخر بي %s (%d مرات).", weekdayAr(wd), top), "INFO"))
			}
		}
	}
	if e.IsLeader {
		if avg, n := s.repo.ResponseMinutes(employeeID); n >= 3 {
			t := "OK"
			if avg > 60 {
				t = "WARN"
			}
			rep.Behavior = append(rep.Behavior, rl(Fmt("سرعة استجابته: من التكليف لتجهيز المواد بالمعدل %s (%d مهمة).", durText(int(avg)), n), t))
		}
		if adj := s.repo.AdjustedInvoices(employeeID); adj > 0 {
			rep.Behavior = append(rep.Behavior, rl(Fmt("%d من فواتيره انعدّلت بالتدقيق هالشهر.", adj), "WARN"))
		}
	}
	if c := s.repo.Complaints(employeeID); c > 0 {
		rep.Behavior = append(rep.Behavior, rl(Fmt("%d شكوى مرتبطة بشغله بآخر ٣٠ يوم.", c), "BAD"))
	}
	if a, err := s.repo.Achievements(employeeID); err == nil {
		if a.Total == 0 {
			rep.Behavior = append(rep.Behavior, rl("ما كتب ولا إنجاز يومي بآخر ٣٠ يوم.", "WARN"))
		} else {
			line := Fmt("الإنجازات اليومية: كتب %d يوم من ٣٠، بمعدل %d حرف.", a.Days, a.AvgLen)
			t := "OK"
			if a.Duplicates > 0 {
				line += Fmt(" كرر نفس النص %d مرة.", a.Duplicates)
				t = "WARN"
			}
			if a.Good+a.NeedsReview > 0 {
				line += Fmt(" تقييمها: %d جيد، %d يحتاج مراجعة.", a.Good, a.NeedsReview)
			}
			rep.Behavior = append(rep.Behavior, rl(line, t))
		}
	}

	// ── تذكيرات ماتركس ──
	if r, err := s.repo.Reminders(employeeID); err == nil && r.Total > 0 {
		t := "OK"
		if r.Escalated > 0 {
			t = "WARN"
		}
		line := Fmt("ماتركس ذكّره %d مرة بآخر ٣٠ يوم: انحل %d، وصعد للمدير %d.", r.Total, r.Resolved, r.Escalated)
		if r.TopKind.Valid {
			line += Fmt(" أكثر شي يتكرر عليه: «%s» (%d).", model.AiActionLabels[r.TopKind.String], r.TopCount)
		}
		rep.Reminders = append(rep.Reminders, rl(line, t))
	}

	s.performance(rep, employeeID)

	rep.Summary, rep.SummaryBy = s.summary(rep), "RULES"
	if s.client != nil {
		if t, err := s.modelSummary(rep); err == nil && t != "" {
			rep.Summary, rep.SummaryBy = t, "MODEL"
		}
	}
	return rep, nil
}

func weekdayAr(d time.Weekday) string {
	return [...]string{"الأحد", "الاثنين", "الثلاثاء", "الأربعاء", "الخميس", "الجمعة", "السبت"}[d]
}

// Fmt اختصار fmt.Sprintf — الجمل كثيرة.
func Fmt(f string, a ...any) string { return fmt.Sprintf(f, a...) }

// summary خلاصة بالقواعد: أسوأ شي، وأحسن شي.
func (s *MatrixEmployeeReportService) summary(r *EmployeeReport) string {
	bad, warn, ok := []string{}, []string{}, 0
	for _, sec := range [][]ReportLine{r.Attendance, r.Performance, r.Leaves, r.Behavior, r.Reminders} {
		for _, l := range sec {
			switch l.Tone {
			case "BAD":
				bad = append(bad, l.Text)
			case "WARN":
				warn = append(warn, l.Text)
			case "OK":
				ok++
			}
		}
	}
	for _, j := range r.Jobs {
		for _, l := range j.Lines {
			if l.Tone == "WARN" {
				warn = append(warn, j.Code+": "+l.Text)
			}
		}
	}
	switch {
	case len(bad) > 0:
		return "يحتاج متابعة: " + bad[0] + func() string {
			if len(warn) > 0 {
				return " وبعد: " + warn[0]
			}
			return ""
		}()
	case len(warn) > 0:
		return "شغله ماشي بس بي ملاحظات: " + strings.Join(firstN(warn, 2), " ")
	case ok > 0:
		return "يومه نظيف — ملتزم وماكو ملاحظة."
	}
	return "ماكو بيانات كافية هاليوم."
}

func firstN(xs []string, n int) []string {
	sort.SliceStable(xs, func(i, j int) bool { return len(xs[i]) < len(xs[j]) })
	if len(xs) > n {
		return xs[:n]
	}
	return xs
}

const reportSystemPrompt = `أنت «ماتركس»، عقل المتابعة بشركة الأماني (العراق). تنطيك أسطر تقرير موظف (بلا اسم).
اكتب خلاصة للمدير بـ٢-٣ جمل باللهجة العراقية: أهم مشكلة بالأرقام، وأهم شي زين، واقتراح عملي واحد.
لا تخترع رقم، ولا تخمّن أسباب شخصية (مرض، سفر، ظروف)، ولا تقترح عقوبة. بلا مقدمات.`

func (s *MatrixEmployeeReportService) modelSummary(r *EmployeeReport) (string, error) {
	var b strings.Builder
	b.WriteString("المجموعة: " + r.Group + "\n")
	for _, sec := range [][]ReportLine{r.Attendance, r.Performance, r.Leaves, r.Behavior, r.Reminders} {
		for _, l := range sec {
			b.WriteString("- " + strings.ReplaceAll(l.Text, r.Name, "الموظف") + "\n")
		}
	}
	for _, j := range r.Jobs {
		for _, l := range j.Lines {
			// سطر التجهيز بي اسم الي جهّز — ما يطلع للنموذج.
			if strings.Contains(l.Text, "جهّزها") {
				continue
			}
			b.WriteString("- حجز: " + l.Text + "\n")
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	resp, err := s.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model: anthropic.Model(s.model), MaxTokens: 300,
		System:   []anthropic.TextBlockParam{{Text: reportSystemPrompt, CacheControl: anthropic.NewCacheControlEphemeralParam()}},
		Messages: []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock(b.String()))},
	})
	if err != nil {
		return "", err
	}
	var out strings.Builder
	for _, block := range resp.Content {
		if tb, ok := block.AsAny().(anthropic.TextBlock); ok {
			out.WriteString(tb.Text)
		}
	}
	return strings.TrimSpace(out.String()), nil
}

// ═══ الأداء التفصيلي — آخر ٣٠ يوم مقابل الـ٣٠ الي قبلها ═══
//
// مثال (ع): «طلع لـ٢٠ حجز، كل وحدة تحتاج ساعتين، وهو طوّل ٤، ونصها ما كملت».
// المتوقع = وسيط مدة نفس الخدمة عند كل الكادر (آخر ١٨٠ يوم)، والمدد الي
// فوگ ١٤ ساعة تنعزل (غالباً نسى يسكّر الحجز) وما تدخل المعدل.

type SlowJob struct {
	BookingID string `json:"bookingId"`
	Code      string `json:"code"`
	Service   string `json:"service"`
	Actual    int    `json:"actual"`
	Expected  int    `json:"expected"`
}

func (s *MatrixEmployeeReportService) performance(rep *EmployeeReport, id string) {
	cur, err := s.repo.Performance(id, 0)
	if err != nil || cur.Total == 0 {
		return
	}
	prev, _ := s.repo.Performance(id, 30)
	p := &rep.Performance
	line := Fmt("طلع لـ%d حجز: كمل %d", cur.Total, cur.Completed)
	if cur.Partial > 0 {
		line += Fmt("، جزئي %d", cur.Partial)
	}
	if open := cur.Total - cur.Completed - cur.Partial; open > 0 {
		line += Fmt("، بعدها مفتوحة %d", open)
	}
	tone := "OK"
	unfinished := cur.Total - cur.Completed
	if cur.Total >= 4 && unfinished*2 >= cur.Total {
		tone = "BAD"
	} else if unfinished*4 >= cur.Total && cur.Total >= 4 {
		tone = "WARN"
	}
	l := rl(line+".", tone)
	if tone != "OK" {
		l = withFix(l, "/bookings", "تأكد إن المواد والعدّة تتجهز قبل الطلعة — الجزئي غالباً من نقص مواد.",
			"راجع الحجوزات المفتوحة ويّاه وحدد سبب كل وحدة.", "قلّل حجوزاته باليوم لحد ما ترجع نسبة الإكمال.")
	}
	*p = append(*p, l)

	if cur.Timed > 0 && cur.AvgExpected > 0 {
		ratio := cur.AvgActual / cur.AvgExpected
		line := Fmt("المدة: المتوقع بالمعدل %s للحجز، وهو ياخذ %s (%d حجز موقوت)", durText(int(cur.AvgExpected)), durText(int(cur.AvgActual)), cur.Timed)
		tone := "OK"
		switch {
		case ratio >= 1.6:
			line += Fmt(" — يعني %.1f ضعف المتوقع.", ratio)
			tone = "BAD"
		case ratio >= 1.25:
			line += Fmt(" — أبطأ بـ%d٪.", int((ratio-1)*100))
			tone = "WARN"
		case ratio <= 0.8:
			line += " — أسرع من المعدل."
		default:
			line += " — ضمن الطبيعي."
		}
		l := rl(line, tone)
		if tone != "OK" {
			l = withFix(l, "", "وزّع حجوزاته الكبيرة على كادر أكبر.", "تدريب على الخدمات الي يطوّل بيها (شوف أبطأ الحجوزات تحت).",
				"اسأله عن العوائق: طريق، مواد ناقصة، زبون غير جاهز.")
		}
		*p = append(*p, l)
		if prev != nil && prev.Timed >= 3 && prev.AvgExpected > 0 {
			pr := prev.AvgActual / prev.AvgExpected
			delta := (ratio - pr) / pr * 100
			switch {
			case delta >= 20:
				*p = append(*p, withFix(rl(Fmt("↓ صار يطوّل بالحجز %d٪ أكثر من الشهر الي قبله.", int(delta)), "WARN"), "",
					"شوف إذا تغيّر نوع شغله أو فريقه هالشهر.", "جلسة قصيرة ويّاه تسأله شنو تغيّر."))
			case delta <= -20:
				*p = append(*p, rl(Fmt("↑ صار أسرع بـ%d٪ من الشهر الي قبله.", int(-delta)), "OK"))
			}
		}
	}
	if cur.Outliers > 0 {
		*p = append(*p, withFix(rl(Fmt("%d حجز مدته فوگ ١٤ ساعة — انعزلت من الحساب (غالباً ما تسكّرت بوقتها).", cur.Outliers), "INFO"), "",
			"ذكّره يسكّر الحجز أول ما يخلص حتى تطلع أرقامه صحيحة."))
	}
	if cur.GroupPartialRate >= 0 && cur.Total >= 4 {
		mine := float64(cur.Partial) / float64(cur.Total)
		if mine > cur.GroupPartialRate*1.5 && mine >= 0.2 {
			*p = append(*p, withFix(rl(Fmt("نسبة الجزئي عنده %d٪ مقابل %d٪ لباقي الكادر.", int(mine*100), int(cur.GroupPartialRate*100)), "WARN"), "",
				"تأكد من تجهيز المواد قبل الطلعة.", "راجع أسباب الجزئي المكتوبة بحجوزاته."))
		}
	}
	for _, j := range cur.Slow {
		rep.SlowJobs = append(rep.SlowJobs, SlowJob{BookingID: j.ID, Code: j.Code, Service: j.Service, Actual: j.Actual, Expected: j.Expected})
	}
}
