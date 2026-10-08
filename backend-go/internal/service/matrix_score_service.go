package service

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ═══ تقييم ماتركس — قرار (ع) 10-05 ═══
// «من يبدي الحجز لحد ما ينتهي ماتركس كون يقيم… الفاتورة بوقتها نقطتين،
// تأخر نقطة وحدة». كل محطة بسلسلة الحجز إلها صاحب، وحكم ماتركس عليها
// (بوقتها / متأخرة / فيها غلط / ما صارت) يصير نقاط لصاحبها:
//
//	بوقتها = ٢ · متأخرة أو بيها غلط = ١ · ما صارت ووقتها فات = ٠ (من أصل ٢)
//
// ونفس الشي يومياً (الحضور، المهام المضافة بموعدها) وأسبوعياً للمشاريع.
// ⚠️ درجة تقييم بس: ماكو فلوس، وما يلمس نقاط الانضباط. المالك ومدير النظام
// ما ينقيّمون. ومفتاح matrix_scoring مطفي لحد ما المالك يشغّله.
//
// التقييم النهائي = ٦٠٪ ماتركس + ٤٠٪ البشر (الليدر، الإداري، المراقب، الجودة).

const (
	ScoreMatrixWeight = 0.6
	ScoreHumanWeight  = 0.4
	scoreLookbackDays = 60
)

// StationPoints نقاط حكم محطة. ok=false: ما تنحسب (بعدها أو ما تنطبق).
func StationPoints(status string) (int, bool) {
	switch status {
	case ChainOK:
		return 2, true
	case ChainLate, ChainIssue:
		return 1, true
	case ChainMissed:
		return 0, true
	}
	return 0, false
}

// TaskPoints مهمة مضافة: بموعدها ٢، خلصت خلال يوم بعد الموعد ١، غير هيچ ٠.
// ok=false: موعدها بعد ما فات ولا خلصت.
func TaskPoints(due time.Time, done *time.Time, now time.Time) (int, bool) {
	if done != nil {
		switch {
		case !done.After(due):
			return 2, true
		case done.Sub(due) <= 24*time.Hour:
			return 1, true
		default:
			return 0, true
		}
	}
	if now.Sub(due) > 24*time.Hour {
		return 0, true
	}
	return 0, false
}

// FinalScore ٠٫٦ ماتركس + ٠٫٤ البشر. بلا تقييم بشري: ماتركس بس.
func FinalScore(matrixPct float64, hasMatrix bool, humanPct float64, hasHuman bool) (float64, bool) {
	switch {
	case hasMatrix && hasHuman:
		return ScoreMatrixWeight*matrixPct + ScoreHumanWeight*humanPct, true
	case hasMatrix:
		return matrixPct, true
	case hasHuman:
		return humanPct, true
	}
	return 0, false
}

type MatrixScoreService struct {
	repo     *repository.MatrixScoreRepository
	chain    *MatrixChainService
	switches *repository.SystemSwitchRepository
	projects func() (*ProjectsReport, error)
	lastRun  time.Time
}

func NewMatrixScoreService(repo *repository.MatrixScoreRepository, chain *MatrixChainService,
	switches *repository.SystemSwitchRepository, projects func() (*ProjectsReport, error)) *MatrixScoreService {
	return &MatrixScoreService{repo: repo, chain: chain, switches: switches, projects: projects}
}

func (s *MatrixScoreService) Repo() *repository.MatrixScoreRepository { return s.repo }

func (s *MatrixScoreService) On() bool {
	all, err := s.switches.All()
	return err == nil && all[model.SwitchMatrixScoring]
}

// RunIfDue كل ساعة (إذا المفتاح شغّال).
func (s *MatrixScoreService) RunIfDue() error {
	if !s.On() || time.Since(s.lastRun) < 55*time.Minute {
		return nil
	}
	s.lastRun = time.Now()
	n, err := s.Run(time.Now())
	if n > 0 {
		log.Printf("matrix score: %d نقطة جديدة", n)
	}
	return err
}

