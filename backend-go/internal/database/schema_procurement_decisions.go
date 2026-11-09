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
	}, {
		// قرار (ع) 10-10: الفاتورة المجانية تنعتمد بلا رقم — ما إلها فاتورة بالنظام الثاني.
		// القيد يظل يمنع أي فاتورة **غير مجانية** تنعتمد بلا رقم.
		Version: "0343_free_invoice_no_number",
		SQL: `
			ALTER TABLE "LeaderInvoice" DROP CONSTRAINT IF EXISTS leader_invoice_approved_needs_number;
			ALTER TABLE "LeaderInvoice" ADD CONSTRAINT leader_invoice_approved_needs_number CHECK (
				status <> 'APPROVED'
				OR ("externalInvoiceNumber" IS NOT NULL AND btrim("externalInvoiceNumber") <> '')
				OR "isFree" OR COALESCE("auditVerdict", '') = 'FREE'
			) NOT VALID;
		`,
	}}
}
