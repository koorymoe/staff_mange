package database

// leaderInvoiceVersionedMigrations يرجّع الترحيلات المرقّمة الخاصة بميزة فوترة
// الليدر (تحل محل شيت جوجل): كتالوج أسعار المنظومات الثمانية، جدول المواد
// (مواد الشد) بكود فريد وسعر بيع وربح، وفاتورة الليدر مع بنودها من المواد.
func leaderInvoiceVersionedMigrations() []Migration {
	return []Migration{
		{
			Version: "0120_create_system_price_catalog",
			SQL: `CREATE TABLE IF NOT EXISTS "SystemPriceCatalog" (
				id TEXT PRIMARY KEY,
				"systemName" TEXT NOT NULL,
				"itemName" TEXT NOT NULL,
				category TEXT NOT NULL CHECK (category IN ('install', 'wiring', 'programming')),
				value NUMERIC NOT NULL,
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				UNIQUE ("systemName", "itemName", category)
			);
			CREATE INDEX IF NOT EXISTS "SystemPriceCatalog_systemName_idx" ON "SystemPriceCatalog"("systemName")`,
		},
		{
			Version: "0121_create_material",
			SQL: `CREATE TABLE IF NOT EXISTS "Material" (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				code TEXT NOT NULL UNIQUE,
				"sellPrice" NUMERIC NOT NULL,
				"profitPerUnit" NUMERIC NOT NULL DEFAULT 0,
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS "Material_code_idx" ON "Material"(code)`,
		},
		{
			Version: "0122_create_leader_invoice",
			SQL: `CREATE TABLE IF NOT EXISTS "LeaderInvoice" (
				id TEXT PRIMARY KEY,
				"bookingId" TEXT REFERENCES "Booking"(id),
				"employeeId" TEXT NOT NULL REFERENCES "Employee"(id),
				"customerName" TEXT,
				"customerPhone" TEXT,
				"customerAddress" TEXT,
				systems JSONB NOT NULL DEFAULT '[]',
				items JSONB NOT NULL DEFAULT '[]',
				"totalDeviceCount" INTEGER NOT NULL DEFAULT 0,
				"executionCost" NUMERIC NOT NULL DEFAULT 0,
				"materialsTotal" NUMERIC NOT NULL DEFAULT 0,
				"discountValue" NUMERIC NOT NULL DEFAULT 0,
				"netTotal" NUMERIC NOT NULL DEFAULT 0,
				"accountingCode" TEXT NOT NULL UNIQUE,
				status TEXT NOT NULL DEFAULT 'DRAFT',
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS "LeaderInvoice_employeeId_idx" ON "LeaderInvoice"("employeeId");
			CREATE INDEX IF NOT EXISTS "LeaderInvoice_bookingId_idx" ON "LeaderInvoice"("bookingId");
			CREATE INDEX IF NOT EXISTS "LeaderInvoice_createdAt_idx" ON "LeaderInvoice"("createdAt")`,
		},
		{
			Version: "0123_create_leader_invoice_material_item",
			SQL: `CREATE TABLE IF NOT EXISTS "LeaderInvoiceMaterialItem" (
				id TEXT PRIMARY KEY,
				"leaderInvoiceId" TEXT NOT NULL REFERENCES "LeaderInvoice"(id) ON DELETE CASCADE,
				"materialId" TEXT REFERENCES "Material"(id),
				name TEXT NOT NULL,
				quantity NUMERIC NOT NULL DEFAULT 1,
				"unitPrice" NUMERIC NOT NULL DEFAULT 0,
				"profitPerUnit" NUMERIC NOT NULL DEFAULT 0,
				"lineTotal" NUMERIC NOT NULL DEFAULT 0,
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS "LeaderInvoiceMaterialItem_leaderInvoiceId_idx" ON "LeaderInvoiceMaterialItem"("leaderInvoiceId")`,
		},
	}
}
