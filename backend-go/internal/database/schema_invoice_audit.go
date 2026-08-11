package database

// ══════════════════════════════════════════════════════════════════
// تدقيق فاتورة الليدر — الحكم، وقفل رقم الفاتورة، وسحب الاعتماد
// ══════════════════════════════════════════════════════════════════
//
// ثلاث فجوات، وحدة منهن كلّفت فلوس فعلاً:
//
// ١. **حكم التدقيق ما إله محل.** صاحب العمل طلب: «بالتدقيق مطابق /
//    غير مطابق / خطأ بالسعر». مطابق = سعر الفاتورة نفس المبلغ
//    الداخل. غير مطابق = يختلف. خطأ بالسعر = الموظف أخذ أعلى أو أوطى
//    من الفاتورة. اليوم ماكو ولا عمود يحفظ هذا الحكم — فالمحاسب
//    يعتمد أو ما يعتمد، والسبب يضيع.
//
// ٢. 🔴 **«سوّه اعتماد لفاتورة وبالغلط راحت لأن ما طلب منه رقم
//    الفاتورة».** الرقم إجباري بكود Go من فترة، **بس ماكو ولا قيد
//    بقاعدة البيانات**. يعني أي مسار (قديم، جديد، أو استعلام مباشر)
//    يكدر يحط status='APPROVED' برقم فاضي وما يوقفه شي. والدليل
//    موجود بالكود نفسه: SetExternalNumber مكتوب صراحةً «للفواتير الي
//    انعتمدت **قبل ما يصير الرقم إجبارياً**».
//
//    ⚠️ القيد يتزرع NOT VALID: يفرض الشرط على كل صف جديد أو منتحدّث،
//    بس ما يفشل الترحيل بسبب فواتير قديمة انعتمدت بلا رقم. لو زرعناه
//    VALID چان الترحيل طاح بالسيرفر وما اشتغل النظام أصلاً — والفواتير
//    القديمة تنعالج بسحب الاعتماد أو بـSetExternalNumber.
//
// ٣. **ماكو طريقة ترجّع فاتورة انعتمدت بالغلط.** الاعتماد كان طريق
//    باتجاه واحد. صاحب العمل: «لازم تخليلي خيار أكدر أرجعله الفواتير
//    الما معتمدة».
//    ⚠️ السحب ما يمحي: يرجّع الحالة SUBMITTED ويحفظ منو سحب وليش
//    ومتى، والرقم القديم يبقى بالسجل. محو الأثر بفاتورة انعتمدت
//    وانسحبت يخلي التدقيق مستحيل.
func invoiceAuditMigration() []Migration {
	return []Migration{
		{
			Version: "0246_leader_invoice_audit",
			SQL: `
				ALTER TABLE "LeaderInvoice"
					-- حكم التدقيق: MATCHED | MISMATCH | PRICE_ERROR
					ADD COLUMN IF NOT EXISTS "auditVerdict" TEXT,
					ADD COLUMN IF NOT EXISTS "auditNote" TEXT,
					ADD COLUMN IF NOT EXISTS "auditedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					ADD COLUMN IF NOT EXISTS "auditedAt" TIMESTAMPTZ,
					-- المبلغ الي دخل فعلاً — أساس المطابقة
					ADD COLUMN IF NOT EXISTS "auditedAmount" NUMERIC(14,2),

					-- سحب الاعتماد
					ADD COLUMN IF NOT EXISTS "revokedAt" TIMESTAMPTZ,
					ADD COLUMN IF NOT EXISTS "revokedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					ADD COLUMN IF NOT EXISTS "revokeReason" TEXT,
					ADD COLUMN IF NOT EXISTS "revokedCount" INT NOT NULL DEFAULT 0;

				-- 🔴 القيد الي كان ناقص: المعتمدة لازم إلها رقم فاتورة.
				-- بلا هذا، الحارس بكود Go وحده — وأي مسار ينساه يفتح
				-- نفس الثغرة الي كلّفتنا فاتورة.
				DO $$
				BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM pg_constraint
						WHERE conname = 'leader_invoice_approved_needs_number'
					) THEN
						ALTER TABLE "LeaderInvoice"
							ADD CONSTRAINT leader_invoice_approved_needs_number
							CHECK (
								status <> 'APPROVED'
								OR ("externalInvoiceNumber" IS NOT NULL
								    AND btrim("externalInvoiceNumber") <> '')
							) NOT VALID;
					END IF;
				END $$;

				CREATE INDEX IF NOT EXISTS "LeaderInvoice_audit_idx"
					ON "LeaderInvoice" ("auditVerdict", "auditedAt" DESC);
			`,
		},
	}
}
