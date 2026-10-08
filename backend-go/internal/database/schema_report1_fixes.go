package database

// ═══ تقرير ١ (10-08) — B19 ═══
// quotation_system مرادف قديم لـquotation_manage_all وما ينمنح بعد. الممنوحين
// ياخذون الجديدة حتى ما يعتمدون على القراءة القديمة (القديمة تبقى احتياط، ما نحذف).
func report1FixMigrations() []Migration {
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
	}}
}
