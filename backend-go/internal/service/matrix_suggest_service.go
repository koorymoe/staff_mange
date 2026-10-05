package service

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"staffmange-api/internal/repository"
)

// ═══ المرحلة الثانية: ماتركس يقترح والمنسق يقرر ═══
//
// قرار (ع) 10-05. ماتركس يقترح **موعد** و**كادر** لكل حجز، والمنسق يوافق أو
// يغيّر — ماتركس ما ينفّذ شي لحاله. كل اقتراح ينحفظ، وبعدين ينقارن بالي
// اختاره المنسق وبشنو صار بالحجز، فتنقاس دقته (٪ قبول، وهل الحجوزات الي
// مشت على اقتراحه طلعت بوقتها أكثر).
//
// كل رقم من بيانات الشركة: مدة الخدمة (الوسيط الحقيقي)، ساعات وأيام الدوام
// المعتادة، منو مشغول بنفس الوقت، منو بإجازة، خبرة كل واحد بنفس الخدمة،
// سرعته بيها، وتقييم الليدرية للفنيين.

const travelBufferMin = 60 // وقت الطريق بين حجزين لنفس الشخص

type MatrixSuggestService struct {
	repo *repository.MatrixSuggestRepository
}

func NewMatrixSuggestService(repo *repository.MatrixSuggestRepository) *MatrixSuggestService {
	return &MatrixSuggestService{repo: repo}
}

type SuggestedSlot struct {
	At          time.Time `json:"at"`
	DurationMin int       `json:"durationMin"`
	FreeLeaders int       `json:"freeLeaders"`
}

type SuggestedPerson struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Why    []string `json:"why"`
	Score  float64  `json:"score"`
	IsLead bool     `json:"isLeader"`
}

type SuggestedCrew struct {
	Leader     *SuggestedPerson  `json:"leader"`
	Techs      []SuggestedPerson `json:"techs"`
	AltLeaders []SuggestedPerson `json:"altLeaders"`
	AltTechs   []SuggestedPerson `json:"altTechs"`
	For        time.Time         `json:"for"`
	CrewSize   int               `json:"crewSize"`
	Solo       bool              `json:"solo"`
}

type BookingSuggestion struct {
	BookingID   string         `json:"bookingId"`
	Code        string         `json:"code"`
	Schedule    *SuggestedSlot `json:"schedule"`
	ScheduleWhy []string       `json:"scheduleWhy"`
	Crew        *SuggestedCrew `json:"crew"`
	CrewWhy     []string       `json:"crewWhy"`
	Note        string         `json:"note"`
}

func (s *MatrixSuggestService) duration(serviceID *string) int {
	per, all := s.repo.ServiceDurations()
	d := all
	if serviceID != nil {
		if v, ok := per[*serviceID]; ok {
			d = v
		}
	}
	if d < 60 {
		d = 60
	}
	if d > 8*60 {
		d = 8 * 60
	}
	return d
}

func overlaps(aStart time.Time, aMin int, bStart time.Time, bMin int) bool {
	aEnd := aStart.Add(time.Duration(aMin+travelBufferMin) * time.Minute)
	bEnd := bStart.Add(time.Duration(bMin+travelBufferMin) * time.Minute)
	return aStart.Before(bEnd) && bStart.Before(aEnd)
}

// busyMap لكل موظف: أوقات حجوزاته الحية حول الفترة.
func (s *MatrixSuggestService) busyMap(from, to time.Time, except string) map[string][]repository.Busy {
	rows, _ := s.repo.BusyBetween(from.Add(-12*time.Hour), to.Add(12*time.Hour), except)
	m := map[string][]repository.Busy{}
	for _, r := range rows {
		m[r.EmployeeID] = append(m[r.EmployeeID], r)
	}
	return m
}

func (s *MatrixSuggestService) isFree(busy map[string][]repository.Busy, id string, at time.Time, dur int, durOf func(*string) int) bool {
	for _, b := range busy[id] {
		if overlaps(at, dur, b.ScheduledAt, durOf(b.ServiceID)) {
			return false
		}
	}
	return true
}

func dayLoad(busy map[string][]repository.Busy, id string, day time.Time) int {
	n := 0
	y, m, d := day.In(debriefLoc).Date()
	for _, b := range busy[id] {
		by, bm, bd := b.ScheduledAt.In(debriefLoc).Date()
		if by == y && bm == m && bd == d {
			n++
		}
	}
	return n
}

