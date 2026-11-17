package repository

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// (دازاي 11-17) وضع التجربة ياخذ كود الحجز (B1206) أو معرّفه.
func TestResolveBookingRef_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()
	const cu, bk = "rbr-cu", "rbr-bk"
	clean := func() {
		db.Exec(`DELETE FROM "Mission" WHERE "bookingId" = $1`, bk)
		db.Exec(`DELETE FROM "Booking" WHERE id = $1`, bk)
		db.Exec(`DELETE FROM "Customer" WHERE id = $1`, cu)
	}
	clean()
	t.Cleanup(clean)
	db.MustExec(`INSERT INTO "Customer"(id, name, phone) VALUES ($1, 'x', '07700001206')`, cu)
	db.MustExec(`INSERT INTO "Booking"(id, code, "customerId", status) VALUES ($1, 'RBR1206', $2, 'PENDING')`, bk, cu)
	r := NewWaRepository(db)
	for _, in := range []string{"RBR1206", " rbr1206 ", bk} {
		if got := r.ResolveBookingRef(in); got != bk {
			t.Fatalf("%q ← %q، المفروض %q", in, got, bk)
		}
	}
	if got := r.ResolveBookingRef("مو-موجود"); got != "مو-موجود" {
		t.Fatalf("الي ما ينلگى يرجع كما هو: %q", got)
	}
}
