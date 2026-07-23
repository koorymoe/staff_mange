package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type KpiCriterionRepository struct {
	db *sqlx.DB
}

func NewKpiCriterionRepository(db *sqlx.DB) *KpiCriterionRepository {
	return &KpiCriterionRepository{db: db}
}

func (r *KpiCriterionRepository) List() ([]model.KpiCriterion, error) {
	criteria := []model.KpiCriterion{}
	err := r.db.Select(&criteria, `SELECT * FROM "KpiCriterion" ORDER BY "createdAt" ASC`)
	return criteria, err
}

func (r *KpiCriterionRepository) Create(label string) (*model.KpiCriterion, error) {
	var c model.KpiCriterion
	err := r.db.Get(&c, `
		INSERT INTO "KpiCriterion" (id, label)
		VALUES (gen_random_uuid()::text, $1)
		RETURNING *
	`, label)
	return &c, err
}

func (r *KpiCriterionRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM "KpiCriterion" WHERE id = $1`, id)
	return err
}
