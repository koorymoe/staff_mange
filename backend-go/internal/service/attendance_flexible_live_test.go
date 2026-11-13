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

// (دازاي 11-13) بدون جدول: بعد ٨ ساعات من الحضور ينسكر انصرافه تلقائياً بوقت
// الثمانية بالضبط — إلا إذا بعده فاتح النظام ويشتغل؛ وما ينجبر يسجّل بوقت معيّن.
func TestFlexibleAutoCheckout_Live(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	flex, busy := "fx-"+uuid.NewString()[:8], "fx-"+uuid.NewString()[:8]
	for _, e := range []string{flex, busy} {
		db.MustExec(`INSERT INTO "Employee" (id, name, role, status, "flexibleHours") VALUES ($1, 'بلا جدول', 'IT_SUPPORT', 'ACTIVE', true)`, e)
	}
	t.Cleanup(func() {
		for _, e := range []string{flex, busy} {
			db.Exec(`DELETE FROM "AttendanceAuto" WHERE "employeeId" = $1`, e)   //nolint:errcheck
			db.Exec(`DELETE FROM "Attendance" WHERE "employeeId" = $1`, e)       //nolint:errcheck
			db.Exec(`DELETE FROM "EmployeePresence" WHERE "employeeId" = $1`, e) //nolint:errcheck
			db.Exec(`DELETE FROM "Employee" WHERE id = $1`, e)                   //nolint:errcheck
		}
	})
	in := time.Now().Add(-9 * time.Hour).Truncate(time.Second)
	ids := map[string]string{}
	for _, e := range []string{flex, busy} {
		ids[e] = uuid.NewString()
		db.MustExec(`INSERT INTO "Attendance" (id, "employeeId", "checkIn", date) VALUES ($1, $2, ($3::timestamptz AT TIME ZONE 'UTC'), $4::date)`,
			ids[e], e, in, in.In(debriefLoc).Format("2006-01-02"))
	}
	// الثاني بعده يشتغل: ضغط بالنظام قبل دقيقتين
	db.MustExec(`INSERT INTO "EmployeePresence" ("employeeId", "lastSeenAt") VALUES ($1, now() - interval '2 minutes')`, busy)
	svc := NewAttendanceService(repository.NewAttendanceRepository(db))
	if _, err := svc.AutoCheckout(time.Now(), nil); err != nil {
		t.Fatal(err)
	}
	var out *time.Time
	db.Get(&out, `SELECT "checkOut" AT TIME ZONE 'UTC' FROM "Attendance" WHERE id = $1`, ids[flex])
	if out == nil || out.Sub(in.Add(8*time.Hour)).Abs() > time.Second {
		t.Fatalf("بلا جدول: ينسكر بعد ٨ ساعات من الحضور بالضبط: %v", out)
	}
	var open int
	db.Get(&open, `SELECT count(*) FROM "Attendance" WHERE id = $1 AND "checkOut" IS NULL`, ids[busy])
	if open != 1 {
		t.Fatal("بعده فاتح النظام ويشتغل — ما ينسكر")
	}
	g, err := svc.Gate(flex, time.Now())
	if err != nil || g.Required {
		t.Fatalf("بلا جدول ما ينجبر يسجّل بوقت معيّن: %+v %v", g, err)
	}
}
