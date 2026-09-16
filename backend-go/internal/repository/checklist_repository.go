package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"staffmange-api/internal/model"
)

type ChecklistRepository struct {
	db *sqlx.DB
}

func NewChecklistRepository(db *sqlx.DB) *ChecklistRepository {
	return &ChecklistRepository{db: db}
}

func (r *ChecklistRepository) hydrate(items []model.ProjectChecklist) {
	for i := range items {
		c := &items[i]
		if c.ProjectID != nil {
			var p model.Project
			if err := r.db.Get(&p, `SELECT * FROM "Project" WHERE id = $1`, *c.ProjectID); err == nil {
				c.Project = &p
			}
		}
		var e model.Employee
		if err := r.db.Get(&e, `SELECT * FROM "Employee" WHERE id = $1`, c.CreatedByID); err == nil {
			c.CreatedBy = &model.EmployeeBrief{ID: e.ID, Name: e.Name}
		}
	}
}

func (r *ChecklistRepository) List() ([]model.ProjectChecklist, error) {
	items := []model.ProjectChecklist{}
	if err := r.db.Select(&items, `SELECT * FROM "ProjectChecklist" ORDER BY "createdAt" DESC`); err != nil {
		return nil, err
	}
	r.hydrate(items)
	return items, nil
}

func (r *ChecklistRepository) Create(id string, projectID *string, title, createdByID string) (*model.ProjectChecklist, error) {
	var c model.ProjectChecklist
	err := r.db.Get(&c, `
		INSERT INTO "ProjectChecklist" (id, "projectId", title, "createdById")
		VALUES ($1, $2, $3, $4)
		RETURNING *
	`, id, projectID, title, createdByID)
	if err != nil {
		return nil, err
	}
	items := []model.ProjectChecklist{c}
	r.hydrate(items)
	return &items[0], nil
}

func (r *ChecklistRepository) AddPhotos(id string, photoUrls []string) (*model.ProjectChecklist, error) {
	var c model.ProjectChecklist
	err := r.db.Get(&c, `
		UPDATE "ProjectChecklist" SET "photoUrls" = "photoUrls" || $2
		WHERE id = $1
		RETURNING *
	`, id, pq.Array(photoUrls))
	if err != nil {
		return nil, err
	}
	items := []model.ProjectChecklist{c}
	r.hydrate(items)
	return &items[0], nil
}
