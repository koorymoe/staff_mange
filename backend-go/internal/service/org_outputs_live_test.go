package service

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
)

// TestOutputs_Live النقطة ٨: نقاط ماتركس تنحسب على فقراتها، المهمة المربوطة والـKPI
// المربوط بمعياره والشكوى يدخلون فقرتهم، والوحدة تجمع كادرها، وأضعف الفقرات تطلع.
func TestOutputs_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	const emp, cust, crit = "__out_emp", "__out_cust", "__out_crit"
	cleanup := func() {
		db.Exec(`DELETE FROM "MatrixScore" WHERE "employeeId" = $1`, emp)
		db.Exec(`DELETE FROM "ExtraTask" WHERE "assignedToId" = $1`, emp)
		db.Exec(`DELETE FROM "KpiEvaluation" WHERE "employeeId" = $1`, emp)
		db.Exec(`DELETE FROM "KpiCriterion" WHERE id = $1`, crit)
		db.Exec(`DELETE FROM "Complaint" WHERE "customerId" = $1`, cust)
		db.Exec(`DELETE FROM "Customer" WHERE id = $1`, cust)
		db.Exec(`DELETE FROM "OrgAssignment" WHERE "employeeId" = $1`, emp)
		db.Exec(`DELETE FROM "OrgPermissionGrant" WHERE "employeeId" = $1`, emp)
		db.Exec(`DELETE FROM "EmployeePermission" WHERE "employeeId" = $1`, emp)
		db.Exec(`DELETE FROM "PermissionAuditLog" WHERE "employeeId" = $1`, emp)
		db.Exec(`DELETE FROM "Employee" WHERE id = $1`, emp)
	}
	cleanup()
	defer cleanup()
	db.MustExec(`INSERT INTO "Employee"(id, name, status, role) VALUES ($1, 'موظف مخرجات', 'ACTIVE', 'TECHNICIAN')`, emp)
	s := NewOrgService(db)
	_ = s.Seed()
	if _, err := s.Assign("", OrgAssignInput{EmployeeID: emp, UnitID: "it.main", Position: "MEMBER"}); err != nil {
		t.Skipf("المكان: %v", err)
	}
	db.MustExec(`UPDATE "OrgAssignment" SET "startedAt" = now() - interval '2 months' WHERE "employeeId" = $1`, emp)
	now := time.Now().In(baghdad)
	at := time.Date(now.Year(), now.Month(), 1, 12, 0, 0, 0, baghdad)
	for i, r := range []struct {
		rule string
		p    int
	}{{"ATTENDANCE", 0}, {"ATTENDANCE", 2}, {"PAPER", 2}, {"PAPER", 2}} {
		db.MustExec(`INSERT INTO "MatrixScore"("employeeId", source, "sourceId", rule, points, "maxPoints", reason, "at") VALUES ($1, 'DAY', $2, $3, $4, 2, 'x', $5)`,
			emp, "__out_"+string(rune('a'+i)), r.rule, r.p, at.Add(time.Duration(i)*time.Hour))
	}
	db.MustExec(`INSERT INTO "ExtraTask"(id, title, "assignedToId", "dueAt", "doneAt", clause) VALUES ('__out_t1', 'تقرير', $1, $2, $3, 4)`, emp, at, at.Add(48*time.Hour))
	db.MustExec(`INSERT INTO "KpiCriterion"(id, label, clause) VALUES ($1, '__تأخير التصاميم', 7)`, crit)
	if err := SetCriterionClause(db, crit, intp(7)); err != nil {
		t.Fatal(err)
	}
	db.MustExec(`INSERT INTO "KpiEvaluation"(id, "employeeId", "evaluatorId", points, reason, "deductionAmount", "createdAt") VALUES ('__out_k1', $1, $1, -1, '__تأخير التصاميم', 10000, $2)`, emp, at)
	db.MustExec(`INSERT INTO "Customer"(id, name, phone) VALUES ($1, 'زبون مخرجات', '07700000097')`, cust)
	db.MustExec(`INSERT INTO "Complaint"(id, "customerId", description, status, "createdByEmployeeId", "relatedEmployeeId", "createdAt") VALUES ('__out_c1', $1, 'x', 'RESOLVED', $2, $2, $3)`, cust, emp, at)

	period := at.Format("2006-01")
	card, err := OutputsScorecard(db, emp, "", "", period)
	if err != nil {
		t.Fatal(err)
	}
	get := func(c int) OutputRow {
		for _, r := range card.Rows {
			if r.Clause == c {
				return r
			}
		}
		t.Fatalf("ماكو فقرة %d", c)
		return OutputRow{}
	}
	if r := get(3); r.MatrixPct == nil || *r.MatrixPct != 50 { // الحضور ← م١/ف٣
		t.Fatalf("م١/ف٣: %+v", r)
	}
	if r := get(15); r.MatrixPct == nil || *r.MatrixPct != 100 { // الورق ← م٤/ف٢
		t.Fatalf("م٤/ف٢: %+v", r)
	}
	if r := get(4); r.TasksLate != 1 || r.Score == nil || *r.Score != 0 { // مهمة متأخرة ← م١/ف٤
		t.Fatalf("م١/ف٤: %+v", r)
	}
	if r := get(7); r.KpiDeducts != 1 || r.Score == nil || *r.Score != 90 {
		t.Fatalf("م٢/ف١ (KPI): %+v", r)
	}
	if r := get(16); r.Complaints != 1 {
		t.Fatalf("م٤/ف٣ (شكوى): %+v", r)
	}
	if len(card.Weakest) == 0 || !strings.Contains(card.Weakest[0], "م1/ف4") {
		t.Fatalf("أضعف الفقرات: %v", card.Weakest)
	}
	unit, err := OutputsScorecard(db, "", "it.main", "", period)
	if err != nil || unit.People < 1 {
		t.Fatalf("الوحدة: %v %+v", err, unit)
	}
	if !s.CanSeeOutputs(emp, "TECHNICIAN", emp, "", "") || s.CanSeeOutputs(emp, "TECHNICIAN", "", "it.main", "") {
		t.Fatal("صلاحية الرؤية: نفسه نعم، الوحدة (عضو) لا")
	}
}
