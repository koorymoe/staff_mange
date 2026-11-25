package service

import (
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/database"
)

// (ل) 11-25 دازاي: سند لحجز (لموظف) وسند لمشروع خارجي (للكاشير)، الرقم يتسلسل، والإلغاء بسبب ما ينحسب.
func TestVouchersLive(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", dsn)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	clean := func() {
		db.Exec(`DELETE FROM "PaymentVoucher" WHERE "createdById" LIKE 'pv-%'`)
		db.Exec(`DELETE FROM "Booking" WHERE id = 'pv-b'`)
		db.Exec(`DELETE FROM "Customer" WHERE id = 'pv-cu'`)
		db.Exec(`DELETE FROM "Employee" WHERE id LIKE 'pv-%'`)
	}
	clean()
	t.Cleanup(clean)
	db.MustExec(`INSERT INTO "Employee"(id, name, role, status) VALUES ('pv-f', 'pv محاسب', 'FINANCE', 'ACTIVE'), ('pv-t', 'pv فني', 'TECHNICIAN', 'ACTIVE')`)
	db.MustExec(`INSERT INTO "Customer"(id, name, phone) VALUES ('pv-cu', 'x', '07700003131')`)
	db.MustExec(`INSERT INTO "Booking"(id, code, "customerId", status) VALUES ('pv-b', 'PV-B1', 'pv-cu', 'CONFIRMED')`)
	s := NewVouchers(db)
	now := time.Now()
	if _, err := s.Create("pv-f", VoucherInput{TargetKind: "BOOKING", Code: "NOPE", Purpose: "شراء كيبل", Amount: 1000, RecipientKind: "CASHIER"}, now); err == nil {
		t.Fatal("حجز مو موجود لازم ينرفض")
	}
	if _, err := s.Create("pv-f", VoucherInput{TargetKind: "EXTERNAL", ExternalName: "x", Purpose: "شراء", Amount: 1000, RecipientKind: "EMPLOYEE"}, now); err == nil {
		t.Fatal("موظف بلا اختيار لازم ينرفض")
	}
	a, err := s.Create("pv-f", VoucherInput{TargetKind: "BOOKING", Code: "pv-b1", Purpose: "شراء كيبل", Amount: 25000, RecipientKind: "EMPLOYEE", RecipientEmployeeID: "pv-t"}, now)
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.Create("pv-f", VoucherInput{TargetKind: "EXTERNAL", ExternalName: "فيلا الجادرية", Purpose: "أجور نقل", Amount: 40000, RecipientKind: "CASHIER"}, now)
	if err != nil {
		t.Fatal(err)
	}
	if b.Number != a.Number+1 || a.Recipient == nil || *a.Recipient != "pv فني" || b.TargetLabel != "فيلا الجادرية" {
		t.Fatalf("السندات غلط: %+v %+v", a, b)
	}
	if err := s.Cancel("pv-f", a.ID, ""); err == nil {
		t.Fatal("الإلغاء يحتاج سبب")
	}
	if err := s.Cancel("pv-f", a.ID, "انصرف بالغلط"); err != nil {
		t.Fatal(err)
	}
	l, _ := s.List("")
	var mine float64
	for _, v := range l.Items {
		if v.ID == a.ID && v.CancelledAt == nil {
			t.Fatal("الملغي لازم يبين ملغي")
		}
		if (v.ID == a.ID || v.ID == b.ID) && v.CancelledAt == nil {
			mine += v.Amount
		}
	}
	if mine != 40000 {
		t.Fatalf("الملغي ما ينحسب: %v", mine)
	}
}
