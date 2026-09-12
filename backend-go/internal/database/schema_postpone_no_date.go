package database

// ══════════════════════════════════════════════════════════════════
// التأجيل بلا موعد
// ══════════════════════════════════════════════════════════════════
//
// التأجيل موجود بس يفرض تاريخ جديد. وبالواقع أكثر التأجيلات تصير
// والزبون **ما محدّد** متى يناسبه: «خابرني بعدين»، «هالأسبوع ما أكدر».
// فالإداري يضطر يحط تاريخ من راسه حتى يمرّر الشاشة — والنتيجة موعد
// كذب بالجدول، والكادر يتحضّر لحجز ماكو.
//
// هسه التأجيل يصير بموعد أو بلا موعد. الي بلا موعد ينزاح من جدول
// اليوم ويروح لقائمة «الحجوزات المؤجلة» حتى الإداري يرجع يحدد له
// موعد بعدين.
//
// ⚠️ ليش عمود جديد ومو نستنتجها من scheduledAt IS NULL؟
// لأن الحجز الجديد بعده ما إله موعد أصلاً. العمود يفرّق بين «لسه ما
// انجدول» و«انجدول وانأجّل بلا موعد» — الثاني عليه قرار مطلوب.
func postponeNoDateMigration() []Migration {
	return []Migration{
		{
			Version: "0235_booking_awaiting_reschedule",
			SQL: `
				ALTER TABLE "Booking"
					ADD COLUMN IF NOT EXISTS "awaitingReschedule" BOOLEAN NOT NULL DEFAULT false;

				CREATE INDEX IF NOT EXISTS "Booking_awaiting_reschedule_idx"
					ON "Booking" ("lastPostponedAt" DESC)
					WHERE "awaitingReschedule" AND "archivedAt" IS NULL;
			`,
		},
		{
			// حاجز فعلي: سجل تغييرات المواعيد يفرض newTime NOT NULL،
			// والتأجيل بلا موعد ماكو عنده وقت جديد. بدون هاي الهجرة
			// العملية تفشل وقت التشغيل مو وقت البناء.
			Version: "0236_schedule_log_nullable_new_time",
			SQL: `
				ALTER TABLE "ScheduleChangeLog" ALTER COLUMN "newTime" DROP NOT NULL;
			`,
		},
	}
}
