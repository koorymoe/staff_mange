package service

import (
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/database"
)

// (دازاي 11-13) الموظفين بنفس معايير المدراء: المهمة الي انطاها المدير بموعدها ترفع
// «اتباع الخطة والتعليمات»، والمتأخرة تنزلها وتدخل بالشكاوي والتأخير؛ المدير مو بقائمتهم.
func TestStaffRankingLive(t *testing.T) {
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
		db.Exec(`DELETE FROM "ExtraTask" WHERE "assignedToId" LIKE 'sr-%'`)
		db.Exec(`DELETE FROM "OrgAssignment" WHERE "employeeId" LIKE 'sr-%'`)
		db.Exec(`DELETE FROM "OrgDivision" WHERE id = 'sr-div'`)
		db.Exec(`DELETE FROM "Employee" WHERE id LIKE 'sr-%'`)
	}
	clean()
	t.Cleanup(clean)
	db.MustExec(`INSERT INTO "Employee"(id, name, role) VALUES ('sr-boss', 'sr مدير', 'SALES'), ('sr-good', 'sr ملتزم', 'SALES'), ('sr-bad', 'sr مو ملتزم', 'SALES')`)
	db.MustExec(`INSERT INTO "OrgDivision"(id, name) VALUES ('sr-div', 'sr إدارة')`)
	db.MustExec(`INSERT INTO "OrgAssignment"(id, "divisionId", "employeeId", position) VALUES
		('sr-a1', 'sr-div', 'sr-boss', 'DIRECTOR'), ('sr-a2', 'sr-div', 'sr-good', 'MEMBER'), ('sr-a3', 'sr-div', 'sr-bad', 'MEMBER')`)
	now := time.Now()
	due := now.Add(-48 * time.Hour)
	db.MustExec(`INSERT INTO "ExtraTask"(id, title, "assignedToId", "assignedById", "dueAt", "doneAt", status) VALUES
		('sr-t1', 'x', 'sr-good', 'sr-boss', $1, $2, 'DONE'), ('sr-t2', 'y', 'sr-bad', 'sr-boss', $1, NULL, 'PENDING')`, due, due.Add(-time.Hour))
	period := now.Format("2006-01")
	if due.Format("2006-01") != period {
		t.Skip("أول الشهر")
	}
	rows := staffRanking(db, period, nil)
	by := map[string]DirectorRankRow{}
	for _, r := range rows {
		by[r.EmployeeID] = r
	}
	if _, ok := by["sr-boss"]; ok {
		t.Fatal("المدير مو بقائمة الموظفين")
	}
	crit := func(id, k string) *float64 {
		for _, c := range by[id].Criteria {
			if c.Key == k {
				return c.Score
			}
		}
		return nil
	}
	g, b := crit("sr-good", "follow"), crit("sr-bad", "follow")
	if g == nil || b == nil || *g != 100 || *b >= *g {
		t.Fatalf("اتباع التعليمات: %v %v", g, b)
	}
	if c := crit("sr-bad", "complaints"); c == nil || *c != 80 {
		t.Fatalf("التأخير: %v", c)
	}
	if by["sr-good"].DivisionName != "sr إدارة · مديره sr مدير" {
		t.Fatalf("المدير: %q", by["sr-good"].DivisionName)
	}
}
