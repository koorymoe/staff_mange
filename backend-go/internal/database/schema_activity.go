package database

// سجل نشاط الموظف — كل حفظ ناجح (إضافة/تعديل/حذف/قرار). (ع): «أدخل على
// المراقب يطلعلي شنو سوّى اليوم». يكتبه RequireAuth لحاله لكل المسارات.
func employeeActivityMigrations() []Migration {
	return []Migration{{
		Version: "0305_employee_activity",
		SQL: `
			CREATE TABLE IF NOT EXISTS "EmployeeActivity" (
				id BIGSERIAL PRIMARY KEY,
				"employeeId" TEXT NOT NULL,
				method TEXT NOT NULL,
				pattern TEXT NOT NULL,
				path TEXT NOT NULL,
				status INT NOT NULL,
				"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now()
			);
			CREATE INDEX IF NOT EXISTS "EmployeeActivity_emp_time_idx" ON "EmployeeActivity"("employeeId", "createdAt");
		`,
	}}
}
