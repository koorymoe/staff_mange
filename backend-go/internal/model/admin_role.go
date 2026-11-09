package model

// IsAdminRole (DRY 11-02) مدير النظام أو المالك — بدل `role == "ADMIN" || role == "OWNER"`
// المكتوبة بعشرات الأماكن. يقبل string أو أي نوع دور مبني عليه.
func IsAdminRole[T ~string](role T) bool {
	return role == "ADMIN" || role == "OWNER"
}

// HasPermSQL (DRY 11-02) شرط SQL «الموظف عنده الصلاحية» — بدل الاستعلام الفرعي
// المكتوب بإيد بكل ملف. emp تعبير عمود الموظف (e.id أو $1)، وcond شرط اسم الصلاحية
// (= 'finance' أو = $2 أو IN ('a','b')). يُستعمل داخل نص SQL بعلامات `.
func HasPermSQL(emp, cond string) string {
	return `EXISTS (SELECT 1 FROM "EmployeePermission" hp_ep JOIN "Permission" hp_p ON hp_p.id = hp_ep."permissionId"
		WHERE hp_ep."employeeId" = ` + emp + ` AND hp_p.name ` + cond + `)`
}
