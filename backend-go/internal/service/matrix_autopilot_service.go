package service

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ═══ ماتركس «موظف ويانه» — ينفّذ البسيط لحاله ═══
//
// قرار (ع): ماتركس يسوي الأشياء الروتينية بنفسه بدل ما ينتظر أحد —
// تذكيرات وتنبيهات بس. **ماكو** غرامة، نقاط، خصم، تعديل فاتورة أو
// سعر، تكليف كادر، أو حذف: هذني تبقى «تنتظر قرارك» بصندوق المدير.
//
// كل فعل: يتسجّل بـAiAction أول (مرة بالفترة، ومحظور ٣٠ يوم إذا المدير
// رفضه)، وبعدين ينبعث الإشعار. ومفتاح matrix_autopilot_enabled يطفي
// كلشي بضغطة.

const (
	autopilotHour     = 10 // بعد فحوصات ماتركس (٨) وتوقّع التأخير (٧)، ووقت كافي لتسجيل الحضور
	autopilotDailyCap = 30
)

type MatrixAutopilotService struct {
	actions *repository.AiActionRepository
	aiRepo  *repository.AiRepository
	notif   *repository.NotificationRepository
	// السيارات الي انذكرت وثائقها ويا إشعار الاستبدال بهالدورة — حتى ما يتكرر إشعارها.
	vehicleNoted map[string]bool
	switches     *repository.SystemSwitchRepository
	delay        *DelayPredictionService
	insights     *MatrixInsightsService
	// proposals عدّاد اقتراحات ماتركس المعلّقة — للإشعار الصباحي والصندوق.
	proposals func() int
	// ResolveFor: آخر فحص لكل موظف (حد ١٠ ثواني).
	resolveMu   sync.Mutex
	resolveLast map[string]time.Time
	// undoHooks: الأفعال التنفيذية ترجّع الي سوّته لمن المدير يتراجع.
	undoHooks map[string]func(a model.AiAction) error
}

// OnUndo يربط ترجيع فعل تنفيذي بنوعه (ماتركس كلّف كادر ← يشيل التكليف).
func (s *MatrixAutopilotService) OnUndo(kind string, f func(a model.AiAction) error) {
	if s.undoHooks == nil {
		s.undoHooks = map[string]func(model.AiAction) error{}
	}
	s.undoHooks[kind] = f
}

func (s *MatrixAutopilotService) SetProposalCounter(f func() int) { s.proposals = f }

func (s *MatrixAutopilotService) pendingProposals() int {
	if s.proposals == nil {
		return 0
	}
	return s.proposals()
}

func NewMatrixAutopilotService(actions *repository.AiActionRepository, aiRepo *repository.AiRepository,
	notif *repository.NotificationRepository, switches *repository.SystemSwitchRepository,
	delay *DelayPredictionService, insights *MatrixInsightsService) *MatrixAutopilotService {
	return &MatrixAutopilotService{actions: actions, aiRepo: aiRepo, notif: notif, switches: switches, delay: delay, insights: insights}
}

// Enabled حالة المفتاح (الافتراضي شغّال).
func (s *MatrixAutopilotService) Enabled() bool {
	all, err := s.switches.All()
	if err != nil {
		return false // ما نعرف؟ ما ننفّذ.
	}
	return all[model.SwitchMatrixAutopilot]
}

// RunIfDue مرة باليوم بعد ١٠ الصبح بغداد.
func (s *MatrixAutopilotService) RunIfDue() error {
	now := time.Now().In(debriefLoc)
	if now.Hour() < autopilotHour {
		return nil
	}
	today := now.Format("2006-01-02")
	claimed, err := s.aiRepo.ClaimDailyMarker("DAILY_MATRIX_AUTOPILOT", today)
	if err != nil || !claimed {
		return err
	}
	done := 0
	if s.Enabled() {
		s.followUp()
		done = s.runActions(now)
	}
	return s.sendMorningSummary(done)
}

