package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type ServiceStudyRepository struct {
	db *sqlx.DB
}

func NewServiceStudyRepository(db *sqlx.DB) *ServiceStudyRepository {
	return &ServiceStudyRepository{db: db}
}

func (r *ServiceStudyRepository) hydrate(items []model.ServiceStudy) {
	for i := range items {
		s := &items[i]
		var creator model.Employee
		if err := r.db.Get(&creator, `SELECT * FROM "Employee" WHERE id = $1`, s.CreatedByID); err == nil {
			s.CreatedBy = &model.EmployeeBrief{ID: creator.ID, Name: creator.Name}
		}
		assigned := []model.EmployeeBrief{}
		_ = r.db.Select(&assigned, `
			SELECT e.id, e.name FROM "ServiceStudyAssignment" a
			JOIN "Employee" e ON e.id = a."employeeId"
			WHERE a."serviceStudyId" = $1
		`, s.ID)
		s.AssignedEmployees = assigned

		reports := []model.ServiceStudyReport{}
		_ = r.db.Select(&reports, `SELECT * FROM "ServiceStudyReport" WHERE "serviceStudyId" = $1 ORDER BY "createdAt" DESC`, s.ID)
		for i := range reports {
			var emp model.Employee
			if err := r.db.Get(&emp, `SELECT * FROM "Employee" WHERE id = $1`, reports[i].EmployeeID); err == nil {
				reports[i].Employee = &model.EmployeeBrief{ID: emp.ID, Name: emp.Name}
			}
		}
		s.Reports = reports
	}
}

func (r *ServiceStudyRepository) List() ([]model.ServiceStudy, error) {
	items := []model.ServiceStudy{}
	if err := r.db.Select(&items, `SELECT * FROM "ServiceStudy" ORDER BY archived ASC, "createdAt" DESC`); err != nil {
		return nil, err
	}
	r.hydrate(items)
	return items, nil
}

func (r *ServiceStudyRepository) FindByID(id string) (*model.ServiceStudy, error) {
	var s model.ServiceStudy
	if err := r.db.Get(&s, `SELECT * FROM "ServiceStudy" WHERE id = $1`, id); err != nil {
		return nil, err
	}
	items := []model.ServiceStudy{s}
	r.hydrate(items)
	return &items[0], nil
}

func (r *ServiceStudyRepository) Create(id, name, createdByID string) (*model.ServiceStudy, error) {
	var s model.ServiceStudy
	err := r.db.Get(&s, `
		INSERT INTO "ServiceStudy" (id, name, "createdById") VALUES ($1, $2, $3) RETURNING *
	`, id, name, createdByID)
	if err != nil {
		return nil, err
	}
	items := []model.ServiceStudy{s}
	r.hydrate(items)
	return &items[0], nil
}

func (r *ServiceStudyRepository) SetAssignments(serviceStudyID string, employeeIDs []string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM "ServiceStudyAssignment" WHERE "serviceStudyId" = $1`, serviceStudyID); err != nil {
		return err
	}
	for _, empID := range employeeIDs {
		if _, err := tx.Exec(`
			INSERT INTO "ServiceStudyAssignment" (id, "serviceStudyId", "employeeId")
			VALUES (gen_random_uuid()::text, $1, $2)
		`, serviceStudyID, empID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *ServiceStudyRepository) IsAssigned(serviceStudyID, employeeID string) (bool, error) {
	var count int
	err := r.db.Get(&count, `
		SELECT COUNT(*) FROM "ServiceStudyAssignment" WHERE "serviceStudyId" = $1 AND "employeeId" = $2
	`, serviceStudyID, employeeID)
	return count > 0, err
}

func (r *ServiceStudyRepository) AddReport(id, serviceStudyID, employeeID, content string) (*model.ServiceStudyReport, error) {
	var rep model.ServiceStudyReport
	err := r.db.Get(&rep, `
		INSERT INTO "ServiceStudyReport" (id, "serviceStudyId", "employeeId", content)
		VALUES ($1, $2, $3, $4)
		RETURNING *
	`, id, serviceStudyID, employeeID, content)
	if err != nil {
		return nil, err
	}
	var emp model.Employee
	if err := r.db.Get(&emp, `SELECT * FROM "Employee" WHERE id = $1`, employeeID); err == nil {
		rep.Employee = &model.EmployeeBrief{ID: emp.ID, Name: emp.Name}
	}
	return &rep, nil
}

func (r *ServiceStudyRepository) Archive(id string) (*model.ServiceStudy, error) {
	var s model.ServiceStudy
	err := r.db.Get(&s, `UPDATE "ServiceStudy" SET archived = true WHERE id = $1 RETURNING *`, id)
	if err != nil {
		return nil, err
	}
	items := []model.ServiceStudy{s}
	r.hydrate(items)
	return &items[0], nil
}
