package database

// ═══ النظام يتعلّم سعر الخدمات الي ما ذاكرينها ═══
//
// (ع): «هاي الفاتورة اليدوية من تضيفها بيها فدشي قوي كلش — أريد
// مقارنة، النظام يسوي مقارنة بين أسعار النظام وبين هاي الفاتورة،
// وبنفس الوقت يدرس ويحلل هاي الأرقام حتى يطلع أرقام ويسعّر الخدمات
// الي احنه ماذاكريهن… يضيف هاي الخدمة ويخلي إلها سعر أفيرج حسب معدل
// الأخذ من الزبائن السابقين».
// وقراره: **النظام يقترح، والمحاسب يراجع، وبعدين المالك يعتمد** ·
// وبعد **٥ عيّنات**.
//
// 🔴 **ليش هاي أعجل من «الذكاء» نفسه**: المعدّل يحتاج عيّنات، والعيّنة
// تنولد لحظة ما تنكتب الفاتورة — **ما تنستخرج بأثر رجعي**. والفاتورة
// اليدوية اليوم تسجّل وصفاً حراً وبس (`manualWork`)، ماكو بيها **أي**
// ربط بخدمة. يعني كل يوم يمر بلا هذا الربط = عيّنات تضيع للأبد.
// فالأساس ينبني أولاً حتى تبدي البيانات تتجمّع من الليلة.
//
// ⚠️ **العيّنة صف مستقل مو عمود على الفاتورة**: الفاتورة تنعدّل
// وتنسحب ويُعاد اعتمادها، والعيّنة لازم تبقى مثل ما انولدت. ومنفصلة
// يعني نقدر نستثني عيّنة شاذة بعدين بلا ما نلمس فاتورة.
//
// ⚠️ **وسعر الاقتراح ما ينخلط بالسعر الحقيقي أبداً**: عمود `source`
// يگول SEED (انزرع بالنظام) أو MANUAL_AVERAGE (اقتراح من العيّنات).
// بلا التفريق، بعد شهر محد يعرف أي سعر قرار إدارة وأي سعر تخمين
// آلة — فالكل يشك بالكل.
//
// ⚠️ **وماكو اعتماد تلقائي**: الاقتراح يبقى اقتراحاً بحالة PROPOSED
// لحد ما يراجعه المحاسب ويعتمده المالك. ولو انفتح الاعتماد التلقائي،
// فاتورتان غلط يخلقان سعراً رسمياً للشركة.

func servicePriceLearningMigrations() []Migration {
	return []Migration{
		{
			Version: "0283_manual_invoice_service_link",
			SQL: `
				ALTER TABLE "LeaderInvoice"
					ADD COLUMN IF NOT EXISTS "manualServiceId" TEXT REFERENCES "Service"(id) ON DELETE SET NULL;

				CREATE INDEX IF NOT EXISTS "LeaderInvoice_manualService_idx"
					ON "LeaderInvoice" ("manualServiceId")
					WHERE "manualServiceId" IS NOT NULL;
			`,
		},
		{
			Version: "0284_service_price_sample",
			SQL: `
				CREATE TABLE IF NOT EXISTS "ServicePriceSample" (
					id           TEXT PRIMARY KEY,
					"serviceId"  TEXT NOT NULL REFERENCES "Service"(id) ON DELETE CASCADE,
					"invoiceId"  TEXT REFERENCES "LeaderInvoice"(id) ON DELETE SET NULL,
					amount       DOUBLE PRECISION NOT NULL,
					"takenById"  TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					-- عيّنة مستثناة من المعدّل (شاذة) — بلا حذفها، حتى
					-- يبقى واضح إنها كانت موجودة ومنو استثناها وليش.
					excluded         BOOLEAN NOT NULL DEFAULT false,
					"excludeReason"  TEXT,
					"createdAt"  TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);

				CREATE INDEX IF NOT EXISTS "ServicePriceSample_service_idx"
					ON "ServicePriceSample" ("serviceId", excluded);
			`,
		},
		{
			Version: "0285_service_price_suggestion",
			SQL: `
				CREATE TABLE IF NOT EXISTS "ServicePriceSuggestion" (
					id            TEXT PRIMARY KEY,
					"serviceId"   TEXT NOT NULL REFERENCES "Service"(id) ON DELETE CASCADE,
					-- المعدّل المحسوب وقت الاقتراح، وعدد العيّنات الي
					-- انبنى عليها — الرقم بلا عدد عيّناته ما ينراجع.
					"avgAmount"   DOUBLE PRECISION NOT NULL,
					"sampleCount" INTEGER NOT NULL,
					"minAmount"   DOUBLE PRECISION NOT NULL,
					"maxAmount"   DOUBLE PRECISION NOT NULL,
					source        TEXT NOT NULL DEFAULT 'MANUAL_AVERAGE',
					-- PROPOSED → REVIEWED (المحاسب) → APPROVED (المالك) / REJECTED
					status        TEXT NOT NULL DEFAULT 'PROPOSED',
					"reviewedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					"reviewedAt"   TIMESTAMPTZ,
					"reviewNote"   TEXT,
					"decidedById"  TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					"decidedAt"    TIMESTAMPTZ,
					"createdAt"    TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);

				-- اقتراح مفتوح واحد لكل خدمة: بلا هذا يتراكم اقتراح كل
				-- فاتورة ويصير الطابور مية صف لنفس الخدمة.
				CREATE UNIQUE INDEX IF NOT EXISTS "ServicePriceSuggestion_open_uniq"
					ON "ServicePriceSuggestion" ("serviceId")
					WHERE status IN ('PROPOSED', 'REVIEWED');
			`,
		},
	}
}
