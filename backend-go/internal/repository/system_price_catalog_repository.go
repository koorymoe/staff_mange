package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// SystemPriceCatalogRepository يخزّن ويرجّع كتالوج أسعار المنظومات الثمانية —
// نفس بيانات شيت "تكاليف المشروع" الي حل محله (ميزة فوترة الليدر).
type SystemPriceCatalogRepository struct {
	db *sqlx.DB
}

func NewSystemPriceCatalogRepository(db *sqlx.DB) *SystemPriceCatalogRepository {
	return &SystemPriceCatalogRepository{db: db}
}

// List يرجّع كل الكتالوج، أو مصفّى حسب اسم منظومة لو تم تمريره.
func (r *SystemPriceCatalogRepository) List(systemName string) ([]model.SystemPriceCatalog, error) {
	rows := []model.SystemPriceCatalog{}
	if systemName == "" {
		err := r.db.Select(&rows, `SELECT * FROM "SystemPriceCatalog" ORDER BY "systemName", category, "itemName"`)
		return rows, err
	}
	err := r.db.Select(&rows, `SELECT * FROM "SystemPriceCatalog" WHERE "systemName" = $1 ORDER BY category, "itemName"`, systemName)
	return rows, err
}

// All يرجّع الكتالوج كامل — يُستخدم من الخدمة لحساب تكاليف التنفيذ.
func (r *SystemPriceCatalogRepository) All() ([]model.SystemPriceCatalog, error) {
	rows := []model.SystemPriceCatalog{}
	err := r.db.Select(&rows, `SELECT * FROM "SystemPriceCatalog"`)
	return rows, err
}
