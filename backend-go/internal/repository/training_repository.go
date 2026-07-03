package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type TrainingRepository struct {
	db *sqlx.DB
}

func NewTrainingRepository(db *sqlx.DB) *TrainingRepository {
	return &TrainingRepository{db: db}
}

func (r *TrainingRepository) getService(id string) (*model.Service, error) {
	var s model.Service
	if err := r.db.Get(&s, `SELECT id, name, category, "createdAt" FROM "Service" WHERE id = $1`, id); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *TrainingRepository) AssignedServiceIDs(employeeID string) ([]string, error) {
	var ids []string
	err := r.db.Select(&ids, `SELECT "serviceId" FROM "EmployeeTrainingAssignment" WHERE "employeeId" = $1`, employeeID)
	return ids, err
}

func (r *TrainingRepository) AssignedServices(employeeID string) ([]model.Service, error) {
	var services []model.Service
	err := r.db.Select(&services, `
		SELECT s.id, s.name, s.category, s."createdAt"
		FROM "EmployeeTrainingAssignment" a
		JOIN "Service" s ON s.id = a."serviceId"
		WHERE a."employeeId" = $1
	`, employeeID)
	return services, err
}

func (r *TrainingRepository) MaterialsForServices(serviceIDs []string) ([]model.TrainingMaterial, error) {
	if len(serviceIDs) == 0 {
		return []model.TrainingMaterial{}, nil
	}
	var materials []model.TrainingMaterial
	query, args, err := sqlx.In(`SELECT * FROM "TrainingMaterial" WHERE "serviceId" IN (?) ORDER BY "serviceId" ASC, "order" ASC`, serviceIDs)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)
	if err := r.db.Select(&materials, query, args...); err != nil {
		return nil, err
	}
	for i := range materials {
		materials[i].Service, _ = r.getService(materials[i].ServiceID)
	}
	return materials, nil
}

func (r *TrainingRepository) SetAssignments(employeeID string, serviceIDs []string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM "EmployeeTrainingAssignment" WHERE "employeeId" = $1`, employeeID); err != nil {
		return err
	}
	for _, sid := range serviceIDs {
		if _, err := tx.Exec(`
			INSERT INTO "EmployeeTrainingAssignment" (id, "employeeId", "serviceId")
			VALUES (gen_random_uuid()::text, $1, $2)
		`, employeeID, sid); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *TrainingRepository) ListMaterials(serviceID string) ([]model.TrainingMaterial, error) {
	var materials []model.TrainingMaterial
	var err error
	if serviceID != "" {
		err = r.db.Select(&materials, `SELECT * FROM "TrainingMaterial" WHERE "serviceId" = $1 ORDER BY "serviceId" ASC, "order" ASC`, serviceID)
	} else {
		err = r.db.Select(&materials, `SELECT * FROM "TrainingMaterial" ORDER BY "serviceId" ASC, "order" ASC`)
	}
	if err != nil {
		return nil, err
	}
	for i := range materials {
		materials[i].Service, _ = r.getService(materials[i].ServiceID)
	}
	return materials, nil
}

func (r *TrainingRepository) CreateMaterial(serviceID, title, url, materialType string, order int) (*model.TrainingMaterial, error) {
	var m model.TrainingMaterial
	err := r.db.Get(&m, `
		INSERT INTO "TrainingMaterial" (id, "serviceId", title, url, type, "order")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5)
		RETURNING *
	`, serviceID, title, url, materialType, order)
	if err != nil {
		return nil, err
	}
	m.Service, _ = r.getService(m.ServiceID)
	return &m, nil
}

func (r *TrainingRepository) UpdateMaterial(id string, title, url, materialType *string, order *int) (*model.TrainingMaterial, error) {
	var m model.TrainingMaterial
	err := r.db.Get(&m, `
		UPDATE "TrainingMaterial" SET
			title = COALESCE($2, title),
			url = COALESCE($3, url),
			type = COALESCE($4, type),
			"order" = COALESCE($5, "order")
		WHERE id = $1
		RETURNING *
	`, id, title, url, materialType, order)
	if err != nil {
		return nil, err
	}
	m.Service, _ = r.getService(m.ServiceID)
	return &m, nil
}

func (r *TrainingRepository) DeleteMaterial(id string) error {
	_, err := r.db.Exec(`DELETE FROM "TrainingMaterial" WHERE id = $1`, id)
	return err
}