// Suggest الاقتراح الكامل لحجز، وينحفظ بسجل الدقة.
func (s *MatrixSuggestService) Suggest(bookingID string) (*BookingSuggestion, error) {
	b, err := s.repo.Booking(bookingID)
	if err != nil {
		return nil, fmt.Errorf("الحجز مو موجود")
	}
	out := &BookingSuggestion{BookingID: b.ID, Code: b.Code, ScheduleWhy: []string{}, CrewWhy: []string{}}
	if b.Status == "COMPLETED" || b.Status == "CANCELLED" || b.Status == "PARTIAL" {
		out.Note = "الحجز منتهي — ماكو اقتراح."
		return out, nil
	}
	per, all := s.repo.ServiceDurations()
	durOf := func(sid *string) int {
		if sid != nil {
			if v, ok := per[*sid]; ok {
				return v
			}
		}
		return all
	}
	dur := s.duration(b.ServiceID)
	now := time.Now()

	// ═══ الموعد ═══
	at := b.ScheduledAt
	if at == nil {
		slot, why := s.suggestSlot(b, dur, durOf, now)
		out.Schedule, out.ScheduleWhy = slot, why
		if slot != nil {
			at = &slot.At
			s.repo.Log(b.ID, "SCHEDULE", map[string]any{"at": slot.At}, strings.Join(why, " "))
		}
	} else {
		out.ScheduleWhy = append(out.ScheduleWhy, "الحجز بيه موعد أصلاً — الاقتراح للكادر بس.")
	}

	// ═══ الكادر ═══
	if at != nil && b.SupervisorID == nil && len(b.TechIDs) == 0 {
		crew, why := s.suggestCrew(b, *at, dur, durOf)
		out.Crew, out.CrewWhy = crew, why
		if crew != nil && (crew.Leader != nil || len(crew.Techs) > 0) {
			ids := []string{}
			for _, t := range crew.Techs {
				ids = append(ids, t.ID)
			}
			lead := ""
			if crew.Leader != nil {
				lead = crew.Leader.ID
			}
			s.repo.Log(b.ID, "CREW", map[string]any{"leader": lead, "techs": ids}, strings.Join(why, " "))
		}
	} else if at != nil {
		out.CrewWhy = append(out.CrewWhy, "الحجز بيه كادر أصلاً.")
	}
	return out, nil
}

func (s *MatrixSuggestService) suggestSlot(b *repository.SuggestBooking, dur int, durOf func(*string) int, now time.Time) (*SuggestedSlot, []string) {
	fromH, toH, off := s.repo.WorkHours()
	isOff := map[int]bool{}
	for _, d := range off {
		isOff[d] = true
	}
	horizon := now.AddDate(0, 0, 10)
	busy := s.busyMap(now, horizon, b.ID)
	people, _ := s.repo.People(b.ServiceID, now)
	leaders := []string{}
	for _, p := range people {
		if p.IsLeader {
			leaders = append(leaders, p.ID)
		}
	}
	earliest := now.Add(2 * time.Hour)
	for d := 0; d <= 10; d++ {
		day := now.In(debriefLoc).AddDate(0, 0, d)
		if isOff[int(day.Weekday())] {
			continue
		}
		for h := fromH; h <= toH; h++ {
			at := time.Date(day.Year(), day.Month(), day.Day(), h, 0, 0, 0, debriefLoc)
			if at.Before(earliest) {
				continue
			}
			free := 0
			for _, id := range leaders {
				if s.isFree(busy, id, at, dur, durOf) {
					free++
				}
			}
			need := 1
			if b.Solo {
				need = 0 // الخدمة الفردية يكفيها فني
			}
			if free >= need && (need > 0 || len(people) > 0) {
				first := fmt.Sprintf("أقرب وقت بيه %d ليدر فاضي.", free)
				if b.Solo {
					first = "أقرب وقت ضمن الدوام (خدمة فردية — الفني ينختار بالكادر)."
				}
				why := []string{
					first,
					fmt.Sprintf("مدة الشغل المتوقعة %s (وسيط الشركة لهالخدمة) + ساعة طريق.", FmtMinutes(dur)),
					fmt.Sprintf("ضمن ساعات المواعيد المعتادة (%d:00–%d:00).", fromH, toH),
				}
				if len(leaders) == 0 {
					why = append(why, "⚠️ ماكو ليدرية فعّالين بالنظام.")
				}
				return &SuggestedSlot{At: at, DurationMin: dur, FreeLeaders: free}, why
			}
		}
	}
	return nil, []string{"ما لگيت وقت فاضي خلال ١٠ أيام — الكوادر كلها مشغولة."}
}

