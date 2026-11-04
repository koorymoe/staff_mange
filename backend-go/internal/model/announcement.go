package model

import "time"

// إعلان يمر بشريط متحرك كدام كل الموظفين — المالك ومدير النظام بس
// يقدرون ينزّلونه.
type Announcement struct {
	ID          string    `db:"id" json:"id"`
	Body        string    `db:"body" json:"body"`
	Active      bool      `db:"active" json:"active"`
	CreatedByID string    `db:"createdById" json:"createdById"`
	CreatedAt   time.Time `db:"createdAt" json:"createdAt"`
	// ExpiresAt فاضي = بلا انتهاء. إعلانات المخالفات تنتهي بعد 3 أيام.
	ExpiresAt *time.Time `db:"expiresAt" json:"expiresAt"`

	CreatedByName string `db:"createdByName" json:"createdByName"`
}

type CreateAnnouncementRequest struct {
	Body string `json:"body"`
	// ExpiresInDays فاضي/صفر = بلا انتهاء
	ExpiresInDays int `json:"expiresInDays"`
}

// AnnouncementPenaltyDays مدة إعلان المخالفة بالشريط.
const AnnouncementPenaltyDays = 3
