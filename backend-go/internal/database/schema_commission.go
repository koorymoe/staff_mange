package database

// commissionVersionedMigrations يرجّع الترحيلات المرقّمة الخاصة بميزة عمولات
// الليدر/الفنيين المحسوبة تلقائياً عند إنشاء فاتورة ليدر، بالإضافة إلى حقل
// "سعر الجملة" على المنتجات والمواد اللازم لاحتساب هامش الربح فعلياً (مو
// مجرد رقم مخزّن يُصدَّق عليه بدون تحقق).
func commissionVersionedMigrations() []Migration {
	return []Migration{
		{
			Version: "0124_add_product_wholesale_price",
			SQL:     `ALTER TABLE "Product" ADD COLUMN IF NOT EXISTS "wholesalePrice" NUMERIC`,
		},
		{
			Version: "0125_add_material_wholesale_price",
			SQL: `ALTER TABLE "Material" ADD COLUMN IF NOT EXISTS "wholesalePrice" NUMERIC NOT NULL DEFAULT 0;
				UPDATE "Material" SET "wholesalePrice" = "sellPrice" - "profitPerUnit" WHERE "wholesalePrice" = 0`,
		},
		{
			Version: "0126_create_employee_commission",
			SQL: `CREATE TABLE IF NOT EXISTS "EmployeeCommission" (
				id TEXT PRIMARY KEY,
				"employeeId" TEXT NOT NULL REFERENCES "Employee"(id),
				"leaderInvoiceId" TEXT NOT NULL REFERENCES "LeaderInvoice"(id) ON DELETE CASCADE,
				role TEXT NOT NULL CHECK (role IN ('LEADER', 'TECHNICIAN')),
				"executionCommission" NUMERIC NOT NULL DEFAULT 0,
				"salesCommission" NUMERIC NOT NULL DEFAULT 0,
				"totalCommission" NUMERIC NOT NULL DEFAULT 0,
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS "EmployeeCommission_employeeId_idx" ON "EmployeeCommission"("employeeId");
			CREATE INDEX IF NOT EXISTS "EmployeeCommission_leaderInvoiceId_idx" ON "EmployeeCommission"("leaderInvoiceId");
			CREATE INDEX IF NOT EXISTS "EmployeeCommission_createdAt_idx" ON "EmployeeCommission"("createdAt")`,
		},
	}
}
