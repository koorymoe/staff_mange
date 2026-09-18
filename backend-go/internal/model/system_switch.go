package model

import "time"

// SystemSwitch مفتاح ميزة على مستوى النظام كله.
type SystemSwitch struct {
	Key       string    `db:"key" json:"key"`
	Enabled   bool      `db:"enabled" json:"enabled"`
	UpdatedAt time.Time `db:"updatedAt" json:"updatedAt"`
	UpdatedBy *string   `db:"updatedBy" json:"updatedBy"`
}

// المفاتيح المعروفة — **قائمة بيضاء مقصودة**.
//
// ⚠️ بلاها أي واحد عنده حساب مالك يكتب مفتاحاً بأي اسم، فيتجمّع
// بالجدول صفوف ما تقرأها ولا شاشة — و(ع) يحسب إنه أطفى شي وهو ما
// انطفى، لأن الاسم انكتب غلط بحرف.
const (
	// SwitchEntity شخصية الكائن — الودجة العائمة وورقة القصة.
	SwitchEntity = "entity_enabled"
	// SwitchAnnouncements شريط الإعلانات فوق كل الشاشات.
	SwitchAnnouncements = "announcements_enabled"
)

// SystemSwitchLabels الاسم العربي لكل مفتاح — للعرض وللسجل.
var SystemSwitchLabels = map[string]string{
	SwitchEntity:        "شخصية الكائن",
	SwitchAnnouncements: "شريط الإعلانات",
}

// KnownSystemSwitch هل هذا مفتاح نعرفه؟
func KnownSystemSwitch(key string) bool {
	_, ok := SystemSwitchLabels[key]
	return ok
}

// systemSwitchAdminAllowed المفاتيح الي مدير النظام (ADMIN) يقدر
// يبدّلها هو الثاني — مو المالك حصراً.
//
// ⚠️ **قائمة بيضاء بالاسم مو قاعدة عامة**: شريط الإعلانات قرار
// تشغيلي يومي، بينما شخصية الكائن مرتبطة بقرارات (ع) الشخصية
// (الصور، الترخيص) وتبقى له حصراً حتى لو مدير النظام موجود.
var systemSwitchAdminAllowed = map[string]bool{
	SwitchAnnouncements: true,
}

// SystemSwitchAdminAllowed هل مدير النظام (ADMIN) مسموح له يبدّل
// هذا المفتاح، إضافة على المالك الي مسموح له كل المفاتيح دايماً؟
func SystemSwitchAdminAllowed(key string) bool {
	return systemSwitchAdminAllowed[key]
}
