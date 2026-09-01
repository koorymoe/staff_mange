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

func (r *PerformanceReviewRepository) Create(employeeID, evaluatorID, rating, reason string, bookingID *string, commitment, speed, quality *int) (*model.PerformanceReview, error) {
	var pr model.PerformanceReview
	// ⚠️ ON CONFLICT يحدّث بدل ما يفشل: الليدر يغيّر رأيه ويعيد التقييم
	// لنفس الحجز — ما يصير نرفضه برسالة «انقيّم قبل» ويضطر يتصل بالدعم.
	// الفهرس الفريد يحمي من تقييمين، وهاي تخلي التعديل طبيعي.
	err := r.db.Get(&pr, `
		INSERT INTO "PerformanceReview" (id, "employeeId", "evaluatorId", rating, reason, "bookingId",
		                                 "commitmentScore", "speedScore", "qualityScore")
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT ("bookingId", "employeeId") WHERE "bookingId" IS NOT NULL
		DO UPDATE SET rating = EXCLUDED.rating, reason = EXCLUDED.reason,
		              "evaluatorId" = EXCLUDED."evaluatorId",
		              -- ⚠️ COALESCE: إعادة التقييم بلا نجوم ما تمحي النجوم
		              -- الي انطاها قبل. الليدر يعدّل الحكم بضغطة وحدة،
		              -- وما يتوقع إن تفاصيل نطّاها أمس تنمسح بصمت.
		              "commitmentScore" = COALESCE(EXCLUDED."commitmentScore", "PerformanceReview"."commitmentScore"),
		              "speedScore"      = COALESCE(EXCLUDED."speedScore",      "PerformanceReview"."speedScore"),
		              "qualityScore"    = COALESCE(EXCLUDED."qualityScore",    "PerformanceReview"."qualityScore")
		RETURNING *
	`, uuid.NewString(), employeeID, evaluatorID, rating, reason, bookingID, commitment, speed, quality)
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
func (r *PerformanceReviewRepository) BookingsAwaitingReview(
	viewerID string, allBookings bool, from, to string,
) ([]model.BookingAwaitingReview, error) {
	// ⚠️ المدير/المراقب مو طرف بأي حجز، فلو خلّينا شرط «لازم يكون هو
	// بالفريق» يشوفون صفراً دائماً — وهاي چانت علّة الشاشة الفارغة.
	// ولنفس السبب شرط «لازم وياه غيره» ينشال إلهم: هما أصلاً برّا
	// الفريق، فحجز الفني الواحد (جي بي اس وداش كام) إله منو يتقيّم.
	var args []any
	var scope, soloGuard, fromP, toP string
	if allBookings {
		args = []any{from, to}
		fromP, toP = "$1", "$2"
		soloGuard = `
		  -- لازم يكون مكلّف بيه أحد أصلاً، ولو واحد
		  AND EXISTS (
		      SELECT 1 FROM "BookingAssignment" any_one
		      WHERE any_one."bookingId" = b.id
		  )`
	} else {
		args = []any{viewerID, from, to}
		fromP, toP = "$2", "$3"
		scope = `
		JOIN "BookingAssignment" mine ON mine."bookingId" = b.id AND mine."employeeId" = $1`
		soloGuard = `
		  -- لازم يكون وياه غيره: حجز طلع بيه لحاله ماكو منو يتقيّم
		  AND EXISTS (
		      SELECT 1 FROM "BookingAssignment" other
		      WHERE other."bookingId" = b.id AND other."employeeId" <> $1
		  )`
	}

	rows := []model.BookingAwaitingReview{}
	err := r.db.Select(&rows, `
		SELECT DISTINCT
			b.id            AS "bookingId",
			b.code          AS code,
			COALESCE(c.name, '—') AS "customerName",
			s.name          AS "serviceName",
			c.phone         AS "customerPhone",
			c.location      AS "customerAddress",
			b."completedAt" AS "completedAt"
		FROM "Booking" b`+scope+`
		LEFT JOIN "Customer" c ON c.id = b."customerId"
		LEFT JOIN "Service"  s ON s.id = b."serviceId"
		WHERE b.status = 'COMPLETED'
		  AND b."archivedAt" IS NULL
		  AND (`+fromP+` = '' OR b."completedAt" >= `+fromP+`::timestamp)
		  AND (`+toP+` = '' OR b."completedAt" < (`+toP+`::date + interval '1 day'))`+soloGuard+`
		ORDER BY b."completedAt" DESC
		LIMIT 100
	`, args...)
	if err != nil {
		return nil, err
	}

	for i := range rows {
		crew := []model.CrewReviewState{}
		if err := r.db.Select(&crew, `
			SELECT e.id AS "employeeId", e.name, e.position,
			       pr.rating, pr.reason,
			       pr."commitmentScore", pr."speedScore", pr."qualityScore"
			FROM "BookingAssignment" ba
			JOIN "Employee" e ON e.id = ba."employeeId"
			LEFT JOIN "PerformanceReview" pr
			       ON pr."bookingId" = ba."bookingId"
			      AND pr."employeeId" = e.id
			WHERE ba."bookingId" = $1 AND e.id <> $2
			ORDER BY e.name
		`, rows[i].BookingID, viewerID); err != nil {
			return nil, err
		}
		rows[i].Crew = crew
	}
	return rows, nil
}
