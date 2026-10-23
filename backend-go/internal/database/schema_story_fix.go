package database

// ══════════════════════════════════════════════════════════════════
// تصحيح جدول القصص — الحذف ما يمحي الدليل
// ══════════════════════════════════════════════════════════════════
//
// ⚠️⚠️ **عيب حقيقي بالترحيل 0272**: كتبته
// `"recipientEmployeeId" ... ON DELETE CASCADE`، يعني حذف حساب موظف
// **يمحي كل دليل إنه انخصم وأقرّ بالاطلاع**.
//
// وهذا يكسر نمط مطبَّق بالمشروع **أربع مرات** (`ComplaintEvent` ·
// `CoordinationAlert` · `DesignAsset` · `TeamInventoryFollowUp`):
// الاسم ينتنسخ نصاً، والمفتاح `ON DELETE SET NULL` — حذف الموظف
// **ما يمحي إن الحدث صار**، ولا يخلّي السطر مجهولاً.
//
// ⚠️ **و`recipientRef` مو زينة**: الفهرس الفريد ما يشتغل على عمود
// يصير `NULL`. فمفتاح الـidempotency ينتقل لعمود **ثابت منسوخ**،
// وتبقى ضمانة «نفس الحدث ما ينشئ قصتين» شغّالة **حتى بعد حذف
// الحساب**. بدونه، حذف موظف يفتح باب تكرار القصة عليه لو انرجع.
func storyFixMigrations() []Migration {
	return []Migration{
		{
			Version: "0273_story_recipient_snapshot",
			SQL: `ALTER TABLE "StoryInstance"
			        ADD COLUMN IF NOT EXISTS "recipientRef"  TEXT,
			        ADD COLUMN IF NOT EXISTS "recipientName" TEXT NOT NULL DEFAULT ''`,
		},
		{
			// الصفوف القديمة تاخذ مرجعها من العمود الموجود قبل ما
			// نفكّ الـCASCADE — وإلا نخسر ربطها.
			Version: "0273_story_recipient_backfill",
			SQL: `UPDATE "StoryInstance"
			      SET "recipientRef" = COALESCE("recipientRef", "recipientEmployeeId")
			      WHERE "recipientRef" IS NULL`,
		},
		{
			Version: "0273_story_recipient_ref_required",
			SQL: `ALTER TABLE "StoryInstance" ALTER COLUMN "recipientRef" SET NOT NULL`,
		},
		{
			// الـFK يصير SET NULL — الصف يبقى، والاسم المنسوخ يخليه
			// مقروءاً. اسم القيد يجي من تسمية بوستگرس التلقائية.
			Version: "0273_story_recipient_fk_set_null",
			SQL: `ALTER TABLE "StoryInstance"
			        DROP CONSTRAINT IF EXISTS "StoryInstance_recipientEmployeeId_fkey";
			      ALTER TABLE "StoryInstance"
			        ALTER COLUMN "recipientEmployeeId" DROP NOT NULL;
			      ALTER TABLE "StoryInstance"
			        ADD CONSTRAINT "StoryInstance_recipientEmployeeId_fkey"
			        FOREIGN KEY ("recipientEmployeeId") REFERENCES "Employee"(id) ON DELETE SET NULL`,
		},
		{
			// ⚠️ **الفهرس يشمل نوع الحدث**: `eventId` يجي من جداول
			// مستقلة (`KpiEvaluation` · `DisciplineEvent` · …)، وما
			// ينفع نفترض إنهن بمساحة أسماء وحدة. احتمال التصادم
			// ضئيل، بس العقد غلط والكلفة صفر.
			Version: "0273_story_event_unique_v2",
			SQL: `DROP INDEX IF EXISTS "story_event_recipient_unique";
			      CREATE UNIQUE INDEX IF NOT EXISTS "story_event_recipient_unique_v2"
			        ON "StoryInstance" ("eventKind", "eventId", "recipientRef")`,
		},
		{
			// الفهارس تشتغل على المرجع الثابت مو على المفتاح الي
			// يصير NULL بعد الحذف.
			Version: "0273_story_indexes_on_ref",
			SQL: `DROP INDEX IF EXISTS "story_pending_idx";
			      CREATE INDEX IF NOT EXISTS "story_pending_ref_idx"
			        ON "StoryInstance" ("recipientRef", priority DESC, "createdAt")
			        WHERE status IN ('QUEUED','DELIVERED','PLAYING');
			      DROP INDEX IF EXISTS "story_group_open_unique";
			      CREATE UNIQUE INDEX IF NOT EXISTS "story_group_open_ref_unique"
			        ON "StoryInstance" ("recipientRef", "groupKey")
			        WHERE "groupKey" IS NOT NULL AND status IN ('QUEUED','DELIVERED','PLAYING')`,
		},
		{
			// ⚠️ **الطوابع حقائق و`status` مشتق منهن**: منع رجوع
			// المرحلة مبني بالكود (`Advance`) ومفحوص حياً — بس بالكود
			// بس. القيد بالقاعدة يخليه ما ينكسر لو انكتب مسار ثانٍ
			// بكرة ما يمرّ على نفس الدالة.
			Version: "0273_story_status_check",
			SQL: `ALTER TABLE "StoryInstance"
			        DROP CONSTRAINT IF EXISTS "story_status_valid";
			      ALTER TABLE "StoryInstance"
			        ADD CONSTRAINT "story_status_valid" CHECK (
			          status IN ('QUEUED','DELIVERED','PLAYING','SEEN','OPENED','ACKNOWLEDGED','FAILED')
			          AND (status <> 'ACKNOWLEDGED' OR "acknowledgedAt" IS NOT NULL)
			        )`,
		},
	}
}
