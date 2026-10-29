package service

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
)

// TestOrgDesk_Live مكتب المدير: الوصول حسب المنصب (الغريب ياخذ «مو موجود»)، المشكلة،
// الخطة ببند مربوط بمقياس (خط الأساس من النظام)، ورأي النائب، وجدول الصلاحيات.
func TestOrgDesk_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة — تخطي اختبار القاعدة الحية")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()
	const dir, dep, stranger, admDir = "__desk_dir", "__desk_dep", "__desk_x", "__desk_adm"
	ids := []any{dir, dep, stranger, admDir}
	cleanup := func() {
		db.Exec(`DELETE FROM "OrgComment" WHERE "authorId" = ANY($1)`, []string{dir, dep, stranger, admDir})
		db.Exec(`DELETE FROM "OrgPlanItem" WHERE "planId" IN (SELECT id FROM "OrgPlan" WHERE "createdById" = $1)`, dir)
		db.Exec(`DELETE FROM "OrgPlan" WHERE "createdById" = $1`, dir)
		db.Exec(`DELETE FROM "OrgIssueReport" WHERE "authorId" IN ($1, $2, $3, $4)`, ids...)
		db.Exec(`DELETE FROM "OrgAuthority" WHERE "updatedById" = $1`, admDir)
		db.Exec(`DELETE FROM "OrgAssignment" WHERE "employeeId" IN ($1, $2, $3, $4)`, ids...)
		db.Exec(`DELETE FROM "OrgPermissionGrant" WHERE "employeeId" IN ($1, $2, $3, $4)`, ids...)
		db.Exec(`DELETE FROM "EmployeePermission" WHERE "employeeId" IN ($1, $2, $3, $4)`, ids...)
		db.Exec(`DELETE FROM "PermissionAuditLog" WHERE "employeeId" IN ($1, $2, $3, $4)`, ids...)
		db.Exec(`DELETE FROM "Employee" WHERE id IN ($1, $2, $3, $4)`, ids...)
	}
	cleanup()
	defer cleanup()
	db.MustExec(`INSERT INTO "Employee"(id, name, status, role) VALUES ($1,'مدير جودة','ACTIVE','TECHNICIAN'),($2,'نائب جودة','ACTIVE','TECHNICIAN'),($3,'غريب','ACTIVE','TECHNICIAN'),($4,'مدير الإدارة','ACTIVE','TECHNICIAN')`, ids...)
	s := NewOrgService(db)
	if err := s.Seed(); err != nil {
		t.Fatal(err)
	}
	for _, a := range []OrgAssignInput{
		{EmployeeID: dir, DivisionID: "quality", Position: "DIRECTOR"},
		{EmployeeID: dep, DivisionID: "quality", Position: "DEPUTY"},
		{EmployeeID: admDir, DivisionID: "admin", Position: "DIRECTOR"},
	} {
		if _, err := s.Assign("", a); err != nil {
			// المكان ممكن مشغول بالقاعدة المحلية من تجارب سابقة
			t.Skipf("ما گدرت أعيّن (%v) — القاعدة بيها مناصب سابقة", err)
		}
	}
	if _, err := s.Desk(stranger, "TECHNICIAN", "quality", ""); !errors.Is(err, ErrOrgHidden) {
		t.Fatalf("الغريب لازم «مو موجود»: %v", err)
	}
	if d, err := s.Desk(admDir, "TECHNICIAN", "quality", ""); err != nil || d.CanEdit {
		t.Fatalf("مدير الإدارة يشوف الكل بس ما يعدّل خطة غيره: %v %+v", err, d)
	}
	issue, err := s.AddIssue(dep, "TECHNICIAN", OrgIssueInput{DivisionID: "quality", Title: "تأخر الرد على الشكاوي", Problem: "الشكاوي تبقى يومين بلا اتصال", Impact: "زبون زعلان", Severity: "HIGH", Principle: intp(9)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.AddIssue(stranger, "TECHNICIAN", OrgIssueInput{DivisionID: "quality", Title: "xxx", Problem: "yyyyyyyyyyyy"}); !errors.Is(err, ErrOrgHidden) {
		t.Fatal("الغريب كتب تقرير")
	}
	if err := s.UpdateIssue(dep, "TECHNICIAN", issue, "RESOLVED", ""); err == nil {
		t.Fatal("الحل بلا شرح انقبل")
	}
	plan, err := s.CreatePlan(dir, "TECHNICIAN", "quality", "", "خطة الشهر")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreatePlan(stranger, "TECHNICIAN", "quality", "", "x"); !errors.Is(err, ErrOrgHidden) {
		t.Fatal("الغريب سوّى خطة")
	}
	cur, _ := s.MetricValue("issues_open", "quality", OrgPeriod(time.Now()))
	target := 0.0
	item, err := s.AddPlanItem(dir, "TECHNICIAN", plan, OrgPlanItemInput{Kind: "PROBLEM", Title: "صفر مشاكل مفتوحة", IssueID: issue, Metric: "issues_open", Target: &target})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.AddPlanItem(dir, "TECHNICIAN", plan, OrgPlanItemInput{Kind: "DEVELOPMENT", Title: "دليل اتصال للزبائن"}); err != nil {
		t.Fatal(err)
	}
	if err := s.AddComment(dep, "TECHNICIAN", "PLAN_ITEM", item, "أقترح نبدي بالشكاوي القديمة"); err != nil {
		t.Fatal(err)
	}
	// المشكلة تنحل ← المقياس ينزل ← النسبة تصعد بلا ما أحد يكتبها
	if err := s.UpdateIssue(dir, "TECHNICIAN", issue, "RESOLVED", "اتصلنا بكل الزبائن"); err != nil {
		t.Fatal(err)
	}
	d, err := s.Desk(dir, "TECHNICIAN", "quality", "")
	if err != nil || d.Plan == nil || len(d.Plan.Items) != 2 || len(d.Comments) == 0 {
		t.Fatalf("المكتب ناقص: %v %+v", err, d)
	}
	var metricItem OrgPlanItem
	for _, it := range d.Plan.Items {
		if it.ID == item {
			metricItem = it
		}
	}
	if !metricItem.ByMatrix || metricItem.Baseline == nil || *metricItem.Baseline != cur {
		t.Fatalf("خط الأساس لازم من النظام (%v): %+v", cur, metricItem)
	}
	if cur >= 1 && metricItem.Progress == 0 {
		t.Fatalf("انحلت المشكلة والنسبة ما تحركت: %+v", metricItem)
	}
	// جدول الصلاحيات
	lim := 50.0
	if err := s.SetAuthority(admDir, "quality", []OrgAuthorityRow{{Decision: "تعويض زبون", Mode: "DECIDE", LimitIQD: &lim}}); err == nil {
		t.Fatal("سقف ٥٠ دينار انقبل")
	}
	lim = 100000
	if err := s.SetAuthority(admDir, "quality", []OrgAuthorityRow{{Decision: "تعويض زبون", Mode: "DECIDE", LimitIQD: &lim}, {Decision: "فصل موظف", Mode: "ESCALATE"}}); err != nil {
		t.Fatal(err)
	}
	d, _ = s.Desk(dir, "TECHNICIAN", "quality", "")
	if len(d.Authority) != 2 {
		t.Fatalf("الصلاحيات: %+v", d.Authority)
	}
}

func intp(v int) *int { return &v }
