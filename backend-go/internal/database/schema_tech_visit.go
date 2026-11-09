package database

// ═══ قرار التقني بعد التواصل (قرار (ع) 10-07) ═══
// بعد ما يحچي ويا الزبون يحدد: انحلّت بالتلفون، لو يحتاج كشف/زيارة ويحدد
// موعد ويطلع. كل خطوة بوقتها حتى المراقب يفتهم وماتركس يحاسب.
func techVisitMigrations() []Migration {
	return []Migration{{
		Version: "0330_tech_visit",
		SQL: `
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "techDecision" TEXT;
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "techDecidedAt" TIMESTAMP;
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "techVisitAt" TIMESTAMP;
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "techVisitedAt" TIMESTAMP;
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "techVisitMoves" INTEGER NOT NULL DEFAULT 0;
		`,
	}}
}
