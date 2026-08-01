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
	{Name: "kpi_criteria_management", Label: "إدارة نقاط الكي بي اي (إضافة/حذف)"},
	{Name: "inventory", Label: "جرد الأدوات"},
	// موافقة/رفض طلبات الأدوات انفصلت بصلاحية مستقلة. قبل، كانت مربوطة بالدور
	// الوظيفي فقط (ADMIN/HR_COORDINATOR/MONITOR) بدون أي منفذ صلاحية، فإداري
	// الكميات — وهو صاحب الشغلة أصلاً — ما كان يقدر يوافق حتى لو انمنحت له
	// صلاحية "جرد الأدوات". هسه تنمنح لوحدها لأي موظف.
	{Name: "tool_requests_approve", Label: "موافقة/رفض طلبات الأدوات"},
	{Name: "complaints", Label: "الشكاوى"},
	{Name: "sales_booking", Label: "إنشاء حجز جديد"},
	{Name: "manage_customers", Label: "إدارة العملاء"},
	{Name: "view_bookings", Label: "عرض الحجوزات"},
	{Name: "coordinator", Label: "تنسيق الحجوزات"},
	{Name: "manage_services", Label: "الخدمات"},
	{Name: "mission_tracking", Label: "تتبع المهام"},
	{Name: "gps_system", Label: "نظام GPS"},
	{Name: "project_management", Label: "إدارة المشاريع (كاملة)"},
	// صلاحية مبسّطة: يضيف مشروع وتفاصيله فقط — ما يشوف الإحصائيات ولا التقارير
	// ولا استمارة الكشف ولا يقدر يرحّل مراحل. تُمنح لأي موظف (حتى ليدر).
	{Name: "project_create_only", Label: "إضافة مشروع فقط (بدون إحصائيات ولا تقارير)"},
	// عروض الأسعار: ثلاث درجات متدرجة (يُمنح واحدة منهن للموظف عادةً، مو أكثر
	// من وحدة مع بعض) — "quotation_system" القديمة تبقى تشتغل بمفعول
	// quotation_manage_all لأي موظف كانت ممنوحة له سابقاً، بس ما تظهر هنا
	// كخيار جديد حتى ما تختلط الدرجات على المدير.
	{Name: "quotation_create", Label: "عروض الأسعار: إضافة فقط (بدون اطلاع على العروض القديمة)"},
	{Name: "quotation_edit_own", Label: "عروض الأسعار: إضافة وتعديل (عروضي فقط)"},
	{Name: "quotation_manage_all", Label: "عروض الأسعار: إضافة وتعديل واطلاع (كل العروض)"},
	{Name: "finance", Label: "المالية"},
	{Name: "expenses", Label: "المصاريف"},
	{Name: "procurement", Label: "المشتريات"},
	{Name: "procurement_personal", Label: "طلب احتياجات شخصية"},
	{Name: "procurement_customer", Label: "طلب منتج للزبون"},
	{Name: "monitoring", Label: "مراقبة (متابعة المهام والحجوزات والموظفين)"},
	{Name: "auditing", Label: "تدقيق (التحقق من جودة العمل والتقارير والحسابات)"},
	{Name: "content_technician", Label: "صلاحية التقني (إدارة المحتوى التدريبي والخدمات والمواد)"},
	// الموردون انفصلوا بصلاحية مستقلة — كانوا مربوطين بـcontent_technician
	// الواسعة، فمنح موظف حق إضافة مورد كان يفتحله التدريب والخدمات والمنتجات
	// كلها معاه. هسه صلاحية وحدها تُمنح لوحدها.
	{Name: "suppliers_management", Label: "إدارة الموردين (إضافة وتعديل)"},
	{Name: "vehicle_management", Label: "إدارة المركبات"},
	{Name: "quality_control", Label: "الجودة (متابعة مشاكل التنفيذ والرقابة)"},
	{Name: "leader_basket", Label: "سلة الليدر (فاتورة الليدر / المواد والمنظومات المختارة)"},
	{Name: "crew_management", Label: "متابعة تنسيق الحجوزات (حجوزات موجّهة قبل التثبيت)"},
	// unit_technicians: صلاحية ظهور "وحدة التقنيين" كاملة بالقائمة الجانبية —
	// منفصلة عمداً عن content_technician (الأخيرة مستخدمة بأماكن ثانية أوسع:
	// مواد التدريب، الموردين، المنتجات) حتى منح صلاحية واحدة ما يفتح كل شي.
	// ما تُمنح لأي دور افتراضياً — المدير حصراً يشوفها إلا إذا مُنحت يدوياً.
	{Name: "privacy_policy_manage", Label: "إضافة وتعديل سياسات الخصوصية"},
	{Name: "unit_technicians", Label: "وحدة التقنيين (ظهور الوحدة كاملة بالقائمة)"},
	// صلاحية ظهور لكل وحدة إدارية بالقائمة الجانبية. منح صلاحية الوحدة يخلي
	// الموظف يشوفها ويشوف بياناتها، وتبقى صلاحيات الصفحات الداخلية شغالة فوقها
	// (يعني الوحدة تنفتح، وشنو يشوف جواها ينحدد بصلاحياته التفصيلية).
	{Name: "unit_service", Label: "وحدة الخدمة"},
	{Name: "unit_design", Label: "وحدة التصميم"},
	{Name: "unit_pr", Label: "وحدة الإعلام والعلاقات العامة"},
	{Name: "unit_quality", Label: "وحدة الجودة والسلامة المهنية"},
	{Name: "unit_monitoring", Label: "وحدة الرقابة"},
	{Name: "unit_procurement", Label: "وحدة المشتريات والمخازن"},
	{Name: "unit_finance", Label: "وحدة الحسابات"},
	{Name: "unit_hr", Label: "وحدة الكوادر التنفيذية"},
	{Name: "unit_projects", Label: "وحدة إدارة المشاريع"},
}

// RoleDefaultPermissions هي الصلاحيات الافتراضية لكل دور وظيفي
var RoleDefaultPermissions = map[string][]string{
	"ADMIN":             {},
	"SALES":             {"sales_booking", "complaints"},
	"HR_COORDINATOR":    {"staff_management", "edit_employee_profile", "coordinator", "manage_customers", "view_bookings", "manage_services", "inventory", "complaints", "mission_tracking", "sales_booking"},
	"TECHNICIAN":        {"expenses"},
	"PROJECT_MANAGER":   {"project_management", "expenses", "mission_tracking"},
	"MONITOR":           {"staff_management", "edit_employee_profile", "kpi_management", "view_bookings", "manage_customers", "manage_services", "mission_tracking", "inventory", "complaints", "finance", "monitoring", "auditing", "quality_control", "gps_system"},
	"FINANCE":           {"finance", "view_bookings"},
	"GPS_ADMIN":         {"gps_system"},
	"QUALITY_ENGINEER":  {"auditing", "complaints", "quality_control", "sales_booking", "kpi_management"},
	"ENGINEER":          {"expenses", "quotation_manage_all", "project_management"},
	"PROCUREMENT_ADMIN": {"procurement", "inventory", "tool_requests_approve"},
	"DESIGNER":          {},
	"SERVICE_MANAGER":   {"content_technician", "manage_services"},
}
