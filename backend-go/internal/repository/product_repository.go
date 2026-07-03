package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type ProductRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) List() ([]model.Product, error) {
	var products []model.Product
	err := r.db.Select(&products, `SELECT * FROM "Product" ORDER BY name ASC`)
	return products, err
}

func (r *ProductRepository) Create(name string, unit *string, defaultPrice *float64, imageBase64 *string) (*model.Product, error) {
	var p model.Product
	err := r.db.Get(&p, `
		INSERT INTO "Product" (id, name, unit, "defaultPrice", "imageBase64")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4)
		RETURNING *
	`, name, unit, defaultPrice, imageBase64)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepository) Update(id string, name, unit *string, defaultPrice *float64, imageBase64 *string) (*model.Product, error) {
	var p model.Product
	err := r.db.Get(&p, `
		UPDATE "Product" SET
			name = COALESCE($2, name),
			unit = COALESCE($3, unit),
			"defaultPrice" = COALESCE($4, "defaultPrice"),
			"imageBase64" = COALESCE($5, "imageBase64")
		WHERE id = $1
		RETURNING *
	`, id, name, unit, defaultPrice, imageBase64)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM "Product" WHERE id = $1`, id)
	return err
}
