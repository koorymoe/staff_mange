package database

// ══════════════════════════════════════════════════════════════════
// تدقيق التكرار — حجوزات وزبائن مكررون بالغلط
// ══════════════════════════════════════════════════════════════════
//
// جزء من "ماتركس" — بس **خارج** خط أنابيب AiSignal/AiVerdict وصندوق
// المراقب عمداً: ذاك التصميم يفترض معرّف واحد يحل لصف واحد لهوية
// واحدة (`MonitorReview.hydrateIdentity`)، وزوج التكرار علاقة بين
// **صفّين** — حشره بنفس الآلية يحتاج تمديد حقيقي بالهوية والواجهة.
// فمساره مستقل وبسيط: يكتشف بالكود، ويعرض الطرفين جنب بعض، وصاحب
// العمل يقرر — بلا حذف أو دمج تلقائي أبداً.
//
// ⚠️ التطبيع بـ`phone_norm` **إجباري** قبل مقارنة أرقام الهاتف —
// نفس الدرس المتعلَّم من سكربتات تسوية دفتر تدقيق الحسابات هذا
// الجلسة: الهاتف المخزون قد يحمل أرقاماً عربية-هندية، رمز الدولة
// `+964`، أو صفراً أول ناقصاً — مقارنة الأعمدة الخام مباشرة تفوّت
// أغلب الحالات الحقيقية.
func duplicateCandidateMigration() []Migration {
	return []Migration{
		{
			Version: "0292_phone_norm_function",
			SQL: `
				CREATE OR REPLACE FUNCTION phone_norm(p text) RETURNS text
				LANGUAGE sql IMMUTABLE PARALLEL SAFE AS $fn$
					SELECT CASE
						WHEN raw = '' THEN ''
						WHEN raw LIKE '964%' AND length(raw) > 10 THEN '0' || substring(raw from 4)
						WHEN length(raw) = 10 AND left(raw, 1) = '7' THEN '0' || raw
						ELSE raw
					END
					FROM (
						SELECT regexp_replace(
							translate(coalesce(p, ''), '٠١٢٣٤٥٦٧٨٩', '0123456789'),
							'[^0-9]', '', 'g'
						) AS raw
					) s
				$fn$;
			`,
		},
		{
			Version: "0293_duplicate_candidate",
			SQL: `
				CREATE TABLE IF NOT EXISTS "DuplicateCandidate" (
					id             TEXT PRIMARY KEY,
					kind           TEXT NOT NULL CHECK (kind IN ('BOOKING', 'CUSTOMER')),
					"entityAId"    TEXT NOT NULL,
					"entityBId"    TEXT NOT NULL,
					"matchReason"  TEXT NOT NULL,
					status         TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'DISMISSED')),
					"reviewedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					"reviewedAt"   TIMESTAMPTZ,
					"detectedAt"   TIMESTAMPTZ NOT NULL DEFAULT now(),
					UNIQUE (kind, "entityAId", "entityBId")
				);
				CREATE INDEX IF NOT EXISTS "DuplicateCandidate_status_idx"
					ON "DuplicateCandidate" (kind, status);

				CREATE INDEX IF NOT EXISTS "Customer_phone_norm_idx"
					ON "Customer" (phone_norm(phone));
			`,
		},
	}
}
