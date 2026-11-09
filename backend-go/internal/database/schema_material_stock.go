package database

// رصيد المواد — بند ٢٠ بالفحص الشامل. المخزن يسجّل الرصيد (جرد)، وماتركس
// يطرح المصروف بفواتير الليدرية من بعد الجرد، فيعرف «يخلص خلال كم يوم».
func materialStockMigrations() []Migration {
	return []Migration{{
		Version: "0304_material_stock",
		SQL: `
			ALTER TABLE "Material" ADD COLUMN IF NOT EXISTS "stockQty" DOUBLE PRECISION;
			ALTER TABLE "Material" ADD COLUMN IF NOT EXISTS "stockCountedAt" TIMESTAMPTZ;
			ALTER TABLE "Material" ADD COLUMN IF NOT EXISTS "stockCountedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL;
		`,
	}}
}
