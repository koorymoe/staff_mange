package database

// عدّاد كلفة الذكاء الاصطناعي — طلب (ع) 10-05: «اريد اقل سعر ممكن». كل نداء
// لهايكو ينحسب هنا (توكنات الداخل والطالع والمحفوظ) لكل ميزة لكل يوم، حتى
// الكلفة الحقيقية تبين بمركز قيادة ماتركس بدل التقدير.
func aiUsageMigrations() []Migration {
	return []Migration{{
		Version: "0318_ai_usage",
		SQL: `CREATE TABLE IF NOT EXISTS "AiUsage" (
			day            DATE NOT NULL,
			feature        TEXT NOT NULL,
			model          TEXT NOT NULL DEFAULT '',
			calls          INTEGER NOT NULL DEFAULT 0,
			"inputTokens"  BIGINT NOT NULL DEFAULT 0,
			"outputTokens" BIGINT NOT NULL DEFAULT 0,
			"cacheRead"    BIGINT NOT NULL DEFAULT 0,
			"cacheWrite"   BIGINT NOT NULL DEFAULT 0,
			PRIMARY KEY (day, feature, model)
		);`,
	}}
}
