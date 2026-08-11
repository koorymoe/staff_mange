package database

// ══════════════════════════════════════════════════════════════════
// تتبّع مراحل الحجز — منو رحّله، ملاحظات الترحيل، ووقت الإلغاء
// ══════════════════════════════════════════════════════════════════
//
// ثلاث فجوات انكشفن سوه:
//
// ١. **منو رحّل الحجز؟** الحجز الي بانتظار التثبيت ما يقول منو أدخله.
//    الإداري يشوف قائمة حجوزات بلا ما يعرف لمنو يرجع لو ناقصة معلومة،
//    و«منو أدخل هذا الحجز؟» تنسأل بالتلفون بدل ما تكون مكتوبة.
//    ماكو ولا عمود يحفظها — createdAt موجود بس createdById لا.
//
// ٢. **ملاحظة الترحيل تروح بالهوا.** الإداري يرحّل الحجز للكادر ويريد
//    يوصّلهم شي («الزبون ما يرد قبل العصر»، «الدرج ضيق جيبوا سلّم
//    قصير») — فيتصل بيهم. adminNotes موجود بس ملاحظات إدارية عامة
//    مو موجّهة، وcrewNotes موجود بالجدول من زمان و**ما ينستعمل ولا
//    بمكان واحد** بالكود.
//    وترحيل المشروع لمدير المشاريع نفس القصة بلا حتى عمود.
//
// ٣. **الإلغاء بلا وقت.** الحالة تصير CANCELLED وبس — ماكو وقت ولا
//    منو ألغى. يعني ما نكدر نفرّق بين حجز انلغى **قبل** التثبيت
//    (زبون بدّل رأيه بالتلفون) وحجز انلغى **بعده** (وعدناه وخلفنا)،
//    وهذا فرق حقيقي بالمسؤولية.
func bookingStageTrackingMigration() []Migration {
	return []Migration{
		{
			Version: "0242_booking_stage_tracking",
			SQL: `
				ALTER TABLE "Booking"
					-- منو أدخل/رحّل الحجز. ON DELETE SET NULL: الموظف ممكن
					-- ينحذف بس الحجز وسجله يبقون.
					ADD COLUMN IF NOT EXISTS "createdById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,

					-- ملاحظة الإداري للكادر المنفّذ (crewNotes موجود أصلاً
					-- بالجدول وما ينستعمل — نستعمله بدل ما نضيف عمود ثالث).
					ADD COLUMN IF NOT EXISTS "crewNotesById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					ADD COLUMN IF NOT EXISTS "crewNotesAt" TIMESTAMPTZ,

					-- ملاحظة الإداري لمدير المشاريع — منفصلة عن ملاحظة
					-- الكادر: الاثنين يقرون أشياء مختلفة، ودمجهن يخلي كل
					-- واحد يقرا كلام مو إله.
					ADD COLUMN IF NOT EXISTS "projectNotes" TEXT,
					ADD COLUMN IF NOT EXISTS "projectNotesById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					ADD COLUMN IF NOT EXISTS "projectNotesAt" TIMESTAMPTZ,

					-- الإلغاء: وقته ومنو وسببه
					ADD COLUMN IF NOT EXISTS "cancelledAt" TIMESTAMPTZ,
					ADD COLUMN IF NOT EXISTS "cancelledById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					ADD COLUMN IF NOT EXISTS "cancelReason" TEXT;

				-- الحجوزات الملغاة الي صارت قبل هذا التتبّع: نحط وقت
				-- الإلغاء = آخر تعديل، حتى ما تطلع كلها بسلّة «قبل
				-- التثبيت» بالغلط.
				-- ⚠️ تقدير مو حقيقة — بس أقرب من الفراغ، ويصير مرة وحدة
				-- بس (WHERE "cancelledAt" IS NULL يخليها idempotent).
				UPDATE "Booking"
				SET "cancelledAt" = COALESCE("lastEditedAt", "updatedAt", "createdAt")
				WHERE status = 'CANCELLED' AND "cancelledAt" IS NULL;
			`,
		},
	}
}
