package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type EmployeeRepository struct {
	db *sqlx.DB
}

func NewEmployeeRepository(db *sqlx.DB) *EmployeeRepository {
	return &EmployeeRepository{db: db}
}

func (r *EmployeeRepository) List() ([]model.Employee, error) {
	var employees []model.Employee
	err := r.db.Select(&employees, `SELECT * FROM "Employee" ORDER BY name ASC`)
	return employees, err
}

func (r *EmployeeRepository) FindByID(id string) (*model.Employee, error) {
	var e model.Employee
	err := r.db.Get(&e, `SELECT * FROM "Employee" WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EmployeeRepository) FindByUsername(username string) (*model.Employee, error) {
	var e model.Employee
	err := r.db.Get(&e, `SELECT * FROM "Employee" WHERE username = $1`, username)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EmployeeRepository) Create(e *model.Employee) error {
	_, err := r.db.NamedExec(`
		INSERT INTO "Employee" (id, name, certificate, position, phone, username, password)
		VALUES (:id, :name, :certificate, :position, :phone, :username, :password)
	`, e)
	return err
}

func (r *EmployeeRepository) Update(e *model.Employee) error {
	_, err := r.db.NamedExec(`
		UPDATE "Employee" SET
			name = :name,
			certificate = :certificate,
			position = :position,
			phone = :phone,
			status = :status,
			role = :role,
			"onDuty" = :onDuty,
			username = :username,
			password = COALESCE(NULLIF(:password, ''), password),
			"hasDrivingLicense" = :hasDrivingLicense,
			"hasSafetyCertificate" = :hasSafetyCertificate,
			"isLeader" = :isLeader,
			"isTrainee" = :isTrainee
		WHERE id = :id
	`, e)
	return err
}
