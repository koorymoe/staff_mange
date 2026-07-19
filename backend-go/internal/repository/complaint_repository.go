package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type ComplaintRepository struct {
	db *sqlx.DB
}

func NewComplaintRepository(db *sqlx.DB) *ComplaintRepository {
	return &ComplaintRepository{db: db}
}

func (r *ComplaintRepository) loadEmployeeBrief(id string) *model.EmployeeBrief {
	var brief model.EmployeeBrief
	if err := r.db.Get(&brief, `SELECT id, name FROM "Employee" WHERE id = $1`, id); err != nil {
		return nil
	}
	return &brief
}

func (r *ComplaintRepository) hydrate(c *model.Complaint) {
	var customer model.Customer
	if err := r.db.Get(&customer, `SELECT * FROM "Customer" WHERE id = $1`, c.CustomerID); err == nil {
		c.Customer = &customer
	}
	if c.BookingID != nil {
		var booking model.Booking
		if err := r.db.Get(&booking, `SELECT * FROM "Booking" WHERE id = $1`, *c.BookingID); err == nil {
			c.Booking = &booking
		}
	}
	c.CreatedByEmployee = r.loadEmployeeBrief(c.CreatedByEmployeeID)
	if c.AssignedToEmployeeID != nil {
		c.AssignedToEmployee = r.loadEmployeeBrief(*c.AssignedToEmployeeID)
	}
}

func (r *ComplaintRepository) List() ([]model.Complaint, error) {
	complaints := []model.Complaint{}
	if err := r.db.Select(&complaints, `SELECT * FROM "Complaint" ORDER BY "createdAt" DESC`); err != nil {
		return nil, err
	}
	for i := range complaints {
		r.hydrate(&complaints[i])
	}
	return complaints, nil
}

func (r *ComplaintRepository) Create(customerID string, bookingID *string, description, createdByEmployeeID string) (*model.Complaint, error) {
	var c model.Complaint
	err := r.db.Get(&c, `
		INSERT INTO "Complaint" (id, "customerId", "bookingId", description, "createdByEmployeeId")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4)
		RETURNING *
	`, customerID, bookingID, description, createdByEmployeeID)
	if err != nil {
		return nil, err
	}
	r.hydrate(&c)
	return &c, nil
}

func (r *ComplaintRepository) Update(id string, status, assignedToEmployeeID, resolution *string) (*model.Complaint, error) {
	var c model.Complaint
	err := r.db.Get(&c, `
		UPDATE "Complaint" SET
			status = COALESCE($2, status),
			"assignedToEmployeeId" = COALESCE($3, "assignedToEmployeeId"),
			resolution = COALESCE($4, resolution)
		WHERE id = $1
		RETURNING *
	`, id, status, assignedToEmployeeID, resolution)
	if err != nil {
		return nil, err
	}
	r.hydrate(&c)
	return &c, nil
}

// StatsByCustomer يرجع عدد الشكاوى لكل زبون اشتكى (مرة واحدة على الأقل)، مرتبة من
// الأكثر شكاوى — تقرير منفصل تماماً عن إحصائيات الحجوزات.
func (r *ComplaintRepository) StatsByCustomer() ([]model.ComplaintCustomerStat, error) {
	stats := []model.ComplaintCustomerStat{}
	err := r.db.Select(&stats, `
		SELECT
			cu.id AS "customerId",
			cu.name AS "customerName",
			cu.phone AS "customerPhone",
			COUNT(c.id) AS "complaintCount",
			COUNT(*) FILTER (WHERE c.status IN ('NEW', 'IN_PROGRESS')) AS "openCount"
		FROM "Complaint" c
		JOIN "Customer" cu ON cu.id = c."customerId"
		GROUP BY cu.id, cu.name, cu.phone
		ORDER BY "complaintCount" DESC
	`)
	return stats, err
}

func (r *ComplaintRepository) Resolve(id string, resolution *string) (*model.Complaint, error) {
	var c model.Complaint
	err := r.db.Get(&c, `
		UPDATE "Complaint" SET status = 'RESOLVED', resolution = COALESCE($2, resolution), "resolvedAt" = now()
		WHERE id = $1
		RETURNING *
	`, id, resolution)
	if err != nil {
		return nil, err
	}
	r.hydrate(&c)
	return &c, nil
}
