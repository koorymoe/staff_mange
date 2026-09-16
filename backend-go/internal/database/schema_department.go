package database

// ══════════════════════════════════════════════════════════════════
// سجل الأقسام ومسؤوليها — للحجز داخل الشركة
// ══════════════════════════════════════════════════════════════════
//
// «مباشرة تطلعله لستة بهاي الاقسام يختار قسم منهن… ونكدر نضيف
// الاقسام يدوياً ونضيف اسماء المسؤولين يدوياً مع قابلية اضافة
// اكثر من شخص يكدر يطلب حجز لنفس القسم، بالاضافة الى ارقام
// المسؤولين، وامكانية التعديل فقط لمالك ومدير النظام».
//
// ⚠️ **جدولان مو عمود نصّي**: القسم بالحجز چان نصاً حراً
// (`internalDepartment`)، يعني «قسم الاجهزه» و«قسم الأجهزة» صفّان
// مختلفان بالتقارير. السجل يخلي الاسم واحداً.
//
// ⚠️ **أكثر من مسؤول للقسم الواحد** — طلبه الصريح، فالمسؤولون
// جدول منفصل بمفتاح للقسم، مو عمود بالقسم.
//
// ⚠️ **`active` مو حذف فعلي**: قسم انلغى يبقى بالسجل حتى الحجوزات
// القديمة تبقى مقروءة — نفس مبدأ الأرشفة بكل النظام.
//
// ⚠️ **حسابات دخول للمسؤولين = مرحلة ثانية** (قراره: «الاثنان على
// مرحلتين»). هنا سجل أسماء وأرقام بس — `employeeId` موجود من هسه
// ويبقى فارغاً، حتى ما نحتاج ترحيلاً ثانياً لمن تجي المرحلة.
func departmentMigrations() []Migration {
	return []Migration{
		{
			Version: "0270_department",
			SQL: `CREATE TABLE IF NOT EXISTS "Department" (
			        id TEXT PRIMARY KEY,
			        name TEXT NOT NULL UNIQUE,
			        active BOOLEAN NOT NULL DEFAULT true,
			        "createdAt" TIMESTAMP NOT NULL DEFAULT now()
			      )`,
		},
		{
			Version: "0270_department_head",
			SQL: `CREATE TABLE IF NOT EXISTS "DepartmentHead" (
			        id TEXT PRIMARY KEY,
			        "departmentId" TEXT NOT NULL REFERENCES "Department"(id) ON DELETE CASCADE,
			        name TEXT NOT NULL,
			        phone TEXT,
			        "employeeId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
			        active BOOLEAN NOT NULL DEFAULT true,
			        "createdAt" TIMESTAMP NOT NULL DEFAULT now()
			      )`,
		},
		{
			Version: "0270_department_head_idx",
			SQL:     `CREATE INDEX IF NOT EXISTS "DepartmentHead_dept_idx" ON "DepartmentHead"("departmentId")`,
		},
		{
			// الأقسام الأربعة والعشرون الي دزّهم (ع) نصاً.
			// ⚠️ `ON CONFLICT DO NOTHING` حتى الترحيل ينعاد بلا ضرر،
			// وأي تعديل صار بالإيد بعدين ما ينمحي.
			Version: "0270_department_seed",
			SQL: `INSERT INTO "Department" (id, name) VALUES
			        (gen_random_uuid()::text, 'قسم المتابعة'),
			        (gen_random_uuid()::text, 'قسم الاجهزه'),
			        (gen_random_uuid()::text, 'قسم المخازن'),
			        (gen_random_uuid()::text, 'قسم السكيورتي'),
			        (gen_random_uuid()::text, 'قسم المشتريات'),
			        (gen_random_uuid()::text, 'قسم الشاشات'),
			        (gen_random_uuid()::text, 'قسم IT (البرمجة)'),
			        (gen_random_uuid()::text, 'قسم الالعاب'),
			        (gen_random_uuid()::text, 'قسم الحاسبات'),
			        (gen_random_uuid()::text, 'قسم الحسابات الآمنة'),
			        (gen_random_uuid()::text, 'قسم الخطوط'),
			        (gen_random_uuid()::text, 'قسم المحافظات'),
			        (gen_random_uuid()::text, 'قسم الموارد البشرية'),
			        (gen_random_uuid()::text, 'قسم التوصيل'),
			        (gen_random_uuid()::text, 'قسم الكول سنتر'),
			        (gen_random_uuid()::text, 'قسم الاكسسوارات'),
			        (gen_random_uuid()::text, 'قسم الأماني أرات'),
			        (gen_random_uuid()::text, 'قسم الديكور'),
			        (gen_random_uuid()::text, 'قسم الكهرباء'),
			        (gen_random_uuid()::text, 'قسم الصيانه'),
			        (gen_random_uuid()::text, 'قسم القانونية'),
			        (gen_random_uuid()::text, 'الشعبة الهندسية'),
			        (gen_random_uuid()::text, 'قسم الخدمات'),
			        (gen_random_uuid()::text, 'قسم الحسابات')
			      ON CONFLICT (name) DO NOTHING`,
		},
		{
			// ملاحظات إداري الكوادر على الحجز الداخلي — «اريد ملاحضات
			// يضيفها اداري الكوادر ع حجوزات داخل الشركة».
			// ⚠️ عمود منفصل عن `notes` العام: ملاحظة الإداري شي وملاحظة
			// الحجز شي، ودمجهن يضيّع منو كتب شنو.
			Version: "0270_booking_internal_hr_note",
			SQL: `ALTER TABLE "Booking"
			      ADD COLUMN IF NOT EXISTS "internalHrNote" TEXT,
			      ADD COLUMN IF NOT EXISTS "internalDepartmentId" TEXT REFERENCES "Department"(id) ON DELETE SET NULL,
			      ADD COLUMN IF NOT EXISTS "internalHeadId" TEXT REFERENCES "DepartmentHead"(id) ON DELETE SET NULL`,
		},
	}
}
