package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type VipCustomerRepository struct {
	db *sqlx.DB
}

func NewVipCustomerRepository(db *sqlx.DB) *VipCustomerRepository {
	return &VipCustomerRepository{db: db}
}

// vipSelect استعلام واحد بـJOIN يجيب كل تفاصيل العرض (اسم الزبون ورقمه، رمز
// الحجز، واسم الموظف الي علّمه) — بدون استعلام إضافي لكل صف.
const vipSelect = `
	SELECT v.*, c.name AS "customerName", c.phone AS "customerPhone",
		b.code AS "bookingCode", e.name AS "markedByName"
	FROM "VipCustomer" v
	JOIN "Customer" c ON c.id = v."customerId"
	LEFT JOIN "Booking" b ON b.id = v."bookingId"
	JOIN "Employee" e ON e.id = v."markedByEmployeeId"
`

func (r *VipCustomerRepository) List() ([]model.VipCustomer, error) {
	items := []model.VipCustomer{}
	err := r.db.Select(&items, vipSelect+` ORDER BY v."createdAt" DESC`)
	return items, err
}

// ListCustomerIDs يرجّع معرّفات الزبائن المعلّمين — تستخدمها الواجهة لتبيّن
// أي زبون معلّم أصلاً (بدون كشف تفاصيل الـVIP لغير المدير).
func (r *VipCustomerRepository) ListCustomerIDs() ([]string, error) {
	ids := []string{}
	err := r.db.Select(&ids, `SELECT "customerId" FROM "VipCustomer"`)
	return ids, err
}

// Mark يعلّم الزبون كشخصية مهمة — لو معلّم أصلاً يحدّث السجل الموجود بدل ما يفشل.
func (r *VipCustomerRepository) Mark(id string, req model.MarkVipCustomerRequest, markedBy string) (*model.VipCustomer, error) {
	var saved model.VipCustomer
	if err := r.db.Get(&saved, `
		INSERT INTO "VipCustomer" (id, "customerId", "bookingId", "requestSummary", note, "markedByEmployeeId")
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT ("customerId") DO UPDATE SET
			"bookingId" = COALESCE(EXCLUDED."bookingId", "VipCustomer"."bookingId"),
			"requestSummary" = COALESCE(EXCLUDED."requestSummary", "VipCustomer"."requestSummary"),
			note = COALESCE(EXCLUDED.note, "VipCustomer".note),
			"markedByEmployeeId" = EXCLUDED."markedByEmployeeId"
		RETURNING *
	`, id, req.CustomerID, req.BookingID, req.RequestSummary, req.Note, markedBy); err != nil {
		return nil, err
	}
	var full model.VipCustomer
	if err := r.db.Get(&full, vipSelect+` WHERE v.id = $1`, saved.ID); err != nil {
		return nil, err
	}
	return &full, nil
}

func (r *VipCustomerRepository) Unmark(customerID string) error {
	_, err := r.db.Exec(`DELETE FROM "VipCustomer" WHERE "customerId" = $1`, customerID)
	return err
}
