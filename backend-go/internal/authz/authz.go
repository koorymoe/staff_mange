// Package authz — 🛡️ محرك الصلاحيات الواحد (تعديلات بند ١، المرحلة ١).
//
// «منو يشوف شنو من البيانات» جان مكتوب بكل handler بطريقته، ومرات يختلف.
// هنا مكان واحد: جدول `SeeAll` (منو يشوف كل صفوف النوع) و`Scope` (شرط SQL
// للي ما يشوف الكل). المرحلة ١ تنقل القواعد الموجودة **كما هي** — نفس
// السلوك بالضبط — وأي تغيير بالقاعدة يصير هنا ويتطبّق بكل مكان.
package authz

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
)

// Resource نوع البيانات.
type Resource string

const (
	Bookings       Resource = "bookings"        // الحجوزات (القائمة، البحث، العدّادات، الخط الزمني)
	LeaderInvoices Resource = "leader_invoices" // فواتير الليدرية
	Expenses       Resource = "expenses"        // المصاريف
	WorkReports    Resource = "work_reports"    // تقارير العمل
	StaffScores    Resource = "staff_scores"    // تقييم الموظفين (اللوحة والتفصيل)
)

// Rule منو يشوف كل صفوف النوع: دور فعلي (EffectiveRole — يشمل المنصب بالهيكل)
// أو صلاحية. الي ما ينطبق عليه يشوف نطاقه هو (Scope).
type Rule struct {
	Roles []string
	Perms []string
	// RawRole: يقارن الدور الأصلي بدل الفعلي (مثل ما جان بالكود القديم).
	RawRole bool
}

// SeeAllRules ⚠️ نُقلت كما هي من الـ handlers (11-04) — لا تغيّرها بلا قرار دازاي.
var SeeAllRules = map[Resource]Rule{
	// booking_handler.canSeeAllBookings — sales_booking منشالة عمداً (تسجيل مو قراءة)
	Bookings: {Roles: []string{"ADMIN", "OWNER", "HR_COORDINATOR", "MONITOR", "FINANCE", "PROJECT_MANAGER", "PROCUREMENT_ADMIN"},
		Perms: []string{"view_bookings", "coordinator", "crew_management", "mission_tracking", "procurement", "monitoring", "finance"}},
	// leader_invoice_handler.canReviewInvoices — دور بس
	LeaderInvoices: {Roles: []string{"ADMIN", "OWNER", "FINANCE", "MONITOR"}},
	// expense_handler.canSeeAll
	Expenses: {Roles: []string{"ADMIN", "OWNER", "FINANCE", "MONITOR"}, Perms: []string{"expenses_manage", "finance"}},
	// work_report_handler.canReviewReports (canSeeOperational + صلاحيات المراجعة)
	WorkReports: {Roles: []string{"OWNER", "ADMIN", "HR_COORDINATOR", "MONITOR", "PROJECT_MANAGER", "SERVICE_MANAGER"},
		Perms: []string{"monitoring", "quality_control", "auditing", "mon_reports"}},
	// staff_score_handler isAdminRole — الأدمن والمالك يشوفون صفّهم؛ الباقي ما يشوف نفسه باللوحة
	StaffScores: {Roles: []string{"ADMIN", "OWNER"}, RawRole: true},
}

// PermLister مصدر صلاحيات الموظف (PermissionRepository).
type PermLister interface {
	Names(employeeID string) ([]string, error)
}

// Decide القرار الصافي (قابل للاختبار): الدور (الفعلي أو الأصلي حسب القاعدة) + الصلاحيات.
func Decide(res Resource, effectiveRole, rawRole string, perms []string) bool {
	rule, ok := SeeAllRules[res]
	if !ok {
		return false // نوع مو معرّف = يضيق، ما يوسّع
	}
	role := effectiveRole
	if rule.RawRole {
		role = rawRole
	}
	for _, r := range rule.Roles {
		if r == role {
			return true
		}
	}
	for _, p := range perms {
		for _, want := range rule.Perms {
			if p == want {
				return true
			}
		}
	}
	return false
}

// SeeAll يشوف كل صفوف النوع؟ خطأ بقراءة الصلاحيات = لا (يضيّق ما يوسّع).
func SeeAll(r *http.Request, perms PermLister, res Resource) bool {
	eff, raw := middleware.EffectiveRole(r), middleware.RoleFromContext(r)
	if Decide(res, eff, raw, nil) {
		return true
	}
	if perms == nil || len(SeeAllRules[res].Perms) == 0 {
		return false
	}
	names, err := perms.Names(middleware.EmployeeIDFromContext(r))
	if err != nil {
		return false
	}
	return Decide(res, eff, raw, names)
}

// IsAdmin مدير النظام أو المالك (الدور الأصلي).
func IsAdmin(r *http.Request) bool { return model.IsAdminRole(middleware.RoleFromContext(r)) }
