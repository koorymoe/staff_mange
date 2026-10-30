package service

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
)

// TestUnitDevelopment_Live تطوير الوحدة (دازاي 10-30): رئيس الوحدة يشخّص ويكتب خطة حل،
// يسوّي خطة تطوير بمواعيد ومسؤول، يرفعها لمدير إدارته، والمدير يرجّعها (بملاحظة) أو يعتمدها.
// ولائحة العمل ترجع ٢٠ فقرة بمادتها.
func TestUnitDevelopment_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	const head, dir, mem, other = "__ud_head", "__ud_dir", "__ud_mem", "__ud_other"
	emps := []string{head, dir, mem, other}
	cleanup := func() {
		db.Exec(`DELETE FROM "OrgPlanItem" WHERE "createdById" = ANY($1)`, emps)
		db.Exec(`DELETE FROM "OrgPlan" WHERE "createdById" = ANY($1)`, emps)
		db.Exec(`DELETE FROM "OrgIssueReport" WHERE "authorId" = ANY($1)`, emps)
		db.Exec(`DELETE FROM "OrgAssignment" WHERE "employeeId" = ANY($1)`, emps)
		db.Exec(`DELETE FROM "OrgPermissionGrant" WHERE "employeeId" = ANY($1)`, emps)
		db.Exec(`DELETE FROM "EmployeePermission" WHERE "employeeId" = ANY($1)`, emps)
		db.Exec(`DELETE FROM "PermissionAuditLog" WHERE "employeeId" = ANY($1)`, emps)
		db.Exec(`DELETE FROM "Employee" WHERE id = ANY($1)`, emps)
	}
	cleanup()
	defer cleanup()
	db.MustExec(`INSERT INTO "Employee"(id, name, status, role) VALUES ($1,'رئيس وحدة','ACTIVE','TECHNICIAN'),($2,'مدير إدارة','ACTIVE','TECHNICIAN'),($3,'عضو','ACTIVE','TECHNICIAN'),($4,'رئيس وحدة ثانية','ACTIVE','TECHNICIAN')`, head, dir, mem, other)
	s := NewOrgService(db)
	if err := s.Seed(); err != nil {
		t.Fatal(err)
	}
	for _, a := range []OrgAssignInput{
		{EmployeeID: dir, DivisionID: "eng", Position: "DIRECTOR"},
		{EmployeeID: head, UnitID: "eng.projects", Position: "DIRECTOR"},
		{EmployeeID: mem, UnitID: "eng.projects", Position: "MEMBER"},
		{EmployeeID: other, UnitID: "eng.design", Position: "DIRECTOR"},
	} {
		if _, err := s.Assign("", a); err != nil {
			t.Skipf("المكان مشغول محلياً: %v", err)
		}
	}
	// لائحة العمل
	regs, err := s.Regulations()
	if err != nil || len(regs) != 20 || regs[6].Code != "م2/ف1" {
		t.Fatalf("اللائحة: %v %d %+v", err, len(regs), regs)
	}
	// التشخيص: رئيس الوحدة يكتب مشكلة وحدته، ورئيس وحدة ثانية ما يگدر
	issue, err := s.AddIssue(head, "TECHNICIAN", OrgIssueInput{DivisionID: "eng", UnitID: "eng.projects", Title: "تأخير المراحل", Problem: "المشاريع تتجاوز حد المرحلة بلا سبب", Principle: intp(7)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.AddIssue(other, "TECHNICIAN", OrgIssueInput{DivisionID: "eng", UnitID: "eng.projects", Title: "xxx", Problem: "yyyyyyyyyyyy"}); !errors.Is(err, ErrOrgHidden) {
		t.Fatal("رئيس وحدة ثانية كتب بوحدة غيره")
	}
	if err := s.SetSolutionPlan(head, "TECHNICIAN", issue, "مراجعة أسبوعية لكل مشروع ويا سبب مكتوب"); err != nil {
		t.Fatal(err)
	}
	// خطة الوحدة
	plan, err := s.CreatePlan(head, "TECHNICIAN", "eng", "eng.projects", "", "تطوير المشاريع")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreatePlan(other, "TECHNICIAN", "eng", "eng.projects", "", "x"); !errors.Is(err, ErrOrgHidden) {
		t.Fatal("رئيس وحدة ثانية سوّى خطة لغير وحدته")
	}
	if _, err := s.SubmitPlan(head, "TECHNICIAN", plan); err == nil {
		t.Fatal("خطة فاضية انرفعت")
	}
	past := time.Now().In(baghdad).AddDate(0, 0, -1).Format("2006-01-02")
	item, err := s.AddPlanItem(head, "TECHNICIAN", plan, OrgPlanItemInput{Kind: "PROBLEM", Title: "سبب مكتوب لكل تأخير", IssueID: issue, DueDate: past, OwnerID: mem})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.AddPlanItem(head, "TECHNICIAN", plan, OrgPlanItemInput{Kind: "DEVELOPMENT", Title: "x y z", OwnerID: "emp_t1"}); err == nil {
		t.Fatal("مسؤول من إدارة ثانية انقبل")
	}
	to, err := s.SubmitPlan(head, "TECHNICIAN", plan)
	if err != nil || len(to) != 1 || to[0] != dir {
		t.Fatalf("الرفع لازم يوصل مدير الإدارة: %v %v", err, to)
	}
	if _, err := s.AddPlanItem(head, "TECHNICIAN", plan, OrgPlanItemInput{Kind: "DEVELOPMENT", Title: "بعد الرفع"}); err == nil {
		t.Fatal("بند انضاف لخطة مرفوعة")
	}
	if _, err := s.ReviewPlan(head, "TECHNICIAN", plan, true, ""); !errors.Is(err, ErrOrgHidden) {
		t.Fatal("رئيس الوحدة اعتمد خطته بنفسه")
	}
	if _, err := s.ReviewPlan(dir, "TECHNICIAN", plan, false, ""); err == nil {
		t.Fatal("إرجاع بلا ملاحظة انقبل")
	}
	if author, err := s.ReviewPlan(dir, "TECHNICIAN", plan, false, "حدد موعد لكل بند"); err != nil || author != head {
		t.Fatalf("الإرجاع: %v %q", err, author)
	}
	if _, err := s.SubmitPlan(head, "TECHNICIAN", plan); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ReviewPlan(dir, "TECHNICIAN", plan, true, ""); err != nil {
		t.Fatal(err)
	}
	// المسؤول عن البند يحدّث نسبته
	pct := 40
	if err := s.UpdatePlanItem(mem, "TECHNICIAN", item, &pct, nil); err != nil {
		t.Fatalf("المسؤول ما گدر يحدّث بنده: %v", err)
	}
	d, err := s.Desk(dir, "TECHNICIAN", "eng", "")
	if err != nil || len(d.UnitPlans) != 1 || d.UnitPlans[0].Status != "APPROVED" {
		t.Fatalf("مكتب الإدارة لازم يشوف خطة الوحدة المعتمدة: %v %+v", err, d)
	}
	it := d.UnitPlans[0].Items[0]
	if !it.Overdue || it.OwnerName == nil || it.Progress != 40 {
		t.Fatalf("البند: متأخر=%v مسؤول=%v نسبة=%d", it.Overdue, it.OwnerName, it.Progress)
	}
	dh, _ := s.Desk(head, "TECHNICIAN", "eng", "")
	if len(dh.EditableUnits) != 1 || dh.EditableUnits[0] != "eng.projects" || dh.CanEdit {
		t.Fatalf("رئيس الوحدة يعدّل وحدته بس: %+v canEdit=%v", dh.EditableUnits, dh.CanEdit)
	}
}
