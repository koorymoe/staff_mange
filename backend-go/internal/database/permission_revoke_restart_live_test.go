package database

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// (دازاي 11-15) «شلت منها الصلاحيات وثاني يوم يرجعن» — كل تشغيل للسيرفر يعيد منح صلاحيات
// الأدوار (grantRolePermission…). المشالة بالإيد لازم تبقى مشالة بعد إعادة التشغيل.
func TestRevokedPermissionSurvivesRestartLive(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	clean := func() {
		db.Exec(`DELETE FROM "PermissionRevoke" WHERE "employeeId" IN ('rr-tech', 'rr-lead')`)
		db.Exec(`DELETE FROM "EmployeePermission" WHERE "employeeId" IN ('rr-tech', 'rr-lead')`)
		db.Exec(`DELETE FROM "Employee" WHERE id IN ('rr-tech', 'rr-lead')`)
	}
	clean()
	t.Cleanup(clean)
	db.MustExec(`INSERT INTO "Employee"(id, name, role, "isLeader") VALUES ('rr-tech', 'rr فني', 'TECHNICIAN', false), ('rr-lead', 'rr ليدر', 'TECHNICIAN', true)`)
	db.MustExec(`INSERT INTO "PermissionRevoke"("employeeId", permission) VALUES ('rr-tech', 'procurement'), ('rr-lead', 'leader_basket'), ('rr-lead', 'procurement_manual')`)
	if err := Migrate(db, "", ""); err != nil { // = إعادة تشغيل
		t.Fatal(err)
	}
	has := func(emp, perm string) bool {
		var ok bool
		db.Get(&ok, `SELECT EXISTS (SELECT 1 FROM "EmployeePermission" ep JOIN "Permission" p ON p.id = ep."permissionId" WHERE ep."employeeId" = $1 AND p.name = $2)`, emp, perm)
		return ok
	}
	if has("rr-tech", "procurement") || has("rr-lead", "leader_basket") || has("rr-lead", "procurement_manual") {
		t.Fatal("صلاحية مشالة رجعت بعد إعادة التشغيل")
	}
	if !has("rr-tech", "procurement_customer") || !has("rr-lead", "procurement") {
		t.Fatal("الي ما انشالت لازم تبقى تنمنح حسب الدور")
	}
}
