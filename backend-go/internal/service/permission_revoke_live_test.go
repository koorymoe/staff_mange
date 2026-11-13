package service

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/database"
	"staffmange-api/internal/repository"
)

// (دازاي 11-13) «جاي اشيل من عدها صلاحيات مو شغلها بس فجأة رجعن» — الدورة كل ٦ ساعات
// چانت ترجّع صلاحيات الدور (والأدوار الثانوية). هسه الي انشال بقصد ما يرجع، والمنح
// اليدوي بعدين يرجّعه طبيعي.
func TestRevokedPermissionStaysRevoked_Live(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	emp := "rv-des"
	clean := func() {
		db.Exec(`DELETE FROM "PermissionRevoke" WHERE "employeeId" = $1`, emp)
		db.Exec(`DELETE FROM "EmployeePermission" WHERE "employeeId" = $1`, emp)
		db.Exec(`DELETE FROM "Employee" WHERE id = $1`, emp)
	}
	clean()
	t.Cleanup(clean)
	// مصممة ومأشّر عليها دور ثانوي (الكوادر) — مصدر الصلاحيات الي مو شغلها
	db.MustExec(`INSERT INTO "Employee"(id, name, role, status, "secondaryRoles") VALUES ($1, 'مصممة', 'DESIGNER', 'ACTIVE', '{HR_COORDINATOR}')`, emp)
	perms := repository.NewPermissionRepository(db)
	svc := NewPermissionService(perms, repository.NewEmployeeRepository(db))
	has := func(name string) bool {
		var n int
		db.Get(&n, `SELECT count(*) FROM "EmployeePermission" ep JOIN "Permission" p ON p.id = ep."permissionId" WHERE ep."employeeId" = $1 AND p.name = $2`, emp, name)
		return n == 1
	}
	if _, err := svc.ApplyDefaults(emp); err != nil {
		t.Fatal(err)
	}
	if !has("staff_management") || !has("design_forms") {
		t.Fatal("أول مرة تاخذ صلاحيات أدوارها")
	}
	// الأدمن يحفظ صلاحياتها بلا staff_management
	ids := []string{}
	db.Select(&ids, `SELECT ep."permissionId" FROM "EmployeePermission" ep JOIN "Permission" p ON p.id = ep."permissionId" WHERE ep."employeeId" = $1 AND p.name <> 'staff_management'`, emp)
	if _, err := svc.SetForEmployee(emp, ids); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ApplyDefaultsAll(); err != nil {
		t.Fatal(err)
	}
	if has("staff_management") {
		t.Fatal("الدورة رجّعت الصلاحية الي انشالت بقصد")
	}
	// المنح اليدوي يرجّعها ويمسح المنع
	var sm string
	db.Get(&sm, `SELECT id FROM "Permission" WHERE name = 'staff_management'`)
	if _, err := svc.SetForEmployee(emp, append(ids, sm)); err != nil {
		t.Fatal(err)
	}
	var n int
	db.Get(&n, `SELECT count(*) FROM "PermissionRevoke" WHERE "employeeId" = $1`, emp)
	if !has("staff_management") || n != 0 {
		t.Fatalf("المنح اليدوي يرجّعها ويمسح المنع: %v %d", has("staff_management"), n)
	}
}
