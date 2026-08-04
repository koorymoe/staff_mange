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

	CreatedByName string `db:"createdByName" json:"createdByName"`
}

type CreateAnnouncementRequest struct {
	Body string `json:"body"`
}
