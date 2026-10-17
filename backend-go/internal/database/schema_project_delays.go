package database

// ═══ تأخير المشاريع (قرار (ع) 10-10) ═══
// «ليش المشرف أو التقني أو المهندس يتأخر؟ … ينحطله لمت على كل مرحلة».
// الحد: ماتركس يتعلّمه من المشاريع، والمالك يگدر يثبّته بإيده (ProjectStageLimit).
// والي يتجاوز يكتب السبب (ProjectDelayReason) والمراقب يشوفه.
func projectDelayMigrations() []Migration {
	return []Migration{{
		Version: "0341_project_delays",
		SQL: `
			CREATE TABLE IF NOT EXISTS "ProjectStageLimit" (
				stage TEXT PRIMARY KEY,
				days INT NOT NULL CHECK (days > 0),
				"updatedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				"updatedAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			CREATE TABLE IF NOT EXISTS "ProjectDelayReason" (
				id TEXT PRIMARY KEY,
				"projectId" TEXT NOT NULL REFERENCES "Project"(id) ON DELETE CASCADE,
				stage TEXT NOT NULL,
				"employeeId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				reason TEXT NOT NULL,
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS "ProjectDelayReason_project" ON "ProjectDelayReason"("projectId", "createdAt");
		`,
	}}
}
