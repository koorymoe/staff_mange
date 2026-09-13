package repository

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// EntityAvatarModel مجسّم ثلاثي الأبعاد مرفوع من داخل النظام.
type EntityAvatarModel struct {
	ID           string     `db:"id" json:"id"`
	Label        string     `db:"label" json:"label"`
	FileKey      string     `db:"fileKey" json:"fileKey"`
	FileType     string     `db:"fileType" json:"fileType"`
	SizeBytes    int64      `db:"sizeBytes" json:"sizeBytes"`
	UploadedByID *string    `db:"uploadedById" json:"uploadedById"`
	ArchivedAt   *time.Time `db:"archivedAt" json:"archivedAt"`
	CreatedAt    time.Time  `db:"createdAt" json:"createdAt"`
}

type EntityModelRepository struct{ db *sqlx.DB }

func NewEntityModelRepository(db *sqlx.DB) *EntityModelRepository {
	return &EntityModelRepository{db: db}
}

// List المجسّمات غير المؤرشفة، الأحدث أول.
func (r *EntityModelRepository) List() ([]EntityAvatarModel, error) {
	rows := []EntityAvatarModel{}
	err := r.db.Select(&rows, `
		SELECT * FROM "EntityAvatarModel"
		WHERE "archivedAt" IS NULL
		ORDER BY "createdAt" DESC`)
	return rows, err
}

// Get صفّ واحد بمعرّفه — نحتاجه قبل الأرشفة وقبل الخدمة.
func (r *EntityModelRepository) Get(id string) (*EntityAvatarModel, error) {
	var row EntityAvatarModel
	if err := r.db.Get(&row, `SELECT * FROM "EntityAvatarModel" WHERE id = $1`, id); err != nil {
		return nil, err
	}
	return &row, nil
}

// Create يسجّل مجسّماً مرفوعاً.
//
// ⚠️ الملف نفسه يكون **مخزوناً** قبل هالنداء، والصف يحمل مفتاحه بس.
func (r *EntityModelRepository) Create(label, fileKey, fileType string, size int64, uploadedBy string) (*EntityAvatarModel, error) {
	label = strings.TrimSpace(label)
	if label == "" {
		return nil, fmt.Errorf("اسم المجسّم مطلوب")
	}
	if strings.TrimSpace(fileKey) == "" {
		return nil, fmt.Errorf("ملف المجسّم مطلوب")
	}
	var row EntityAvatarModel
	err := r.db.Get(&row, `
		INSERT INTO "EntityAvatarModel" (id, label, "fileKey", "fileType", "sizeBytes", "uploadedById")
		VALUES ($1,$2,$3,$4,$5,$6) RETURNING *`,
		uuid.NewString(), label, fileKey, fileType, size, uploadedBy)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// Archive أرشفة ناعمة — ما نحذف ملفاً ممكن يكون معروضاً.
func (r *EntityModelRepository) Archive(id string) error {
	res, err := r.db.Exec(`
		UPDATE "EntityAvatarModel" SET "archivedAt" = CURRENT_TIMESTAMP
		WHERE id = $1 AND "archivedAt" IS NULL`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("المجسّم غير موجود أو مؤرشف أصلاً")
	}
	return nil
}