func (s *MatrixSuggestService) suggestCrew(b *repository.SuggestBooking, at time.Time, dur int, durOf func(*string) int) (*SuggestedCrew, []string) {
	people, err := s.repo.People(b.ServiceID, at.In(debriefLoc))
	if err != nil {
		return nil, []string{"تعذر جلب الكوادر."}
	}
	busy := s.busyMap(at, at, b.ID)
	size := s.repo.CrewSize(b.ServiceID)
	crew := &SuggestedCrew{For: at, CrewSize: size, Solo: b.Solo, Techs: []SuggestedPerson{}, AltLeaders: []SuggestedPerson{}, AltTechs: []SuggestedPerson{}}
	why := []string{}

	score := func(p repository.SuggestPerson) SuggestedPerson {
		sp := SuggestedPerson{ID: p.ID, Name: p.Name, IsLead: p.IsLeader, Why: []string{}}
		if p.SameDone > 0 {
			sp.Score += math.Min(float64(p.SameDone), 10) * 2
			sp.Why = append(sp.Why, fmt.Sprintf("سوّى %d حجز بنفس الخدمة", p.SameDone))
		} else {
			sp.Why = append(sp.Why, "ما عنده خبرة مسجّلة بهالخدمة")
		}
		if p.WorkMed != nil && *p.WorkMed > 0 && dur > 0 && p.SameDone >= 2 {
			r := *p.WorkMed / float64(dur)
			sp.Score += (1 - r) * 8 // أسرع من الوسيط = أعلى
			if r <= 0.85 {
				sp.Why = append(sp.Why, fmt.Sprintf("أسرع من المعتاد (%s مقابل %s)", FmtMinutes(int(*p.WorkMed)), FmtMinutes(dur)))
			} else if r >= 1.3 {
				sp.Why = append(sp.Why, fmt.Sprintf("أبطأ من المعتاد (%s مقابل %s)", FmtMinutes(int(*p.WorkMed)), FmtMinutes(dur)))
			}
		}
		if p.RatingAvg != nil && p.RatingN >= 2 {
			sp.Score += (*p.RatingAvg - 3) * 3
			sp.Why = append(sp.Why, fmt.Sprintf("تقييم الليدرية %.1f/5", *p.RatingAvg))
		}
		load := dayLoad(busy, p.ID, at)
		sp.Score -= float64(load) * 3
		if load > 0 {
			sp.Why = append(sp.Why, fmt.Sprintf("عنده %d حجز ثاني بنفس اليوم", load))
		} else {
			sp.Why = append(sp.Why, "يومه فاضي")
		}
		return sp
	}

	var leaders, techs []SuggestedPerson
	busyNames := 0
	for _, p := range people {
		if !s.isFree(busy, p.ID, at, dur, durOf) {
			busyNames++
			continue
		}
		sp := score(p)
		if p.IsLeader {
			leaders = append(leaders, sp)
		} else {
			techs = append(techs, sp)
		}
	}
	sort.SliceStable(leaders, func(i, j int) bool { return leaders[i].Score > leaders[j].Score })
	sort.SliceStable(techs, func(i, j int) bool { return techs[i].Score > techs[j].Score })

	if !b.Solo {
		if len(leaders) > 0 {
			crew.Leader = &leaders[0]
			if len(leaders) > 1 {
				crew.AltLeaders = leaders[1:minInt(3, len(leaders))]
			}
		} else {
			why = append(why, "⚠️ ماكو ليدر فاضي بهالوقت — غيّر الموعد أو حرّر ليدر.")
		}
	} else {
		size = 1
		crew.CrewSize = 1
		why = append(why, "خدمة فردية (داش كام/جي بي اس) — فني واحد بلا ليدر.")
	}
	n := minInt(size, len(techs))
	crew.Techs = techs[:n]
	if len(techs) > n {
		crew.AltTechs = techs[n:minInt(n+3, len(techs))]
	}
	if n < size {
		why = append(why, fmt.Sprintf("⚠️ المعتاد %d فني، والفاضي بس %d.", size, n))
	}
	if b.Solo {
		why = append(why, fmt.Sprintf("%d شخص مشغول بنفس الوقت انشال.", busyNames))
	} else {
		why = append(why, fmt.Sprintf("المعتاد لهالخدمة %d فني ويا الليدر. %d شخص مشغول بنفس الوقت انشال.", size, busyNames))
	}
	return crew, why
}

