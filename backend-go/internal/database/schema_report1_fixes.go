package database

// ═══ تقرير ١ (10-08) — B19 ═══
// quotation_system مرادف قديم لـquotation_manage_all وما ينمنح بعد. الممنوحين
// ياخذون الجديدة حتى ما يعتمدون على القراءة القديمة (القديمة تبقى احتياط، ما نحذف).
func report1FixMigrations() []Migration {
	grant := func(role, perm string) string {
		return `INSERT INTO "EmployeePermission" (id, "employeeId", "permissionId")
			SELECT gen_random_uuid()::text, e.id, p.id FROM "Employee" e CROSS JOIN "Permission" p
			WHERE e.role::text = '` + role + `' AND p.name = '` + perm + `'
			  AND NOT EXISTS (SELECT 1 FROM "EmployeePermission" x WHERE x."employeeId" = e.id AND x."permissionId" = p.id);`
	}
	return []Migration{{
		Version: "0334_quotation_system_to_manage_all",
		SQL: `
			INSERT INTO "EmployeePermission" (id, "employeeId", "permissionId")
			SELECT gen_random_uuid()::text, ep."employeeId", pn.id
			FROM "EmployeePermission" ep
			JOIN "Permission" po ON po.id = ep."permissionId" AND po.name = 'quotation_system'
			JOIN "Permission" pn ON pn.name = 'quotation_manage_all'
			WHERE NOT EXISTS (SELECT 1 FROM "EmployeePermission" x WHERE x."employeeId" = ep."employeeId" AND x."permissionId" = pn.id);
		`,
	}, {
		// تقرير ٢: المحاسب يحتاج وحدة الحسابات حتى يشوف شاشات التدقيق،
		// والعلاقات العامة يضيفون شخصيات مهمة (قرار (ع) 10-08).
		Version: "0335_report2_role_defaults",
		SQL:     grant("FINANCE", "unit_finance") + grant("PUBLIC_RELATIONS", "vip_manual_add"),
	}}
}
