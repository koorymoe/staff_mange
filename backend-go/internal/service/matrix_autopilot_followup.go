package service

import (
	"fmt"
	"log"
	"time"

	"staffmange-api/internal/model"
)

// ═══ ماتركس: التذكيرات الجديدة + المتابعة والتصعيد والتعلّم ═══
//
// التذكير بلا متابعة نص شغل: ماتركس يرجع كل يوم يشوف هل الشي الي ذكّر
// بي انحل. إذا انحل يسجّله (هذا دليل إن التذكير نفع)، وإذا فات ٣ أيام
// وما انحل يصعده للمدير بـ«ينتظر قرارك». وإذا المدير رفض نفس النوع ٣
// مرات بشهر، يفهم إنه مو مرغوب ويوقفه لحاله لحد ما المدير يرجّعه.

// الأنواع الي إلها «انحل» واضح نگدر نفحصه.
var followableKinds = []string{
	model.AiActionPaperworkReminder, model.AiActionUnstaffedAlert, model.AiActionExtraTaskOverdue,
	model.AiActionGpsExpiry, model.AiActionVehicleDocExpiry, model.AiActionInvoiceApproval, model.AiActionLowStock,
}

// ٧. اشتراك جي بي اس يخلص خلال ١٤ يوم — للموظف الي باعه، مرة بالأسبوع.
func (s *MatrixAutopilotService) gpsExpiry(week string, dayStart time.Time) (int, error) {
	rows, err := s.actions.GpsExpiringSoon(14)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, r := range rows {
		emp := r.EmployeeID
		name := r.CustomerName
		if name == "" {
			name = "زبون بلا اسم"
		}
		msg := fmt.Sprintf("🤖 ماتركس — اشتراك جي بي اس للزبون «%s» يخلص بعد %d يوم (%s). اتصل بي للتجديد.",
			name, r.DaysLeft, r.SubscriptionEnd.Format("2006-01-02"))
		if s.act(model.AiAction{Kind: model.AiActionGpsExpiry, EntityType: "GPS_SUBSCRIPTION", EntityID: r.ID,
			Period: week, TargetEmployeeID: &emp, TargetLabel: s.actions.EmployeeName(emp),
			Summary: fmt.Sprintf("ذكّر البائع بتجديد اشتراك «%s» (%d يوم)", name, r.DaysLeft),
			Details: why(map[string]any{"customer": name, "subscriptionEnd": r.SubscriptionEnd, "daysLeft": r.DaysLeft})}, dayStart,
			func() error { return s.notif.Create(emp, "AI_AUTOPILOT", msg) }) {
			n++
		}
	}
	return n, nil
}

// ٨. وثيقة سيارة تخلص خلال ٣٠ يوم — للأسطول، مرة بالأسبوع.
func (s *MatrixAutopilotService) vehicleDocs(week string, dayStart time.Time) (int, error) {
	rows, err := s.actions.VehicleDocsExpiring(30)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, d := range rows {
		state := "تخلص"
		if d.ExpiryDate.Before(time.Now()) {
			state = "خالصة من"
		}
		msg := fmt.Sprintf("🤖 ماتركس — %s السيارة «%s» (%s) %s %s. جدّدوها قبل ما تطلع بالشارع.",
			d.DocumentType, d.VehicleName, d.PlateNumber, state, d.ExpiryDate.Format("2006-01-02"))
		if s.act(model.AiAction{Kind: model.AiActionVehicleDocExpiry, EntityType: "VEHICLE_DOCUMENT", EntityID: d.ID,
			Period: week, TargetLabel: "الأسطول",
			Summary: fmt.Sprintf("نبّه الأسطول: %s «%s» %s %s", d.DocumentType, d.VehicleName, state, d.ExpiryDate.Format("2006-01-02")),
			Details: why(map[string]any{"document": d.DocumentType, "vehicle": d.VehicleName, "plate": d.PlateNumber, "expiryDate": d.ExpiryDate})}, dayStart,
			func() error { return s.notif.CreateForPermission("vehicle_management", "AI_AUTOPILOT", msg) }) {
			n++
		}
	}
	return n, nil
}

