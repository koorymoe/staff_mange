package service

import (
	"fmt"
	"math"
	"sort"
	"time"

	"staffmange-api/internal/repository"
)

// ═══ ماتركس — الكادر الأنسب، منحنى الموظف الجديد، المخزون، الاستبدال ═══
//
// كلها اقتراحات تنعرض للإنسان — ماكو غرامة ولا نقاط ولا تكليف تلقائي.
// لو البيانات ما تكفي نگول «ماكو بيانات كافية» بدل ما نخترع رقم.

type MatrixInsightsService struct {
	aiRepo *repository.AiRepository
}

func NewMatrixInsightsService(aiRepo *repository.AiRepository) *MatrixInsightsService {
	return &MatrixInsightsService{aiRepo: aiRepo}
}

const insufficientNote = "ماكو بيانات كافية"

// ───────────── ٣) توصية الكادر الأنسب ─────────────
//
// النقاط (قاعدة ثابتة، مو تعلّم):
//   - مهارة الخدمة: +٣٠ (بس لو الخدمة إلها مهارات معرّفة أصلاً)
//   - خبرة: +١.٥ لكل حجز منجز من نفس الخدمة بآخر سنة (سقف ٢٠ حجز = ٣٠)
//   - مشاكل: −(نسبة الحجوزات الي عليها توقف أو شكوى × ٤٠) — بس لو عنده ٥ منجز فأكثر
//   - عدالة الحمل: −١٠ لكل حجز ثاني عنده بنفس اليوم
// التعادل: الأقل حمل، بعدين بالاسم. نرجّع أعلى ٣ ليدرية وأعلى ٣ فنيين.

const (
	crewSkillPoints    = 30.0
	crewPerDonePoints  = 1.5
	crewDoneCap        = 20
	crewProblemWeight  = 40.0
	crewProblemMinDone = 5
	crewLoadPenalty    = 10.0
	crewTopN           = 3
)

type CrewSuggestion struct {
	EmployeeID     string   `json:"employeeId"`
	Name           string   `json:"name"`
	IsLeader       bool     `json:"isLeader"`
	Score          float64  `json:"score"`
	HasSkill       bool     `json:"hasSkill"`
	DoneCount      int      `json:"doneCount"`
	ProblemRatePct *int     `json:"problemRatePct,omitempty"`
	DayLoad        int      `json:"dayLoad"`
	Reasons        []string `json:"reasons"`
}

type CrewRecommendation struct {
	BookingID    string           `json:"bookingId"`
	ServiceName  string           `json:"serviceName"`
	Day          string           `json:"day"`
	Insufficient bool             `json:"insufficient"`
	Note         string           `json:"note,omitempty"`
	Leaders      []CrewSuggestion `json:"leaders"`
	Technicians  []CrewSuggestion `json:"technicians"`
}

// scoreCrew نقاط موظف واحد مع أسبابها.
func scoreCrew(c repository.CrewCandidateRow) CrewSuggestion {
	s := CrewSuggestion{EmployeeID: c.EmployeeID, Name: c.Name, IsLeader: c.IsLeader, HasSkill: c.HasSkill,
		DoneCount: c.DoneCount, DayLoad: c.DayLoad, Reasons: []string{}}
	if c.ServiceHasSkills {
		if c.HasSkill {
			s.Score += crewSkillPoints
			s.Reasons = append(s.Reasons, "عنده مهارة الخدمة")
		} else {
			s.Reasons = append(s.Reasons, "⚠️ ما عنده مهارة الخدمة")
		}
	}
	done := c.DoneCount
	if done > crewDoneCap {
		done = crewDoneCap
	}
	s.Score += float64(done) * crewPerDonePoints
	if c.DoneCount > 0 {
		s.Reasons = append(s.Reasons, fmt.Sprintf("نفّذ %d حجز من هالخدمة بآخر سنة", c.DoneCount))
	}
	if c.DoneCount >= crewProblemMinDone {
		rate := float64(c.ProblemCount) / float64(c.DoneCount)
		s.Score -= rate * crewProblemWeight
		pct := int(math.Round(rate * 100))
		s.ProblemRatePct = &pct
		s.Reasons = append(s.Reasons, fmt.Sprintf("توقف/شكوى %d%% (%d من %d)", pct, c.ProblemCount, c.DoneCount))
	}
	s.Score -= float64(c.DayLoad) * crewLoadPenalty
	if c.DayLoad > 0 {
		s.Reasons = append(s.Reasons, fmt.Sprintf("عنده %d حجز ثاني بنفس اليوم", c.DayLoad))
	} else {
		s.Reasons = append(s.Reasons, "فاضي بنفس اليوم")
	}
	s.Score = math.Round(s.Score*10) / 10
	return s
}

