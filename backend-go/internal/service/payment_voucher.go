package service

import (
	"errors"
	"math"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/timeutil"
)

// ═══ 🧾 سندات الصرف (ل) 11-25 دازاي ═══
// «إذا صرفت سند — لأي مشروع أو حجز أو مشروع خارجي (مو شرط حجز) — كلفة السند، ووقت الصرف،
//  والمبلغ لمنو انسلّم لو انحط بالكاشير». رقم تسلسلي، وما ينحذف: يتلغى بسبب.

type PaymentVoucher struct {
	ID            string     `db:"id" json:"id"`
	Number        int64      `db:"number" json:"number"`
	TargetKind    string     `db:"targetKind" json:"targetKind"`
	TargetLabel   string     `db:"targetLabel" json:"targetLabel"`
	Purpose       string     `db:"purpose" json:"purpose"`
	Amount        float64    `db:"amount" json:"amount"`
	PaidAt        time.Time  `db:"paidAt" json:"paidAt"`
	RecipientKind string     `db:"recipientKind" json:"recipientKind"`
	Recipient     *string    `db:"recipient" json:"recipient"`
	Note          *string    `db:"note" json:"note"`
	CreatedBy     string     `db:"createdBy" json:"createdBy"`
	CancelledAt   *time.Time `db:"cancelledAt" json:"cancelledAt"`
	CancelReason  *string    `db:"cancelReason" json:"cancelReason"`
}

type VoucherInput struct {
	TargetKind          string  `json:"targetKind"` // PROJECT | BOOKING | EXTERNAL
	Code                string  `json:"code"`       // كود المشروع أو الحجز
	ExternalName        string  `json:"externalName"`
	Purpose             string  `json:"purpose"`
	Amount              float64 `json:"amount"`
	PaidAt              string  `json:"paidAt"` // ISO — فاضي = هسه
	RecipientKind       string  `json:"recipientKind"`
	RecipientEmployeeID string  `json:"recipientEmployeeId"`
	Note                string  `json:"note"`
}

type Vouchers struct{ db *sqlx.DB }

func NewVouchers(db *sqlx.DB) *Vouchers { return &Vouchers{db: db} }

const voucherCols = `SELECT v.id, v.number, v."targetKind",
	CASE v."targetKind" WHEN 'PROJECT' THEN COALESCE(p.code || ' — ', '') || COALESCE(p.name, '')
		WHEN 'BOOKING' THEN COALESCE(b.code, '') || COALESCE(' — ' || c.name, '') ELSE v."externalName" END AS "targetLabel",
	v.purpose, v.amount::float8 AS amount, v."paidAt", v."recipientKind", r.name AS recipient, v.note, e.name AS "createdBy",
	v."cancelledAt", v."cancelReason"
	FROM "PaymentVoucher" v JOIN "Employee" e ON e.id = v."createdById"
	LEFT JOIN "Employee" r ON r.id = v."recipientEmployeeId"
	LEFT JOIN "Project" p ON p.id = v."projectId"
	LEFT JOIN "Booking" b ON b.id = v."bookingId" LEFT JOIN "Customer" c ON c.id = b."customerId"`

