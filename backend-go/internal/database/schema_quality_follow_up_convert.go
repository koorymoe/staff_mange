package database

// ══════════════════════════════════════════════════════════════════
// متابعة الجودة: تحويل حقيقي لحجز صيانة + اتصال ثانٍ بعد الإنجاز
// ══════════════════════════════════════════════════════════════════
//
// المشكلة: زر «تحويل لحجز جديد» الموجود أصلاً كان يفتح شاشة إنشاء
// حجز جديد ويأشّر المتابعة CONVERTED **بلا أي ربط حقيقي** — مهندس
// الجودة يفتح حجزاً ويروح، وماكو أثر يرجعه لمتابعة الزبون لمن الحجز
// يخلص. صاحب العمل طلب صراحة: يبقى الحجز معلّق يم مهندس الجودة لحد
// ما ينجز، وبعدها يرجعله يتصل بالزبون مرة ثانية يتأكد الحل انسوى.
//
// الحل: عمود ربط + حالة جديدة `RECONTACT` تنفعّل تلقائياً (كنسة
// خلفية) لمن الحجز المربوط يوصل COMPLETED.
func qualityFollowUpConvertMigration() []Migration {
	return []Migration{
		{
			Version: "0294_quality_follow_up_linked_booking",
			SQL: `
				ALTER TABLE "QualityFollowUp"
					ADD COLUMN IF NOT EXISTS "linkedBookingId" TEXT
						REFERENCES "Booking"(id) ON DELETE SET NULL;

				CREATE INDEX IF NOT EXISTS "QualityFollowUp_linkedBookingId_idx"
					ON "QualityFollowUp" ("linkedBookingId")
					WHERE "linkedBookingId" IS NOT NULL;
			`,
		},
	}
}
