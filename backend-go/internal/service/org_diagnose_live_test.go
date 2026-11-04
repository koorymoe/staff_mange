package service

import (
	"os"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
)

// TestOrgDiagnosis_Live النقطة ٤: شكوى فاتت ٤٨ ساعة ← ماتركس يفتح مشكلة بوحدة الجودة
// مربوطة بـم٤/ف٣ ويبلّغ رئيس الوحدة؛ الجولة الثانية تحدّث ما تكرر؛ الإشارة تختفي ← تتسكّر.
func TestOrgDiagnosis_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	const head, cust, cmp = "__dg_head", "__dg_cust", "__dg_cmp"
	var openLate int
	db.Get(&openLate, `SELECT count(*) FROM "Complaint" WHERE status IN ('NEW','IN_PROGRESS') AND "createdAt" < now() - interval '48 hours' AND id <> $1`, cmp)
	if openLate > 0 {
		t.Skip("القاعدة المحلية بيها شكاوي متأخرة — الإشارة مو معزولة")
	}
	cleanup := func() {
		db.Exec(`DELETE FROM "OrgIssueReport" WHERE signal = 'complaints_late'`)
		db.Exec(`DELETE FROM "Complaint" WHERE id = $1`, cmp)
		db.Exec(`DELETE FROM "Customer" WHERE id = $1`, cust)
		db.Exec(`DELETE FROM "OrgAssignment" WHERE "employeeId" = $1`, head)
		db.Exec(`DELETE FROM "OrgPermissionGrant" WHERE "employeeId" = $1`, head)
		db.Exec(`DELETE FROM "EmployeePermission" WHERE "employeeId" = $1`, head)
		db.Exec(`DELETE FROM "PermissionAuditLog" WHERE "employeeId" = $1`, head)
		db.Exec(`DELETE FROM "Employee" WHERE id = $1`, head)
	}
	cleanup()
	defer cleanup()
	db.MustExec(`INSERT INTO "Employee"(id, name, status, role) VALUES ($1, 'رئيس جودة', 'ACTIVE', 'TECHNICIAN')`, head)
	db.MustExec(`INSERT INTO "Customer"(id, name, phone) VALUES ($1, 'زبون تشخيص', '07700000098')`, cust)
	s := NewOrgService(db)
	_ = s.Seed()
	if _, err := s.Assign("", OrgAssignInput{EmployeeID: head, UnitID: "quality.followup", Position: "DIRECTOR"}); err != nil {
		t.Skipf("المكان مشغول محلياً: %v", err)
	}
	db.MustExec(`INSERT INTO "Complaint"(id, "customerId", description, status, "createdByEmployeeId", "createdAt") VALUES ($1, $2, 'x', 'NEW', $3, now() - interval '60 hours')`, cmp, cust, head)
	got := []string{}
	s.RunDiagnosis(nil, func(emp, msg string) {
		if emp == head {
			got = append(got, msg)
		}
	})
	var issue struct {
		ID        string `db:"id"`
		Unit      string `db:"unitId"`
		Principle int    `db:"principle"`
		Source    string `db:"source"`
		Problem   string `db:"problem"`
	}
	if err := db.Get(&issue, `SELECT id, "unitId", principle, source, problem FROM "OrgIssueReport" WHERE signal = 'complaints_late' AND status <> 'RESOLVED'`); err != nil {
		t.Fatalf("ماتركس ما فتح المشكلة: %v", err)
	}
	if issue.Unit != "quality.followup" || issue.Principle != 16 || issue.Source != "MATRIX" || !strings.Contains(issue.Problem, "1 شكوى") {
		t.Fatalf("المشكلة: %+v", issue)
	}
	if len(got) != 1 || !strings.Contains(got[0], "م4/ف3") {
		t.Fatalf("تبليغ رئيس الوحدة: %v", got)
	}
	s.RunDiagnosis(nil, func(string, string) { t.Fatal("الجولة الثانية بلّغت مرة ثانية") })
	var n int
	db.Get(&n, `SELECT count(*) FROM "OrgIssueReport" WHERE signal = 'complaints_late'`)
	if n != 1 {
		t.Fatalf("تكررت: %d", n)
	}
	// انحلت الشكوى والإشارة صار إلها ٣ أيام مختفية ← تتسكّر
	db.MustExec(`UPDATE "Complaint" SET status = 'RESOLVED', "resolvedAt" = now() WHERE id = $1`, cmp)
	db.MustExec(`UPDATE "OrgIssueReport" SET "lastSeenAt" = now() - interval '4 days' WHERE id = $1`, issue.ID)
	_, closed := s.RunDiagnosis(nil, nil)
	var st string
	db.Get(&st, `SELECT status FROM "OrgIssueReport" WHERE id = $1`, issue.ID)
	if closed < 1 || st != "RESOLVED" {
		t.Fatalf("ما تسكّرت: %d %s", closed, st)
	}
}
