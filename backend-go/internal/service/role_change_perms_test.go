package service

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/database"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// (ل) 11-22 دازاي: تغيير الدور يشيل صلاحيات الدور القديم
func TestSameRoles(t *testing.T) {
	if !sameRoles([]string{"A", "B"}, []string{"B", "A"}) || sameRoles([]string{"A"}, []string{"B"}) || sameRoles([]string{"A"}, []string{"A", "B"}) {
		t.Fatal("sameRoles غلط")
	}
}

// (ل) 11-22: مراقب ينقلب مصمم → صلاحيات المراقب تنشال، اليدوية تبقى؛ والمصممة تشوف سيتها
func TestRoleChangeDropsOldDefaults_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db := sqlx.MustConnect("postgres", dsn)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	const e = "__rc_emp"
	clean := func() {
		db.Exec(`DELETE FROM "PermissionAuditLog" WHERE "employeeId" = $1`, e) //nolint:errcheck
		db.Exec(`DELETE FROM "DesignSet" WHERE "createdById" = $1`, e)         //nolint:errcheck
		db.Exec(`DELETE FROM "Employee" WHERE id = $1`, e)                     //nolint:errcheck
	}
	clean()
	t.Cleanup(clean)
	db.MustExec(`INSERT INTO "Employee"(id, name, status, role) VALUES ($1,'نقل دور','ACTIVE','MONITOR')`, e)
	ps := NewPermissionService(repository.NewPermissionRepository(db), repository.NewEmployeeRepository(db))
	if _, err := ps.ApplyDefaults(e); err != nil {
		t.Fatal(err)
	}
	db.MustExec(`INSERT INTO "EmployeePermission"(id,"employeeId","permissionId") SELECT gen_random_uuid()::text,$1,id FROM "Permission" WHERE name='design_gallery' ON CONFLICT DO NOTHING`, e)
	es := NewEmployeeService(repository.NewEmployeeRepository(db))
	r := "DESIGNER"
	if _, err := es.Update(e, model.UpdateEmployeeRequest{Role: &r}); err != nil {
		t.Fatal(err)
	}
	if _, err := ps.ApplyDefaults(e); err != nil {
		t.Fatal(err)
	}
	got := []string{}
	db.Select(&got, `SELECT p.name FROM "EmployeePermission" ep JOIN "Permission" p ON p.id = ep."permissionId" WHERE ep."employeeId" = $1 ORDER BY 1`, e) //nolint:errcheck
	want := map[string]bool{}
	for _, n := range model.RoleDefaultPermissions["DESIGNER"] {
		want[n] = true
	}
	for _, n := range got {
		if !want[n] {
			t.Fatalf("بقت صلاحية من الدور القديم: %s (الكل %v)", n, got)
		}
	}
	// سيت سوّته هي وما مكلّفة بقسم منه → يطلعلها
	db.MustExec(`INSERT INTO "DesignSet"(id, title, "createdById") VALUES ('__rc_set','سيت نرجس التجريبي',$1)`, e)
	ds := NewDesignSetService(db, repository.NewExtraTaskRepository(db), repository.NewNotificationRepository(db))
	rows, err := ds.List(e, "DESIGNER", got)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, x := range rows {
		found = found || x.ID == "__rc_set"
	}
	if !found {
		t.Fatal("السيت الي سوّته ما طلعلها")
	}
	// المصممة بلا تسجيل بالهيكل تشوف كل السيتات وتشتغل على سيت المخططات
	if !ds.CanSeeAll(e, "DESIGNER", nil) || !ds.CanSeeAll(e, "TECHNICIAN", []string{"unit_design"}) || ds.CanSeeAll(e, "TECHNICIAN", nil) {
		t.Fatal("شوف المصمم غلط")
	}
	pl := NewPlanSetService(db, ds)
	if !pl.CanSee(e, "DESIGNER", nil) || !pl.CanManage(e, "DESIGNER", nil) {
		t.Fatal("المصممة ما تشوف سيت المخططات")
	}
}
