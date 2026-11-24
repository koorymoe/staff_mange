package service

import (
	"math"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/database"
)

// (ل) 11-24 دازاي: مثال مكتوب باليد — فنيين اثنين نفس الراتب والساعات، إداري 700، مصاريف 300،
// إداري حجوزات صباحي 400 (على فني الصباح بس)، وربح 10٪.
// (ل) 11-24 دازاي: «الأساسي 350 والباقي حسب الهدف»
func TestExpectedSalary(t *testing.T) {
	for _, c := range []struct{ sal, base, got, target, want float64 }{
		{600000, 350000, 800, 1000, 550000},  // 80٪ ← 350 + 250×0.8
		{600000, 350000, 0, 1000, 350000},    // صفر ← الأساسي
		{600000, 350000, 2000, 1000, 600000}, // عبر هدفه ← الكامل بس
		{300000, 350000, 500, 1000, 300000},  // أساسي أكبر من الراتب ← الراتب
	} {
		if g := ExpectedSalary(c.sal, c.base, c.got, c.target); g != c.want {
			t.Fatalf("%+v ← %v", c, g)
		}
	}
}

func TestComputeTargets(t *testing.T) {
	m, e := "MORNING", "EVENING"
	emps := []CostEmployee{
		{EmployeeID: "t1", Kind: "EARNER", Shift: &m, Salary: 600000, HoursPerDay: 8, DaysPerMonth: 26},
		{EmployeeID: "t2", Kind: "EARNER", Shift: &e, Salary: 600000, HoursPerDay: 8, DaysPerMonth: 26},
		{EmployeeID: "a1", Kind: "ADMIN", Salary: 700000, HoursPerDay: 8, DaysPerMonth: 26},
		{EmployeeID: "b1", Kind: "BOOKING_ADMIN", Shift: &m, Salary: 400000, HoursPerDay: 8, DaysPerMonth: 26},
	}
	ts, _, _, _ := ComputeTargets(emps, 300000, 10, 0, 30)
	if len(ts) != 2 {
		t.Fatalf("لازم هدفين: %+v", ts)
	}
	by := map[string]EmployeeTarget{ts[0].EmployeeID: ts[0], ts[1].EmployeeID: ts[1]}
	// t1: (600 + نص 1000 + كل 400) × 1.1 = 1,650,000 · t2: (600 + 500) × 1.1 = 1,210,000
	if by["t1"].Target != 1650000 || by["t2"].Target != 1210000 {
		t.Fatalf("الأهداف غلط: %+v", ts)
	}
	if by["t1"].Daily != math.Round(1650000.0/26) || by["t1"].Hourly != math.Round(math.Round(1650000.0/26)/8) {
		t.Fatalf("اليومي/الساعة غلط: %+v", by["t1"])
	}
	// فترة توقف ١٠ أيام من ٣٠ ← أيام الشغل تقل (26 − 26×10/30) والهدف اليومي يزيد، الشهري نفسه
	ts2, _, _, _ := ComputeTargets(emps, 300000, 10, 10, 30)
	for _, x := range ts2 {
		if x.EmployeeID == "t1" && (x.Target != 1650000 || x.WorkDays > 17.4 || x.Daily <= by["t1"].Daily) {
			t.Fatalf("التوقف ما انحسب: %+v", x)
		}
	}
	// إداري حجوزات مسائي وماكو فني مسائي ← يتوزع على الكل
	emps[1].Shift = &m
	emps[3].Shift = &e
	ts3, _, _, _ := ComputeTargets(emps, 300000, 0, 0, 30)
	if ts3[0].Target+ts3[1].Target != 600000*2+700000+300000+400000 {
		t.Fatalf("لازم كل التكاليف تتغطى: %+v", ts3)
	}
}

func TestCompanyCostsLive(t *testing.T) {
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
		db.Exec(`DELETE FROM "CostEmployee" WHERE "employeeId" LIKE 'cc-%'`)
		db.Exec(`DELETE FROM "CostExpense" WHERE month = '2031-02'`)
		db.Exec(`DELETE FROM "DowntimePeriod" WHERE reason LIKE 'cc-%'`)
		db.Exec(`DELETE FROM "Employee" WHERE id LIKE 'cc-%'`)
	}
	clean()
	t.Cleanup(clean)
	db.MustExec(`INSERT INTO "Employee"(id, name, role) VALUES ('cc-t', 'cc فني', 'TECHNICIAN'), ('cc-a', 'cc إداري', 'HR_COORDINATOR')`)
	c := NewCompanyCosts(db)
	if err := c.SaveEmployee(CostEmployee{EmployeeID: "cc-t", Kind: "EARNER", Salary: 600000, HoursPerDay: 8, DaysPerMonth: 24}); err != nil {
		t.Fatal(err)
	}
	if err := c.SaveEmployee(CostEmployee{EmployeeID: "cc-a", Kind: "ADMIN", Salary: 500000, HoursPerDay: 8, DaysPerMonth: 24}); err != nil {
		t.Fatal(err)
	}
	if err := c.SaveEmployee(CostEmployee{EmployeeID: "cc-a", Kind: "BOSS", HoursPerDay: 8, DaysPerMonth: 24}); err == nil {
		t.Fatal("نوع غلط لازم ينرفض")
	}
	if err := c.AddExpense(CostExpense{Month: "2031-02", Kind: "FUEL", Amount: 100000}, ""); err != nil {
		t.Fatal(err)
	}
	// شباط 2031 = 28 يوم؛ توقف 7 أيام (من 25 شباط لـ3 آذار ← 4 أيام بس جوه الشهر) + فترة متداخلة ما تنعد مرتين
	if err := c.AddDowntime(DowntimePeriod{StartDate: "2031-02-25", EndDate: "2031-03-03", Reason: "cc-عيد"}, ""); err != nil {
		t.Fatal(err)
	}
	c.AddDowntime(DowntimePeriod{StartDate: "2031-02-26", EndDate: "2031-02-27", Reason: "cc-متداخلة"}, "")
	if err := c.AddDowntime(DowntimePeriod{StartDate: "2031-02-10", EndDate: "2031-02-01", Reason: "cc-غلط"}, ""); err == nil {
		t.Fatal("النهاية قبل البداية لازم تنرفض")
	}
	rep, err := c.Report("2031-02", time.Date(2031, 2, 15, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if rep.DowntimeDays != 4 || len(rep.Targets) != 1 || rep.Targets[0].Target != 1200000 || rep.TotalCost != 1200000 {
		t.Fatalf("التقرير غلط: توقف %d، أهداف %+v، تكلفة %v", rep.DowntimeDays, rep.Targets, rep.TotalCost)
	}
}

func TestTargetMessage(t *testing.T) {
	m := TargetMessage(EmployeeTarget{Brought: 400000, Target: 1000000, Pct: 40, Remaining: 600000, DailyNeeded: 50000, Forecast: 800000, ExpectedSalary: 550000})
	for _, w := range []string{"400,000", "1,000,000", "40٪", "50,000", "800,000", "550,000"} {
		if !strings.Contains(m, w) {
			t.Fatalf("الرسالة ناقصها %q: %s", w, m)
		}
	}
}