// rankCrew يرتّب ويرجّع أعلى n. informative=false لو ولا واحد عنده
// مهارة أو خبرة بالخدمة (يعني الترتيب يصير على الحمل بس — ما نقترح).
func rankCrew(rows []repository.CrewCandidateRow, n int) (out []CrewSuggestion, informative bool) {
	all := make([]CrewSuggestion, 0, len(rows))
	for _, r := range rows {
		if (r.ServiceHasSkills && r.HasSkill) || r.DoneCount > 0 {
			informative = true
		}
		all = append(all, scoreCrew(r))
	}
	sort.SliceStable(all, func(i, j int) bool {
		if all[i].Score != all[j].Score {
			return all[i].Score > all[j].Score
		}
		if all[i].DayLoad != all[j].DayLoad {
			return all[i].DayLoad < all[j].DayLoad
		}
		return all[i].Name < all[j].Name
	})
	if len(all) > n {
		all = all[:n]
	}
	return all, informative
}

func (s *MatrixInsightsService) CrewRecommendation(bookingID string) (*CrewRecommendation, error) {
	info, err := s.aiRepo.CrewBookingInfo(bookingID)
	if err != nil {
		return nil, err
	}
	out := &CrewRecommendation{BookingID: bookingID, ServiceName: info.ServiceName, Day: info.Day,
		Leaders: []CrewSuggestion{}, Technicians: []CrewSuggestion{}}
	if info.ServiceID == "" {
		out.Insufficient, out.Note = true, insufficientNote+" — الحجز بلا خدمة محددة"
		return out, nil
	}
	rows, err := s.aiRepo.CrewCandidates(bookingID, info.ServiceID, info.Day)
	if err != nil {
		return nil, err
	}
	var leaders, techs []repository.CrewCandidateRow
	for _, r := range rows {
		if r.IsLeader {
			leaders = append(leaders, r)
		} else {
			techs = append(techs, r)
		}
	}
	l, okL := rankCrew(leaders, crewTopN)
	t, okT := rankCrew(techs, crewTopN)
	if okL {
		out.Leaders = l
	}
	if okT {
		out.Technicians = t
	}
	if !okL && !okT {
		out.Insufficient, out.Note = true, insufficientNote+" — ماكو أحد بالدوام عنده مهارة أو خبرة مسجّلة بهالخدمة"
	}
	return out, nil
}

// ───────────── ٧) منحنى تعلّم الموظف الجديد ─────────────
//
// أسابيع من يوم بدايته (hireDate وإلا إنشاء الحساب)، ٧ أيام لكل أسبوع.
// نقاط الأسبوع = منجز − ٢×توقف − ٣×شكوى (الورق يُعرض بس، ما يدخل النقاط
// لأنه يخص الليدرية وحدهم).
// «تحسّن» الأسبوع = نقاطه أعلى من الأسبوع الي قبله، أو ثابت على أعلى
// قمّة وصلها (>٠) — الثابت على القمة مو تراجع.
// «يحتاج متابعة» = آخر ٣ أسابيع **مكتملة** كلها بلا تحسّن. أقل من ٤
// أسابيع مكتملة = «ماكو بيانات كافية بعد».

const (
	curveMaxWeeks       = 13
	curveStallWeeks     = 3
	CurveStatusOK       = "OK"
	CurveStatusFollowUp = "NEEDS_FOLLOWUP"
	CurveStatusTooEarly = "INSUFFICIENT"
)