// ٩. مهمة إضافية فات موعدها — للموظف نفسه، مرة باليوم.
func (s *MatrixAutopilotService) overdueTasks(today string, dayStart time.Time) (int, error) {
	rows, err := s.actions.OverdueExtraTasks()
	if err != nil {
		return 0, err
	}
	n := 0
	for _, t := range rows {
		emp := t.AssignedToID
		days := int(time.Since(t.DueAt).Hours() / 24)
		msg := fmt.Sprintf("🤖 ماتركس — المهمة «%s» فات موعدها (%s). كمّلها أو بلّغ الي كلّفك إذا تحتاج وقت.",
			t.Title, t.DueAt.In(debriefLoc).Format("2006-01-02 15:04"))
		if s.act(model.AiAction{Kind: model.AiActionExtraTaskOverdue, EntityType: "EXTRA_TASK", EntityID: t.ID,
			Period: today, TargetEmployeeID: &emp, TargetLabel: s.actions.EmployeeName(emp),
			Summary: fmt.Sprintf("ذكّر بمهمة «%s» المتأخرة", t.Title),
			Details: why(map[string]any{"title": t.Title, "dueAt": t.DueAt, "daysLate": days})}, dayStart,
			func() error { return s.notif.Create(emp, "AI_AUTOPILOT", msg) }) {
			n++
		}
	}
	return n, nil
}

// ١٠. عنده شغل اليوم وما سجّل حضور — تذكير لطيف، مو غياب.
func (s *MatrixAutopilotService) attendanceNudges(today string, dayStart time.Time) (int, error) {
	rows, err := s.actions.WorkTodayNoCheckIn()
	if err != nil {
		return 0, err
	}
	n := 0
	for _, r := range rows {
		emp := r.EmployeeID
		msg := fmt.Sprintf("🤖 ماتركس — عندك حجز %s اليوم وبعدك ما مسجّل حضور. سجّله من النظام حتى يبين شغلك.", r.BookingCode)
		if s.act(model.AiAction{Kind: model.AiActionAttendanceNudge, EntityType: "EMPLOYEE", EntityID: emp,
			Period: today, TargetEmployeeID: &emp, TargetLabel: s.actions.EmployeeName(emp),
			Summary: fmt.Sprintf("ذكّر بتسجيل الحضور (عنده حجز %s)", r.BookingCode),
			Details: why(map[string]any{"bookingCode": r.BookingCode})}, dayStart,
			func() error { return s.notif.Create(emp, "AI_AUTOPILOT", msg) }) {
			n++
		}
	}
	return n, nil
}

// ١١. أداة رصيدها ٢٠٪ أو أقل — لصاحب صلاحية الجرد، مرة بالأسبوع.
func (s *MatrixAutopilotService) lowStock(week string, dayStart time.Time) (int, error) {
	rows, err := s.actions.LowStockTools()
	if err != nil {
		return 0, err
	}
	n := 0
	for _, t := range rows {
		msg := fmt.Sprintf("🤖 ماتركس — الأداة «%s» باقي منها %d من %d بالمخزن. فكّروا بالتزويد.", t.Name, t.Available, t.Total)
		if s.act(model.AiAction{Kind: model.AiActionLowStock, EntityType: "TOOL", EntityID: t.ID,
			Period: week, TargetLabel: "المخزن",
			Summary: fmt.Sprintf("نبّه المخزن: «%s» باقي %d من %d", t.Name, t.Available, t.Total),
			Details: why(map[string]any{"tool": t.Name, "available": t.Available, "total": t.Total})}, dayStart,
			func() error { return s.notif.CreateForPermission("inventory", "AI_AUTOPILOT", msg) }) {
			n++
		}
	}
	return n, nil
}

// ١٢. فاتورة ليدر مرفوعة وما انعتمدت من ٣ أيام — للمحاسبة، مرة باليوم.
func (s *MatrixAutopilotService) staleInvoices(today string, dayStart time.Time) (int, error) {
	rows, err := s.actions.StaleSubmittedInvoices(3)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, r := range rows {
		msg := fmt.Sprintf("🤖 ماتركس — فاتورة حجز %s تنتظر الاعتماد من %d يوم.", r.BookingCode, r.Days)
		if s.act(model.AiAction{Kind: model.AiActionInvoiceApproval, EntityType: "LEADER_INVOICE", EntityID: r.ID,
			Period: today, TargetLabel: "المحاسبة",
			Summary: fmt.Sprintf("ذكّر المحاسبة بفاتورة %s (%d يوم)", r.BookingCode, r.Days),
			Details: why(map[string]any{"bookingCode": r.BookingCode, "submittedAt": r.CreatedAt, "days": r.Days})}, dayStart,
			func() error {
				return s.notif.CreateForRolesOrPermission([]string{"FINANCE"}, "finance_audit", "AI_AUTOPILOT", msg)
			}) {
			n++
		}
	}
	return n, nil
}

