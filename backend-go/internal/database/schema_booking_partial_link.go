package database

// ══════════════════════════════════════════════════════════════════
// ربط حجزين تاريخيين منفصلين كإنجاز جزئي لنفس الشغلة
// ══════════════════════════════════════════════════════════════════
//
// المشكلة: شغلة حقيقية طوّلت أكثر من يوم استوردت (`import-history.sh`)
// كصفوف `Booking` منفصلة تماماً — لكل يوم كوده الخاص، حالته
// `COMPLETED` من أول استيراد، وبلا أي علاقة بينها بالجدول. صاحب
// العمل يشوفها بشاشة تدقيق الحسابات وتبين له تكراراً (نفس الزبون
// ونفس الخدمة)، بس هذا **مو تكرار** — يوم ١ ويوم ٢ من نفس الشغلة.
//
// ⚠️ وهذا **غير** حالة `PARTIAL` الحية (`schema_partial_completion.go`):
// ذيچ تخص حجزاً واحداً وهو شغل جارٍ (نفس معرّف الحجز طول الوقت،
// الليدر يبلّغ يومياً). هذا الحجزين مستوردين تاريخياً، الاثنين
// `COMPLETED` من الأساس، وصفّان منفصلان كلياً بالجدول — ما تنطبق
// عليهم آلية `PartialComplete` الحية إطلاقاً.
//
// الحل: عمود ربط ذاتي بسيط. الحجز التابع (يوم ٢ أو ٣ أو ٤) يشير
// لمعرّف حجز اليوم الأول (أو أي حجز بنفس الشغلة يختاره صاحب العمل)،
// فلو الشغلة طوّلت ٤ أيام، الثلاثة التابعين كلهم يشيرون **لنفس
// المعرّف** — يعني الاستعلام العكسي `WHERE id = $anchor OR
// "partialJobBookingId" = $anchor` يرجّع المجموعة كاملة بلا حاجة
// لبنية شجرية.
func bookingPartialLinkMigration() []Migration {
	return []Migration{
		{
			Version: "0289_booking_partial_job_link",
			SQL: `
				ALTER TABLE "Booking"
					ADD COLUMN IF NOT EXISTS "partialJobBookingId" TEXT
						REFERENCES "Booking"(id) ON DELETE SET NULL;

				ALTER TABLE "Booking"
					DROP CONSTRAINT IF EXISTS booking_partial_job_not_self;
				ALTER TABLE "Booking"
					ADD CONSTRAINT booking_partial_job_not_self
						CHECK ("partialJobBookingId" IS NULL OR "partialJobBookingId" <> id);

				CREATE INDEX IF NOT EXISTS "Booking_partialJobBookingId_idx"
					ON "Booking" ("partialJobBookingId")
					WHERE "partialJobBookingId" IS NOT NULL;
			`,
		},
	}
}
