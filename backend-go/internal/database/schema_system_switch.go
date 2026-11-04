package database

// ═══ مفاتيح النظام — إطفاء ميزة لكل الموظفين بضغطة ═══
//
// 🔴 **ليش بالخادم مو بالمتصفح**: (ع) كال «ماريد أي شخصية تظهر» —
// يعني لكل الموظفين، مو على شاشته هو. ولو خزّنا المفتاح
// بـ`localStorage`، ينطفي عند (ع) وحده ويبقى شغّالاً عند ٧ موظفين
// وهو يحسب إنه أطفاه.
//
// ⚠️ **وجدول عام مو عمودين**: ثاني مرة يطلب إطفاء شي (والطلبات
// جايّة) نضيف **صفاً** مو ترحيلاً وعموداً وحقلاً بالموديل — وكل
// عمود جديد على جدول يُجلب بـ`SELECT *` مخاطرة بحد ذاته.
//
// ⚠️ **والافتراضي «شغّال» لمّا ما يكون اكو صف**: الميزة الي انبنت
// تبقى شغّالة، والإطفاء قرار صريح مكتوب بالجدول. والعكس (الافتراضي
// مطفي) يعني أي خلل بالجلب يطفّي النظام كله بهدوء.

func systemSwitchMigrations() []Migration {
	return []Migration{
		{
			Version: "0279_system_switch",
			SQL: `
				CREATE TABLE IF NOT EXISTS "SystemSwitch" (
					key         TEXT PRIMARY KEY,
					enabled     BOOLEAN NOT NULL DEFAULT true,
					"updatedAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					"updatedBy" TEXT
				);
			`,
		},
	}
}
