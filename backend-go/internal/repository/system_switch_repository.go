package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type SystemSwitchRepository struct {
	db *sqlx.DB
}

func NewSystemSwitchRepository(db *sqlx.DB) *SystemSwitchRepository {
	return &SystemSwitchRepository{db: db}
}

// All يرجّع حالة **كل** المفاتيح المعروفة — والي ماكو له صف يطلع
// «شغّال».
//
// ⚠️ **ليش نكمّل الناقص هنا مو بالواجهة**: الواجهة الي تفترض
// الافتراضي تتفرّق عن الخادم بأول تعديل. والمصدر الوحيد للحقيقة
// لازم يجاوب سؤال «شنو شغّال الآن؟» كاملاً.
func (r *SystemSwitchRepository) All() (map[string]bool, error) {
	rows := []model.SystemSwitch{}
	if err := r.db.Select(&rows, `SELECT * FROM "SystemSwitch"`); err != nil {
		return nil, err
	}
	out := make(map[string]bool, len(model.SystemSwitchLabels))
	for key := range model.SystemSwitchLabels {
		out[key] = true
	}
	for _, row := range rows {
		if model.KnownSystemSwitch(row.Key) {
			out[row.Key] = row.Enabled
		}
	}
	return out, nil
}

// Set يكتب حالة مفتاح (upsert) ويسجّل منو بدّلها.
func (r *SystemSwitchRepository) Set(key string, enabled bool, byEmployeeID string) error {
	var by *string
	if byEmployeeID != "" {
		by = &byEmployeeID
	}
	_, err := r.db.Exec(`
		INSERT INTO "SystemSwitch" (key, enabled, "updatedAt", "updatedBy")
		VALUES ($1, $2, CURRENT_TIMESTAMP, $3)
		ON CONFLICT (key) DO UPDATE
			SET enabled = EXCLUDED.enabled,
			    "updatedAt" = CURRENT_TIMESTAMP,
			    "updatedBy" = EXCLUDED."updatedBy"`, key, enabled, by)
	return err
}
