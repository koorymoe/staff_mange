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

func (r *ProjectRepository) Create(code, name string, rep, phone, location *string, mapLatitude, mapLongitude *float64, workType, refPerson *string, priority string, deliveryDate, bookingID, responsibleEmployeeID, surveyorEmployeeID *string) (*model.Project, error) {
	var p model.Project
	err := r.db.Get(&p, `
		INSERT INTO "Project" (id, code, name, rep, phone, location, "mapLatitude", "mapLongitude", "workType", "refPerson", priority, "deliveryDate", stage, "bookingId", "responsibleEmployeeId", "surveyorEmployeeId", "updatedAt")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, now())
		RETURNING *
	`, code, name, rep, phone, location, mapLatitude, mapLongitude, workType, refPerson, priority, deliveryDate, firstStage, bookingID, responsibleEmployeeID, surveyorEmployeeID)
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
			"mapLatitude" = COALESCE($6, "mapLatitude"),
			"mapLongitude" = COALESCE($7, "mapLongitude"),
			"workType" = COALESCE($8, "workType"),
			"refPerson" = COALESCE($9, "refPerson"),
			stage = COALESCE($10, stage),
			price = COALESCE($11, price),
			staff = COALESCE($12, staff),
			time = COALESCE($13, time),
			task = COALESCE($14, task),
			priority = COALESCE($15, priority),
			"deliveryDate" = COALESCE($16, "deliveryDate"),
			survey = COALESCE($17::jsonb, survey),
			"contractPdfBase64" = COALESCE($18, "contractPdfBase64"),
			"signedContractPdfBase64" = COALESCE($19, "signedContractPdfBase64"),
			"responsibleEmployeeId" = COALESCE($20, "responsibleEmployeeId"),
			"surveyorEmployeeId" = COALESCE($21, "surveyorEmployeeId"),
			"updatedAt" = now()
		WHERE id = $1
		RETURNING *
	`, id, req.Name, req.Rep, req.Phone, req.Location, req.MapLatitude, req.MapLongitude, req.WorkType, req.RefPerson, req.Stage,
		req.Price, req.Staff, req.Time, req.Task, req.Priority, req.DeliveryDate, req.Survey,
		req.ContractPdfBase64, req.SignedContractPdfBase64, req.ResponsibleEmployeeID, req.SurveyorEmployeeID)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProjectRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM "Project" WHERE id = $1`, id)
	return err
}
