package repository

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
)

// (دازاي 11-15) المطابقة الثلاثية: تقرير الليدر «انجز» ↔ حالة الحجز والفلوس ↔ اتصال الجودة.
func TestCrossCheckLive(t *testing.T) {
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
		db.Exec(`DELETE FROM "QualityFollowUp" WHERE "bookingId" LIKE 'xc-%'`)
		db.Exec(`DELETE FROM "WorkReport" WHERE "bookingId" LIKE 'xc-%'`)
		db.Exec(`DELETE FROM "LeaderInvoice" WHERE "bookingId" LIKE 'xc-%'`)
		db.Exec(`DELETE FROM "Booking" WHERE id LIKE 'xc-%'`)
		db.Exec(`DELETE FROM "Customer" WHERE id = 'xc-cu'`)
		db.Exec(`DELETE FROM "Employee" WHERE id = 'xc-ld'`)
	}
	clean()
	t.Cleanup(clean)
	db.MustExec(`INSERT INTO "Employee"(id, name, role) VALUES ('xc-ld', 'xc ليدر', 'TECHNICIAN')`)
	db.MustExec(`INSERT INTO "Customer"(id, name, phone) VALUES ('xc-cu', 'x', '07700007711')`)
	for _, b := range [][3]string{{"xc-open", "IN_PROGRESS", "0"}, {"xc-money", "COMPLETED", "100000"}, {"xc-q", "COMPLETED", "500000"}, {"xc-ok", "COMPLETED", "500000"}} {
		db.MustExec(`INSERT INTO "Booking"(id, code, "customerId", status, "amountCollected", "completedAt") VALUES ($1, 'B'||$1, 'xc-cu', $2, $3::numeric, now())`, b[0], b[1], b[2])
		db.MustExec(`INSERT INTO "WorkReport"(id, "bookingId", "employeeId", "workStatus", "createdAt") VALUES ($1||'-w', $1, 'xc-ld', 'COMPLETED', now() - interval '3 days')`, b[0])
	}
	for _, b := range []string{"xc-money", "xc-q", "xc-ok"} {
		db.MustExec(`INSERT INTO "LeaderInvoice"(id, "bookingId", "employeeId", systems, items, "netTotal", "accountingCode", status)
			VALUES ($1||'-i', $1, 'xc-ld', '[]', '[]', 500000, $1||'-c', 'SUBMITTED')`, b)
	}
	db.MustExec(`INSERT INTO "QualityFollowUp"(id, "bookingId", "customerId", status, "contactNotes") VALUES ('xc-qf', 'xc-q', 'xc-cu', 'CONTACTED_ISSUE', 'الكاميرا ما تشتغل')`)
	db.MustExec(`INSERT INTO "QualityFollowUp"(id, "bookingId", "customerId", status) VALUES ('xc-qok', 'xc-ok', 'xc-cu', 'CONTACTED_OK')`)

	rows, err := NewMatrixWatchRepository(db).CrossCheck()
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, r := range rows {
		if len(r.BookingID) > 3 && r.BookingID[:3] == "xc-" {
			got[r.BookingID] += r.Kind + " "
		}
	}
	want := map[string]string{"xc-open": "REPORT_DONE_NOT_DONE ", "xc-money": "MONEY_GAP ", "xc-q": "QUALITY_CONTRADICTS "}
	for k, v := range want {
		if got[k] != v {
			t.Fatalf("%s: got %q want %q (all %v)", k, got[k], v, got)
		}
	}
	if got["xc-ok"] != "" {
		t.Fatalf("الحجز المطابق ما يطلع: %q", got["xc-ok"])
	}
}
