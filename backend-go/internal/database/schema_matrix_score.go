package database

// ═══ تقييم ماتركس + تقييمات البشر بالسلسلة (قرار (ع) 10-05) ═══
//
//   - MatrixScore: كل نقطة يعطيها ماتركس بسببها — لكل محطة حجز، لكل يوم دوام،
//     ولكل أسبوع مشروع. ⚠️ درجة تقييم بس: ماكو فلوس ولا تلمس نقاط الانضباط.
//     الفريد يمنع التكرار لو الدورة انعادت.
//   - StaffRating: تقييم بشري ١–٥ بمحطته (الإداري يقيّم الليدر، المراقب
//     بالتدقيق، الجودة بعد الاتصال، والمراقب دورياً للمكاتب).
//   - التقييم النهائي = ٠٫٦ ماتركس + ٠٫٤ البشر (يتحسب وقت العرض).
func matrixScoreMigrations() []Migration {
	return []Migration{{
		Version: "0321_matrix_score",
		SQL: `
			CREATE TABLE IF NOT EXISTS "MatrixScore" (
				id              TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
				"employeeId"    TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				source          TEXT NOT NULL,             -- BOOKING | DAY | PROJECT
				"sourceId"      TEXT NOT NULL,
				"sourceLabel"   TEXT,
				rule            TEXT NOT NULL,
				points          INTEGER NOT NULL,
				"maxPoints"     INTEGER NOT NULL,
				reason          TEXT NOT NULL,
				"at"            TIMESTAMPTZ NOT NULL,      -- وقت الحدث نفسه (للشهر)
				"createdAt"     TIMESTAMPTZ NOT NULL DEFAULT now(),
				"cancelledAt"   TIMESTAMPTZ,
				"cancelledById" TEXT REFERENCES "Employee"(id),
				"cancelNote"    TEXT,
				UNIQUE ("employeeId", source, "sourceId", rule)
			);
			CREATE INDEX IF NOT EXISTS "MatrixScore_emp_at_idx" ON "MatrixScore"("employeeId", "at");
			CREATE TABLE IF NOT EXISTS "StaffRating" (
				id           TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
				"raterId"    TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				"rateeId"    TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				stage        TEXT NOT NULL,              -- COORD_LEADER | AUDIT | QUALITY_CALL | MONITOR_PERIODIC
				"bookingId"  TEXT NOT NULL DEFAULT '',   -- '' للدوري
				period       TEXT NOT NULL DEFAULT '',   -- الدوري: «2026-10-A»
				score        INTEGER NOT NULL CHECK (score BETWEEN 1 AND 5),
				note         TEXT,
				"createdAt"  TIMESTAMPTZ NOT NULL DEFAULT now(),
				UNIQUE ("raterId", "rateeId", stage, "bookingId", period)
			);
			CREATE INDEX IF NOT EXISTS "StaffRating_ratee_idx" ON "StaffRating"("rateeId", "createdAt");
		`,
	}}
}
