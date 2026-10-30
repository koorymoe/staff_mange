package model

import "time"

// ═══ 🏛️ الهيكل الجديد (دازاي 10-29) — مدراء ووحدات ونواب ═══
// «الشغل الداخلي (الحجز والفاتورة والتقرير) ما نلمسه — يتغير الهيكل الخارجي».
// طبقة فوك الأدوار: الدور القديم يبقى (ألف فحص يعتمد عليه)، والمنصب بالوحدة هو
// الي ينطي الصلاحيات (مصدر ORG). الخطة: /root/.claude/plans (الهيكل الجديد).
//
//   مدير الشركة (المالك) + مدير النظام  ← فوك الكل (ما يحتاجون تعيين)
//        └ مدير الإدارة                  ← يراقب كل الوحدات + الحسابات + الرقابة
//             └ العمليات · تطوير الأعمال · الهندسية · الجودة · سلاسل التوريد · العقود والإعلام
//   IT وحدة مستقلة (قرار دازاي: «IT يضل وحده»)

const (
	OrgDirector = "DIRECTOR"
	OrgDeputy   = "DEPUTY"
	OrgMember   = "MEMBER"
)

// OrgDivision إدارة (يرأسها مدير ونائب).
type OrgDivision struct {
	ID       string  `db:"id" json:"id"`
	Name     string  `db:"name" json:"name"`
	ParentID *string `db:"parentId" json:"parentId"`
	Sort     int     `db:"sort" json:"sort"`
	Summary  string  `db:"summary" json:"summary"`
}

// OrgUnit وحدة تحت إدارة — permissions حزمة الصلاحيات الي ينطيها المنصب بيها.
type OrgUnit struct {
	ID          string   `db:"id" json:"id"`
	DivisionID  string   `db:"divisionId" json:"divisionId"`
	Name        string   `db:"name" json:"name"`
	Sort        int      `db:"sort" json:"sort"`
	Permissions []string `db:"-" json:"permissions"`
	// HeadPermissions الجزء من Permissions الي لرئيس الوحدة/نائبها بس
	HeadPermissions []string `db:"-" json:"headPermissions"`
}

// OrgAssignment منصب موظف. ما ينحذف أبداً — ينتهي (endedAt) وينحفظ السجل حتى التراجع يرجّعه.
type OrgAssignment struct {
	ID           string     `db:"id" json:"id"`
	EmployeeID   string     `db:"employeeId" json:"employeeId"`
	EmployeeName string     `db:"employeeName" json:"employeeName"`
	DivisionID   string     `db:"divisionId" json:"divisionId"`
	UnitID       *string    `db:"unitId" json:"unitId"`
	Position     string     `db:"position" json:"position"`
	AllowanceIQD float64    `db:"allowanceIQD" json:"allowanceIQD"`
	Acting       bool       `db:"acting" json:"acting"`
	ActingUntil  *time.Time `db:"actingUntil" json:"actingUntil"`
	ReplacesID   *string    `db:"replacesId" json:"replacesId"`
	StartedAt    time.Time  `db:"startedAt" json:"startedAt"`
	EndedAt      *time.Time `db:"endedAt" json:"endedAt"`
	EndedByID    *string    `db:"endedById" json:"endedById"`
	EndReason    *string    `db:"endReason" json:"endReason"`
	Note         *string    `db:"note" json:"note"`
	// أسماء المكان — تنملي بـ/api/org/me بس (الموظف ما يگدر يطلب الشجرة)
	DivisionName string  `db:"divisionName" json:"divisionName,omitempty"`
	UnitName     *string `db:"unitName" json:"unitName,omitempty"`
}

type orgSeedUnit struct {
	ID, Name string
	Perms    []string
	// Head من Perms: للرئيس ونائبه بس (الموافقات والإدارة) — العضو ياخذ الباقي (دازاي 10-30)
	Head []string
}

type orgSeedDivision struct {
	ID, Name, Parent, Summary string
	// Perms صلاحيات منصب الإدارة نفسه (فوك حزم وحداتها) — مثلاً مراقبة الكل لمدير الإدارة
	Perms []string
	// HeadAll: حزمة منصب الإدارة كلها للمدير والنائب بس
	HeadAll bool
	Units   []orgSeedUnit
}

