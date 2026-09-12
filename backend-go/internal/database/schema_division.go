package database

// divisionVersionedMigrations يرجّع الترحيلات المرقّمة لميزة "الشُّعبة" (division) —
// طلب صاحب العمل يفصل موظفي الشعبة الهندسية (كاميرات/شبكات/GPS...، الموجودين
// أصلاً بالنظام) عن موظفي شعبة الديكور الجديدة (حدادة/نجارة/صباغة/سيراميك/لبخ/
// تأسيس ماء ومجاري/جبس بورد). كل موظف وكل خدمة (Service) يصير عندهم عمود
// "division" (TEXT، NOT NULL، افتراضي 'ENGINEERING') — هذا الافتراضي يرجّع كل
// السجلات القديمة تلقائياً (backfill) لـ ENGINEERING بدون أي تدخل يدوي، فما
// يتأثر أي موظف/خدمة موجودة أصلاً.
func divisionVersionedMigrations() []Migration {
	return []Migration{
		{
			Version: "0131_add_employee_division",
			SQL:     `ALTER TABLE "Employee" ADD COLUMN IF NOT EXISTS division TEXT NOT NULL DEFAULT 'ENGINEERING'`,
		},
		{
			Version: "0132_add_service_division",
			SQL:     `ALTER TABLE "Service" ADD COLUMN IF NOT EXISTS division TEXT NOT NULL DEFAULT 'ENGINEERING'`,
		},
	}
}
