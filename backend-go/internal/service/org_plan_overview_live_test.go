package service

import (
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/database"
	"staffmange-api/internal/repository"
)

// (دازاي 11-08) كل الخطط بمكان واحد، والإرجاع بعد الاعتماد، والبند الجديد على
// خطة معتمدة يصير مهمة — والمدير العادي ما يكدر يسوي الثنتين الأوليات.
func TestPlansOverviewReopen_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة")
	}
	db := sqlx.MustConnect("postgres", dsn)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	const head, adm = "__ov_head", "__ov_admin"
	emps := []string{head, adm}
	cleanup := func() {
		db.Exec(`DELETE FROM "OrgPlanItem" WHERE "createdById" = ANY($1)`, emps)  //nolint:errcheck
		db.Exec(`DELETE FROM "OrgPlan" WHERE "createdById" = ANY($1)`, emps)      //nolint:errcheck
		db.Exec(`DELETE FROM "ExtraTask" WHERE "assignedToId" = ANY($1)`, emps)   //nolint:errcheck
		db.Exec(`DELETE FROM "OrgAssignment" WHERE "employeeId" = ANY($1)`, emps) //nolint:errcheck
		db.Exec(`DELETE FROM "Employee" WHERE id = ANY($1)`, emps)                //nolint:errcheck
	}
	cleanup()
	t.Cleanup(cleanup)
	db.MustExec(`INSERT INTO "Employee"(id, name, status, role) VALUES ($1,'مدير وحدة','ACTIVE','TECHNICIAN'),($2,'مدير نظام','ACTIVE','ADMIN')`, head, adm)
	s := NewOrgService(db)
	if err := s.Seed(); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Assign("", OrgAssignInput{EmployeeID: head, UnitID: "eng.design", Position: "DIRECTOR"}); err != nil {
		t.Skipf("المكان مشغول محلياً: %v", err)
	}
	s.SetTaskCreator(repository.NewExtraTaskRepository(db).CreateFor)
	period := OrgPeriod(time.Now().In(baghdad))
	due := time.Now().In(baghdad).AddDate(0, 0, 5).Format("2006-01-02")

	plan, err := s.CreatePlan(head, "TECHNICIAN", "eng", "eng.design", period, "خطة")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.AddPlanItem(head, "TECHNICIAN", plan, OrgPlanItemInput{Kind: "PROBLEM", Title: "بند أول", DueDate: due, OwnerID: head}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SubmitPlan(head, "TECHNICIAN", plan); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ReviewPlan(adm, "ADMIN", plan, true, ""); err != nil {
		t.Fatal(err)
	}

	// النظرة العامة: للمدير العادي ممنوعة، ولمدير النظام تطلّع الوحدة معتمدة
	if _, err := s.PlansOverview("TECHNICIAN", period); err == nil {
		t.Fatal("المدير العادي ما لازم يشوف كل الخطط")
	}
	rows, err := s.PlansOverview("ADMIN", period)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, r := range rows {
		if r.UnitID != nil && *r.UnitID == "eng.design" {
			found = r.Status == "APPROVED" && r.Items == 1
		}
	}
	if !found {
		t.Fatal("الخطة المعتمدة لازم تطلع بالنظرة العامة")
	}

	// مدير النظام يضيف بند على خطة معتمدة ← يصير مهمة هسه؛ المدير العادي ممنوع
	if _, err := s.AddPlanItem(head, "TECHNICIAN", plan, OrgPlanItemInput{Kind: "DEVELOPMENT", Title: "ممنوع", DueDate: due, OwnerID: head}); err == nil {
		t.Fatal("المدير العادي ما يضيف على خطة معتمدة")
	}
	item, err := s.AddPlanItem(adm, "ADMIN", plan, OrgPlanItemInput{Kind: "DEVELOPMENT", Title: "بند من مدير النظام", DueDate: due, OwnerID: head})
	if err != nil {
		t.Fatal(err)
	}
	var hasTask bool
	db.Get(&hasTask, `SELECT "taskId" IS NOT NULL FROM "OrgPlanItem" WHERE id = $1`, item)
	if !hasTask {
		t.Fatal("البند الجديد على خطة معتمدة لازم يصير مهمة")
	}

	// الإرجاع بعد الاعتماد: للمدير العادي ممنوع، ولمدير النظام بملاحظة
	if _, err := s.ReopenPlan(head, "TECHNICIAN", plan, "عدّل المواعيد"); err == nil {
		t.Fatal("المدير العادي ما يرجّع خطة معتمدة")
	}
	if _, err := s.ReopenPlan(adm, "ADMIN", plan, "قص"); err == nil {
		t.Fatal("الإرجاع لازم بملاحظة واضحة")
	}
	author, err := s.ReopenPlan(adm, "ADMIN", plan, "عدّل المواعيد رجاءً")
	if err != nil || author != head {
		t.Fatalf("الإرجاع: %v %q", err, author)
	}
	var st string
	db.Get(&st, `SELECT status FROM "OrgPlan" WHERE id = $1`, plan)
	if st != "RETURNED" {
		t.Fatalf("الحالة: %s", st)
	}
	// حالة المهمة تطلع بالبند
	p, err := s.planFor("eng", strPtr("eng.design"), period)
	if err != nil || p == nil || len(p.Items) != 2 || p.Items[0].TaskStatus == nil {
		t.Fatalf("حالة المهمة لازم تطلع بالبنود: %v", err)
	}
}

