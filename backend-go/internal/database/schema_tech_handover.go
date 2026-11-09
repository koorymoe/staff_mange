package database

// ═══ ترحيل الحجز للتقني (قرار (ع) 10-07) ═══
// مشكلة ما يعرفون الفنيين يحلّوها: الإداري يرحّل الحجز لتقني أو مسؤول خدمة.
// الإداري ينحاسب على الاستجابة والتواصل والترحيل نفسه، وبعدها كلشي على التقني:
// يتواصل ويا الزبون، يكتب الكشف، ويعالج بنفسه أو يطلب طاقم.
func techHandoverMigrations() []Migration {
	return []Migration{{
		Version: "0326_tech_handover",
		SQL: `
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "handoverToId" TEXT REFERENCES "Employee"(id);
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "handoverById" TEXT REFERENCES "Employee"(id);
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "handoverAt" TIMESTAMP;
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "handoverReason" TEXT;
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "techContactedAt" TIMESTAMP;
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "techDiagnosis" TEXT;
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "techDiagnosedAt" TIMESTAMP;
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "techCrewRequestedAt" TIMESTAMP;
			CREATE INDEX IF NOT EXISTS "Booking_handoverToId_idx" ON "Booking" ("handoverToId") WHERE "handoverToId" IS NOT NULL;
		`,
	}}
}
