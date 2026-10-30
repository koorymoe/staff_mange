package service

import (
	"os"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
)

// TestOrgAIContext_Live النقطة ٢: الموجز بيه الإدارات واللائحة ومكان المستخدم بلا أسماء،
// والأدوات تشتغل والأسماء تنخفى لماتركس، والموظف ما ياخذ unit_status.
func TestOrgAIContext_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	const emp = "__ai_ctx_emp"
	cleanup := func() {
		db.Exec(`DELETE FROM "OrgAssignment" WHERE "employeeId" = $1`, emp)
		db.Exec(`DELETE FROM "OrgPermissionGrant" WHERE "employeeId" = $1`, emp)
		db.Exec(`DELETE FROM "EmployeePermission" WHERE "employeeId" = $1`, emp)
		db.Exec(`DELETE FROM "PermissionAuditLog" WHERE "employeeId" = $1`, emp)
		db.Exec(`DELETE FROM "Employee" WHERE id = $1`, emp)
	}
	cleanup()
	defer cleanup()
	db.MustExec(`INSERT INTO "Employee"(id, name, status, role) VALUES ($1, 'زيد الحكيم التجريبي', 'ACTIVE', 'TECHNICIAN')`, emp)
	s := NewOrgService(db)
	_ = s.Seed()
	if _, err := s.Assign("", OrgAssignInput{EmployeeID: emp, UnitID: "supply.procurement", Position: "MEMBER"}); err != nil {
		t.Fatal(err)
	}
	brief := OrgBrief(db, emp)
	for _, want := range []string{"العمليات", "سلاسل التوريد", "التقنية IT", "م2/ف1 الدقة والسرعة", "م5/ف4", "عضو — سلاسل التوريد / المشتريات والمخزن", "بالوكالة"} {
		if !strings.Contains(brief, want) {
			t.Errorf("الموجز ناقص «%s»", want)
		}
	}
	if strings.Contains(brief, "زيد الحكيم") {
		t.Fatal("الموجز بيه اسم")
	}
	a := NewMatrixAgent(db, nil, "")
	a.registerOrgTools()
	run := func(name string, in map[string]any) string {
		for _, tl := range a.tools {
			if tl.Name == name {
				out, err := tl.Run(in, a.newMask())
				if err != nil {
					t.Fatalf("%s: %v", name, err)
				}
				return out
			}
		}
		t.Fatalf("ماكو أداة %s", name)
		return ""
	}
	st := run("org_structure", nil)
	if strings.Contains(st, "زيد الحكيم") || !strings.Contains(st, "موظف#") || !strings.Contains(st, "الكراسي الفاضية") {
		t.Fatalf("org_structure لازم بلا أسماء وبيه الكراسي الفاضية:\n%s", st)
	}
	if r := run("regulations", map[string]any{"article": float64(3)}); !strings.Contains(r, "م3/ف1") || strings.Contains(r, "م2/ف1") {
		t.Fatalf("regulations بالمادة:\n%s", r)
	}
	if u := run("unit_status", map[string]any{"division": "supply"}); !strings.Contains(u, "المشاكل المفتوحة") || !strings.Contains(u, "قرارات ماتركس الحاكم") {
		t.Fatalf("unit_status:\n%s", u)
	}
	f := MatrixToolFilter("TECHNICIAN", []string{"agent_matrix"})
	if !f("org_structure") || !f("regulations") || f("unit_status") {
		t.Fatal("صلاحيات الأدوات للموظف غلط")
	}
}
