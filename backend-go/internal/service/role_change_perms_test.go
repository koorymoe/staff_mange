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

// (ل) 11-22: الخانات الإضافية بالفورما بس بصلاحية plan_sheet_fields
func TestPlanSheetExtra_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db := sqlx.MustConnect("postgres", dsn)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	const e = "__px_emp"
	clean := func() {
		db.Exec(`DELETE FROM "PlanSet" WHERE "createdById" = $1`, e) //nolint:errcheck
		db.Exec(`DELETE FROM "Employee" WHERE id = $1`, e)           //nolint:errcheck
	}
	clean()
	t.Cleanup(clean)
	db.MustExec(`INSERT INTO "Employee"(id, name, status, role) VALUES ($1,'مصممة','ACTIVE','DESIGNER')`, e)
	db.MustExec(`INSERT INTO "PlanSet"(id, title, "createdById") VALUES ('__px_set','سيت',$1)`, e)
	var fid string
	if err := db.Get(&fid, `INSERT INTO "PlanFile"("setId","typeCode",title,"fileKey","fileType","uploadedById") VALUES ('__px_set','1.1','x','k','image/png',$1) RETURNING id`, e); err != nil {
		t.Fatal(err)
	}
	ds := NewDesignSetService(db, repository.NewExtraTaskRepository(db), repository.NewNotificationRepository(db))
	pl := NewPlanSetService(db, ds)
	ex := []SheetExtraField{{Label: "المهندس المشرف", Value: "م. علي", After: "package"}, {Label: "x", After: "زبالة"}}
	if err := pl.UpdateSheet(e, "DESIGNER", nil, fid, PlanSheetInput{Name: "لوحة", Extra: &ex}); err != ErrDesignHidden {
		t.Fatalf("بلا صلاحية لازم ينرفض: %v", err)
	}
	if err := pl.UpdateSheet(e, "DESIGNER", []string{"plan_sheet_fields"}, fid, PlanSheetInput{Name: "لوحة", Extra: &ex}); err != nil {
		t.Fatal(err)
	}
	var raw string
	db.Get(&raw, `SELECT "sheetExtra"::text FROM "PlanFile" WHERE id = $1`, fid) //nolint:errcheck
	if raw != `[{"after": "package", "label": "المهندس المشرف", "value": "م. علي"}, {"after": "notes", "label": "x", "value": ""}]` {
		t.Fatalf("انحفظ غلط: %s", raw)
	}
	// حفظ بلا extra (مصممة بلا صلاحية) ما يمسح الخانات
	if err := pl.UpdateSheet(e, "DESIGNER", nil, fid, PlanSheetInput{Name: "لوحة 2"}); err != nil {
		t.Fatal(err)
	}
	var n int
	db.Get(&n, `SELECT jsonb_array_length("sheetExtra") FROM "PlanFile" WHERE id = $1`, fid) //nolint:errcheck
	if n != 2 {
		t.Fatalf("الخانات انمسحت: %d", n)
	}
}
