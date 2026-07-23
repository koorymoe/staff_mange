package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type KpiRepository struct {
	db *sqlx.DB
}

func NewKpiRepository(db *sqlx.DB) *KpiRepository {
	return &KpiRepository{db: db}
}

func (r *KpiRepository) loadEmployeeBrief(id string) *model.EmployeeBrief {
	var brief model.EmployeeBrief
	if err := r.db.Get(&brief, `SELECT id, name FROM "Employee" WHERE id = $1`, id); err != nil {
		return nil
	}
	return &brief
}

func (r *KpiRepository) hydrate(e *model.KpiEvaluation) {
	e.Employee = r.loadEmployeeBrief(e.EmployeeID)
	e.Evaluator = r.loadEmployeeBrief(e.EvaluatorID)
	if e.CancelledByEmployeeID != nil {
		e.CancelledByEmployee = r.loadEmployeeBrief(*e.CancelledByEmployeeID)
	}
}

func (r *KpiRepository) List() ([]model.KpiEvaluation, error) {
	evals := []model.KpiEvaluation{}
	if err := r.db.Select(&evals, `SELECT * FROM "KpiEvaluation" ORDER BY "createdAt" DESC`); err != nil {
		return nil, err
	}
	for i := range evals {
		r.hydrate(&evals[i])
	}
	return evals, nil
}

func (r *KpiRepository) ListForEmployee(employeeID string) ([]model.KpiEvaluation, error) {
	evals := []model.KpiEvaluation{}
	if err := r.db.Select(&evals, `SELECT * FROM "KpiEvaluation" WHERE "employeeId" = $1 ORDER BY "createdAt" DESC`, employeeID); err != nil {
		return nil, err
	}
	for i := range evals {
		r.hydrate(&evals[i])
	}
	return evals, nil
}

func (r *KpiRepository) Create(employeeID, evaluatorID string, points int, reason string, deductionAmount float64) (*model.KpiEvaluation, error) {
	var e model.KpiEvaluation
	err := r.db.Get(&e, `
		INSERT INTO "KpiEvaluation" (id, "employeeId", "evaluatorId", points, reason, "deductionAmount")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5)
		RETURNING *
	`, employeeID, evaluatorID, points, reason, deductionAmount)
	if err != nil {
		return nil, err
	}
	r.hydrate(&e)
	return &e, nil
}

func (r *KpiRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM "KpiEvaluation" WHERE id = $1`, id)
	return err
}

// Cancel "يرجّع" نقطة كي بي اي — ما تنحذف نهائياً، تنعلّم "ملغاة" حتى يضل
// تاريخها موجود ويشوفه المراقب، بس تأثيرها المالي يوقف يحسب بالمجاميع.
func (r *KpiRepository) Cancel(id, cancelledByEmployeeID string) (*model.KpiEvaluation, error) {
	var e model.KpiEvaluation
	err := r.db.Get(&e, `
		UPDATE "KpiEvaluation" SET
			cancelled = true,
			"cancelledAt" = now(),
			"cancelledByEmployeeId" = $2
		WHERE id = $1
		RETURNING *
	`, id, cancelledByEmployeeID)
	if err != nil {
		return nil, err
	}
	r.hydrate(&e)
	return &e, nil
}

// RoleLeaderboard يرجع ترتيب موظفي دور معيّن حسب مجموع نقاط الـKPI ضمن فترة زمنية،
// حتى يشوف كل موظف ترتيبه بين نظرائه بنفس الدور (فني مع الفنيين، إداري مع الإداريين).
func (r *KpiRepository) RoleLeaderboard(role string, since string) ([]model.KpiLeaderboardEntry, error) {
	entries := []model.KpiLeaderboardEntry{}
	err := r.db.Select(&entries, `
		SELECT
			e.id AS "employeeId",
			e.name AS "employeeName",
			COALESCE(SUM(k.points), 0) AS points,
			COUNT(k.id) AS "evaluationCount",
			COALESCE((
				SELECT COUNT(*) FROM "BookingAssignment" ba
				JOIN "Booking" b ON b.id = ba."bookingId"
				WHERE ba."employeeId" = e.id AND b.status = 'COMPLETED' AND b."completedAt" >= $2::timestamp
			), 0) AS "completedBookings"
		FROM "Employee" e
		LEFT JOIN "KpiEvaluation" k ON k."employeeId" = e.id AND k."createdAt" >= $2::timestamp AND k.cancelled = false
		WHERE e.role = $1 AND e.status = 'ACTIVE'
		GROUP BY e.id, e.name
		ORDER BY points DESC, "completedBookings" DESC
	`, role, since)
	return entries, err
}