// Run يحسب كل الي ما انحسب — الإعادة آمنة (الفريد يمنع التكرار).
func (s *MatrixScoreService) Run(now time.Time) (int, error) {
	n := 0
	var errs []string
	add := func(r repository.MatrixScoreRow) {
		ok, err := s.repo.Add(r)
		if err != nil {
			errs = append(errs, err.Error())
		} else if ok {
			n++
		}
	}
	if err := s.scoreBookings(now, add); err != nil {
		errs = append(errs, err.Error())
	}
	if err := s.scoreDays(now, add); err != nil {
		errs = append(errs, err.Error())
	}
	if err := s.scoreTasks(now, add); err != nil {
		errs = append(errs, err.Error())
	}
	s.scoreProjects(now, add)
	if len(errs) > 0 {
		return n, errors.New(strings.Join(errs, "; "))
	}
	return n, nil
}

func (s *MatrixScoreService) scoreBookings(now time.Time, add func(repository.MatrixScoreRow)) error {
	facts, err := s.chain.Repo().Since(now.AddDate(0, 0, -scoreLookbackDays))
	if err != nil {
		return err
	}
	_ = s.repo.DropGoneBookings()
	// المحطة الي ما اشتغل عليها أحد: إذا دورها إله موظف واحد بس، هو المسؤول.
	sole := map[string]string{
		ChainRoleAccountant: s.repo.SoleHolder("FINANCE", "finance_audit"),
		ChainRoleQuality:    s.repo.SoleHolder("QUALITY_ENGINEER", "quality_control"),
		ChainRoleMonitor:    s.repo.SoleHolder("MONITOR", "monitoring"),
	}
	// كل حجز: نخلي بس (محطة، موظف) المحسوبة هسه — غيرها ينمسح (صاحب تغيّر،
	// محطة انشالت بالترحيل للتقني، أو رجعت تنتظر).
	var seen, keep []string
	defer func() { _ = s.repo.KeepOnlyStations(seen, keep) }()
	for i := range facts {
		f := &facts[i]
		seen = append(seen, f.ID)
		ch := s.chain.chainOf(f, now)
		label := "حجز " + ch.Code
		for _, st := range ch.Stations {
			if st.Status == ChainNA || st.Status == ChainWaiting {
				continue
			}
			if len(st.Owners) == 0 && st.Status == ChainMissed && sole[st.Role] != "" {
				st.Owners = []ChainOwner{{ID: sole[st.Role]}}
				st.Verdict += " (محسوبة على المسؤول الوحيد بهالدور.)"
			}
			at := f.CreatedAt
			if t := latestOf(st.EndAt, st.StartAt); t != nil {
				at = *t
			}
			for _, o := range st.Owners {
				status := st.Status
				if o.Status != "" {
					status = o.Status
				}
				pts, ok := StationPoints(status)
				if !ok || o.ID == "" {
					continue
				}
				keep = append(keep, f.ID+"|"+st.Key+"|"+o.ID)
				add(repository.MatrixScoreRow{EmployeeID: o.ID, Source: "BOOKING", SourceID: f.ID, SourceLabel: &label,
					Rule: st.Key, Points: pts, MaxPoints: 2, At: at,
					Reason: fmt.Sprintf("%s · %s: %s", label, st.Title, scoreVerdict(status, st.Verdict))})
			}
		}
	}
	return nil
}

func scoreVerdict(status, verdict string) string {
	head := map[string]string{ChainOK: "صارت بوقتها", ChainLate: "صارت بس متأخرة", ChainIssue: "صارت بس بيها غلط", ChainMissed: "ما صارت ووقتها فات"}[status]
	if verdict == "" || verdict == head {
		return head
	}
	return head + " — " + verdict
}

