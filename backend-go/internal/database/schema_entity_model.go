package database

// مجسّمات الكيان ثلاثية الأبعاد — تُرفَع من داخل النظام
//
// ⚠️ **ليش جدول ومو خانة إعداد واحدة؟**
// قبلها كان المجسّم **ملفاً ثابتاً** بـ`frontend/public`، ومعناها إن
// كل تغيير يحتاج: نسخ الملف للسيرفر بالإيد → بناء حاوية → نشر. وهاي
// طلبناها من مالك النظام فعلاً وما انفهمت — وهي أصلاً شغلة مبرمج.
//
// وبجدول نگدر **نرفع أكثر من مجسّم ونبدّل بينهم ونقارنهم جنب بعض**
// بلا نشر ولا SSH. والمقارنة هي بالضبط الي يحتاجها القرار: منو أحسن.
//
// 🔒 **ونخزّن المفتاح مو الملف**: الملف يروح لطبقة التخزين
// (`internal/storage`)، والصف يحمل مفتاحه بس. والتخزين يُخدَم **بعد
// تسجيل الدخول بوسم موقَّع** — وهذا ألزم لرخصة المجسّمات المشتراة
// من أسواق ثلاثية الأبعاد، لأنها تمنع **إتاحة المجسّم كملف** للعموم،
// والملف الثابت بـ`public/` كان مكشوفاً لأي أحد.
// 🔴 و**خانة «نشط»** (ترحيل 0277) هي الي تخلي التبديل فعلياً: قبلها
// كان اسم الملف مكتوباً **ثابتاً** بـ`EntityAvatar.tsx`، فالمجسّم
// المرفوع ما يوصل لأي موظف — يبين بشاشة المختبر وحدها. ومالك النظام
// شكى إنه **ما شاف الشخصية الجديدة أبداً بالنظام**، وكان محقاً.
func entityModelMigrations() []Migration {
	return []Migration{
		{
			Version: "0276_entity_avatar_model",
			SQL: `
CREATE TABLE IF NOT EXISTS "EntityAvatarModel" (
  id TEXT PRIMARY KEY,
  label TEXT NOT NULL,
  "fileKey" TEXT NOT NULL,
  "fileType" TEXT NOT NULL,
  "sizeBytes" BIGINT NOT NULL DEFAULT 0,
  -- ⚠️ SET NULL مو CASCADE: لو انحذف الموظف الي رفعه، المجسّم يبقى
  -- (النظام يعتمد عليه بالعرض) ويضيع اسم الرافع بس.
  "uploadedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
  -- الأرشفة ناعمة: ما نحذف ملفاً ممكن يكون معروضاً لموظفين
  "archivedAt" TIMESTAMP,
  "createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS "EntityAvatarModel_active_idx"
  ON "EntityAvatarModel" ("archivedAt", "createdAt" DESC);
`,
		},
		{
			Version: "0277_entity_avatar_model_active",
			SQL: `
ALTER TABLE "EntityAvatarModel"
  ADD COLUMN IF NOT EXISTS "isActive" BOOLEAN NOT NULL DEFAULT false;
-- 🔴 نشط **واحد حصراً** — والقيد بقاعدة البيانات مو بالكود: لو صارت
-- نشطين، الودجة تعرض مجسّماً يتغيّر حسب ترتيب الاستعلام، وهذا عيب
-- ما يبيّن إلا بالإنتاج.
CREATE UNIQUE INDEX IF NOT EXISTS "EntityAvatarModel_one_active_idx"
  ON "EntityAvatarModel" ("isActive") WHERE "isActive";
`,
		},
	}
}
