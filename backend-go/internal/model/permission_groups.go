package model

// ═══ مجموعات الصلاحيات (دازاي 10-25: «الصلاحيات لازم تتقسم حسب كل دور») ═══
// كل صلاحية بمجموعة وحدة، والمجموعة الي إلها Bundle تنمنح كاملة بضغطة. أي صلاحية
// مو هنا تطلع تحت «أخرى» — حتى صلاحية جديدة ما تختفي من الشاشة.

type PermissionGroup struct {
	Key    string   `json:"key"`
	Title  string   `json:"title"`
	Bundle string   `json:"bundle,omitempty"` // الصلاحية الي تعني «الحزمة كاملة» (إن وجدت)
	Perms  []string `json:"perms"`
}

var PermissionGroups = []PermissionGroup{
	{Key: "monitor", Title: "👁️ المراقب", Bundle: "monitoring",
		Perms: append([]string{"monitoring"}, MonitorSubPermissions...)},
	{Key: "quality", Title: "⭐ الجودة والشكاوى",
		Perms: []string{"quality_control", "complaints", "complaint_contact", "performance_review"}},
	{Key: "audit", Title: "🔍 التدقيق", Bundle: "auditing",
		Perms: []string{"auditing", "audit_issues", "duplicate_review", "leader_invoices_view", "crew_management"}},
	{Key: "finance", Title: "💵 الحسابات",
		Perms: []string{"finance", "finance_audit", "expenses", "expenses_manage", "revolving_fund", "revolving_fund_amount", "fund_discharge",
			"gps_install_costs", "invoice_internal", "invoice_manual", "invoice_gps", "invoice_dashcam", "execution_cost", "booking_partial_link"}},
	{Key: "staff", Title: "👥 الموظفين",
		Perms: []string{"staff_management", "employee_create", "edit_employee_profile", "edit_employee_identity", "employee_edit_all", "employee_suspend",
			"employee_stats", "staff_stats", "staff_requests", "extra_tasks_assign", "training_manage", "kpi_management", "kpi_criteria_management",
			"work_schedule_manage", "remote_hours_manage", "remote_hours_self", "leave_approve_morning", "leave_approve_evening", "departments_manage"}},
	{Key: "bookings", Title: "📋 الحجوزات والزبائن",
		Perms: []string{"sales_booking", "create_booking", "view_bookings", "coordinator", "manage_customers", "bookings_archive", "booking_internal",
			"booking_reopen", "booking_delete_request", "booking_delete_approve", "mission_tracking", "paperwork_carryover", "vip_manual_add"}},
	{Key: "projects", Title: "🏗️ المشاريع وعروض الأسعار",
		Perms: []string{"project_management", "project_create_only", "my_projects", "quotation_create", "quotation_edit_own", "quotation_manage_all",
			"quotation_system", "network_prices", "solar_system", "gps_system", "gps_requests", "content_technician", "service_managers", "manage_services"}},
	{Key: "stores", Title: "📦 المخزن والمشتريات",
		Perms: []string{"inventory", "inventory_follow", "tool_requests", "tool_requests_approve", "procurement", "procurement_personal", "procurement_customer",
			"procurement_manual", "suppliers_management", "supplier_relations", "tech_suppliers", "leader_basket", "device_maintenance"}},
	{Key: "vehicles", Title: "🚗 المركبات", Perms: []string{"vehicle_management"}},
	{Key: "agents", Title: "🤖 الوكلاء", Perms: []string{"agent_matrix", "agent_quality", "agent_booking"}},
	{Key: "units", Title: "🏢 الوحدات والأقسام",
		Perms: []string{"unit_technicians", "unit_service", "unit_design", "unit_pr", "unit_quality", "unit_monitoring", "unit_procurement", "unit_finance",
			"unit_hr", "unit_projects", "unit_it", "design_forms", "design_gallery", "media", "it_assets", "it_stats", "privacy_policy_manage", "work_reports"}},
}

// PermissionOwnerOnly صلاحيات «مفتاح المالك» — المالك يمنحها بيده، والمدير يشوفها مقفولة ويا السبب.
var PermissionOwnerOnly = map[string]string{
	"finance_audit":   "مفتاح التدقيق المالي — المالك يمنحه بيده",
	"employee_create": "إضافة موظفين — قرار المالك",
	"employee_suspend": "إيقاف الحسابات — قرار المالك",
}
