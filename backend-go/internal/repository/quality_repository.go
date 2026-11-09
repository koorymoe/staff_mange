package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type QualityRepository struct {
	db *sqlx.DB
}

func NewQualityRepository(db *sqlx.DB) *QualityRepository {
	return &QualityRepository{db: db}
}

func (r *QualityRepository) loadEmployeeBrief(id *string) *model.EmployeeBrief {
	if id == nil {
		return nil
	}
	var brief model.EmployeeBrief
	if err := r.db.Get(&brief, `SELECT id, name FROM "Employee" WHERE id = $1`, *id); err != nil {
		return nil
	}
	return &brief
}

func (r *QualityRepository) hydrate(i *model.QualityIssue) {
	i.ResponsibleEmployee = r.loadEmployeeBrief(i.ResponsibleEmployeeID)
	i.ReportedBy = r.loadEmployeeBrief(i.ReportedByID)
}

func (r *QualityRepository) List(category string) ([]model.QualityIssue, error) {
	issues := []model.QualityIssue{}
	query := `SELECT * FROM "QualityIssue" WHERE 1=1`
	args := []any{}
	if category != "" {
		args = append(args, category)
		query += ` AND category = $1`
	}
	query += ` ORDER BY "createdAt" DESC`
	if err := r.db.Select(&issues, query, args...); err != nil {
		return nil, err
	}
	for i := range issues {
		r.hydrate(&issues[i])
	}
	return issues, nil
}

func (r *QualityRepository) Create(req model.CreateQualityIssueRequest, reportedByID string) (*model.QualityIssue, error) {
	var i model.QualityIssue
	err := r.db.Get(&i, `
		INSERT INTO "QualityIssue" (id, category, title, description, "responsibleEmployeeId", "bookingId", "reportedById")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $6)
		RETURNING *
	`, req.Category, req.Title, req.Description, req.ResponsibleEmployeeID, req.BookingID, reportedByID)
	if err != nil {
		return nil, err
	}
	r.hydrate(&i)
	return &i, nil
}

func (r *QualityRepository) Update(id string, req model.UpdateQualityIssueRequest) (*model.QualityIssue, error) {
	var resolvedAtExpr string
	if req.Status != nil && *req.Status == "RESOLVED" {
		resolvedAtExpr = "now()"
	} else {
		resolvedAtExpr = `"resolvedAt"`
	}
	var i model.QualityIssue
	err := r.db.Get(&i, `
		UPDATE "QualityIssue" SET status = COALESCE($2, status), "resolvedAt" = `+resolvedAtExpr+`
		WHERE id = $1
		RETURNING *
	`, id, req.Status)
	if err != nil {
		return nil, err
	}
	r.hydrate(&i)
	return &i, nil
}
