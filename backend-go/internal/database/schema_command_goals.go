package database

// ═══ مركز القيادة: الأهداف والخطة (قرار (ع) 10-10) ═══
// بيانات جديدة ما موجودة بنظام الشركة: أهداف المالك للفصل/السنة، ولكل هدف
// مراحل بموعد. «المسؤول» نص حر — الطبقتين ما يتشاركن بيانات الموظفين.
func commandGoalsMigrations() []Migration {
	return []Migration{{
		Version: "0340_command_goals",
		SQL: `
			CREATE TABLE IF NOT EXISTS "CommandGoal" (
				id TEXT PRIMARY KEY,
				title TEXT NOT NULL,
				description TEXT,
				category TEXT NOT NULL DEFAULT 'عام',
				"ownerLabel" TEXT,
				"targetDate" DATE,
				status TEXT NOT NULL DEFAULT 'ACTIVE',
				"createdById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			CREATE TABLE IF NOT EXISTS "CommandGoalStep" (
				id TEXT PRIMARY KEY,
				"goalId" TEXT NOT NULL REFERENCES "CommandGoal"(id) ON DELETE CASCADE,
				title TEXT NOT NULL,
				"dueDate" DATE,
				"doneAt" TIMESTAMP,
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS "CommandGoalStep_goal" ON "CommandGoalStep"("goalId");
		`,
	}}
}
