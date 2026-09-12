package database

// ═══ تقييم الزبون وتدقيق مهندسي الجودة ═══
//
// صاحب النظام دزّ تصميم شاشة الشكاوى وبيه عمود «تقييم الزبون»
// (٤.٣ · ٢.٥ · ٥) وبطاقة «متوسط التقييم ٤.٦/٥» وعمود «متابعة
// المدقق». ولا وحدة منهن موجودة: **ماكو حقل تقييم بالنظام بالمرّة**،
// وماكو أي أثر لحكم المدقق على الشكوى.
//
// يعني البطاقة لو انرسمت بلا هذا الترحيل تكون **رقماً مخترعاً**.
//
// **منو يسجّل التقييم**: مهندس الجودة وقت التواصل — قرار صاحب
// النظام. يعني ينضاف لمسار التواصل الموجود بدل مسار جديد، ويشتغل
// من أول يوم.
//
// ⚠️⚠️ **`CHECK` إجباري مو زينة**: التقييم يدخل بمتوسط يشوفه المدير
// ويحاسب عليه. رقم برّا ١..٥ — من نداء مباشر أو غلط بالواجهة —
// **يسمّم المتوسط**، والمتوسط المسموم أسوأ من ماكو متوسط، لأن
// الأول يُتّخذ عليه قرار والثاني ما يُتّخذ.
//
// ⚠️ **والحقل nullable بقصد**: «ما انسأل» **مو** «تقييمه صفر».
// الصفر يهبّط المتوسط ويظلم مهندس الجودة على شكوى ماكو بيها تقييم
// أصلاً. فالفاضي يبقى فاضي، والمتوسط يتجاهله (`FILTER`).
//
// ⚠️ **`ON DELETE SET NULL` مو `CASCADE`** على `auditedById`: حذف
// موظف ما يصير يمحي **إن التدقيق صار**. الحكم واقعة حصلت، وهوية
// مَن حكم تنفقد — أما الواقعة نفسها فتبقى.
//
// ⚠️⚠️ **وترحيل مستقل مو تعديل `CREATE TABLE`**: جدول `Complaint`
// موجود بقاعدة الإنتاج، و`CREATE TABLE IF NOT EXISTS` **ما تسوي شي**
// عليها.
//
// ⚠️⚠️ **و`SELECT *` بالمستودع**: هاي الخمس حقول **لازم** تنضاف
// لـstruct الشكوى بنفس الدفعة، وإلا كل استعلام على الشكاوى يفشل
// بـ«missing destination name» — يعني الشاشة تنهار كلها.
func complaintQualityMigrations() []Migration {
	return []Migration{
		{
			Version: "0260_complaint_quality",
			SQL: `
				ALTER TABLE "Complaint"
					ADD COLUMN IF NOT EXISTS "customerRating" INT,
					ADD COLUMN IF NOT EXISTS "auditVerdict"   TEXT,
					ADD COLUMN IF NOT EXISTS "auditNote"      TEXT,
					ADD COLUMN IF NOT EXISTS "auditedAt"      TIMESTAMP,
					ADD COLUMN IF NOT EXISTS "auditedById"    TEXT
						REFERENCES "Employee"(id) ON DELETE SET NULL;

				-- القيد ينضاف لحاله حتى يكون إعادة التشغيل آمنة: لو
				-- انضاف قبل، ما نحاول ثانية ونفشل الترحيل كله.
				DO $$
				BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM pg_constraint
						WHERE conname = 'Complaint_customerRating_range'
					) THEN
						ALTER TABLE "Complaint"
							ADD CONSTRAINT "Complaint_customerRating_range"
							CHECK ("customerRating" IS NULL
							       OR "customerRating" BETWEEN 1 AND 5);
					END IF;
				END $$;

				-- المدقق يدوّر على الي ما انعتمد — وهذا العمود الي
				-- يترشّح عليه بكل فتحة للشاشة.
				CREATE INDEX IF NOT EXISTS "Complaint_auditVerdict_idx"
					ON "Complaint" ("auditVerdict");
			`,
		},
	}
}
