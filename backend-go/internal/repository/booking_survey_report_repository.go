package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type BookingSurveyReportRepository struct {
	db *sqlx.DB
}

func NewBookingSurveyReportRepository(db *sqlx.DB) *BookingSurveyReportRepository {
	return &BookingSurveyReportRepository{db: db}
}

func (r *BookingSurveyReportRepository) hydrate(sr *model.BookingSurveyReport) {
	var emp model.EmployeeBrief
	if err := r.db.Get(&emp, `SELECT id, name FROM "Employee" WHERE id = $1`, sr.EmployeeID); err == nil {
		sr.Employee = &emp
	}
}

func (r *BookingSurveyReportRepository) Create(employeeID string, req model.CreateBookingSurveyReportRequest) (*model.BookingSurveyReport, error) {
	var sr model.BookingSurveyReport
	err := r.db.Get(&sr, `
		INSERT INTO "BookingSurveyReport" (id, "bookingId", "employeeId", "customerWants", "siteAreaSqm", "siteDetails", "otherNotes")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $6)
		RETURNING *
	`, req.BookingID, employeeID, req.CustomerWants, req.SiteAreaSqm, req.SiteDetails, req.OtherNotes)
	if err != nil {
		return nil, err
	}
	r.hydrate(&sr)
	return &sr, nil
}

func (r *BookingSurveyReportRepository) List(bookingID string) ([]model.BookingSurveyReport, error) {
	reports := []model.BookingSurveyReport{}
	if err := r.db.Select(&reports, `
		SELECT * FROM "BookingSurveyReport" WHERE "bookingId" = $1 ORDER BY "createdAt" DESC
	`, bookingID); err != nil {
		return nil, err
	}
	for i := range reports {
		r.hydrate(&reports[i])
	}
	return reports, nil
}
