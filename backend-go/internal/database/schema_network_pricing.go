package database

// ══════════════════════════════════════════════════════════════════
// تسعيرة الشبكات — فاتورة الشبكات
// ══════════════════════════════════════════════════════════════════
//
// الشبكات موجودة بالنظام كخدمة وكشعبة، بس ماكو إلها تسعيرة: الاكسل
// الي انبنت منه بقية الحاسبات (كاميرات، تكاليف المشروع) ما بيه ولا
// سعر شبكات. فالنتيجة الليدر يحسب فاتورة الشبكات براسه أو بالتلفون،
// وكل واحد يطلع برقم غير.
//
// ⚠️ ليش الأسعار بجدول مو بالكود؟
// أسعار الكاميرات منسوخة حرفياً من شيت ثابت وما تتغير إلا بقرار.
// الشبكات لسه تتبني — صاحب العمل راح يزوّد فقرات ويعدّل أرقام. لو
// حطيناها بالكود، كل سعر يتغير يريد تعديل ونشر. بالجدول: المالك
// ومدير النظام يعدلونها من الشاشة بنفسهم.
//
// ثلاث أنماط تسعير، لأن هذول الي موجودين فعلاً:
//
//	FLAT    = السعر × الكمية (تنصيب راوتر ١٧٬٠٠٠ للواحد)
//	TIERED  = مبلغ مقطوع لحد كمية معيّنة، وبعدها زيادة لكل وحدة
//	          تسليك كيبل: لحد ٢٠ متر ١٢٬٠٠٠، وكل متر زايد ١٬٤٠٠
//	          → ٣٠ متر = ١٢٬٠٠٠ + (١٠ × ١٬٤٠٠) = ٢٦٬٠٠٠
//	BRACKET = سعر الوحدة نفسه يتغيّر حسب حجم الكمية (تسعيرة جملة)
//	          تنظيم الراك: سويتج ٨ بورت → ٥٬٠٠٠ للبورت،
//	          ١٦ بورت → ٤٬٥٠٠، ٢٤ بورت → ٤٬٠٠٠. المجموع = العدد × سعر
//	          الشريحة. نفس منطق installMinimumPerDevice بحاسبة الكاميرات.
func networkPricingMigration() []Migration {
	return []Migration{
		{
			Version: "0233_network_price_item",
			SQL: `
				CREATE TABLE IF NOT EXISTS "NetworkPriceItem" (
					id            TEXT PRIMARY KEY,
					label         TEXT NOT NULL,
					-- وحدة القياس تنعرض بالاستمارة: متر / نقطة / جهاز...
					unit          TEXT NOT NULL DEFAULT 'قطعة',
					-- FLAT | TIERED | BRACKET
					"pricingMode" TEXT NOT NULL DEFAULT 'FLAT',
					-- FLAT: سعر الوحدة. TIERED: المبلغ المقطوع للكمية المشمولة.
					"basePrice"   NUMERIC(14,2) NOT NULL DEFAULT 0,
					-- TIERED بس: الكمية المشمولة بالمبلغ المقطوع
					"includedQty" NUMERIC(14,2) NOT NULL DEFAULT 0,
					-- TIERED بس: سعر كل وحدة بعد الكمية المشمولة
					"extraPerUnit" NUMERIC(14,2) NOT NULL DEFAULT 0,
					-- BRACKET بس: [{"upTo":8,"unitPrice":5000},...] وآخر شريحة
					-- upTo=0 يعني «وأكثر». مخزّنة JSON لأن عدد الشرائح يتغيّر
					-- من فقرة لفقرة، وجدول منفصل لثلاث صفوف مبالغة.
					brackets      JSONB NOT NULL DEFAULT '[]'::jsonb,
					note          TEXT,
					"sortOrder"   INTEGER NOT NULL DEFAULT 0,
					active        BOOLEAN NOT NULL DEFAULT true,
					"updatedAt"   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					"createdAt"   TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);

				CREATE INDEX IF NOT EXISTS "NetworkPriceItem_active_idx"
					ON "NetworkPriceItem" (active, "sortOrder");

				-- الفقرات الي عدنا أسعارها مؤكدة من صاحب العمل. الباقي
				-- ينضاف من الشاشة — ما نخمّن أسعار.
				INSERT INTO "NetworkPriceItem"
					(id, label, unit, "pricingMode", "basePrice", "includedQty", "extraPerUnit", brackets, note, "sortOrder")
				SELECT gen_random_uuid()::text, 'تسليك كيبل', 'متر', 'TIERED', 12000, 20, 1400, '[]'::jsonb,
					'لحد ٢٠ متر ١٢٬٠٠٠ مقطوعة، وكل متر زايد ١٬٤٠٠ (٣٠ متر = ٢٦٬٠٠٠)', 10
				WHERE NOT EXISTS (SELECT 1 FROM "NetworkPriceItem" WHERE label = 'تسليك كيبل');

				INSERT INTO "NetworkPriceItem"
					(id, label, unit, "pricingMode", "basePrice", "includedQty", "extraPerUnit", brackets, note, "sortOrder")
				SELECT gen_random_uuid()::text, 'تنصيب راوتر', 'جهاز', 'FLAT', 17000, 0, 0, '[]'::jsonb,
					'١٧٬٠٠٠ للراوتر الواحد', 20
				WHERE NOT EXISTS (SELECT 1 FROM "NetworkPriceItem" WHERE label = 'تنصيب راوتر');

				-- تنظيم الراك: سطر لكل سويتج، والكمية = عدد بورتاته.
				INSERT INTO "NetworkPriceItem"
					(id, label, unit, "pricingMode", "basePrice", "includedQty", "extraPerUnit", brackets, note, "sortOrder")
				SELECT gen_random_uuid()::text, 'تنظيم الراك', 'بورت', 'BRACKET', 0, 0, 0,
					'[{"upTo":8,"unitPrice":5000},{"upTo":16,"unitPrice":4500},{"upTo":0,"unitPrice":4000}]'::jsonb,
					'سعر البورت حسب حجم السويتج: ٨ بورت → ٥٬٠٠٠، ١٦ بورت → ٤٬٥٠٠، ٢٤ فما فوق → ٤٬٠٠٠. سطر لكل سويتج والكمية عدد بورتاته.', 30
				WHERE NOT EXISTS (SELECT 1 FROM "NetworkPriceItem" WHERE label = 'تنظيم الراك');
			`,
		},
	}
}