// followUp يفحص الأفعال المفتوحة: انحلت؟ ولا صار وقت تصعيدها؟
func (s *MatrixAutopilotService) followUp() {
	open, err := s.actions.OpenForFollowUp(followableKinds)
	if err != nil {
		log.Printf("[ai] متابعة ماتركس: %v", err)
		return
	}
	for _, a := range open {
		ok, err := s.actions.IsResolved(a)
		if err != nil {
			log.Printf("[ai] فحص انحلال %s: %v", a.ID, err)
			continue
		}
		if ok {
			_ = s.actions.MarkResolved(a.ID)
			continue
		}
		if a.EscalatedAt == nil && time.Since(a.CreatedAt) >= model.AiActionEscalateAfter {
			_ = s.actions.MarkEscalated(a.ID)
			// المهمة المتأخرة: الي كلّفها لازم يدري، مو بس المدير.
			if a.Kind == model.AiActionExtraTaskOverdue {
				if by := s.actions.TaskAssigner(a.EntityID); by != "" {
					_ = s.notif.Create(by, "AI_AUTOPILOT",
						fmt.Sprintf("🤖 ماتركس — %s، وذكّرته من %d يوم وبعدها ما انحلت.", a.Summary, int(time.Since(a.CreatedAt).Hours()/24)))
				}
			}
		}
	}
}

// afterUndo التعلّم: ٣ رفضات لنفس النوع بشهر؟ يوقفه ويبلّغ.
func (s *MatrixAutopilotService) afterUndo(actionID string) {
	kind, err := s.actions.KindOf(actionID)
	if err != nil {
		return
	}
	n, err := s.actions.RejectsSince(kind, time.Now().AddDate(0, 0, -30))
	if err != nil || n < model.AiActionPauseAfterRejects {
		return
	}
	label := model.AiActionLabels[kind]
	reason := fmt.Sprintf("رفضته %d مرات خلال ٣٠ يوم", n)
	if err := s.actions.Pause(kind, reason); err != nil {
		return
	}
	_ = s.notif.CreateForRolesOrPermission([]string{"OWNER", "ADMIN"}, "", "AI_DECISIONS",
		fmt.Sprintf("🤖 ماتركس — وقّفت «%s» لأنك %s. تگدر ترجّعه من صندوق القرارات.", label, reason))
}

func (s *MatrixAutopilotService) Resume(kind string) error { return s.actions.Resume(kind) }

// ═══ عين ماتركس ═══
//
// العين تعكس حالة حقيقية بس — ما تعاقب ولا تخترع. هادئة: ماكو شي
// مفتوح. منتبهة: تذكير مفتوح. حمرة: تذكير صعد أو ٣ مفتوحة فأكثر.
// والتنبيه اللحظي (حضور اليوم، خطر التأخير) ما ينحسب — هذني مو تقصير.

type WatchItem struct {
	Summary   string    `json:"summary"`
	Kind      string    `json:"kind"`
	Label     string    `json:"label"`
	Escalated bool      `json:"escalated"`
	Since     time.Time `json:"since"`
}

type WatchState struct {
	Level string      `json:"level"` // CALM | ALERT | RED
	Open  int         `json:"open"`
	Items []WatchItem `json:"items"`
}

const watchRedOpen = 3

func (s *MatrixAutopilotService) WatchState(employeeID string) (*WatchState, error) {
	rows, err := s.actions.OpenForEmployee(employeeID)
	if err != nil {
		return nil, err
	}
	out := &WatchState{Level: "CALM", Items: []WatchItem{}}
	escalated := false
	for _, a := range rows {
		esc := a.EscalatedAt != nil
		escalated = escalated || esc
		out.Items = append(out.Items, WatchItem{Summary: a.Summary, Kind: a.Kind,
			Label: model.AiActionLabels[a.Kind], Escalated: esc, Since: a.CreatedAt})
	}
	out.Open = len(rows)
	switch {
	case escalated || out.Open >= watchRedOpen:
		out.Level = "RED"
	case out.Open > 0:
		out.Level = "ALERT"
	}
	return out, nil
}
