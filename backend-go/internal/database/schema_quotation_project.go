package database

// ═══ عرض السعر مربوط بالمشروع (قرار (ع) 10-09) ═══
// «اعمل عرض سعر» من إدارة المشاريع چان يسوي عرض جديد كل مرة، فالمشروع
// يصير عنده ٤ عروض بعد ٣ تعديلات. هسه عرض واحد لكل مشروع والتعديلات
// تنأرشف بـQuotationVersion. العروض القديمة تنربط بمطابقة كود المشروع
// ببداية projectName (الأحدث بس) — والمكررة ما تنمسح.
func quotationProjectMigrations() []Migration {
	return []Migration{{
		Version: "0337_quotation_project_link",
		SQL: `
			ALTER TABLE "Quotation" ADD COLUMN IF NOT EXISTS "projectId" TEXT REFERENCES "Project"(id) ON DELETE SET NULL;
			CREATE INDEX IF NOT EXISTS "Quotation_projectId_idx" ON "Quotation"("projectId");
			UPDATE "Quotation" q SET "projectId" = m.pid
			FROM (
				SELECT DISTINCT ON (p.id) p.id AS pid, q2.id AS qid
				FROM "Project" p JOIN "Quotation" q2 ON q2."projectName" LIKE p.code || ' —%'
				ORDER BY p.id, q2."createdAt" DESC
			) m
			WHERE q.id = m.qid AND q."projectId" IS NULL;
		`,
	}}
}
