package service

import (
	"fmt"
	"log"
	"strings"
	"time"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
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
	model.AiActionAttendanceNudge, model.AiActionCrewRating, model.AiActionAfterInventory,
	model.AiActionMonitorBacklog,
}

// ResolveFor فحص لحظي بعد كل عملية حفظ للموظف: التذكيرات الي نفّذها تنسكّر
// وإشعارها يتعلّم مقروء، فالعين ترجع هادئة بنفس اللحظة.
// (ع): «ينبّهه سجّل حضور، يسجّل، والإشعار ما يروح والعين ما تتغير».
// محدود: مرة لكل موظف كل ١٠ ثواني حتى ما نضرب القاعدة بكل طلب.
func (s *MatrixAutopilotService) ResolveFor(employeeID string) {
	if employeeID == "" {
		return
	}
	s.resolveMu.Lock()
	if s.resolveLast == nil {
		s.resolveLast = map[string]time.Time{}
	}
	if t, ok := s.resolveLast[employeeID]; ok && time.Since(t) < 10*time.Second {
		s.resolveMu.Unlock()
		return
	}
	s.resolveLast[employeeID] = time.Now()
	s.resolveMu.Unlock()

	open, err := s.actions.OpenFollowableFor(employeeID, followableKinds)
	if err != nil {
		return
	}
	for _, a := range open {
		if ok, err := s.actions.IsResolved(a); err == nil && ok {
			_ = s.actions.MarkResolved(a.ID)
			_ = s.notif.MarkReadNear(employeeID, "AI_AUTOPILOT", a.CreatedAt, 2*time.Minute)
		}
	}
}

// ٧. اشتراك جي بي اس يخلص خلال ١٤ يوم — لمهندس الجودة، مرة بالأسبوع.
//
// قرار (ع): التذكير لمهندس الجودة (هو الي يتصل بالزبون)، مو للبائع — والبائع
// بالاشتراكات المستوردة هو المالك، فچانت تتكدّس بعين المالك. إذا ماكو مهندس
// جودة، يروح لصاحب صلاحية الجودة كإشعار بلا ما ينحسب على شخص.
func (s *MatrixAutopilotService) gpsExpiry(week string, dayStart time.Time) (int, error) {
	rows, err := s.actions.GpsExpiringSoon(14)
	if err != nil {
		return 0, err
	}
	qe := s.ReassignGpsToQuality()
	n := 0
	for _, r := range rows {
		name := r.CustomerName
		if name == "" {
			name = "زبون بلا اسم"
		}
		msg := fmt.Sprintf("🤖 ماتركس — اشتراك جي بي اس للزبون «%s» يخلص بعد %d يوم (%s). اتصلوا بي للتجديد.",
			name, r.DaysLeft, r.SubscriptionEnd.Format("2006-01-02"))
		a := model.AiAction{Kind: model.AiActionGpsExpiry, EntityType: "GPS_SUBSCRIPTION", EntityID: r.ID,
			Period: week, TargetLabel: "الجودة",
			Summary: fmt.Sprintf("ذكّر الجودة بتجديد اشتراك «%s» (%d يوم)", name, r.DaysLeft),
			Details: why(map[string]any{"customer": name, "subscriptionEnd": r.SubscriptionEnd, "daysLeft": r.DaysLeft})}
		send := func() error {
			return s.notif.CreateForRolesOrPermission([]string{"QUALITY_ENGINEER"}, "quality_control", "AI_AUTOPILOT", msg)
		}
		if qe != "" {
			target := qe
			a.TargetEmployeeID, a.TargetLabel = &target, s.actions.EmployeeName(qe)
			send = func() error { return s.notif.Create(target, "AI_AUTOPILOT", msg) }
		}
		if s.act(a, dayStart, send) {
			n++
		}
	}
	return n, nil
}

