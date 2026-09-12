package database

// ══════════════════════════════════════════════════════════════════
// الكلفة اليدوية بفاتورة الليدر
// ══════════════════════════════════════════════════════════════════
//
// المشكلة الي انحلّت: جدول الكلفة (`SystemPriceCatalog`) يغطي المنظومات
// الثمانية ببنودها المعروفة. وأي شغل **برّا الجدول** — صيانة غير
// نمطية، تمديد ما إله بند، ترتيب موقع، شغل مركّب — ما إله سعر يطلع
// من الحساب، فالليدر إما يجبره على أقرب بند (فيطلع سعر غلط)، أو ما
// يسوي فاتورة أصلاً (فينغرم على ورق متأخر).
//
// ⚠️⚠️ **والسعر اليدوي يشيل الحارس الوحيد على التسعير**، فينبني
// بثلاثة قيود مكتوبة بالقاعدة نفسها لا بالكود بس:
//
//	① `pricingMode` عمود صريح — الفاتورة اليدوية **تتميّز للأبد**
//	   بكل شاشة وكل تقرير. بدونه تندسّ بين فواتير الجدول وما يعرف
//	   المحاسب ليش سعرها هيج.
//	② `manualWork` **إجباري** بالخدمة — سعر حر بلا وصف شنو انعمل
//	   هو رقم بلا مرجع، وبعد شهر محد يقدر يدقّقه.
//	③ `CHECK` على القيم المسموحة — عمود نصي حر يقبل أي شي يخلي
//	   التصفية تكذب بصمت.
//
// ⚠️ **والبذرة الرجعية مقصودة**: فواتير الجي بي اس والداش كام
// (`CreateServiceInvoice`) **يدوية فعلاً من يوم بنيناها** — سعرها
// بالإيد وبنودها فاضية. فتنوسم `MANUAL` بأثر رجعي حتى التصنيف
// يقول الحقيقة عن الماضي هم، مو عن الجديد بس.
func manualInvoiceMigrations() []Migration {
	return []Migration{
		{
			Version: "0274_leader_invoice_manual_pricing",
			SQL: `ALTER TABLE "LeaderInvoice"
			        ADD COLUMN IF NOT EXISTS "pricingMode"     TEXT NOT NULL DEFAULT 'CATALOG',
			        ADD COLUMN IF NOT EXISTS "manualWork"      TEXT,
			        ADD COLUMN IF NOT EXISTS "manualPriceNote" TEXT`,
		},
		{
			// الفواتير الي انبنت بلا بنود تنفيذ سعرها انكتب بالإيد
			// أصلاً — نوسمها بحقيقتها.
			Version: "0274_leader_invoice_manual_backfill",
			SQL: `UPDATE "LeaderInvoice"
			        SET "pricingMode" = 'MANUAL'
			      WHERE "pricingMode" = 'CATALOG'
			        AND COALESCE(items::text, '[]') IN ('[]', 'null')`,
		},
		{
			Version: "0274_leader_invoice_pricing_mode_check",
			SQL: `DO $$
			      BEGIN
			        IF NOT EXISTS (
			          SELECT 1 FROM pg_constraint WHERE conname = 'leader_invoice_pricing_mode_valid'
			        ) THEN
			          ALTER TABLE "LeaderInvoice"
			            ADD CONSTRAINT leader_invoice_pricing_mode_valid
			            CHECK ("pricingMode" IN ('CATALOG','MANUAL'));
			        END IF;
			      END $$`,
		},
		{
			// المحاسب والمراقب يحتاجون «وريني اليدوية بس» بسرعة —
			// فهرس جزئي لأن اليدوية أقلية.
			Version: "0274_leader_invoice_manual_idx",
			SQL: `CREATE INDEX IF NOT EXISTS leader_invoice_manual_idx
			        ON "LeaderInvoice" ("createdAt" DESC)
			        WHERE "pricingMode" = 'MANUAL'`,
		},
	}
}
