package database

// ═══ 🧠 تحليل ماتركس للتقرير (11-15: «اريد ماتركس يقرا كلشي ويحلل كلشي ويشوف كلشي») ═══
// كل تقرير إنجاز يومي ينحلّل بالذكاء مقابل كل شي سجّله النظام للموظف ذاك اليوم.
// استشاري بس — ما يغيّر التقييم. bodyHash حتى ما يعيد التحليل إلا إذا التقرير تغيّر.
func dailyReportAnalysisMigrations() []Migration {
	return []Migration{{
		Version: "0499_daily_report_analysis",
		SQL: `
			CREATE TABLE IF NOT EXISTS "DailyReportAnalysis" (
				id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
				"reportId" TEXT NOT NULL REFERENCES "DailyWorkReport"(id) ON DELETE CASCADE,
				"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				day DATE NOT NULL,
				verdict TEXT NOT NULL CHECK (verdict IN ('CONSISTENT', 'GAPS', 'CONTRADICTS', 'NO_EVIDENCE')),
				summary TEXT NOT NULL DEFAULT '',
				claims JSONB NOT NULL DEFAULT '[]'::jsonb,
				missing JSONB NOT NULL DEFAULT '[]'::jsonb,
				notes TEXT NOT NULL DEFAULT '',
				"bodyHash" TEXT NOT NULL,
				"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now(),
				"updatedAt" TIMESTAMPTZ NOT NULL DEFAULT now(),
				CONSTRAINT "DailyReportAnalysis_reportId_key" UNIQUE ("reportId")
			);
			CREATE INDEX IF NOT EXISTS "DailyReportAnalysis_day_idx" ON "DailyReportAnalysis" (day);
		`,
	}}
}
