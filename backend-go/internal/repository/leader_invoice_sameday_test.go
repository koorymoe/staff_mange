package repository

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/model"
)

// TestLeaderInvoice_OnePerBookingPerDay_Live قرار (ع) 10-22: فاتورة وحدة للحجز
// باليوم، حتى من ليدر ثاني؛ ويوم ثاني أو بعد السحب تنقبل.
func TestLeaderInvoice_OnePerBookingPerDay_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة — تخطي اختبار القاعدة الحية")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()
	r := NewLeaderInvoiceRepository(db)
	const l1, l2, cust, bk = "__inv_l1", "__inv_l2", "__inv_cust", "__inv_bk"
	cleanup := func() {
		db.Exec(`DELETE FROM "LeaderInvoice" WHERE "bookingId" = $1`, bk)
		db.Exec(`DELETE FROM "Booking" WHERE id = $1`, bk)
		db.Exec(`DELETE FROM "Customer" WHERE id = $1`, cust)
		db.Exec(`DELETE FROM "Employee" WHERE id IN ($1, $2)`, l1, l2)
	}
	cleanup()
	defer cleanup()
	db.MustExec(`INSERT INTO "Employee"(id, name, status) VALUES ($1, 'ليدر ١', 'ACTIVE'), ($2, 'ليدر ٢', 'ACTIVE')`, l1, l2)
	db.MustExec(`INSERT INTO "Customer"(id, "customerCode", name, phone, "createdAt") VALUES ($1, 987655, 'زبون', '07709998877', now())`, cust)
	db.MustExec(`INSERT INTO "Booking"(id, code, "customerId", status, "createdAt", "updatedAt") VALUES ($1, 'INV-1', $2, 'COMPLETED', now(), now())`, bk, cust)

	b := bk
	mk := func(emp string) error {
		_, err := r.Create(&model.LeaderInvoice{BookingID: &b, EmployeeID: emp, NetTotal: 1000, Status: "SUBMITTED"}, nil)
		return err
	}
	if err := mk(l1); err != nil {
		t.Fatalf("الأولى: %v", err)
	}
	if err := mk(l1); err != ErrInvoiceSameDay {
		t.Fatalf("الضغطة الثانية لازم ترفض: %v", err)
	}
	if err := mk(l2); err != ErrInvoiceSameDay {
		t.Fatalf("ليدر ثاني بنفس اليوم لازم يرفض: %v", err)
	}
	// الحجز انطلعله يوم ثاني: الأولى صارت أمس
	db.MustExec(`UPDATE "LeaderInvoice" SET "createdAt" = "createdAt" - interval '1 day' WHERE "bookingId" = $1`, bk)
	if err := mk(l2); err != nil {
		t.Fatalf("يوم ثاني لازم تنقبل: %v", err)
	}
	// السحب يفتح المجال من جديد
	db.MustExec(`UPDATE "LeaderInvoice" SET "revokedAt" = now() WHERE "bookingId" = $1`, bk)
	if err := mk(l1); err != nil {
		t.Fatalf("بعد السحب لازم تنقبل: %v", err)
	}
}
