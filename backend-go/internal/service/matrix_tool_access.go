package service

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// ═══ 🔐 أدوات ماتركس حسب صلاحيات الموظف (دازاي 10-29) ═══
// «اريد اخلي بعض الموظفين يراسلون ماتركس… الصلاحيات تكون مقسمة بصورة صحيحة».
// صلاحية agent_matrix تفتح المحادثة، وكل أداة تنفتح بس إذا عنده صلاحية مجالها —
// الكود يضمن هذا مو البرومبت. المالك والمدير ياخذون الكل. أداة مو بالخريطة = ممنوعة.

var matrixToolAccess = map[string][]string{
	// للكل (ويا agent_matrix)
	"system_guide": nil, "lookup_code": nil, "booking_check": nil, "suggest_improvement": nil, "my_stats": nil,
	// الحجوزات والميدان
	"late_bookings":    {"view_bookings", "coordinator", "crew_management", "monitoring", "mon_field", "mon_attendance"},
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
	"employee_signals": {"monitoring", "mon_staff", "staff_management"}, "team_performance": {"monitoring", "mon_staff", "staff_management", "kpi_management"},
	// query_data / describe_data: مو بالخريطة = للمالك والمدير بس
}

// MatrixToolFilter يرجّع فلتر أدوات لموظف؛ nil للمالك والمدير (كلشي).
func MatrixToolFilter(role string, perms []string) func(string) bool {
	if role == "OWNER" || role == "ADMIN" {
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

// StaffTools أدوات خاصة بهالمستخدم: أرقامه هو، وللموظف قراءة بيانات مجاله.
func (a *MatrixAgent) StaffTools(db *sqlx.DB, userID string, top bool, perms []string) []agentTool {
	out := MatrixSelfTools(db, userID)
	if !top {
		out = append(out, a.ScopedDataTools(MatrixDataScope(perms))...)
	}
	return out
}
