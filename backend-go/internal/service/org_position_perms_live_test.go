package service

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// TestOrgPositionPerms_Live الصلاحيات حسب المنصب (دازاي 10-30): عضو الوحدة ما ياخذ
// صلاحيات الرئيس، ورئيسها ياخذها، ومدير الإدارة ياخذ حزم الرؤساء لكل وحداتها.
func TestOrgPositionPerms_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة — تخطي اختبار القاعدة الحية")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()
	ids := []string{"__pp_mem", "__pp_head", "__pp_dir"}
	cleanup := func() {
		for _, id := range ids {
			db.Exec(`DELETE FROM "OrgAssignment" WHERE "employeeId" = $1`, id)
			db.Exec(`DELETE FROM "OrgPermissionGrant" WHERE "employeeId" = $1`, id)
			db.Exec(`DELETE FROM "EmployeePermission" WHERE "employeeId" = $1`, id)
			db.Exec(`DELETE FROM "PermissionAuditLog" WHERE "employeeId" = $1`, id)
			db.Exec(`DELETE FROM "Employee" WHERE id = $1`, id)
		}
	}
	cleanup()
	defer cleanup()
	for _, id := range ids {
		db.MustExec(`INSERT INTO "Employee"(id, name, status, role) VALUES ($1, 'تجربة منصب', 'ACTIVE', 'TECHNICIAN')`, id)
	}
	s := NewOrgService(db)
	if err := s.Seed(); err != nil {
		t.Fatal(err)
	}
	has := func(emp, perm string) bool {
		var n int
		db.Get(&n, `SELECT count(*) FROM "EmployeePermission" ep JOIN "Permission" p ON p.id = ep."permissionId"
			WHERE ep."employeeId" = $1 AND p.name = $2`, emp, perm)
		return n > 0
	}
	if _, err := s.Assign("", OrgAssignInput{EmployeeID: "__pp_mem", UnitID: "ops.staff", Position: "MEMBER"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Assign("", OrgAssignInput{EmployeeID: "__pp_head", UnitID: "ops.staff", Position: "DIRECTOR"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Assign("", OrgAssignInput{EmployeeID: "__pp_dir", DivisionID: "ops", Position: "DIRECTOR"}); err != nil {
		t.Fatal(err)
	}
	if !has("__pp_mem", "staff_requests") || !has("__pp_mem", "unit_hr") {
		t.Fatal("العضو لازم ياخذ الشغل اليومي")
	}
	for _, p := range []string{"leave_approve_morning", "kpi_management", "staff_management"} {
		if has("__pp_mem", p) {
			t.Fatalf("العضو أخذ صلاحية رئيس: %s", p)
		}
		if !has("__pp_head", p) {
			t.Fatalf("الرئيس ما أخذ: %s", p)
		}
		if !has("__pp_dir", p) {
			t.Fatalf("مدير الإدارة ما أخذ: %s", p)
		}
	}
	if !has("__pp_dir", "vehicle_management") || !has("__pp_dir", "manage_customers") {
		t.Fatal("مدير الإدارة لازم ياخذ حزم رؤساء كل وحداته")
	}
	// إعادة البذر ما تكرر شي ولا تغيّر
	if err := s.Seed(); err != nil {
		t.Fatal(err)
	}
	if has("__pp_mem", "kpi_management") {
		t.Fatal("إعادة البذر خربت التقسيم")
	}
}
