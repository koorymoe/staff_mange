// Package authz — 🛡️ محرك الصلاحيات الواحد (تعديلات بند ١، المرحلة ١).
//
// «منو يشوف شنو من البيانات» جان مكتوب بكل handler بطريقته، ومرات يختلف.
// هنا مكان واحد: جدول `SeeAll` (منو يشوف كل صفوف النوع) و`Scope` (شرط SQL
// للي ما يشوف الكل). المرحلة ١ تنقل القواعد الموجودة **كما هي** — نفس
// السلوك بالضبط — وأي تغيير بالقاعدة يصير هنا ويتطبّق بكل مكان.
package authz

import (
	"log"
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
	// ١ج (11-04) — حراس الفلوس الي جانت قوائم بـ main.go
	Credits                Resource = "credits"                  // الكريدت وتدقيق الحجوزات
	Funds                  Resource = "funds"                    // الدوار
	FundAmount             Resource = "fund_amount"              // مبلغ الدوار وتعبئته
	FundDischarge          Resource = "fund_discharge"           // إبراء ذمة الدوار
	ProjectPaymentsFinance Resource = "project_payments_finance" // تدقيق دفعات المشاريع
	// المرحلة ٤ (11-04) — قرارات كانت جوّا project_payment_handler
	ProjectContractValue    Resource = "project_contract_value"    // قيمة العقد + تسجيل دفعة لأي مشروع
	ProjectPaymentsOverview Resource = "project_payments_overview" // فلوس كل المشاريع
	ProjectPaymentsAudit    Resource = "project_payments_audit"    // سجل الدفعات وتصحيح المبلغ (المالك)
	SurveyPhotos            Resource = "survey_photos"             // صور الكشف
	AchievementVoice        Resource = "achievement_voice"         // فويس الإنجاز وتحليله
	EmployeeRecords         Resource = "employee_records"          // سجلات موظف (دوام، تقارير شهرية)
	Projects                Resource = "projects"                  // مشروع مو إلك
	Quotations              Resource = "quotations"                // عروض الأسعار — القاعدة بـ QuotationService (حلقة استيراد)، الرفض من هنا
	InternalBookings        Resource = "internal_bookings"         // الحجوزات داخل الشركة (كلها)
	InternalBookingCreate   Resource = "internal_booking_create"   // تسجيل حجز داخل الشركة
	ServiceInvoices         Resource = "service_invoices"          // فواتير خدمة جي بي اس/داش كام (الصلاحية حسب النوع بالمعالج)
)

// Rule منو يشوف كل صفوف النوع: دور فعلي (EffectiveRole — يشمل المنصب بالهيكل)
// أو صلاحية. الي ما ينطبق عليه يشوف نطاقه هو (Scope).
type Rule struct {
	Roles []string
	Perms []string
	// RawRole: يقارن الدور الأصلي بدل الفعلي (مثل ما جان بالكود القديم).
	RawRole bool
	// Guard: نفس منطق حرّاس المسارات بالضبط — الأدمن والمالك (الأصلي) يعبرون،
	// والدور يطابق الأصلي أو المراقب أو **أي** دور من المنصب (middleware.HasRole).
	Guard bool
	// OwnerOnly: المالك (الدور الأصلي) بس — ولا حتى مدير النظام.
	OwnerOnly bool
}