// scoreDays الحضور لآخر ٣ أيام دوام (الجمعة عطلة، والإجازة المعتمدة ما تنحسب).
func (s *MatrixScoreService) scoreDays(now time.Time, add func(repository.MatrixScoreRow)) error {
	people, err := s.repo.Scorables()
	if err != nil {
		return err
	}
	today := now.In(debriefLoc)
	for back := 1; back <= 3; back++ {
		d := today.AddDate(0, 0, -back)
		if d.Weekday() == time.Friday {
			continue
		}
		day := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, debriefLoc)
		present, leave, err := s.repo.DayFacts(day)
		if err != nil {
			return err
		}
		key := day.Format("2006-01-02")
		label := "يوم " + key
		for _, p := range people {
			if leave[p.ID] || p.CreatedAt.After(day.Add(9*time.Hour)) {
				continue
			}
			r := repository.MatrixScoreRow{EmployeeID: p.ID, Source: "DAY", SourceID: key, SourceLabel: &label,
				Rule: "ATTENDANCE", MaxPoints: 2, At: day.Add(12 * time.Hour).UTC()}
			if _, ok := present[p.ID]; ok {
				r.Points, r.Reason = 2, label+": سجّل حضور."
			} else {
				r.Points, r.Reason = 0, label+": ما سجّل حضور وما عنده إجازة معتمدة."
			}
			add(r)
		}
	}
	return nil
}

func (s *MatrixScoreService) scoreTasks(now time.Time, add func(repository.MatrixScoreRow)) error {
	tasks, err := s.repo.TasksDue(now.AddDate(0, 0, -scoreLookbackDays), now)
	if err != nil {
		return err
	}
	for _, t := range tasks {
		pts, ok := TaskPoints(t.DueAt, t.DoneAt, now)
		if !ok {
			continue
		}
		label := "مهمة «" + t.Title + "»"
		why := map[int]string{2: "خلصت بموعدها.", 1: "خلصت بس بعد الموعد (خلال يوم).", 0: "ما خلصت بموعدها."}[pts]
		add(repository.MatrixScoreRow{EmployeeID: t.AssignedToID, Source: "TASK", SourceID: t.ID, SourceLabel: &label,
			Rule: "TASK_DUE", Points: pts, MaxPoints: 2, At: t.DueAt, Reason: label + ": " + why})
	}
	return nil
}

// scoreProjects مرة بالأسبوع لكل مشروع مفتوح: حكم عين المشاريع على مشرفه.
func (s *MatrixScoreService) scoreProjects(now time.Time, add func(repository.MatrixScoreRow)) {
	if s.projects == nil {
		return
	}
	rep, err := s.projects()
	if err != nil || rep == nil {
		return
	}
	week := PeerWeek(now).Format("2006-01-02")
	for _, p := range rep.Projects {
		pts, ok := StationPoints(p.Status)
		if !ok || p.OwnerID == "" {
			continue
		}
		label := "مشروع «" + p.Name + "»"
		add(repository.MatrixScoreRow{EmployeeID: p.OwnerID, Source: "PROJECT", SourceID: p.ID + "@" + week, SourceLabel: &label,
			Rule: "PROJECT_WEEK", Points: pts, MaxPoints: 2, At: now,
			Reason: fmt.Sprintf("%s (أسبوع %s): %s", label, week, p.Verdict)})
	}
}

// ═══ العرض ═══

type RuleLoss struct {
	Rule   string `json:"rule"`
	Title  string `json:"title"`
	Lost   int    `json:"lost"`
	Count  int    `json:"count"`
	Advice string `json:"advice"`
}

type StaffScore struct {
	Group       string            `json:"group"`
	GroupLabel  string            `json:"groupLabel"`
	Reliability *float64          `json:"reliability"`
	RelParts    []ReliabilityPart `json:"reliabilityParts"`
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Role        string            `json:"role"`
	Earned      int               `json:"earned"`
	Max         int               `json:"max"`
	MatrixPct   *float64          `json:"matrixPct"`
	HumanAvg    *float64          `json:"humanAvg"`
	HumanCount  int               `json:"humanCount"`
	HumanPct    *float64          `json:"humanPct"`
	Final       *float64          `json:"final"`
	PrevFinal   *float64          `json:"prevFinal"`
	NoHuman     bool              `json:"noHuman"`
	TopLosses   []RuleLoss        `json:"topLosses"`
	BySource    map[string][2]int `json:"bySource"` // مكسوب/أقصى لكل مصدر
	// معلومات تنعرض بـ«مؤشرات أداء الموظف» وما تدخل بالنسبة (قرار (ع) 10-08).
	Skills     int  `json:"skills"`
	HasLicense bool `json:"hasLicense"`
	IsLeader   bool `json:"isLeader"`
}

