package database

// ═══ سجل توقع ماتركس للإيراد (قرار (ع) 10-07: «التوقع يكون قريب من النتائج») ═══
// كل يوم ينحفظ التوقع، وبنهاية الشهر نقارنه بالنتيجة الفعلية — حتى ماتركس
// يعرف شكد چان دقيق، والمدير يشوف الدقة بعينه مو بالثقة.
func forecastLogMigrations() []Migration {
	return []Migration{{
		Version: "0329_forecast_log",
		SQL: `
			CREATE TABLE IF NOT EXISTS "MatrixForecastLog" (
				day       DATE PRIMARY KEY,
				month     TEXT NOT NULL,
				expected  DOUBLE PRECISION NOT NULL,
				pipeline  DOUBLE PRECISION NOT NULL,
				pace      DOUBLE PRECISION,
				mtd       DOUBLE PRECISION NOT NULL,
				"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now()
			);
			CREATE INDEX IF NOT EXISTS "MatrixForecastLog_month_idx" ON "MatrixForecastLog"(month);
		`,
	}}
}
