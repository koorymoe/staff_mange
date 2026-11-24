package repository

import (
	"math"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
)

// (ل) 11-24 دازاي: الإحصائيات — ليدر 50٪ وفنيين 50٪ يتقاسمونها، بلا ليدر الفني 30٪، مسؤول الخدمة صفر.
// والإيراد — ليدر 40٪، فني 30٪، الشركة الباقي.
func TestMoneySplitLive(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db := sqlx.MustConnect("postgres", dsn)
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	ids := []string{"ms_l", "ms_t1", "ms_t2", "ms_sm"}
	invs := []string{"ms_i1", "ms_i2"}
	var saved *MoneySplit
	if m := LoadMoneySplit(db); m.UpdatedAt != nil {
		saved = &m
	}
	cleanup := func() {
		db.Exec(`DELETE FROM "EmployeeCommission" WHERE "leaderInvoiceId" = ANY($1)`, invs)
		db.Exec(`DELETE FROM "LeaderInvoice" WHERE id = ANY($1)`, invs)
		db.Exec(`DELETE FROM "Employee" WHERE id = ANY($1)`, ids)
	}
	cleanup()
	db.Exec(`DELETE FROM "MoneySplitConfig"`)
	t.Cleanup(func() {
		cleanup()
		db.Exec(`DELETE FROM "MoneySplitConfig"`)
		if saved != nil {
			SaveMoneySplit(db, *saved, "")
		}
		db.Close()
	})
	db.MustExec(`INSERT INTO "Employee" (id, name, username, password, role, status, "isLeader") VALUES
		('ms_l', 'ms_l', 'ms_l', 'x', 'TECHNICIAN', 'ACTIVE', true), ('ms_t1', 'ms_t1', 'ms_t1', 'x', 'TECHNICIAN', 'ACTIVE', false),
		('ms_t2', 'ms_t2', 'ms_t2', 'x', 'TECHNICIAN', 'ACTIVE', false), ('ms_sm', 'ms_sm', 'ms_sm', 'x', 'SERVICE_MANAGER', 'ACTIVE', false)`)
	// فاتورة ١: ليدر + فنيين اثنين، مليون
	db.MustExec(`INSERT INTO "LeaderInvoice"(id, "employeeId", systems, items, "netTotal", "accountingCode", status) VALUES
		('ms_i1', 'ms_l', '[]', '[]', 1000000, 'ms_c1', 'SUBMITTED'), ('ms_i2', 'ms_sm', '[]', '[]', 200000, 'ms_c2', 'SUBMITTED')`)
	db.MustExec(`INSERT INTO "EmployeeCommission"(id, "employeeId", "leaderInvoiceId", role, "totalCommission") VALUES
		('ms_c_l', 'ms_l', 'ms_i1', 'LEADER', 1), ('ms_c_1', 'ms_t1', 'ms_i1', 'TECHNICIAN', 1), ('ms_c_2', 'ms_t2', 'ms_i1', 'TECHNICIAN', 1),
		-- فاتورة ٢ (GPS): مسؤول الخدمة فوترها وفني واحد بلا ليدر، ٢٠٠ ألف
		('ms_c_s', 'ms_sm', 'ms_i2', 'TECHNICIAN', 1), ('ms_c_3', 'ms_t1', 'ms_i2', 'TECHNICIAN', 1)`)
	repo := NewMatrixScoreRepository(db)
	from, to := time.Now().Add(-time.Hour), time.Now().Add(time.Hour)
	rows, err := repo.Revenue(from, to, "")
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]float64{}
	for _, r := range rows {
		if r.Kind == "WORK" {
			got[r.EmployeeID] += r.Amount
		}
	}
	want := map[string]float64{
		"ms_l":  500000,          // 50٪ من مليون
		"ms_t1": 250000 + 60000, // 50٪÷2 من المليون + 30٪ من الـ200 (بلا ليدر)
		"ms_t2": 250000,
		"ms_sm": 0, // مسؤول الخدمة ما ينحسبله شي
	}
	for id, w := range want {
		if math.Abs(got[id]-w) > 0.5 {
			t.Fatalf("%s: لازم %v طلع %v (الكل %v)", id, w, got[id], got)
		}
	}
	p, err := repo.CompanyProfit(from, to)
	if err != nil {
		t.Fatal(err)
	}
	// مليون: ليدر 400، فني 300 (يتقاسمون)، شركة 300 · ٢٠٠: بلا ليدر، فني 60، شركة 140
	if p.Invoices < 2 || math.Abs(p.Company-p.Total+p.Leaders+p.Techs) > 1 {
		t.Fatalf("الربح ما يتطابق: %+v", p)
	}
	// النسب من الإعدادات: «لكل فني» يرفع حصة الفنيين
	m := DefaultMoneySplit
	m.RevenueTechEach = true
	m.StatsLeaderPct = 60
	if err := SaveMoneySplit(db, m, ""); err != nil {
		t.Fatal(err)
	}
	p2, _ := repo.CompanyProfit(from, to)
	if p2.Techs-p.Techs < 299999 { // فاتورة المليون: فنيين اثنين ← 300 ألف زيادة
		t.Fatalf("«لكل فني» ما انطبق: قبل %+v بعد %+v", p, p2)
	}
	rows, _ = repo.Revenue(from, to, "ms_l")
	if len(rows) == 0 || math.Abs(rows[0].Amount-600000) > 0.5 {
		t.Fatalf("نسبة الليدر من الإعدادات ما انطبقت: %+v", rows)
	}
}
