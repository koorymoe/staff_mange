package service

import (
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// (دازاي 11-13) «جاي يروح للتقني يگله روح گل لليدر خل يقيّم»: واجب تقييم الليدر
// (COORD_LEADER) على الإداري الي ثبّت/سجّل الحجز بس — الفني الي سجّل حجز ما يتطالب.
func TestRatingDutyCoordRaterNotTechnician_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db := sqlx.MustConnect("postgres", dsn)
	defer db.Close()
	clean := func() {
		db.Exec(`DELETE FROM "BookingAssignment" WHERE "bookingId" LIKE 'rdr-%'`)
		db.Exec(`DELETE FROM "Booking" WHERE id LIKE 'rdr-%'`)
		db.Exec(`DELETE FROM "Customer" WHERE id = 'rdr-cu'`)
		db.Exec(`DELETE FROM "Employee" WHERE id LIKE 'rdr-%'`)
	}
	clean()
	defer clean()
	db.MustExec(`INSERT INTO "Employee"(id, name, role, "isLeader") VALUES ('rdr-lead', 'rdr-lead', 'TECHNICIAN', true),
		('rdr-tech', 'rdr-tech', 'TECHNICIAN', false), ('rdr-sales', 'rdr-sales', 'SALES', false)`)
	db.MustExec(`INSERT INTO "Customer"(id, name, phone) VALUES ('rdr-cu', 'x', '07700009191')`)
	db.MustExec(`INSERT INTO "Booking"(id, code, "customerId", status, "completedAt", "createdById") VALUES
		('rdr-b1', 'RDR-1', 'rdr-cu', 'COMPLETED', now(), 'rdr-tech'), ('rdr-b2', 'RDR-2', 'rdr-cu', 'COMPLETED', now(), 'rdr-sales')`)
	db.MustExec(`INSERT INTO "BookingAssignment"(id, "bookingId", "employeeId", role) VALUES
		(gen_random_uuid()::text, 'rdr-b1', 'rdr-lead', 'TECH_1'), (gen_random_uuid()::text, 'rdr-b2', 'rdr-lead', 'TECH_1')`)
	loc, _ := time.LoadLocation("Asia/Baghdad")
	day := time.Now().In(loc).Format("2006-01-02")
	d := NewRatingDuty(db)
	count := func(rater string) int {
		items, err := d.Owed(rater, day)
		if err != nil {
			t.Fatal(err)
		}
		n := 0
		for _, it := range items {
			if it.Stage == "COORD_LEADER" {
				n++
			}
		}
		return n
	}
	if n := count("rdr-tech"); n != 0 {
		t.Fatalf("الفني ما يتطالب بتقييم الليدر: %d", n)
	}
	if n := count("rdr-sales"); n != 1 {
		t.Fatalf("الإداري الي سجّل يقيّم الليدر: %d", n)
	}
}
