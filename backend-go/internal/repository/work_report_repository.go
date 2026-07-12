package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type WorkReportRepository struct {
	db *sqlx.DB
}

func NewWorkReportRepository(db *sqlx.DB) *WorkReportRepository {
	return &WorkReportRepository{db: db}
}

func (r *WorkReportRepository) hydrate(wr *model.WorkReport) {
	var emp model.EmployeeBrief
	if err := r.db.Get(&emp, `SELECT id, name FROM "Employee" WHERE id = $1`, wr.EmployeeID); err == nil {
		wr.Employee = &emp
	}
	var booking model.WorkReportBookingBrief
	err := r.db.Get(&booking, `
		SELECT b.id, b.code, c.name AS "customerName"
		FROM "Booking" b JOIN "Customer" c ON c.id = b."customerId"
		WHERE b.id = $1
	`, wr.BookingID)
	if err == nil {
		wr.Booking = &booking
	}
}

func (r *WorkReportRepository) Create(employeeID string, req model.CreateWorkReportRequest) (*model.WorkReport, error) {
	var wr model.WorkReport
	err := r.db.Get(&wr, `
		INSERT INTO "WorkReport" (id, "bookingId", "employeeId", "workStatus", events, "extraRequests", "cleanedSite", "gaveInfo", "tookPhotos", "stopReason", notes)
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING *
	`, req.BookingID, employeeID, req.WorkStatus, req.Events, req.ExtraRequests, req.CleanedSite, req.GaveInfo, req.TookPhotos, req.StopReason, req.Notes)
	if err != nil {
		return nil, err
	}
	r.hydrate(&wr)
	return &wr, nil
}

func (r *WorkReportRepository) List(employeeID string) ([]model.WorkReport, error) {
	reports := []model.WorkReport{}
	query := `SELECT * FROM "WorkReport" WHERE 1=1`
	args := []any{}
	if employeeID != "" {
		args = append(args, employeeID)
		query += ` AND "employeeId" = $1`
	}
	query += ` ORDER BY "createdAt" DESC LIMIT 200`
	if err := r.db.Select(&reports, query, args...); err != nil {
		return nil, err
	}
	for i := range reports {
		r.hydrate(&reports[i])
	}
	return reports, nil
}
