package repository

import (
	"database/sql"
	"errors"
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
	note := "↩️ رجع للكادر " + time.Now().Format("2006-01-02 15:04") + " — بواسطة " + byName + ": " + reason
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
