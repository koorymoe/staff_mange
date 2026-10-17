package database

// ═══ شروط وأحكام عرض السعر — طلب (ع) 10-08 ═══
// كانت خمس أسطر ثابتة بالواجهة. هسه جدول ينعدل منه (صاحب صلاحية عرض السعر
// يضيف/يعدّل/يحذف/يرتّب)، ويتعبّى أول مرة بنفس الأسطر القديمة.
func quotationTermsMigrations() []Migration {
	return []Migration{
		{
			Version: "0333_quotation_terms",
			SQL: `
				CREATE TABLE IF NOT EXISTS "QuotationTerm" (
					id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
					text        TEXT NOT NULL,
					"sortOrder" INT NOT NULL DEFAULT 0,
					"updatedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					"updatedAt" TIMESTAMP NOT NULL DEFAULT (now() AT TIME ZONE 'UTC')
				);
				INSERT INTO "QuotationTerm" (text, "sortOrder")
				SELECT t, o FROM (VALUES
					('الأسعار المذكورة أعلاه لا تشمل أجور النقل والتركيب ما لم يُذكر خلاف ذلك.', 1),
					('عرض السعر ساري المفعول لمدة 15 يوم من تاريخه.', 2),
					('الدفع: 50% مقدم والباقي عند التسليم.', 3),
					('مدة التنفيذ تبدأ من تاريخ استلام الدفعة الأولى.', 4),
					('الأسعار قابلة للتغيير حسب تقلبات السوق.', 5)
				) v(t, o) WHERE NOT EXISTS (SELECT 1 FROM "QuotationTerm");`,
		},
	}
}
