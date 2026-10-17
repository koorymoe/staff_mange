package model

import "time"

// PrivacyPolicyPoint نقطة وحدة من سياسة الخصوصية. السياسة كلها عبارة عن نقاط
// مرقّمة يضيفها صاحب صلاحية privacy_policy_manage، ونحتفظ بمنو أضاف كل نقطة
// حتى يشوفها المالك ومدير النظام.
type PrivacyPolicyPoint struct {
	ID                  string    `db:"id" json:"id"`
	Content             string    `db:"content" json:"content"`
	Order               int       `db:"order" json:"order"`
	IsActive            bool      `db:"isActive" json:"isActive"`
	CreatedByEmployeeID *string   `db:"createdByEmployeeId" json:"createdByEmployeeId"`
	CreatedAt           time.Time `db:"createdAt" json:"createdAt"`
	UpdatedAt           time.Time `db:"updatedAt" json:"updatedAt"`

	// اسم الي أضافها — يرجع فقط للمالك ومدير النظام (يُصفّر لغيرهم بالخدمة)
	CreatedByName *string `db:"createdByName" json:"createdByName"`
}

type UpsertPrivacyPolicyPointRequest struct {
	Content  string `json:"content"`
	Order    *int   `json:"order"`
	IsActive *bool  `json:"isActive"`
}

// PrivacyPolicyStatus حالة موافقة الموظف الحالي — الواجهة تستخدمها حتى تقرر
// إذا تعرض نافذة الموافقة أول ما يسجل دخول.
type PrivacyPolicyStatus struct {
	// NeedsAcceptance يصير true لو ما وافق أبداً، أو لو انضافت نقاط جديدة
	// بعد آخر موافقة (نقارن عدد النقاط الفعّالة بالعدد وقت الموافقة).
	NeedsAcceptance bool                 `json:"needsAcceptance"`
	Points          []PrivacyPolicyPoint `json:"points"`
	AcceptedAt      *time.Time           `json:"acceptedAt"`
}
