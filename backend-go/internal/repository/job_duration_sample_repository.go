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

// EmployeeSpeedRatio يقيس سرعة الموظف بشهر معيّن.
//
// الفكرة: كل عيّنة زمنية مربوطة بحجز، والحجز إله كادر. نقارن الزمن الي
// أخذه هذا الكادر بمتوسط الزمن الي تاخذه نفس المنظومة بنفس نوع الشغل
// عند بقية الكوادر.
//
//	النسبة = المتوسط العام ÷ زمن هذا الموظف
//
// أكبر من ١ يعني أسرع من المتوسط، أصغر من ١ يعني أبطأ. اخترت هذا
// الاتجاه حتى الرقم الأعلى دائماً أفضل — متل باقي مؤشرات الصفحة، فما
// يحتاج المدير ينتبه إن هذا العمود بالعكس.
//
// المقارنة تصير على نفس المنظومة ونفس نوع الشغل بس: تركيب كاميرات مو
// نفس تركيب جي بي اس، ومقارنتهن ببعض تطلع رقم بلا معنى.
//
// نحتاج على الأقل عيّنتين بالمجموعة حتى يصير عدنا «متوسط» نقارن بيه —
// عيّنة وحدة تقارن نفسها بنفسها وتطلع ١ دائماً.
func (r *JobDurationSampleRepository) EmployeeSpeedRatio(employeeID, month string) (*float64, int, error) {
	var row struct {
		Ratio   *float64 `db:"ratio"`
		Samples int      `db:"samples"`
	}
	err := r.db.Get(&row, `
		WITH averages AS (
			-- متوسط زمن كل منظومة/نوع شغل عند الكل، معدّل على حجم الكادر
			-- (شغل ساعتين بفني واحد مو نفس شغل ساعتين بثلاث فنيين)
			SELECT "systemName", "jobType",
			       AVG("durationMinutes"::numeric * GREATEST("crewSize", 1) / GREATEST("itemCount", 1)) AS avg_effort,
			       COUNT(*) AS n
			FROM "JobDurationSample"
			GROUP BY "systemName", "jobType"
			HAVING COUNT(*) >= 2
		),
		mine AS (
			SELECT j."systemName", j."jobType",
			       j."durationMinutes"::numeric * GREATEST(j."crewSize", 1) / GREATEST(j."itemCount", 1) AS effort
			FROM "JobDurationSample" j
			JOIN "Booking" b ON b.id = j."bookingId"
			JOIN "BookingAssignment" ba ON ba."bookingId" = b.id
			WHERE ba."employeeId" = $1
			  AND to_char(j."createdAt", 'YYYY-MM') = $2
		)
		SELECT AVG(a.avg_effort / NULLIF(m.effort, 0))::float8 AS ratio,
		       COUNT(*)::int AS samples
		FROM mine m
		JOIN averages a ON a."systemName" = m."systemName" AND a."jobType" = m."jobType"
	`, employeeID, month)
	if err != nil {
		return nil, 0, err
	}
	return row.Ratio, row.Samples, nil
}