type StaffScoreBoard struct {
	Month    string       `json:"month"`
	On       bool         `json:"on"`
	Staff    []StaffScore `json:"staff"`
	Insights []string     `json:"insights"`
}

type StaffScoreDetail struct {
	StaffScore
	Month  string                      `json:"month"`
	Points []repository.MatrixScoreRow `json:"points"`
	Human  []repository.HumanRow       `json:"human"`
}

var ruleTitles = map[string]string{
	"ATTENDANCE": "تسجيل الحضور", "TASK_DUE": "المهام المضافة بموعدها", "PROJECT_WEEK": "المشاريع",
}

var ruleAdvice = map[string]string{
	"ATTENDANCE":   "سجّل حضورك كل يوم دوام أول ما توصل.",
	"TASK_DUE":     "خلّص المهام قبل موعدها، وإذا ما تلحق گول للي وجّهها قبل الموعد.",
	"PROJECT_WEEK": "حرّك المشروع: كل أسبوع لازم يتقدم خطوة ويتسجل عليه شي.",
	"PAPER":        "سوّي الفاتورة والتقرير نفس يوم الإنجاز.",
	"CONTACT":      "اتصل بالزبون أول ما يتسجل الحجز.",
	"CONFIRM":      "ثبّت الموعد ويا الزبون بسرعة.",
	"CREW":         "حدد الكادر قبل يوم الحجز.",
	"RATING":       "قيّم فنيّينك بعد كل حجز.",
	"INVENTORY":    "جرد العدّة بعد كل حجز.",
	"ACCOUNT":      "دقّق واعتمد فواتير الحجوزات خلال يوم.",
	"QUALITY":      "اتصل بالزبون خلال يومين من الإنجاز.",
	"MONITOR":      "احسم بنود الحجز بالتدقيق بسرعة.",
	"ROUTE":        "انطلق للموقع بوقتك.",
	"WORK":         "ابدأ الشغل أول ما توصل.",
	"COMPLETE":     "سجّل الإنجاز أول ما يخلص الشغل.",
	"RECEIVE":      "استلم المواد وجهّزها قبل الانطلاق.",
}

func ruleTitle(rule string) string {
	if t, ok := ruleTitles[rule]; ok {
		return t
	}
	return chainStationTitle(rule)
}

// MonthRange حدود الشهر بتوقيت بغداد. month = «2026-10» أو فارغ للحالي.
func MonthRange(month string, now time.Time) (string, time.Time, time.Time) {
	t := now.In(debriefLoc)
	if m, err := time.ParseInLocation("2006-01", month, debriefLoc); err == nil {
		t = m
	}
	from := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, debriefLoc)
	return from.Format("2006-01"), from, from.AddDate(0, 1, 0)
}

func (s *MatrixScoreService) Board(month string) (*StaffScoreBoard, error) {
	key, from, to := MonthRange(month, time.Now())
	people, err := s.repo.Scorables()
	if err != nil {
		return nil, err
	}
	cur, err := s.compute(people, from, to)
	if err != nil {
		return nil, err
	}
	prev, _ := s.compute(people, from.AddDate(0, -1, 0), from)
	prevBy := map[string]*float64{}
	for _, p := range prev {
		prevBy[p.ID] = p.Final
	}
	for i := range cur {
		cur[i].PrevFinal = prevBy[cur[i].ID]
	}
	sort.SliceStable(cur, func(i, j int) bool {
		a, b := cur[i].Final, cur[j].Final
		if a == nil || b == nil {
			return a != nil
		}
		return *a > *b
	})
	return &StaffScoreBoard{Month: key, On: s.On(), Staff: cur, Insights: boardInsights(cur)}, nil
}

