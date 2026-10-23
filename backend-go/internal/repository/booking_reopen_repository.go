package repository

import (
	"database/sql"
	"errors"
	"staffmange-api/internal/timeutil"
	"time"

	"github.com/jmoiron/sqlx"
)

// إرجاع حجز «منجز» بالغلط للكادر — حتى يسجّل إنجاز جزئي.
type BookingReopenRepository struct{ db *sqlx.DB }

func NewBookingReopenRepository(db *sqlx.DB) *BookingReopenRepository {
	return &BookingReopenRepository{db: db}
}

var ErrNotCompleted = errors.New("الحجز مو منجز — الإرجاع بس للحجز الي انضغط عليه «تم الإنجاز»")

type ReopenResult struct {
	Code       string
	HasInvoice bool
	CrewIDs    []string
}

func (r *BookingReopenRepository) Reopen(bookingID, byName, reason string) (*ReopenResult, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var b struct {
		Code   string `db:"code"`
		Status string `db:"status"`
	}
	if err := tx.Get(&b, `SELECT code, status::text AS status FROM "Booking" WHERE id = $1 FOR UPDATE`, bookingID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("الحجز مو موجود")
		}
		return nil, err
	}
	if b.Status != "COMPLETED" {
		return nil, ErrNotCompleted
	}
	note := "↩️ رجع للكادر " + time.Now().In(timeutil.CompanyLocation()).Format("2006-01-02 15:04") + " — بواسطة " + byName + ": " + reason
	if _, err := tx.Exec(`UPDATE "Booking" SET status = 'IN_PROGRESS', "completedAt" = NULL, "updatedAt" = now(),
		"adminNotes" = CASE WHEN COALESCE(btrim("adminNotes"), '') = '' THEN $2 ELSE "adminNotes" || E'\n' || $2 END
		WHERE id = $1`, bookingID, note); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(`UPDATE "Mission" SET "completedAt" = NULL WHERE id = (
		SELECT id FROM "Mission" WHERE "bookingId" = $1 ORDER BY "assignedAt" DESC NULLS LAST LIMIT 1)`, bookingID); err != nil {
		return nil, err
	}
	res := &ReopenResult{Code: b.Code, CrewIDs: []string{}}
	_ = tx.Get(&res.HasInvoice, `SELECT EXISTS (SELECT 1 FROM "LeaderInvoice" WHERE "bookingId" = $1 AND "revokedAt" IS NULL)`, bookingID)
	_ = tx.Select(&res.CrewIDs, `SELECT DISTINCT x FROM (
		SELECT "employeeId" AS x FROM "BookingAssignment" WHERE "bookingId" = $1
		UNION SELECT "leaderId" FROM "Mission" WHERE "bookingId" = $1 AND "leaderId" IS NOT NULL) t WHERE x IS NOT NULL`, bookingID)
	return res, tx.Commit()
}

// ═══ التراجع عن «بدء العمل» — طلب (ع) 10-19 ═══
// «موظف مرحلين إله حجز باچر وسوّه بالغلط بدأ العمل اليوم». الحجز يرجع
// «مثبّت» بموعده، وأوقات البدء والانطلاق والوصول تنمسح حتى ما تحسب
// عليه تأخير ولا تطلع بتقرير اليوم. المنجز ما يتراجع (عنده «إرجاع»).
var ErrNotStarted = errors.New("الحجز مو «قيد العمل» — ماكو بدء عمل نتراجع عنه")

func (r *BookingReopenRepository) Unstart(bookingID, byName, reason string) (*ReopenResult, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var b struct {
		Code   string `db:"code"`
		Status string `db:"status"`
	}
	if err := tx.Get(&b, `SELECT code, status::text AS status FROM "Booking" WHERE id = $1 FOR UPDATE`, bookingID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("الحجز مو موجود")
		}
		return nil, err
	}
	if b.Status != "IN_PROGRESS" {
		return nil, ErrNotStarted
	}
	note := "⏪ تراجع عن بدء العمل " + time.Now().In(timeutil.CompanyLocation()).Format("2006-01-02 15:04") + " — بواسطة " + byName + ": " + reason
	if _, err := tx.Exec(`UPDATE "Booking" SET status = 'CONFIRMED', "startedAt" = NULL, "departedAt" = NULL,
		"departOrigin" = NULL, "arrivedAt" = NULL, "updatedAt" = now(),
		"adminNotes" = CASE WHEN COALESCE(btrim("adminNotes"), '') = '' THEN $2 ELSE "adminNotes" || E'\n' || $2 END
		WHERE id = $1`, bookingID, note); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(`UPDATE "Mission" SET "departedAt" = NULL, "arrivedAt" = NULL, "workStartedAt" = NULL
		WHERE "bookingId" = $1 AND "completedAt" IS NULL`, bookingID); err != nil {
		return nil, err
	}
	// (ع) 10-19: أثر البدء بسجل الموظف (بدأ/انطلق/وصل) ينشال هم — وإلا يبقى
	// بسجله وينعدّ بواجب «سجّل مراحل حجوزاتك». الاستلام والمواد يبقون: صحيحة.
	if _, err := tx.Exec(`DELETE FROM "BookingStageProof" WHERE "bookingId" = $1
		AND stage IN ('STARTED', 'DEPARTED', 'ARRIVED')`, bookingID); err != nil {
		return nil, err
	}
	res := &ReopenResult{Code: b.Code, CrewIDs: []string{}}
	_ = tx.Select(&res.CrewIDs, `SELECT DISTINCT x FROM (
		SELECT "employeeId" AS x FROM "BookingAssignment" WHERE "bookingId" = $1
		UNION SELECT "leaderId" FROM "Mission" WHERE "bookingId" = $1 AND "leaderId" IS NOT NULL) t WHERE x IS NOT NULL`, bookingID)
	return res, tx.Commit()
}
