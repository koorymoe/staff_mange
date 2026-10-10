package database

// ═══ قرارات أبو الكميات تنسجّل (قرار (ع) 10-10) ═══
// «كل شغلة يشتغلها لازم تكون موجودة بالنظام» — رفض طلب المواد أو تحويله «قيد
// التنفيذ» ورفض طلب الأداة ما چانوا يسجّلون منو ووكت، فما ينقاس سرعة الرد.
func procurementDecisionMigrations() []Migration {
	return []Migration{{
		Version: "0342_procurement_decisions",
		SQL: `
			ALTER TABLE "ProcurementRequest" ADD COLUMN IF NOT EXISTS "decidedAt" TIMESTAMP,
				ADD COLUMN IF NOT EXISTS "decidedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL;
			UPDATE "ProcurementRequest" SET "decidedAt" = "fulfilledAt", "decidedById" = "fulfilledById"
				WHERE "decidedAt" IS NULL AND "fulfilledAt" IS NOT NULL;
			ALTER TABLE "ToolRequest" ADD COLUMN IF NOT EXISTS "rejectedAt" TIMESTAMP,
				ADD COLUMN IF NOT EXISTS "rejectedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL;
		`,
	}}
}
