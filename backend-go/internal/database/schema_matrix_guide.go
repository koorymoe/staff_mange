package database

// ═══ تعليمات ماتركس («الخطوات التعليمية») ═══
//
// قرار (ع): ماتركس يوجّه الموظف ويا أي إجراء، والتعليمات **مو بالكود**:
// تنحفظ هنا والمدير يضيف ويعدّل بنفسه. وإذا مفتاح النموذج موجود، النموذج
// ياخذها كتعليمات ويكتب التوجيه بنفسه حسب الموقف.
//
// route: بادئة المسار ('/coordinator'). match: كلمات بنص الزر مفصولة بـ|
// (فاضي = أي زر). groups: مجموعات الأدوار (فاضي = الكل). text: التوجيه،
// ويقبل {left} و{work} — يتعوّضن بأرقام الموظف الحقيقية.
func matrixGuideMigrations() []Migration {
	return []Migration{
		{
			Version: "0300_matrix_guide_rule",
			SQL: `
				CREATE TABLE IF NOT EXISTS "MatrixGuideRule" (
					id          TEXT PRIMARY KEY,
					route       TEXT NOT NULL,
					match       TEXT NOT NULL DEFAULT '',
					groups      TEXT NOT NULL DEFAULT '',
					text        TEXT NOT NULL,
					"onlyIfPending" BOOLEAN NOT NULL DEFAULT false,
					priority    INT NOT NULL DEFAULT 0,
					enabled     BOOLEAN NOT NULL DEFAULT true,
					"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now()
				);
				INSERT INTO "MatrixGuideRule" (id, route, match, groups, text, "onlyIfPending", priority) VALUES
				 ('seed-leave', '/attendance', 'إجازة|إرسال الطلب', '', 'عندك {work}. كمّل شغلك أول، وبعدها اطلب الإجازة.', true, 10),
				 ('seed-confirm', '/coordinator', 'تثبيت|تأكيد|ثبّت', '', 'تواصل ويا الزبون أول وتأكد من الموعد والعنوان. إذا توصلت اضغط «تم»، وإذا ما رد سجّلها «الزبون ما رد».', false, 10),
				 ('seed-survey', '/coordinator', 'كشف', '', 'الكشف يروحله ليدر متفرّغ. تأكد إنه يرفق صور الموقع بالتقرير.', false, 9),
				 ('seed-sales', '/sales', 'حجز|حفظ|إنشاء', '', 'تأكد من رقم الزبون والعنوان ونوع الخدمة قبل الحفظ. الحجز الناقص يرجعلك.', false, 10),
				 ('seed-invoice', '/leader-invoices/new', '', '', 'اكتب المواد بالكمية الي انصرفت فعلاً. ماتركس يقارن كل مادة بمعدل نفس الخدمة.', false, 8),
				 ('seed-audit', '/daily-audit', 'تدقيق|دقق|طابق', '', 'طابق المبلغ ويا الوصل قبل ما تأشّر. باقي عليك {left}.', false, 10),
				 ('seed-desk', '/monitor-desk', '', '', 'كل صف قرار: سليم أو عندك ملاحظة. لا تتركه معلّق. باقي عليك {left}.', false, 5),
				 ('seed-expense', '/expenses', '', '', 'كل مصروف لازم وياه وصل. بلا وصل ينرفض.', false, 5),
				 ('seed-any', '/', '', '', 'عندك {work}. كمّل شغلك أول.', true, 0)
				ON CONFLICT (id) DO NOTHING;
			`,
		},
		{
			// ماتركس يقترح ويتعلّم: اقتراحات تنتظر قرار المدير، وعدّاد استعمال
			// لكل تعليمة (يغذّي اقتراح وقف التعليمات الي ما تنفع).
			Version: "0301_matrix_proposal",
			SQL: `
				ALTER TABLE "MatrixGuideRule" ADD COLUMN IF NOT EXISTS hits INT NOT NULL DEFAULT 0;
				ALTER TABLE "MatrixGuideRule" ADD COLUMN IF NOT EXISTS "lastHitAt" TIMESTAMPTZ;
				ALTER TABLE "MatrixGuideRule" ADD COLUMN IF NOT EXISTS source TEXT NOT NULL DEFAULT 'ADMIN';
				CREATE TABLE IF NOT EXISTS "MatrixProposal" (
					id            TEXT PRIMARY KEY,
					kind          TEXT NOT NULL,
					title         TEXT NOT NULL,
					rationale     TEXT NOT NULL DEFAULT '',
					evidence      JSONB,
					payload       JSONB,
					signature     TEXT NOT NULL,
					source        TEXT NOT NULL DEFAULT 'RULES',
					status        TEXT NOT NULL DEFAULT 'PENDING',
					"decidedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					"decidedAt"   TIMESTAMPTZ,
					note          TEXT,
					"createdAt"   TIMESTAMPTZ NOT NULL DEFAULT now()
				);
				CREATE UNIQUE INDEX IF NOT EXISTS "MatrixProposal_pending_sig"
					ON "MatrixProposal" (signature) WHERE status = 'PENDING';
				CREATE INDEX IF NOT EXISTS "MatrixProposal_status_idx" ON "MatrixProposal" (status, "createdAt" DESC);
			`,
		},
	}
}
