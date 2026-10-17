package database

// ══════════════════════════════════════════════════════════════════
// تذكير معاودة الاتصال بالزبون الي ما رد
// ══════════════════════════════════════════════════════════════════
//
// حالة «في الانتظار» موجودة وتعدّ المحاولات، بس ماكو شي يرجع ينبّه
// الإداري «اتصل عليه مرة ثانية». فالحجز يقعد بالطابور بلا حركة —
// مو ملغي ومو شغّال، وبعد شهر أحد يسأل «شنو صار بحجز فلان؟».
//
// هذول عمودين للحد من الإزعاج مو للتذكير نفسه: بدونهم الكنسة تدزّ
// نفس التذكير كل ساعة وينتحول لضجيج ينتجاهل — وتذكير ينتجاهل أسوأ
// من ماكو تذكير.
//
//	lastWaitingReminderAt  آخر مرة ذكّرنا → فاصل ٢٤ ساعة بين تذكيرين
//	waitingReminderCount   عدد التذكيرات → يوقف بعد خمسة
//
// الخمسة مقصودة: زبون ما رد خمس مرات هذا **قرار** مو انتظار، ولازم
// الإداري يقرر يلغي أو يتابع بطريقة ثانية، مو النظام يضل ينقّ عليه.
func waitingReminderMigration() []Migration {
	return []Migration{
		{
			Version: "0238_booking_waiting_reminder",
			SQL: `
				ALTER TABLE "Booking"
					ADD COLUMN IF NOT EXISTS "lastWaitingReminderAt" TIMESTAMPTZ,
					ADD COLUMN IF NOT EXISTS "waitingReminderCount" INTEGER NOT NULL DEFAULT 0;

				CREATE INDEX IF NOT EXISTS "Booking_waiting_reminder_idx"
					ON "Booking" ("waitingSince")
					WHERE status = 'WAITING' AND "archivedAt" IS NULL;
			`,
		},
	}
}
