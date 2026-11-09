package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"staffmange-api/internal/errlog"
	"staffmange-api/internal/model"
)

type ChecklistRepository struct {
	db *sqlx.DB
}

func NewChecklistRepository(db *sqlx.DB) *ChecklistRepository {
	return &ChecklistRepository{db: db}
}

// hydrate (٧ل + ٧ن، 11-04): چانت تسوي `SELECT *` على المشروع لكل بند — يعني تنزّل
// العقد والعقد الموقّع (PDF base64) لكل سطر وهي تحتاج الاسم بس. هسه بأعمدة القائمة
// (بلا الملفات) وبدفعة وحدة.
func (r *ChecklistRepository) hydrate(items []model.ProjectChecklist) {
	projIDs, empIDs := []string{}, []string{}
	for _, c := range items {
		if c.ProjectID != nil {
			projIDs = append(projIDs, *c.ProjectID)
		}
		empIDs = append(empIDs, c.CreatedByID)
	}
	projBy := map[string]*model.Project{}
	if len(projIDs) > 0 {
		ps := []model.Project{}
		errlog.Err(r.db.Select(&ps, `SELECT `+projectListColumns+` FROM "Project" WHERE id = ANY($1)`, pq.Array(projIDs)))
		for i := range ps {
			projBy[ps[i].ID] = &ps[i]
		}
	}
	briefs := []model.EmployeeBrief{}
	errlog.Err(r.db.Select(&briefs, `SELECT id, name FROM "Employee" WHERE id = ANY($1)`, pq.Array(empIDs)))
	empBy := map[string]model.EmployeeBrief{}
	for _, b := range briefs {
		empBy[b.ID] = b
	}
	for i := range items {
		c := &items[i]
		if c.ProjectID != nil {
			c.Project = projBy[*c.ProjectID]
		}
		if b, ok := empBy[c.CreatedByID]; ok {
			c.CreatedBy = &model.EmployeeBrief{ID: b.ID, Name: b.Name}
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
