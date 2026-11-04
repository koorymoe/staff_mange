package repository

import (
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// TestReportsReconcile_Live (ع) 10-23: «٣٣ حجز شلون ٤٢ تقرير؟» — الفرق يتفسّر.
func TestReportsReconcile_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة — تخطي اختبار القاعدة الحية")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()
	const emp, cust = "__rc_e", "__rc_c"
	bks := []string{"__rc_b1", "__rc_b2", "__rc_b3", "__rc_b4"}
	cleanup := func() {
		for _, b := range bks {
			db.Exec(`DELETE FROM "WorkReport" WHERE "bookingId" = $1`, b)
			db.Exec(`DELETE FROM "BookingAssignment" WHERE "bookingId" = $1`, b)
			db.Exec(`DELETE FROM "Booking" WHERE id = $1`, b)
		}
		db.Exec(`DELETE FROM "Customer" WHERE id = $1`, cust)
		db.Exec(`DELETE FROM "Employee" WHERE id = $1`, emp)
	}
	cleanup()
	defer cleanup()
	db.MustExec(`INSERT INTO "Employee"(id, name, status, role) VALUES ($1, 'فني', 'ACTIVE', 'TECHNICIAN')`, emp)
	db.MustExec(`INSERT INTO "Customer"(id, "customerCode", name, phone, "createdAt") VALUES ($1, 987740, 'ز', '07700000740', now())`, cust)
	// b1 خلص وبيه ٣ زيارات (٣ تقارير) · b2 خلص بتقرير · b3 ما خلص بس بيه تقرير · b4 خلص بلا تقرير
	for i, st := range []string{"COMPLETED", "COMPLETED", "CONFIRMED", "COMPLETED"} {
		db.MustExec(`INSERT INTO "Booking"(id, code, "customerId", status, "completedAt", "createdAt", "updatedAt")
			VALUES ($1, $2, $3, $4::"BookingStatus", CASE WHEN $4 = 'COMPLETED' THEN now() AT TIME ZONE 'UTC' END, now(), now())`,
			bks[i], "RC-"+string(rune('1'+i)), cust, st)
		db.MustExec(`INSERT INTO "BookingAssignment"(id, "bookingId", "employeeId", role) VALUES (gen_random_uuid()::text, $1, $2, 'TECH_1')`, bks[i], emp)
	}
	for _, b := range []string{"__rc_b1", "__rc_b1", "__rc_b1", "__rc_b2", "__rc_b3"} {
		db.MustExec(`INSERT INTO "WorkReport"(id, "bookingId", "employeeId", "workStatus", "createdAt") VALUES (gen_random_uuid()::text, $1, $2, 'COMPLETED', now() AT TIME ZONE 'UTC')`, b, emp)
	}
	now := time.Now().UTC()
	rc, err := NewSmartKpiRepository(db).ReportsReconcile(emp, now.Add(-time.Hour), now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	want := ReportsReconcile{Bookings: 3, ExtraVisits: 2, MultiBookings: 1, OutsideMonth: 1, NoReport: 1}
	if rc != want {
		t.Fatalf("got %+v want %+v", rc, want)
	}
}
