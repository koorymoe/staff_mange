package repository

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type StaffRequestRepository struct {
	db *sqlx.DB
}

func NewStaffRequestRepository(db *sqlx.DB) *StaffRequestRepository {
	return &StaffRequestRepository{db: db}
}

func (r *StaffRequestRepository) Create(requesterID string, req model.CreateStaffRequest) (*model.StaffRequest, error) {
	id := uuid.NewString()
	if _, err := r.db.Exec(`
		INSERT INTO "StaffRequest" (id, "requesterId", "projectId", "neededAt", "durationHours", notes)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, id, requesterID, req.ProjectID, req.NeededAt, req.DurationHours, req.Notes); err != nil {
		return nil, err
	}
	for _, empID := range req.EmployeeIDs {
		if _, err := r.db.Exec(`
			INSERT INTO "StaffRequestEmployee" (id, "requestId", "employeeId")
			VALUES ($1, $2, $3) ON CONFLICT DO NOTHING
		`, uuid.NewString(), id, empID); err != nil {
			return nil, err
		}
	}
	return r.Get(id)
}

func (r *StaffRequestRepository) Get(id string) (*model.StaffRequest, error) {
	var sr model.StaffRequest
	if err := r.db.Get(&sr, `SELECT * FROM "StaffRequest" WHERE id = $1`, id); err != nil {
		return nil, err
	}
	r.hydrate(&sr)
	return &sr, nil
}

// List يرجع الطلبات: كلها (لإدارة الكوادر/الأدمن) أو مال طالب معيّن فقط
func (r *StaffRequestRepository) List(requesterID string) ([]model.StaffRequest, error) {
	list := []model.StaffRequest{}
	var err error
	if requesterID == "" {
		err = r.db.Select(&list, `SELECT * FROM "StaffRequest" ORDER BY "createdAt" DESC`)
	} else {
		err = r.db.Select(&list, `SELECT * FROM "StaffRequest" WHERE "requesterId" = $1 ORDER BY "createdAt" DESC`, requesterID)
	}
	if err != nil {
		return nil, err
	}
	for i := range list {
		r.hydrate(&list[i])
	}
	return list, nil
}

func (r *StaffRequestRepository) UpdateStatus(id, status, handledByID string) (*model.StaffRequest, error) {
	if _, err := r.db.Exec(`
		UPDATE "StaffRequest" SET status = $2, "handledById" = $3, "handledAt" = CURRENT_TIMESTAMP WHERE id = $1
	`, id, status, handledByID); err != nil {
		return nil, err
	}
	return r.Get(id)
}

func (r *StaffRequestRepository) hydrate(sr *model.StaffRequest) {
	var requester model.EmployeeBrief
	if err := r.db.Get(&requester, `SELECT id, name, position FROM "Employee" WHERE id = $1`, sr.RequesterID); err == nil {
		sr.Requester = &requester
	}
	if sr.HandledByID != nil {
		var handler model.EmployeeBrief
		if err := r.db.Get(&handler, `SELECT id, name, position FROM "Employee" WHERE id = $1`, *sr.HandledByID); err == nil {
			sr.HandledBy = &handler
		}
	}
	if sr.ProjectID != nil {
		var name string
		if err := r.db.Get(&name, `SELECT name FROM "Project" WHERE id = $1`, *sr.ProjectID); err == nil {
			sr.ProjectName = &name
		}
	}
	sr.Employees = []model.EmployeeBrief{}
	_ = r.db.Select(&sr.Employees, `
		SELECT e.id, e.name, e.position FROM "StaffRequestEmployee" sre
		JOIN "Employee" e ON e.id = sre."employeeId"
		WHERE sre."requestId" = $1 ORDER BY e.name
	`, sr.ID)
}
