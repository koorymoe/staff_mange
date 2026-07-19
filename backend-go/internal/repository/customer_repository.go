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
	customers := []model.Customer{}
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

func (r *CustomerRepository) Create(name, phone string, location *string, lat, lng *float64) (*model.Customer, error) {
	var c model.Customer
	err := r.db.Get(&c, `
		INSERT INTO "Customer" (id, name, phone, location, "mapLatitude", "mapLongitude")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5)
		RETURNING *
	`, name, phone, location, lat, lng)
	return &c, err
}

func (r *CustomerRepository) FindByID(id string) (*model.Customer, error) {
	var c model.Customer
	err := r.db.Get(&c, `SELECT * FROM "Customer" WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// CountBookings يرجع عدد الحجوزات السابقة لزبون معيّن — يستخدم لتمييز "زبون جديد" عن
// "زبون قديم" وقت تسجيل حجز جديد له.
func (r *CustomerRepository) CountBookings(customerID string) (int, error) {
	var count int
	err := r.db.Get(&count, `SELECT COUNT(*) FROM "Booking" WHERE "customerId" = $1`, customerID)
	return count, err
}

// UpdateLocation يحدّث موقع الزبون المحفوظ كلما ثبّت موظف المبيعات موقعاً جديداً بحجز،
// حتى تنترحل آخر نقطة معروفة تلقائياً بالمرة الجاية.
func (r *CustomerRepository) UpdateLocation(id string, location *string, lat, lng *float64) error {
	_, err := r.db.Exec(`
		UPDATE "Customer" SET
			location = COALESCE($2, location),
			"mapLatitude" = COALESCE($3, "mapLatitude"),
			"mapLongitude" = COALESCE($4, "mapLongitude")
		WHERE id = $1
	`, id, location, lat, lng)
	return err
}
