// Package people — الموظفين على «ون بيس»: الملف، الهوية، الدوام، الإيقاف، النطاق،
// التسليم. كل حارس قديم صار قاعدة بنفس المعنى بالضبط (منقولة 11-10).
//
// القراءات العامة (القائمة، الملف) «أي موظف» عن قصد: المعالج يقص الحقول حسب منو
// يسأل (writeEmployeeFor / ViewEmployees) — الراتب ما يطلع لغير الإداري.
package people

import "staffmange-api/internal/core"

func New() *core.SetApp {
	admin, perm := core.AdminRule, core.PermRule
	any := func(l string) core.Rule { return core.Rule{Label: l, Anyone: true} }
	adminOr := func(l, p string) core.Rule { return core.RolePermRule(l, []string{"ADMIN"}, p) }
	return core.NewSetApp("people", map[core.Action]core.Rule{
		// قبل: requireAuth بس والفحص جوّا المعالج. المدير بلا الصلاحية ما يضيف (Strict).
		"people.create":            {Label: "يضيف موظف جديد", Strict: true, Roles: []string{"OWNER"}, Perms: []string{"employee_create"}},
		"people.list":              any("يشوف قائمة الموظفين (الحقول حسب منو يسأل)"),
		"people.match":             any("يدوّر موظف بالاسم"),
		"people.supervisors":       any("يشوف المشرفين (ViewEmployees)"),
		"people.view":              any("يشوف ملف موظف (writeEmployeeFor: الإداري كامل، غيره مقصوص)"),
		"people.my_scope":          any("يشوف نطاقه الإداري هو"),
		"people.leader_skills":     any("يشوف مهارات الليدر"),
		"people.archived":          admin("يشوف الموظفين المؤرشفين"),
		"people.update":            admin("يعدّل موظف"),
		"people.access_view":       admin("يشوف وصول موظف"),
		"people.access_set":        admin("يغيّر وصول موظف"),
		"people.scope_view":        admin("يشوف نطاق مدير"),
		"people.scope_set":         admin("يغيّر نطاق مدير"),
		"people.handover_view":     admin("يشوف تسليم شغل موظف لغيره"),
		"people.handover":          admin("يسلّم شغل موظف لغيره"),
		"people.link_history":      admin("يربط سجلات تاريخية بموظف"),
		"people.edit_all":          adminOr("يعدّل كل بيانات موظف", "employee_edit_all"),
		"people.suspended":         adminOr("يشوف الحسابات الموقوفة", "employee_suspend"),
		"people.suspend":           adminOr("يوقف حساب موظف", "employee_suspend"),
		"people.reactivate":        adminOr("يرجّع تفعيل حساب", "employee_suspend"),
		"people.schedule":          adminOr("يغيّر دوام موظف", "work_schedule_manage"),
		"people.identity":          adminOr("يعدّل اسم/دور موظف", "edit_employee_identity"),
		"people.profile":           adminOr("يعدّل ملف موظف (الراتب ومزايا النفس محمية — selfPerkEdit)", "edit_employee_profile"),
		"people.skills":            perm("يعدّل مهارات موظف", "staff_management"),
		"people.training_done":     perm("يعلّم تدريب موظف مكتمل", "content_technician"),
		"people.vehicle_rating":    perm("يشوف تقييم السائق", "vehicle_management"),
		"people.dossier":           core.RolePermRule("يشوف ملف أداء موظف", []string{"ADMIN", "MONITOR", "HR_COORDINATOR"}, "monitoring", "mon_staff", "staff_management"),
		"people.leader_skills_set": core.RolePermRule("يقيّم مهارات الليدر", []string{"HR_COORDINATOR"}),
	})
}
