package service

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/database"
	"staffmange-api/internal/repository"
)

// (دازاي 11-28) «الصلاحيات الي انطيهن للموظف هل يروحن عند اعاده تشغيل الدوكر؟»
// نحاكي إعادة التشغيل كاملة (الترحيلات + زرع الكتالوج + زرع الهيكل + إعادة مزامنة المناصب + الافتراضيات)
// ونتأكد: اليدوي يبقى، المشال بالإيد ما يرجع، صلاحيات المنصب تنضاف وتنسحب ويا المنصب، واليدوي ما ينسحب ويا المنصب.
func TestPermissionsSurviveRestartLive(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", dsn)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	clean := func() {
		for _, q := range []string{`DELETE FROM "OrgAssignment" WHERE "employeeId" LIKE 'pr-%'`, `DELETE FROM "OrgPermissionGrant" WHERE "employeeId" LIKE 'pr-%'`,
			`DELETE FROM "PermissionRevoke" WHERE "employeeId" LIKE 'pr-%'`, `DELETE FROM "PermissionAuditLog" WHERE "employeeId" LIKE 'pr-%'`,
			`DELETE FROM "EmployeePermission" WHERE "employeeId" LIKE 'pr-%'`, `DELETE FROM "Employee" WHERE id LIKE 'pr-%'`} {
			db.Exec(q)
		}
	}
	clean()
	t.Cleanup(clean)
	db.MustExec(`INSERT INTO "Employee"(id, name, role) VALUES ('pr-e', 'pr موظف', 'TECHNICIAN')`)
	grant := func(p string) {
		db.MustExec(`INSERT INTO "EmployeePermission"(id, "employeeId", "permissionId") SELECT gen_random_uuid()::text, 'pr-e', id FROM "Permission" WHERE name = $1 ON CONFLICT DO NOTHING`, p)
	}
	has := func(p string) bool {
		var ok bool
		db.Get(&ok, `SELECT EXISTS (SELECT 1 FROM "EmployeePermission" ep JOIN "Permission" p ON p.id = ep."permissionId" WHERE ep."employeeId" = 'pr-e' AND p.name = $1)`, p)
		return ok
	}
	permRepo := repository.NewPermissionRepository(db)
	org := NewOrgService(db)
	restart := func() {
		if err := database.Migrate(db, "", ""); err != nil {
			t.Fatal(err)
		}
		if err := permRepo.EnsureSeeded(); err != nil {
			t.Fatal(err)
		}
		if err := org.Seed(); err != nil {
			t.Fatal(err)
		}
		if err := org.resyncAll("اختبار"); err != nil {
			t.Fatal(err)
		}
		if _, err := NewPermissionService(permRepo, repository.NewEmployeeRepository(db)).ApplyDefaultsAll(); err != nil {
			t.Fatal(err)
		}
	}

	// ١) يدوي + المشال بالإيد
	grant("oversight_view")
	grant("company_costs")
	restart()
	if !has("expenses") {
		t.Fatal("صلاحية الدور الافتراضية (expenses) لازم تنضاف")
	}
	db.MustExec(`DELETE FROM "EmployeePermission" ep USING "Permission" p WHERE ep."permissionId" = p.id AND ep."employeeId" = 'pr-e' AND p.name = 'expenses'`)
	db.MustExec(`INSERT INTO "PermissionRevoke"("employeeId", permission) VALUES ('pr-e', 'expenses')`)
	restart()
	restart()
	if !has("oversight_view") || !has("company_costs") {
		t.Fatal("اليدوي راح بإعادة التشغيل")
	}
	if has("expenses") {
		t.Fatal("المشالة بالإيد رجعت بإعادة التشغيل")
	}

	// ٢) منصب: عضو بوحدة التدقيق — ياخذ حزمة العضو (مو صلاحيات الرئيس)
	db.MustExec(`INSERT INTO "OrgAssignment"("employeeId", "divisionId", "unitId", position) VALUES ('pr-e', 'ops', 'ops.audit', 'MEMBER')`)
	restart()
	if !has("auditing") || !has("leader_invoices_view") {
		t.Fatal("صلاحيات المنصب ما انضافت")
	}
	if has("audit_issues") {
		t.Fatal("صلاحية الرئيس انطت للعضو")
	}
	// يدوي فوگ المنصب: نفس صلاحية بالحزمة انطت باليد (المنح اليدوي يشيل علامة المنصب)
	db.MustExec(`DELETE FROM "OrgPermissionGrant" WHERE "employeeId" = 'pr-e' AND permission = 'auditing'`)
	// ينتهي المنصب
	db.MustExec(`UPDATE "OrgAssignment" SET "endedAt" = now() WHERE "employeeId" = 'pr-e'`)
	tx := db.MustBegin()
	if err := syncOrgPermissions(tx, "pr-e", "", "اختبار"); err != nil {
		t.Fatal(err)
	}
	tx.Commit()
	restart()
	if has("leader_invoices_view") {
		t.Fatal("صلاحية المنصب ما انسحبت بعد انتهاء المنصب")
	}
	if !has("auditing") || !has("oversight_view") {
		t.Fatal("اليدوي انسحب ويا المنصب")
	}
}

