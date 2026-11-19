package repository

import (
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// (دازاي 11-19) فاتورة بلا حجز، بيها عمولة لليدر وفني: الفلوس تنقسم بيناتهم بالتساوي —
// قبل چانت الفاتورة بلا حجز ما تنحسب لأحد، والفني يطلع «دخّل ٠» وعنده عمولة.
func TestRevenueCrewFromCommissionLive(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db := sqlx.MustConnect("postgres", dsn)
	ids := []string{"rv_ldr", "rv_tech"}
	cleanup := func() {
		db.Exec(`DELETE FROM "EmployeeCommission" WHERE "leaderInvoiceId" = 'rv_inv'`)
		db.Exec(`DELETE FROM "LeaderInvoice" WHERE id = 'rv_inv'`)
		db.Exec(`DELETE FROM "Employee" WHERE id = ANY($1)`, ids)
	}
	cleanup()
	t.Cleanup(func() { cleanup(); db.Close() })
	for _, id := range ids {
		db.MustExec(`INSERT INTO "Employee" (id, name, username, password, role, status) VALUES ($1, $1, $1, 'x', 'TECHNICIAN', 'ACTIVE')`, id)
	}
	db.MustExec(`INSERT INTO "LeaderInvoice"(id, "employeeId", systems, items, "netTotal", "accountingCode", status) VALUES ('rv_inv', 'rv_ldr', '[]', '[]', 1000000, 'rv_c1', 'SUBMITTED')`)
	db.MustExec(`INSERT INTO "EmployeeCommission"(id, "employeeId", "leaderInvoiceId", role, "totalCommission") VALUES ('rv_c_l', 'rv_ldr', 'rv_inv', 'LEADER', 1), ('rv_c_t', 'rv_tech', 'rv_inv', 'TECHNICIAN', 1)`)
	rows, err := NewMatrixScoreRepository(db).Revenue(time.Now().Add(-time.Hour), time.Now().Add(time.Hour), "rv_tech")
	if err != nil {
		t.Fatal(err)
	}
	got := 0.0
	for _, r := range rows {
		if r.Kind == "WORK" {
			got += r.Amount
		}
	}
	if got != 500000 {
		t.Fatalf("الفني لازم ياخذ نص الفاتورة: %v (%+v)", got, rows)
	}
}
