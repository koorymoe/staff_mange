package database

// nightFeaturesVersionedMigrations يرجّع الترحيلات المرقّمة لميزتين أضيفتا بنفس
// الليلة: (1) طابع "تم" تواصل الإداري مع الزبون قبل تثبيت الحجز (يفصل خطوة
// "تواصلت مع الزبون" عن تغيير حالة الحجز نفسها، حتى يقدر المراقب بصلاحية
// crew_management يدقق هل صار التواصل قبل التثبيت الفعلي)، و(2) لقطة الأدوات
// الشخصية الناقصة عند لحظة استلام الحجز (بديل سريع عن جرد كامل منفصل).
func nightFeaturesVersionedMigrations() []Migration {
	return []Migration{
		{
			Version: "0129_add_booking_confirmation_contacted",
			SQL: `ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "confirmationContactedAt" TIMESTAMP;
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "confirmationContactedById" TEXT`,
		},
		{
			Version: "0130_create_booking_tool_check",
			SQL: `CREATE TABLE IF NOT EXISTS "BookingToolCheck" (
				id TEXT PRIMARY KEY,
				"bookingId" TEXT NOT NULL REFERENCES "Booking"(id) ON DELETE CASCADE,
				"employeeId" TEXT NOT NULL REFERENCES "Employee"(id),
				"missingItems" TEXT,
				"checkedAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS "BookingToolCheck_bookingId_idx" ON "BookingToolCheck"("bookingId");
			CREATE INDEX IF NOT EXISTS "BookingToolCheck_employeeId_idx" ON "BookingToolCheck"("employeeId")`,
		},
	}
}
