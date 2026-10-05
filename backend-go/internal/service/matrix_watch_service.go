package service

import (
	"fmt"
	"math"
	"sort"
	"time"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ═══ عيون الرقابة: الكميات، السيارات، تقييمات الزبائن ═══
//
// طلب (ع) 10-05: «لازم اكو رقابه بكل مكان»، و«هاي ايضاً تروح للمراقب ويا
// تحليل وتفكير ماتركس». كل عين ترجع بنود بحكم ماتركس (شنو الغلط، شكد،
// منو المسؤول، وشنو يقترح)، ويشوفها المراقب والمدير والمالك.
// كلها قواعد على الجداول — ماكو ذكاء اصطناعي، فماكو كلفة. وماكو عقوبة.

const (
	WatchHigh   = "HIGH"
	WatchMedium = "MEDIUM"
	WatchLow    = "LOW"
)

type EyeItem struct {
	Eye       string     `json:"eye"`
	Kind      string     `json:"kind"`
	Severity  string     `json:"severity"`
	Title     string     `json:"title"`
	Detail    string     `json:"detail"`
	Advice    string     `json:"advice"`
	OwnerID   string     `json:"ownerId,omitempty"`
	OwnerName string     `json:"ownerName,omitempty"`
	BookingID string     `json:"bookingId,omitempty"`
	At        *time.Time `json:"at,omitempty"`
}

type EyeGroup struct {
	Key     string    `json:"key"`
	Title   string    `json:"title"`
	Items   []EyeItem `json:"items"`
	Summary string    `json:"summary"`
	High    int       `json:"high"`
}

type CustomerScore struct {
	repository.CustomerScoreRow
	HappyPct *int     `json:"happyPct"`
	Notes    []string `json:"notes"`
}

type EyesReport struct {
	Eyes      []EyeGroup      `json:"eyes"`
	Customers []CustomerScore `json:"customers"`
	TeamHappy *int            `json:"teamHappyPct"`
	Insights  []string        `json:"insights"`
}

type MatrixWatchService struct {
	repo  *repository.MatrixWatchRepository
	auto  *MatrixAutopilotService
	notif *repository.NotificationRepository
}

func NewMatrixWatchService(repo *repository.MatrixWatchRepository, auto *MatrixAutopilotService, notif *repository.NotificationRepository) *MatrixWatchService {
	return &MatrixWatchService{repo: repo, auto: auto, notif: notif}
}

func num(f float64) string {
	if f == math.Trunc(f) {
		return fmt.Sprintf("%.0f", f)
	}
	return fmt.Sprintf("%.1f", f)
}

func ptr(t time.Time) *time.Time { return &t }

func plate(name string, p *string) string {
	if p != nil && *p != "" {
		return name + " (" + *p + ")"
	}
	return name
}

func (s *MatrixWatchService) stockEye() EyeGroup {
	eye := EyeGroup{Key: "STOCK", Title: "الكميات والمخزن", Items: []EyeItem{}}
	if rows, err := s.repo.NegativeStock(); err == nil {
		for _, r := range rows {
			left := r.StockQty - r.Used
			eye.Items = append(eye.Items, EyeItem{Eye: eye.Key, Kind: "NEGATIVE_STOCK", Severity: WatchHigh,
				Title:  fmt.Sprintf("«%s» رصيده بالحساب صار %s", r.Name, num(left)),
				Detail: fmt.Sprintf("آخر جرد چان %s، وانصرف منه بالفواتير بعدها %s. يعني انصرف أكثر من الموجود.", num(r.StockQty), num(r.Used)),
				Advice: "يا أكو شراء ما انسجّل، يا الجرد غلط، يا أكو صرف بالفواتير أكثر من الحقيقي. لازم جرد جديد ويقارنه ويا وصولات الشراء.",
				At:     r.CountedAt})
		}
	}
	if rows, err := s.repo.StaleCount(); err == nil {
		for _, r := range rows {
			when := "ما انجرد أبداً"
			if r.CountedAt != nil {
				when = fmt.Sprintf("آخر جرد قبل %d يوم", int(time.Since(*r.CountedAt).Hours()/24))
			}
			eye.Items = append(eye.Items, EyeItem{Eye: eye.Key, Kind: "STALE_COUNT", Severity: WatchLow,
				Title:  fmt.Sprintf("«%s» يصرف وما انجرد", r.Name),
				Detail: fmt.Sprintf("انصرف منه %s بآخر ٣٠ يوم، و%s.", num(r.Used), when),
				Advice: "جرده هالأسبوع حتى الرصيد يصير حقيقي."})
		}
	}
	if rows, err := s.repo.ToolsNotReturned(); err == nil {
		for _, r := range rows {
			days := int(time.Since(r.ApprovedAt).Hours() / 24)
			sev := WatchMedium
			if days > 21 {
				sev = WatchHigh
			}
			eye.Items = append(eye.Items, EyeItem{Eye: eye.Key, Kind: "TOOL_NOT_RETURNED", Severity: sev,
				Title:   fmt.Sprintf("«%s» ويا %s من %d يوم وما رجعت", r.Tool, r.Employee, days),
				Detail:  "الأداة انطلعت بطلب معتمد وبعد ما انسجّل رجوعها.",
				Advice:  "ماتركس يذكّره كل يوم يرجعها، وإذا ضايعة يسجّلها «بدل مفقود».",
				OwnerID: r.EmployeeID, OwnerName: r.Employee, At: ptr(r.ApprovedAt)})
		}
	}
	if rows, err := s.repo.ToolCountsWrong(); err == nil {
		for _, r := range rows {
			eye.Items = append(eye.Items, EyeItem{Eye: eye.Key, Kind: "TOOL_COUNT", Severity: WatchMedium,
				Title:  fmt.Sprintf("أرقام «%s» ما تركب", r.Name),
				Detail: fmt.Sprintf("المتوفر %d والكلي %d.", r.Available, r.Total),
				Advice: "أكو رجوع أو صرف انسجّل مرتين. لازم يتصحح الرقم بالمخزن."})
		}
	}
	if rows, err := s.repo.ShortagesOpen(); err == nil {
		for _, r := range rows {
			what := "ناقص شي"
			if r.Missing != nil && *r.Missing != "" {
				what = *r.Missing
			}
			when := "قبل الحجز"
			if r.Source == "AFTER" {
				when = "بعد الحجز"
			}
			code := ""
			if r.Code != nil {
				code = " (" + *r.Code + ")"
			}
			it := EyeItem{Eye: eye.Key, Kind: "SHORTAGE", Severity: WatchMedium,
				Title:   fmt.Sprintf("نقص بعدّة %s %s%s", r.Employee, when, code),
				Detail:  "الناقص: " + what + ".",
				Advice:  "مسؤول المخزن يعوّضه أو يتحقق وين راح، ويأشّره «انحل».",
				OwnerID: r.EmployeeID, OwnerName: r.Employee, At: ptr(r.At)}
			if r.BookingID != nil {
				it.BookingID = *r.BookingID
			}
			eye.Items = append(eye.Items, it)
		}
	}
	if rows, err := s.repo.OddInvoiceQty(); err == nil {
		for _, r := range rows {
			it := EyeItem{Eye: eye.Key, Kind: "ODD_QTY", Severity: WatchMedium,
				Title:   fmt.Sprintf("فاتورة %s بيها «%s» بكمية %s", r.Leader, r.Material, num(r.Qty)),
				Detail:  fmt.Sprintf("المعتاد لهالمادة %s بالفاتورة، يعني هاي %s أضعاف.", num(r.Median), num(r.Qty/r.Median)),
				Advice:  "يتأكد المحاسب إذا الشغل فعلاً احتاج هالكمية، أو انكتب رقم غلط.",
				OwnerID: r.LeaderID, OwnerName: r.Leader, At: ptr(r.At)}
			if r.Booking != nil {
				it.BookingID = *r.Booking
			}
			eye.Items = append(eye.Items, it)
		}
	}
	return finishEye(eye, "المخزن ماشي مضبوط حسب البيانات.")
}

func (s *MatrixWatchService) vehicleEye() EyeGroup {
	eye := EyeGroup{Key: "VEHICLES", Title: "السيارات", Items: []EyeItem{}}
	name := func(p *string) string {
		if p == nil {
			return ""
		}
		return *p
	}
	if rows, err := s.repo.FuelAbnormal(); err == nil {
		for _, r := range rows {
			eye.Items = append(eye.Items, EyeItem{Eye: eye.Key, Kind: "FUEL", Severity: WatchHigh,
				Title:   fmt.Sprintf("%s صرفت وقود أكثر من عادتها", plate(r.Vehicle, r.Plate)),
				Detail:  fmt.Sprintf("%s لتر لـ%s كم (%s لتر لكل ١٠٠ كم)، ومعدل نفس السيارة %s.", num(r.Liters), num(r.Km), num(r.Per100), num(r.Avg100)),
				Advice:  "يتأكد من وصل التعبئة وقراءة العدّاد، أو أكو عطل بالسيارة.",
				OwnerID: name(r.ByID), OwnerName: name(r.By), At: ptr(r.At)})
		}
	}
	if rows, err := s.repo.OdometerGaps(); err == nil {
		for _, r := range rows {
			it := EyeItem{Eye: eye.Key, Kind: "ODOMETER", Severity: WatchHigh, OwnerID: name(r.DriverID), OwnerName: name(r.Driver), At: ptr(r.At)}
			if r.Start < r.PrevEnd {
				it.Title = fmt.Sprintf("عدّاد %s رجع لورا", plate(r.Vehicle, r.Plate))
				it.Detail = fmt.Sprintf("المهمة الي قبلها خلصت على %d، وهاي بدت على %d.", r.PrevEnd, r.Start)
				it.Advice = "يا رقم انكتب غلط، يا أكو تلاعب بالقراءة. يتأكد المراقب."
			} else {
				it.Title = fmt.Sprintf("%s مشت %d كم بلا مهمة مسجّلة", plate(r.Vehicle, r.Plate), r.Start-r.PrevEnd)
				it.Detail = fmt.Sprintf("آخر مهمة خلصت على %d، والمهمة الي بعدها بدت على %d.", r.PrevEnd, r.Start)
				it.Advice = "منو طلع بالسيارة بين المهمتين؟ لازم كل طلعة تنسجّل."
				it.Severity = WatchMedium
			}
			eye.Items = append(eye.Items, it)
		}
	}
	if rows, err := s.repo.MissionsUnbooked(); err == nil {
		for _, r := range rows {
			eye.Items = append(eye.Items, EyeItem{Eye: eye.Key, Kind: "UNBOOKED", Severity: WatchLow,
				Title:   fmt.Sprintf("%s طلعت بلا حجز سيارة معتمد", plate(r.Vehicle, r.Plate)),
				Detail:  "المهمة انسجّلت بس ماكو حجز للسيارة بنفس الوقت.",
				Advice:  "الطلعات لازم تمر بحجز السيارة حتى تنعرف منو وليش.",
				OwnerID: name(r.DriverID), OwnerName: name(r.Driver), At: ptr(r.At)})
		}
	}
	if rows, err := s.repo.MissionsOpen(); err == nil {
		for _, r := range rows {
			eye.Items = append(eye.Items, EyeItem{Eye: eye.Key, Kind: "OPEN_MISSION", Severity: WatchMedium,
				Title:   fmt.Sprintf("مهمة %s مفتوحة من %s", plate(r.Vehicle, r.Plate), FmtMinutes(int(time.Since(r.At).Minutes()))),
				Detail:  "السائق ما سكّر المهمة ولا سجّل قراءة العدّاد بالنهاية.",
				Advice:  "يسكّرها ويكتب العدّاد، وإلا حساب الكيلومترات يخرب.",
				OwnerID: name(r.DriverID), OwnerName: name(r.Driver), At: ptr(r.At)})
		}
	}
	if rows, err := s.repo.MaintenanceOverdue(); err == nil {
		types := map[string]string{"OIL_CHANGE": "تبديل الدهن", "MAINTENANCE": "الصيانة", "CLEANING": "التنظيف"}
		for _, r := range rows {
			t := types[r.Type]
			if t == "" {
				t = r.Type
			}
			d := ""
			if r.DueAt != nil && r.DueAt.Before(time.Now()) {
				d = fmt.Sprintf("موعده چان %s.", r.DueAt.In(debriefLoc).Format("2006-01-02"))
			}
			if r.DueOdo != nil && r.Odometer != nil && *r.Odometer > *r.DueOdo {
				d += fmt.Sprintf(" تجاوز العدّاد بـ%d كم.", *r.Odometer-*r.DueOdo)
			}
			eye.Items = append(eye.Items, EyeItem{Eye: eye.Key, Kind: "MAINTENANCE", Severity: WatchMedium,
				Title: fmt.Sprintf("%s فاتها %s", plate(r.Vehicle, r.Plate), t), Detail: d,
				Advice: "تتسوّى هالأسبوع قبل ما تصير عطلة أغلى."})
		}
	}
	if rows, err := s.repo.FuelNoProof(); err == nil {
		for _, r := range rows {
			why := []string{}
			if r.NoReceipt {
				why = append(why, "بلا صورة وصل")
			}
			if r.NoMission {
				why = append(why, "والي عبّى ما عنده مهمة بالسيارة بنفس اليوم")
			}
			cost := ""
			if r.Cost != nil {
				cost = fmt.Sprintf(" بـ%s د.ع", num(*r.Cost))
			}
			eye.Items = append(eye.Items, EyeItem{Eye: eye.Key, Kind: "FUEL_PROOF", Severity: WatchLow,
				Title:   fmt.Sprintf("تعبئة %s%s", plate(r.Vehicle, r.Plate), cost),
				Detail:  joinAr(why) + ".",
				Advice:  "كل تعبئة لازم وياها وصل، وتكون بطلعة مسجّلة.",
				OwnerID: name(r.ByID), OwnerName: name(r.By), At: ptr(r.At)})
		}
	}
	return finishEye(eye, "السيارات ماشية مضبوطة حسب البيانات.")
}

func joinAr(xs []string) string {
	out := ""
	for i, x := range xs {
		if i > 0 {
			out += " "
		}
		out += x
	}
	return out
}

func (s *MatrixWatchService) customerEye() (EyeGroup, []CustomerScore, *int) {
	eye := EyeGroup{Key: "CUSTOMERS", Title: "رأي الزبائن", Items: []EyeItem{}}
	rows, err := s.repo.CustomerScores()
	scores := []CustomerScore{}
	if err != nil {
		return finishEye(eye, ""), scores, nil
	}
	happy, total := 0, 0
	for _, r := range rows {
		happy += r.Happy
		total += r.Happy + r.Unhappy
	}
	var team *int
	if total > 0 {
		t := int(math.Round(float64(happy) * 100 / float64(total)))
		team = &t
	}
	for _, r := range rows {
		cs := CustomerScore{CustomerScoreRow: r, Notes: []string{}}
		if n := r.Happy + r.Unhappy; n > 0 {
			p := int(math.Round(float64(r.Happy) * 100 / float64(n)))
			cs.HappyPct = &p
			if r.PrevHappyPct != nil {
				d := p - int(math.Round(*r.PrevHappyPct))
				if d >= 10 {
					cs.Notes = append(cs.Notes, fmt.Sprintf("📈 رضا الزبائن عنه زاد من %.0f٪ إلى %d٪.", *r.PrevHappyPct, p))
				} else if d <= -10 {
					cs.Notes = append(cs.Notes, fmt.Sprintf("📉 رضا الزبائن عنه نزل من %.0f٪ إلى %d٪.", *r.PrevHappyPct, p))
				}
			}
			if team != nil && n >= 5 && p <= *team-20 {
				eye.Items = append(eye.Items, EyeItem{Eye: eye.Key, Kind: "LOW_SATISFACTION", Severity: WatchMedium,
					Title:   fmt.Sprintf("رضا الزبائن عن %s أقل من زملائه", r.Name),
					Detail:  fmt.Sprintf("%d٪ من زبائنه راضين (من %d متابعة)، والفريق %d٪.", p, n, *team),
					Advice:  "مهندس الجودة يراجع ملاحظات زبائنه، والمراقب يحچي وياه بشنو يتكرر.",
					OwnerID: r.EmployeeID, OwnerName: r.Name})
			}
		}
		if r.Complaints30 >= 2 {
			avg := ""
			if r.AvgRating != nil {
				avg = fmt.Sprintf("، ومعدل تقييم الزبائن بشكاواهم %.1f", *r.AvgRating)
			}
			eye.Items = append(eye.Items, EyeItem{Eye: eye.Key, Kind: "COMPLAINTS", Severity: WatchHigh,
				Title:   fmt.Sprintf("%d شكاوى على %s بآخر ٣٠ يوم", r.Complaints30, r.Name),
				Detail:  fmt.Sprintf("مجموع شكاواه بآخر ٦٠ يوم %d%s.", r.Complaints, avg),
				Advice:  "المراقب يشوف إذا الشكاوى من نفس السبب، ويحچي ويا الموظف قبل ما تكبر.",
				OwnerID: r.EmployeeID, OwnerName: r.Name})
		}
		scores = append(scores, cs)
	}
	return finishEye(eye, "ماكو شي يقلق برأي الزبائن."), scores, team
}

var sevRank = map[string]int{WatchHigh: 0, WatchMedium: 1, WatchLow: 2}

func finishEye(eye EyeGroup, calm string) EyeGroup {
	sort.SliceStable(eye.Items, func(i, j int) bool { return sevRank[eye.Items[i].Severity] < sevRank[eye.Items[j].Severity] })
	for _, it := range eye.Items {
		if it.Severity == WatchHigh {
			eye.High++
		}
	}
	switch {
	case len(eye.Items) == 0:
		eye.Summary = calm
	case eye.High > 0:
		eye.Summary = fmt.Sprintf("%d شغلة تحتاج انتباه، منها %d مهمة.", len(eye.Items), eye.High)
	default:
		eye.Summary = fmt.Sprintf("%d شغلة تحتاج متابعة، ماكو شي خطير.", len(eye.Items))
	}
	return eye
}

func (s *MatrixWatchService) itEye() EyeGroup {
	eye := EyeGroup{Key: "IT", Title: "الآيتي وصيانة الأجهزة", Items: []EyeItem{}}
	str := func(p *string) string {
		if p == nil {
			return ""
		}
		return *p
	}
	if rows, err := s.repo.ItRepeatRepairs(); err == nil {
		for _, r := range rows {
			cost := ""
			if r.Cost != nil && *r.Cost > 0 {
				cost = fmt.Sprintf("، وكلّفت %s د.ع", num(*r.Cost))
			}
			eye.Items = append(eye.Items, EyeItem{Eye: eye.Key, Kind: "IT_REPEAT", Severity: WatchMedium,
				Title:   fmt.Sprintf("جهاز «%s» انصلح %d مرات بآخر ٣ أشهر", r.Name, r.Count),
				Detail:  "التصليح يتكرر" + cost + ".",
				Advice:  "قارن كلفة التصليحات ويا سعر جهاز جديد، يمكن الاستبدال أوفر.",
				OwnerID: str(r.HolderID), OwnerName: str(r.Holder), At: r.At})
		}
	}
	if rows, err := s.repo.ItInUseNoHolder(); err == nil {
		for _, r := range rows {
			eye.Items = append(eye.Items, EyeItem{Eye: eye.Key, Kind: "IT_NO_HOLDER", Severity: WatchLow,
				Title:  fmt.Sprintf("جهاز «%s» شغّال بلا مسؤول", r.Name),
				Detail: "الجهاز مأشّر مستعمل، بس ما مربوط بأي موظف.",
				Advice: "الآيتي يحدد منو مستلمه، حتى إذا صار بيه شي ينعرف منو يُسأل."})
		}
	}
	if rows, err := s.repo.RepairTicketsLate(); err == nil {
		for _, r := range rows {
			eye.Items = append(eye.Items, EyeItem{Eye: eye.Key, Kind: "REPAIR_LATE", Severity: WatchMedium,
				Title:   fmt.Sprintf("جهاز زبون (%s) عدنا من %d يوم وما انسلّم", str(r.Device), int(time.Since(r.Received).Hours()/24)),
				Detail:  "انستلم للتصليح وبعده ما انسجّل تسليمه.",
				Advice:  "يتواصل المسؤول ويا الزبون ويبلّغه بالموعد، أو يسجّل التسليم إذا انسلّم.",
				OwnerID: str(r.EmpID), OwnerName: str(r.Employee), At: ptr(r.Received)})
		}
	}
	if rows, err := s.repo.RepairRepeats(); err == nil {
		for _, r := range rows {
			eye.Items = append(eye.Items, EyeItem{Eye: eye.Key, Kind: "REPAIR_REPEAT", Severity: WatchMedium,
				Title:  fmt.Sprintf("جهاز زبون (%s) رجع للتصليح %d مرات", str(r.Device), r.Count),
				Detail: "نفس الرقم التسلسلي: " + str(r.Serial) + ".",
				Advice: "التصليح الأول ما حل المشكلة. يتراجع شنو انسوّى بيه أول مرة."})
		}
	}
	return finishEye(eye, "الآيتي وصيانة الأجهزة ماشية مضبوطة.")
}

func (s *MatrixWatchService) pendingEye() EyeGroup {
	eye := EyeGroup{Key: "PENDING", Title: "طلبات تنتظر قرار", Items: []EyeItem{}}
	rows, err := s.repo.PendingRequests()
	if err == nil {
		for _, r := range rows {
			days := int(time.Since(r.Since).Hours() / 24)
			sev := WatchLow
			if days >= 7 {
				sev = WatchMedium
			}
			who := ""
			if r.Who != nil {
				who = " — " + *r.Who
			}
			it := EyeItem{Eye: eye.Key, Kind: r.Kind, Severity: sev,
				Title:  fmt.Sprintf("%s%s: صارله %d يوم بلا قرار", r.Title, who, days),
				Detail: "القرار عند: " + r.Decider + ".",
				Advice: "ماتركس يقترح يتحسم هالطلب (موافقة أو رفض ويا السبب)، لأن الانتظار الطويل يضايق صاحبه.",
				At:     ptr(r.Since)}
			if r.WhoID != nil {
				it.OwnerID = *r.WhoID
			}
			if r.Who != nil {
				it.OwnerName = *r.Who
			}
			eye.Items = append(eye.Items, it)
		}
	}
	return finishEye(eye, "ماكو طلبات متأخرة.")
}

func (s *MatrixWatchService) Report() *EyesReport {
	cust, scores, team := s.customerEye()
	rep := &EyesReport{Eyes: []EyeGroup{s.stockEye(), s.vehicleEye(), cust, s.itEye(), s.pendingEye()}, Customers: scores, TeamHappy: team, Insights: []string{}}
	high := 0
	for _, e := range rep.Eyes {
		high += e.High
	}
	if high == 0 {
		rep.Insights = append(rep.Insights, "ماكو شي خطير بالمخزن والسيارات ورأي الزبائن والآيتي والطلبات.")
	} else {
		rep.Insights = append(rep.Insights, fmt.Sprintf("أكو %d شغلة مهمة تحتاج قرار.", high))
	}
	if team != nil {
		rep.Insights = append(rep.Insights, fmt.Sprintf("رضا الزبائن العام بآخر ٦٠ يوم %d٪.", *team))
	}
	return rep
}

// Remind تذكير يومي: الأداة الي ما رجعت لصاحبها، وملخّص البنود المهمة للمراقب.
func (s *MatrixWatchService) Remind(today string, dayStart time.Time) {
	if s.auto == nil || !s.auto.Enabled() {
		return
	}
	if rows, err := s.repo.ToolsNotReturned(); err == nil {
		for _, r := range rows {
			emp := r.EmployeeID
			msg := fmt.Sprintf("🤖 ماتركس — «%s» ويّاك من %d يوم وبعد ما رجعتها. رجّعها للمخزن، وإذا ضايعة بلّغ مسؤول المخزن.",
				r.Tool, int(time.Since(r.ApprovedAt).Hours()/24))
			s.auto.Act(model.AiAction{Kind: model.AiActionToolReturn, EntityType: "TOOL_REQUEST", EntityID: r.ID, Period: today,
				TargetEmployeeID: &emp, TargetLabel: r.Employee, Summary: fmt.Sprintf("ذكّر %s يرجّع «%s»", r.Employee, r.Tool),
				Details: why(map[string]any{"tool": r.Tool, "approvedAt": r.ApprovedAt})}, dayStart,
				func() error { return s.notif.Create(emp, "AI_AUTOPILOT", msg) })
		}
	}
	rep := s.Report()
	high := 0
	for _, e := range rep.Eyes {
		high += e.High
	}
	if high > 0 {
		msg := fmt.Sprintf("🤖 ماتركس — عيون الرقابة لگت %d شغلة مهمة. التفاصيل ويا اقتراحاتي بمكتب المراقب ← عيون الرقابة.", high)
		s.auto.Act(model.AiAction{Kind: model.AiActionWatchDigest, EntityType: "WATCH", EntityID: "watch", Period: today,
			TargetLabel: "المراقب", Summary: fmt.Sprintf("بلّغ المراقب بـ%d شغلة مهمة بعيون الرقابة", high),
			Details: why(map[string]any{"high": high})}, dayStart,
			func() error {
				return s.notif.CreateForRolesOrPermission([]string{"MONITOR"}, "monitoring", "AI_AUTOPILOT", msg)
			})
	}
}

// BaghdadLoc وَBaghdadMidnight للمؤقّتات بـmain.
func BaghdadLoc() *time.Location           { return debriefLoc }
func BaghdadMidnight(day string) time.Time { return baghdadMidnight(day) }
