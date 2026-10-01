package model

import "time"

// MatrixGuideRule تعليمة توجيه لماتركس — يعدّلها المدير.
type MatrixGuideRule struct {
	ID            string    `db:"id" json:"id"`
	Route         string    `db:"route" json:"route"`
	Match         string    `db:"match" json:"match"`
	Groups        string    `db:"groups" json:"groups"`
	Text          string    `db:"text" json:"text"`
	OnlyIfPending bool      `db:"onlyIfPending" json:"onlyIfPending"`
	Priority      int       `db:"priority" json:"priority"`
	Enabled       bool      `db:"enabled" json:"enabled"`
	CreatedAt     time.Time `db:"createdAt" json:"createdAt"`
}