func (s *MatrixScoreService) Detail(employeeID, month string) (*StaffScoreDetail, error) {
	key, from, to := MonthRange(month, time.Now())
	people, err := s.repo.Scorables()
	if err != nil {
		return nil, err
	}
	var one []repository.Scorable
	for _, p := range people {
		if p.ID == employeeID {
			one = append(one, p)
		}
	}
	if len(one) == 0 {
		return nil, errors.New("هذا الموظف ما ينقيّم (المالك ومدير النظام مستثنين)")
	}
	rows, err := s.compute(one, from, to)
	if err != nil {
		return nil, err
	}
	pts, err := s.repo.Scores(from, to, employeeID)
	if err != nil {
		return nil, err
	}
	hum, err := s.repo.Human(from, to, employeeID)
	if err != nil {
		return nil, err
	}
	return &StaffScoreDetail{StaffScore: rows[0], Month: key, Points: pts, Human: hum}, nil
}

func (s *MatrixScoreService) compute(people []repository.Scorable, from, to time.Time) ([]StaffScore, error) {
	pts, err := s.repo.Scores(from, to, "")
	if err != nil {
		return nil, err
	}
	hum, err := s.repo.Human(from, to, "")
	if err != nil {
		return nil, err
	}
	rel, _ := s.repo.ReliabilityFacts(from, to)
	return BuildScores(people, pts, hum, rel), nil
}

// BuildScores يجمع النقاط والتقييمات لكل موظف (منفصلة حتى تنفحص).
// reliabilityOnlyRules قواعد تدخل بالاعتمادية بس (حضور، جرد) — ما تنزّل التقييم.
var reliabilityOnlyRules = map[string]bool{"ATTENDANCE": true, "INVENTORY": true}

func BuildScores(people []repository.Scorable, pts []repository.MatrixScoreRow, hum []repository.HumanRow, rel map[string]repository.ReliabilityFact) []StaffScore {
	groupOf := map[string]string{}
	for _, p := range people {
		groupOf[p.ID] = ScoreGroup(p)
	}
	type acc struct {
		earned, max int
		loss        map[string]*RuleLoss
		src         map[string][2]int
		rules       map[string][2]int
		hSum, hN    int
	}
	by := map[string]*acc{}
	get := func(id string) *acc {
		if by[id] == nil {
			by[id] = &acc{loss: map[string]*RuleLoss{}, src: map[string][2]int{}, rules: map[string][2]int{}}
		}
		return by[id]
	}
	for _, p := range pts {
		if p.CancelledAt != nil {
			continue
		}
		a := get(p.EmployeeID)
		rv := a.rules[p.Rule]
		a.rules[p.Rule] = [2]int{rv[0] + p.Points, rv[1] + p.MaxPoints}
		// قرار (ع) 10-06: الاعتمادية رقم منفصل — الحضور والجرد يدخلون بيها بس، مو بالتقييم.
		if reliabilityOnlyRules[p.Rule] {
			continue
		}
		a.earned += p.Points
		a.max += p.MaxPoints
		v := a.src[p.Source]
		a.src[p.Source] = [2]int{v[0] + p.Points, v[1] + p.MaxPoints}
		if lost := p.MaxPoints - p.Points; lost > 0 {
			l := a.loss[p.Rule]
			if l == nil {
				l = &RuleLoss{Rule: p.Rule, Title: ruleTitle(p.Rule), Advice: ruleAdvice[p.Rule]}
				a.loss[p.Rule] = l
			}
			l.Lost += lost
			l.Count++
		}
	}
	for _, h := range hum {
		// قرار (ع) 10-06: كل مجموعة ينحسبلها بس مقيّميها المسموحين.
		if g, ok := groupOf[h.RateeID]; ok && !groupRaterStages[g][h.Stage] {
			continue
		}
		a := get(h.RateeID)
		a.hSum += h.Score
		a.hN++
	}
	out := make([]StaffScore, 0, len(people))
	for _, p := range people {
		a := get(p.ID)
		g := groupOf[p.ID]
		s := StaffScore{ID: p.ID, Name: p.Name, Role: p.Role, Earned: a.earned, Max: a.max, HumanCount: a.hN,
			TopLosses: []RuleLoss{}, BySource: a.src, Group: g, GroupLabel: ScoreGroupLabels[g]}
		s.RelParts, s.Reliability = BuildReliability(a.rules, rel[p.ID])
		s.Skills, s.HasLicense, s.IsLeader = rel[p.ID].Skills, rel[p.ID].HasLicense, p.IsLeader
		if a.max > 0 {
			v := float64(a.earned) * 100 / float64(a.max)
			s.MatrixPct = &v
		}
		if a.hN > 0 {
			avg := float64(a.hSum) / float64(a.hN)
			pct := avg * 20
			s.HumanAvg, s.HumanPct = &avg, &pct
		}
		mp, hp := 0.0, 0.0
		if s.MatrixPct != nil {
			mp = *s.MatrixPct
		}
		if s.HumanPct != nil {
			hp = *s.HumanPct
		}
		if f, ok := FinalScore(mp, s.MatrixPct != nil, hp, s.HumanPct != nil); ok {
			s.Final = &f
		}
		s.NoHuman = s.MatrixPct != nil && s.HumanPct == nil
		for _, l := range a.loss {
			s.TopLosses = append(s.TopLosses, *l)
		}
		sort.Slice(s.TopLosses, func(i, j int) bool { return s.TopLosses[i].Lost > s.TopLosses[j].Lost })
		if len(s.TopLosses) > 3 {
			s.TopLosses = s.TopLosses[:3]
		}
		out = append(out, s)
	}
	return out
}

