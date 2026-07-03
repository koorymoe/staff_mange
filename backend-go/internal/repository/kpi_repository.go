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
