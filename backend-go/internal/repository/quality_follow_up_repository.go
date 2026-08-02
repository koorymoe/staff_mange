package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type QualityFollowUpRepository struct {
	db *sqlx.DB
}

func NewQualityFollowUpRepository(db *sqlx.DB) *QualityFollowUpRepository {
	return &QualityFollowUpRepository{db: db}
}

func (r *QualityFollowUpRepository) hydrate(q *model.QualityFollowUp) {
	var booking model.Booking
	if err := r.db.Get(&booking, `SELECT * FROM "Booking" WHERE id = $1`, q.BookingID); err == nil {
		q.Booking = &booking
	}
	var customer model.Customer
	if err := r.db.Get(&customer, `SELECT * FROM "Customer" WHERE id = $1`, q.CustomerID); err == nil {
		q.Customer = &customer
	}
	if q.ContactedByEmployeeID != nil {
		var brief model.EmployeeBrief
		if err := r.db.Get(&brief, `SELECT id, name FROM "Employee" WHERE id = $1`, *q.ContactedByEmployeeID); err == nil {
			q.ContactedByEmployee = &brief
		}
	}
	q.Financials = r.loadFinancials(q.BookingID)
}

// loadFinancials يجمع تفاصيل الحجز ومشروعه المرتبط وأمواله باستعلام واحد.
//
// الفارق يُحسب هنا بالسيرفر (مو بالواجهة) حتى يبقى تعريف واحد للفارق بكل
// الشاشات — الفارق = المتفق عليه ناقص (العربون + المستلم).
func (r *QualityFollowUpRepository) loadFinancials(bookingID string) *model.QualityFollowUpFinancials {
	var f model.QualityFollowUpFinancials
	err := r.db.Get(&f, `
		SELECT
			b.code                       AS "bookingCode",
			s.name                       AS "serviceName",
			b.address                    AS "location",
			b.notes                      AS "workDetails",
			p.id                         AS "projectId",
			p.code                       AS "projectCode",
			p.name                       AS "projectName",
			p.stage                      AS "projectStage",
			b."quotedPrice"              AS "quotedPrice",
			p.price                      AS "projectPrice",
			b."advancePaid"              AS "advancePaid",
			b."amountCollected"          AS "amountCollected"
		FROM "Booking" b
		LEFT JOIN "Service" s ON s.id = b."serviceId"
		LEFT JOIN "Project" p ON p."bookingId" = b.id
		WHERE b.id = $1`, bookingID)
	if err != nil {
		return nil
	}
	// السعر المعتمد: سعر المشروع إذا الحجز انرحّل لمشروع، وإلا السعر المتفق بالحجز
	if f.ProjectPrice != nil && *f.ProjectPrice > 0 {
		f.AgreedTotal = *f.ProjectPrice
	} else if f.QuotedPrice != nil {
		f.AgreedTotal = *f.QuotedPrice
	}
	if f.AdvancePaid != nil {
		f.ReceivedTotal += *f.AdvancePaid
	}
	if f.AmountCollected != nil {
		f.ReceivedTotal += *f.AmountCollected
	}
	f.Difference = f.AgreedTotal - f.ReceivedTotal
	return &f
}

// CreateForBooking تنشئ سطر متابعة جودة تلقائياً لحجز اكتمل (idempotent — ما تكرر لو
// تكرر استدعاء Complete على نفس الحجز بالخطأ).
func (r *QualityFollowUpRepository) CreateForBooking(bookingID, customerID string) error {
	_, err := r.db.Exec(`
		INSERT INTO "QualityFollowUp" (id, "bookingId", "customerId")
		VALUES (gen_random_uuid()::text, $1, $2)
		ON CONFLICT ("bookingId") DO NOTHING
	`, bookingID, customerID)
	return err
}

func (r *QualityFollowUpRepository) List() ([]model.QualityFollowUp, error) {
	items := []model.QualityFollowUp{}
	if err := r.db.Select(&items, `SELECT * FROM "QualityFollowUp" ORDER BY "createdAt" DESC`); err != nil {
		return nil, err
	}
	for i := range items {
		r.hydrate(&items[i])
	}
	return items, nil
}

func (r *QualityFollowUpRepository) Update(id, status string, contactedByEmployeeID string, contactNotes *string) (*model.QualityFollowUp, error) {
	var q model.QualityFollowUp
	err := r.db.Get(&q, `
		UPDATE "QualityFollowUp" SET
			status = $2,
			"contactNotes" = COALESCE($3, "contactNotes"),
			"contactedByEmployeeId" = $4,
			"contactedAt" = now()
		WHERE id = $1
		RETURNING *
	`, id, status, contactNotes, contactedByEmployeeID)
	if err != nil {
		return nil, err
	}
	r.hydrate(&q)
	return &q, nil
}
