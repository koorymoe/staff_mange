package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type CartRepository struct {
	db *sqlx.DB
}

func NewCartRepository(db *sqlx.DB) *CartRepository {
	return &CartRepository{db: db}
}

func (r *CartRepository) ListForBooking(bookingID string) ([]model.CartItem, error) {
	var items []model.CartItem
	err := r.db.Select(&items, `SELECT * FROM "CartItem" WHERE "bookingId" = $1 ORDER BY "createdAt" ASC`, bookingID)
	return items, err
}

func (r *CartRepository) Create(bookingID, productName string, quantity, unitPrice float64, notes *string) (*model.CartItem, error) {
	var item model.CartItem
	err := r.db.Get(&item, `
		INSERT INTO "CartItem" (id, "bookingId", "productName", quantity, "unitPrice", "totalPrice", notes)
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $6)
		RETURNING *
	`, bookingID, productName, quantity, unitPrice, quantity*unitPrice, notes)
	return &item, err
}

func (r *CartRepository) Update(id string, req model.UpdateCartItemRequest) (*model.CartItem, error) {
	var totalPrice *float64
	if req.Quantity != nil && req.UnitPrice != nil {
		t := *req.Quantity * *req.UnitPrice
		totalPrice = &t
	}

	var item model.CartItem
	err := r.db.Get(&item, `
		UPDATE "CartItem" SET
			"productName" = COALESCE($2, "productName"),
			quantity = COALESCE($3, quantity),
			"unitPrice" = COALESCE($4, "unitPrice"),
			"totalPrice" = COALESCE($5, "totalPrice"),
			notes = COALESCE($6, notes)
		WHERE id = $1
		RETURNING *
	`, id, req.ProductName, req.Quantity, req.UnitPrice, totalPrice, req.Notes)
	return &item, err
}

func (r *CartRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM "CartItem" WHERE id = $1`, id)
	return err
}