// ═══ الدقة: المنسق قبل الاقتراح لو غيّره؟ وشنو صار بعدين؟ ═══

// Evaluate يحسم الاقتراحات المعلّقة الي صار بحجزها موعد/كادر.
func (s *MatrixSuggestService) Evaluate() {
	rows, err := s.repo.Pending()
	if err != nil {
		return
	}
	for _, r := range rows {
		b, err := s.repo.Booking(r.BookingID)
		if err != nil {
			continue
		}
		switch r.Kind {
		case "SCHEDULE":
			if b.ScheduledAt == nil || time.Since(r.CreatedAt) < 10*time.Minute {
				continue
			}
			var sg struct {
				At time.Time `json:"at"`
			}
			_ = json.Unmarshal(r.Suggested, &sg)
			diff := math.Abs(b.ScheduledAt.Sub(sg.At).Minutes())
			out := "CHANGED"
			switch {
			case diff <= 60:
				out = "ACCEPTED"
			case sameDay(*b.ScheduledAt, sg.At):
				out = "PARTIAL"
			}
			s.repo.Decide(r.ID, out, map[string]any{"at": b.ScheduledAt})
		case "CREW":
			// ننتظر ساعة بعد أول تكليف حتى المنسق يكمّل الكادر.
			if (b.SupervisorID == nil && len(b.TechIDs) == 0) || time.Since(r.CreatedAt) < time.Hour {
				continue
			}
			var sg struct {
				Leader string   `json:"leader"`
				Techs  []string `json:"techs"`
			}
			_ = json.Unmarshal(r.Suggested, &sg)
			leaderOK := sg.Leader == "" || (b.SupervisorID != nil && *b.SupervisorID == sg.Leader)
			have := map[string]bool{}
			for _, t := range b.TechIDs {
				have[t] = true
			}
			hit := 0
			for _, t := range sg.Techs {
				if have[t] {
					hit++
				}
			}
			techOK := len(sg.Techs) == 0 || hit*2 >= len(sg.Techs)
			out := "CHANGED"
			switch {
			case leaderOK && techOK:
				out = "ACCEPTED"
			case leaderOK || hit > 0:
				out = "PARTIAL"
			}
			lead := ""
			if b.SupervisorID != nil {
				lead = *b.SupervisorID
			}
			s.repo.Decide(r.ID, out, map[string]any{"leader": lead, "techs": []string(b.TechIDs)})
		}
	}
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.In(debriefLoc).Date()
	by, bm, bd := b.In(debriefLoc).Date()
	return ay == by && am == bm && ad == bd
}

type SuggestKindStats struct {
	Kind      string `json:"kind"`
	Title     string `json:"title"`
	Decided   int    `json:"decided"`
	Accepted  int    `json:"accepted"`
	Partial   int    `json:"partial"`
	Changed   int    `json:"changed"`
	AcceptPct int    `json:"acceptPct"`
	OnTimeAcc *int   `json:"onTimeAccepted"` // ٪ الي طلعت بوقتها لمن مشى اقتراحه
	OnTimeChg *int   `json:"onTimeChanged"`  // ٪ الي طلعت بوقتها لمن انغيّر
	SampleAcc int    `json:"sampleAccepted"`
	SampleChg int    `json:"sampleChanged"`
}

type SuggestAccuracy struct {
	Days     int                `json:"days"`
	Kinds    []SuggestKindStats `json:"kinds"`
	Pending  int                `json:"pending"`
	Recent   []map[string]any   `json:"recent"`
	Insights []string           `json:"insights"`
}

