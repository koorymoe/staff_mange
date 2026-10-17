package database

// ═══ ورق الأعمال داخل الشركة (قرار (ع) 10-07) ═══
// الإداري المسؤول عن الحجوزات هو الي يسوي فاتورة وتقرير العمل الداخلي بعد
// ما يخلص الفني — فإداريّو الحجوزات الموجودين ياخذون صلاحية الفاتورة الداخلية.
func internalPaperworkMigrations() []Migration {
	return []Migration{{
		Version: "0331_internal_paperwork",
		SQL: `
			INSERT INTO "EmployeePermission" (id, "employeeId", "permissionId")
			SELECT gen_random_uuid()::text, e.id, p.id
			FROM "Employee" e CROSS JOIN "Permission" p
			WHERE e.role::text = 'HR_COORDINATOR' AND p.name = 'invoice_internal'
			  AND NOT EXISTS (SELECT 1 FROM "EmployeePermission" x WHERE x."employeeId" = e.id AND x."permissionId" = p.id);
		`,
	}}
}