func (s *MatrixAutopilotService) runActions(now time.Time) int {
	today := now.Format("2006-01-02")
	week := isoWeekMonday(now)
	dayStart := baghdadMidnight(today)
	n := 0
	steps := []func() (int, error){
		func() (int, error) { return s.paperworkReminders(week, dayStart) },
		func() (int, error) { return s.delayWarnings(today, dayStart) },
		func() (int, error) { return s.customerFollowUps(week, dayStart) },
		func() (int, error) { return s.unstaffedAlerts(now.AddDate(0, 0, 1).Format("2006-01-02"), dayStart) },
		func() (int, error) { return s.replacementAlerts(week, dayStart) },
		func() (int, error) { return s.gpsExpiry(week, dayStart) },
		func() (int, error) { return s.vehicleDocs(week, dayStart) },
		func() (int, error) { return s.overdueTasks(today, dayStart) },
		func() (int, error) { return s.attendanceNudges(today, dayStart) },
		func() (int, error) { return s.lowStock(week, dayStart) },
		func() (int, error) { return s.staleInvoices(today, dayStart) },
		func() (int, error) { return s.crewRatings(today, dayStart) },
		func() (int, error) { return s.afterInventory(today, dayStart) },
		func() (int, error) { return s.monitorBacklog(today, dayStart) },
	}
	for _, step := range steps {
		c, err := step()
		if err != nil {
			log.Printf("[ai] ماتركس التنفيذي: %v", err)
		}
		n += c
	}
	return n
}

// act يسجّل الفعل وبعدين ينفّذه — بحد يومي لكل نوع.
// Act نفس act للخدمات الثانية (الأفعال التنفيذية).
func (s *MatrixAutopilotService) Act(a model.AiAction, dayStart time.Time, send func() error) bool {
	return s.act(a, dayStart, send)
}

func (s *MatrixAutopilotService) act(a model.AiAction, dayStart time.Time, send func() error) bool {
	if c, err := s.actions.CountSince(a.Kind, dayStart); err != nil || c >= autopilotDailyCap {
		return false
	}
	ok, err := s.actions.Claim(a)
	if err != nil || !ok {
		return false
	}
	if err := send(); err != nil {
		log.Printf("[ai] تعذر إرسال فعل %s: %v", a.Kind, err)
	}
	return true
}

// why يحوّل أسباب الفعل لـJSON يطلع بزر «ليش؟».
func why(v map[string]any) model.NullJSON {
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}

// ١. الورق المتأخر — تذكير لليدر نفسه، مرة بالأسبوع.
func (s *MatrixAutopilotService) paperworkReminders(week string, dayStart time.Time) (int, error) {
	rows, err := s.aiRepo.LatePaperworkRows("")
	if err != nil {
		return 0, err
	}
	n := 0
	for leader, items := range groupLatePaperwork(rows) {
		// حماية: الورق مسؤولية الليدر بس — الفني العادي ما يتذكّر بيه أبداً.
		if sub, err := s.actions.Subject(leader); err != nil || !sub.IsLeader {
			continue
		}
		all := make([]string, 0, len(items))
		for _, it := range items {
			all = append(all, it.BookingCode)
		}
		codes := []string{}
		for i, it := range items {
			if i == 8 {
				codes = append(codes, "…")
				break
			}
			codes = append(codes, it.BookingCode)
		}
		msg := fmt.Sprintf("🤖 ماتركس — عندك %d حجز منجز من أكثر من يومين وناقصه فاتورة أو تقرير: %s. كمّلها اليوم رجاءً.",
			len(items), strings.Join(codes, "، "))
		leaderID := leader
		if s.act(model.AiAction{Kind: model.AiActionPaperworkReminder, EntityType: "EMPLOYEE", EntityID: leader,
			Period: week, TargetEmployeeID: &leaderID, TargetLabel: s.actions.EmployeeName(leader),
			Summary: fmt.Sprintf("ذكّر الليدر بـ%d حجز ناقصه ورق", len(items)),
			Details: why(map[string]any{"bookingCodes": all, "items": items})}, dayStart,
			func() error { return s.notif.Create(leaderID, "AI_AUTOPILOT", msg) }) {
			n++
		}
	}
	return n, nil
}

