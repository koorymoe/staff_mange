package database

// تقييم الليدر لفنيّيه بعد كل حجز — طلب (ع) 10-05 («ماتركس ٢٠٥٠»).
// ماتركس يذكّر الليدر لحد ما يقيّم. التقييم يطلع بتقرير الفني بس — ماكو نقاط.
func crewRatingMigrations() []Migration {
	return []Migration{{
		Version: "0312_crew_rating",
		SQL: `CREATE TABLE IF NOT EXISTS "CrewRating" (
			id             TEXT PRIMARY KEY,
			"bookingId"    TEXT NOT NULL REFERENCES "Booking"(id) ON DELETE CASCADE,
			"leaderId"     TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
			"technicianId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
			score          INTEGER NOT NULL CHECK (score BETWEEN 1 AND 5),
			note           TEXT,
			"createdAt"    TIMESTAMPTZ NOT NULL DEFAULT now(),
			UNIQUE ("bookingId", "technicianId")
		);
		CREATE INDEX IF NOT EXISTS "CrewRating_technician_idx" ON "CrewRating" ("technicianId", "createdAt" DESC);`,
	}}
}
