package database

// إيقاف حساب موظف ترك الشركة — بلا حذف: سجله وحجوزاته تبقى، ودخوله
// ينقطع فوراً (RequireAuth يفحص الحالة بكل طلب). هنا نحفظ منو أوقفه وليش.
func employeeSuspendMigrations() []Migration {
	return []Migration{{
		Version: "0302_employee_suspend_reason",
		SQL: `
			ALTER TABLE "Employee" ADD COLUMN IF NOT EXISTS "suspendedReason" TEXT;
			ALTER TABLE "Employee" ADD COLUMN IF NOT EXISTS "suspendedAt" TIMESTAMPTZ;
			ALTER TABLE "Employee" ADD COLUMN IF NOT EXISTS "suspendedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL;
		`,
	}}
}
