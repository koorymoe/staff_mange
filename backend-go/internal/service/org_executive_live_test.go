package service

import (
	"os"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
)

// TestExecutiveOpinion_Live النقطة ٦: حقائق الموظف من النظام، أدوات كمال تشتغل بالأسماء،
// والرأي ينحفظ ويتحدّث (واحد بالشهر) ويرجع بالسجل.
func TestExecutiveOpinion_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	const emp, boss = "__ex_emp", "__ex_boss"
	cleanup := func() {
		db.Exec(`DELETE FROM "OrgExecutiveOpinion" WHERE "employeeId" = $1`, emp)
		db.Exec(`DELETE FROM "OrgAssignment" WHERE "employeeId" = $1`, emp)
		db.Exec(`DELETE FROM "OrgPermissionGrant" WHERE "employeeId" = $1`, emp)
		db.Exec(`DELETE FROM "EmployeePermission" WHERE "employeeId" = $1`, emp)
		db.Exec(`DELETE FROM "PermissionAuditLog" WHERE "employeeId" = $1`, emp)
		db.Exec(`DELETE FROM "Employee" WHERE id IN ($1, $2)`, emp, boss)
	}
	cleanup()
	defer cleanup()
	db.MustExec(`INSERT INTO "Employee"(id, name, status, role) VALUES ($1, 'سامر الخبير التجريبي', 'ACTIVE', 'TECHNICIAN'), ($2, 'مدير تجريبي', 'ACTIVE', 'OWNER')`, emp, boss)
	s := NewOrgService(db)
	_ = s.Seed()
	if _, err := s.Assign("", OrgAssignInput{EmployeeID: emp, UnitID: "ops.fleet", Position: "DEPUTY"}); err != nil {
		t.Skipf("المكان مشغول محلياً: %v", err)
	}
	f, err := ExecutiveFacts(db, emp)
	if err != nil || len(f.Positions) != 1 || !strings.Contains(f.Text(), "نائب — العمليات / العُدد والمركبات") {
		t.Fatalf("الحقائق: %v %+v", err, f)
	}
	a := NewMatrixAgent(db, nil, "")
	a.ShowNames()
	registerExecutiveTools(a, &SecretaryTurnCtx{UserID: boss})
	run := func(name string, in map[string]any) (string, error) {
		for _, tl := range a.tools {
			if tl.Name == name {
				return tl.Run(in, a.newMask())
			}
		}
		t.Fatalf("ماكو أداة %s", name)
		return "", nil
	}
	out, err := run("executive_review", map[string]any{"who": "سامر الخبير"})
	if err != nil || !strings.Contains(out, "سامر الخبير التجريبي") || !strings.Contains(out, "تقييم ماتركس") {
		t.Fatalf("executive_review: %v\n%s", err, out)
	}
	if out, _ := run("executive_review", map[string]any{"unit": "ops.fleet"}); !strings.Contains(out, "سامر") {
		t.Fatalf("كادر الوحدة:\n%s", out)
	}
	if _, err := run("save_executive_opinion", map[string]any{"who": "سامر", "rating": float64(9), "recommendation": "يستحق ترقية واضحة"}); err == nil {
		t.Fatal("تقدير ٩ انقبل")
	}
	for _, r := range []float64{3, 4} {
		if _, err := run("save_executive_opinion", map[string]any{"who": "سامر", "rating": r, "strengths": "ملتزم", "weaknesses": "تأخير م١/ف٣", "recommendation": "يستمر نائب ويتابع الدوام"}); err != nil {
			t.Fatal(err)
		}
	}
	ops, err := ExecutiveOpinions(db, emp)
	if err != nil || len(ops) != 1 || ops[0].Rating != 4 {
		t.Fatalf("الرأي لازم واحد بالشهر ويتحدّث: %v %+v", err, ops)
	}
}
