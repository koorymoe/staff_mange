package database

// ══════════════════════════════════════════════════════════════════
// سجل تعديلات فاتورة الليدر — قبل وبعد
// ══════════════════════════════════════════════════════════════════
//
// تعديل المحاسب على المبالغ كان **يمحي الأصل**: التحديث يدوس على
// executionCost وmaterialsTotal وdiscountValue وnetTotal بمكانهن،
// وماكو ولا عمود يحفظ الي كان قبل.
//
// ثلاث فجوات انفتحن من هذا:
//
//	١. الأرقام الأصلية تروح للأبد
//	٢. ماكو adjustedById — النظام ما يعرف **منو** المحاسب الي عدّل
//	٣. التعديل الثاني يدوس على سبب ووقت الأول
//
// يعني المراقب يشوف الرقم الجديد ويظن إنه الأصلي، وماكو أثر يقول
// شنو كان ولا منو غيّره. وصاحب العمل طلب صراحةً إن المراقب يشوف
// «الفاتورة قبل التدقيق وبعده والملاحظات الي كتبها المحاسب».
//
// ⚠️ ليش جدول مو أعمدة على الفاتورة؟
// لأن التعديل يتكرر. الأعمدة تحفظ آخر تعديل بس وتدوس على الي قبله —
// يعني نفس المشكلة بشكل ثاني. الجدول يحفظ كل تعديل بسطر مستقل.
//
// ⚠️ التعديلات الي صارت **قبل** هذا الجدول ضاعت أرقامها ولا تنرجع.
// السجل يبدي من لحظة تشغيله وطالع.
func invoiceAdjustmentMigration() []Migration {
	return []Migration{
		{
			Version: "0241_leader_invoice_adjustment",
			SQL: `
				CREATE TABLE IF NOT EXISTS "LeaderInvoiceAdjustment" (
					id TEXT PRIMARY KEY,
					"invoiceId" TEXT NOT NULL REFERENCES "LeaderInvoice"(id) ON DELETE CASCADE,

					-- قبل وبعد للمبالغ الأربعة. نحفظ الأربعة حتى لو ما
					-- انتغيّر إلا واحد — المقارنة لازم تكون كاملة.
					"oldExecutionCost"  NUMERIC(14,2) NOT NULL DEFAULT 0,
					"newExecutionCost"  NUMERIC(14,2) NOT NULL DEFAULT 0,
					"oldMaterialsTotal" NUMERIC(14,2) NOT NULL DEFAULT 0,
					"newMaterialsTotal" NUMERIC(14,2) NOT NULL DEFAULT 0,
					"oldDiscountValue"  NUMERIC(14,2) NOT NULL DEFAULT 0,
					"newDiscountValue"  NUMERIC(14,2) NOT NULL DEFAULT 0,
					"oldNetTotal"       NUMERIC(14,2) NOT NULL DEFAULT 0,
					"newNetTotal"       NUMERIC(14,2) NOT NULL DEFAULT 0,

					reason TEXT NOT NULL,
					-- منو عدّل. ON DELETE SET NULL: الموظف ممكن ينحذف،
					-- بس التعديل لازم يبقى بالسجل.
					"adjustedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					"createdAt" TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);

				CREATE INDEX IF NOT EXISTS "LeaderInvoiceAdjustment_invoice_idx"
					ON "LeaderInvoiceAdjustment" ("invoiceId", "createdAt" DESC);
			`,
		},
	}
}
