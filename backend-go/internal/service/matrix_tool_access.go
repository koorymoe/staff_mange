package service

import (
	"fmt"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"

	"github.com/jmoiron/sqlx"
)

// ═══ 🔐 أدوات ماتركس حسب صلاحيات الموظف (دازاي 10-29) ═══
// «اريد اخلي بعض الموظفين يراسلون ماتركس… الصلاحيات تكون مقسمة بصورة صحيحة».
// صلاحية agent_matrix تفتح المحادثة، وكل أداة تنفتح بس إذا عنده صلاحية مجالها —
// الكود يضمن هذا مو البرومبت. المالك والمدير ياخذون الكل. أداة مو بالخريطة = ممنوعة.

var matrixToolAccess = map[string][]string{
	// للكل (ويا agent_matrix)
	"system_guide": nil, "lookup_code": nil, "booking_check": nil, "suggest_improvement": nil, "system_changelog": nil, "report_gap": nil, "my_performance": nil, "propose_lesson": nil, "my_stats": nil, "raise_manager_request": nil,
	// الهيكل الجديد ولائحة العمل (10-30) — للكل؛ unit_status (تفاصيل الإدارات) للمالك والمدير بس
	"org_structure": nil, "regulations": nil,
	// الحجوزات والميدان
	"late_bookings":    {"view_bookings", "coordinator", "crew_management", "monitoring", "mon_field", "mon_attendance"},
	"field_timing":     {"view_bookings", "coordinator", "crew_management", "monitoring", "mon_field"},
	"late_paperwork":   {"monitoring", "mon_reports", "auditing", "finance", "leader_invoices_view", "coordinator"},
	"attendance_today": {"monitoring", "mon_attendance", "staff_management", "coordinator"},
	// الزبائن والجودة
	"customer_profile": {"manage_customers", "quality_control", "complaints"},
	// المشاريع والمشتريات
	"project_delays":    {"project_management", "my_projects"},
	"procurement_watch": {"procurement", "inventory", "inventory_follow"},
	// الفلوس
	"money_month": {"finance", "finance_audit", "auditing"}, "invoices_status": {"finance", "finance_audit", "auditing", "leader_invoices_view"},
	"revenue_by_responsible": {"finance", "finance_audit", "auditing"}, "service_economics": {"finance", "finance_audit"},
	// الموظفين وتقييمهم
	"staff_scores": {"monitoring", "mon_staff", "staff_management", "kpi_management"}, "employee_profile": {"monitoring", "mon_staff", "staff_management"},
	"employee_dossier":   {"monitoring", "mon_staff", "staff_management"},
	"training_study":     {"monitoring", "mon_staff", "staff_management"},
	"daily_report_match": {"monitoring", "mon_staff", "staff_management"},
	"employee_signals":   {"monitoring", "mon_staff", "staff_management"}, "team_performance": {"monitoring", "mon_staff", "staff_management", "kpi_management"},
	// (ل) 11-10 نماذج عروض الأسعار — بصلاحيات العروض نفسها
	"quotation_template_save": {QuotationPermCreate, QuotationPermEditOwn, QuotationPermManageAll, "quotation_system"},
	"quotation_template_list": {QuotationPermCreate, QuotationPermEditOwn, QuotationPermManageAll, "quotation_system"},
	"quotation_from_template": {QuotationPermCreate, QuotationPermEditOwn, QuotationPermManageAll, "quotation_system"},
	// query_data / describe_data: مو بالخريطة = للمالك والمدير بس
}

// MatrixToolFilter يرجّع فلتر أدوات لموظف؛ nil للمالك والمدير (كلشي).
func MatrixToolFilter(role string, perms []string) func(string) bool {
	if model.IsAdminRole(role) {
		return nil
	}
	has := map[string]bool{}
	for _, p := range perms {
		has[p] = true
	}
	return func(name string) bool {
		need, ok := matrixToolAccess[name]
		if !ok {
			return false
		}
		if len(need) == 0 {
			return true
		}
		for _, p := range need {
			if has[p] {
				return true
			}
		}
		return false
	}
}

// MatrixSelfTools أدوات تخص المستخدم نفسه (my_stats).
func MatrixSelfTools(db *sqlx.DB, userID string) []agentTool {
	return []agentTool{{Name: "my_stats",
		Desc: "أرقامك إنت (الي يحاچيك): حضورك وتأخيرك وإنجازك آخر ١٤ يوم مقابل الي قبلها. لأي سؤال «شلون شغلي/أدائي/حضوري».",
		Run: func(map[string]any, *nameMask) (string, error) {
			rows, err := EmployeeMetrics(db, userID)
			if err != nil {
				return "", err
			}
			if len(rows) == 0 {
				return "ماكو أرقام إلك (أو حسابك مو موظف ميداني).", nil
			}
			m := rows[0]
			return fmt.Sprintf("%s\nإجازة %d يوم، وأطول دوام متواصل %d يوم.", metricsLine(m), m.Leave, m.Streak), nil
		}}}
}

// raiseManagerRequestTool (ع) 11-04: ماتركس يرفع طلب للمدير باسم الي يحاچيه —
// كتابة وحيدة ومحصورة: طلب يسجّل باسمه هو، والمدير يقرر. ما يلمس أي بيانات ثانية.
func raiseManagerRequestTool(db *sqlx.DB, userID string) agentTool {
	return agentTool{Name: "raise_manager_request",
		Desc: "يرفع طلب أو مهمة للمدير باسم الموظف الي يحاچيك (يوصل للمدير والمالك كإشعار، ويبقى مسجّل لحد ما ينحسم). استعملها **بس** لما الموظف يطلب صراحة يرفع شي للمدير، وأكّد وياه العنوان والتفاصيل قبل. attach_inventory=true يرفق أرقام المخازن الحقيقية (طلبات معلّقة، نواقص جرد، سيارات، صيانة).",
		Props: map[string]any{
			"title":            map[string]any{"type": "string", "description": "عنوان قصير"},
			"details":          map[string]any{"type": "string", "description": "شنو المطلوب بالضبط"},
			"priority":         map[string]any{"type": "string", "enum": []string{"NORMAL", "HIGH", "URGENT"}},
			"attach_inventory": map[string]any{"type": "boolean"},
		},
		Req: []string{"title", "details"},
		Run: func(in map[string]any, _ *nameMask) (string, error) {
			s := func(k string) string { v, _ := in[k].(string); return v }
			ctx := ""
			if b, _ := in["attach_inventory"].(bool); b {
				ctx = repository.InventoryContext(db)
			}
			if _, err := repository.NewManagerRequestRepository(db, repository.NewNotificationRepository(db)).
				Create(userID, s("title"), s("details"), s("priority"), ctx, "MATRIX"); err != nil {
				return "", err
			}
			return "انرفع الطلب للمدير وانبعثله إشعار. الموظف يتابعه من «طلباتي للمدير» بشاشة المخازن، ويوصله إشعار لما ينحسم.", nil
		}}
}

// StaffTools أدوات خاصة بهالمستخدم: أرقامه هو، وللموظف قراءة بيانات مجاله.
func (a *MatrixAgent) StaffTools(db *sqlx.DB, userID string, top bool, perms []string) []agentTool {
	out := append(MatrixSelfTools(db, userID), raiseManagerRequestTool(db, userID))
	out = append(out, quotationTemplateTools(db, userID, "")...)
	if !top {
		out = append(out, a.ScopedDataTools(MatrixDataScope(perms))...)
	}
	return out
}
