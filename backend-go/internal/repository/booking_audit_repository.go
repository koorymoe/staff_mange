package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type BookingAuditRepository struct {
	db *sqlx.DB
}

func NewBookingAuditRepository(db *sqlx.DB) *BookingAuditRepository {
	return &BookingAuditRepository{db: db}
}

const auditIssueSelect = `SELECT i.*, b.code AS "bookingCode",
		COALESCE(c.name, '') AS "customerName", e.name AS "raisedByName"
	FROM "BookingAuditIssue" i
	JOIN "Booking" b ON b.id = i."bookingId"
	LEFT JOIN "Customer" c ON c.id = b."customerId"
	JOIN "Employee" e ON e.id = i."raisedById"`

func decorateAuditIssues(rows []model.BookingAuditIssue) []model.BookingAuditIssue {
	for i := range rows {
		rows[i].KindLabel = model.AuditIssueLabels[rows[i].Kind]
		rows[i].RoutedTo = model.AuditRoutedLabel(rows[i].Kind)
	}
	return rows
}

// Verify يأشر الحجز مدقق — بس إذا عنده مبلغ فعلاً.
//
// الشرط داخل الـUPDATE مو بالكود: لو تحققنا بالكود ثم حدّثنا، ممكن
// موظف ثاني يفضّي المبلغ بينهما. وهذا هو بيت القصيد — «مدقق» بلا
// مبلغ يخرب كل الأرباح والإحصائيات.
func (r *BookingAuditRepository) Verify(bookingID string, amount, advance *float64) error {
	var got string
	err := r.db.Get(&got, `
		UPDATE "Booking" SET
			"amountCollected" = COALESCE($2, "amountCollected"),
			"advancePaid" = COALESCE($3, "advancePaid"),
			"amountVerified" = true
		WHERE id = $1
		  AND status = 'COMPLETED'
		  AND COALESCE($2, "amountCollected", 0) > 0
		RETURNING id`, bookingID, amount, advance)
	if err != nil {
		return fmt.Errorf("ما تكدر تدقق حجز بلا مبلغ — اكتب المبلغ من الفاتورة أو أشّر خطأ")
	}
	return nil
}

// SetAmount يصحّح المبلغ بلا تدقيق — للحجوزات القديمة الي المحاسب
// يمشي عليها وحدة وحدة من فواتير النظام القديم.
func (r *BookingAuditRepository) SetAmount(bookingID string, amount, advance *float64) error {
	_, err := r.db.Exec(`
		UPDATE "Booking" SET
			"amountCollected" = COALESCE($2, "amountCollected"),
			"advancePaid" = COALESCE($3, "advancePaid")
		WHERE id = $1`, bookingID, amount, advance)
	return err
}

// RaiseIssue يسجّل بلاغ خطأ ويترك الحجز *غير* مدقق.
func (r *BookingAuditRepository) RaiseIssue(bookingID, kind string, note *string, expected, actual *float64, byID string) (*model.BookingAuditIssue, error) {
	var id string
	err := r.db.Get(&id, `
		INSERT INTO "BookingAuditIssue" (id, "bookingId", kind, note, "expectedAmount", "actualAmount", "raisedById")
		VALUES (gen_random_uuid()::text, $1, $2, NULLIF($3,''), $4, $5, $6)
		RETURNING id`, bookingID, kind, derefStr(note), expected, actual, byID)
	if err != nil {
		return nil, err
	}
	var row model.BookingAuditIssue
	if err := r.db.Get(&row, auditIssueSelect+` WHERE i.id = $1`, id); err != nil {
		return nil, err
	}
	return &decorateAuditIssues([]model.BookingAuditIssue{row})[0], nil
}

// List بلاغات الأخطاء. kinds فاضية = الكل.
func (r *BookingAuditRepository) List(status string, kinds []string) ([]model.BookingAuditIssue, error) {
	rows := []model.BookingAuditIssue{}
	q := auditIssueSelect + ` WHERE 1=1`
	args := []any{}
	if status != "" {
		args = append(args, status)
		q += fmt.Sprintf(` AND i.status = $%d`, len(args))
	}
	if len(kinds) > 0 {
		args = append(args, kinds)
		q += fmt.Sprintf(` AND i.kind = ANY($%d)`, len(args))
	}
	q += ` ORDER BY (i.status = 'OPEN') DESC, i."createdAt" DESC LIMIT 500`
	err := r.db.Select(&rows, q, args...)
	return decorateAuditIssues(rows), err
}

func (r *BookingAuditRepository) Resolve(id string) error {
	_, err := r.db.Exec(`
		UPDATE "BookingAuditIssue" SET status = 'RESOLVED', "resolvedAt" = now()
		WHERE id = $1 AND status = 'OPEN'`, id)
	return err
}

// KindOf نوع البلاغ — نحتاجه حتى نتأكد إن الي يغلقه من الجهة المعنية.
func (r *BookingAuditRepository) KindOf(id string) (string, error) {
	var kind string
	err := r.db.Get(&kind, `SELECT kind FROM "BookingAuditIssue" WHERE id = $1`, id)
	return kind, err
}

// PendingZeroAmount الحجوزات المنجزة الي مبلغها صفر أو فاضي — هذي
// الي المحاسب لازم يمشي عليها بفواتير النظام القديم.
func (r *BookingAuditRepository) PendingZeroAmount() (int, error) {
	var n int
	err := r.db.Get(&n, `
		SELECT COUNT(*) FROM "Booking"
		WHERE status = 'COMPLETED' AND COALESCE("amountCollected", 0) = 0`)
	return n, err
}
