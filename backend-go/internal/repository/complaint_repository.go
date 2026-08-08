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
	if c.RelatedEmployeeID != nil {
		c.RelatedEmployee = r.loadEmployeeBrief(*c.RelatedEmployeeID)
	}
	// منو اتصل بالزبون — لازم يبين كدام الشكوى مثل ما انطلب
	if c.ContactedByID != nil {
		if e := r.loadEmployeeBrief(*c.ContactedByID); e != nil {
			name := e.Name
			c.ContactedByName = &name
		}
	}
}

// SetContacted يأشر إن أحد اتصل بالزبون (أو يشيل التأشير)، ويخزن منو
// اتصل ومتى. أي موظف يقدر — المهم نعرف منو.
func (r *ComplaintRepository) SetContacted(id string, contacted bool, byID string) (*model.Complaint, error) {
	var q string
	var args []any
	if contacted {
		q = `UPDATE "Complaint" SET "contactedAt" = now(), "contactedById" = $2 WHERE id = $1`
		args = []any{id, byID}
	} else {
		q = `UPDATE "Complaint" SET "contactedAt" = NULL, "contactedById" = NULL WHERE id = $1`
		args = []any{id}
	}
	if _, err := r.db.Exec(q, args...); err != nil {
		return nil, err
	}
	return r.Find(id)
}

// SetNotes ملاحظات الزبون — نستفاد منها بالتحسين.
func (r *ComplaintRepository) SetNotes(id, notes string) (*model.Complaint, error) {
	if _, err := r.db.Exec(`UPDATE "Complaint" SET notes = NULLIF($2,'') WHERE id = $1`, id, notes); err != nil {
		return nil, err
	}
	return r.Find(id)
}

func (r *ComplaintRepository) Find(id string) (*model.Complaint, error) {
	var c model.Complaint
	if err := r.db.Get(&c, `SELECT * FROM "Complaint" WHERE id = $1`, id); err != nil {
		return nil, err
	}
	r.hydrate(&c)
	return &c, nil
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

func (r *ComplaintRepository) Create(customerID string, bookingID *string, complaintType, description, createdByEmployeeID string, relatedEmployeeID *string) (*model.Complaint, error) {
	var c model.Complaint
	err := r.db.Get(&c, `
		INSERT INTO "Complaint" (id, "customerId", "bookingId", type, description, "createdByEmployeeId", "relatedEmployeeId")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $6)
		RETURNING *
	`, customerID, bookingID, complaintType, description, createdByEmployeeID, relatedEmployeeID)
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
			COUNT(*) FILTER (WHERE c.status IN ('NEW', 'IN_PROGRESS')) AS "openCount",
		       COUNT(*) FILTER (WHERE c."contactedAt" IS NULL) AS "notContactedCount"
		FROM "Complaint" c
		JOIN "Customer" cu ON cu.id = c."customerId"
		GROUP BY cu.id, cu.name, cu.phone
		ORDER BY "complaintCount" DESC
	`)
	return stats, err
}

// CountForEmployeeMonth يرجّع عدد الشكاوى المرتبطة بموظف معيّن (relatedEmployeeId)
// خلال شهر معيّن (monthPrefix بصيغة "YYYY-MM").
func (r *ComplaintRepository) CountForEmployeeMonth(employeeID, monthPrefix string) (int, error) {
	var count int
	err := r.db.Get(&count, `
		SELECT COUNT(*) FROM "Complaint"
		WHERE "relatedEmployeeId" = $1 AND to_char("createdAt", 'YYYY-MM') = $2
	`, employeeID, monthPrefix)
	return count, err
}

// CountForEmployeeRange نفس CountForEmployeeMonth لكن لمدى تاريخ حر (from/to
// بصيغة "YYYY-MM-DD").
func (r *ComplaintRepository) CountForEmployeeRange(employeeID, from, to string) (int, error) {
	var count int
	err := r.db.Get(&count, `
		SELECT COUNT(*) FROM "Complaint"
		WHERE "relatedEmployeeId" = $1 AND baghdad_date("createdAt") BETWEEN $2::date AND $3::date
	`, employeeID, from, to)
	return count, err
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
