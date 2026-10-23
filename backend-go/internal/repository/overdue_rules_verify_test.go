package repository

import (
	"os"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// TestOverdueRules_Live تعريف «المتأخر» (ع) 10-23: كل مرحلة من لحظتها، وبانتظار
// قرار الزبون مو متأخر ويتأرشف استفسار بعد ٥ أيام، والكادر يطلع كامل بالليدر.
func TestOverdueRules_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة — تخطي اختبار القاعدة الحية")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()
	const cust, ldr, tech = "__od_c", "__od_l", "__od_t"
	ids := []string{"__od_contact", "__od_fresh", "__od_wait", "__od_depart", "__od_future", "__od_paper", "__od_inq"}
	cleanup := func() {
		for _, id := range ids {
			db.Exec(`DELETE FROM "BookingAssignment" WHERE "bookingId" = $1`, id)
			db.Exec(`DELETE FROM "Booking" WHERE id = $1`, id)
		}
		db.Exec(`DELETE FROM "Customer" WHERE id = $1`, cust)
		db.Exec(`DELETE FROM "Employee" WHERE id IN ($1, $2)`, ldr, tech)
	}
	cleanup()
	defer cleanup()
	db.MustExec(`INSERT INTO "Employee"(id, name, status, role, "isLeader") VALUES ($1, 'ليدر الاختبار', 'ACTIVE', 'TECHNICIAN', true), ($2, 'فني الاختبار', 'ACTIVE', 'TECHNICIAN', false)`, ldr, tech)
	db.MustExec(`INSERT INTO "Customer"(id, "customerCode", name, phone, "createdAt") VALUES ($1, 987711, 'ز', '07700000077', now())`, cust)
	ins := func(id, code, extra string) {
		db.MustExec(`INSERT INTO "Booking"(id, code, "customerId", status, "createdAt", "updatedAt") VALUES ($1, $2, $3, 'PENDING', now() AT TIME ZONE 'UTC', now())`, id, code, cust)
		if extra != "" {
			db.MustExec(`UPDATE "Booking" SET `+extra+` WHERE id = $1`, id)
		}
	}
	utc := `(now() AT TIME ZONE 'UTC')`
	ins("__od_contact", "OD-1", `"createdAt" = `+utc+` - interval '3 hours'`)
	ins("__od_fresh", "OD-2", `"createdAt" = `+utc+` - interval '1 hour'`)
	ins("__od_wait", "OD-3", `"createdAt" = `+utc+` - interval '3 days', "waitingSince" = now() - interval '2 days', "waitingKind" = 'CUSTOMER_DECISION', status = 'WAITING'`)
	ins("__od_depart", "OD-4", `status = 'CONFIRMED', "confirmedAt" = `+utc+`, "scheduledAt" = `+utc+` - interval '2 hours'`)
	ins("__od_future", "OD-5", `status = 'CONFIRMED', "confirmedAt" = `+utc+`, "scheduledAt" = `+utc+` + interval '1 day'`)
	ins("__od_paper", "OD-6", `status = 'COMPLETED', "completedAt" = `+utc+` - interval '30 hours'`)
	ins("__od_inq", "OD-7", `"waitingSince" = now() - interval '6 days', "waitingKind" = 'CUSTOMER_DECISION', status = 'WAITING'`)
	for _, b := range []string{"__od_depart", "__od_future"} {
		db.MustExec(`INSERT INTO "BookingAssignment"(id, "bookingId", "employeeId", role) VALUES (gen_random_uuid()::text, $1, $2, 'TECH_1'), (gen_random_uuid()::text, $1, $3, 'TECH_2')`, b, ldr, tech)
	}

	n, err := NewBookingRepository(db).ArchiveStaleInquiries()
	if err != nil || n < 1 {
		t.Fatalf("ArchiveStaleInquiries: %d %v", n, err)
	}
	var reason string
	db.Get(&reason, `SELECT COALESCE("archiveReason", '') FROM "Booking" WHERE id = '__od_inq'`)
	if !strings.Contains(reason, "استفسار فقط") {
		t.Fatalf("الاستفسار +٥ أيام لازم يتأرشف: %q", reason)
	}

	res, err := Drill(db, "overdue", 500)
	if err != nil {
		t.Fatalf("Drill: %v", err)
	}
	got := map[string]DrillRow{}
	for _, r := range res.Rows {
		got[r.Code] = r
	}
	want := map[string]string{"OD-1": "تواصل", "OD-4": "ما طلع", "OD-6": "فاتورة"}
	for code, kind := range want {
		r, ok := got[code]
		if !ok || !strings.Contains(r.Station, kind) {
			t.Fatalf("%s لازم يطلع متأخر «%s»: %+v", code, kind, r)
		}
	}
	for _, code := range []string{"OD-2", "OD-3", "OD-5", "OD-7"} {
		if _, ok := got[code]; ok {
			t.Fatalf("%s مو متأخر", code)
		}
	}
	if c := got["OD-4"].Crew; c == nil || !strings.HasPrefix(*c, "⭐ ليدر الاختبار") || !strings.Contains(*c, "فني الاختبار") {
		t.Fatalf("الكادر لازم كامل والليدر أول: %v", c)
	}
}
