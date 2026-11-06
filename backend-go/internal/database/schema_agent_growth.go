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
	}, {
		// (دازاي 11-06) خطة الوحدة: ماتركس يكتب المسودة، والبنود المعتمدة تصير مهام للمدير والنائب
		Version: "0450_org_plan_item_task",
		SQL: `
			ALTER TABLE "OrgPlanItem" ADD COLUMN IF NOT EXISTS "taskId" TEXT REFERENCES "ExtraTask"(id) ON DELETE SET NULL;
			ALTER TABLE "OrgPlanItem" ADD COLUMN IF NOT EXISTS "byMatrix" BOOLEAN NOT NULL DEFAULT false;
			ALTER TABLE "OrgPlanItem" ADD COLUMN IF NOT EXISTS reason TEXT;
		`,
	}, {
		// امتحان الوكلاء (دازاي 11-06، فكرة ٥): نتيجة كل جولة أسئلة ثابتة.
		Version: "0453_agent_eval_run",
		SQL: `
			CREATE TABLE IF NOT EXISTS "AgentEvalRun" (
				id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
				agent TEXT NOT NULL,
				build TEXT NOT NULL,
				total INT NOT NULL,
				passed INT NOT NULL,
				details JSONB NOT NULL DEFAULT '[]'::jsonb,
				"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now()
			);
			CREATE INDEX IF NOT EXISTS "AgentEvalRun_agent_idx" ON "AgentEvalRun" (agent, "createdAt" DESC);
		`,
	}}
}
