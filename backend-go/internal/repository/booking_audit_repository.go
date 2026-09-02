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
		COALESCE(c.name, '') AS "customerName", e.name AS "raisedByName",
		ld.name AS "leaderName"
	FROM "BookingAuditIssue" i
	JOIN "Booking" b ON b.id = i."bookingId"
	LEFT JOIN "Customer" c ON c.id = b."customerId"
	JOIN "Employee" e ON e.id = i."raisedById"
	LEFT JOIN LATERAL (
		SELECT le.name FROM "LeaderInvoice" li
		JOIN "Employee" le ON le.id = li."employeeId"
		WHERE li."bookingId" = b.id ORDER BY li."createdAt" DESC LIMIT 1
	) ld ON true`

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
//
// raisedByID مو فاضي = بلاغات هذا الموظف بس. استعماله الوحيد المحاسب:
// شغله يختلف عن شغل المراقب — البلاغات عنده **صادر** سجّله هو، وعند
// المراقب **وارد** يدقق بيه على الليدر.
func (r *BookingAuditRepository) List(status string, kinds []string, raisedByID string) ([]model.BookingAuditIssue, error) {
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
	if raisedByID != "" {
		args = append(args, raisedByID)
		q += fmt.Sprintf(` AND i."raisedById" = $%d`, len(args))
	}
	q += ` ORDER BY (i.status = 'OPEN') DESC, i."createdAt" DESC LIMIT 500`
	err := r.db.Select(&rows, q, args...)
	return decorateAuditIssues(rows), err
}

// Resolve يغلق البلاغ **مع** أثره: منو وليش وشنو سوّى.
//
// ⚠️ چان يكتب `status` و`resolvedAt` بس — فالبلاغ يختفي وماكو
// طريقة تعرف منو سكّره ولا ليش.
func (r *BookingAuditRepository) Resolve(
	id, byID, byName, action, reason string,
) error {
	_, err := r.db.Exec(`
		UPDATE "BookingAuditIssue"
		SET status = 'RESOLVED', "resolvedAt" = now(),
		    "resolvedById" = $2, "resolvedByName" = NULLIF($3,''),
		    "actionKind" = $4, "resolveReason" = NULLIF($5,'')
		WHERE id = $1 AND status = 'OPEN'`, id, byID, byName, action, reason)
	return err
}

// LeaderIDForBooking ليدر الحجز — من فاتورته، نفس منطق اسم الليدر
// بالاستعلام فوگ. البلاغ يتابع **على الليدر** مو على المحاسب.
//
// ⚠️ يرجّع فاضياً بلا خطأ لمن ماكو فاتورة: حجز بلا فاتورة ليدر
// ماكو عليه منو ينعاقب، والمستدعي يقرر شنو يسوي.
func (r *BookingAuditRepository) LeaderIDForBooking(bookingID string) (string, error) {
	var id string
	err := r.db.Get(&id, `
		SELECT "employeeId" FROM "LeaderInvoice"
		WHERE "bookingId" = $1 ORDER BY "createdAt" DESC LIMIT 1`, bookingID)
	if err != nil {
		return "", nil
	}
	return id, nil
}

// Find بلاغ واحد — نحتاجه حتى نعرف الليدر والحجز قبل الإغلاق.
func (r *BookingAuditRepository) Find(id string) (*model.BookingAuditIssue, error) {
	var issue model.BookingAuditIssue
	if err := r.db.Get(&issue, auditIssueSelect+` WHERE i.id = $1`, id); err != nil {
		return nil, err
	}
	return &issue, nil
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
