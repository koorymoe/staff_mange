package repository

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type SurveyPhoto struct {
	ID           string    `db:"id" json:"id"`
	BookingID    *string   `db:"bookingId" json:"bookingId"`
	ProjectID    *string   `db:"projectId" json:"projectId"`
	FileKey      string    `db:"fileKey" json:"-"`
	ContentType  string    `db:"contentType" json:"contentType"`
	FileName     string    `db:"fileName" json:"fileName"`
	UploadedByID string    `db:"uploadedById" json:"uploadedById"`
	UploadedBy   string    `db:"uploadedByName" json:"uploadedByName"`
	CreatedAt    time.Time `db:"createdAt" json:"createdAt"`
}

type SurveyPhotoRepository struct {
	db *sqlx.DB
}

func NewSurveyPhotoRepository(db *sqlx.DB) *SurveyPhotoRepository {
	return &SurveyPhotoRepository{db: db}
}

const surveyPhotoSelect = `
	SELECT p.id, p."bookingId", p."projectId", p."fileKey", p."contentType", p."fileName",
	       p."uploadedById", COALESCE(e.name, '') AS "uploadedByName", p."createdAt"
	FROM "SurveyPhoto" p
	LEFT JOIN "Employee" e ON e.id = p."uploadedById"`

func (r *SurveyPhotoRepository) Create(bookingID, projectID *string, fileKey, contentType, fileName, uploadedByID string) (*SurveyPhoto, error) {
	var id string
	if err := r.db.Get(&id, `
		INSERT INTO "SurveyPhoto" (id, "bookingId", "projectId", "fileKey", "contentType", "fileName", "uploadedById")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $6)
		RETURNING id
	`, bookingID, projectID, fileKey, contentType, fileName, uploadedByID); err != nil {
		return nil, err
	}
	return r.Get(id)
}

func (r *SurveyPhotoRepository) Get(id string) (*SurveyPhoto, error) {
	var p SurveyPhoto
	if err := r.db.Get(&p, surveyPhotoSelect+` WHERE p.id = $1`, id); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *SurveyPhotoRepository) ListForBooking(bookingID string) ([]SurveyPhoto, error) {
	out := []SurveyPhoto{}
	err := r.db.Select(&out, surveyPhotoSelect+` WHERE p."bookingId" = $1 ORDER BY p."createdAt"`, bookingID)
	return out, err
}

func (r *SurveyPhotoRepository) ListForProject(projectID string) ([]SurveyPhoto, error) {
	out := []SurveyPhoto{}
	err := r.db.Select(&out, surveyPhotoSelect+` WHERE p."projectId" = $1 ORDER BY p."createdAt"`, projectID)
	return out, err
}

func (r *SurveyPhotoRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM "SurveyPhoto" WHERE id = $1`, id)
	return err
}

// IsBookingCrew: مكلّف بالحجز أو مشرفه — هذا الي طلع للكشف فعلاً.
func (r *SurveyPhotoRepository) IsBookingCrew(bookingID, employeeID string) (bool, error) {
	var ok bool
	err := r.db.Get(&ok, `
		SELECT EXISTS (SELECT 1 FROM "BookingAssignment" WHERE "bookingId" = $1 AND "employeeId" = $2)
		    OR EXISTS (SELECT 1 FROM "Booking" WHERE id = $1 AND "projectSupervisorId" = $2)
	`, bookingID, employeeID)
	return ok, err
}

// IsProjectSurveyTeam: منفّذ الكشف أو المسؤول أو الي انسلّم إله المشروع.
func (r *SurveyPhotoRepository) IsProjectSurveyTeam(projectID, employeeID string) (bool, error) {
	var ok bool
	err := r.db.Get(&ok, `
		SELECT EXISTS (
			SELECT 1 FROM "Project" WHERE id = $1 AND $2 IN (
				COALESCE("surveyorEmployeeId", ''), COALESCE("responsibleEmployeeId", ''), COALESCE("delegatedToEmployeeId", '')
			)
		)
	`, projectID, employeeID)
	return ok, err
}
