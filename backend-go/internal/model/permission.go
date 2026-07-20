package model

type Permission struct {
	ID    string `db:"id" json:"id"`
	Name  string `db:"name" json:"name"`
	Label string `db:"label" json:"label"`
}

type SetPermissionsRequest struct {
	PermissionIDs []string `json:"permissionIds"`
}

// DefaultPermissions هي القائمة الثابتة لكل الصلاحيات المتاحة بالنظام
var DefaultPermissions = []Permission{
	{Name: "staff_management", Label: "إدارة الكوادر"},
	{Name: "edit_employee_profile", Label: "تعديل ملف الموظف (الراتب/الدوام/الإجازات)"},
	{Name: "kpi_management", Label: "تقييم الأداء (KPI)"},
	{Name: "inventory", Label: "جرد الأدوات"},
	{Name: "complaints", Label: "الشكاوى"},
	{Name: "sales_booking", Label: "إنشاء حجز جديد"},
	{Name: "manage_customers", Label: "إدارة العملاء"},
	{Name: "view_bookings", Label: "عرض الحجوزات"},
	{Name: "coordinator", Label: "تنسيق الحجوزات"},
	{Name: "manage_services", Label: "الخدمات"},
	{Name: "mission_tracking", Label: "تتبع المهام"},
	{Name: "gps_system", Label: "نظام GPS"},
	{Name: "project_management", Label: "إدارة المشاريع"},
	{Name: "quotation_system", Label: "نظام عروض الأسعار"},
	{Name: "finance", Label: "المالية"},
	{Name: "expenses", Label: "المصاريف"},
	{Name: "procurement", Label: "المشتريات"},
	{Name: "monitoring", Label: "مراقبة (متابعة المهام والحجوزات والموظفين)"},
	{Name: "auditing", Label: "تدقيق (التحقق من جودة العمل والتقارير والحسابات)"},
	{Name: "content_technician", Label: "صلاحية التقني (إدارة المحتوى التدريبي والخدمات والموردين والمواد)"},
	{Name: "vehicle_management", Label: "إدارة المركبات"},
	{Name: "quality_control", Label: "الجودة (متابعة مشاكل التنفيذ والرقابة)"},
}

// RoleDefaultPermissions هي الصلاحيات الافتراضية لكل دور وظيفي
var RoleDefaultPermissions = map[string][]string{
	"ADMIN":            {},
	"SALES":            {"sales_booking", "complaints"},
	"HR_COORDINATOR":   {"staff_management", "edit_employee_profile", "coordinator", "manage_customers", "view_bookings", "manage_services", "inventory", "complaints", "mission_tracking", "sales_booking"},
	"TECHNICIAN":       {"expenses"},
	"PROJECT_MANAGER":  {"project_management", "expenses", "mission_tracking"},
	"MONITOR":          {"staff_management", "kpi_management", "view_bookings", "manage_customers", "manage_services", "mission_tracking", "inventory", "complaints", "finance", "monitoring", "auditing", "quality_control"},
	"FINANCE":          {"finance", "view_bookings"},
	"GPS_ADMIN":        {"gps_system"},
	"QUALITY_ENGINEER": {"auditing", "complaints", "quality_control", "sales_booking"},
	"ENGINEER":         {"expenses", "quotation_system", "project_management"},
}
