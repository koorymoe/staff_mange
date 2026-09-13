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
	}
}
