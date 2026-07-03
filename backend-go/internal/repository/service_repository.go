package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type ServiceRepository struct {
	db *sqlx.DB
}

func NewServiceRepository(db *sqlx.DB) *ServiceRepository {
	return &ServiceRepository{db: db}
}

func (r *ServiceRepository) List() ([]model.Service, error) {
	services := []model.Service{}
	if err := r.db.Select(&services, `SELECT * FROM "Service" ORDER BY name ASC`); err != nil {
		return nil, err
	}

	for i := range services {
		skills := []model.Skill{}
		if err := r.db.Select(&skills, `SELECT * FROM "Skill" WHERE "serviceId" = $1 ORDER BY name ASC`, services[i].ID); err != nil {
			return nil, err
		}
		services[i].Skills = skills
	}

	return services, nil
}

func (r *ServiceRepository) Create(s *model.Service) error {
	_, err := r.db.NamedExec(`
		INSERT INTO "Service" (id, name, category)
		VALUES (:id, :name, :category)
	`, s)
	return err
}
