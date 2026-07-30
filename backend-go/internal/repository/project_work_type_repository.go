package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type ProjectWorkTypeRepository struct {
	db *sqlx.DB
}

func NewProjectWorkTypeRepository(db *sqlx.DB) *ProjectWorkTypeRepository {
	return &ProjectWorkTypeRepository{db: db}
}

func (r *ProjectWorkTypeRepository) List() ([]model.ProjectWorkType, error) {
	items := []model.ProjectWorkType{}
	err := r.db.Select(&items, `SELECT * FROM "ProjectWorkType" ORDER BY "createdAt" ASC`)
	return items, err
}

func (r *ProjectWorkTypeRepository) Create(id, name string) (*model.ProjectWorkType, error) {
	var item model.ProjectWorkType
	err := r.db.Get(&item, `
		INSERT INTO "ProjectWorkType" (id, name) VALUES ($1, $2)
		RETURNING *
	`, id, name)
	return &item, err
}

func (r *ProjectWorkTypeRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM "ProjectWorkType" WHERE id = $1`, id)
	return err
}