// ٢. خطر التأخير اليوم — تنبيه لليدر المكلّف (الإداريين يوصلهم تنبيه الصبح أصلاً).
func (s *MatrixAutopilotService) delayWarnings(today string, dayStart time.Time) (int, error) {
	risks, err := s.delay.Risks(today, today)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, r := range risks {
		if r.LeaderID == "" {
			continue
		}
		leaderID := r.LeaderID
		msg := fmt.Sprintf("🤖 ماتركس — حجز %s اليوم متوقع ياخذ ~%s والوقت المتاح %s. انتبه للوقت، وإذا تحتاج دعم بلّغ التنسيق من وقت.",
			r.BookingCode, fmtHours(r.ExpectedMinutes), fmtHours(r.AvailableMinutes))
		if s.act(model.AiAction{Kind: model.AiActionDelayWarning, EntityType: "BOOKING", EntityID: r.BookingID,
			Period: today, TargetEmployeeID: &leaderID, TargetLabel: s.actions.EmployeeName(leaderID),
			Summary: fmt.Sprintf("نبّه الليدر إن حجز %s ممكن يطوّل", r.BookingCode),
			Details: why(map[string]any{"bookingCode": r.BookingCode, "expectedMinutes": r.ExpectedMinutes,
				"availableMinutes": r.AvailableMinutes, "samples": r.Samples, "basis": r.Basis, "limitedBy": r.LimitedBy})}, dayStart,
			func() error { return s.notif.Create(leaderID, "AI_AUTOPILOT", msg) }) {
			n++
		}
	}
	return n, nil
}

// ٣. زبون قريب يزعل — الجودة تتصل بي، مرة بالأسبوع لكل زبون.
func (s *MatrixAutopilotService) customerFollowUps(week string, dayStart time.Time) (int, error) {
	rows, err := s.aiRepo.CustomerRiskRows()
	if err != nil {
		return 0, err
	}
	n := 0
	for _, r := range rows {
		factors := customerRiskFactors(r)
		if len(factors) < riskMinFactors {
			continue
		}
		msg := fmt.Sprintf("🤖 ماتركس — الزبون رقم %d (آخر حجز %s) عليه أكثر من علامة زعل. اتصلوا بي اليوم واسألوا عن رضاه.",
			r.CustomerCode, r.LatestBookingCode)
		if s.act(model.AiAction{Kind: model.AiActionCustomerFollowUp, EntityType: "CUSTOMER", EntityID: r.CustomerID,
			Period: week, TargetLabel: "الجودة",
			Summary: fmt.Sprintf("طلب من الجودة تتصل بالزبون رقم %d", r.CustomerCode),
			Details: why(map[string]any{"customerCode": r.CustomerCode, "latestBooking": r.LatestBookingCode, "factors": factors})}, dayStart,
			func() error {
				return s.notif.CreateForRolesOrPermission([]string{}, "quality_control", "AI_AUTOPILOT", msg)
			}) {
			n++
		}
	}
	return n, nil
}