type CurveWeek struct {
	Week       int  `json:"week"` // ١ = أول أسبوع
	Complete   bool `json:"complete"`
	Completed  int  `json:"completed"`
	Stops      int  `json:"stops"`
	Complaints int  `json:"complaints"`
	PaperDue   int  `json:"paperDue"`
	PaperOK    int  `json:"paperOk"`
	PaperPct   *int `json:"paperPct,omitempty"`
	Score      int  `json:"score"`
}

type NewEmployeeCurve struct {
	EmployeeID string      `json:"employeeId"`
	Name       string      `json:"name"`
	IsLeader   bool        `json:"isLeader"`
	StartDate  string      `json:"startDate"`
	DaysIn     int         `json:"daysIn"`
	Weeks      []CurveWeek `json:"weeks"`
	Status     string      `json:"status"`
	StatusNote string      `json:"statusNote"`
}

func curveScore(w CurveWeek) int { return w.Completed - 2*w.Stops - 3*w.Complaints }

// curveStatus القاعدة نفسها على نقاط الأسابيع المكتملة بالترتيب.
func curveStatus(scores []int) string {
	if len(scores) < curveStallWeeks+1 {
		return CurveStatusTooEarly
	}
	improved := func(i int) bool {
		peak := scores[0]
		for _, v := range scores[:i] {
			if v > peak {
				peak = v
			}
		}
		return scores[i] > scores[i-1] || (scores[i] >= peak && scores[i] > 0)
	}
	for i := len(scores) - curveStallWeeks; i < len(scores); i++ {
		if improved(i) {
			return CurveStatusOK
		}
	}
	return CurveStatusFollowUp
}

// bucketCurve يوزّع الأحداث على أسابيع من start لحد now.
func bucketCurve(start, now time.Time, events []repository.EmployeeEvent) []CurveWeek {
	weeks := []CurveWeek{}
	for i := 0; i < curveMaxWeeks; i++ {
		ws := start.AddDate(0, 0, 7*i)
		if !ws.Before(now) {
			break
		}
		weeks = append(weeks, CurveWeek{Week: i + 1, Complete: !ws.AddDate(0, 0, 7).After(now)})
	}
	for _, e := range events {
		if e.At.Before(start) {
			continue
		}
		i := int(e.At.Sub(start).Hours() / (24 * 7))
		if i < 0 || i >= len(weeks) {
			continue
		}
		w := &weeks[i]
		switch e.Kind {
		case "DONE":
			w.Completed++
			if e.PaperDue {
				w.PaperDue++
				if e.PaperOK {
					w.PaperOK++
				}
			}
		case "STOP":
			w.Stops++
		case "COMPLAINT":
			w.Complaints++
		}
	}
	for i := range weeks {
		if weeks[i].PaperDue > 0 {
			p := int(math.Round(100 * float64(weeks[i].PaperOK) / float64(weeks[i].PaperDue)))
			weeks[i].PaperPct = &p
		}
		weeks[i].Score = curveScore(weeks[i])
	}
	return weeks
}

func (s *MatrixInsightsService) NewEmployeeCurves() ([]NewEmployeeCurve, error) {
	emps, err := s.aiRepo.NewFieldEmployees()
	if err != nil {
		return nil, err
	}
	out := []NewEmployeeCurve{}
	if len(emps) == 0 {
		return out, nil
	}
	now := time.Now().In(debriefLoc)
	ids := make([]string, 0, len(emps))
	earliest := now
	starts := map[string]time.Time{}
	for _, e := range emps {
		d, err := time.ParseInLocation("2006-01-02", e.Start, debriefLoc)
		if err != nil {
			continue
		}
		starts[e.ID] = d
		ids = append(ids, e.ID)
		if d.Before(earliest) {
			earliest = d
		}
	}
	events, err := s.aiRepo.EmployeeEvents(ids, earliest, now.Add(time.Minute))
	if err != nil {
		return nil, err
	}
	byEmp := map[string][]repository.EmployeeEvent{}
	for _, ev := range events {
		byEmp[ev.EmployeeID] = append(byEmp[ev.EmployeeID], ev)
	}
	for _, e := range emps {
		st, ok := starts[e.ID]
		if !ok {
			continue
		}
		weeks := bucketCurve(st, now, byEmp[e.ID])
		scores := []int{}
		for _, w := range weeks {
			if w.Complete {
				scores = append(scores, w.Score)
			}
		}
		c := NewEmployeeCurve{EmployeeID: e.ID, Name: e.Name, IsLeader: e.IsLeader, StartDate: e.Start,
			DaysIn: int(now.Sub(st).Hours() / 24), Weeks: weeks, Status: curveStatus(scores)}
		switch c.Status {
		case CurveStatusOK:
			c.StatusNote = "ماشي طبيعي"
		case CurveStatusFollowUp:
			c.StatusNote = "يحتاج متابعة — ٣ أسابيع متتالية بلا تحسّن"
		default:
			c.StatusNote = insufficientNote + " بعد (أقل من ٤ أسابيع مكتملة)"
		}
		out = append(out, c)
	}
	return out, nil
}

