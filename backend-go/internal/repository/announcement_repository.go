package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type AnnouncementRepository struct {
	db *sqlx.DB
}

func NewAnnouncementRepository(db *sqlx.DB) *AnnouncementRepository {
	return &AnnouncementRepository{db: db}
}

const announcementSelect = `SELECT a.*, e.name AS "createdByName"
	FROM "Announcement" a JOIN "Employee" e ON e.id = a."createdById"`

// ListActive الإعلانات الشغالة — هذي الي تمر بالشريط.
func (r *AnnouncementRepository) ListActive() ([]model.Announcement, error) {
	rows := []model.Announcement{}
	err := r.db.Select(&rows, announcementSelect+` WHERE a.active ORDER BY a."createdAt" DESC LIMIT 20`)
	return rows, err
}

// ListAll للإدارة — تشمل المخفية حتى يكدر يرجعها.
func (r *AnnouncementRepository) ListAll() ([]model.Announcement, error) {
	rows := []model.Announcement{}
	err := r.db.Select(&rows, announcementSelect+` ORDER BY a."createdAt" DESC LIMIT 200`)
	return rows, err
}

func (r *AnnouncementRepository) Create(body, byID string) (*model.Announcement, error) {
	var id string
	if err := r.db.Get(&id, `
		INSERT INTO "Announcement" (id, body, "createdById")
		VALUES (gen_random_uuid()::text, $1, $2) RETURNING id`, body, byID); err != nil {
		return nil, err
	}
	var a model.Announcement
	err := r.db.Get(&a, announcementSelect+` WHERE a.id = $1`, id)
	return &a, err
}

func (r *AnnouncementRepository) SetActive(id string, active bool) error {
	_, err := r.db.Exec(`UPDATE "Announcement" SET active = $2 WHERE id = $1`, id, active)
	return err
}

func (r *AnnouncementRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM "Announcement" WHERE id = $1`, id)
	return err
}
