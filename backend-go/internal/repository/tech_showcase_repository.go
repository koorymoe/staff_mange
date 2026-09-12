package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"staffmange-api/internal/model"
)

type TechShowcaseRepository struct {
	db *sqlx.DB
}

func NewTechShowcaseRepository(db *sqlx.DB) *TechShowcaseRepository {
	return &TechShowcaseRepository{db: db}
}

func (r *TechShowcaseRepository) hydrate(items []model.TechShowcaseItem) {
	for i := range items {
		it := &items[i]
		var e model.Employee
		if err := r.db.Get(&e, `SELECT * FROM "Employee" WHERE id = $1`, it.EmployeeID); err == nil {
			it.Employee = &model.EmployeeBrief{ID: e.ID, Name: e.Name}
		}
	}
}

func (r *TechShowcaseRepository) List() ([]model.TechShowcaseItem, error) {
	items := []model.TechShowcaseItem{}
	if err := r.db.Select(&items, `SELECT * FROM "TechShowcaseItem" ORDER BY "createdAt" DESC`); err != nil {
		return nil, err
	}
	r.hydrate(items)
	return items, nil
}

func (r *TechShowcaseRepository) Create(id, employeeID, title string, description *string) (*model.TechShowcaseItem, error) {
	var it model.TechShowcaseItem
	err := r.db.Get(&it, `
		INSERT INTO "TechShowcaseItem" (id, "employeeId", title, description)
		VALUES ($1, $2, $3, $4)
		RETURNING *
	`, id, employeeID, title, description)
	if err != nil {
		return nil, err
	}
	items := []model.TechShowcaseItem{it}
	r.hydrate(items)
	return &items[0], nil
}

func (r *TechShowcaseRepository) AddMedia(id string, mediaUrls []string) (*model.TechShowcaseItem, error) {
	var it model.TechShowcaseItem
	err := r.db.Get(&it, `
		UPDATE "TechShowcaseItem" SET "mediaUrls" = "mediaUrls" || $2
		WHERE id = $1
		RETURNING *
	`, id, pq.Array(mediaUrls))
	if err != nil {
		return nil, err
	}
	items := []model.TechShowcaseItem{it}
	r.hydrate(items)
	return &items[0], nil
}
