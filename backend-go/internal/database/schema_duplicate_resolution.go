package database

// تدقيق التكرار: الموظف يحل الزوج (دمج زبونين / طلب حذف حجز مكرر) — نحفظ
// شنو انسوّى حتى يبقى أثر بعد ما المكرر ينشال.
func duplicateResolutionMigrations() []Migration {
	return []Migration{{
		Version: "0309_duplicate_resolution",
		SQL: `ALTER TABLE "DuplicateCandidate" ADD COLUMN IF NOT EXISTS resolution TEXT;
		      ALTER TABLE "DuplicateCandidate" ADD COLUMN IF NOT EXISTS "resolutionNote" TEXT;`,
	}}
}
