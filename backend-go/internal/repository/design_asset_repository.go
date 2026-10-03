package repository

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type DesignAssetRepository struct {
	db *sqlx.DB
}

func NewDesignAssetRepository(db *sqlx.DB) *DesignAssetRepository {
	return &DesignAssetRepository{db: db}
}

// List المعرض. `includeArchived` للي يريد يشوف المؤرشف — الافتراضي
// الحيّ بس.
func (r *DesignAssetRepository) List(includeArchived bool) ([]model.DesignAsset, error) {
	q := `SELECT * FROM "DesignAsset"`
	if !includeArchived {
		q += ` WHERE "archivedAt" IS NULL`
	}
	q += ` ORDER BY "createdAt" DESC`
	rows := []model.DesignAsset{}
	err := r.db.Select(&rows, q)
	return rows, err
}

func (r *DesignAssetRepository) Create(req model.CreateDesignAssetRequest, byID, byName string) (*model.DesignAsset, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, fmt.Errorf("اكتب عنوان التصميم")
	}
	if strings.TrimSpace(req.FileKey) == "" {
		return nil, fmt.Errorf("ارفع ملف التصميم")
	}
	// ⚠️ التصنيف يتحقّق من الخريطة: قيمة برّا القائمة تخلي المعرض
	// يعرض تصنيفاً بلا تسمية، والفلترة ما تلگيه.
	if _, ok := model.DesignCategoryLabels[req.Category]; !ok {
		return nil, fmt.Errorf("اختر تصنيف التصميم")
	}
	row := model.DesignAsset{}
	err := r.db.Get(&row, `
		INSERT INTO "DesignAsset"
			(id, title, category, notes, "fileKey", "fileType", "uploadedById", "uploadedByName")
		VALUES ($1,$2,$3,NULLIF($4,''),$5,NULLIF($6,''),NULLIF($7,''),$8)
		RETURNING *`,
		uuid.NewString(), title, req.Category, derefStr(req.Notes),
		strings.TrimSpace(req.FileKey), derefStr(req.FileType), byID, byName)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// SetArchived أرشفة أو إرجاع.
//
// ⚠️ **ما ينمحي الملف من التخزين**: الأرشفة قرار عرض، والتراجع عنها
// لازم يرجّع التصميم كامل. محو الملف يخلي «رجّعه» يعرض مربعاً فاضياً.
func (r *DesignAssetRepository) SetArchived(id string, archived bool) (*model.DesignAsset, error) {
	row := model.DesignAsset{}
	q := `UPDATE "DesignAsset" SET "archivedAt" = now() WHERE id = $1 RETURNING *`
	if !archived {
		q = `UPDATE "DesignAsset" SET "archivedAt" = NULL WHERE id = $1 RETURNING *`
	}
	if err := r.db.Get(&row, q, id); err != nil {
		return nil, err
	}
	return &row, nil
}
