package database

// عدّاد ساعات البيت الحي + تصفير الفترة (يوم ٢٧ وبعده).
// التصفير ما يحذف: يسكّر السجلات (closedAt) فالعدّادات تبدي من صفر،
// والسجلات القديمة تبقى بالإكسل.
func remoteTimerMigrations() []Migration {
	return []Migration{{
		Version: "0307_remote_timer",
		SQL: `
			CREATE TABLE IF NOT EXISTS "RemoteTimer" (
				"employeeId" TEXT PRIMARY KEY REFERENCES "Employee"(id) ON DELETE CASCADE,
				"startedAt" TIMESTAMPTZ,
				"accumulated" INT NOT NULL DEFAULT 0,
				"updatedAt" TIMESTAMPTZ NOT NULL DEFAULT now()
			);
			ALTER TABLE "RemoteWorkEntry" ADD COLUMN IF NOT EXISTS "closedAt" TIMESTAMPTZ;
			ALTER TABLE "RemoteWorkEntry" ADD COLUMN IF NOT EXISTS "closedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL;
			ALTER TABLE "RemoteWorkEntry" ADD COLUMN IF NOT EXISTS source TEXT NOT NULL DEFAULT 'MANUAL';
		`,
	}}
}