// ───────────── ١٠) المخزون قبل ما يخلص ─────────────
//
// معدل الاستهلاك اليومي = مجموع الكمية بفواتير الليدرية (غير الملغاة)
// بآخر ٦٠ يوم ÷ ٦٠، للمواد المستعملة ١٠ مرات فأكثر.
// ⚠️ جدول Material ما بيه رصيد/كمية حالية، وماكو جدول حركة مخزون
// للمواد (StockIntake للأدوات، مو المواد) — فما نكدر نحسب «يخلص
// خلال كم يوم». نگولها صريحة ونعرض الأكثر استهلاكاً بمعدلاتهم.

const (
	stockWindowDays = 60
	stockMinUsages  = 10
	stockHorizon    = 14
)

type StockItem struct {
	Key           string  `json:"key"`
	Name          string  `json:"name"`
	Usages        int     `json:"usages"`
	TotalQuantity float64 `json:"totalQuantity"`
	AvgDaily      float64 `json:"avgDaily"`
	Need14Days    float64 `json:"need14Days"`
}

type StockForecast struct {
	StockTracked bool        `json:"stockTracked"`
	Note         string      `json:"note"`
	WindowDays   int         `json:"windowDays"`
	MinUsages    int         `json:"minUsages"`
	HorizonDays  int         `json:"horizonDays"`
	Items        []StockItem `json:"items"`
}

func avgDaily(total float64, days int) float64 {
	if days <= 0 {
		return 0
	}
	return math.Round(total/float64(days)*100) / 100
}

func (s *MatrixInsightsService) StockForecast() (*StockForecast, error) {
	rows, err := s.aiRepo.MaterialConsumptionSince(time.Now().AddDate(0, 0, -stockWindowDays), stockMinUsages)
	if err != nil {
		return nil, err
	}
	out := &StockForecast{
		StockTracked: false,
		Note:         "ماكو رصيد مخزون مسجّل للمواد بالنظام، فما نكدر نحسب «يخلص خلال كم يوم». هذي المواد الأكثر استهلاكاً ومعدلها — قارنها بالرف.",
		WindowDays:   stockWindowDays, MinUsages: stockMinUsages, HorizonDays: stockHorizon, Items: []StockItem{},
	}
	for _, r := range rows {
		a := avgDaily(r.Quantity, stockWindowDays)
		out.Items = append(out.Items, StockItem{Key: r.Key, Name: r.Name, Usages: r.Usages, TotalQuantity: r.Quantity,
			AvgDaily: a, Need14Days: math.Ceil(a * stockHorizon)})
	}
	if len(out.Items) == 0 {
		out.Note = insufficientNote + fmt.Sprintf(" — ولا مادة انصرفت %d مرات بآخر %d يوم، وماكو رصيد مخزون للمواد أصلاً.", stockMinUsages, stockWindowDays)
	}
	return out, nil
}

// ───────────── ١١) اقتراح استبدال ─────────────
//
// IT: جهاز غير متقاعد عليه ٣ سجلات تصليح/صيانة فأكثر بآخر ١٨٠ يوم.
// مركبة: (صيانة + حوادث بآخر ١٢ شهر) ≥ ٢× وسيط الأسطول — بس لو الأسطول
// ٣ مركبات فأكثر والوسيط > ٠ — أو ٣ حوادث فأكثر بآخر ١٨٠ يوم.

