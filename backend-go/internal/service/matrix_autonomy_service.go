package service

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strings"
	"sync"
	"time"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ═══ المرحلة الثالثة: ماتركس ينفّذ البسيط لحاله — بأقفال أمان ═══
//
// قرار (ع) 10-05. كل فعل تنفيذي:
//   ١. إله مفتاح عند المالك، **مطفي افتراضياً** (model.SystemSwitchDefault).
//   ٢. ما يشتغل إلا بعد ما تثبت دقة اقتراحات نوعه: ≥٢٠ اقتراح محسوم بآخر ٣٠
//      يوم وقبول ≥٨٠٪ (AutonomyGate).
//   ٣. يتسجّل بـAiAction ويا «ليش؟»، وينرجع بضغطة (OnUndo يرجّع الي سوّاه).
//   ٤. ٣ رفضات بشهر لنفس النوع = يوقف لحاله (منطق الطيار الآلي الموجود).
//   ٥. حد ٥ أفعال من النوع باليوم.
// ماكو فلوس ولا غرامات ولا نقاط ولا رسالة للزبون.

const (
	AutonomyMinDecided   = 20
	AutonomyMinAcceptPct = 80
	autonomyDailyMax     = 5
	autoCrewHour         = 18 // ٦ المسا بغداد
)

// AutonomyGate البوابة: مؤهل؟ (دالة صافية — تنفحص باختبار وحدة).
func AutonomyGate(decided, accepted int) (eligible bool, pct int) {
	if decided > 0 {
		pct = int(math.Round(float64(accepted) * 100 / float64(decided)))
	}
	return decided >= AutonomyMinDecided && pct >= AutonomyMinAcceptPct, pct
}

type MatrixAutonomyService struct {
	auto     *MatrixAutopilotService
	actions  *repository.AiActionRepository
	notif    *repository.NotificationRepository
	switches *repository.SystemSwitchRepository
	suggest  *MatrixSuggestService
	bookings *BookingService

	mu      sync.Mutex
	lastRun map[string]string
}

func NewMatrixAutonomyService(auto *MatrixAutopilotService, actions *repository.AiActionRepository,
	notif *repository.NotificationRepository, switches *repository.SystemSwitchRepository,
	suggest *MatrixSuggestService, bookings *BookingService) *MatrixAutonomyService {
	s := &MatrixAutonomyService{auto: auto, actions: actions, notif: notif, switches: switches,
		suggest: suggest, bookings: bookings, lastRun: map[string]string{}}
	auto.OnUndo(model.AiActionAutoCrew, s.undoAutoCrew)
	return s
}

func (s *MatrixAutonomyService) on(key string) bool {
	all, err := s.switches.All()
	return err == nil && all[key]
}

// CrewGate حالة بوابة الكادر: كم اقتراح انحسم وكم انقبل بآخر ٣٠ يوم.
func (s *MatrixAutonomyService) CrewGate() (decided, accepted, pct int, eligible bool) {
	rows, err := s.suggest.Repo().Decided(30)
	if err != nil {
		return
	}
	for _, r := range rows {
		if r.Kind != "CREW" {
			continue
		}
		decided++
		if r.Outcome == "ACCEPTED" {
			accepted++
		}
	}
	eligible, pct = AutonomyGate(decided, accepted)
	return
}

// Start كل ١٥ دقيقة يشوف إذا وقت فعل حان (مرة باليوم لكل فعل).
func (s *MatrixAutonomyService) Start() {
	go func() {
		for {
			time.Sleep(15 * time.Minute)
			s.Tick(time.Now())
		}
	}()
}

func (s *MatrixAutonomyService) once(kind, day string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.lastRun[kind] == day {
		return false
	}
	s.lastRun[kind] = day
	return true
}

func (s *MatrixAutonomyService) Tick(now time.Time) {
	local := now.In(debriefLoc)
	today := local.Format("2006-01-02")
	dayStart := baghdadMidnight(today)
	if local.Hour() >= autoCrewHour && s.once(model.AiActionAutoCrew, today) {
		if n, err := s.autoCrew(local, dayStart); err != nil {
			log.Printf("[ai] ماتركس يكلّف: %v", err)
		} else if n > 0 {
			log.Printf("[ai] ماتركس كلّف كادر %d حجز", n)
		}
	}
	if local.Hour() >= 10 && s.once(model.AiActionMonitorEscalate, today) {
		if err := s.monitorEscalate(today, dayStart); err != nil {
			log.Printf("[ai] تصعيد المراقب: %v", err)
		}
	}
}

type autoCrewDetails struct {
	BookingCode string            `json:"bookingCode"`
	ScheduledAt time.Time         `json:"scheduledAt"`
	Leader      string            `json:"leader"`
	LeaderName  string            `json:"leaderName"`
	Techs       map[string]string `json:"techs"` // role → employeeId
	TechNames   []string          `json:"techNames"`
	Why         []string          `json:"why"`
	Gate        string            `json:"gate"`
}

