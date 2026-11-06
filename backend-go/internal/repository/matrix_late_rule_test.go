package repository

import (
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/database"
)

// TestMatrixLateItemsRule: ماتركس يمشي بنفس قاعدة فحص المتأخرات — يومين فما فوق، والانتظار مو تأخير.
func TestMatrixLateItemsRule(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	cust := uuid.NewString()
	db.MustExec(`INSERT INTO "Customer" (id, name, phone) VALUES ($1, 'زبون فحص ماتركس', '07700000002')`, cust)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM "Booking" WHERE "customerId" = $1`, cust) //nolint:errcheck
		db.Exec(`DELETE FROM "Customer" WHERE id = $1`, cust)          //nolint:errcheck
	})
	loc, _ := time.LoadLocation("Asia/Baghdad")
	now := time.Now().In(loc)
	d := func(days, hour int) time.Time {
		y, m, dd := now.AddDate(0, 0, -days).Date()
		return time.Date(y, m, dd, hour, 0, 0, 0, loc)
	}
	p := "ML" + uuid.NewString()[:5]
	mk := func(code, status string, sched time.Time, done *time.Time, waiting bool) {
		id := uuid.NewString()
		db.MustExec(`INSERT INTO "Booking" (id, code, "customerId", status, "scheduledAt", "completedAt") VALUES ($1, $2, $3, $4::"BookingStatus", $5, $6)`,
			id, code, cust, status, sched, done)
		if waiting {
			db.MustExec(`UPDATE "Booking" SET "waitingSince" = now(), "waitingKind" = 'NO_ANSWER', "contactAttempts" = 2 WHERE id = $1`, id)
		}
	}
	next := d(5, 0).Add(32 * time.Hour)
	far := d(1, 12)
	mk(p+"-1day", "COMPLETED", d(5, 16), &next, false)
	mk(p+"-4days", "COMPLETED", d(5, 16), &far, false)
	mk(p+"-open1", "CONFIRMED", d(1, 16), nil, false)
	mk(p+"-open3", "CONFIRMED", d(3, 16), nil, false)
	mk(p+"-wait", "CONFIRMED", d(4, 16), nil, true)

	items, err := NewMatrixCommandRepository(db).LateItems(6, 0, 500)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, x := range items {
		got[x.Code] = true
	}
	for code, want := range map[string]bool{p + "-1day": false, p + "-4days": true, p + "-open1": false, p + "-open3": true, p + "-wait": false} {
		if got[code] != want {
			t.Errorf("%s: متأخر=%v، المتوقع %v", code, got[code], want)
		}
	}
}
