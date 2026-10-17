package database

// ═══ التقييم مو إيجابي/سلبي بس ═══
//
// كان عدنا قيمتين: POSITIVE و NEGATIVE. والسلبي يعني شي واحد —
// «يحتاج تدريب» — فيتحوّل الموظف متدرب تلقائياً.
//
// وهاي تخلط شغلتين مختلفات تماماً:
//
//	الفني ما يعرف يشتغل  → نقص **مهارة**، علاجه تدريب
//	الفني أسلوبه غلط      → مخالفة **سلوك**، علاجها إجراء إداري
//
// حطهن بخانة وحدة يعني: واحد أسلوبه سيّئ ينتصنّف «يحتاج تدريب»
// وينزل بدورة فنية ما تعالج شي، وواحد ناقصه مهارة ينحسب مخالف.
// والاثنين ينظلمون.
//
// الحالات الجديدة:
//
//	POSITIVE        شغل زين
//	NEEDS_TRAINING  نقص مهارة → يتحوّل متدرب (نفس سلوك NEGATIVE)
//	MISCONDUCT      مخالفة سلوك → بلاغ للإدارة
//	COMMITMENT      تأخر أو التزام بالمواعيد → بلاغ للإدارة
//
// ⚠️ NEGATIVE القديمة تنتحوّل لـNEEDS_TRAINING: هذا **بالضبط** الي
// كانت تسويه (SetTrainee)، فالتحويل يحفظ المعنى الأصلي بدل ما
// يخمّن. تحويلها لـMISCONDUCT يتهم موظفين بمخالفات ما ارتكبوها.
func reviewKindMigrations() []Migration {
	return []Migration{
		{
			Version: "0244_review_kinds",
			SQL: `
				UPDATE "PerformanceReview"
				SET rating = 'NEEDS_TRAINING'
				WHERE rating = 'NEGATIVE';

				-- ⚠️ القيد بقاعدة البيانات مو بالكود بس: قيمة غريبة
				-- تنكتب بأمر SQL مباشر تخرب كل شاشة تعرض التسمية.
				DO $$ BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM pg_constraint WHERE conname = 'performance_review_rating_valid'
					) THEN
						ALTER TABLE "PerformanceReview"
							ADD CONSTRAINT performance_review_rating_valid
							CHECK (rating IN ('POSITIVE','NEEDS_TRAINING','MISCONDUCT','COMMITMENT'));
					END IF;
				END $$;
			`,
		},
	}
}