// SeeAllRules ⚠️ نُقلت كما هي من الـ handlers (11-04) — لا تغيّرها بلا قرار دازاي.
var SeeAllRules = map[Resource]Rule{
	// booking_handler.canSeeAllBookings — sales_booking منشالة عمداً (تسجيل مو قراءة)
	Bookings: {Roles: []string{"ADMIN", "OWNER", "HR_COORDINATOR", "MONITOR", "FINANCE", "PROJECT_MANAGER", "PROCUREMENT_ADMIN"},
		Perms: []string{"view_bookings", "coordinator", "crew_management", "mission_tracking", "procurement", "monitoring", "finance"}},
	// leader_invoice_handler.canReviewInvoices — دور بس
	// (11-04) + صلاحية finance: نفس المصاريف — صاحب صلاحية المالية يشوف فواتير الليدرية
	LeaderInvoices: {Roles: []string{"ADMIN", "OWNER", "FINANCE", "MONITOR"}, Perms: []string{"finance"}},
	// expense_handler.canSeeAll
	Expenses: {Roles: []string{"ADMIN", "OWNER", "FINANCE", "MONITOR"}, Perms: []string{"expenses_manage", "finance"}},
	// work_report_handler.canReviewReports (canSeeOperational + صلاحيات المراجعة)
	// (11-04) + المحاسب وصلاحية finance: يطابق التقرير ويا الفاتورة مثل باقي الفلوس
	WorkReports: {Roles: []string{"OWNER", "ADMIN", "HR_COORDINATOR", "MONITOR", "PROJECT_MANAGER", "SERVICE_MANAGER", "FINANCE"},
		Perms: []string{"monitoring", "quality_control", "auditing", "mon_reports", "finance"}},
	// staff_score_handler isAdminRole — الأدمن والمالك يشوفون صفّهم؛ الباقي ما يشوف نفسه باللوحة
	StaffScores: {Roles: []string{"ADMIN", "OWNER"}, RawRole: true},
	// ١ج: منقولة من main.go كما هي (requireVerifyBooking، requireFund، requireFundAmount، requireDischarge)
	Credits:       {Roles: []string{"ADMIN", "OWNER", "FINANCE", "MONITOR"}, Perms: []string{"finance", "finance_audit"}, Guard: true},
	Funds:         {Perms: []string{"revolving_fund"}, Guard: true},
	FundAmount:    {Roles: []string{"ADMIN", "OWNER"}, Perms: []string{"revolving_fund_amount"}, Guard: true},
	FundDischarge: {Roles: []string{"ADMIN", "OWNER", "FINANCE"}, Perms: []string{"fund_discharge"}, Guard: true},
	// project_payment_handler.isFinance (has: الأدمن الأصلي + HasRole + الصلاحيات)
	ProjectPaymentsFinance: {Roles: []string{"FINANCE"}, Perms: []string{"finance", "finance_audit"}, Guard: true},
	// المرحلة ٤: منقولة كما هي من has(...) بـ project_payment_handler
	ProjectContractValue:    {Roles: []string{"FINANCE", "PROJECT_MANAGER"}, Perms: []string{"finance", "finance_audit", "project_management"}, Guard: true},
	ProjectPaymentsOverview: {Roles: []string{"FINANCE", "PROJECT_MANAGER", "MONITOR"}, Perms: []string{"finance", "finance_audit", "project_management", "monitoring"}, Guard: true},
	ProjectPaymentsAudit:    {OwnerOnly: true},
	// booking_handler.ListInternal — قرار دازاي: غيرهم يشوف الي توجّه إله بس
	InternalBookings: {Roles: []string{"ADMIN", "OWNER", "FINANCE", "MONITOR", "HR_COORDINATOR"}},
	// booking_handler.Create (داخل الشركة)
	InternalBookingCreate: {Roles: []string{"ADMIN", "OWNER", "HR_COORDINATOR", "MONITOR"}, Perms: []string{"booking_internal"}},
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
	if SeeAllRules[res].OwnerOnly {
		return raw == "OWNER"
	}
	if rule := SeeAllRules[res]; rule.Guard {
		if model.IsAdminRole(raw) || (len(rule.Roles) > 0 && middleware.HasRole(r, rule.Roles...)) {
			return true
		}
	} else if Decide(res, eff, raw, nil) {
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

// Guard حارس المسار مبني من نفس القاعدة — نفس الدالة القديمة بالضبط حسب شكل القاعدة
// (صلاحية وحدة بلا أدوار ← Permission؛ صلاحية وحدة ← RoleOrPermission؛ غيرها ← RoleOrAnyPermission).
func Guard(g *middleware.Guards, res Resource) func(http.Handler) http.Handler {
	rule := SeeAllRules[res]
	switch {
	case len(rule.Roles) == 0 && len(rule.Perms) == 1:
		return g.Permission(rule.Perms[0])
	case len(rule.Perms) == 1:
		return g.RoleOrPermission(rule.Roles, rule.Perms[0])
	default:
		return g.RoleOrAnyPermission(rule.Roles, rule.Perms...)
	}
}

// Capabilities جدول «شنو تگدر تشوف» للواجهة (/auth/me → can) — المرحلة ٢.
func Capabilities(r *http.Request, perms PermLister) map[Resource]bool {
	out := make(map[Resource]bool, len(SeeAllRules))
	var names []string
	loaded := false
	for res := range SeeAllRules {
		rule := SeeAllRules[res]
		eff, raw := middleware.EffectiveRole(r), middleware.RoleFromContext(r)
		ok := false
		if rule.OwnerOnly {
			out[res] = raw == "OWNER"
			continue
		}
		if rule.Guard {
			ok = model.IsAdminRole(raw) || (len(rule.Roles) > 0 && middleware.HasRole(r, rule.Roles...))
		} else {
			ok = Decide(res, eff, raw, nil)
		}
		if !ok && len(rule.Perms) > 0 && perms != nil {
			if !loaded {
				names, _ = perms.Names(middleware.EmployeeIDFromContext(r))
				loaded = true
			}
			ok = Decide(res, "", "", names) // الصلاحيات بس
		}
		out[res] = ok
	}
	return out
}

// Can (المرحلة ٤) يگدر يسوي العملية؟ نفس جدول القواعد. owns: ملكية الشي نفسه
// (مثلاً مشرف المشروع) تفتح لو القاعدة ما تفتح.
func Can(r *http.Request, perms PermLister, res Resource, owns ...bool) bool {
	for _, o := range owns {
		if o {
			return true
		}
	}
	return SeeAll(r, perms, res)
}

// Deny (المبدأ ٦و) الرفض يرجّع 404 — ما نكشف إن الشي موجود — ويتسجّل بالسجل.
func Deny(w http.ResponseWriter, r *http.Request, res Resource) {
	log.Printf("[authz] deny emp=%s role=%s %s %s res=%s", middleware.EmployeeIDFromContext(r), middleware.RoleFromContext(r), r.Method, r.URL.Path, res)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte(`{"error":"مو موجود"}`))
}
