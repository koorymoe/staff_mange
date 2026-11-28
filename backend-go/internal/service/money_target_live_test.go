package service

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/database"
	"staffmange-api/internal/timeutil"
)

// (دازاي 11-28) «تنحسب الفلوس ليا شفت حسب المنسق»: حجز ثبّته منسّق الصباحي وانفوتر بالليل يروح للصباحي.
// والهدف اليومي اليدوي يغلب الحساب، وبند الفلوس يطلع بالتقييم.
func TestMoneyTargetsLive(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", dsn)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	var oldShift string
	_ = db.Get(&oldShift, `SELECT COALESCE("shiftTargets"::text, '{}') FROM "CostConfig" WHERE id = 1`)
	clean := func() {
		db.Exec(`DELETE FROM "LeaderInvoice" WHERE id LIKE 'mt-%'`)
		db.Exec(`DELETE FROM "Booking" WHERE id LIKE 'mt-%'`)
		db.Exec(`DELETE FROM "Customer" WHERE id LIKE 'mt-%'`)
		db.Exec(`DELETE FROM "CostEmployee" WHERE "employeeId" LIKE 'mt-%'`)
		db.Exec(`DELETE FROM "Employee" WHERE id LIKE 'mt-%'`)
	}
	clean()
	t.Cleanup(func() {
		clean()
		if oldShift != "" {
			db.Exec(`UPDATE "CostConfig" SET "shiftTargets" = $1::jsonb WHERE id = 1`, oldShift)
		}
	})
	db.MustExec(`INSERT INTO "Employee"(id, name, role) VALUES ('mt-t', 'mt فني', 'TECHNICIAN'), ('mt-m', 'mt منسق صباحي', 'HR_COORDINATOR'), ('mt-e', 'mt منسق مسائي', 'HR_COORDINATOR')`)
	c := NewCompanyCosts(db)
	m, e := "MORNING", "EVENING"
	for _, x := range []CostEmployee{
		{EmployeeID: "mt-t", Kind: "EARNER", Shift: &m, Salary: 600000, HoursPerDay: 8, DaysPerMonth: 24, DailyTarget: 30000},
		{EmployeeID: "mt-m", Kind: "BOOKING_ADMIN", Shift: &m, Salary: 500000, HoursPerDay: 8, DaysPerMonth: 24},
		{EmployeeID: "mt-e", Kind: "BOOKING_ADMIN", Shift: &e, Salary: 500000, HoursPerDay: 8, DaysPerMonth: 24},
	} {
		if err := c.SaveEmployee(x); err != nil {
			t.Fatal(err)
		}
	}
	if err := c.SaveEmployee(CostEmployee{EmployeeID: "mt-t", Kind: "EARNER", Salary: 1, HoursPerDay: 8, DaysPerMonth: 24, DailyTarget: -1}); err == nil {
		t.Fatal("هدف سالب لازم ينرفض")
	}
	if err := c.SetShiftTargets(map[string]float64{"MORNING": 100000, "EVENING": 80000, "X": 5}); err != nil {
		t.Fatal(err)
	}
	// حجز ثبّته منسّق الصباحي — انفوتر بالليل (الساعة ما تفرق)
	db.MustExec(`INSERT INTO "Customer"(id, name, phone) VALUES ('mt-cu', 'x', '07700009191')`)
	db.MustExec(`INSERT INTO "Booking"(id, code, "customerId", status, "confirmedByEmployeeId") VALUES ('mt-b', 'MTB1', 'mt-cu', 'COMPLETED', 'mt-m')`)
	db.MustExec(`INSERT INTO "LeaderInvoice"(id, "bookingId", "employeeId", systems, items, "netTotal", "accountingCode", status, "createdAt")
		VALUES ('mt-i', 'mt-b', 'mt-t', '[]', '[]', 250000, 'mt-c1', 'SUBMITTED', now())`)

	now := time.Now()
	rep, err := c.Report(timeutil.ThisPeriod(), now)
	if err != nil {
		t.Fatal(err)
	}
	for _, x := range rep.Targets {
		if x.EmployeeID == "mt-t" && (!x.Manual || x.Daily != 30000) {
			t.Fatalf("الهدف اليدوي ما انحسب: %+v", x)
		}
	}
	got := map[string]ShiftProgress{}
	for _, s := range rep.Shifts {
		got[s.Shift] = s
	}
	if got["MORNING"].Brought < 250000 || got["MORNING"].Daily != 100000 || got["EVENING"].Daily != 80000 {
		t.Fatalf("الشفتات غلط: %+v", rep.Shifts)
	}
	from, _, _ := timeutil.PeriodStart(timeutil.ThisPeriod())
	facts := c.MoneyFacts(from, now)
	if f, ok := facts["mt-m"]; !ok || !strings.Contains(f.Detail, "الصباحي") {
		t.Fatalf("منسّق الصباحي ما طلعله بند الفلوس: %+v", facts["mt-m"])
	}
	if f, ok := facts["mt-t"]; !ok || f.Pct < 0 || f.Pct > 100 || !strings.Contains(f.Detail, "30 ألف") {
		t.Fatalf("بند فلوس الفني غلط: %+v", f)
	}
}

func TestMoneyPartInScore(t *testing.T) {
	rows := []StaffScore{{ID: "x", Group: GroupTechs}}
	applyRoleScoresM(rows, nil, nil, map[string]MoneyFact{"x": {Pct: 50, Detail: "دخّل نص"}})
	found := false
	for _, p := range rows[0].TechParts {
		if p.Key == "MONEY" {
			found = true
		}
	}
	if !found {
		t.Fatalf("بند الفلوس ما طلع: %+v", rows[0].TechParts)
	}
}
