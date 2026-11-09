package database

// ═══ موردين التقنيين (قرار (ع) 10-06) ═══
// قائمة موردين ثانية غير موردين الشركة: التقنيين ومسؤولي الخدمات يضيفون،
// والمدير يختار لكل تقني شنو الموردين الي يطلعوله. المضاف ما يطلع لأحد —
// حتى لصاحبه — لحد ما المدير يختاره إله.
func techSupplierMigrations() []Migration {
	return []Migration{{
		Version: "0325_tech_suppliers",
		SQL: `
			CREATE TABLE IF NOT EXISTS "TechSupplier" (
				id            TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
				"companyName" TEXT NOT NULL,
				"ownerName"   TEXT,
				phone         TEXT NOT NULL,
				address       TEXT,
				"locationUrl" TEXT,
				specialty     TEXT,
				notes         TEXT,
				"createdById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				"createdAt"   TIMESTAMPTZ NOT NULL DEFAULT now(),
				"updatedAt"   TIMESTAMPTZ NOT NULL DEFAULT now()
			);
			CREATE TABLE IF NOT EXISTS "TechSupplierAccess" (
				"supplierId" TEXT NOT NULL REFERENCES "TechSupplier"(id) ON DELETE CASCADE,
				"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				"createdAt"  TIMESTAMPTZ NOT NULL DEFAULT now(),
				PRIMARY KEY ("supplierId", "employeeId")
			);
			CREATE INDEX IF NOT EXISTS "TechSupplierAccess_emp_idx" ON "TechSupplierAccess"("employeeId");`,
	}}
}
