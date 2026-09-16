package repository

import (
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// NetworkPriceRepository قائمة أسعار الشبكات — يعدّلها المالك ومدير
// النظام من الشاشة، فما نحتاج نشر لكل تغيير سعر.
type NetworkPriceRepository struct {
	db *sqlx.DB
}

func NewNetworkPriceRepository(db *sqlx.DB) *NetworkPriceRepository {
	return &NetworkPriceRepository{db: db}
}

// List الفقرات — activeOnly للاستمارة، والكل لشاشة الإعدادات.
func (r *NetworkPriceRepository) List(activeOnly bool) ([]model.NetworkPriceItem, error) {
	rows := []model.NetworkPriceItem{}
	q := `SELECT * FROM "NetworkPriceItem"`
	if activeOnly {
		q += ` WHERE active = true`
	}
	q += ` ORDER BY "sortOrder", label`
	if err := r.db.Select(&rows, q); err != nil {
		return nil, err
	}
	for i := range rows {
		decodeBrackets(&rows[i])
	}
	return rows, nil
}

func (r *NetworkPriceRepository) FindByID(id string) (*model.NetworkPriceItem, error) {
	var it model.NetworkPriceItem
	err := r.db.Get(&it, `SELECT * FROM "NetworkPriceItem" WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	decodeBrackets(&it)
	return &it, nil
}

func (r *NetworkPriceRepository) Create(req model.SaveNetworkPriceItemRequest) (*model.NetworkPriceItem, error) {
	var it model.NetworkPriceItem
	err := r.db.Get(&it, `
		INSERT INTO "NetworkPriceItem"
			(id, label, unit, "pricingMode", "basePrice", "includedQty", "extraPerUnit", brackets, note, "sortOrder", active)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8::jsonb,NULLIF($9,''),$10,COALESCE($11,true))
		RETURNING *`,
		uuid.NewString(), req.Label, req.Unit, req.PricingMode, req.BasePrice, req.IncludedQty,
		req.ExtraPerUnit, encodeBrackets(req.Brackets), derefStr(req.Note), req.SortOrder, req.Active)
	if err != nil {
		return nil, err
	}
	decodeBrackets(&it)
	return &it, nil
}

func (r *NetworkPriceRepository) Update(id string, req model.SaveNetworkPriceItemRequest) (*model.NetworkPriceItem, error) {
	var it model.NetworkPriceItem
	err := r.db.Get(&it, `
		UPDATE "NetworkPriceItem"
		SET label = $2, unit = $3, "pricingMode" = $4, "basePrice" = $5, "includedQty" = $6,
			"extraPerUnit" = $7, brackets = $8::jsonb, note = NULLIF($9,''), "sortOrder" = $10,
			active = COALESCE($11, active), "updatedAt" = now()
		WHERE id = $1
		RETURNING *`,
		id, req.Label, req.Unit, req.PricingMode, req.BasePrice, req.IncludedQty,
		req.ExtraPerUnit, encodeBrackets(req.Brackets), derefStr(req.Note), req.SortOrder, req.Active)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("الفقرة مو موجودة")
	}
	if err != nil {
		return nil, err
	}
	decodeBrackets(&it)
	return &it, nil
}

// Deactivate ما نمحي: الفواتير القديمة تشير للفقرة، ومحوها يخلي
// فاتورة منجزة بلا تفسير لمبلغها.
func (r *NetworkPriceRepository) Deactivate(id string) error {
	res, err := r.db.Exec(`UPDATE "NetworkPriceItem" SET active = false, "updatedAt" = now() WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("الفقرة مو موجودة")
	}
	return nil
}

func encodeBrackets(b []model.NetworkBracket) string {
	if len(b) == 0 {
		return "[]"
	}
	raw, err := json.Marshal(b)
	if err != nil {
		return "[]"
	}
	return string(raw)
}

func decodeBrackets(it *model.NetworkPriceItem) {
	it.Brackets = []model.NetworkBracket{}
	if len(it.BracketsJSON) == 0 {
		return
	}
	_ = json.Unmarshal(it.BracketsJSON, &it.Brackets)
}
