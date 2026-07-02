package repository

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type CustomerRepository struct {
	db *sqlx.DB
}

func NewCustomerRepository(db *sqlx.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

func (r *CustomerRepository) List() ([]model.Customer, error) {
	var customers []model.Customer
	err := r.db.Select(&customers, `SELECT * FROM "Customer" ORDER BY "customerCode" ASC`)
	return customers, err
}

func (r *CustomerRepository) FindByPhone(phone string) (*model.Customer, error) {
	var c model.Customer
	err := r.db.Get(&c, `SELECT * FROM "Customer" WHERE phone = $1`, phone)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CustomerRepository) Create(name, phone string, location *string) (*model.Customer, error) {
	var c model.Customer
	err := r.db.Get(&c, `
		INSERT INTO "Customer" (id, name, phone, location)
		VALUES (gen_random_uuid()::text, $1, $2, $3)
		RETURNING *
	`, name, phone, location)
	return &c, err
}
