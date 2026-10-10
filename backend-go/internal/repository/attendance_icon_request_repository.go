package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type AttendanceIconRequestRepository struct {
	db *sqlx.DB
}

func NewAttendanceIconRequestRepository(db *sqlx.DB) *AttendanceIconRequestRepository {
	return &AttendanceIconRequestRepository{db: db}
}

func (r *AttendanceIconRequestRepository) hydrate(items []model.AttendanceIconRequest) {
	for i := range items {
		var e model.Employee
		if err := r.db.Get(&e, `SELECT * FROM "Employee" WHERE id = $1`, items[i].EmployeeID); err == nil {
			items[i].Employee = &model.EmployeeBrief{ID: e.ID, Name: e.Name}
		}
	}
}

func (r *AttendanceIconRequestRepository) ListPending() ([]model.AttendanceIconRequest, error) {
	items := []model.AttendanceIconRequest{}
	if err := r.db.Select(&items, `SELECT * FROM "AttendanceIconRequest" WHERE status = 'PENDING' ORDER BY "createdAt" ASC`); err != nil {
		return nil, err
	}
	r.hydrate(items)
	return items, nil
}

func (r *AttendanceIconRequestRepository) Create(id, employeeID, requestedIcon string) (*model.AttendanceIconRequest, error) {
	var req model.AttendanceIconRequest
	err := r.db.Get(&req, `
		INSERT INTO "AttendanceIconRequest" (id, "employeeId", "requestedIcon")
		VALUES ($1, $2, $3)
		RETURNING *
	`, id, employeeID, requestedIcon)
	if err != nil {
		return nil, err
	}
	items := []model.AttendanceIconRequest{req}
	r.hydrate(items)
	return &items[0], nil
}

func (r *AttendanceIconRequestRepository) FindByID(id string) (*model.AttendanceIconRequest, error) {
	var req model.AttendanceIconRequest
	if err := r.db.Get(&req, `SELECT * FROM "AttendanceIconRequest" WHERE id = $1`, id); err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *AttendanceIconRequestRepository) Resolve(id, status, resolvedByID string) error {
	_, err := r.db.Exec(`
		UPDATE "AttendanceIconRequest" SET status = $2, "resolvedAt" = now(), "resolvedById" = $3
		WHERE id = $1
	`, id, status, resolvedByID)
	return err
}
