package database

// جرد العدّة **بعد** الحجز — طلب (ع) 10-05: «الفنيين ويا الليدر يتحاسبون على
// جرد العدّة بعد الشغل». الجرد الموجود (InventoryCheck) يصير **قبل** الحجز،
// فما يكشف أداة ضاعت عند الزبون. جدول مستقل حتى ما نلمس InventoryCheck
// (بيه SELECT * و RETURNING *).
func afterInventoryMigrations() []Migration {
	return []Migration{{
		Version: "0316_after_inventory",
		SQL: `CREATE TABLE IF NOT EXISTS "BookingAfterInventory" (
			"bookingId"    TEXT NOT NULL REFERENCES "Booking"(id) ON DELETE CASCADE,
			"employeeId"   TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
			complete       BOOLEAN NOT NULL,
			"missingItems" TEXT,
			"checkedAt"    TIMESTAMPTZ NOT NULL DEFAULT now(),
			PRIMARY KEY ("bookingId", "employeeId")
		);
		CREATE INDEX IF NOT EXISTS "BookingAfterInventory_emp_idx" ON "BookingAfterInventory" ("employeeId", "checkedAt" DESC);`,
	}}
}
