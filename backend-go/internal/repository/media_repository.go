package repository

import (
	"time"

	"github.com/jmoiron/sqlx"
)

// ═══ الإعلام والعلاقات العامة: المشاريع المحوّلة للتصوير ═══

type MediaRepository struct{ db *sqlx.DB }

func NewMediaRepository(db *sqlx.DB) *MediaRepository { return &MediaRepository{db: db} }

type MediaBrief struct {
	ID            string     `db:"id" json:"id"`
	ProjectID     string     `db:"projectId" json:"projectId"`
	ProjectCode   string     `db:"projectCode" json:"projectCode"`
	ProjectName   string     `db:"projectName" json:"projectName"`
	WorkType      *string    `db:"workType" json:"workType"`
	Stage         string     `db:"stage" json:"stage"`
	Location      *string    `db:"location" json:"location"`
	LocationURL   *string    `db:"locationUrl" json:"locationUrl"`
	Lat           *float64   `db:"lat" json:"lat"`
	Lng           *float64   `db:"lng" json:"lng"`
	DeliveryDate  *string    `db:"deliveryDate" json:"deliveryDate"`
	ProjectTask   *string    `db:"projectTask" json:"projectTask"`
	EngineerID    *string    `db:"engineerId" json:"engineerId"`
	EngineerName  *string    `db:"engineerName" json:"engineerName"`
	EngineerPhone *string    `db:"engineerPhone" json:"engineerPhone"`
	StartAt       *time.Time `db:"startAt" json:"startAt"`
	ExpectedEndAt *time.Time `db:"expectedEndAt" json:"expectedEndAt"`
	Duration      *string    `db:"duration" json:"duration"`
	Notes         *string    `db:"notes" json:"notes"`
	Status        string     `db:"status" json:"status"`
	ShootAt       *time.Time `db:"shootAt" json:"shootAt"`
	MediaEmpID    *string    `db:"mediaEmployeeId" json:"mediaEmployeeId"`
	MediaEmpName  *string    `db:"mediaEmployeeName" json:"mediaEmployeeName"`
	PublishedURL  *string    `db:"publishedUrl" json:"publishedUrl"`
	MediaNotes    *string    `db:"mediaNotes" json:"mediaNotes"`
	CreatedByName *string    `db:"createdByName" json:"createdByName"`
	CreatedAt     time.Time  `db:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time  `db:"updatedAt" json:"updatedAt"`
}

const mediaSelect = `
	SELECT m.id, m."projectId", p.code AS "projectCode", p.name AS "projectName", p."workType", p.stage,
	       p.location, p."locationUrl", p."mapLatitude" AS lat, p."mapLongitude" AS lng, p."deliveryDate", p.task AS "projectTask",
	       m."engineerId", en.name AS "engineerName", en.phone AS "engineerPhone",
	       m."startAt", m."expectedEndAt", m.duration, m.notes, m.status, m."shootAt",
	       m."mediaEmployeeId", me.name AS "mediaEmployeeName", m."publishedUrl", m."mediaNotes",
	       cb.name AS "createdByName", m."createdAt", m."updatedAt"
	FROM "MediaBrief" m JOIN "Project" p ON p.id = m."projectId"
	LEFT JOIN "Employee" en ON en.id = m."engineerId"
	LEFT JOIN "Employee" me ON me.id = m."mediaEmployeeId"
	LEFT JOIN "Employee" cb ON cb.id = m."createdById"`

func (r *MediaRepository) List() ([]MediaBrief, error) {
	rows := []MediaBrief{}
	err := r.db.Select(&rows, mediaSelect+` ORDER BY CASE m.status WHEN 'NEW' THEN 0 WHEN 'SCHEDULED' THEN 1 WHEN 'SHOT' THEN 2 ELSE 3 END, m."createdAt" DESC LIMIT 300`)
	return rows, err
}

func (r *MediaRepository) Get(id string) (*MediaBrief, error) {
	var b MediaBrief
	if err := r.db.Get(&b, mediaSelect+` WHERE m.id = $1`, id); err != nil {
		return nil, err
	}
	return &b, nil
}

// OpenForProject تحويل مفتوح لنفس المشروع (حتى ما يتكرر).
func (r *MediaRepository) OpenForProject(projectID string) string {
	var id string
	_ = r.db.Get(&id, `SELECT id FROM "MediaBrief" WHERE "projectId" = $1 AND status NOT IN ('PUBLISHED', 'CANCELLED') ORDER BY "createdAt" DESC LIMIT 1`, projectID)
	return id
}

type MediaCreate struct {
	ProjectID     string
	CreatedByID   string
	EngineerID    *string
	StartAt       *time.Time
	ExpectedEndAt *time.Time
	Duration      *string
	Notes         *string
}

func (r *MediaRepository) Create(c MediaCreate) (string, error) {
	var id string
	err := r.db.Get(&id, `INSERT INTO "MediaBrief" ("projectId", "createdById", "engineerId", "startAt", "expectedEndAt", duration, notes)
		VALUES ($1, NULLIF($2, ''), $3, $4, $5, $6, $7) RETURNING id`,
		c.ProjectID, c.CreatedByID, c.EngineerID, c.StartAt, c.ExpectedEndAt, c.Duration, c.Notes)
	return id, err
}

type MediaUpdate struct {
	Status       string
	ShootAt      *time.Time
	MediaEmpID   *string
	PublishedURL *string
	MediaNotes   *string
}

func (r *MediaRepository) Update(id string, u MediaUpdate) error {
	_, err := r.db.Exec(`UPDATE "MediaBrief" SET status = $2, "shootAt" = $3, "mediaEmployeeId" = COALESCE($4, "mediaEmployeeId"),
		"publishedUrl" = $5, "mediaNotes" = $6, "updatedAt" = now() WHERE id = $1`,
		id, u.Status, u.ShootAt, u.MediaEmpID, u.PublishedURL, u.MediaNotes)
	return err
}

// ProjectEngineer المهندس المشرف: المسؤول، أو المتوجّه إله، أو منفّذ الكشف.
func (r *MediaRepository) ProjectEngineer(projectID string) *string {
	var id *string
	_ = r.db.Get(&id, `SELECT COALESCE("responsibleEmployeeId", "delegatedToEmployeeId", "surveyorEmployeeId") FROM "Project" WHERE id = $1`, projectID)
	return id
}

func (r *MediaRepository) SetProjectStage(projectID, stage string) error {
	_, err := r.db.Exec(`UPDATE "Project" SET stage = $2, "updatedAt" = CURRENT_TIMESTAMP WHERE id = $1`, projectID, stage)
	return err
}

// StaleNew تحويلات للإعلام بعدها «جديدة» من +٣ أيام — لعين الطلبات المعلّقة.
func (r *MediaRepository) StaleNew() ([]MediaBrief, error) {
	rows := []MediaBrief{}
	err := r.db.Select(&rows, mediaSelect+` WHERE m.status = 'NEW' AND m."createdAt" < now() - interval '3 days' ORDER BY m."createdAt"`)
	return rows, err
}
