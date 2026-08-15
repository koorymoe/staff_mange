package repository

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type PerformanceReviewRepository struct {
	db *sqlx.DB
}

func NewPerformanceReviewRepository(db *sqlx.DB) *PerformanceReviewRepository {
	return &PerformanceReviewRepository{db: db}
}

func (r *PerformanceReviewRepository) Create(employeeID, evaluatorID, rating, reason string, bookingID *string) (*model.PerformanceReview, error) {
	var pr model.PerformanceReview
	// ⚠️ ON CONFLICT يحدّث بدل ما يفشل: الليدر يغيّر رأيه ويعيد التقييم
	// لنفس الحجز — ما يصير نرفضه برسالة «انقيّم قبل» ويضطر يتصل بالدعم.
	// الفهرس الفريد يحمي من تقييمين، وهاي تخلي التعديل طبيعي.
	err := r.db.Get(&pr, `
		INSERT INTO "PerformanceReview" (id, "employeeId", "evaluatorId", rating, reason, "bookingId")
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT ("bookingId", "employeeId") WHERE "bookingId" IS NOT NULL
		DO UPDATE SET rating = EXCLUDED.rating, reason = EXCLUDED.reason,
		              "evaluatorId" = EXCLUDED."evaluatorId"
		RETURNING *
	`, uuid.NewString(), employeeID, evaluatorID, rating, reason, bookingID)
	if err != nil {
		return nil, err
	}
	r.hydrate(&pr)
	return &pr, nil
}

func (r *PerformanceReviewRepository) ListForEmployee(employeeID string) ([]model.PerformanceReview, error) {
	rows := []model.PerformanceReview{}
	if err := r.db.Select(&rows, `SELECT * FROM "PerformanceReview" WHERE "employeeId" = $1 ORDER BY "createdAt" DESC`, employeeID); err != nil {
		return nil, err
	}
	for i := range rows {
		r.hydrate(&rows[i])
	}
	return rows, nil
}

func (r *PerformanceReviewRepository) List() ([]model.PerformanceReview, error) {
	rows := []model.PerformanceReview{}
	if err := r.db.Select(&rows, `SELECT * FROM "PerformanceReview" ORDER BY "createdAt" DESC`); err != nil {
		return nil, err
	}
	for i := range rows {
		r.hydrate(&rows[i])
	}
	return rows, nil
}

func (r *PerformanceReviewRepository) hydrate(pr *model.PerformanceReview) {
	var emp model.EmployeeBrief
	if err := r.db.Get(&emp, `SELECT id, name, position FROM "Employee" WHERE id = $1`, pr.EmployeeID); err == nil {
		pr.Employee = &emp
	}
	var evaluator model.EmployeeBrief
	if err := r.db.Get(&evaluator, `SELECT id, name, position FROM "Employee" WHERE id = $1`, pr.EvaluatorID); err == nil {
		pr.Evaluator = &evaluator
	}
}


// ═══ حجوزات الليدر الي تنتظر تقييم كادرها ═══
//
// «الليدر يكدر يقيّم فريقه لكل حجز يطلعوله، مو مرة وحدة باليوم».
//
// نرجّع الحجوزات المنجزة الي الليدر طرف بيها، ومعاها كل واحد طلع
// وياه وحالة تقييمه.
//
// ⚠️ نستثني الليدر نفسه من قائمة الكادر: ما ينفع يقيّم روحه.
//
// ⚠️ ونحدّها بآخر ٣٠ يوم: تقييم شغلة صارت قبل ثلاثة أشهر ما يفيد —
// الليدر ما يتذكرها، والملاحظة تطلع من الذاكرة مو من الواقع. والقائمة
// الي تطول للأبد تنتجاهل كلها.
func (r *PerformanceReviewRepository) BookingsAwaitingReview(leaderID string) ([]model.BookingAwaitingReview, error) {
	rows := []model.BookingAwaitingReview{}
	err := r.db.Select(&rows, `
		SELECT DISTINCT
			b.id            AS "bookingId",
			b.code          AS code,
			COALESCE(c.name, '—') AS "customerName",
			s.name          AS "serviceName",
			b."completedAt" AS "completedAt"
		FROM "Booking" b
		JOIN "BookingAssignment" mine ON mine."bookingId" = b.id AND mine."employeeId" = $1
		LEFT JOIN "Customer" c ON c.id = b."customerId"
		LEFT JOIN "Service"  s ON s.id = b."serviceId"
		WHERE b.status = 'COMPLETED'
		  AND b."archivedAt" IS NULL
		  AND b."completedAt" >= now() - interval '30 days'
		  -- لازم يكون وياه غيره: حجز طلع بيه لحاله ماكو منو يتقيّم
		  AND EXISTS (
		      SELECT 1 FROM "BookingAssignment" other
		      WHERE other."bookingId" = b.id AND other."employeeId" <> $1
		  )
		ORDER BY b."completedAt" DESC
		LIMIT 100
	`, leaderID)
	if err != nil {
		return nil, err
	}

	for i := range rows {
		crew := []model.CrewReviewState{}
		if err := r.db.Select(&crew, `
			SELECT e.id AS "employeeId", e.name, e.position,
			       pr.rating, pr.reason
			FROM "BookingAssignment" ba
			JOIN "Employee" e ON e.id = ba."employeeId"
			LEFT JOIN "PerformanceReview" pr
			       ON pr."bookingId" = ba."bookingId"
			      AND pr."employeeId" = e.id
			WHERE ba."bookingId" = $1 AND e.id <> $2
			ORDER BY e.name
		`, rows[i].BookingID, leaderID); err != nil {
			return nil, err
		}
		rows[i].Crew = crew
	}
	return rows, nil
}
