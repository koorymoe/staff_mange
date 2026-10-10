package database

// «منو فاتح النظام هسه؟» — آخر ظهور لكل موظف (طلب (ع) 10-04). جدول مستقل
// حتى ما نلمس "Employee" (SELECT * يتكسر إذا انضاف عمود بلا حقل بالنموذج).
func presenceMigrations() []Migration {
	return []Migration{{
		Version: "0311_employee_presence",
		SQL: `CREATE TABLE IF NOT EXISTS "EmployeePresence" (
			"employeeId" TEXT PRIMARY KEY REFERENCES "Employee"(id) ON DELETE CASCADE,
			"lastSeenAt" TIMESTAMPTZ NOT NULL DEFAULT now()
		);`,
	}}
}
