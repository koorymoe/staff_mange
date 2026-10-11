package repository

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type ServiceManagerRepository struct {
	db *sqlx.DB
}

func NewServiceManagerRepository(db *sqlx.DB) *ServiceManagerRepository {
	return &ServiceManagerRepository{db: db}
}

// List يرجع كل موظف مسؤول خدمة، مع الخدمة المرتبطة — يستخدم لبناء "أي موظف مسؤول عن أي خدمة"
func (r *ServiceManagerRepository) List() ([]model.ServiceManager, error) {
	rows := []model.ServiceManager{}
	if err := r.db.Select(&rows, `SELECT * FROM "ServiceManager" ORDER BY "createdAt"`); err != nil {
		return nil, err
	}
	for i := range rows {
		r.hydrate(&rows[i])
	}
	return rows, nil
}

// SetForEmployee يستبدل مجموعة الخدمات اللي الموظف مسؤول عنها بالكامل (حذف القديم وإدخال الجديد)
func (r *ServiceManagerRepository) SetForEmployee(employeeID string, serviceIDs []string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM "ServiceManager" WHERE "employeeId" = $1`, employeeID); err != nil {
		return err
	}
	for _, serviceID := range serviceIDs {
		if _, err := tx.Exec(`
			INSERT INTO "ServiceManager" (id, "employeeId", "serviceId")
			VALUES ($1, $2, $3) ON CONFLICT DO NOTHING
		`, uuid.NewString(), employeeID, serviceID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *ServiceManagerRepository) hydrate(sm *model.ServiceManager) {
	var emp model.EmployeeBrief
	if err := r.db.Get(&emp, `SELECT id, name, position FROM "Employee" WHERE id = $1`, sm.EmployeeID); err == nil {
		sm.Employee = &emp
	}
	var svc model.Service
	if err := r.db.Get(&svc, `SELECT * FROM "Service" WHERE id = $1`, sm.ServiceID); err == nil {
		sm.Service = &svc
	}
}
