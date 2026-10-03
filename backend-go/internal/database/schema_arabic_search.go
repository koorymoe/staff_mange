package database

// ══════════════════════════════════════════════════════════════════
// البحث العربي بالسيرفر — نفس تطبيع الواجهة بالضبط
// ══════════════════════════════════════════════════════════════════
//
// الواجهة تطبّع النص قبل المقارنة (frontend/src/utils/search.ts) حتى
// «احمد» تلكه «أحمد». بس البحث الي يصير بالسيرفر (لما القائمة تكبر وما
// نقدر نجيبها كلها للمتصفح) لازم يطبّع بنفس الطريقة بالضبط — وإلا
// يصير عدنا بحثين بسلوكين مختلفين، والموظف يلكه الزبون بشاشة وما
// يلكه بشاشة ثانية بنفس الكلمة.
//
// لهذا هاي الدالة نسخة طبق الأصل من الي بالواجهة. أي تعديل بوحدة
// منهن لازم ينعمل بالثانية.
func arabicSearchMigration() []Migration {
	return []Migration{
		{
			Version: "0224_arabic_normalize_function",
			SQL: `
				CREATE OR REPLACE FUNCTION ar_norm(t text) RETURNS text
				LANGUAGE sql IMMUTABLE PARALLEL SAFE AS $fn$
					SELECT btrim(regexp_replace(
						lower(translate(
							-- التشكيل والتطويل ينحذفون أول
							regexp_replace(coalesce(t, ''), '[ً-ْـ]', '', 'g'),
							-- همزات → ا · تاء مربوطة → ه · مقصورة → ي
							-- همزة على واو → و · على ياء → ي · أرقام عربية → إنجليزية
							'أإآٱةىؤئ٠١٢٣٤٥٦٧٨٩',
							'اااا' || 'هيوي' || '0123456789'
						)),
						'\s+', ' ', 'g'))
				$fn$;
			`,
		},
		{
			// pg_trgm يخلي البحث الجزئي ('%كلمة%') يستعمل فهرس بدل ما
			// يقرا كل الجدول. بدونه كل بحث بالزبائن = مسح كامل للجدول،
			// والوقت يكبر خطياً مع كل زبون جديد.
			Version: "0225_trigram_search_indexes",
			SQL: `
				CREATE EXTENSION IF NOT EXISTS pg_trgm;

				CREATE INDEX IF NOT EXISTS "Customer_name_trgm_idx"
					ON "Customer" USING gin (ar_norm(name) gin_trgm_ops);

				CREATE INDEX IF NOT EXISTS "Customer_phone_trgm_idx"
					ON "Customer" USING gin (phone gin_trgm_ops);

				CREATE INDEX IF NOT EXISTS "Quotation_customerName_trgm_idx"
					ON "Quotation" USING gin (ar_norm("customerName") gin_trgm_ops);
			`,
		},
		{
			// ⚠️ الإصدار الأول من ar_norm كان فيه حرف ألف زايد بخريطة
			// translate، فانزاحت كل المقابلات بحرف: «فاطمة» صارت «فاطما»
			// و«مصطفى» صارت «مصطفه». لكيناها بفحص التطبيع قبل النشر.
			//
			// نصلّحها هنا بدل ما نعدّل 0224 بس، لأن أي قاعدة شغّلت النسخة
			// الغلط لازم تتصحح — و REINDEX ضروري: فهارس trigram مبنية على
			// ناتج الدالة القديمة، وتبقى غلط لو ما أعدنا بناءها.
			Version: "0226_fix_arabic_normalize",
			SQL: `
				CREATE OR REPLACE FUNCTION ar_norm(t text) RETURNS text
				LANGUAGE sql IMMUTABLE PARALLEL SAFE AS $fn$
					SELECT btrim(regexp_replace(
						lower(translate(
							regexp_replace(coalesce(t, ''), '[ً-ْـ]', '', 'g'),
							'أإآٱةىؤئ٠١٢٣٤٥٦٧٨٩',
							'اااا' || 'هيوي' || '0123456789'
						)),
						'\s+', ' ', 'g'))
				$fn$;

				REINDEX INDEX "Customer_name_trgm_idx";
				REINDEX INDEX "Quotation_customerName_trgm_idx";
			`,
		},
	}
}
