package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type JobDurationSampleRepository struct {
	db *sqlx.DB
}

func NewJobDurationSampleRepository(db *sqlx.DB) *JobDurationSampleRepository {
	return &JobDurationSampleRepository{db: db}
}

// Create يحفظ عيّنة زمنية جديدة (تركيب أو صيانة) — تُستدعى تلقائياً من نقاط الإنجاز
// الحقيقية (فاتورة ليدر مكتملة / تذكرة صيانة مُسلَّمة)، لا يوجد إدخال يدوي لها.
func (r *JobDurationSampleRepository) Create(s model.JobDurationSample) error {
	_, err := r.db.Exec(`
		INSERT INTO "JobDurationSample"
			(id, "systemName", "jobType", "itemCount", "crewSize", "durationMinutes", "bookingId", "deviceMaintenanceTicketId")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $6, $7)
	`, s.SystemName, s.JobType, s.ItemCount, s.CrewSize, s.DurationMinutes, s.BookingID, s.DeviceMaintenanceTicketID)
	return err
}

// ListBySystemAndJobType يرجّع كل العيّنات المطابقة لنفس المنظومة ونوع العمل —
// أساس حساب الوتيرة المتعلَّمة (learnedPacePerItemPerWorker).
func (r *JobDurationSampleRepository) ListBySystemAndJobType(systemName, jobType string) ([]model.JobDurationSample, error) {
	samples := []model.JobDurationSample{}
	err := r.db.Select(&samples, `
		SELECT * FROM "JobDurationSample" WHERE "systemName" = $1 AND "jobType" = $2
	`, systemName, jobType)
	return samples, err
}
