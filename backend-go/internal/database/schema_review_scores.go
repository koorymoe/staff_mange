package database

// ═══ درجات التقييم الثلاث ═══
//
// التقييم كان قيمة وحدة: «إيجابي» أو «يحتاج تدريب» أو مخالفة.
// وهاي تجاوب سؤال واحد بس — «زين لو لا؟» — وما تگول **وين** بالضبط.
//
// فني يوصل متأخر بس شغله ممتاز، وفني يوصل بوقته بس شغله ناقص:
// الاثنين ينكتبون «يحتاج تدريب»، وبعد شهر ماكو بالسجل شي يفرّقهم
// ولا شي يگول شنو التدريب المطلوب.
//
// ثلاث درجات (١-٥) تفصلهن:
//
//	commitmentScore  الالتزام     — يجي بوقته؟ يلتزم بالتعليمات؟
//	speedScore       السرعة       — يخلص بوقت معقول؟
//	qualityScore     جودة التنفيذ — الشغل نفسه شلون طلع؟
//
// ⚠️ كلهن **اختيارية** (NULL مسموح) عن قصد. الليدر راجع من شغل ميداني
// وبيده تلفون؛ لو أجبرناه على ١٥ ضغطة نجوم لثلاثة فنيين راح ما يقيّم
// أبداً — وصفر تقييمات أسوأ من تقييم بلا نجوم. الحكم السريع
// (إيجابي/يحتاج تدريب) يبقى كافي لوحده، والنجوم تفصيل يزيده الي عنده
// وقت.
//
// والقيد بقاعدة البيانات مو بالكود بس: درجة ٧ أو ٠ تنكتب بأمر SQL
// مباشر تخرب كل متوسط ينحسب منها.
func reviewScoreMigrations() []Migration {
	return []Migration{
		{
			Version: "0245_review_scores",
			SQL: `
				ALTER TABLE "PerformanceReview"
					ADD COLUMN IF NOT EXISTS "commitmentScore" SMALLINT,
					ADD COLUMN IF NOT EXISTS "speedScore" SMALLINT,
					ADD COLUMN IF NOT EXISTS "qualityScore" SMALLINT;

				DO $$ BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM pg_constraint WHERE conname = 'performance_review_scores_range'
					) THEN
						ALTER TABLE "PerformanceReview"
							ADD CONSTRAINT performance_review_scores_range CHECK (
								("commitmentScore" IS NULL OR "commitmentScore" BETWEEN 1 AND 5) AND
								("speedScore"      IS NULL OR "speedScore"      BETWEEN 1 AND 5) AND
								("qualityScore"    IS NULL OR "qualityScore"    BETWEEN 1 AND 5)
							);
					END IF;
				END $$;
			`,
		},
	}
}