// ٨. وثيقة سيارة تخلص خلال ٣٠ يوم — للأسطول، مرة بالأسبوع.
//
// 🔴 چانت كل وثيقة إشعار، وفوقها اقتراح الاستبدال بإشعار ثاني — نفس السيارة
// تطلّع ٢-٣ إشعارات. هسه إشعار واحد لكل سيارة بكل وثائقها، وإذا انقترح
// استبدالها بنفس الدورة، الوثائق تنضم لإشعار الاستبدال (vehicleNoted).
// كل وثيقة تبقى فعل مستقل بالسجل، حتى المتابعة تعرف أي وحدة انحلت.
func (s *MatrixAutopilotService) vehicleDocs(week string, dayStart time.Time) (int, error) {
	rows, err := s.actions.VehicleDocsExpiring(30)
	if err != nil {
		return 0, err
	}
	type group struct {
		name, plate string
		parts       []string
	}
	order := []string{}
	groups := map[string]*group{}
	n := 0
	for _, d := range rows {
		state := "تخلص"
		if d.ExpiryDate.Before(time.Now()) {
			state = "خالصة من"
		}
		part := fmt.Sprintf("%s %s %s", d.DocumentType, state, d.ExpiryDate.Format("2006-01-02"))
		if !s.act(model.AiAction{Kind: model.AiActionVehicleDocExpiry, EntityType: "VEHICLE_DOCUMENT", EntityID: d.ID,
			Period: week, TargetLabel: "الأسطول",
			Summary: fmt.Sprintf("نبّه الأسطول: %s «%s» %s %s", d.DocumentType, d.VehicleName, state, d.ExpiryDate.Format("2006-01-02")),
			Details: why(map[string]any{"document": d.DocumentType, "vehicle": d.VehicleName, "plate": d.PlateNumber, "expiryDate": d.ExpiryDate})}, dayStart,
			func() error { return nil }) {
			continue
		}
		n++
		g, ok := groups[d.PlateNumber]
		if !ok {
			g = &group{name: d.VehicleName, plate: d.PlateNumber}
			groups[d.PlateNumber] = g
			order = append(order, d.PlateNumber)
		}
		g.parts = append(g.parts, part)
	}
	for _, plate := range order {
		if s.vehicleNoted[plate] {
			continue // انذكرت ويا إشعار الاستبدال
		}
		g := groups[plate]
		msg := fmt.Sprintf("🤖 ماتركس — السيارة «%s» (%s): %s. جدّدوها قبل ما تطلع بالشارع.", g.name, g.plate, strings.Join(g.parts, "، "))
		if err := s.notif.CreateForPermission("vehicle_management", "AI_AUTOPILOT", msg); err != nil {
			log.Printf("[ai] تعذر إرسال وثائق السيارة %s: %v", plate, err)
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
	Level    string         `json:"level"` // CALM | ALERT | RED
	Mood     string         `json:"mood"`  // CALM | PLEASED | ALERT | ANGRY — شعور ماتركس
	Group    string         `json:"group"` // مجموعة الدور: تحدد شكل العين ولونها
	Open     int            `json:"open"`
	Items    []WatchItem    `json:"items"`
	Workload []WorkloadItem `json:"workload"`
	Name     string         `json:"name,omitempty"`
	ID       string         `json:"id,omitempty"`
}

// WorkloadItem «شغلك اليوم»: منجز وباقي، وشاشته.
type WorkloadItem struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Done  int    `json:"done"`
	Left  int    `json:"left"`
	Route string `json:"route"`
	Verb  string `json:"verb"` // الفعل بالكلام: «دققت»، «اعتمدت»…
	// Codes أكواد الحجوزات المعنية — العين تطلّعها أزرار تودّي للحجز نفسه.
	Codes []string `json:"codes,omitempty"`
	// Links لكل حجز رابط مباشر لشاشته (مثلاً فاتورته).
	Links []WorkLink `json:"links,omitempty"`
}

type WorkLink struct {
	Label string `json:"label"`
	To    string `json:"to"`
}

const watchRedOpen = 3

func (s *MatrixAutopilotService) WatchState(employeeID string) (*WatchState, error) {
	subj, err := s.actions.Subject(employeeID)
	if err != nil {
		return nil, err
	}
	return s.watchFor(*subj)
}

func (s *MatrixAutopilotService) watchFor(subj repository.WatchSubject) (*WatchState, error) {
	employeeID := subj.ID
	rows, err := s.actions.OpenForEmployee(employeeID)
	if err != nil {
		return nil, err
	}
	out := &WatchState{Level: "CALM", Items: []WatchItem{}, Group: WatchGroup(subj), Workload: s.workload(subj)}
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
	// الشعور: الغضب والانتباه من التذكيرات، والرضا لمن خلّص كل شغل اليوم.
	switch out.Level {
	case "RED":
		out.Mood = "ANGRY"
	case "ALERT":
		out.Mood = "ALERT"
	default:
		out.Mood = "CALM"
		done, left := 0, 0
		for _, w := range out.Workload {
			done += w.Done
			left += w.Left
		}
		if done > 0 && left == 0 {
			out.Mood = "PLEASED"
		}
	}
	return out, nil
}

func hasPerm(subj repository.WatchSubject, names ...string) bool {
	for _, p := range subj.Perms {
		for _, n := range names {
			if p == n {
				return true
			}
		}
	}
	return false
}

// WatchGroup مجموعة الموظف — كل مجموعة إلها عين بلونها وشكلها.
func WatchGroup(subj repository.WatchSubject) string {
	switch {
	case subj.Role == "OWNER" || subj.Role == "ADMIN":
		return "ADMINS"
	case subj.Role == "MONITOR":
		return "MONITORS"
	case subj.Role == "FINANCE":
		return "FINANCE"
	case subj.Role == "HR_COORDINATOR":
		return "COORDINATORS"
	case subj.Role == "QUALITY_ENGINEER" || hasPerm(subj, "quality_control"):
		return "QUALITY"
	case subj.Role == "IT_SUPPORT":
		return "IT"
	case subj.Role == "DESIGNER":
		return "DESIGN"
	case subj.IsLeader:
		return "LEADERS"
	case subj.Role == "TECHNICIAN" || subj.Role == "ENGINEER":
		// الفني العادي: جرد وحضور وأداء — ماكو فواتير ولا تقارير.
		return "TECHS"
	case hasPerm(subj, "coordinator"):
		return "COORDINATORS"
	case hasPerm(subj, "finance_audit", "auditing"):
		return "MONITORS"
	// عيون خاصة (چانوا يوقعون بـSTAFF وينقاسون بعمليات النظام بس) — بعد
	// صلاحيات الشغل: بائع عنده صلاحية التنسيق شغله تنسيق.
	case subj.Role == "SALES":
		return "SALES"
	case subj.Role == "PROJECT_MANAGER":
		return "PROJECTS"
	case subj.Role == "GPS_ADMIN" || subj.Role == "GPS_ENGINEER":
		return "GPS"
	default:
		return "STAFF"
	}
}

// workload «شغلك اليوم» حسب الدور والصلاحيات — أرقام حقيقية بس.
func (s *MatrixAutopilotService) workload(subj repository.WatchSubject) []WorkloadItem {
	out := []WorkloadItem{}
	r := s.actions
	if subj.Role == "MONITOR" || hasPerm(subj, "finance_audit", "auditing") {
		out = append(out, WorkloadItem{Key: "AUDIT", Label: "تدقيق حجوزات اليوم", Verb: "دققت",
			Done: r.AuditedByToday(subj.ID), Left: r.AuditLeftToday(), Route: "/daily-audit"})
	}
	if subj.Role == "MONITOR" {
		out = append(out, WorkloadItem{Key: "MONITOR_INBOX", Label: "صندوق المراقب", Verb: "راجعت",
			Done: r.ReviewedByToday(subj.ID), Left: r.MonitorPending(), Route: "/monitor-desk"})
	}
	if subj.Role == "HR_COORDINATOR" || hasPerm(subj, "coordinator") {
		tomorrow := time.Now().In(debriefLoc).AddDate(0, 0, 1).Format("2006-01-02")
		un, _ := r.UnstaffedOn(tomorrow)
		out = append(out,
			WorkloadItem{Key: "CONFIRM", Label: "حجوزات تنتظر التثبيت", Verb: "ثبّتت", Done: r.ConfirmedByToday(subj.ID), Left: r.BookingsPendingConfirm(), Route: "/coordinator"},
			WorkloadItem{Key: "STAFF", Label: "حجوزات باچر بلا كادر", Verb: "كلّفت", Left: len(un), Route: "/coordinator"})
	}
	if subj.Role == "FINANCE" || hasPerm(subj, "finance") {
		out = append(out, WorkloadItem{Key: "APPROVE", Label: "فواتير تنتظر الاعتماد", Verb: "اعتمدت",
			Done: r.ApprovedByToday(subj.ID), Left: r.InvoicesAwaitingApproval(), Route: "/leader-invoices"})
	}
	if subj.Role == "QUALITY_ENGINEER" || hasPerm(subj, "quality_control") {
		out = append(out, WorkloadItem{Key: "QUALITY", Label: "متابعات جودة معلّقة", Verb: "تابعت",
			Done: r.QualityDoneToday(subj.ID), Left: r.QualityPending(), Route: "/quality-follow-ups"})
	}
	// الليدر: حجوزاته، وتجهيز المواد، والورق (فاتورة وتقرير) — هو بس.
	if subj.IsLeader {
		total, done := r.LeaderToday(subj.ID)
		if total > 0 {
			out = append(out, WorkloadItem{Key: "JOBS", Label: "حجوزات اليوم", Verb: "أنجزت",
				Done: done, Left: total - done, Route: "/my-tasks"})
		}
		if codes := r.LeaderMaterialsPending(subj.ID); len(codes) > 0 {
			out = append(out, WorkloadItem{Key: "MATERIALS", Label: "حجوزات اليوم موادها ما تجهزت", Verb: "جهّزت",
				Left: len(codes), Route: "/my-tasks", Codes: codes})
		}
		if late, err := s.aiRepo.LatePaperworkRows(subj.ID); err == nil && len(late) > 0 {
			codes := make([]string, 0, len(late))
			links := make([]WorkLink, 0, len(late))
			for _, l := range late {
				codes = append(codes, l.BookingCode)
				what := "فاتورة"
				to := "/leader-invoices/new?mode=booking&bookingId=" + l.BookingID
				if !l.MissingInvoice && l.MissingReport {
					what, to = "تقرير", "/work-reports"
				}
				if len(links) < 12 {
					links = append(links, WorkLink{Label: l.BookingCode + " · " + what, To: to})
				}
			}
			out = append(out, WorkloadItem{Key: "PAPERWORK", Label: "حجوزات ناقصها فاتورة أو تقرير", Verb: "كمّلت",
				Left: len(late), Route: "/leader-invoices/new", Codes: codes, Links: links})
		}
	} else if subj.Role == "TECHNICIAN" || subj.Role == "ENGINEER" {
		// الفني العادي: حضور وانصراف، وحجوزاته، وعهدته — بلا ورق.
		in, outd := r.CheckedInToday(subj.ID)
		att := WorkloadItem{Key: "ATTEND", Label: "تسجيل الحضور اليوم", Verb: "سجّلت", Route: "/attendance"}
		if in {
			att.Done = 1
		} else {
			att.Left = 1
		}
		out = append(out, att)
		total, done := r.LeaderToday(subj.ID)
		if total > 0 {
			out = append(out, WorkloadItem{Key: "JOBS", Label: "حجوزات اليوم", Verb: "أنجزت",
				Done: done, Left: total - done, Route: "/my-tasks"})
		}
		if held := r.ToolsHeld(subj.ID); held > 0 {
			out = append(out, WorkloadItem{Key: "TOOLS", Label: "أدوات بعهدتك (جرد)", Verb: "رجّعت",
				Left: 0, Done: held, Route: "/my-inventory"})
		}
		_ = outd
	}
	// المبيعات والجي بي اس والدعم الفني — چانوا بلا «شغلك اليوم» أصلاً.
	if subj.Role == "SALES" {
		out = append(out, WorkloadItem{Key: "SALES", Label: "حجوزات سجّلتها اليوم", Verb: "سجّلت",
			Done: r.BookingsCreatedToday(subj.ID), Route: "/bookings"})
	}
	if subj.Role == "GPS_ADMIN" || subj.Role == "GPS_ENGINEER" {
		out = append(out, WorkloadItem{Key: "GPS_RENEW", Label: "اشتراكات تخلص خلال ١٤ يوم", Verb: "اتصلت",
			Done: r.GpsCallsToday(subj.ID), Left: r.GpsExpiringCount(), Route: "/gps/devices"})
	}
	if subj.Role == "IT_SUPPORT" {
		heavy := 0
		if rows, err := s.aiRepo.ItRepairCounts(30, 3); err == nil {
			heavy = len(rows)
		}
		out = append(out, WorkloadItem{Key: "IT", Label: "أجهزة تتصلّح بكثرة (٣+ بالشهر)", Verb: "سجّلت",
			Done: r.ItLogsToday(subj.ID), Left: heavy, Route: "/it-assets"})
	}
	if subj.Role == "DESIGNER" {
		out = append(out, WorkloadItem{Key: "DESIGN", Label: "أعمال تصميم رفعتها اليوم", Verb: "رفعت",
			Done: r.DesignUploadsToday(subj.ID), Route: "/design-gallery"})
	}
	return out
}

// RoleGroupReport تقرير عين مجموعة للمدير.
type RoleGroupReport struct {
	Group     string        `json:"group"`
	Employees []*WatchState `json:"employees"`
	Red       int           `json:"red"`
	Alert     int           `json:"alert"`
}

// RoleWatch كل المجموعات بموظفيها وعيونهم — ADMIN/OWNER.
func (s *MatrixAutopilotService) RoleWatch() ([]RoleGroupReport, error) {
	subs, err := s.actions.ActiveSubjects()
	if err != nil {
		return nil, err
	}
	order := []string{"MONITORS", "COORDINATORS", "FINANCE", "LEADERS", "TECHS", "SALES", "PROJECTS", "GPS", "DESIGN", "QUALITY", "IT", "ADMINS", "STAFF"}
	byGroup := map[string]*RoleGroupReport{}
	for _, g := range order {
		byGroup[g] = &RoleGroupReport{Group: g, Employees: []*WatchState{}}
	}
	for _, sub := range subs {
		w, err := s.watchFor(sub)
		if err != nil {
			continue
		}
		w.Name, w.ID = sub.Name, sub.ID
		g := byGroup[w.Group]
		g.Employees = append(g.Employees, w)
		switch w.Level {
		case "RED":
			g.Red++
		case "ALERT":
			g.Alert++
		}
	}
	out := []RoleGroupReport{}
	for _, g := range order {
		if len(byGroup[g].Employees) > 0 {
			out = append(out, *byGroup[g])
		}
	}
	return out, nil
}

// ReassignGpsToQuality تذكيرات التجديد المفتوحة على المالك/المدير تروح للجودة.
// تنادى بكل دورة وعند تشغيل الخادم — مو بس وقت تذكير جديد.
func (s *MatrixAutopilotService) ReassignGpsToQuality() string {
	qe := s.actions.QualityEngineerForGps()
	label := ""
	if qe != "" {
		label = s.actions.EmployeeName(qe)
	}
	s.actions.ReassignOpenGps(model.AiActionGpsExpiry, qe, label)
	return qe
}

// ١٣. الليدر خلّص حجز وما قيّم فنيّيه — طلب (ع) 10-05: «ماتركس يبقى يذكّره
// لحد ما يقيّم». مرة باليوم، وينحل لحاله أول ما يقيّم.
func (s *MatrixAutopilotService) crewRatings(today string, dayStart time.Time) (int, error) {
	rows, err := s.actions.LeadersPendingRating()
	if err != nil {
		return 0, err
	}
	n := 0
	for _, r := range rows {
		emp := r.LeaderID
		codes := strings.Join(r.Codes, "، ")
		msg := fmt.Sprintf("🤖 ماتركس — خلّصت الحجوزات (%s) وبعدك ما قيّمت الفنيين الي طلعوا وياك. قيّمهم من «مهامي» — تقييمك يبين بتقريرهم.", codes)
		if s.act(model.AiAction{Kind: model.AiActionCrewRating, EntityType: "EMPLOYEE", EntityID: emp,
			Period: today, TargetEmployeeID: &emp, TargetLabel: s.actions.EmployeeName(emp),
			Summary: fmt.Sprintf("ذكّر الليدر يقيّم فنيّيه (%d حجز)", len(r.Codes)),
			Details: why(map[string]any{"bookingCodes": []string(r.Codes)})}, dayStart,
			func() error { return s.notif.Create(emp, "AI_AUTOPILOT", msg) }) {
			n++
		}
	}
	return n, nil
}

// ١٤. الفني خلّص حجز ويا ليدر وما جرد عدّته بعده — طلب (ع) 10-05. مرة باليوم
// لحد ما يجرد، وينحل لحاله أول ما يجرد.
func (s *MatrixAutopilotService) afterInventory(today string, dayStart time.Time) (int, error) {
	rows, err := s.actions.TechsPendingAfterInventory()
	if err != nil {
		return 0, err
	}
	n := 0
	for _, r := range rows {
		emp := r.LeaderID
		codes := strings.Join(r.Codes, "، ")
		msg := fmt.Sprintf("🤖 ماتركس — خلّصت الحجوزات (%s) وبعدك ما جردت عدّتك بعدها. جردها من «مهامي» حتى إذا أكو شي ناقص ينعرف بوقته.", codes)
		if s.act(model.AiAction{Kind: model.AiActionAfterInventory, EntityType: "EMPLOYEE", EntityID: emp,
			Period: today, TargetEmployeeID: &emp, TargetLabel: s.actions.EmployeeName(emp),
			Summary: fmt.Sprintf("ذكّر الفني يجرد عدّته بعد الحجز (%d حجز)", len(r.Codes)),
			Details: why(map[string]any{"bookingCodes": []string(r.Codes)})}, dayStart,
			func() error { return s.notif.Create(emp, "AI_AUTOPILOT", msg) }) {
			n++
		}
	}
	return n, nil
}

// ١٥. ماتركس على المراقب — بنود بالصندوق تنتظر حكمه من أكثر من ٢٤ ساعة.
// مرة باليوم لكل مراقب، وينحل لحاله أول ما الصندوق يفرغ من المتأخر.
// (ع): «المدير راح يراقب المراقب من خلال ماتركس» — التذكير والنتيجة تطلع
// بتقرير المراقبين للمدير والمالك.
func (s *MatrixAutopilotService) monitorBacklog(today string, dayStart time.Time) (int, error) {
	b, err := s.actions.MonitorBacklog()
	if err != nil || b.Count == 0 {
		return 0, err
	}
	monitors, err := s.actions.Monitors()
	if err != nil {
		return 0, err
	}
	age := ""
	if b.Oldest != nil {
		age = fmt.Sprintf("، أقدمها صارله %s", FmtMinutes(int(time.Since(*b.Oldest).Minutes())))
	}
	n := 0
	for _, id := range monitors {
		emp := id
		msg := fmt.Sprintf("🤖 ماتركس — بصندوق المراقب %d بند ينتظر حكمك من أكثر من يوم%s. احكم عليهن حتى الشغل ما يوقف.", b.Count, age)
		if s.act(model.AiAction{Kind: model.AiActionMonitorBacklog, EntityType: "EMPLOYEE", EntityID: emp,
			Period: today, TargetEmployeeID: &emp, TargetLabel: s.actions.EmployeeName(emp),
			Summary: fmt.Sprintf("ذكّر المراقب: %d بند متأخر بالصندوق", b.Count),
			Details: why(map[string]any{"pending": b.Count, "oldest": b.Oldest})}, dayStart,
			func() error { return s.notif.Create(emp, "AI_AUTOPILOT", msg) }) {
			n++
		}
	}
	return n, nil
}