// autoCrew: حجوزات باچر المثبّتة بلا كادر لحد ٦ المسا — ماتركس يكلّف الاقتراح
// الكامل بس (ليدر + العدد المعتاد، كلهم فاضين).
func (s *MatrixAutonomyService) autoCrew(local time.Time, dayStart time.Time) (int, error) {
	if !s.on(model.SwitchMatrixAutoCrew) || !s.auto.Enabled() {
		return 0, nil
	}
	decided, accepted, pct, ok := s.CrewGate()
	if !ok {
		log.Printf("[ai] ماتركس ما يكلّف: الدقة %d٪ من %d (يحتاج %d٪ من %d)", pct, decided, AutonomyMinAcceptPct, AutonomyMinDecided)
		return 0, nil
	}
	gate := fmt.Sprintf("دقة اقتراحاته للكادر %d٪ (%d من %d بآخر ٣٠ يوم)", pct, accepted, decided)
	tomorrow := local.AddDate(0, 0, 1).Format("2006-01-02")
	rows, err := s.actions.UnstaffedOn(tomorrow)
	if err != nil {
		return 0, err
	}
	n := 0
	roles := []string{"TECH_1", "TECH_2", "TECH_3"}
	for _, ub := range rows {
		if n >= autonomyDailyMax {
			break
		}
		b, crew, why, err := s.suggest.CrewFor(ub.ID)
		if err != nil || crew == nil || b.SupervisorID != nil || len(b.TechIDs) > 0 {
			continue
		}
		complete := len(crew.Techs) == crew.CrewSize && (b.Solo || crew.Leader != nil)
		if !complete || len(crew.Techs) > len(roles) {
			continue // اقتراح ناقص — نخليه للإنسان
		}
		d := autoCrewDetails{BookingCode: b.Code, ScheduledAt: *b.ScheduledAt, Techs: map[string]string{}, Why: why, Gate: gate}
		if crew.Leader != nil {
			d.Leader, d.LeaderName = crew.Leader.ID, crew.Leader.Name
		}
		for i, t := range crew.Techs {
			d.Techs[roles[i]] = t.ID
			d.TechNames = append(d.TechNames, t.Name)
		}
		names := d.TechNames
		if d.LeaderName != "" {
			names = append([]string{"الليدر " + d.LeaderName}, names...)
		}
		failed := false
		a := model.AiAction{Kind: model.AiActionAutoCrew, EntityType: "BOOKING", EntityID: b.ID, Period: tomorrow,
			TargetLabel: "التنسيق", Summary: fmt.Sprintf("كلّف كادر حجز %s باچر: %s", b.Code, strings.Join(names, "، ")),
			Details: why2(d)}
		done := s.auto.Act(a, dayStart, func() error {
			if err := s.apply(b.ID, d); err != nil {
				failed = true
				return err
			}
			return nil
		})
		if !done {
			continue
		}
		if failed {
			s.actions.Forget(model.AiActionAutoCrew, b.ID, tomorrow)
			continue
		}
		n++
		at := d.ScheduledAt.In(debriefLoc).Format("15:04")
		msg := fmt.Sprintf("🤖 ماتركس — حجز %s باچر الساعة %s ما انحدد إله كادر لحد المسا، فكلّفت: %s. السبب: %s. %s. إذا ما يناسب، تراجع من «تنسيق الحجوزات».",
			b.Code, at, strings.Join(names, "، "), strings.Join(why, " "), gate)
		_ = s.notif.CreateForRolesOrPermission([]string{"HR_COORDINATOR"}, "coordinator", "AI_AUTOPILOT", msg)
		crewMsg := fmt.Sprintf("🤖 ماتركس — انكلّفت بحجز %s باچر الساعة %s. التفاصيل بـ«مهامي».", b.Code, at)
		if d.Leader != "" {
			_ = s.notif.Create(d.Leader, "AI_AUTOPILOT", crewMsg)
		}
		for _, id := range d.Techs {
			_ = s.notif.Create(id, "AI_AUTOPILOT", crewMsg)
		}
	}
	return n, nil
}

func why2(d autoCrewDetails) model.NullJSON {
	b, _ := json.Marshal(d)
	return b
}

// apply ينفّذ عبر نفس مسارات الخدمة الي يستعملها المنسق. إذا فشل بالنص، يرجّع الي انسوّى.
func (s *MatrixAutonomyService) apply(bookingID string, d autoCrewDetails) error {
	if d.Leader != "" {
		id := d.Leader
		if _, err := s.bookings.SetSupervisor(bookingID, &id); err != nil {
			return err
		}
	}
	done := []string{}
	for role, emp := range d.Techs {
		if _, err := s.bookings.Assign(bookingID, model.AssignBookingRequest{EmployeeID: emp, Role: role}, ""); err != nil {
			for _, r := range done {
				_, _ = s.bookings.Unassign(bookingID, r)
			}
			if d.Leader != "" {
				_, _ = s.bookings.SetSupervisor(bookingID, nil)
			}
			return err
		}
		done = append(done, role)
	}
	return nil
}

