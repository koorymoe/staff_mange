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

func (r *DesignFormRepository) ListForms() ([]model.DesignForm, error) {
	forms := []model.DesignForm{}
	err := r.db.Select(&forms, `SELECT * FROM "DesignForm" ORDER BY "createdAt" DESC`)
	return forms, err
}

func (r *DesignFormRepository) GetForm(id string) (*model.DesignForm, error) {
	var f model.DesignForm
	err := r.db.Get(&f, `SELECT * FROM "DesignForm" WHERE id = $1`, id)
	return &f, err
}

func (r *DesignFormRepository) GetFormByToken(token string) (*model.DesignForm, error) {
	var f model.DesignForm
	err := r.db.Get(&f, `SELECT * FROM "DesignForm" WHERE "publicToken" = $1`, token)
	return &f, err
}

func (r *DesignFormRepository) CreateForm(id, name, publicToken string) (*model.DesignForm, error) {
	var f model.DesignForm
	err := r.db.Get(&f, `
		INSERT INTO "DesignForm" (id, name, "publicToken")
		VALUES ($1, $2, $3)
		RETURNING *
	`, id, name, publicToken)
	return &f, err
}

func (r *DesignFormRepository) DeleteForm(id string) error {
	_, err := r.db.Exec(`DELETE FROM "DesignForm" WHERE id = $1`, id)
	return err
}

func (r *DesignFormRepository) List(formID string) ([]model.DesignFormQuestion, error) {
	items := []model.DesignFormQuestion{}
	err := r.db.Select(&items, `SELECT * FROM "DesignFormQuestion" WHERE "formId" = $1 ORDER BY "order" ASC, "createdAt" ASC`, formID)
	return items, err
}

func (r *DesignFormRepository) Create(id, formID, label, qType string, options []string, required bool, order int) (*model.DesignFormQuestion, error) {
	if options == nil {
		options = []string{}
	}
	var q model.DesignFormQuestion
	err := r.db.Get(&q, `
		INSERT INTO "DesignFormQuestion" (id, "formId", label, type, options, required, "order")
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING *
	`, id, formID, label, qType, pq.Array(options), required, order)
	return &q, err
}

func (r *DesignFormRepository) NextOrder(formID string) (int, error) {
	var max int
	err := r.db.Get(&max, `SELECT COALESCE(MAX("order"), -1) + 1 FROM "DesignFormQuestion" WHERE "formId" = $1`, formID)
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

func (r *DesignFormRepository) CreateSubmission(id, formID string, answers []byte) (*model.DesignFormSubmission, error) {
	var s model.DesignFormSubmission
	err := r.db.Get(&s, `
		INSERT INTO "DesignFormSubmission" (id, "formId", answers)
		VALUES ($1, $2, $3::jsonb)
		RETURNING *
	`, id, formID, answers)
	return &s, err
}

func (r *DesignFormRepository) ListSubmissions(formID string) ([]model.DesignFormSubmission, error) {
	items := []model.DesignFormSubmission{}
	err := r.db.Select(&items, `SELECT * FROM "DesignFormSubmission" WHERE "formId" = $1 ORDER BY "submittedAt" DESC`, formID)
	return items, err
}