// OrgSeed الهيكل الي شرحه دازاي (10-29). الحزم من خريطة الصلاحيات الموجودة؛ مفاتيح
// المالك (finance_audit، employee_create، employee_suspend) ما تدخل أي حزمة — يمنحها بيده.
var OrgSeed = []orgSeedDivision{
	{ID: "admin", Name: "الإدارة", Summary: "مراقبة شغل كل الوحدات وتقاريرهم، والحسابات، والرقابة",
		Perms:   []string{"monitoring", "mon_inbox", "mon_reports", "mon_justifications", "mon_attendance", "mon_staff", "mon_matrix", "mon_field", "agent_matrix", "unit_monitoring", "unit_matrix"},
		HeadAll: true,
		Units: []orgSeedUnit{
			{ID: "admin.monitoring", Name: "مراقبة الوحدات", Perms: []string{"unit_monitoring", "unit_matrix", "monitoring", "mon_inbox", "mon_reports", "mon_justifications", "mon_attendance", "mon_staff", "mon_matrix", "mon_field", "duplicate_review"}, Head: []string{"mon_justifications", "mon_staff", "mon_matrix", "duplicate_review"}},
			{ID: "admin.accounts", Name: "الحسابات", Perms: []string{"unit_finance", "unit_accounts", "finance", "expenses", "expenses_manage", "revolving_fund", "revolving_fund_amount", "fund_discharge", "invoice_internal", "invoice_manual", "execution_cost", "booking_partial_link", "leader_invoices_view"}, Head: []string{"expenses_manage", "revolving_fund", "revolving_fund_amount", "fund_discharge", "invoice_manual", "booking_partial_link", "finance"}},
			{ID: "admin.oversight", Name: "الرقابة", Perms: []string{"unit_monitoring", "audit_issues", "privacy_policy_manage"}, Head: []string{"privacy_policy_manage"}},
		}},
	{ID: "ops", Name: "العمليات", Parent: "admin", Summary: "التدقيق، والعُدد والمركبات، وإدارة الكوادر، والحجوزات والتنسيق",
		Units: []orgSeedUnit{
			{ID: "ops.audit", Name: "التدقيق", Perms: []string{"auditing", "audit_issues", "duplicate_review", "leader_invoices_view", "crew_management"}, Head: []string{"audit_issues", "duplicate_review", "crew_management"}},
			{ID: "ops.fleet", Name: "العُدد والمركبات", Perms: []string{"unit_fleet", "vehicle_management", "tool_requests", "tool_requests_approve", "device_maintenance"}, Head: []string{"vehicle_management", "tool_requests_approve"}},
			{ID: "ops.staff", Name: "إدارة الكوادر", Perms: []string{"unit_hr", "staff_management", "edit_employee_profile", "work_schedule_manage", "leave_approve_morning", "leave_approve_evening", "staff_requests", "extra_tasks_assign", "kpi_management", "performance_review", "employee_stats", "staff_stats", "remote_hours_manage"}, Head: []string{"staff_management", "edit_employee_profile", "work_schedule_manage", "leave_approve_morning", "leave_approve_evening", "extra_tasks_assign", "kpi_management", "performance_review", "remote_hours_manage"}},
			{ID: "ops.bookings", Name: "الحجوزات والتنسيق", Perms: []string{"sales_booking", "create_booking", "view_bookings", "coordinator", "manage_customers", "bookings_archive", "booking_internal", "mission_tracking", "paperwork_carryover"}, Head: []string{"manage_customers", "bookings_archive", "sales_booking"}},
		}},
	{ID: "bizdev", Name: "تطوير الأعمال", Parent: "admin", Summary: "وحدة التقنيين (ومعها الخدمات: GPS والشمسية والداش كام)، والتدريب",
		Units: []orgSeedUnit{
			{ID: "bizdev.tech", Name: "التقنيين والخدمات", Perms: []string{"unit_technicians", "unit_service", "content_technician", "tech_suppliers", "supplier_relations", "gps_system", "gps_requests", "solar_system", "invoice_gps", "invoice_dashcam", "service_managers", "network_prices", "gps_install_costs", "work_reports"}, Head: []string{"tech_suppliers", "supplier_relations", "invoice_gps", "invoice_dashcam", "service_managers", "network_prices", "gps_install_costs"}},
			{ID: "bizdev.training", Name: "التدريب", Perms: []string{"unit_training", "training_manage"}, Head: []string{"training_manage"}},
		}},
	{ID: "eng", Name: "الهندسية", Parent: "admin", Summary: "إدارة المشاريع والتصميم",
		Units: []orgSeedUnit{
			{ID: "eng.projects", Name: "إدارة المشاريع", Perms: []string{"unit_projects", "project_management", "my_projects", "quotation_create", "quotation_edit_own", "quotation_manage_all", "quotation_system"}, Head: []string{"project_management", "quotation_manage_all", "quotation_system"}},
			{ID: "eng.design", Name: "التصميم", Perms: []string{"unit_design", "design_forms", "design_gallery"}},
		}},
	{ID: "quality", Name: "الجودة", Parent: "admin", Summary: "متابعة الجودة والشكاوي",
		Units: []orgSeedUnit{
			{ID: "quality.followup", Name: "متابعة الجودة والشكاوي", Perms: []string{"unit_quality", "quality_control", "complaints", "complaint_contact", "agent_quality"}, Head: []string{"quality_control", "agent_quality"}},
		}},
	{ID: "supply", Name: "سلاسل التوريد", Parent: "admin", Summary: "المشتريات والمخزن",
		Units: []orgSeedUnit{
			{ID: "supply.procurement", Name: "المشتريات والمخزن", Perms: []string{"unit_procurement", "procurement", "procurement_personal", "procurement_customer", "procurement_manual", "suppliers_management", "leader_basket", "inventory", "inventory_follow"}, Head: []string{"procurement_manual", "suppliers_management", "inventory", "procurement"}},
		}},
	{ID: "contracts", Name: "العقود والشؤون التجارية والإعلامية", Parent: "admin", Summary: "الإعلام والعلاقات العامة",
		Units: []orgSeedUnit{
			{ID: "contracts.media", Name: "الإعلام والعلاقات العامة", Perms: []string{"unit_pr", "media", "vip_manual_add", "agent_booking"}, Head: []string{"vip_manual_add", "agent_booking"}},
		}},
	{ID: "it", Name: "التقنية IT", Summary: "وحدة مستقلة (قرار دازاي)",
		Units: []orgSeedUnit{
			{ID: "it.main", Name: "التقنية IT", Perms: []string{"unit_it", "it_assets", "it_stats"}, Head: []string{"it_assets"}},
		}},
}
