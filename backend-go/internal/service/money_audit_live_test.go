package service

import (
	"os"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
	"staffmange-api/internal/repository"
)

// (دازاي 11-26) فحص الأرقام: كل استعلام يشتغل، الفاتورة البلا ليدر تنمسك، والحصص تطابق النسب.
func TestMoneyAuditLive(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db := sqlx.MustConnect("postgres", dsn)
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	ids := []string{"ma_l", "ma_t1", "ma_t2"}
	invs := []string{"ma_i1", "ma_i2"}
	cleanup := func() {
		db.Exec(`DELETE FROM "EmployeeCommission" WHERE "leaderInvoiceId" = ANY($1)`, invs)
		db.Exec(`DELETE FROM "LeaderInvoice" WHERE id = ANY($1)`, invs)
		db.Exec(`DELETE FROM "Employee" WHERE id = ANY($1)`, ids)
	}
	cleanup()
	t.Cleanup(func() { cleanup(); db.Close() })
	db.MustExec(`INSERT INTO "Employee" (id, name, username, password, role, status, "isLeader") VALUES
		('ma_l', 'ma_l', 'ma_l', 'x', 'TECHNICIAN', 'ACTIVE', true), ('ma_t1', 'ma_t1', 'ma_t1', 'x', 'TECHNICIAN', 'ACTIVE', false),
		('ma_t2', 'ma_t2', 'ma_t2', 'x', 'TECHNICIAN', 'ACTIVE', false)`)
	db.MustExec(`INSERT INTO "LeaderInvoice"(id, "employeeId", systems, items, "netTotal", "accountingCode", status) VALUES
		('ma_i1', 'ma_l', '[]', '[]', 1000000, 'ma_c1', 'SUBMITTED'), ('ma_i2', 'ma_t2', '[]', '[]', 200000, 'ma_c2', 'SUBMITTED')`)
	db.MustExec(`INSERT INTO "EmployeeCommission"(id, "employeeId", "leaderInvoiceId", role, "totalCommission") VALUES
		('ma_c_l', 'ma_l', 'ma_i1', 'LEADER', 1), ('ma_c_1', 'ma_t1', 'ma_i1', 'TECHNICIAN', 1),
		('ma_c_2', 'ma_t2', 'ma_i2', 'TECHNICIAN', 1)`)

	scores := NewMatrixScoreService(repository.NewMatrixScoreRepository(db), nil, repository.NewSystemSwitchRepository(db), nil)
	a, err := NewMoneyAuditService(db, scores).Run("")
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]AuditCheck{}
	for _, s := range a.Sections {
		for _, c := range s.Checks {
			found[c.Key] = c
			for _, it := range c.Items {
				if strings.HasPrefix(it.Label, "تعذر الفحص") {
					t.Fatalf("%s: %s", c.Key, it.Label)
				}
			}
		}
	}
	nl := found["no_leader"]
	ok := false
	for _, it := range nl.Items {
		ok = ok || it.Ref == "ma_c2"
	}
	if !ok {
		t.Fatalf("الفاتورة البلا ليدر ما انمسكت: %+v", nl.Items)
	}
	if !found["split_match"].OK {
		t.Fatalf("الحصص ما تطابق: %s", found["split_match"].Value)
	}
	if !found["split_sum"].OK {
		t.Fatalf("مجموع التوزيع: %s", found["split_sum"].Value)
	}
	total := 0.0
	for _, l := range a.Money {
		if l.Label == "إدخال الشركة الكلي" {
			total = l.Amount
		}
	}
	if total < 1200000 {
		t.Fatalf("إدخال الشركة: %+v", a.Money)
	}
	t.Logf("period %s %s..%s problems=%d", a.Period, a.From, a.To, a.Problems)
}