func (s *Vouchers) Create(by string, in VoucherInput, now time.Time) (*PaymentVoucher, error) {
	in.Purpose = strings.TrimSpace(in.Purpose)
	if len([]rune(in.Purpose)) < 3 {
		return nil, errors.New("اكتب السند لشنو (مثلاً: شراء كيبل، أجور نقل)")
	}
	if in.Amount <= 0 || math.IsNaN(in.Amount) || in.Amount > 1e11 {
		return nil, errors.New("المبلغ غلط")
	}
	paid := now
	if strings.TrimSpace(in.PaidAt) != "" {
		t, err := time.Parse(time.RFC3339, in.PaidAt)
		if err != nil {
			return nil, errors.New("وقت الصرف غلط")
		}
		if t.After(now.Add(10 * time.Minute)) {
			return nil, errors.New("وقت الصرف ما يصير بالمستقبل")
		}
		paid = t
	}
	var projectID, bookingID, ext *string
	switch in.TargetKind {
	case "PROJECT":
		var id string
		if err := s.db.Get(&id, `SELECT id FROM "Project" WHERE upper(code) = upper(btrim($1))`, in.Code); err != nil {
			return nil, errors.New("المشروع مو موجود — اكتب كوده")
		}
		projectID = &id
	case "BOOKING":
		var id string
		if err := s.db.Get(&id, `SELECT id FROM "Booking" WHERE upper(code) = upper(btrim($1))`, in.Code); err != nil {
			return nil, errors.New("الحجز مو موجود — اكتب كوده")
		}
		bookingID = &id
	case "EXTERNAL":
		n := strings.TrimSpace(in.ExternalName)
		if n == "" {
			return nil, errors.New("اكتب اسم المشروع الخارجي")
		}
		ext = &n
	default:
		return nil, errors.New("اختار السند لمشروع لو حجز لو مشروع خارجي")
	}
	var rec *string
	switch in.RecipientKind {
	case "EMPLOYEE":
		if in.RecipientEmployeeID == "" {
			return nil, errors.New("اختار الموظف الي استلم المبلغ")
		}
		rec = &in.RecipientEmployeeID
	case "CASHIER":
	default:
		return nil, errors.New("المبلغ انسلّم لموظف لو انحط بالكاشير؟")
	}
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.Exec(`SELECT pg_advisory_xact_lock(hashtext('payment-voucher-number'))`); err != nil {
		return nil, err
	}
	var id string
	err = tx.Get(&id, `INSERT INTO "PaymentVoucher" (number, "targetKind", "projectId", "bookingId", "externalName", purpose, amount, "paidAt",
			"recipientKind", "recipientEmployeeId", note, "createdById")
		VALUES ((SELECT COALESCE(max(number), 0) + 1 FROM "PaymentVoucher"), $1, $2, $3, $4, $5, $6, $7, $8, $9, NULLIF(btrim($10), ''), $11)
		RETURNING id`, in.TargetKind, projectID, bookingID, ext, in.Purpose, in.Amount, paid, in.RecipientKind, rec, in.Note, by)
	if err != nil {
		if strings.Contains(err.Error(), "recipientEmployeeId_fkey") {
			return nil, errors.New("الموظف مو موجود")
		}
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	var v PaymentVoucher
	if err := s.db.Get(&v, voucherCols+` WHERE v.id = $1`, id); err != nil {
		return nil, err
	}
	return &v, nil
}

type VoucherList struct {
	Month     string           `json:"month"`
	Items     []PaymentVoucher `json:"items"`
	Total     float64          `json:"total"`
	ToCashier float64          `json:"toCashier"`
	ToStaff   float64          `json:"toStaff"`
}

func (s *Vouchers) List(month string) (*VoucherList, error) {
	if month == "" {
		month = timeutil.ThisPeriod()
	}
	from, to, ok := timeutil.PeriodStart(month)
	if !ok {
		return nil, errors.New("الشهر غلط")
	}
	out := &VoucherList{Month: month, Items: []PaymentVoucher{}}
	if err := s.db.Select(&out.Items, voucherCols+` WHERE v."paidAt" >= $1 AND v."paidAt" < $2 ORDER BY v."paidAt" DESC, v.number DESC`, from, to); err != nil {
		return nil, err
	}
	for _, v := range out.Items {
		if v.CancelledAt != nil {
			continue
		}
		out.Total += v.Amount
		if v.RecipientKind == "CASHIER" {
			out.ToCashier += v.Amount
		} else {
			out.ToStaff += v.Amount
		}
	}
	return out, nil
}

// Cancel يلغي السند بسبب (ما ينحذف — يبقى بالسجل).
func (s *Vouchers) Cancel(by, id, reason string) error {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return errors.New("اكتب سبب الإلغاء")
	}
	res, err := s.db.Exec(`UPDATE "PaymentVoucher" SET "cancelledAt" = now(), "cancelledById" = $2, "cancelReason" = $3 WHERE id = $1 AND "cancelledAt" IS NULL`, id, by, reason)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrDesignHidden
	}
	return nil
}
