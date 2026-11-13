package repository

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
)

// بلاغ #38 (11-13): حجز انفوتر وهو بعده مو «منجز» لازم يطلع لمسؤول الخدمة
// حتى يسوي تقريره؛ وينشال بعد التقرير؛ والملغي ما يطلع.
func TestServicePaperworkInvoicedNotCompletedLive(t *testing.T) {
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
		db.Exec(`DELETE FROM "WorkReport" WHERE "bookingId" LIKE 'sp38-%'`)
		db.Exec(`DELETE FROM "LeaderInvoice" WHERE "bookingId" LIKE 'sp38-%'`)
		db.Exec(`DELETE FROM "BookingService" WHERE "bookingId" LIKE 'sp38-%'`)
		db.Exec(`DELETE FROM "Booking" WHERE id LIKE 'sp38-%'`)
		db.Exec(`DELETE FROM "Service" WHERE id IN ('sp38-svc', 'sp38-gps')`)
		db.Exec(`DELETE FROM "Customer" WHERE id = 'sp38-cu'`)
		db.Exec(`DELETE FROM "Employee" WHERE id = 'sp38-emp'`)
	}
	clean()
	t.Cleanup(clean)
	db.MustExec(`INSERT INTO "Employee"(id, name, role) VALUES ('sp38-emp', 'sp38', 'SERVICE_MANAGER')`)
	db.MustExec(`INSERT INTO "Customer"(id, name, phone) VALUES ('sp38-cu', 'x', '07700003838')`)
	db.MustExec(`INSERT INTO "Service"(id, name, "serviceKind") VALUES ('sp38-svc', 'sp38 داش', 'DASHCAM')`)
	for _, b := range [][2]string{{"sp38-run", "IN_PROGRESS"}, {"sp38-can", "CANCELLED"}, {"sp38-noinv", "IN_PROGRESS"}} {
		db.MustExec(`INSERT INTO "Booking"(id, code, "customerId", status, "serviceId") VALUES ($1, 'B'||$1, 'sp38-cu', $2, 'sp38-svc')`, b[0], b[1])
	}
	for _, b := range []string{"sp38-run", "sp38-can"} {
		db.MustExec(`INSERT INTO "LeaderInvoice"(id, "bookingId", "employeeId", systems, items, "netTotal", "accountingCode", status)
			VALUES ($1||'-i', $1, 'sp38-emp', '[]', '[]', 1000, $1||'-c', 'SUBMITTED')`, b)
	}
	repo := NewBookingRepository(db)
	has := func() map[string]bool {
		list, err := repo.ListServicePaperworkByKind("DASHCAM", 500)
		if err != nil {
			t.Fatal(err)
		}
		m := map[string]bool{}
		for _, b := range list {
			m[b.ID] = true
		}
		return m
	}
	m := has()
	if !m["sp38-run"] {
		t.Fatal("المفوتر الجاري لازم يطلع للتقرير")
	}
	if m["sp38-can"] {
		t.Fatal("الملغي ما يطلع")
	}
	if m["sp38-noinv"] {
		t.Fatal("الجاري بلا فاتورة ما يطلع")
	}
	// (11-13) مختلط جي بي اس + داش كام: فاتورة الجي بي اس ما تشيله من قائمة الداش كام.
	db.MustExec(`INSERT INTO "Service"(id, name, "serviceKind") VALUES ('sp38-gps', 'sp38 جي بي اس', 'GPS')`)
	db.MustExec(`INSERT INTO "Booking"(id, code, "customerId", status, "serviceId", "completedAt") VALUES ('sp38-mix', 'Bsp38-mix', 'sp38-cu', 'COMPLETED', 'sp38-gps', '2026-09-01')`)
	db.MustExec(`INSERT INTO "BookingService"(id, "bookingId", "serviceId") VALUES (gen_random_uuid()::text, 'sp38-mix', 'sp38-svc')`)
	db.MustExec(`INSERT INTO "LeaderInvoice"(id, "bookingId", "employeeId", systems, items, "netTotal", "accountingCode", status)
		VALUES ('sp38-mix-i', 'sp38-mix', 'sp38-emp', '["فاتورة الجي بي اس"]', '[]', 1000, 'sp38-mix-c', 'SUBMITTED')`)
	db.MustExec(`INSERT INTO "WorkReport"(id, "bookingId", "employeeId", "workStatus") VALUES ('sp38-mix-wr', 'sp38-mix', 'sp38-emp', 'COMPLETED')`)
	if !has()["sp38-mix"] {
		t.Fatal("المختلط يبقى بقائمة الداش كام لحد فاتورتها")
	}
	db.MustExec(`INSERT INTO "LeaderInvoice"(id, "bookingId", "employeeId", systems, items, "netTotal", "accountingCode", status)
		VALUES ('sp38-mix-i2', 'sp38-mix', 'sp38-emp', '["فاتورة الداش كام"]', '[]', 1000, 'sp38-mix-c2', 'SUBMITTED')`)
	if has()["sp38-mix"] {
		t.Fatal("بعد فاتورة الداش كام ينشال")
	}
	db.MustExec(`INSERT INTO "WorkReport"(id, "bookingId", "employeeId", "workStatus") VALUES ('sp38-wr', 'sp38-run', 'sp38-emp', 'COMPLETED')`)
	if has()["sp38-run"] {
		t.Fatal("بعد التقرير ينشال")
	}
}
