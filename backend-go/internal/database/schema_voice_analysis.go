package database

// تحليل الفويس — قرار (ع) 10-05. الصوت يتحول لنص **داخل سيرفرنا** (Whisper
// بحاوية محلية) وما يطلع أبداً. النص بس (بعد شيل الأسماء) يروح لهايكو:
// ملخّص، الشغل المعدود، طريقة الكلام، والعبارات العراقية — والأخيرة تتجمع
// بقاموس يتعلّم منه ماتركس اللهجة.
func voiceAnalysisMigrations() []Migration {
	return []Migration{{
		Version: "0315_voice_analysis",
		SQL: `
			ALTER TABLE "AchievementVoice"
				ADD COLUMN IF NOT EXISTS "analysisStatus" TEXT NOT NULL DEFAULT 'PENDING',
				ADD COLUMN IF NOT EXISTS attempts INTEGER NOT NULL DEFAULT 0,
				ADD COLUMN IF NOT EXISTS transcript TEXT,
				ADD COLUMN IF NOT EXISTS summary TEXT,
				ADD COLUMN IF NOT EXISTS tasks JSONB,
				ADD COLUMN IF NOT EXISTS style TEXT,
				ADD COLUMN IF NOT EXISTS "analysisError" TEXT,
				ADD COLUMN IF NOT EXISTS "analyzedAt" TIMESTAMPTZ;

			CREATE TABLE IF NOT EXISTS "IraqiPhrase" (
				phrase      TEXT PRIMARY KEY,
				meaning     TEXT NOT NULL,
				uses        INTEGER NOT NULL DEFAULT 1,
				"firstSeen" TIMESTAMPTZ NOT NULL DEFAULT now(),
				"lastSeen"  TIMESTAMPTZ NOT NULL DEFAULT now()
			);`,
	}}
}