// ٤. حجز باچر مثبّت وبلا كادر — تنبيه للتنسيق ويه اقتراح. **ما يكلّف.**
func (s *MatrixAutopilotService) unstaffedAlerts(tomorrow string, dayStart time.Time) (int, error) {
	rows, err := s.actions.UnstaffedOn(tomorrow)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, b := range rows {
		hint, suggested := "", ""
		if rec, err := s.insights.CrewRecommendation(b.ID); err == nil && rec != nil && len(rec.Leaders) > 0 {
			suggested = rec.Leaders[0].Name
			hint = fmt.Sprintf(" الأنسب حسب ماتركس: %s.", suggested)
		}
		msg := fmt.Sprintf("🤖 ماتركس — حجز %s باچر الساعة %s مثبّت وماكو عليه كادر.%s",
			b.Code, b.ScheduledAt.In(debriefLoc).Format("15:04"), hint)
		if s.act(model.AiAction{Kind: model.AiActionUnstaffedAlert, EntityType: "BOOKING", EntityID: b.ID,
			Period: tomorrow, TargetLabel: "التنسيق",
			Summary: fmt.Sprintf("نبّه التنسيق إن حجز %s باچر بلا كادر", b.Code),
			Details: why(map[string]any{"bookingCode": b.Code, "scheduledAt": b.ScheduledAt, "suggestedLeader": suggested})}, dayStart,
			func() error {
				return s.notif.CreateForRolesOrPermission([]string{"HR_COORDINATOR"}, "coordinator", "AI_AUTOPILOT", msg)
			}) {
			n++
		}
	}
	return n, nil
}

// ٥. أجهزة وسيارات تكلّف — تنبيه لأصحابها، مرة بالأسبوع لكل واحد.
func (s *MatrixAutopilotService) replacementAlerts(week string, dayStart time.Time) (int, error) {
	rep, err := s.insights.Replacements(true, true)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, it := range rep.It {
		msg := fmt.Sprintf("🤖 ماتركس — الجهاز «%s»: %s. فكّروا باستبداله بدل التصليح.", it.Name, it.Reason)
		if s.act(model.AiAction{Kind: model.AiActionReplacementAlert, EntityType: "IT_ASSET", EntityID: it.ID,
			Period: week, TargetLabel: "تقنية المعلومات", Summary: fmt.Sprintf("اقترح استبدال الجهاز «%s»", it.Name),
			Details: why(map[string]any{"repairCount": it.RepairCount, "repairCost": it.RepairCost, "reason": it.Reason})}, dayStart,
			func() error { return s.notif.CreateForPermission("it_assets", "AI_AUTOPILOT", msg) }) {
			n++
		}
	}
	// وثائق السيارة الي تخلص تنضم لنفس إشعار الاستبدال — إشعار واحد للسيارة.
	s.vehicleNoted = map[string]bool{}
	docs := map[string][]string{}
	if rows, err := s.actions.VehicleDocsExpiring(30); err == nil {
		for _, d := range rows {
			docs[d.PlateNumber] = append(docs[d.PlateNumber], fmt.Sprintf("%s تخلص %s", d.DocumentType, d.ExpiryDate.Format("2006-01-02")))
		}
	}
	for _, v := range rep.Vehicles {
		msg := fmt.Sprintf("🤖 ماتركس — السيارة «%s» (%s): %s. فكّروا باستبدالها.", v.Name, v.PlateNumber, strings.Join(v.Reasons, "، "))
		if d := docs[v.PlateNumber]; len(d) > 0 {
			msg += " وكذلك: " + strings.Join(d, "، ") + "."
		}
		if s.act(model.AiAction{Kind: model.AiActionReplacementAlert, EntityType: "VEHICLE", EntityID: v.ID,
			Period: week, TargetLabel: "الأسطول", Summary: fmt.Sprintf("اقترح استبدال السيارة «%s»", v.Name),
			Details: why(map[string]any{"cost12m": v.Cost12m, "incidents180d": v.Incidents180d, "reasons": v.Reasons})}, dayStart,
			func() error { return s.notif.CreateForPermission("vehicle_management", "AI_AUTOPILOT", msg) }) {
			n++
			s.vehicleNoted[v.PlateNumber] = true
		}
	}
	return n, nil
}

