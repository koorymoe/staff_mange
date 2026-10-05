package service

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"staffmange-api/internal/repository"
)

// ═══ ماتركس ٢٠٥٠ — «سلسلة الحجز» ═══
//
// طلب (ع) 10-05: «الخطوة الاولى انو يراقب وينطيني تقارير وايفادات ومعلومات،
// ولازم المعلومات ماتكون سهله ابد». كل حجز يتقسّم ١٤ محطة، ولكل محطة: منو
// المسؤول، متى بدت، متى خلصت، شكد أخذت، وهل صارت صح. ومن المحطات تنبني
// تقارير كل دور (الحجوزات، المنسقين، الليدرية، الفنيين، المحاسبة، الجودة،
// المراقبين).
//
// ⚠️ ماكو غرامات ولا نقاط — قراءة بس. وكل رقم من الجداول.
// ⚠️ «متأخرة» مو رقم ثابت: وسيط الشركة لنفس المحطة (ولنفس الخدمة إذا عندها
// عينات كافية) بآخر ٩٠ يوم × ١٫٥، وبحد أدنى معقول. والحد يطلع ويا كل حكم.

const (
	ChainOK      = "OK"      // ✅ صارت بوقتها
	ChainLate    = "LATE"    // 🟠 صارت بس متأخرة
	ChainMissed  = "MISSED"  // 🔴 ما صارت ووقتها فات
	ChainWaiting = "WAITING" // ⏳ بعدها، ووقتها ما فات
	ChainIssue   = "ISSUE"   // ⚠️ صارت بس بيها غلط
	ChainNA      = "NA"      // ➖ ما تنطبق على هذا الحجز
)

// أدوار عين ماتركس.
const (
	ChainRoleBooking     = "BOOKING"
	ChainRoleCoordinator = "COORDINATOR"
	ChainRoleLeader      = "LEADER"
	ChainRoleTech        = "TECH"
	ChainRoleAccountant  = "ACCOUNTANT"
	ChainRoleQuality     = "QUALITY"
	ChainRoleMonitor     = "MONITOR"
)

var ChainRoleTitles = map[string]string{
	ChainRoleBooking:     "موظفي الحجوزات",
	ChainRoleCoordinator: "المنسقين",
	ChainRoleLeader:      "الليدرية",
	ChainRoleTech:        "الفنيين",
	ChainRoleAccountant:  "المحاسبة",
	ChainRoleQuality:     "الجودة",
	ChainRoleMonitor:     "المراقبين",
}

var ChainRoleOrder = []string{ChainRoleBooking, ChainRoleCoordinator, ChainRoleLeader, ChainRoleTech,
	ChainRoleAccountant, ChainRoleQuality, ChainRoleMonitor}

type chainStationDef struct {
	Key   string
	Title string
	Floor int // أقل حد «متأخرة» بالدقايق — حتى الوسيط الصغير جداً ما يظلم
}

var chainStations = []chainStationDef{
	{"CREATE", "تسجيل الحجز", 0},
	{"CONTACT", "التواصل ويا الزبون", 30},
	{"CONFIRM", "التثبيت والموعد", 60},
	{"CREW", "تحديد الكادر", 60},
	{"RECEIVE", "الاستلام وتجهيز المواد", 60},
	{"ROUTE", "الانطلاق والطريق", 30},
	{"WORK", "العمل", 60},
	{"COMPLETE", "الإنجاز (تام/جزئي)", 24 * 60},
	{"PAPER", "الورق: الفاتورة والتقرير", 3 * 60},
	{"RATING", "تقييم الليدر لفنيّيه", 24 * 60},
	{"INVENTORY", "جرد العدّة", 0},
	{"ACCOUNT", "المحاسب: التدقيق والاعتماد", 24 * 60},
	{"QUALITY", "اتصال الجودة بالزبون", 48 * 60},
	{"MONITOR", "حكم المراقب", 4 * 60},
}

func chainStationTitle(key string) string {
	for _, d := range chainStations {
		if d.Key == key {
			return d.Title
		}
	}
	return key
}

type ChainOwner struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status,omitempty"` // إذا اختلف عن حكم المحطة (الجرد لكل فني)
}

type ChainStation struct {
	No        int          `json:"no"`
	Key       string       `json:"key"`
	Title     string       `json:"title"`
	Role      string       `json:"role"`
	RoleTitle string       `json:"roleTitle"`
	Owners    []ChainOwner `json:"owners"`
	StartAt   *time.Time   `json:"startAt"`
	EndAt     *time.Time   `json:"endAt"`
	Minutes   *int         `json:"minutes"`
	Threshold *int         `json:"threshold"`
	Status    string       `json:"status"`
	Verdict   string       `json:"verdict"` // حكم ماتركس بجملة
	Issues    []string     `json:"issues"`
	Facts     []string     `json:"facts"` // تفاصيل إضافية (مدة الطريق، الموعد…)
}

type BookingChain struct {
	BookingID string                     `json:"bookingId"`
	Code      string                     `json:"code"`
	Service   string                     `json:"service"`
	Solo      bool                       `json:"solo"`
	Legacy    bool                       `json:"legacy"`
	Stations  []ChainStation             `json:"stations"`
	Summary   string                     `json:"summary"`
	Score     ChainScore                 `json:"score"`
	Ratings   []repository.CrewRatingRow `json:"ratings"`
}

type ChainScore struct {
	OK      int `json:"ok"`
	Late    int `json:"late"`
	Missed  int `json:"missed"`
	Issue   int `json:"issue"`
	Waiting int `json:"waiting"`
}

type MatrixChainService struct {
	repo  *repository.MatrixChainRepository
	names func(id string) string

	mu       sync.Mutex
	thAt     time.Time
	thGlobal map[string]int
	thSvc    map[string]map[string]int
	thMed    map[string]int
	thN      map[string]int
	thSvcN   int
}

func NewMatrixChainService(repo *repository.MatrixChainRepository, names func(string) string) *MatrixChainService {
	return &MatrixChainService{repo: repo, names: names}
}

func (s *MatrixChainService) Repo() *repository.MatrixChainRepository { return s.repo }

// ═══ الحدود: وسيط الشركة ═══

func chainMedian(v []int) int {
	if len(v) == 0 {
		return 0
	}
	c := append([]int(nil), v...)
	sort.Ints(c)
	return c[len(c)/2]
}

