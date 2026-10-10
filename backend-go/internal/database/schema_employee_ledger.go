package database

// ══════════════════════════════════════════════════════════════════
// دفتر ذمة الموظف — المحفظة
// ══════════════════════════════════════════════════════════════════
//
// طلب صاحب النظام حرفياً:
//
//	«لو الموظف الليدر اجه وخله فلوس بالكاشير بدون ما يسوي أي فاتورة،
//	 هنا راح أضل أني أفرفر بذاناتي ما أدري شسوي ليش؟»
//
// وفحص النظام أثبت الحالة: الفلوس **الطالعة** متتبّعة (`RevolvingFund`
// — عهدة وتسوية بوصولات)، والفلوس **الداخلة** ماكو إلها أثر أبداً.
// يعني فلوس الزبائن تدخل الشركة والنظام أعمى عنها تماماً.
//
// ⚠️⚠️ **والي يمسكه ماتركس مو التحويل — هو غيابه.** «موظف عنده ٣
// حجوزات اليوم وولا فاتورة وولا تحصيل» هي الاكتشاف. وما تنمسك إلا
// بوجود **توقّع** — والتوقّع موجود أصلاً بـ`ServicePriceSample`
// (يتعلّم السعر من الفواتير الحقيقية) و`SystemPriceCatalog`.
//
// ⚠️ وليش دفتر مو عمود «رصيد» بجدول الموظف؟ لأن الرقم وحده ما يثبت
// شي. «عليه ٣٤٠ ألف» بلا تفاصيل كلام؛ «استلم ٥٠ من حجز س يوم كذا،
// وسلّم ٢٠٠ يوم كذا» دليل. والرصيد ينحسب من القيود، فما يكذب أبداً.
func employeeLedgerMigrations() []Migration {
	return []Migration{
		{
			Version: "0288_employee_ledger",
			SQL: `
				CREATE TABLE IF NOT EXISTS "EmployeeLedgerEntry" (
					id TEXT PRIMARY KEY,
					"employeeId" TEXT NOT NULL REFERENCES "Employee"(id),

					-- COLLECTED    استلم من زبون        → عليه
					-- HANDED_OVER  سلّم للشركة          → انبرأت ذمته
					-- ADJUSTMENT   تصحيح بقرار إنسان
					-- WRITE_OFF    إعفاء بقرار المالك
					kind TEXT NOT NULL CHECK (kind IN
						('COLLECTED','HANDED_OVER','ADJUSTMENT','WRITE_OFF')),

					-- ⚠️ موجب دائماً. الاتجاه يجي من kind مو من إشارة
					-- المبلغ: مبلغ سالب بقيد «استلم» يعني قيداً يقرا
					-- بالعكس حسب منو يجمعه، والدفتر يصير غامضاً.
					amount NUMERIC(14,2) NOT NULL CHECK (amount > 0),

					-- ⚠️⚠️ **الحجز مطلوب بالتحصيل** (الشرط تحت). هذا
					-- الي يحل مشكلة صاحب النظام فعلاً: المبلغ ما يگدر
					-- يوجد بلا ما يگول من أي حجز — فالفلوس اليتيمة
					-- تختفي **بالتصميم** مو بالرقابة.
					"bookingId" TEXT REFERENCES "Booking"(id),
					"invoiceId" TEXT REFERENCES "LeaderInvoice"(id) ON DELETE SET NULL,

					-- CASH | CARD | TRANSFER — البطاقة **طريقة دفع** مو
					-- نظام: التسليم واحد سواء چان كاشاً أو زين كاش أو
					-- حوالة، فالدفتر يشتغل اليوم بلا انتظار أي مزوّد.
					method TEXT CHECK (method IN ('CASH','CARD','TRANSFER')),
					reference TEXT,

					-- متى صارت الحركة فعلاً (مو متى انسجّلت): موظف
					-- يسجّل تحصيل أمس اليوم، والفرق بين التاريخين
					-- نفسه دليل.
					"occurredAt" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					"recordedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					note TEXT,

					-- ⚠️ التصحيح **قيد جديد** يشاور على القديم — مو
					-- تعديل. الاثنان يبقون ظاهرين، فالأثر ما ينمحي.
					"reversesEntryId" TEXT REFERENCES "EmployeeLedgerEntry"(id),

					"createdAt" TIMESTAMPTZ NOT NULL DEFAULT NOW(),

					-- تحصيل بلا حجز = فلوس يتيمة، وهاي بالضبط الي
					-- نبني الدفتر حتى نمنعها.
					CONSTRAINT employee_ledger_collected_needs_booking CHECK (
						kind <> 'COLLECTED' OR "bookingId" IS NOT NULL
					),
					-- القيد العكسي لازم يگول ليش — تصحيح بلا سبب
					-- يخلّي الدفتر يوازن وما يفسّر.
					CONSTRAINT employee_ledger_reversal_needs_note CHECK (
						"reversesEntryId" IS NULL
						OR (note IS NOT NULL AND btrim(note) <> '')
					)
				);

				CREATE INDEX IF NOT EXISTS "EmployeeLedgerEntry_employee_idx"
					ON "EmployeeLedgerEntry" ("employeeId", "occurredAt" DESC);
				CREATE INDEX IF NOT EXISTS "EmployeeLedgerEntry_booking_idx"
					ON "EmployeeLedgerEntry" ("bookingId");
				-- نفس القيد ما ينعكس مرتين.
				CREATE UNIQUE INDEX IF NOT EXISTS "EmployeeLedgerEntry_reverse_unique"
					ON "EmployeeLedgerEntry" ("reversesEntryId")
					WHERE "reversesEntryId" IS NOT NULL;
			`,
		},
		{
			// ═══ الحماية الحقيقية: الدفتر ما ينعدّل ولا ينمسح ═══
			//
			// ⚠️⚠️ **وليش بقاعدة البيانات مو بالكود؟** لأن الحارس بالكود
			// يحميه مسار واحد: أي خدمة جديدة، أو سكربت، أو psql بالإيد
			// يتجاوزه بلا ما ينتبه أحد. وهذا **دفتر فلوس** — أحكام
			// ماتركس والخصم من الراتب تنبني عليه. صف ينعدّل بهدوء يعني
			// الدليل ينقلب لتزوير، ومحد يعرف.
			//
			// الغلط ينصحّح **بقيد عكسي** (`reversesEntryId`) والاثنان
			// يبقون ظاهرين — نفس منطق دفاتر المحاسبة من قرون.
			Version: "0288_employee_ledger_immutable",
			SQL: `
				CREATE OR REPLACE FUNCTION employee_ledger_immutable()
				RETURNS TRIGGER AS $$
				BEGIN
					RAISE EXCEPTION
						'دفتر ذمة الموظف ما ينعدّل ولا ينمسح — التصحيح يصير بقيد عكسي جديد (reversesEntryId)';
				END;
				$$ LANGUAGE plpgsql;

				DROP TRIGGER IF EXISTS employee_ledger_no_update ON "EmployeeLedgerEntry";
				CREATE TRIGGER employee_ledger_no_update
					BEFORE UPDATE ON "EmployeeLedgerEntry"
					FOR EACH ROW EXECUTE FUNCTION employee_ledger_immutable();

				DROP TRIGGER IF EXISTS employee_ledger_no_delete ON "EmployeeLedgerEntry";
				CREATE TRIGGER employee_ledger_no_delete
					BEFORE DELETE ON "EmployeeLedgerEntry"
					FOR EACH ROW EXECUTE FUNCTION employee_ledger_immutable();

				-- 🔴 **وهاي الثغرة الي انكشفت بالتجربة**: المُشغّل الي
				-- يشتغل «لكل صف» **ما يشتغل على TRUNCATE** أبداً، لأنها
				-- عملية على الجدول مو على صفوفه. جرّبناها على قاعدة
				-- تجربة: TRUNCATE مسحت الدفتر كله (٥ قيود ← صفر) وهي
				-- تمر من فوق المُشغّلين الي فوق بلا ما تلمسهم.
				--
				-- ⚠️ فلازم مُشغّل **على مستوى الجملة** يمسكها — بلاه
				-- الحماية كلها تنهار بأمر واحد.
				DROP TRIGGER IF EXISTS employee_ledger_no_truncate ON "EmployeeLedgerEntry";
				CREATE TRIGGER employee_ledger_no_truncate
					BEFORE TRUNCATE ON "EmployeeLedgerEntry"
					FOR EACH STATEMENT EXECUTE FUNCTION employee_ledger_immutable();
			`,
		},
	}
}
