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

func (r *PerformanceReviewRepository) Create(employeeID, evaluatorID, rating, reason string) (*model.PerformanceReview, error) {
	var pr model.PerformanceReview
	err := r.db.Get(&pr, `
		INSERT INTO "PerformanceReview" (id, "employeeId", "evaluatorId", rating, reason)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING *
	`, uuid.NewString(), employeeID, evaluatorID, rating, reason)
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
