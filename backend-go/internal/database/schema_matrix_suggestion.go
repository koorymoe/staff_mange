package database

// ماتركس يقترح والإنسان يقرر — المرحلة الثانية (قرار (ع) 10-05).
// كل اقتراح (موعد أو كادر) ينحفظ، وبعدين ماتركس يقارنه بالي اختاره المنسق
// فعلاً، وبشنو صار بالحجز — حتى تنقاس دقته قبل ما ينطى أي صلاحية تنفيذ.
func matrixSuggestionMigrations() []Migration {
	return []Migration{{
		Version: "0317_matrix_suggestion",
		SQL: `CREATE TABLE IF NOT EXISTS "MatrixSuggestion" (
			id           TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
			"bookingId"  TEXT NOT NULL REFERENCES "Booking"(id) ON DELETE CASCADE,
			kind         TEXT NOT NULL CHECK (kind IN ('SCHEDULE', 'CREW')),
			suggested    JSONB NOT NULL,
			reason       TEXT NOT NULL DEFAULT '',
			outcome      TEXT NOT NULL DEFAULT 'PENDING' CHECK (outcome IN ('PENDING', 'ACCEPTED', 'PARTIAL', 'CHANGED')),
			actual       JSONB,
			"createdAt"  TIMESTAMPTZ NOT NULL DEFAULT now(),
			"decidedAt"  TIMESTAMPTZ,
			UNIQUE ("bookingId", kind)
		);
		CREATE INDEX IF NOT EXISTS "MatrixSuggestion_outcome_idx" ON "MatrixSuggestion" (outcome, "createdAt");`,
	}}
}
