package repository

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"staffmange-api/internal/model"
)

// ═══ جدول أسعار الديكور — (ع) 10-16 ═══
// المدير يحط لكل مهنة بنودها بسعر (متر/قطعة/يوم)، وفاتورة الديكور تنحسب منه.
type DecorRateRepository struct {
	db *sqlx.DB
}

func NewDecorRateRepository(db *sqlx.DB) *DecorRateRepository { return &DecorRateRepository{db: db} }

const decorRateSelect = `SELECT r.*, s.name AS "serviceName" FROM "DecorRate" r JOIN "Service" s ON s.id = r."serviceId"`

func (r *DecorRateRepository) List(activeOnly bool) ([]model.DecorRate, error) {
	rows := []model.DecorRate{}
	q := decorRateSelect
	if activeOnly {
		q += ` WHERE r.active`
	}
	err := r.db.Select(&rows, q+` ORDER BY s.name, r.name`)
	return rows, err
}

func (r *DecorRateRepository) ByIDs(ids []string) (map[string]model.DecorRate, error) {
	rows := []model.DecorRate{}
	if err := r.db.Select(&rows, decorRateSelect+` WHERE r.id = ANY($1) AND r.active`, pq.Array(ids)); err != nil {
		return nil, err
	}
	out := map[string]model.DecorRate{}
	for _, x := range rows {
		out[x.ID] = x
	}
	return out, nil
}

// Save ينشئ بنداً جديداً (id فارغ) أو يعدّل الموجود.
func (r *DecorRateRepository) Save(x model.DecorRate) error {
	if x.ID == "" {
		_, err := r.db.Exec(`INSERT INTO "DecorRate" (id, "serviceId", name, unit, price) VALUES ($1,$2,$3,$4,$5)`,
			uuid.NewString(), x.ServiceID, x.Name, x.Unit, x.Price)
		return err
	}
	_, err := r.db.Exec(`UPDATE "DecorRate" SET "serviceId" = $2, name = $3, unit = $4, price = $5, active = $6, "updatedAt" = now() WHERE id = $1`,
		x.ID, x.ServiceID, x.Name, x.Unit, x.Price, x.Active)
	return err
}
