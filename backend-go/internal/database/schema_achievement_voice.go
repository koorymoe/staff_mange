package database

// فويس مسج للإنجاز — طلب (ع) 10-05. الصوت ينحفظ بتخزين الملفات عدنا بس،
// وما يطلع لأي مزوّد خارجي (قاعدة البيانات الشخصية) لحد ما (ع) يقرر.
// جدول مستقل حتى ما نلمس "Achievement" (SELECT *).
func achievementVoiceMigrations() []Migration {
	return []Migration{{
		Version: "0314_achievement_voice",
		SQL: `CREATE TABLE IF NOT EXISTS "AchievementVoice" (
			"achievementId" TEXT PRIMARY KEY REFERENCES "Achievement"(id) ON DELETE CASCADE,
			"fileKey"       TEXT NOT NULL,
			"contentType"   TEXT NOT NULL,
			seconds         INTEGER NOT NULL DEFAULT 0,
			"createdAt"     TIMESTAMPTZ NOT NULL DEFAULT now()
		);`,
	}}
}
