package database

// «اسأل ماتركس» — سجل كل سؤال وجواب للمراجعة (المرحلة ١: للمدير والمالك بس).
func matrixChatMigrations() []Migration {
	return []Migration{{
		Version: "0303_matrix_chat",
		SQL: `
			CREATE TABLE IF NOT EXISTS "MatrixChatMessage" (
				id TEXT PRIMARY KEY,
				"employeeId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				question TEXT NOT NULL,
				answer TEXT NOT NULL,
				source TEXT NOT NULL DEFAULT 'RULES',
				"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now()
			);
			CREATE INDEX IF NOT EXISTS "MatrixChatMessage_emp_idx" ON "MatrixChatMessage"("employeeId", "createdAt");
		`,
	}}
}
