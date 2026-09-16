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