// sendMorningSummary إشعار واحد للمالك والمدير: شكد سوّى وشكد ينتظر.
func (s *MatrixAutopilotService) sendMorningSummary(done int) error {
	pending, err := s.actions.PendingAiVerdicts(200)
	if err != nil {
		return err
	}
	escalated, _ := s.actions.Escalated()
	props := s.pendingProposals()
	if done == 0 && len(pending) == 0 && len(escalated) == 0 && props == 0 {
		return nil
	}
	msg := fmt.Sprintf("🤖 ماتركس — اليوم سوّيت %d تذكير لحالي، وأكو %d حكم ينتظر قرارك، و%d تذكير ما انحل وصعدته إلك، وعندي %d اقتراح. افتح «صندوق قرارات ماتركس».",
		done, len(pending), len(escalated), props)
	today := time.Now().In(debriefLoc).Format("2006-01-02")
	if risks, err := s.delay.Risks(today, today); err == nil && len(risks) > 0 {
		msg += fmt.Sprintf(" ⏱️ و%d حجز اليوم ممكن يطوّل — التنسيق والليدرية تبلّغوا.", len(risks))
	}
	if !s.Enabled() {
		msg = fmt.Sprintf("🤖 ماتركس — التنفيذ التلقائي مطفي. أكو %d حكم ينتظر قرارك بصندوق القرارات.", len(pending))
	}
	return s.notif.CreateForRolesOrPermission([]string{"OWNER", "ADMIN"}, "", "AI_DECISIONS", msg)
}

// ── صندوق القرارات ──

type DecisionsBox struct {
	Day              string                    `json:"day"`
	AutopilotEnabled bool                      `json:"autopilotEnabled"`
	Done             []model.AiAction          `json:"done"`
	Pending          []model.PendingAiDecision `json:"pending"`
	Unstaffed        []model.UnstaffedBooking  `json:"unstaffed"`
	Escalated        []model.AiAction          `json:"escalated"`
	Paused           []model.AiActionKindPause `json:"paused"`
	Accuracy         *model.MatrixAccuracy     `json:"accuracy"`
	Proposals        int                       `json:"proposals"`
	Labels           map[string]string         `json:"labels"`
}

// Decisions محتوى الصندوق ليوم (بغداد) — افتراضياً اليوم.
func (s *MatrixAutopilotService) Decisions(day string) (*DecisionsBox, error) {
	now := time.Now().In(debriefLoc)
	if _, err := time.Parse("2006-01-02", day); err != nil {
		day = now.Format("2006-01-02")
	}
	from := baghdadMidnight(day)
	done, err := s.actions.ListBetween(from, from.Add(24*time.Hour))
	if err != nil {
		return nil, err
	}
	pending, err := s.actions.PendingAiVerdicts(100)
	if err != nil {
		return nil, err
	}
	unstaffed, err := s.actions.UnstaffedOn(now.AddDate(0, 0, 1).Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	escalated, err := s.actions.Escalated()
	if err != nil {
		return nil, err
	}
	paused, err := s.actions.ListPaused()
	if err != nil {
		return nil, err
	}
	acc, err := s.actions.Accuracy()
	if err != nil {
		return nil, err
	}
	return &DecisionsBox{Day: day, AutopilotEnabled: s.Enabled(), Done: done, Pending: pending,
		Unstaffed: unstaffed, Escalated: escalated, Paused: paused, Accuracy: acc, Labels: model.AiActionLabels,
		Proposals: s.pendingProposals()}, nil
}

// Undo المدير يرفض فعلاً — وماتركس يتعلّم منه.
func (s *MatrixAutopilotService) Undo(id, by string) error {
	a, _ := s.actions.Find(id)
	if err := s.actions.Undo(id, by); err != nil {
		return err
	}
	if a != nil {
		if hook := s.undoHooks[a.Kind]; hook != nil {
			if err := hook(*a); err != nil {
				log.Printf("[ai] ترجيع %s: %v", a.Kind, err)
			}
		}
	}
	s.afterUndo(id)
	return nil
}