func boardInsights(rows []StaffScore) []string {
	out := []string{}
	scored, noHuman := 0, 0
	var low []string
	for _, r := range rows {
		if r.Final == nil {
			continue
		}
		scored++
		if r.NoHuman {
			noHuman++
		}
		if *r.Final < 60 {
			low = append(low, r.Name)
		}
	}
	if scored == 0 {
		return []string{"بعد ماكو نقاط هالشهر. أول ما يشتغل التقييم، ماتركس يحسب كل حجز ويوم دوام ومشروع."}
	}
	out = append(out, fmt.Sprintf("انحسب تقييم %d موظف هالشهر.", scored))
	if len(low) > 0 {
		more := ""
		if len(low) > 4 {
			more = fmt.Sprintf(" و%d غيرهم", len(low)-4)
			low = low[:4]
		}
		out = append(out, "تحت ٦٠٪: "+strings.Join(low, "، ")+more+".")
	}
	if noHuman > 0 {
		out = append(out, fmt.Sprintf("%d موظف بعد ما قيّمه أحد من البشر، فتقييمه هسه من ماتركس بس.", noHuman))
	}
	return out
}

// ═══ تقييمات البشر بمحطاتها ═══

var ratingStages = map[string][]string{
	"COORD_LEADER": {"LEADER"},                // الإداري يقيّم الليدر لمن يرجع
	"AUDIT":        {"LEADER", "COORDINATOR"}, // المراقب لمن يدقق الحجز
	"QUALITY_CALL": {"LEADER", "COORDINATOR"}, // الجودة بعد الاتصال بالزبون
}

// PeriodKey فترة التقييم الدوري: نص الشهر («2026-10-A» أو «-B»).
func PeriodKey(now time.Time) string {
	t := now.In(debriefLoc)
	half := "A"
	if t.Day() > 15 {
		half = "B"
	}
	return t.Format("2006-01") + "-" + half
}

func validScore(score int) error {
	if score < 1 || score > 5 {
		return errors.New("التقييم من ١ لـ٥")
	}
	return nil
}

