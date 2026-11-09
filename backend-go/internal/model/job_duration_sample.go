package model

import "time"

// JobDurationSample عيّنة زمنية حقيقية واحدة من عمل مُنجز فعلاً (تركيب أو صيانة) —
// تُستخدم لتعليم النظام وتيرة العمل الحقيقية تلقائياً (بدون رقم مفروض يدوياً)،
// مقسّمة حسب نوع المنظومة (systemName) ونوع العمل (jobType: تركيب/صيانة) لأن
// كل منظومة ونوع عمل لهما وتيرة مختلفة تماماً.
type JobDurationSample struct {
	ID                        string    `db:"id" json:"id"`
	SystemName                string    `db:"systemName" json:"systemName"`
	JobType                   string    `db:"jobType" json:"jobType"` // INSTALL | MAINTENANCE
	ItemCount                 int       `db:"itemCount" json:"itemCount"`
	CrewSize                  int       `db:"crewSize" json:"crewSize"`
	DurationMinutes           int       `db:"durationMinutes" json:"durationMinutes"`
	BookingID                 *string   `db:"bookingId" json:"bookingId"`
	DeviceMaintenanceTicketID *string   `db:"deviceMaintenanceTicketId" json:"deviceMaintenanceTicketId"`
	CreatedAt                 time.Time `db:"createdAt" json:"createdAt"`
}

const (
	JobTypeInstall     = "INSTALL"
	JobTypeMaintenance = "MAINTENANCE"
)
