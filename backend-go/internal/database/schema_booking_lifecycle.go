package database

// bookingLifecycleVersionedMigrations ثلاث ثغرات بدورة حياة الحجز طلعن
// من الاستعمال اليومي.
func bookingLifecycleVersionedMigrations() []Migration {
	return []Migration{
		{
			// ═══ الحذف صار أرشفة ═══
			// الزبون يلغي الحجز، والإداري يحذفه — وكان يروح فعلاً:
			// DELETE FROM "Booking". يعني الحجز وتاريخه وتعييناته وسبب
			// إلغائه ينمحون للأبد، وما يبقى ولا أثر يجاوب «ليش هذا الزبون
			// ألغى؟» ولا «شكد حجز ينلغي بالشهر؟».
			//
			// وأخطر من هيچ: المحو كان يشيل الصف بالـCASCADE — يعني حتى
			// طلب الحذف نفسه (منو طلبه وليش ومنو وافق) ينمحي معاه، فما
			// يبقى دليل على القرار أصلاً.
			//
			// هسه ننشّم الحجز بـ"archivedAt". يختفي من الحجوزات ومن تنسيق
			// الحجوزات، ويضل بالأرشيف بكل تفاصيله وسبب حذفه ومنو حذفه.
			Version: "0210_booking_archive",
			SQL: `
				ALTER TABLE "Booking"
					ADD COLUMN IF NOT EXISTS "archivedAt" TIMESTAMPTZ,
					ADD COLUMN IF NOT EXISTS "archivedById" TEXT REFERENCES "Employee"(id),
					ADD COLUMN IF NOT EXISTS "archiveReason" TEXT;

				-- كل شاشة عاملة (حجوزات/تنسيق/مهامي) تسأل «مو مؤرشف»،
				-- والأرشيف يسأل العكس — فهرس جزئي لكل اتجاه
				CREATE INDEX IF NOT EXISTS "Booking_not_archived_idx"
					ON "Booking" ("createdAt" DESC) WHERE "archivedAt" IS NULL;
				CREATE INDEX IF NOT EXISTS "Booking_archived_idx"
					ON "Booking" ("archivedAt" DESC) WHERE "archivedAt" IS NOT NULL;
			`,
		},
		{
			// ═══ حالة «في الانتظار» ═══
			// بعد التثبيت نتصل بالزبون حتى نطلعله — والزبون ما يرد. شنو
			// نسوي بالحجز؟ ما ينلغى (الزبون ممكن يرد بكرة)، وما يضل
			// مثبّت (لأنه يطلع بقائمة الجاهز للتوجيه ويربك التنسيق).
			//
			// فصارت حالة قائمة بذاتها: الحجز ينزاح من طابور الشغل ويضل
			// محفوظ لحد ما الزبون يرد أو الإداري يقرر يلغيه.
			//
			// ملاحظة: ALTER TYPE ADD VALUE مسموحة داخل معاملة من
			// بوستكرس ١٢ فما فوق (إحنا على ١٦)، بشرط ما نستعمل القيمة
			// الجديدة بنفس المعاملة — وهنا ما نستعملها، بس نضيفها.
			Version: "0211_booking_status_waiting",
			SQL:     `ALTER TYPE "BookingStatus" ADD VALUE IF NOT EXISTS 'WAITING'`,
		},
		{
			// ═══ تفاصيل الانتظار والتأجيل ═══
			//
			// الانتظار: نسجّل متى دخل ومنو حطه وكم مرة حاولنا نتصل. عدد
			// المحاولات مهم — زبون ما رد مرة غير زبون ما رد خمس مرات،
			// والثاني لازم يطلع للإداري كقرار مو كطابور.
			//
			// التأجيل: الموعد إله سجل تغييرات أصلاً ("ScheduleChangeLog")،
			// بس ماكو فرق بين «الإداري رتّب الجدول» و«الزبون أجّل». هذولا
			// عمودين يخلون التأجيل مرئي: كم مرة تأجل الحجز وآخر سبب.
			// حجز تأجل أربع مرات علامة على شي غلط، ولازم ينشاف.
			Version: "0212_booking_waiting_and_postpone",
			SQL: `
				ALTER TABLE "Booking"
					ADD COLUMN IF NOT EXISTS "waitingSince" TIMESTAMPTZ,
					ADD COLUMN IF NOT EXISTS "waitingNote" TEXT,
					ADD COLUMN IF NOT EXISTS "contactAttempts" INTEGER NOT NULL DEFAULT 0,
					ADD COLUMN IF NOT EXISTS "lastContactAttemptAt" TIMESTAMPTZ,
					ADD COLUMN IF NOT EXISTS "waitingById" TEXT REFERENCES "Employee"(id),
					ADD COLUMN IF NOT EXISTS "postponeCount" INTEGER NOT NULL DEFAULT 0,
					ADD COLUMN IF NOT EXISTS "lastPostponedAt" TIMESTAMPTZ,
					ADD COLUMN IF NOT EXISTS "postponeReason" TEXT;

				CREATE INDEX IF NOT EXISTS "Booking_waiting_idx"
					ON "Booking" ("waitingSince" DESC) WHERE "waitingSince" IS NOT NULL;
			`,
		},
		{
			// سبب التأجيل ينحفظ بسجل المواعيد بعد، حتى يبقى مربوط بالتغيير
			// نفسه مو بس بآخر حالة للحجز.
			Version: "0213_schedule_log_reason",
			SQL: `
				ALTER TABLE "ScheduleChangeLog"
					ADD COLUMN IF NOT EXISTS reason TEXT,
					ADD COLUMN IF NOT EXISTS kind TEXT NOT NULL DEFAULT 'SCHEDULE';
			`,
		},
	}
}