// RateOnBooking تقييم بمحطة حجز — الطرف لازم يكون فعلاً ليدر/إداري الحجز.
func (s *MatrixScoreService) RateOnBooking(stage, raterID, bookingID, rateeID string, score int, note *string) error {
	roles, ok := ratingStages[stage]
	if !ok {
		return errors.New("محطة تقييم غير معروفة")
	}
	if err := validScore(score); err != nil {
		return err
	}
	if rateeID == raterID {
		return errors.New("ما تقدر تقيّم نفسك")
	}
	// نفس شروط الطوابير: الحجز خالص، والإداري يقيّم بس ليدر حجز هو ثبّته أو سجّله.
	done, handled, err := s.repo.BookingRateState(bookingID, raterID)
	if err != nil {
		return errors.New("الحجز مو موجود")
	}
	if !done {
		return errors.New("الحجز بعده ما خلص — التقييم بعد الإنجاز")
	}
	if stage == "COORD_LEADER" && !handled {
		return errors.New("تقيّم بس ليدرية الحجوزات الي إنت ثبّتها أو سجّلتها")
	}
	parties, err := s.repo.BookingParties(bookingID)
	if err != nil {
		return err
	}
	for _, p := range parties {
		if p.RateeID != rateeID {
			continue
		}
		for _, r := range roles {
			if p.Role == r {
				return s.repo.SaveRating(repository.StaffRatingInput{RaterID: raterID, RateeID: rateeID, Stage: stage,
					BookingID: bookingID, Score: score, Note: note})
			}
		}
	}
	return errors.New("هذا الموظف مو طرف بهالحجز بهالمحطة")
}

// RatePeriodic المراقب يقيّم المكاتب (المحاسب، المصممة، التقنيين…) كل نص شهر.
func (s *MatrixScoreService) RatePeriodic(raterID, raterRole, rateeID string, score int, note *string) error {
	if err := validScore(score); err != nil {
		return err
	}
	if rateeID == raterID {
		return errors.New("ما تقدر تقيّم نفسك")
	}
	people, err := s.repo.Scorables()
	if err != nil {
		return err
	}
	for _, p := range people {
		if p.ID == rateeID {
			stage := periodicStage(raterRole)
			if !periodicAllowed(stage, p) {
				return errors.New("هذا الموظف يقيّمه غيرك (حسب مجموعته)")
			}
			return s.repo.SaveRating(repository.StaffRatingInput{RaterID: raterID, RateeID: rateeID, Stage: stage,
				Period: PeriodKey(time.Now()), Score: score, Note: note})
		}
	}
	return errors.New("هذا الموظف ما ينقيّم")
}

type BookingRatingState struct {
	Parties []repository.PendingStaffRating `json:"parties"`
	Mine    map[string]int                  `json:"mine"`
}

func (s *MatrixScoreService) BookingRatings(stage, raterID, bookingID string) (*BookingRatingState, error) {
	roles := ratingStages[stage]
	all, err := s.repo.BookingParties(bookingID)
	if err != nil {
		return nil, err
	}
	out := &BookingRatingState{Parties: []repository.PendingStaffRating{}}
	for _, p := range all {
		for _, r := range roles {
			if p.Role == r && p.RateeID != raterID {
				out.Parties = append(out.Parties, p)
			}
		}
	}
	out.Mine, err = s.repo.RatedOn(raterID, bookingID, stage)
	return out, err
}

type PeriodicState struct {
	Period string                `json:"period"`
	People []repository.Scorable `json:"people"`
	Mine   map[string]int        `json:"mine"`
}

func (s *MatrixScoreService) Periodic(raterID, raterRole string) (*PeriodicState, error) {
	people, err := s.repo.Scorables()
	if err != nil {
		return nil, err
	}
	out := &PeriodicState{Period: PeriodKey(time.Now()), People: []repository.Scorable{}}
	stage := periodicStage(raterRole)
	for _, p := range people {
		if p.ID != raterID && periodicAllowed(stage, p) {
			out.People = append(out.People, p)
		}
	}
	out.Mine, err = s.repo.StageRated(raterID, stage, out.Period)
	return out, err
}
