package service

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
)

// TestOrgActing_Live المبدأ ٢٠: المدير بإجازة موافق عليها ← نائبه «مدير بالوكالة» بصلاحياته،
// ما يتكرر، والإجازة تنلغى ← الوكالة تنتهي والنائب يبقى نائب.
func TestOrgActing_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	const dir, dep = "__act_dir", "__act_dep"
	emps := []string{dir, dep}
	cleanup := func() {
		db.Exec(`DELETE FROM "LeaveRequest" WHERE "employeeId" = ANY($1)`, emps)
		db.Exec(`DELETE FROM "OrgAssignment" WHERE "employeeId" = ANY($1)`, emps)
		db.Exec(`DELETE FROM "OrgPermissionGrant" WHERE "employeeId" = ANY($1)`, emps)
		db.Exec(`DELETE FROM "EmployeePermission" WHERE "employeeId" = ANY($1)`, emps)
		db.Exec(`DELETE FROM "PermissionAuditLog" WHERE "employeeId" = ANY($1)`, emps)
		db.Exec(`DELETE FROM "OrgGovernorState" WHERE key LIKE 'ACTING_NONE:%'`)
		db.Exec(`DELETE FROM "Employee" WHERE id = ANY($1)`, emps)
	}
	cleanup()
	defer cleanup()
	db.MustExec(`INSERT INTO "Employee"(id, name, status, role) VALUES ($1,'مدير تصميم','ACTIVE','TECHNICIAN'),($2,'نائب تصميم','ACTIVE','TECHNICIAN')`, dir, dep)
	s := NewOrgService(db)
	if err := s.Seed(); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Assign("", OrgAssignInput{EmployeeID: dir, UnitID: "eng.design", Position: "DIRECTOR"}); err != nil {
		t.Skipf("المكان مشغول محلياً: %v", err)
	}
	if _, err := s.Assign("", OrgAssignInput{EmployeeID: dep, UnitID: "eng.design", Position: "DEPUTY"}); err != nil {
		t.Skipf("المكان مشغول محلياً: %v", err)
	}
	db.MustExec(`INSERT INTO "LeaveRequest"(id, "employeeId", "startDate", "endDate", route, status)
		VALUES ('__act_leave', $1, current_date - 1, current_date + 2, 'ADMIN', 'APPROVED')`, dir)
	notes := 0
	s.RunActing(func(string, string) { notes++ })
	acting := func() int {
		var n int
		db.Get(&n, `SELECT count(*) FROM "OrgAssignment" WHERE "employeeId" = $1 AND acting AND position = 'DIRECTOR' AND "endedAt" IS NULL`, dep)
		return n
	}
	if acting() != 1 || notes != 2 {
		t.Fatalf("الوكالة: %d، الإشعارات %d", acting(), notes)
	}
	s.RunActing(func(string, string) {})
	if acting() != 1 {
		t.Fatal("الوكالة تكررت")
	}
	// الإجازة انلغت ← الوكالة تنتهي والنائب يبقى نائب
	db.MustExec(`UPDATE "LeaveRequest" SET status = 'CANCELLED' WHERE id = '__act_leave'`)
	s.RunActing(func(string, string) {})
	if acting() != 0 {
		t.Fatal("الوكالة ما انتهت بعد إلغاء الإجازة")
	}
	mine, _ := s.Mine(dep)
	if len(mine) != 1 || mine[0].Position != "DEPUTY" {
		t.Fatalf("النائب لازم يبقى نائب: %+v", mine)
	}
	// بلا نائب ← تنبيه للمشرفين (مرة باليوم)
	db.MustExec(`UPDATE "LeaveRequest" SET status = 'APPROVED' WHERE id = '__act_leave'`)
	db.MustExec(`UPDATE "OrgAssignment" SET "endedAt" = now() WHERE "employeeId" = $1`, dep)
	got := 0
	s.RunActing(func(string, string) { got++ })
	s.RunActing(func(string, string) { got += 100 })
	if got == 0 || got >= 100 {
		t.Fatalf("تنبيه «ماكو نائب» لازم مرة وحدة: %d", got)
	}
}
