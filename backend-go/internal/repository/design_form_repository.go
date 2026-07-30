package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"staffmange-api/internal/model"
)

type DesignFormRepository struct {
	db *sqlx.DB
}

func NewDesignFormRepository(db *sqlx.DB) *DesignFormRepository {
	return &DesignFormRepository{db: db}
}

func (r *DesignFormRepository) List() ([]model.DesignFormQuestion, error) {
	items := []model.DesignFormQuestion{}
	err := r.db.Select(&items, `SELECT * FROM "DesignFormQuestion" ORDER BY "order" ASC, "createdAt" ASC`)
	return items, err
}

func (r *DesignFormRepository) Create(id, label, qType string, options []string, required bool, order int) (*model.DesignFormQuestion, error) {
	if options == nil {
		options = []string{}
	}
	var q model.DesignFormQuestion
	err := r.db.Get(&q, `
		INSERT INTO "DesignFormQuestion" (id, label, type, options, required, "order")
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING *
	`, id, label, qType, pq.Array(options), required, order)
	return &q, err
}

func (r *DesignFormRepository) NextOrder() (int, error) {
	var max int
	err := r.db.Get(&max, `SELECT COALESCE(MAX("order"), -1) + 1 FROM "DesignFormQuestion"`)
	return max, err
}

func (r *DesignFormRepository) Update(id string, label, qType *string, options []string, required *bool) (*model.DesignFormQuestion, error) {
	var q model.DesignFormQuestion
	err := r.db.Get(&q, `
		UPDATE "DesignFormQuestion" SET
			label = COALESCE($2, label),
			type = COALESCE($3, type),
			options = COALESCE($4, options),
			required = COALESCE($5, required)
		WHERE id = $1
		RETURNING *
	`, id, label, qType, pq.Array(options), required)
	return &q, err
}

func (r *DesignFormRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM "DesignFormQuestion" WHERE id = $1`, id)
	return err
}

func (r *DesignFormRepository) Reorder(questionIDs []string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i, id := range questionIDs {
		if _, err := tx.Exec(`UPDATE "DesignFormQuestion" SET "order" = $2 WHERE id = $1`, id, i); err != nil {
			return err
		}
	}
	return tx.Commit()
}
