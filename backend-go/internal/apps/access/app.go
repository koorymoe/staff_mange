// Package access — الصلاحيات نفسها على «ون بيس» (منو يگدر يوصل لوين).
// أخطر قسم بالنظام: كلها لمدير النظام، إلا صلاحيات الموظف نفسه (يشوفها هو ومشرفه).
package access

import "staffmange-api/internal/core"

func New() *core.SetApp {
	admin := core.AdminRule
	return core.NewSetApp("access", map[core.Action]core.Rule{
		"access.list":           admin("يشوف كل الصلاحيات"),
		"access.audit_defaults": admin("يدقق الصلاحيات الافتراضية للأدوار"),
		"access.role_defaults":  {Label: "يشوف الصلاحيات الافتراضية لكل دور (للواجهة)", Anyone: true},
		"access.holders":        admin("يشوف منو عنده صلاحية معيّنة"),
		"access.employee_view":  {Label: "يشوف صلاحيات موظف — نفسه أو مشرفه (requireSelfOrSupervisor بالمعالج)", Anyone: true},
		"access.employee_set":   admin("يغيّر صلاحيات موظف (صلاحيات المالك للمالك بس — legacyPermLock)"),
		"access.employee_reset": admin("يرجّع صلاحيات موظف للافتراضي"),
		"access.reset_all":      admin("يرجّع صلاحيات الكل للافتراضي"),
		"access.catalog":        admin("يشوف دليل الصلاحيات"),
		"access.role_drift":     admin("يشوف الموظفين الي صلاحياتهم طالعة عن دورهم"),
		"access.role_drift_fix": admin("يرجّع موظف لصلاحيات دوره"),
		"access.exception_add":  admin("يضيف استثناء صلاحية"),
		"access.exception_del":  admin("يشيل استثناء صلاحية"),
		"access.audit_log":      admin("يشوف سجل تغييرات الصلاحيات"),
	})
}
