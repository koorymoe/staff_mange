package database

// ══════════════════════════════════════════════════════════════════
// الشغل المجاني — فاتورة بصفر، بس بسبب مكتوب
// ══════════════════════════════════════════════════════════════════
//
// أكو حجوزات تنشتغل مجاناً: ضمان، إعادة عمل بسبب خلل من عدنا،
// تعويض عن شكوى، أو مجاملة لزبون مهم. وقبل هذا التعديل، الليدر عنده
// خياران وكلاهما يخرّب الحسابات:
//
//	يسوي فاتورة بصفر بلا سبب → تطلع بالحسابات فاتورة بصفر ومحد
//	  يعرف ليش. وبعد شهر، لا المحاسب ولا المالك يقدر يميّز بين
//	  «شغل مجاني مقصود» و«الليدر نسى يحط الأسعار».
//
//	ما يسوي فاتورة أصلاً → الشغل ينضاع من السجل كله: ما ينحسب
//	  بإنتاجية الليدر، ولا يظهر إن الزبون أخذ خدمة، ولا نعرف شكد
//	  يكلفنا الضمان بالسنة.
//
// الحل: الفاتورة تنسوّى عادي بصفر، بس **بسبب من قائمة**. والقائمة
// بجدول مو ثابتة بالكود حتى صاحب العمل يعدّلها بنفسه لما تتغير
// سياسته، بدون ما يحتاج مبرمج.
//
// والسبب من قائمة مو نص حر: النص الحر ما ينجمّع ولا ينحسب. لما
// يكون من قائمة نقدر نجاوب «شكد كلّفنا الضمان هالسنة؟» بسؤال واحد.
func freeWorkMigration() []Migration {
	return []Migration{
		{
			Version: "0229_free_work_reason",
			SQL: `
				CREATE TABLE IF NOT EXISTS "FreeWorkReason" (
					id          TEXT PRIMARY KEY,
					label       TEXT NOT NULL UNIQUE,
					-- الترتيب بالقائمة — الأكثر استعمالاً فوق
					"sortOrder" INTEGER NOT NULL DEFAULT 0,
					-- بدل الحذف: السبب القديم لازم يبقى مفهوم بالفواتير
					-- القديمة الي تشاور عليه، بس ما ينعرض للجداد.
					active      BOOLEAN NOT NULL DEFAULT TRUE,
					-- يطلب توضيح مكتوب إضافي (مثل «أخرى»)
					"needsNote" BOOLEAN NOT NULL DEFAULT FALSE,
					"createdAt" TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);

				INSERT INTO "FreeWorkReason" (id, label, "sortOrder", "needsNote") VALUES
					('fwr_warranty',    'ضمان',                              1, false),
					('fwr_rework',      'إعادة عمل — خلل من عدنا',            2, true),
					('fwr_complaint',   'تعويض عن شكوى زبون',                 3, true),
					('fwr_goodwill',    'مجاملة / علاقات عامة',                4, true),
					('fwr_vip',         'شخصية مهمة',                          5, false),
					('fwr_promo',       'عرض ترويجي',                          6, false),
					('fwr_maintenance', 'صيانة دورية ضمن الاتفاق',             7, false),
					('fwr_other',       'سبب آخر',                             99, true)
				ON CONFLICT (label) DO NOTHING;
			`,
		},
		{
			Version: "0230_leader_invoice_free",
			SQL: `
				ALTER TABLE "LeaderInvoice"
					ADD COLUMN IF NOT EXISTS "isFree" BOOLEAN NOT NULL DEFAULT FALSE,
					ADD COLUMN IF NOT EXISTS "freeReasonId" TEXT REFERENCES "FreeWorkReason"(id) ON DELETE SET NULL,
					ADD COLUMN IF NOT EXISTS "freeReasonNote" TEXT;

				CREATE INDEX IF NOT EXISTS "LeaderInvoice_free_idx"
					ON "LeaderInvoice" ("freeReasonId") WHERE "isFree";
			`,
		},
	}
}
