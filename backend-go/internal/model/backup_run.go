package model

import "time"

// BackupRun نتيجة تشغيل واحد لسكربت النسخ الاحتياطي.
//
// ⚠️ هذي أعمدة الجدول كاملة. الجلب يستعمل SELECT * وأي عمود بلا حقل
// هنا يفشّل الاستعلام كله بالسكوت — إذا ضفت عمود بمايكريشن، ضيف حقله
// هنا بنفس اللحظة.
//
// ⚠️ ما يُعرض إلا للمالك (RequireOwner). شوف schema_backup.go.
type BackupRun struct {
	ID            string     `db:"id" json:"id"`
	StartedAt     time.Time  `db:"startedAt" json:"startedAt"`
	FinishedAt    *time.Time `db:"finishedAt" json:"finishedAt"`
	OK            bool       `db:"ok" json:"ok"`
	FileName      *string    `db:"fileName" json:"fileName"`
	SizeBytes     int64      `db:"sizeBytes" json:"sizeBytes"`
	TableCount    int        `db:"tableCount" json:"tableCount"`
	Encrypted     bool       `db:"encrypted" json:"encrypted"`
	Offsite       bool       `db:"offsite" json:"offsite"`
	OffsiteTarget *string    `db:"offsiteTarget" json:"offsiteTarget"`
	HasUploads    bool       `db:"hasUploads" json:"hasUploads"`
	HasEnv        bool       `db:"hasEnv" json:"hasEnv"`
	Warnings      *string    `db:"warnings" json:"warnings"`
	Error         *string    `db:"error" json:"error"`
	KeptCount     int        `db:"keptCount" json:"keptCount"`
}
