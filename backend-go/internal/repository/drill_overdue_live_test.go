package repository

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
)

// (دازاي 11-13) المتأخر حالتين بس: يوم الموعد خلص والكادر ما طلع، أو حجز جديد صارله
// يومين بلا تثبيت. الورق الناقص قائمة لحاله، وحجوزات المشاريع بـ«متابعة المشاريع».
func TestOverdueDefinition_Live(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	clean := func() {
		db.Exec(`DELETE FROM "Project" WHERE "bookingId" LIKE 'od-%'`)
		db.Exec(`DELETE FROM "BookingAssignment" WHERE "bookingId" LIKE 'od-%'`)
		db.Exec(`DELETE FROM "Booking" WHERE id LIKE 'od-%'`)
		db.Exec(`DELETE FROM "Customer" WHERE id = 'od-cu'`)
		db.Exec(`DELETE FROM "Employee" WHERE id = 'od-t'`)
	}
	clean()
	t.Cleanup(clean)
	db.MustExec(`INSERT INTO "Employee"(id, name, role) VALUES ('od-t', 'od', 'TECHNICIAN')`)
	db.MustExec(`INSERT INTO "Customer"(id, name, phone) VALUES ('od-cu', 'x', '07700006060')`)
	ins := func(id, status, extra string) {
		db.MustExec(`INSERT INTO "Booking"(id, code, "customerId", status) VALUES ($1, 'OD'||$1, 'od-cu', $2::"BookingStatus")`, id, status)
		if extra != "" {
			db.MustExec(`UPDATE "Booking" SET `+extra+` WHERE id = $1`, id)
		}
	}
	ins("od-new1", "PENDING", `"createdAt" = now() - interval '1 day'`)
	ins("od-new3", "PENDING", `"createdAt" = now() - interval '3 days'`)
	ins("od-today", "CONFIRMED", `"confirmedAt" = now() - interval '2 days', "scheduledAt" = date_trunc('day', now() AT TIME ZONE 'Asia/Baghdad') AT TIME ZONE 'Asia/Baghdad' AT TIME ZONE 'UTC' + interval '1 minute'`)
	ins("od-yday", "CONFIRMED", `"confirmedAt" = now() - interval '3 days', "scheduledAt" = now() - interval '30 hours'`)
	for _, b := range []string{"od-today", "od-yday"} {
		db.MustExec(`INSERT INTO "BookingAssignment"(id, "bookingId", "employeeId", role) VALUES (gen_random_uuid()::text, $1, 'od-t', 'TECH_1')`, b)
	}
	ins("od-paper", "COMPLETED", `"completedAt" = now() - interval '3 days'`)
	db.MustExec(`INSERT INTO "BookingAssignment"(id, "bookingId", "employeeId", role) VALUES (gen_random_uuid()::text, 'od-paper', 'od-t', 'TECH_1')`)
	ins("od-proj", "CONFIRMED", `"confirmedAt" = now() - interval '90 days', "transferToProjects" = true`)
	ins("od-rej", "CONFIRMED", `"confirmedAt" = now() - interval '90 days', "transferToProjects" = true`)
	db.MustExec(`INSERT INTO "Project"(id, code, name, stage, "bookingId") VALUES ('od-p1', 'OD-P1', 'x', '❌ مرفوض', 'od-rej')`)

	in := func(key, id string) bool {
		var n int
		if err := db.Get(&n, `SELECT `+DrillCountSQL(key)+` FROM "Booking" b WHERE b.id = $1 AND `+BookingCountableSQL("b"), id); err != nil {
			t.Fatal(err)
		}
		return n == 1
	}
	cases := []struct {
		key, id string
		want    bool
		why     string
	}{
		{"overdue", "od-new1", false, "حجز جديد بيوم واحد مو متأخر"},
		{"overdue", "od-new3", true, "حجز جديد ٣ أيام بلا تثبيت متأخر"},
		{"overdue", "od-today", false, "موعده اليوم — اليوم بعده ما خلص"},
		{"overdue", "od-yday", true, "موعده البارحة والكادر ما طلع"},
		{"overdue", "od-paper", false, "الورق الناقص مو بالمتأخرات"},
		{"paper_missing", "od-paper", true, "الورق الناقص بقائمته"},
		{"needs_coord", "od-proj", false, "حجز المشاريع مو «تنتظر تنسيق»"},
		{"projects_follow", "od-proj", true, "حجز المشاريع بـ«متابعة المشاريع»"},
		{"projects_follow", "od-rej", false, "المشروع المرفوض ما يطلع"},
	}
	for _, c := range cases {
		if got := in(c.key, c.id); got != c.want {
			t.Errorf("%s / %s: %v — %s", c.key, c.id, got, c.why)
		}
	}
}
