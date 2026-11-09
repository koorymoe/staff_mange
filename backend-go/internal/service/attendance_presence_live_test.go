package service

import (
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/database"
	"staffmange-api/internal/repository"
)

// TestAutoCheckoutRespectsPresence (دازاي 11-06): فاتح النظام ويضغط = ما ينسكر خروجه حتى لو عبر
// وقت دوامه؛ ومن يسد النظام، ينسكر بوقت آخر ضغطة إله.
func TestAutoCheckoutRespectsPresence(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	emp := "ap-" + uuid.NewString()[:8]
	db.MustExec(`INSERT INTO "Employee" (id, name, role, status) VALUES ($1, 'فني فحص', 'TECHNICIAN', 'ACTIVE')`, emp)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM "Attendance" WHERE "employeeId" = $1`, emp)       //nolint:errcheck
		db.Exec(`DELETE FROM "EmployeePresence" WHERE "employeeId" = $1`, emp) //nolint:errcheck
		db.Exec(`DELETE FROM "Employee" WHERE id = $1`, emp)                   //nolint:errcheck
	})
	// حضور ٨ الصبح قبل يومين (الحد ٤ العصر، والسماح ٣ ساعات عبر من زمان)
	y, m, d := time.Now().In(debriefLoc).AddDate(0, 0, -2).Date()
	in := time.Date(y, m, d, 8, 0, 0, 0, debriefLoc)
	id := uuid.NewString()
	db.MustExec(`INSERT INTO "Attendance" (id, "employeeId", "checkIn", date) VALUES ($1, $2, ($3::timestamptz AT TIME ZONE 'UTC'), $4::date)`,
		id, emp, in, in.Format("2006-01-02"))
	svc := NewAttendanceService(repository.NewAttendanceRepository(db))
	open := func() bool {
		var n int
		db.Get(&n, `SELECT count(*) FROM "Attendance" WHERE id = $1 AND "checkOut" IS NULL`, id)
		return n == 1
	}

	// ١) ضغط شي قبل ٥ دقايق = بعده يشتغل، ما ينسكر
	db.MustExec(`INSERT INTO "EmployeePresence" ("employeeId", "lastSeenAt") VALUES ($1, now() - interval '5 minutes')`, emp)
	if _, err := svc.AutoCheckout(time.Now(), nil); err != nil {
		t.Fatal(err)
	}
	if !open() {
		t.Fatal("فاتح النظام ويضغط — ما لازم ينسكر خروجه")
	}

	// ٢) سد النظام: آخر ضغطة قبل ساعة → ينسكر بوقت آخر ضغطة
	db.MustExec(`UPDATE "EmployeePresence" SET "lastSeenAt" = now() - interval '1 hour' WHERE "employeeId" = $1`, emp)
	var last time.Time
	db.Get(&last, `SELECT "lastSeenAt" FROM "EmployeePresence" WHERE "employeeId" = $1`, emp)
	if _, err := svc.AutoCheckout(time.Now(), nil); err != nil {
		t.Fatal(err)
	}
	if open() {
		t.Fatal("سد النظام من ساعة — لازم ينسكر")
	}
	var out time.Time
	db.Get(&out, `SELECT "checkOut" AT TIME ZONE 'UTC' FROM "Attendance" WHERE id = $1`, id)
	if d := out.Sub(last); d > time.Second || d < -time.Second {
		t.Fatalf("الانصراف لازم بوقت آخر ضغطة: %v مقابل %v", out, last)
	}
}
