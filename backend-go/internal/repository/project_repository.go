package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

const firstStage = "1. اتصال بالزبون"

type ProjectRepository struct {
	db *sqlx.DB
}

func NewProjectRepository(db *sqlx.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) List() ([]model.Project, error) {
	projects := []model.Project{}
	err := r.db.Select(&projects, `SELECT * FROM "Project" ORDER BY "createdAt" DESC`)
	return projects, err
}

func (r *ProjectRepository) CountAll() (int, error) {
	var count int
	err := r.db.Get(&count, `SELECT COUNT(*) FROM "Project"`)
	return count, err
}

func (r *ProjectRepository) Create(code, name string, rep, phone, location, workType, refPerson *string, priority string, deliveryDate, bookingID *string) (*model.Project, error) {
	var p model.Project
	err := r.db.Get(&p, `
		INSERT INTO "Project" (id, code, name, rep, phone, location, "workType", "refPerson", priority, "deliveryDate", stage, "bookingId", "updatedAt")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, now())
		RETURNING *
	`, code, name, rep, phone, location, workType, refPerson, priority, deliveryDate, firstStage, bookingID)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProjectRepository) Update(id string, req model.UpdateProjectRequest) (*model.Project, error) {
	var p model.Project
	err := r.db.Get(&p, `
		UPDATE "Project" SET
			name = COALESCE($2, name),
			rep = COALESCE($3, rep),
			phone = COALESCE($4, phone),
			location = COALESCE($5, location),
			"workType" = COALESCE($6, "workType"),
			"refPerson" = COALESCE($7, "refPerson"),
			stage = COALESCE($8, stage),
			price = COALESCE($9, price),
			staff = COALESCE($10, staff),
			time = COALESCE($11, time),
			task = COALESCE($12, task),
			priority = COALESCE($13, priority),
			"deliveryDate" = COALESCE($14, "deliveryDate"),
			survey = COALESCE($15::jsonb, survey),
			"sentToGroup" = COALESCE($16, "sentToGroup"),
			"updatedAt" = now()
		WHERE id = $1
		RETURNING *
	`, id, req.Name, req.Rep, req.Phone, req.Location, req.WorkType, req.RefPerson, req.Stage,
		req.Price, req.Staff, req.Time, req.Task, req.Priority, req.DeliveryDate, req.Survey, req.SentToGroup)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProjectRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM "Project" WHERE id = $1`, id)
	return err
}