// undoAutoCrew يشيل بس الي كلّفه ماتركس **وبعده نفسه** — إذا المنسق غيّر
// شي بعده، تغييره ما ينلمس. ويبلّغ الكادر بالإلغاء.
func (s *MatrixAutonomyService) undoAutoCrew(a model.AiAction) error {
	var d autoCrewDetails
	if err := json.Unmarshal(a.Details, &d); err != nil {
		return err
	}
	b, err := s.suggest.Repo().Booking(a.EntityID)
	if err != nil {
		return err
	}
	msg := fmt.Sprintf("🤖 ماتركس — انلغى تكليفك بحجز %s (المنسق راجع القرار).", d.BookingCode)
	if d.Leader != "" && b.SupervisorID != nil && *b.SupervisorID == d.Leader {
		if _, err := s.bookings.SetSupervisor(a.EntityID, nil); err == nil {
			_ = s.notif.Create(d.Leader, "AI_AUTOPILOT", msg)
		}
	}
	current := s.suggest.Repo().AssignmentRoles(a.EntityID)
	for role, emp := range d.Techs {
		if current[role] == emp {
			if _, err := s.bookings.Unassign(a.EntityID, role); err == nil {
				_ = s.notif.Create(emp, "AI_AUTOPILOT", msg)
			}
		}
	}
	return nil
}

// monitorEscalate بنود المراقب +٤٨ ساعة ← المدير (إشعار بس، ماكو تغيير بيانات).
func (s *MatrixAutonomyService) monitorEscalate(today string, dayStart time.Time) error {
	if !s.on(model.SwitchMatrixMonitorEscalate) || !s.auto.Enabled() {
		return nil
	}
	n, err := s.actions.MonitorOverdue()
	if err != nil || n == 0 {
		return err
	}
	msg := fmt.Sprintf("🤖 ماتركس — %d بند بصندوق المراقب صارله أكثر من يومين بلا حكم. التفاصيل بتقرير المراقبين (مكتب المدير ← ماتركس).", n)
	s.auto.Act(model.AiAction{Kind: model.AiActionMonitorEscalate, EntityType: "MONITOR_INBOX", EntityID: "monitor-inbox",
		Period: today, TargetLabel: "المدير", Summary: fmt.Sprintf("صعّد %d بند متأخر بصندوق المراقب للمدير", n),
		Details: why(map[string]any{"overdue": n})}, dayStart,
		func() error { return s.notif.CreateForRolesOrPermission([]string{"ADMIN"}, "", "AI_AUTOPILOT", msg) })
	return nil
}

// ═══ الشاشة ═══

type AutonomyStatus struct {
	AutoCrew struct {
		On       bool `json:"on"`
		Decided  int  `json:"decided"`
		Accepted int  `json:"accepted"`
		Pct      int  `json:"pct"`
		Eligible bool `json:"eligible"`
		Need     int  `json:"needDecided"`
		NeedPct  int  `json:"needPct"`
	} `json:"autoCrew"`
	MonitorEscalate struct {
		On      bool `json:"on"`
		Overdue int  `json:"overdue"`
	} `json:"monitorEscalate"`
	AutopilotOn bool             `json:"autopilotOn"`
	Recent      []model.AiAction `json:"recent"`
}

func (s *MatrixAutonomyService) Status() AutonomyStatus {
	var st AutonomyStatus
	st.AutopilotOn = s.auto.Enabled()
	st.AutoCrew.On = s.on(model.SwitchMatrixAutoCrew)
	st.AutoCrew.Decided, st.AutoCrew.Accepted, st.AutoCrew.Pct, st.AutoCrew.Eligible = s.CrewGate()
	st.AutoCrew.Need, st.AutoCrew.NeedPct = AutonomyMinDecided, AutonomyMinAcceptPct
	st.MonitorEscalate.On = s.on(model.SwitchMatrixMonitorEscalate)
	st.MonitorEscalate.Overdue, _ = s.actions.MonitorOverdue()
	st.Recent, _ = s.actions.RecentOfKinds([]string{model.AiActionAutoCrew, model.AiActionMonitorEscalate}, 15)
	return st
}

// ActiveAutoCrew تكليفات ماتركس الحية (مو مرجوعة) — شارة المنسق.
func (s *MatrixAutonomyService) ActiveAutoCrew() []map[string]string {
	rows, _ := s.actions.RecentOfKinds([]string{model.AiActionAutoCrew}, 50)
	out := []map[string]string{}
	for _, a := range rows {
		if a.Status == model.AiActionDone {
			out = append(out, map[string]string{"actionId": a.ID, "bookingId": a.EntityID, "summary": a.Summary})
		}
	}
	return out
}

// UndoAutoCrew المنسق يتراجع عن تكليف ماتركس (بس هالنوع).
func (s *MatrixAutonomyService) UndoAutoCrew(actionID, by string) error {
	a, err := s.actions.Find(actionID)
	if err != nil || a.Kind != model.AiActionAutoCrew {
		return fmt.Errorf("مو تكليف من ماتركس")
	}
	return s.auto.Undo(actionID, by)
}
