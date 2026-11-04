package database

// ═══ دور «مطوّر الخدمات» (قرار (ع) 10-10) ═══
// يكتشفون خدمات جديدة ويطوّرون الموجودة — شغلهم دراسات الخدمات، مو شغل التقني.
func serviceDeveloperMigrations() []Migration {
	return []Migration{{Version: "0340_employee_role_service_developer", SQL: `ALTER TYPE "EmployeeRole" ADD VALUE IF NOT EXISTS 'SERVICE_DEVELOPER';`}}
}
