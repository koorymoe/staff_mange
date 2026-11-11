package repository

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// (ل) 11-11 طلب المخازن: أولوية نواقص الجرد — حجز اليوم/باچر = عاجل، +٧ أيام = مهم، واليدوي يغلب.
func TestShortagePriority_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	clean := func() {
		db.Exec(`DELETE FROM "InventoryCheck" WHERE "employeeId" IN ('ip-a','ip-b','ip-c')`)
		db.Exec(`DELETE FROM "BookingAssignment" WHERE "bookingId" IN ('ip-bk','ip-bk2')`)
		db.Exec(`DELETE FROM "Booking" WHERE id IN ('ip-bk','ip-bk2')`)
		db.Exec(`DELETE FROM "Customer" WHERE id = 'ip-cu'`)
		db.Exec(`DELETE FROM "Employee" WHERE id IN ('ip-a','ip-b','ip-c')`)
	}
	clean()
	defer clean()
	db.MustExec(`INSERT INTO "Employee"(id, name) VALUES ('ip-a','فني أ'),('ip-b','فني ب'),('ip-c','فني ج')`)
	db.MustExec(`INSERT INTO "Customer"(id, name, phone) VALUES ('ip-cu','زبون','07000000991')`)
	// فني أ: حجز باچر. فني ج: حجز بعد ٥ أيام (مو عاجل).
	db.MustExec(`INSERT INTO "Booking"(id, code, "customerId", "scheduledAt") VALUES
		('ip-bk','IP1','ip-cu', (baghdad_today() + 1)::timestamp + interval '7 hours'),
		('ip-bk2','IP2','ip-cu', (baghdad_today() + 5)::timestamp + interval '7 hours')`)
	db.MustExec(`INSERT INTO "BookingAssignment"(id, "bookingId", "employeeId", role) VALUES
		(gen_random_uuid()::text,'ip-bk','ip-a','TECH_1'),(gen_random_uuid()::text,'ip-bk2','ip-c','TECH_1')`)
	db.MustExec(`INSERT INTO "InventoryCheck"(id, "employeeId", complete, "missingItems", "checkedAt") VALUES
		('ip-ca','ip-a',false,'دريل', now() - interval '1 day'),
		('ip-cb','ip-b',false,'سلم', now() - interval '8 days'),
		('ip-cc','ip-c',false,'شريط', now() - interval '1 day')`)
	r := NewInventoryRepository(db)
	pri := func() map[string]string {
		rows, err := r.OpenShortages()
		if err != nil {
			t.Fatal(err)
		}
		m := map[string]string{}
		for _, c := range rows {
			if c.ID == "ip-ca" || c.ID == "ip-cb" || c.ID == "ip-cc" {
				m[c.ID] = c.EffPriority
			}
		}
		return m
	}
	if m := pri(); m["ip-ca"] != "URGENT" || m["ip-cb"] != "HIGH" || m["ip-cc"] != "NORMAL" {
		t.Fatalf("المحسوبة غلط: %v", m)
	}
	var urgent int
	if err := db.Get(&urgent, UrgentShortagesSQL()+` AND c."employeeId" LIKE 'ip-%'`); err != nil || urgent != 1 {
		t.Fatalf("عدد العاجل للملخص: %d %v", urgent, err)
	}
	if ok, err := r.SetShortagePriority("ip-cc", "URGENT", "ip-b"); !ok || err != nil {
		t.Fatalf("اليدوي: %v %v", ok, err)
	}
	if ok, _ := r.SetShortagePriority("ip-ca", "NORMAL", ""); !ok {
		t.Fatal("اليدوي")
	}
	if m := pri(); m["ip-cc"] != "URGENT" || m["ip-ca"] != "NORMAL" {
		t.Fatalf("اليدوي لازم يغلب: %v", m)
	}
	r.SetShortagePriority("ip-ca", "", "") //nolint:errcheck
	if m := pri(); m["ip-ca"] != "URGENT" {
		t.Fatalf("الفارغ يرجّعها محسوبة: %v", m)
	}
	db.MustExec(`UPDATE "InventoryCheck" SET resolved = true WHERE id = 'ip-cb'`)
	if ok, _ := r.SetShortagePriority("ip-cb", "HIGH", ""); ok {
		t.Fatal("المسكّر ما يتغيّر")
	}
}
