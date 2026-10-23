package database

// التصفير الشهري (٢٧ الساعة ١١ بالليل): التقييم يتسجّل «رجع بالتصفير
// الشهري» بدل ما ينحذف — السجل يبقى.
func monthlyResetMigrations() []Migration {
	return []Migration{{
		Version: "0308_kpi_cancel_note",
		SQL:     `ALTER TABLE "KpiEvaluation" ADD COLUMN IF NOT EXISTS "cancelNote" TEXT;`,
	}}
}
