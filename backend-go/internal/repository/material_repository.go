package repository

import (
	"database/sql"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// MaterialRepository يخزّن ويرجّع أرشيف "مواد الشد" — بحث بالكود يُشابه VLOOKUP
// بالشيت القديم.
type MaterialRepository struct {
	db *sqlx.DB
}

func NewMaterialRepository(db *sqlx.DB) *MaterialRepository {
	return &MaterialRepository{db: db}
}

// List يرجّع المواد، مصفّاة بالكود لو تم تمريره (مطابقة تامة أو بادئة).
func (r *MaterialRepository) List(code string) ([]model.Material, error) {
	rows := []model.Material{}
	if code == "" {
		err := r.db.Select(&rows, `SELECT * FROM "Material" ORDER BY name ASC`)
		return rows, err
	}
	err := r.db.Select(&rows, `SELECT * FROM "Material" WHERE code ILIKE $1 ORDER BY name ASC`, code+"%")
	return rows, err
}

// GetByCode يبحث عن مادة واحدة بالكود المطابق تماماً — نفس منطق VLOOKUP بالشيت.
func (r *MaterialRepository) GetByCode(code string) (*model.Material, error) {
	var m model.Material
	err := r.db.Get(&m, `SELECT * FROM "Material" WHERE code = $1`, code)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// GetByID يبحث عن مادة واحدة بالمعرف.
func (r *MaterialRepository) GetByID(id string) (*model.Material, error) {
	var m model.Material
	err := r.db.Get(&m, `SELECT * FROM "Material" WHERE id = $1`, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}