const (
	itRepairDays       = 180
	itRepairMin        = 3
	vehicleCostFactor  = 2.0
	vehicleMinFleet    = 3
	vehicleIncidentMin = 3
)

type ItReplacement struct {
	repository.ItRepairHeavy
	Reason string `json:"reason"`
}

type VehicleReplacement struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	PlateNumber   string   `json:"plateNumber"`
	Cost12m       float64  `json:"cost12m"`
	Incidents180d int      `json:"incidents180d"`
	Reasons       []string `json:"reasons"`
}

type ReplacementSuggestions struct {
	ItIncluded       bool                 `json:"itIncluded"`
	VehiclesIncluded bool                 `json:"vehiclesIncluded"`
	It               []ItReplacement      `json:"it"`
	Vehicles         []VehicleReplacement `json:"vehicles"`
	FleetMedian      *float64             `json:"fleetMedian,omitempty"`
	FleetNote        string               `json:"fleetNote,omitempty"`
}

func median(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	s := append([]float64(nil), xs...)
	sort.Float64s(s)
	m := len(s) / 2
	if len(s)%2 == 1 {
		return s[m]
	}
	return (s[m-1] + s[m]) / 2
}

// flagVehicles القاعدة على كل الأسطول. medianOK=false لو الأسطول صغير
// أو الوسيط صفر (قاعدة الكلفة تنطفي، وقاعدة الحوادث تبقى).
func flagVehicles(rows []repository.VehicleCostRow) (out []VehicleReplacement, med float64, medianOK bool) {
	costs := make([]float64, len(rows))
	for i, r := range rows {
		costs[i] = r.MaintCost12m + r.IncidentCost
	}
	med = median(costs)
	medianOK = len(rows) >= vehicleMinFleet && med > 0
	out = []VehicleReplacement{}
	for i, r := range rows {
		reasons := []string{}
		if medianOK && costs[i] >= vehicleCostFactor*med {
			reasons = append(reasons, fmt.Sprintf("كلفة صيانة وحوادث سنة %s د.ع = %.1f× وسيط الأسطول", fmtIQD(costs[i]), costs[i]/med))
		}
		if r.Incidents180d >= vehicleIncidentMin {
			reasons = append(reasons, fmt.Sprintf("%d حوادث بآخر ١٨٠ يوم", r.Incidents180d))
		}
		if len(reasons) > 0 {
			out = append(out, VehicleReplacement{ID: r.ID, Name: r.Name, PlateNumber: r.PlateNumber,
				Cost12m: costs[i], Incidents180d: r.Incidents180d, Reasons: reasons})
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Cost12m > out[j].Cost12m })
	return out, med, medianOK
}

func fmtIQD(v float64) string {
	n := int64(math.Round(v))
	s := fmt.Sprintf("%d", n)
	out := []byte{}
	for i := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, s[i])
	}
	return string(out)
}

func (s *MatrixInsightsService) Replacements(includeIT, includeVehicles bool) (*ReplacementSuggestions, error) {
	out := &ReplacementSuggestions{ItIncluded: includeIT, VehiclesIncluded: includeVehicles,
		It: []ItReplacement{}, Vehicles: []VehicleReplacement{}}
	if includeIT {
		rows, err := s.aiRepo.ItRepairCounts(itRepairDays, itRepairMin)
		if err != nil {
			return nil, err
		}
		for _, r := range rows {
			out.It = append(out.It, ItReplacement{ItRepairHeavy: r,
				Reason: fmt.Sprintf("%d تصليح/صيانة بآخر ١٨٠ يوم (كلفة %s د.ع)", r.RepairCount, fmtIQD(r.RepairCost))})
		}
	}
	if includeVehicles {
		rows, err := s.aiRepo.VehicleCosts()
		if err != nil {
			return nil, err
		}
		v, med, ok := flagVehicles(rows)
		out.Vehicles = v
		if ok {
			out.FleetMedian = &med
		} else {
			out.FleetNote = insufficientNote + " لمقارنة الكلفة بالأسطول (أقل من ٣ مركبات أو ماكو كلف مسجّلة) — بس قاعدة الحوادث شغّالة."
		}
	}
	return out, nil
}