// rawMinutes قياس المحطة بدون حكم — للوسيط.
func rawMinutes(f *repository.ChainFacts) map[string]*int {
	m := map[string]*int{}
	m["CONTACT"] = minsBetween(&f.CreatedAt, f.ContactedAt)
	m["CONFIRM"] = minsBetween(firstOf(f.ContactedAt, &f.CreatedAt), f.ConfirmedAt)
	m["CREW"] = minsBetween(f.ConfirmedAt, f.FirstAssignAt)
	m["RECEIVE"] = minsBetween(firstOf(f.MissionAt, f.FirstAssignAt), f.MaterialsAt)
	m["ROUTE"] = minsBetween(f.ScheduledAt, firstOf(f.DepartedAt, f.StartedAt))
	m["WORK"] = minsBetween(firstOf(f.WorkStartAt, f.StartedAt, f.ArrivedAt), f.CompletedAt)
	m["COMPLETE"] = minsBetween(f.ScheduledAt, f.CompletedAt)
	m["PAPER"] = minsBetween(f.CompletedAt, latestOf(f.InvoiceAt, f.ReportAt))
	m["RATING"] = minsBetween(f.CompletedAt, f.RatedAt)
	m["ACCOUNT"] = minsBetween(firstOf(f.InvoiceAt, f.CompletedAt), f.ApprovedAt)
	m["QUALITY"] = minsBetween(f.CompletedAt, f.QualityAt)
	if f.MonitorMinutes != nil {
		v := int(*f.MonitorMinutes)
		m["MONITOR"] = &v
	}
	return m
}

func (s *MatrixChainService) thresholds() (map[string]int, map[string]map[string]int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.thGlobal != nil && time.Since(s.thAt) < time.Hour {
		return s.thGlobal, s.thSvc
	}
	facts, err := s.repo.Since(time.Now().AddDate(0, 0, -90))
	g := map[string][]int{}
	sv := map[string]map[string][]int{}
	if err == nil {
		for i := range facts {
			f := &facts[i]
			for k, v := range rawMinutes(f) {
				if v == nil {
					continue
				}
				g[k] = append(g[k], *v)
				if f.ServiceID != nil {
					if sv[*f.ServiceID] == nil {
						sv[*f.ServiceID] = map[string][]int{}
					}
					sv[*f.ServiceID][k] = append(sv[*f.ServiceID][k], *v)
				}
			}
		}
	}
	floor := map[string]int{}
	for _, d := range chainStations {
		floor[d.Key] = d.Floor
	}
	mk := func(k string, vals []int) int {
		th := int(math.Round(float64(chainMedian(vals)) * 1.5))
		if th < floor[k] {
			th = floor[k]
		}
		return th
	}
	s.thGlobal, s.thMed, s.thN = map[string]int{}, map[string]int{}, map[string]int{}
	for _, d := range chainStations {
		s.thGlobal[d.Key] = mk(d.Key, g[d.Key])
		s.thMed[d.Key] = chainMedian(g[d.Key])
		s.thN[d.Key] = len(g[d.Key])
	}
	// الجودة: قرار (ع) — «بعد يوم من الإنجاز». حد ثابت يومين، مو وسيط.
	s.thGlobal["QUALITY"] = 48 * 60
	s.thSvc = map[string]map[string]int{}
	for svc, byKey := range sv {
		s.thSvc[svc] = map[string]int{}
		for k, vals := range byKey {
			// الخدمة تاخذ حدها الخاص بس إذا عندها ٨ عينات أو أكثر.
			if len(vals) >= 8 && k != "QUALITY" {
				s.thSvc[svc][k] = mk(k, vals)
			}
		}
	}
	s.thSvcN = 0
	for _, m := range s.thSvc {
		if len(m) > 0 {
			s.thSvcN++
		}
	}
	s.thAt = time.Now()
	return s.thGlobal, s.thSvc
}

func (s *MatrixChainService) threshold(f *repository.ChainFacts, key string) int {
	g, sv := s.thresholds()
	if f.ServiceID != nil {
		if v, ok := sv[*f.ServiceID][key]; ok {
			return v
		}
	}
	return g[key]
}

// ═══ أدوات الوقت ═══

func minsBetween(from, to *time.Time) *int {
	if from == nil || to == nil || from.IsZero() || to.IsZero() {
		return nil
	}
	d := int(to.Sub(*from).Minutes())
	if d < 0 {
		d = 0
	}
	return &d
}

func firstOf(ts ...*time.Time) *time.Time {
	for _, t := range ts {
		if t != nil && !t.IsZero() {
			return t
		}
	}
	return nil
}

func latestOf(ts ...*time.Time) *time.Time {
	var out *time.Time
	for _, t := range ts {
		if t == nil || t.IsZero() {
			return nil // لازم الاثنين
		}
		if out == nil || t.After(*out) {
			out = t
		}
	}
	return out
}

// FmtMinutes «٣ ساعات و١٠ دقايق» بشكل مختصر.
func FmtMinutes(m int) string {
	switch {
	case m <= 0:
		return "فوراً"
	case m < 60:
		return fmt.Sprintf("%d د", m)
	case m < 24*60:
		h, r := m/60, m%60
		if r == 0 {
			return fmt.Sprintf("%d س", h)
		}
		return fmt.Sprintf("%d س %d د", h, r)
	default:
		d, h := m/(24*60), (m%(24*60))/60
		if h == 0 {
			return fmt.Sprintf("%d يوم", d)
		}
		return fmt.Sprintf("%d يوم %d س", d, h)
	}
}