func (s *MatrixSuggestService) Accuracy(days int) (*SuggestAccuracy, error) {
	if days <= 0 || days > 180 {
		days = 30
	}
	s.Evaluate()
	rows, err := s.repo.Decided(days)
	if err != nil {
		return nil, err
	}
	pend, _ := s.repo.Pending()
	out := &SuggestAccuracy{Days: days, Pending: len(pend), Recent: []map[string]any{}, Insights: []string{}}
	titles := map[string]string{"SCHEDULE": "اقتراح الموعد", "CREW": "اقتراح الكادر"}
	type acc struct {
		st               SuggestKindStats
		onA, nA, onC, nC int
	}
	m := map[string]*acc{"SCHEDULE": {st: SuggestKindStats{Kind: "SCHEDULE", Title: titles["SCHEDULE"]}}, "CREW": {st: SuggestKindStats{Kind: "CREW", Title: titles["CREW"]}}}
	outLabel := map[string]string{"ACCEPTED": "✅ اعتمده", "PARTIAL": "🟡 اعتمد جزء منه", "CHANGED": "✏️ غيّره"}
	for i, r := range rows {
		a := m[r.Kind]
		if a == nil {
			continue
		}
		a.st.Decided++
		switch r.Outcome {
		case "ACCEPTED":
			a.st.Accepted++
		case "PARTIAL":
			a.st.Partial++
		case "CHANGED":
			a.st.Changed++
		}
		// شنو صار بعدين: طلع بوقته؟ (بدأ خلال ٣٠ دقيقة من الموعد)
		if r.StartedAt != nil && r.ScheduledAt != nil {
			onTime := r.StartedAt.Sub(*r.ScheduledAt) <= 30*time.Minute
			if r.Outcome == "ACCEPTED" {
				a.nA++
				if onTime {
					a.onA++
				}
			} else if r.Outcome == "CHANGED" {
				a.nC++
				if onTime {
					a.onC++
				}
			}
		}
		if i < 15 {
			out.Recent = append(out.Recent, map[string]any{"bookingId": r.BookingID, "code": r.Code, "kind": titles[r.Kind],
				"outcome": outLabel[r.Outcome], "at": r.DecidedAt})
		}
	}
	for _, k := range []string{"SCHEDULE", "CREW"} {
		a := m[k]
		if a.st.Decided > 0 {
			a.st.AcceptPct = int(math.Round(float64(a.st.Accepted) * 100 / float64(a.st.Decided)))
		}
		if a.nA > 0 {
			p := a.onA * 100 / a.nA
			a.st.OnTimeAcc, a.st.SampleAcc = &p, a.nA
		}
		if a.nC > 0 {
			p := a.onC * 100 / a.nC
			a.st.OnTimeChg, a.st.SampleChg = &p, a.nC
		}
		out.Kinds = append(out.Kinds, a.st)
		if a.st.Decided == 0 {
			out.Insights = append(out.Insights, fmt.Sprintf("%s: لحد هسه ماكو اقتراح قرر بيه المنسق.", a.st.Title))
			continue
		}
		line := fmt.Sprintf("%s: المنسقين اعتمدوه %d مرة من أصل %d (%d٪).", a.st.Title, a.st.Accepted, a.st.Decided, a.st.AcceptPct)
		if a.st.OnTimeAcc != nil && a.st.OnTimeChg != nil && a.nA >= 5 && a.nC >= 5 {
			line += fmt.Sprintf(" الحجوزات الي اعتمدوا بيها اقتراحه طلعت بوقتها %d٪، والي غيّروه %d٪.", *a.st.OnTimeAcc, *a.st.OnTimeChg)
		} else if a.st.Decided < 20 {
			line += " العدد بعده قليل، نحتاج ٢٠ اقتراح على الأقل حتى نحكم على دقته."
		}
		out.Insights = append(out.Insights, line)
	}
	return out, nil
}

// CrewFor اقتراح الكادر لحجز بموعده الحالي — للفعل التنفيذي (بلا تسجيل اقتراح).
func (s *MatrixSuggestService) CrewFor(bookingID string) (*repository.SuggestBooking, *SuggestedCrew, []string, error) {
	b, err := s.repo.Booking(bookingID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("الحجز مو موجود")
	}
	if b.ScheduledAt == nil {
		return b, nil, nil, fmt.Errorf("الحجز بلا موعد")
	}
	per, all := s.repo.ServiceDurations()
	durOf := func(sid *string) int {
		if sid != nil {
			if v, ok := per[*sid]; ok {
				return v
			}
		}
		return all
	}
	crew, why := s.suggestCrew(b, *b.ScheduledAt, s.duration(b.ServiceID), durOf)
	return b, crew, why, nil
}

// Repo للخدمات الثانية.
func (s *MatrixSuggestService) Repo() *repository.MatrixSuggestRepository { return s.repo }
