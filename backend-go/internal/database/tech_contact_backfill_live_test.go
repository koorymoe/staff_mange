package database

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// (ل) 11-24 بلاغ #58: الحجز المحوّل الي انحط «بالانتظار» بملاحظة ينعلّم «تم الاتصال»؛ الي ما بيه شي يبقى.
func TestTechContactBackfillLive(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", dsn)
	t.Cleanup(func() { db.Close() })
	if err := Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	var sql string
	for _, m := range autoMigrations() {
		if m.Version == "0524_data_tech_contact_backfill" {
			sql = m.SQL
		}
	}
	if sql == "" {
		t.Fatal("الترحيل مو موجود")
	}
	clean := func() {
		db.Exec(`DELETE FROM "Booking" WHERE id LIKE 'tcb-%'`)
		db.Exec(`DELETE FROM "Customer" WHERE id = 'tcb-cu'`)
		db.Exec(`DELETE FROM "Employee" WHERE id = 'tcb-t'`)
	}
	clean()
	t.Cleanup(clean)
	db.MustExec(`INSERT INTO "Employee"(id, name, role) VALUES ('tcb-t', 'tcb', 'TECHNICAL')`)
	db.MustExec(`INSERT INTO "Customer"(id, name, phone) VALUES ('tcb-cu', 'x', '07700005858')`)
	db.MustExec(`INSERT INTO "Booking"(id, code, "customerId", status, "handoverToId", "handoverAt", "handoverReason", "waitingSince", "waitingNote") VALUES
		('tcb-w', 'TCB-W', 'tcb-cu', 'WAITING', 'tcb-t', now() - interval '3 days', 'x', now() - interval '2 days', 'تواصلت ويه الزبون والليدر'),
		('tcb-n', 'TCB-N', 'tcb-cu', 'CONFIRMED', 'tcb-t', now() - interval '3 days', 'x', NULL, NULL)`)
	db.MustExec(sql)
	var w, n bool
	var note string
	db.Get(&w, `SELECT "techContactedAt" IS NOT NULL FROM "Booking" WHERE id = 'tcb-w'`)
	db.Get(&note, `SELECT COALESCE("techContactNote", '') FROM "Booking" WHERE id = 'tcb-w'`)
	db.Get(&n, `SELECT "techContactedAt" IS NOT NULL FROM "Booking" WHERE id = 'tcb-n'`)
	if !w || note != "تواصلت ويه الزبون والليدر" {
		t.Fatalf("الحجز المنتظر ما انعلّم: %v %q", w, note)
	}
	if n {
		t.Fatal("الحجز الي ما بيه اتصال انعلّم بالغلط")
	}
}