func str(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// ═══ بناء السلسلة ═══

func (s *MatrixChainService) Build(bookingID string) (*BookingChain, error) {
	f, err := s.repo.ForBooking(bookingID)
	if err != nil {
		return nil, fmt.Errorf("الحجز مو موجود")
	}
	ch := s.chainOf(f, time.Now())
	if r, err := s.repo.RatingsByBooking(bookingID); err == nil {
		ch.Ratings = r
	}
	return ch, nil
}

func (s *MatrixChainService) owner(id *string, name *string) []ChainOwner {
	if id == nil || *id == "" {
		return nil
	}
	n := str(name)
	if n == "" && s.names != nil {
		n = s.names(*id)
	}
	return []ChainOwner{{ID: *id, Name: n}}
}

func (s *MatrixChainService) chainOf(f *repository.ChainFacts, now time.Time) *BookingChain {
	ch := &BookingChain{BookingID: f.ID, Code: f.Code, Service: str(f.ServiceName), Solo: f.Solo,
		Legacy: strings.HasPrefix(strings.ToUpper(f.Code), "OLD")}
	cancelled := f.Status == "CANCELLED"
	done := f.CompletedAt != nil && (f.Status == "COMPLETED" || f.Status == "PARTIAL")
	survey := f.BookingType == "SURVEY"

	// المسؤول الميداني: الليدر، وبالخدمات الفردية (داش كام/جي بي اس) الفني الوحيد.
	fieldRole := ChainRoleLeader
	fieldOwner := s.owner(f.LeaderID, f.LeaderName)
	if f.Solo {
		fieldRole = ChainRoleTech
		if len(fieldOwner) == 0 && len(f.CrewIDs) > 0 {
			id := f.CrewIDs[0]
			fieldOwner = s.owner(&id, nil)
		}
	}

	// judge: حكم موحّد لمحطة «من A لـB».
	judge := func(st *ChainStation, start, end *time.Time, th int, deadlineFrom *time.Time) {
		st.StartAt, st.EndAt = start, end
		st.Threshold = &th
		if end != nil {
			st.Minutes = minsBetween(start, end)
			if st.Minutes != nil && th > 0 && *st.Minutes > th {
				st.Status = ChainLate
				st.Verdict = fmt.Sprintf("أخذت %s، والمعتاد بالشركة ما يتجاوز %s.", FmtMinutes(*st.Minutes), FmtMinutes(th))
			} else {
				st.Status = ChainOK
				if st.Minutes != nil {
					st.Verdict = fmt.Sprintf("صارت خلال %s (المعتاد ما يتجاوز %s).", FmtMinutes(*st.Minutes), FmtMinutes(th))
				} else {
					st.Verdict = "صارت."
				}
			}
			return
		}
		from := deadlineFrom
		if from == nil {
			from = start
		}
		if from == nil || cancelled {
			st.Status = ChainNA
			st.Verdict = "الحجز بعده ما وصل لهالمحطة."
			return
		}
		// محطة قبل الإنجاز والحجز خلص بدونها — انتخطّت، مو «بعدها تنتظر».
		if done && (st.Key == "CONTACT" || st.Key == "CONFIRM" || st.Key == "CREW" || st.Key == "RECEIVE" || st.Key == "WORK") {
			st.Status = ChainMissed
			st.Verdict = "الحجز خلص، بس هالخطوة ما انسجّلت بالنظام."
			return
		}
		el := int(now.Sub(*from).Minutes())
		if el > th {
			st.Status = ChainMissed
			st.Verdict = fmt.Sprintf("ما صارت لحد هسه. صارلها %s تنتظر، والمعتاد ما يتجاوز %s.", FmtMinutes(el), FmtMinutes(th))
		} else {
			st.Status = ChainWaiting
			st.Verdict = fmt.Sprintf("تنتظر. صارلها %s، والمعتاد ما يتجاوز %s.", FmtMinutes(maxInt(el, 0)), FmtMinutes(th))
		}
	}
	add := func(st ChainStation) {
		st.No = len(ch.Stations) + 1
		st.Title = chainStationTitle(st.Key)
		st.RoleTitle = ChainRoleTitles[st.Role]
		if st.Issues == nil {
			st.Issues = []string{}
		}
		if st.Facts == nil {
			st.Facts = []string{}
		}
		if st.Owners == nil {
			st.Owners = []ChainOwner{}
		}
		// غلط بمحطة «بوقتها» يقلبها ⚠️ — الوقت زين بس الشغل بيه نقص.
		if st.Status == ChainOK && len(st.Issues) > 0 {
			st.Status = ChainIssue
		}
		ch.Stations = append(ch.Stations, st)
	}

	// ١. تسجيل الحجز
	{
		st := ChainStation{Key: "CREATE", Role: ChainRoleBooking, Owners: s.owner(f.CreatedByID, f.CreatedByName)}
		created := f.CreatedAt
		st.StartAt, st.EndAt = &created, &created
		st.Status = ChainOK
		st.Verdict = "انسجّل."
		if f.NameWords > 0 && f.NameWords < 3 {
			st.Issues = append(st.Issues, fmt.Sprintf("اسم الزبون %d كلمة بس، مو ثلاثي.", f.NameWords))
		}
		if !f.HasAddress && !f.HasLocation {
			st.Issues = append(st.Issues, "بلا عنوان ولا موقع.")
		}
		if f.ServiceID == nil && f.BookingType != "INTERNAL" {
			st.Issues = append(st.Issues, "بلا خدمة محددة.")
		}
		if f.Duplicate {
			st.Issues = append(st.Issues, "ماتركس لگاه مشكوك تكرار ويا حجز ثاني.")
		}
		if len(st.Owners) == 0 {
			st.Facts = append(st.Facts, "ما معروف منو سجّله (حجز قديم قبل تسجيل المُنشئ).")
		}
		add(st)
	}

	// ٢. التواصل
	{
		st := ChainStation{Key: "CONTACT", Role: ChainRoleCoordinator}
		created := f.CreatedAt
		if f.ContactedAt != nil {
			st.Owners = s.owner(f.ContactedByID, f.ContactedBy)
			judge(&st, &created, f.ContactedAt, s.threshold(f, "CONTACT"), nil)
		} else if f.ConfirmedAt != nil {
			// قرار (ع) 10-05: «الإداري الي يثبّت الحجز هو نفسه لازم يضغط تواصلت
			// ويا الزبون — إذا ما ضغط يعني هو الي ما ضاغط». فالمحطة عليه وما صارت.
			st.Owners = s.owner(f.ConfirmedByID, f.ConfirmedBy)
			st.StartAt, st.EndAt = &created, nil
			th := s.threshold(f, "CONTACT")
			st.Threshold = &th
			st.Status = ChainMissed
			st.Verdict = "ثبّت الحجز بدون ما يضغط «تواصلت ويا الزبون»، فهالخطوة محسوبة عليه."
		} else {
			judge(&st, &created, nil, s.threshold(f, "CONTACT"), nil)
		}
		if f.ContactTries > 1 {
			st.Facts = append(st.Facts, fmt.Sprintf("حاول %d مرات (الزبون ما رد).", f.ContactTries))
		}
		add(st)
	}

	// ٣. التثبيت
	{
		st := ChainStation{Key: "CONFIRM", Role: ChainRoleCoordinator, Owners: s.owner(f.ConfirmedByID, f.ConfirmedBy)}
		judge(&st, firstOf(f.ContactedAt, &f.CreatedAt), f.ConfirmedAt, s.threshold(f, "CONFIRM"), nil)
		if f.ScheduledAt != nil {
			st.Facts = append(st.Facts, "الموعد: "+f.ScheduledAt.In(debriefLoc).Format("2006-01-02 15:04"))
		} else if f.ConfirmedAt != nil {
			st.Issues = append(st.Issues, "انثبّت بلا موعد.")
		}
		if f.PostponeCount > 0 {
			st.Facts = append(st.Facts, fmt.Sprintf("انأجّل %d مرة.", f.PostponeCount))
		}
		if f.ScheduleMoves > 1 {
			st.Facts = append(st.Facts, fmt.Sprintf("الموعد انتغيّر %d مرات.", f.ScheduleMoves))
		}
		add(st)
	}

	// ٤. الكادر
	{
		st := ChainStation{Key: "CREW", Role: ChainRoleCoordinator, Owners: s.owner(f.ConfirmedByID, f.ConfirmedBy)}
		judge(&st, f.ConfirmedAt, f.FirstAssignAt, s.threshold(f, "CREW"), nil)
		if f.FirstAssignAt != nil {
			st.Facts = append(st.Facts, fmt.Sprintf("عدد الكادر: %d.", len(f.CrewIDs)))
			if !f.Solo && !survey && f.LeaderID == nil {
				st.Issues = append(st.Issues, "الكادر بلا ليدر.")
			}
			if f.ScheduledAt != nil {
				if f.FirstAssignAt.After(*f.ScheduledAt) {
					st.Issues = append(st.Issues, "الكادر انحدد بعد الموعد.")
				} else {
					st.Facts = append(st.Facts, "انحدد قبل الموعد بـ"+FmtMinutes(int(f.ScheduledAt.Sub(*f.FirstAssignAt).Minutes()))+".")
				}
			}
		} else if f.ScheduledAt != nil && now.After(*f.ScheduledAt) && !cancelled && !done {
			st.Status = ChainMissed
			st.Verdict = "وصل الموعد والحجز بلا كادر."
		}
		add(st)
	}

	// ٥. الاستلام والمواد
	{
		st := ChainStation{Key: "RECEIVE", Role: fieldRole, Owners: fieldOwner}
		start := firstOf(f.MissionAt, f.FirstAssignAt)
		if f.MaterialsAt == nil && done {
			st.Status = ChainNA
			st.Verdict = "الحجز خلص بدون ما ينسجّل وقت استلام المواد."
			st.Issues = append(st.Issues, "ما أشّر «المواد جاهزة».")
			st.StartAt = start
			add(st)
		} else {
			judge(&st, start, f.MaterialsAt, s.threshold(f, "RECEIVE"), nil)
			add(st)
		}
	}

	// ٦. الانطلاق والطريق
	{
		st := ChainStation{Key: "ROUTE", Role: fieldRole, Owners: fieldOwner}
		dep := firstOf(f.DepartedAt, f.StartedAt)
		if f.ScheduledAt == nil {
			st.Status = ChainNA
			st.Verdict = "الحجز بلا موعد، فما نگدر نقيس التأخير."
		} else {
			judge(&st, f.ScheduledAt, dep, s.threshold(f, "ROUTE"), nil)
			if dep != nil && st.Minutes != nil {
				if *st.Minutes == 0 {
					st.Verdict = "طلع بموعده أو قبله."
				} else if st.Status == ChainOK {
					st.Verdict = fmt.Sprintf("طلع بعد الموعد بـ%s، وهذا ضمن المعتاد (لحد %s).", FmtMinutes(*st.Minutes), FmtMinutes(*st.Threshold))
				} else {
					st.Verdict = fmt.Sprintf("طلع متأخر عن الموعد بـ%s، والمعتاد ما يتجاوز %s.", FmtMinutes(*st.Minutes), FmtMinutes(*st.Threshold))
				}
			}
		}
		if f.DepartedAt != nil && f.ArrivedAt != nil {
			if r := minsBetween(f.DepartedAt, f.ArrivedAt); r != nil {
				st.Facts = append(st.Facts, "مدة الطريق: "+FmtMinutes(*r)+".")
			}
		} else if f.DepartedAt == nil && f.StartedAt != nil {
			st.Issues = append(st.Issues, "ما سجّل «انطلقت» — القياس من بدء الشغل.")
		}
		add(st)
	}

	// ٧. العمل
	{
		st := ChainStation{Key: "WORK", Role: fieldRole, Owners: fieldOwner}
		start := firstOf(f.WorkStartAt, f.StartedAt, f.ArrivedAt)
		judge(&st, start, f.CompletedAt, s.threshold(f, "WORK"), start)
		if f.StoppedAt != nil {
			st.Facts = append(st.Facts, "توقف العمل "+f.StoppedAt.In(debriefLoc).Format("01-02 15:04")+".")
		}
		if start == nil && done {
			st.Issues = append(st.Issues, "ما سجّل بدء العمل.")
		}
		add(st)
	}

	// ٨. الإنجاز
	{
		st := ChainStation{Key: "COMPLETE", Role: fieldRole, Owners: fieldOwner}
		judge(&st, f.ScheduledAt, f.CompletedAt, s.threshold(f, "COMPLETE"), f.ScheduledAt)
		if f.Status == "PARTIAL" {
			st.Facts = append(st.Facts, fmt.Sprintf("إنجاز جزئي (%d يوم).", f.PartialCount))
		} else if f.PartialCount > 0 {
			st.Facts = append(st.Facts, fmt.Sprintf("خلص بعد %d إنجاز جزئي.", f.PartialCount))
		}
		if f.StoppedAt != nil && done && f.Status == "COMPLETED" && f.PartialCount == 0 {
			st.Facts = append(st.Facts, "توقف ثم انأشّر «تام» بلا جزئي.")
		}
		add(st)
	}

	// ٩. الورق
	{
		owners := fieldOwner
		if len(owners) == 0 {
			owners = s.owner(f.InvoiceByID, nil)
		}
		st := ChainStation{Key: "PAPER", Role: fieldRole, Owners: owners}
		if survey || ch.Legacy {
			st.Status = ChainNA
			st.Verdict = "معفي من الورق (كشف أو حجز قديم)."
		} else {
			judge(&st, f.CompletedAt, latestOf(f.InvoiceAt, f.ReportAt), s.threshold(f, "PAPER"), f.CompletedAt)
			if f.CompletedAt != nil {
				if f.InvoiceAt == nil {
					st.Issues = append(st.Issues, "ماكو فاتورة.")
				}
				if f.ReportAt == nil {
					st.Issues = append(st.Issues, "ماكو تقرير عمل.")
				}
				if st.Status != ChainOK && len(st.Issues) > 0 {
					st.Verdict += " " + strings.Join(st.Issues, " ")
					st.Issues = []string{}
				}
			}
		}
		add(st)
	}

	// ١٠. تقييم الفنيين
	{
		st := ChainStation{Key: "RATING", Role: ChainRoleLeader, Owners: s.owner(f.LeaderID, f.LeaderName)}
		techs := 0
		for _, t := range f.TechIDs {
			if f.LeaderID == nil || t != *f.LeaderID {
				techs++
			}
		}
		if f.Solo || f.LeaderID == nil || techs == 0 {
			st.Status = ChainNA
			st.Verdict = "ماكو ليدر وفنيين بهالحجز حتى ينقيّمون."
		} else {
			var end *time.Time
			if f.RatedCount >= techs {
				end = f.RatedAt
			}
			judge(&st, f.CompletedAt, end, s.threshold(f, "RATING"), f.CompletedAt)
			st.Facts = append(st.Facts, fmt.Sprintf("قيّم %d من %d.", f.RatedCount, techs))
		}
		add(st)
	}

	// ١١. جرد العدّة **بعد** الحجز — لكل فني طلع ويا ليدر.
	{
		st := ChainStation{Key: "INVENTORY", Role: ChainRoleTech}
		inv, short := map[string]bool{}, map[string]bool{}
		for _, id := range f.InventoryIDs {
			inv[id] = true
		}
		for _, id := range f.InventoryMiss {
			short[id] = true
		}
		if f.Solo || len(f.TechIDs) == 0 || f.LeaderID == nil || survey || ch.Legacy {
			st.Status = ChainNA
			st.Verdict = "ماكو فنيين طالعين ويا ليدر بهالحجز."
		} else {
			missing, lacking := []string{}, []string{}
			for _, id := range f.TechIDs {
				if id == *f.LeaderID {
					continue
				}
				o := s.owner(&id, nil)[0]
				switch {
				case inv[id] && short[id]:
					o.Status = ChainIssue
					lacking = append(lacking, o.Name)
				case inv[id]:
					o.Status = ChainOK
				case done:
					o.Status = ChainMissed
					missing = append(missing, o.Name)
				default:
					o.Status = ChainWaiting
				}
				st.Owners = append(st.Owners, o)
			}
			if len(lacking) > 0 {
				st.Issues = append(st.Issues, "لگوا نقص بعدّتهم بعد الشغل: "+strings.Join(lacking, "، ")+".")
			}
			switch {
			case len(missing) > 0:
				st.Status = ChainMissed
				st.Verdict = "ما جردوا عدّتهم بعد الحجز: " + strings.Join(missing, "، ") + "."
			case done:
				st.Status = ChainOK
				st.Verdict = "كل الفنيين جردوا عدّتهم بعد الحجز."
			default:
				st.Status = ChainWaiting
				st.Verdict = "ينتظر نهاية الحجز."
			}
		}
		add(st)
	}

	// ١٢. المحاسب
	{
		owners := s.owner(f.ApprovedByID, f.ApprovedBy)
		if len(owners) == 0 {
			owners = s.owner(f.AuditedByID, f.AuditedBy)
		}
		st := ChainStation{Key: "ACCOUNT", Role: ChainRoleAccountant, Owners: owners}
		if f.InvoiceAt == nil {
			st.Status = ChainNA
			st.Verdict = "بعد ماكو فاتورة حتى توصل للمحاسب."
		} else {
			judge(&st, f.InvoiceAt, f.ApprovedAt, s.threshold(f, "ACCOUNT"), f.InvoiceAt)
			if f.AuditedAt != nil {
				if a := minsBetween(f.InvoiceAt, f.AuditedAt); a != nil {
					st.Facts = append(st.Facts, "التدقيق بعد "+FmtMinutes(*a)+" من الفاتورة.")
				}
			}
			if f.InvoiceNet != nil && f.Collected != nil {
				diff := math.Abs(*f.InvoiceNet - *f.Collected)
				if diff >= 1000 {
					st.Issues = append(st.Issues, fmt.Sprintf("المبلغ المستلم %.0f والفاتورة %.0f، يعني أكو فرق %.0f د.ع.", *f.Collected, *f.InvoiceNet, diff))
				} else {
					st.Facts = append(st.Facts, "المستلم مطابق للفاتورة.")
				}
			}
			if f.ApprovedAt != nil && !f.AmountChecked {
				st.Facts = append(st.Facts, "المبلغ ما انأشّر «متحقق».")
			}
		}
		add(st)
	}

	// ١٣. الجودة — «بعد يوم من الإنجاز»
	{
		st := ChainStation{Key: "QUALITY", Role: ChainRoleQuality, Owners: s.owner(f.QualityByID, f.QualityBy)}
		if f.CompletedAt == nil {
			st.Status = ChainNA
			st.Verdict = "الحجز بعده ما خلص."
		} else {
			judge(&st, f.CompletedAt, f.QualityAt, s.threshold(f, "QUALITY"), f.CompletedAt)
			if f.QualityStatus != nil {
				st.Facts = append(st.Facts, "النتيجة: "+qualityStatusLabel(*f.QualityStatus)+".")
			}
			if f.QualityCreated == nil {
				st.Facts = append(st.Facts, "ما انفتحت متابعة جودة لهذا الحجز.")
			}
		}
		add(st)
	}

	// ١٤. المراقب
	{
		st := ChainStation{Key: "MONITOR", Role: ChainRoleMonitor, Owners: s.owner(f.MonitorByID, f.MonitorBy)}
		if f.MonitorRows == 0 {
			st.Status = ChainNA
			st.Verdict = "ما وصل صندوق المراقب."
		} else {
			th := s.threshold(f, "MONITOR")
			st.Threshold = &th
			st.StartAt = f.MonitorFirstAt
			if f.MonitorPending == 0 {
				st.EndAt = f.MonitorLastAt
			}
			if f.MonitorMinutes != nil {
				v := int(*f.MonitorMinutes)
				st.Minutes = &v
			}
			st.Facts = append(st.Facts, fmt.Sprintf("%d بند، منها %d بعدها تنتظر حكم.", f.MonitorRows, f.MonitorPending))
			switch {
			case f.MonitorPending > 0 && f.MonitorFirstAt != nil && now.Sub(*f.MonitorFirstAt).Minutes() > float64(th):
				st.Status = ChainMissed
				st.Verdict = fmt.Sprintf("أكو بنود ما انحكمت، وفاتها المعتاد (%s).", FmtMinutes(th))
			case f.MonitorPending > 0:
				st.Status = ChainWaiting
				st.Verdict = "أكو بنود تنتظر حكم المراقب."
			case st.Minutes != nil && *st.Minutes > th:
				st.Status = ChainLate
				st.Verdict = fmt.Sprintf("حكم بمعدل %s، والمعتاد ما يتجاوز %s.", FmtMinutes(*st.Minutes), FmtMinutes(th))
			default:
				st.Status = ChainOK
				st.Verdict = "حكم على كل البنود بوقتها."
			}
		}
		add(st)
	}

	for _, st := range ch.Stations {
		switch st.Status {
		case ChainOK:
			ch.Score.OK++
		case ChainLate:
			ch.Score.Late++
		case ChainMissed:
			ch.Score.Missed++
		case ChainIssue:
			ch.Score.Issue++
		case ChainWaiting:
			ch.Score.Waiting++
		}
	}
	ch.Summary = chainSummary(ch)
	return ch
}

func qualityStatusLabel(s string) string {
	switch s {
	case "CONTACTED_OK":
		return "الزبون راضي"
	case "CONTACTED_ISSUE":
		return "الزبون عنده ملاحظة"
	case "CONVERTED":
		return "تحوّلت لشكوى"
	case "RECONTACT":
		return "يحتاج إعادة اتصال"
	case "CLOSED":
		return "انسدّت"
	case "PENDING":
		return "تنتظر الاتصال"
	}
	return s
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// chainSummary جملة ماتركس عن الحجز: وين انكسرت السلسلة ومنو.
func chainSummary(ch *BookingChain) string {
	if ch.Legacy {
		return "حجز قديم (قبل النظام) — السلسلة للعلم بس، ما ينحاسب عليها أحد."
	}
	var worst []string
	for _, st := range ch.Stations {
		if st.Status == ChainMissed || st.Status == ChainLate {
			who := ""
			if len(st.Owners) == 1 && st.Owners[0].Name != "" {
				who = " (" + st.Owners[0].Name + ")"
			}
			worst = append(worst, st.Title+who)
		}
	}
	measured := ch.Score.OK + ch.Score.Late + ch.Score.Missed + ch.Score.Issue
	if len(worst) == 0 {
		if measured == 0 {
			return "الحجز بأوله — بعد ماكو محطات تنقاس."
		}
		return fmt.Sprintf("السلسلة سليمة لحد هسه: %d محطة بوقتها%s.", ch.Score.OK+ch.Score.Issue,
			map[bool]string{true: fmt.Sprintf("، %d بيها ملاحظة", ch.Score.Issue), false: ""}[ch.Score.Issue > 0])
	}
	return fmt.Sprintf("السلسلة انكسرت بـ%d محطة: %s.", len(worst), strings.Join(worst, "، "))
}

// ═══ تقرير الدور ═══

type ChainStationStat struct {
	Key       string `json:"key"`
	Title     string `json:"title"`
	Count     int    `json:"count"`
	OK        int    `json:"ok"`
	Late      int    `json:"late"`
	Missed    int    `json:"missed"`
	Issue     int    `json:"issue"`
	MedianMin *int   `json:"medianMin"`
	AvgMin    *int   `json:"avgMin"`
	TeamMed   *int   `json:"teamMedianMin"`
}

type ChainCodeRef struct {
	ID      string `json:"id"`
	Code    string `json:"code"`
	Station string `json:"station"`
	Status  string `json:"status"`
	Note    string `json:"note"`
}

type ChainEmployeeReport struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Total       int                `json:"total"`
	OnTime      int                `json:"onTime"`
	Late        int                `json:"late"`
	Missed      int                `json:"missed"`
	Issues      int                `json:"issues"`
	OnTimePct   int                `json:"onTimePct"`
	PrevPct     *int               `json:"prevPct"`
	Stations    []ChainStationStat `json:"stations"`
	Worst       string             `json:"worst"`
	Bad         []ChainCodeRef     `json:"bad"`
	LateWeekday string             `json:"lateWeekday"`
	LateHour    *int               `json:"lateHour"`
	RatingAvg   *float64           `json:"ratingAvg"`
	RatingCount int                `json:"ratingCount"`
	RatingNotes []string           `json:"ratingNotes"`
	Insights    []string           `json:"insights"`
}

type RoleChainReport struct {
	Role      string                `json:"role"`
	Title     string                `json:"title"`
	Days      int                   `json:"days"`
	Bookings  int                   `json:"bookings"`
	TeamPct   int                   `json:"teamPct"`
	PrevPct   *int                  `json:"prevPct"`
	Stations  []ChainStationStat    `json:"stations"`
	Employees []ChainEmployeeReport `json:"employees"`
	Insights  []string              `json:"insights"`
}

type chainAgg struct {
	total, ok, late, missed, issue int
	st                             map[string]*chainStAgg
	bad                            []ChainCodeRef
	wd                             map[time.Weekday]int
	hr                             map[int]int
}

type chainStAgg struct {
	ok, late, missed, issue int
	mins                    []int
}

func newAgg() *chainAgg {
	return &chainAgg{st: map[string]*chainStAgg{}, wd: map[time.Weekday]int{}, hr: map[int]int{}}
}

func (a *chainAgg) add(ch *BookingChain, st *ChainStation, status string) {
	if status == ChainNA || status == ChainWaiting || status == "" {
		return
	}
	sa := a.st[st.Key]
	if sa == nil {
		sa = &chainStAgg{}
		a.st[st.Key] = sa
	}
	a.total++
	switch status {
	case ChainOK:
		a.ok++
		sa.ok++
	case ChainIssue:
		a.ok++
		a.issue++
		sa.issue++
		sa.ok++
	case ChainLate:
		a.late++
		sa.late++
	case ChainMissed:
		a.missed++
		sa.missed++
	}
	if st.Minutes != nil && st.EndAt != nil {
		sa.mins = append(sa.mins, *st.Minutes)
	}
	if status == ChainLate || status == ChainMissed || status == ChainIssue {
		note := st.Verdict
		if status == ChainIssue {
			note = strings.Join(st.Issues, " ")
		}
		a.bad = append(a.bad, ChainCodeRef{ID: ch.BookingID, Code: ch.Code, Station: st.Title, Status: status, Note: note})
		if status != ChainIssue && st.StartAt != nil {
			t := st.StartAt.In(debriefLoc)
			a.wd[t.Weekday()]++
			a.hr[t.Hour()]++
		}
	}
}

func (a *chainAgg) pct() int {
	if a.total == 0 {
		return 0
	}
	return int(math.Round(float64(a.ok) * 100 / float64(a.total)))
}

var arWeekdays = map[time.Weekday]string{
	time.Saturday: "السبت", time.Sunday: "الأحد", time.Monday: "الاثنين", time.Tuesday: "الثلاثاء",
	time.Wednesday: "الأربعاء", time.Thursday: "الخميس", time.Friday: "الجمعة",
}

// collect يمرّ على حجوزات الفترة ويجمع لكل موظف بالدور.
func (s *MatrixChainService) collect(facts []repository.ChainFacts, role string, from, to time.Time, now time.Time) (map[string]*chainAgg, *chainAgg, map[string]string, int) {
	per := map[string]*chainAgg{}
	team := newAgg()
	names := map[string]string{}
	n := 0
	for i := range facts {
		f := &facts[i]
		if f.CreatedAt.Before(from) || !f.CreatedAt.Before(to) {
			continue
		}
		n++
		ch := s.chainOf(f, now)
		for j := range ch.Stations {
			st := &ch.Stations[j]
			if st.Role != role {
				continue
			}
			if len(st.Owners) == 0 {
				// محطة فاتت/تأخرت وما معروف منو مسؤولها (ما صارت أصلاً) —
				// تنحسب على الفريق وتطلع بالقائمة «بلا مسؤول».
				if per[""] == nil {
					per[""] = newAgg()
				}
				names[""] = "— ما معروف منو المسؤول (الخطوة ما انسجّلت)"
				per[""].add(ch, st, st.Status)
				team.add(ch, st, st.Status)
				continue
			}
			for _, o := range st.Owners {
				status := st.Status
				if o.Status != "" {
					status = o.Status
				}
				if per[o.ID] == nil {
					per[o.ID] = newAgg()
				}
				names[o.ID] = o.Name
				per[o.ID].add(ch, st, status)
				team.add(ch, st, status)
			}
		}
	}
	return per, team, names, n
}

func statsOf(a *chainAgg, team *chainAgg) []ChainStationStat {
	out := []ChainStationStat{}
	for _, d := range chainStations {
		sa := a.st[d.Key]
		if sa == nil {
			continue
		}
		st := ChainStationStat{Key: d.Key, Title: d.Title, OK: sa.ok, Late: sa.late, Missed: sa.missed, Issue: sa.issue,
			Count: sa.ok + sa.late + sa.missed}
		if len(sa.mins) > 0 {
			m := chainMedian(sa.mins)
			sum := 0
			for _, v := range sa.mins {
				sum += v
			}
			avg := sum / len(sa.mins)
			st.MedianMin, st.AvgMin = &m, &avg
		}
		if team != nil {
			if ts := team.st[d.Key]; ts != nil && len(ts.mins) > 0 {
				tm := chainMedian(ts.mins)
				st.TeamMed = &tm
			}
		}
		out = append(out, st)
	}
	return out
}

// RoleReport التقرير العميق لدور: كل موظف ومحطاته، ويا المقارنة والاتجاه والنمط.
func (s *MatrixChainService) RoleReport(role string, days int) (*RoleChainReport, error) {
	if _, ok := ChainRoleTitles[role]; !ok {
		return nil, fmt.Errorf("دور غير معروف")
	}
	if days <= 0 || days > 180 {
		days = 30
	}
	now := time.Now()
	from := now.AddDate(0, 0, -days)
	prevFrom := from.AddDate(0, 0, -days)
	facts, err := s.repo.Since(prevFrom)
	if err != nil {
		return nil, err
	}
	per, team, names, n := s.collect(facts, role, from, now.Add(time.Minute), now)
	prevPer, prevTeam, _, _ := s.collect(facts, role, prevFrom, from, now)

	rep := &RoleChainReport{Role: role, Title: ChainRoleTitles[role], Days: days, Bookings: n,
		TeamPct: team.pct(), Stations: statsOf(team, nil), Employees: []ChainEmployeeReport{}, Insights: []string{}}
	if prevTeam.total > 0 {
		p := prevTeam.pct()
		rep.PrevPct = &p
	}

	// تقييمات الليدرية للفنيين — لتقرير الفنيين بس.
	ratings := map[string][]repository.CrewRatingRow{}
	if role == ChainRoleTech {
		if rows, err := s.repo.RatingsSince(from); err == nil {
			for _, r := range rows {
				ratings[r.TechnicianID] = append(ratings[r.TechnicianID], r)
				if _, ok := per[r.TechnicianID]; !ok {
					per[r.TechnicianID] = newAgg()
					names[r.TechnicianID] = s.names(r.TechnicianID)
				}
			}
		}
	}

	for id, a := range per {
		e := ChainEmployeeReport{ID: id, Name: names[id], Total: a.total, OnTime: a.ok - a.issue, Late: a.late,
			Missed: a.missed, Issues: a.issue, OnTimePct: a.pct(), Stations: statsOf(a, team), Insights: []string{},
			RatingNotes: []string{}}
		if e.Name == "" && s.names != nil && id != "" {
			e.Name = s.names(id)
		}
		if pa := prevPer[id]; pa != nil && pa.total > 0 {
			p := pa.pct()
			e.PrevPct = &p
		}
		// أسوأ محطة: أكثر نسبة تأخير/فوات.
		worstRate := 0.0
		for _, st := range e.Stations {
			if st.Count == 0 {
				continue
			}
			r := float64(st.Late+st.Missed) / float64(st.Count)
			if r > worstRate {
				worstRate, e.Worst = r, st.Title
			}
		}
		sort.Slice(a.bad, func(i, j int) bool { return a.bad[i].Code > a.bad[j].Code })
		e.Bad = a.bad
		if len(e.Bad) > 300 {
			e.Bad = e.Bad[:300]
		}
		// النمط: إذا ٤٠٪ أو أكثر من تأخيراته بنفس اليوم/الساعة.
		lates := a.late + a.missed
		if lates >= 3 {
			bw, bc := time.Sunday, 0
			for d, c := range a.wd {
				if c > bc {
					bw, bc = d, c
				}
			}
			if bc*10 >= lates*4 {
				e.LateWeekday = arWeekdays[bw]
			}
			bh, hc := 0, 0
			for h, c := range a.hr {
				if c > hc {
					bh, hc = h, c
				}
			}
			if hc*10 >= lates*4 {
				e.LateHour = &bh
			}
		}
		if rs := ratings[id]; len(rs) > 0 {
			sum := 0
			for _, r := range rs {
				sum += r.Score
				if r.Note != nil && *r.Note != "" && len(e.RatingNotes) < 5 {
					e.RatingNotes = append(e.RatingNotes, fmt.Sprintf("%s (%s، %d/5): %s", r.BookingCode, r.LeaderName, r.Score, *r.Note))
				}
			}
			avg := math.Round(float64(sum)*10/float64(len(rs))) / 10
			e.RatingAvg, e.RatingCount = &avg, len(rs)
		}
		e.Insights = employeeInsights(&e, rep.TeamPct)
		rep.Employees = append(rep.Employees, e)
	}
	sort.Slice(rep.Employees, func(i, j int) bool {
		if rep.Employees[i].Total == 0 || rep.Employees[j].Total == 0 {
			return rep.Employees[i].Total > rep.Employees[j].Total
		}
		return rep.Employees[i].OnTimePct < rep.Employees[j].OnTimePct
	})
	rep.Insights = teamInsights(rep)
	return rep, nil
}

// employeeInsights جمل ماتركس عن الموظف — مقارنة، اتجاه، نمط. بلا أسماء لمزوّد خارجي (كلها محلية).
func employeeInsights(e *ChainEmployeeReport, teamPct int) []string {
	out := []string{}
	if e.Total < 3 {
		if e.Total > 0 {
			out = append(out, fmt.Sprintf("شغله بهالفترة قليل (%d خطوة)، فبعد وكت نحكم عليه.", e.Total))
		}
	} else {
		d := e.OnTimePct - teamPct
		switch {
		case d <= -15:
			out = append(out, fmt.Sprintf("التزامه بالوقت أقل من زملائه بـ%d نقطة (%d٪، والفريق %d٪).", -d, e.OnTimePct, teamPct))
		case d >= 15:
			out = append(out, fmt.Sprintf("التزامه بالوقت أحسن من زملائه بـ%d نقطة (%d٪، والفريق %d٪).", d, e.OnTimePct, teamPct))
		}
	}
	if e.PrevPct != nil && e.Total >= 3 {
		d := e.OnTimePct - *e.PrevPct
		if d >= 10 {
			out = append(out, fmt.Sprintf("📈 تحسّن: الفترة الفاتت چان %d٪ وهسه %d٪.", *e.PrevPct, e.OnTimePct))
		} else if d <= -10 {
			out = append(out, fmt.Sprintf("📉 تراجع: الفترة الفاتت چان %d٪ وهسه %d٪.", *e.PrevPct, e.OnTimePct))
		}
	}
	for _, st := range e.Stations {
		if st.MedianMin != nil && st.TeamMed != nil && *st.TeamMed > 0 && st.Count >= 3 {
			r := float64(*st.MedianMin) / float64(*st.TeamMed)
			if r >= 1.5 {
				out = append(out, fmt.Sprintf("بـ«%s» ياخذ %s، وزملاؤه ياخذون %s، يعني أبطأ منهم %.1f مرة.", st.Title, FmtMinutes(*st.MedianMin), FmtMinutes(*st.TeamMed), r))
			} else if r <= 0.6 {
				out = append(out, fmt.Sprintf("بـ«%s» أسرع من زملائه: ياخذ %s وهم %s.", st.Title, FmtMinutes(*st.MedianMin), FmtMinutes(*st.TeamMed)))
			}
		}
		if st.Missed >= 2 {
			out = append(out, fmt.Sprintf("«%s» ما صارت %d مرات.", st.Title, st.Missed))
		}
	}
	if e.LateWeekday != "" {
		out = append(out, "أغلب تأخيره يصير يوم "+e.LateWeekday+".")
	}
	if e.LateHour != nil {
		out = append(out, fmt.Sprintf("وبالذات حوالي الساعة %d:00.", *e.LateHour))
	}
	if e.Issues >= 2 {
		out = append(out, fmt.Sprintf("%d مرة خلّص بوقته بس الشغل بيه نقص (الحجوزات مذكورة تحت).", e.Issues))
	}
	if e.RatingAvg != nil {
		out = append(out, fmt.Sprintf("الليدرية قيّموه %.1f من 5 (على %d حجز).", *e.RatingAvg, e.RatingCount))
	}
	return out
}

func teamInsights(rep *RoleChainReport) []string {
	out := []string{}
	if rep.Bookings == 0 {
		return append(out, "ماكو حجوزات بهالفترة.")
	}
	measured := 0
	for _, st := range rep.Stations {
		measured += st.Count
	}
	if measured == 0 {
		return append(out, fmt.Sprintf("%d حجز بآخر %d يوم، وبعد ماكو شغل وصل لـ%s حتى نقيسه.", rep.Bookings, rep.Days, rep.Title))
	}
	out = append(out, fmt.Sprintf("%d حجز بآخر %d يوم. %s مسؤولين عن %d خطوة، صار منها بوقتها %d٪.", rep.Bookings, rep.Days, rep.Title, measured, rep.TeamPct))
	if rep.PrevPct != nil {
		d := rep.TeamPct - *rep.PrevPct
		if d != 0 {
			arrow := "📈"
			if d < 0 {
				arrow = "📉"
			}
			out = append(out, fmt.Sprintf("%s الفترة الفاتت چانت %d٪.", arrow, *rep.PrevPct))
		}
	}
	// أضعف محطة بالفريق.
	worst, wr := "", 0.0
	for _, st := range rep.Stations {
		if st.Count < 3 {
			continue
		}
		r := float64(st.Late+st.Missed) / float64(st.Count)
		if r > wr {
			worst, wr = st.Title, r
		}
	}
	if worst != "" && wr > 0 {
		out = append(out, fmt.Sprintf("أضعف خطوة: «%s»، %d٪ منها تأخرت أو ما صارت.", worst, int(math.Round(wr*100))))
	}
	return out
}

// ═══ «شنو تعلّم ماتركس» ═══
// الحدود مو مكتوبة بالكود: ماتركس يتعلّمها من شغل الشركة الحقيقي (آخر ٩٠ يوم)
// ويعيد حسابها كل ساعة. هنا يعرضها: الوسيط، الحد، وعدد العينات.
type ChainLearned struct {
	Key       string `json:"key"`
	Title     string `json:"title"`
	Samples   int    `json:"samples"`
	MedianMin int    `json:"medianMin"`
	LimitMin  int    `json:"limitMin"`
	Source    string `json:"source"`
}

type ChainLearning struct {
	Stations       []ChainLearned `json:"stations"`
	ServicesOwnLim int            `json:"servicesOwnLimits"`
	LearnedAt      time.Time      `json:"learnedAt"`
}

func (s *MatrixChainService) Learning() ChainLearning {
	g, _ := s.thresholds()
	s.mu.Lock()
	defer s.mu.Unlock()
	out := ChainLearning{Stations: []ChainLearned{}, ServicesOwnLim: s.thSvcN, LearnedAt: s.thAt}
	for _, d := range chainStations {
		if d.Key == "CREATE" || d.Key == "INVENTORY" {
			continue
		}
		l := ChainLearned{Key: d.Key, Title: d.Title, Samples: s.thN[d.Key], MedianMin: s.thMed[d.Key], LimitMin: g[d.Key]}
		switch {
		case d.Key == "QUALITY":
			l.Source = "قرار (ع): يومين بعد الإنجاز"
		case l.Samples == 0:
			l.Source = "بعد ماكو حجوزات كافية، فحطّينا حد أدنى"
		case l.LimitMin == d.Floor && l.MedianMin*3/2 < d.Floor:
			l.Source = "الشغل سريع بالعادة، فانطبق الحد الأدنى"
		default:
			l.Source = "تعلّمه من المعتاد بالشركة × ١٫٥"
		}
		out.Stations = append(out.Stations, l)
	}
	return out
}
