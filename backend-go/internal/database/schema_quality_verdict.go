package database

// ══════════════════════════════════════════════════════════════════
// حكم الجودة: تقرير إيجابي/سلبي، والكشف قبل الغرامة
// ══════════════════════════════════════════════════════════════════
//
// قبل هذا، مهندس الجودة يتصل بالزبون ويضغط «تواصلت — اكو مشكلة»
// ويكتب ملاحظة. وبس. المشكلة تنسجّل بنص حر ومحد يتحاسب عليها، فالشكوى
// تروح بالهوا والليدر ما يعرف إن زبونه اشتكى عليه أصلاً.
//
// صار الحكم إله أثر حقيقي:
//
//   - تقرير إيجابي → خير على خير، ينسجّل وبس.
//   - تقرير سلبي   → تنخصم نقطة من معيار «شكوى الزبائن» (وحدة من
//     معايير الكي بي اي الثمانية) من **الليدر** — لأنه المسؤول عن
//     شغل كادره قدام الزبون.
//
// ── وليش «يحتاج كشف»؟ ──
//
// لأن الزبون أحياناً يجذب. يشتكي حتى يوصّل تخفيض، أو ينقنق على شغلة
// مو ذنب الكادر. ولو خصمنا بمجرد الشكوى، نظلم الليدر وننطي الزبون
// سلاح يستعمله كل مرة.
//
// فمهندس الجودة عنده خيار ثالث: «يحتاج كشف» — يوقف الغرامة لحد ما
// يطلع أحد يشوف بعينه. وبعد الكشف:
//
//	كلام الزبون صح  → تنزل الغرامة (متأخرة، بس عادلة)
//	كلام الزبون كذب → ما ينغرم أحد، وتنسجّل علامة **على الزبون**
//
// وعداد الشكاوى الكاذبة على الزبون مو للانتقام — هو حتى المبيعات
// والإداري يعرفون شنو يتعاملون وياه قبل ما يوعدوه بشي.
func qualityVerdictMigration() []Migration {
	return []Migration{
		{
			Version: "0227_quality_followup_verdict",
			SQL: `
				ALTER TABLE "QualityFollowUp"
					-- POSITIVE | NEGATIVE — فاضي يعني لسه ما انبتّ بيه
					ADD COLUMN IF NOT EXISTS "reportType" TEXT,
					-- مسار الكشف: NONE | PENDING | DONE
					ADD COLUMN IF NOT EXISTS "inspectionStatus" TEXT NOT NULL DEFAULT 'NONE',
					-- نتيجة الكشف: CUSTOMER_RIGHT | CUSTOMER_WRONG
					ADD COLUMN IF NOT EXISTS "inspectionResult" TEXT,
					ADD COLUMN IF NOT EXISTS "inspectionNotes" TEXT,
					ADD COLUMN IF NOT EXISTS "inspectedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					ADD COLUMN IF NOT EXISTS "inspectedAt" TIMESTAMPTZ,
					-- منو انغرم فعلاً وبأي تقييم — حتى الغرامة تنتتبّع
					-- وتنلغى لو انكشف إنها غلط، بدل ما تضيع بالسجل.
					ADD COLUMN IF NOT EXISTS "penalizedEmployeeId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					ADD COLUMN IF NOT EXISTS "kpiEvaluationId" TEXT;

				CREATE INDEX IF NOT EXISTS "QualityFollowUp_inspection_idx"
					ON "QualityFollowUp" ("inspectionStatus")
					WHERE "inspectionStatus" = 'PENDING';
			`,
		},
		{
			Version: "0228_customer_false_claims",
			SQL: `
				ALTER TABLE "Customer"
					-- كم مرة انكشف إن شكواه ما كانت صحيحة
					ADD COLUMN IF NOT EXISTS "falseClaimCount" INTEGER NOT NULL DEFAULT 0,
					ADD COLUMN IF NOT EXISTS "lastFalseClaimAt" TIMESTAMPTZ,
					ADD COLUMN IF NOT EXISTS "falseClaimNote" TEXT;
			`,
		},
	}
}
