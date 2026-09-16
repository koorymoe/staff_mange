package database

// ═══ الفاتورة تروح للمراقب — وترجع للمحاسب ═══
//
// طلبان من صاحب النظام:
//
// ١) المحاسب بعد ما يأشّر حكمه، يطلعله **مع الاعتماد** خيار ثاني:
//    **يرسلها للمراقب** حتى يراجعها ويدققها. يعني الفاتورة ما تنحصر
//    بين «اعتمدها» و«اتركها» — اكو طريق ثالث للشك.
//
// ٢) والفواتير الي مرّت **قبل** فصل الطوابير: المالك وحده يگدر
//    يرجّعهن للمحاسب حتى يرتّبهن من جديد.
//
// ⚠️⚠️ **الحالة تنشتق من طوابع زمنية — مو من عمود «مرحلة» جديد.**
// عمود مرحلة منفصل يعني حقيقتين لنفس الشي: فاتورة رجعت من المراقب
// ومرحلتها لسه «بانتظار المراقب» لأن أحداً نسى يحدّثه. الطابع الزمني
// **يجاوب سؤالين بجواب واحد**: وين هي الآن، ومتى انتقلت.
//
// ⚠️ ورقم الفاتورة المحاسبية **ما ينمحي** بالإرجاع — نفس قاعدة سحب
// الاعتماد الموجودة. الرقم صادر من نظام المحاسب فعلاً، ومحوه عدنا
// ما يمحيه عنده: يقطع الخيط بين النظامين ويبقى رقم يتيم ما يدل على
// شي.

func invoiceMonitorMigrations() []Migration {
	return []Migration{
		{
			Version: "0258_invoice_monitor_review",
			SQL: `
				ALTER TABLE "LeaderInvoice"
					-- المحاسب أرسلها للمراقب: الطابع موجود = هي عنده الآن
					ADD COLUMN IF NOT EXISTS "monitorRequestedAt"  TIMESTAMPTZ,
					ADD COLUMN IF NOT EXISTS "monitorRequestedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					ADD COLUMN IF NOT EXISTS "monitorRequestNote"   TEXT,

					-- المراقب بتّ بيها: الطابع موجود = رجعت للمحاسب بحكمها
					ADD COLUMN IF NOT EXISTS "monitorDecidedAt"    TIMESTAMPTZ,
					ADD COLUMN IF NOT EXISTS "monitorDecidedById"  TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					-- OK | FLAGGED — نفس مفردات صندوق المراقب حتى ما
					-- تختلف التسمية بين مكانين
					ADD COLUMN IF NOT EXISTS "monitorVerdict"      TEXT,
					ADD COLUMN IF NOT EXISTS "monitorNote"         TEXT,

					-- المالك رجّعها للمحاسب حتى يرتّبها من جديد
					ADD COLUMN IF NOT EXISTS "returnedAt"          TIMESTAMPTZ,
					ADD COLUMN IF NOT EXISTS "returnedById"        TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					ADD COLUMN IF NOT EXISTS "returnReason"        TEXT,
					-- ⚠️ عدّاد مو علم: فاتورة رجعت ثلاث مرات مشكلة
					-- تختلف عن فاتورة رجعت مرة، والعلم يخفي الفرق.
					ADD COLUMN IF NOT EXISTS "returnedCount"       INTEGER NOT NULL DEFAULT 0;

				-- طابور المراقب: الفواتير الي عنده الآن
				CREATE INDEX IF NOT EXISTS "LeaderInvoice_monitor_pending_idx"
					ON "LeaderInvoice" ("monitorRequestedAt")
					WHERE "monitorRequestedAt" IS NOT NULL AND "monitorDecidedAt" IS NULL;
			`,
		},
	}
}
