package service

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/repository"
)

// TestOrgEffectiveRoles_Live الدور من المنصب (دازاي 11-01): فني بوحدة الحسابات ياخذ دور
// المحاسب الفعلي، وإشعارات «المحاسب» توصله، وبلا منصب ما ياخذ شي.
func TestOrgEffectiveRoles_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة — تخطي اختبار القاعدة الحية")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()
	const emp = "__orgrole_tech"
	cleanup := func() {
		db.Exec(`DELETE FROM "Notification" WHERE "employeeId" = $1`, emp)
		db.Exec(`DELETE FROM "OrgAssignment" WHERE "employeeId" = $1`, emp)
		db.Exec(`DELETE FROM "OrgPermissionGrant" WHERE "employeeId" = $1`, emp)
		db.Exec(`DELETE FROM "EmployeePermission" WHERE "employeeId" = $1`, emp)
		db.Exec(`DELETE FROM "PermissionAuditLog" WHERE "employeeId" = $1`, emp)
		db.Exec(`DELETE FROM "Employee" WHERE id = $1`, emp)
	}
	cleanup()
	defer cleanup()
	db.MustExec(`INSERT INTO "Employee"(id, name, status, role) VALUES ($1, 'فني أدوار', 'ACTIVE', 'TECHNICIAN')`, emp)
	s := NewOrgService(db)
	_ = s.Seed()
	if r := repository.OrgEffectiveRoles(db, emp); len(r) != 0 {
		t.Fatalf("بلا منصب ما ياخذ دور: %v", r)
	}
	if _, err := s.Assign("", OrgAssignInput{EmployeeID: emp, UnitID: "admin.accounts", Position: "MEMBER"}); err != nil {
		t.Fatal(err)
	}
	if !repository.HasOrgRole(db, emp, "FINANCE") || repository.HasOrgRole(db, emp, "ADMIN", "OWNER", "MONITOR") {
		t.Fatalf("الأدوار غلط: %v", repository.OrgEffectiveRoles(db, emp))
	}
	notif := repository.NewNotificationRepository(db)
	_ = notif.CreateForRole("FINANCE", "test_orgrole", "اختبار")
	_ = notif.CreateForRolesOrPermission([]string{"FINANCE"}, "", "test_orgrole2", "اختبار")
	var n int
	db.Get(&n, `SELECT count(*) FROM "Notification" WHERE "employeeId" = $1 AND type IN ('test_orgrole', 'test_orgrole2')`, emp)
	if n != 2 {
		t.Fatalf("إشعارات المحاسب ما وصلت لصاحب المنصب: %d", n)
	}
	db.Exec(`DELETE FROM "Notification" WHERE type IN ('test_orgrole', 'test_orgrole2')`)
	var hasRemote int
	db.Get(&hasRemote, `SELECT count(*) FROM "EmployeePermission" ep JOIN "Permission" p ON p.id = ep."permissionId" WHERE ep."employeeId" = $1 AND p.name = 'remote_hours_self'`, emp)
	if hasRemote != 0 {
		t.Fatal("العمل من البيت ما ينعطي بالمنصب — بس يدوي")
	}
}
