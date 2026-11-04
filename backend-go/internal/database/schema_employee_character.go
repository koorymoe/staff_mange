package database

// ══════════════════════════════════════════════════════════════════
// الكيان — شخصية لكل موظف تراقبه وتساعده
// ══════════════════════════════════════════════════════════════════
//
// طلب صاحب النظام: «كل شخص راح يكون عنده إيموجي حسب شخصيته… يكون
// عبارة عن كيان يهابه ويخافه الموظف، أول ما يفتح النظام يرحّب بيه:
// آني المراقب عليك والمساعد بنفس الوقت… كأنما بشر يراقب بشر».
//
// ⚠️ ليش جدول مستقل مو أعمدة على "Employee"؟
// الشخصية تتولّد وتتبدّل وتفشل وتُعاد — حالة وصور ونص وصف ووقت
// توليد ومنو ولّدها. حشرها بصف الموظف يعني ستة أعمدة تتغيّر بمعزل
// تام عن بيانات الموظف، وكل استعلام يجيب الموظف يسحبها وياه بلا
// حاجة (نفس علّة الصور base64 الي انشالت من القاعدة سابقاً).
//
// ⚠️ والصور ما تنخزن هنا — بس **مفاتيح** التخزين (storage.Store)،
// نفس ما تنخزن صورة الموظف بالضبط. الصورة الوحدة بميغا، وثلاث صور
// لكل موظف داخل القاعدة تنفخها بلا فايدة.
//
// ⚠️ ثلاث ملامح مو وحدة: هادئ (المراقبة) · مبتسم (نظيف ومنجز) ·
// غاضب (تأخر وانخصم منه). التعبير هو الي يخلي الكيان يحس حي —
// وصورة وحدة ثابتة تخليه أيقونة بلا روح.
func employeeCharacterMigrations() []Migration {
	return []Migration{
		{
			Version: "0267_employee_character",
			SQL: `
				CREATE TABLE IF NOT EXISTS "EmployeeCharacter" (
					id TEXT PRIMARY KEY,

					-- ⚠️ فريد: كيان واحد لكل موظف. وCASCADE لأن شخصية
					-- بلا صاحبها ما إلها أي معنى.
					"employeeId" TEXT NOT NULL UNIQUE
						REFERENCES "Employee"(id) ON DELETE CASCADE,

					-- وصف الشخصية بالعربي، مشتق من بيانات الموظف
					-- الحقيقية (دوره، انضباطه، إنجازه) — هو الي يغذّي
					-- طلب توليد الصورة وطريقة كلام الكيان.
					persona TEXT,

					-- مفاتيح التخزين للملامح الثلاث (storage.Store)
					"calmKey"  TEXT,
					"happyKey" TEXT,
					"angryKey" TEXT,

					-- الطلب الي انولدت بيه — حتى نعرف ليش طلعت هيچ
					-- ونقدر نعيد نفس التوليد بالضبط
					prompt TEXT,

					-- PENDING (قيد التوليد) · READY · FAILED
					status TEXT NOT NULL DEFAULT 'PENDING',
					error TEXT,

					"generatedAt" TIMESTAMP,
					-- ⚠️ SET NULL: حذف المدير ما يمحي إن الشخصية انولدت
					"generatedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,

					"createdAt" TIMESTAMP NOT NULL DEFAULT now(),
					"updatedAt" TIMESTAMP NOT NULL DEFAULT now()
				);
				CREATE INDEX IF NOT EXISTS "EmployeeCharacter_status_idx"
					ON "EmployeeCharacter"(status);
			`,
		},
	}
}
