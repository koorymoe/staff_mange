package database

// ═══ سجل أفعال ماتركس ═══
//
// ماتركس صار ينفّذ أشياء بسيطة لحاله (تذكيرات وتنبيهات بس — بلا مال
// وبلا عقوبة). كل فعل ينكتب هنا قبل ما ينفّذ، حتى المدير يشوف «شنو
// سوّى» بصندوق القرارات ويگدر يرفضه.
//
// الفريد (kind, entityId, period): نفس الفعل على نفس الشي ما يتكرر
// بنفس الفترة (يوم أو أسبوع حسب النوع).
func aiActionMigrations() []Migration {
	return []Migration{
		{
			Version: "0298_ai_action",
			SQL: `
				CREATE TABLE IF NOT EXISTS "AiAction" (
					id                 TEXT PRIMARY KEY,
					kind               TEXT NOT NULL,
					"entityType"       TEXT NOT NULL,
					"entityId"         TEXT NOT NULL,
					period             TEXT NOT NULL,
					"targetEmployeeId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					"targetLabel"      TEXT NOT NULL DEFAULT '',
					summary            TEXT NOT NULL,
					status             TEXT NOT NULL DEFAULT 'DONE',
					"createdAt"        TIMESTAMPTZ NOT NULL DEFAULT now(),
					"undoneById"       TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					"undoneAt"         TIMESTAMPTZ,
					UNIQUE (kind, "entityId", period)
				);
				CREATE INDEX IF NOT EXISTS "AiAction_createdAt_idx" ON "AiAction" ("createdAt" DESC);
			`,
		},
	}
}
