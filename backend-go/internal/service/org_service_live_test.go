package service

import (
	"os"
	"sort"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// TestOrgPermissionSync_Live الهيكل الجديد (دازاي 10-29): المنصب ينطي حزمته، والإنهاء
// يسحب بس الي انطت بسببه — اليدوي والقالب والاستثناء ما ينمسون، وشيل الأدمن يتحترم.
func TestOrgPermissionSync_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة — تخطي اختبار القاعدة الحية")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()
	const emp, susp = "__org_emp", "__org_susp"
	cleanup := func() {
		db.Exec(`DELETE FROM "OrgAssignment" WHERE "employeeId" IN ($1, $2)`, emp, susp)
		db.Exec(`DELETE FROM "OrgPermissionGrant" WHERE "employeeId" IN ($1, $2)`, emp, susp)
		db.Exec(`DELETE FROM "PermissionException" WHERE "employeeId" IN ($1, $2)`, emp, susp)
		db.Exec(`DELETE FROM "EmployeePermission" WHERE "employeeId" IN ($1, $2)`, emp, susp)
		db.Exec(`DELETE FROM "PermissionAuditLog" WHERE "employeeId" IN ($1, $2)`, emp, susp)
		db.Exec(`DELETE FROM "Employee" WHERE id IN ($1, $2)`, emp, susp)
	}
	cleanup()
	defer cleanup()
	db.MustExec(`INSERT INTO "Employee"(id, name, status, role) VALUES ($1, 'تجربة هيكل', 'ACTIVE', 'TECHNICIAN'), ($2, 'موقوف', 'SUSPENDED', 'TECHNICIAN')`, emp, susp)
	s := NewOrgService(db)
	if err := s.Seed(); err != nil {
		t.Fatal(err)
	}
	perms := func() []string {
		var out []string
		db.Select(&out, `SELECT p.name FROM "EmployeePermission" ep JOIN "Permission" p ON p.id = ep."permissionId" WHERE ep."employeeId" = $1 ORDER BY 1`, emp)
		return out
	}
	has := func(n string) bool {
		for _, p := range perms() {
			if p == n {
				return true
			}
		}
		return false
	}
	give := func(n string) {
		db.MustExec(`INSERT INTO "EmployeePermission"(id, "employeeId", "permissionId") SELECT gen_random_uuid()::text, $1, id FROM "Permission" WHERE name = $2 ON CONFLICT DO NOTHING`, emp, n)
	}
	// قبل: صلاحية يدوية من ضمن حزمة العُدد
	give("vehicle_management")
	before := perms()

	if _, err := s.Assign("", OrgAssignInput{EmployeeID: susp, UnitID: "ops.fleet", Position: "MEMBER"}); err == nil {
		t.Fatal("الموقوف ما ينعيّن")
	}
	a, err := s.Assign("", OrgAssignInput{EmployeeID: emp, UnitID: "ops.fleet", Position: "DEPUTY", AllowanceIQD: 150000})
	if err != nil {
		t.Fatal(err)
	}
	if !has("tool_requests_approve") || !has("unit_fleet") {
		t.Fatalf("الحزمة ما انطت: %v", perms())
	}
	var viaOrgVM int
	db.Get(&viaOrgVM, `SELECT count(*) FROM "OrgPermissionGrant" WHERE "employeeId" = $1 AND permission = 'vehicle_management'`, emp)
	if viaOrgVM != 0 {
		t.Fatal("اليدوية السابقة انحسبت من المنصب")
	}
	// نائب ثاني بنفس المكان ينرفض
	if _, err := s.Assign("", OrgAssignInput{EmployeeID: emp, UnitID: "ops.fleet", Position: "DEPUTY"}); err == nil {
		t.Fatal("تكرار بنفس المكان")
	}
	// الأدمن يشيل tool_requests بيده — المزامنة ما ترجّعها (تعيين ثاني يشغّل المزامنة)
	db.MustExec(`DELETE FROM "EmployeePermission" WHERE "employeeId" = $1 AND "permissionId" = (SELECT id FROM "Permission" WHERE name = 'tool_requests')`, emp)
	b, err := s.Assign("", OrgAssignInput{EmployeeID: emp, UnitID: "it.main", Position: "MEMBER"})
	if err != nil {
		t.Fatal(err)
	}
	if has("tool_requests") {
		t.Fatal("المزامنة رجّعت صلاحية شالها الأدمن")
	}
	// استثناء مثبّت على صلاحية من المنصب (عضو IT ياخذ unit_it، مو it_assets الي للرئيس) — الإنهاء ما يشيلها
	db.MustExec(`INSERT INTO "PermissionException"("employeeId", permission, reason) VALUES ($1, 'unit_it', 'تجربة')`, emp)
	if _, err := s.End("", b.ID, "تجربة إنهاء"); err != nil {
		t.Fatal(err)
	}
	if !has("unit_it") {
		t.Fatal("الاستثناء المثبّت انسحب")
	}
	if has("it_stats") {
		t.Fatal("صلاحية المنصب المنتهي ما انسحبت")
	}
	// إنهاء المنصب الأول يرجّعه لحاله + الاستثناء
	if _, err := s.End("", a.ID, "تجربة"); err != nil {
		t.Fatal(err)
	}
	after := perms()
	want := append(append([]string{}, before...), "unit_it")
	sort.Strings(want)
	if len(after) != len(want) {
		t.Fatalf("بعد الإنهاء %v والمتوقع %v", after, want)
	}
	for i := range want {
		if after[i] != want[i] {
			t.Fatalf("بعد الإنهاء %v والمتوقع %v", after, want)
		}
	}
	// التراجع يرجّع المنصب وحزمته
	if err := s.Restore("", a.ID); err != nil {
		t.Fatal(err)
	}
	if !has("unit_fleet") {
		t.Fatal("التراجع ما رجّع الحزمة")
	}
	mine, _ := s.Mine(emp)
	if len(mine) != 1 || mine[0].DivisionName == "" || mine[0].UnitName == nil {
		t.Fatalf("مكاني بلا أسماء: %+v", mine)
	}
}
