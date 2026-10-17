package database

// ═══ عروض الأسعار: ترتيب البنود وأرشيف النسخ ═══
//
// (ع): «أريد أرتب أماكن الأجهزة، أريد أخلي الأول أخير والأخير أول» ·
// «أكو عروض أسعار نسويهن وبعد فترة نرجع نريد نعدّل عليهن — أريد
// ينحفظ بدل القديم بس القديم يضل مؤرشف بغير مكان».
//
// 🔴 **الترتيب ماكان موجوداً إطلاقاً**: بنود العرض تُقرا بلا أي
// ORDER BY، فترتيبهن هو الي ترجّعه القاعدة — ويتبدّل بين قراءة
// وقراءة وبعد أي تعديل. يعني العرض الي ينطبع للزبون ترتيب أجهزته
// **غير مضمون** أصلاً، قبل ما نتكلم عن سحب وإفلات.
//
// 🔴 **والتعديل يمحي النسخة القديمة**: التحديث يسوي
// DELETE على كل البنود ويعيد إدخالهن. فالعرض الي انرسل للزبون
// بتاريخ معيّن — بأسعاره وبنوده — **يضيع للأبد** أول ما أحد يعدّل.
// وهاي خسارة مستمرة: كل تعديل يمحي وثيقة انرسلت فعلاً.
//
// ⚠️ **لقطة JSONB مو جداول ظل**: النسخة القديمة وثيقة تاريخية
// تُقرا ما تُعدَّل. لو أرشفناها بجدولين (رأس وبنود) نصير ملزمين
// نحدّثهم كل ما يتبدّل شكل العرض، وأي عمود ننساه يخلي الأرشيف
// ناقصاً بهدوء. اللقطة تنحفظ كاملة مثل ما كانت بذاك اليوم.
//
// ⚠️ وrestrict على الموظف مقصودة بالعكس: ON DELETE SET NULL —
// الموظف ينمسح والنسخة تبقى، لأن الوثيقة أهم من نسبتها.

func quotationVersionMigrations() []Migration {
	return []Migration{
		{
			Version: "0281_quotation_item_sort_index",
			SQL: `
				ALTER TABLE "QuotationItem"
					ADD COLUMN IF NOT EXISTS "sortIndex" INTEGER NOT NULL DEFAULT 0;

				CREATE INDEX IF NOT EXISTS "QuotationItem_order_idx"
					ON "QuotationItem" ("quotationId", "sortIndex");
			`,
		},
		{
			Version: "0282_quotation_version_archive",
			SQL: `
				CREATE TABLE IF NOT EXISTS "QuotationVersion" (
					id             TEXT PRIMARY KEY,
					"quotationId"  TEXT NOT NULL REFERENCES "Quotation"(id) ON DELETE CASCADE,
					version        INTEGER NOT NULL,
					snapshot       JSONB NOT NULL,
					"archivedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					"archivedAt"   TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);

				CREATE UNIQUE INDEX IF NOT EXISTS "QuotationVersion_uniq"
					ON "QuotationVersion" ("quotationId", version);
			`,
		},
	}
}
