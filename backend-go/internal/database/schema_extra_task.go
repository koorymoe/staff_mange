package database

// ══════════════════════════════════════════════════════════════════
// المهام الإضافية — المدير يوجّه شغل لموظف
// ══════════════════════════════════════════════════════════════════
//
// طلب صاحب العمل: «يجي المدير يوجّه أحد الموظفين شيسوي؟ يختار موظف
// ويوجّهله عمل يطلعله بالنظام… مثلاً أريد أوجّهه على تخريج فواتير.
// هاي راح تطلع يم منو؟ يم الموظف الي انتوجّه له العمل».
//
// الفجوة قبلها: كل شغل بالنظام مربوط **بحجز**. الشغل الي مو حجز —
// «رتّب المخزن»، «خرّج فواتير الشهر»، «راجع عقود الزبائن» — ينقال
// بالتلفون أو بالمجموعة، وما ينحفظ ولا ينتابع. النتيجة: المدير ما
// يتذكر منو وجّه بشنو، والموظف يگول «ما وصلني»، وماكو أثر يفصل.
//
// ⚠️ ليش جدول مستقل مو حجز بنوع خاص؟
// الحجز إله زبون وموعد وكادر وفاتورة وتقرير — والمهمة الإضافية ما
// إلها ولا وحدة منهن. حشرها بجدول الحجوزات يعني أعمدة فاضية بكل صف،
// وتلوّث كل إحصائية تعدّ الحجوزات (شكد حجز عدنا هذا الشهر؟ يطلع رقم
// يشمل «رتّب المخزن»).
//
// ⚠️ وما تمس ولا حساب: ماكو فاتورة ولا عمولة ولا غرامة تنبني عليها
// بهاي المرحلة. توجيه ومتابعة بس.
func extraTaskMigration() []Migration {
	return []Migration{
		{
			Version: "0245_extra_task",
			SQL: `
				CREATE TABLE IF NOT EXISTS "ExtraTask" (
					id TEXT PRIMARY KEY,
					title TEXT NOT NULL,
					description TEXT,

					-- منو ينفّذها ومنو وجّهها.
					-- ⚠️ ON DELETE CASCADE للمكلّف: مهمة بلا منفّذ ما
					-- إلها معنى. أما الموجِّه فـSET NULL — المهمة تبقى
					-- بالسجل حتى لو المدير انحذف.
					"assignedToId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
					"assignedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,

					priority TEXT NOT NULL DEFAULT 'NORMAL',   -- NORMAL|URGENT
					"dueAt" TIMESTAMPTZ,

					-- NEW: وصلت وما فتحها · IN_PROGRESS: بدأ · DONE · CANCELLED
					status TEXT NOT NULL DEFAULT 'NEW',
					"seenAt" TIMESTAMPTZ,
					"startedAt" TIMESTAMPTZ,
					"doneAt" TIMESTAMPTZ,
					-- شنو سوّى بالضبط. إجباري وقت الإنجاز: «تم» بلا وصف
					-- ما تنفع لا للمتابعة ولا للتقييم.
					"doneNote" TEXT,

					-- الإلغاء بسببه — نفس منطق إلغاء الحجز.
					"cancelledAt" TIMESTAMPTZ,
					"cancelReason" TEXT,

					"createdAt" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					"updatedAt" TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);

				-- الموظف يفتح شاشته: مهامه المفتوحة أول.
				CREATE INDEX IF NOT EXISTS "ExtraTask_assignee_idx"
					ON "ExtraTask" ("assignedToId", status, "createdAt" DESC);
				-- المدير يتابع الي وجّهه.
				CREATE INDEX IF NOT EXISTS "ExtraTask_assigner_idx"
					ON "ExtraTask" ("assignedById", "createdAt" DESC);
			`,
		},
	}
}
