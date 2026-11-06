package database

// agentGrowthMigrations (دازاي 11-06): «الوكلاء يطورون نفسهم ويطلبون تحديثات».
// قياس الأدوات، والأسئلة الي ما انجاوبت، و👍/👎 على الأجوبة، ودروس لكل وكيل ويا
// دروس يقترحها الوكيل نفسه. مخطط بس (قرار ٠٠٠٩).
func agentGrowthMigrations() []Migration {
	return []Migration{{
		Version: "0446_agent_tool_stat",
		SQL: `
			CREATE TABLE IF NOT EXISTS "AgentToolStat" (
				day        DATE NOT NULL,
				agent      TEXT NOT NULL,
				tool       TEXT NOT NULL,
				calls      INT NOT NULL DEFAULT 0,
				failures   INT NOT NULL DEFAULT 0,
				"totalMs"  BIGINT NOT NULL DEFAULT 0,
				"lastError" TEXT,
				PRIMARY KEY (day, agent, tool)
			);
		`,
	}, {
		Version: "0447_agent_gap",
		SQL: `
			CREATE TABLE IF NOT EXISTS "AgentGap" (
				id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
				agent TEXT NOT NULL,
				kind TEXT NOT NULL,
				missing TEXT NOT NULL DEFAULT '',
				"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now()
			);
			CREATE INDEX IF NOT EXISTS "AgentGap_kind_idx" ON "AgentGap" (agent, kind, "createdAt" DESC);
		`,
	}, {
		Version: "0448_agent_feedback",
		SQL: `
			CREATE TABLE IF NOT EXISTS "AgentFeedback" (
				id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
				agent TEXT NOT NULL,
				"messageId" TEXT NOT NULL,
				"employeeId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				good BOOLEAN NOT NULL,
				note TEXT NOT NULL DEFAULT '',
				"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now(),
				UNIQUE ("messageId", "employeeId")
			);
		`,
	}, {
		Version: "0449_lesson_agent",
		SQL: `
			ALTER TABLE "MatrixLesson" ADD COLUMN IF NOT EXISTS agent TEXT NOT NULL DEFAULT 'MATRIX';
			ALTER TABLE "MatrixLesson" ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'ACTIVE'
				CHECK (status IN ('ACTIVE', 'PROPOSED'));
		`,
	}}
}