// (دازاي 11-28) «مدير سلاسل التوريد وبنفس الوقت نائب مدير الادارة»: صلاحياته تجي من المنصبين لحالها.
func TestTwoPositionsPermissionsLive(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", dsn)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	clean := func() {
		for _, q := range []string{`DELETE FROM "OrgAssignment" WHERE "employeeId" = 'pr-2'`, `DELETE FROM "OrgPermissionGrant" WHERE "employeeId" = 'pr-2'`,
			`DELETE FROM "PermissionAuditLog" WHERE "employeeId" = 'pr-2'`, `DELETE FROM "EmployeePermission" WHERE "employeeId" = 'pr-2'`, `DELETE FROM "Employee" WHERE id = 'pr-2'`} {
			db.Exec(q)
		}
	}
	clean()
	t.Cleanup(clean)
	org := NewOrgService(db)
	if err := org.Seed(); err != nil {
		t.Fatal(err)
	}
	db.MustExec(`INSERT INTO "Employee"(id, name, role) VALUES ('pr-2', 'pr مدير توريد ونائب إدارة', 'TECHNICIAN')`)
	db.MustExec(`INSERT INTO "OrgAssignment"("employeeId", "divisionId", position) VALUES ('pr-2', 'supply', 'DIRECTOR'), ('pr-2', 'admin', 'DEPUTY')`)
	tx := db.MustBegin()
	if err := syncOrgPermissions(tx, "pr-2", "", "اختبار"); err != nil {
		t.Fatal(err)
	}
	tx.Commit()
	repo := repository.NewPermissionRepository(db)
	full, _ := repo.Names("pr-2")
	view, _ := repo.ViewNames("pr-2")
	in := func(xs []string, x string) bool {
		for _, y := range xs {
			if y == x {
				return true
			}
		}
		return false
	}
	// مدير سلاسل التوريد: كامل (حتى صلاحيات الرئيس)
	for _, p := range []string{"procurement", "inventory", "suppliers_management", "unit_procurement"} {
		if !in(full, p) {
			t.Fatalf("مدير التوريد ناقصه %s: %v", p, full)
		}
	}
	// نائب الإدارة: تعديل العضو (إحصائيات، معالجة الحالات)، ومشاهدة التقييمات والتفاصيل والتدقيق
	for _, p := range []string{"employee_stats", "action_plan"} {
		if !in(full, p) {
			t.Fatalf("نائب الإدارة ناقصه %s: %v", p, full)
		}
	}
	for _, p := range []string{"mon_staff", "monitoring", "auditing", "staff_stats", "performance_review", "finance"} {
		if !in(view, p) && !in(full, p) {
			t.Fatalf("نائب الإدارة ما يشوف %s: %v", p, view)
		}
	}
	// موافقة المدير ما تنطى للنائب كاملة
	if in(full, "mon_justifications") || in(full, "booking_delete_approve") {
		t.Fatalf("النائب أخذ صلاحية موافقة كاملة: %v", full)
	}
	if in(view, "finance_audit") {
		t.Fatal("صلاحية مالك انطت من المنصب")
	}
}
