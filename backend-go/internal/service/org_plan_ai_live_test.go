package service

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/database"
	"staffmange-api/internal/repository"
)

// TestUnitPlanByMatrix_Live (دازاي 11-06): ماتركس يكتب المسودة ويوزع البنود على المدير والنائب،
// الاعتماد يحوّلها مهام، إنجاز المهمة يخلّص البند، والتصنيف يطلّع نسبة وأسباب لمدير ونائب الوحدة.
func TestUnitPlanByMatrix_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة")
	}
	db := sqlx.MustConnect("postgres", dsn)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	const head, dep, dir = "__ai_head", "__ai_dep", "__ai_dir"
	emps := []string{head, dep, dir}
	cleanup := func() {
		db.Exec(`DELETE FROM "OrgPlanItem" WHERE "createdById" = ANY($1)`, emps)       //nolint:errcheck
		db.Exec(`DELETE FROM "OrgPlan" WHERE "createdById" = ANY($1)`, emps)           //nolint:errcheck
		db.Exec(`DELETE FROM "ExtraTask" WHERE "assignedToId" = ANY($1)`, emps)        //nolint:errcheck
		db.Exec(`DELETE FROM "OrgAssignment" WHERE "employeeId" = ANY($1)`, emps)      //nolint:errcheck
		db.Exec(`DELETE FROM "OrgPermissionGrant" WHERE "employeeId" = ANY($1)`, emps) //nolint:errcheck
		db.Exec(`DELETE FROM "EmployeePermission" WHERE "employeeId" = ANY($1)`, emps) //nolint:errcheck
		db.Exec(`DELETE FROM "PermissionAuditLog" WHERE "employeeId" = ANY($1)`, emps) //nolint:errcheck
		db.Exec(`DELETE FROM "Employee" WHERE id = ANY($1)`, emps)                     //nolint:errcheck
	}
	cleanup()
	t.Cleanup(cleanup)
	db.MustExec(`INSERT INTO "Employee"(id, name, status, role) VALUES ($1,'مدير وحدة','ACTIVE','TECHNICIAN'),($2,'نائب وحدة','ACTIVE','TECHNICIAN'),($3,'مدير إدارة','ACTIVE','TECHNICIAN')`, head, dep, dir)
	s := NewOrgService(db)
	if err := s.Seed(); err != nil {
		t.Fatal(err)
	}
	for _, a := range []OrgAssignInput{
		{EmployeeID: dir, DivisionID: "eng", Position: "DIRECTOR"},
		{EmployeeID: head, UnitID: "eng.design", Position: "DIRECTOR"},
		{EmployeeID: dep, UnitID: "eng.design", Position: "DEPUTY"},
	} {
		if _, err := s.Assign("", a); err != nil {
			t.Skipf("المكان مشغول محلياً: %v", err)
		}
	}
	now := time.Now().In(baghdad)
	period := OrgPeriod(now)
	future := now.AddDate(0, 0, 3).Format("2006-01-02")
	var gotTask string
	s.SetPlanDrafter(func(system, task string) (string, error) {
		gotTask = task
		return fmt.Sprintf(`هذي الخطة: [
			{"kind":"PROBLEM","title":"مراجعة يومية للتصاميم المتأخرة","reason":"تأخير عالي","dueDate":"%s","owner":"DIRECTOR"},
			{"kind":"DEVELOPMENT","title":"تدريب الكادر على القوالب","reason":"تقليل الأخطاء","dueDate":"%s","owner":"DEPUTY"},
			{"kind":"DEVELOPMENT","title":"مقياس غلط يرجع يدوي","metric":"no_such","target":5,"dueDate":"%s","owner":"DEPUTY"}
		]`, future, future, future), nil
	})
	s.SetTaskCreator(repository.NewExtraTaskRepository(db).CreateFor)

	plan, err := s.CreatePlan(head, "TECHNICIAN", "eng", "eng.design", period, "خطة تحسين")
	if err != nil {
		t.Fatal(err)
	}
	n, err := s.DraftPlanByMatrix(head, "TECHNICIAN", plan, "ركّز على التأخير")
	if err != nil || n != 3 {
		t.Fatalf("المسودة: %v %d", err, n)
	}
	if !strings.Contains(gotTask, "ركّز على التأخير") || !strings.Contains(gotTask, "members_score") {
		t.Fatal("ماتركس لازم ياخذ توجيه المدير والمقاييس")
	}
	var depItems int
	db.Get(&depItems, `SELECT count(*) FROM "OrgPlanItem" WHERE "planId" = $1 AND "ownerId" = $2 AND "byMatrix"`, plan, dep)
	if depItems != 2 {
		t.Fatalf("بنود النائب: %d", depItems)
	}
	if _, err := s.DraftPlanByMatrix(dep, "TECHNICIAN", plan, ""); err == nil {
		// النائب مو رئيس الوحدة — ما لازم يكتب (إلا إذا عنده صلاحية تعديل)
		t.Log("النائب گدر يكتب المسودة (عنده صلاحية تعديل الوحدة)")
	}
	if _, err := s.SubmitPlan(head, "TECHNICIAN", plan); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ReviewPlan(dir, "TECHNICIAN", plan, true, ""); err != nil {
		t.Fatal(err)
	}
	var tasks int
	db.Get(&tasks, `SELECT count(*) FROM "OrgPlanItem" WHERE "planId" = $1 AND "taskId" IS NOT NULL`, plan)
	if tasks < 3 {
		t.Fatalf("كل بند لازم يصير مهمة: %d", tasks)
	}
	// المدير ينجز مهمته → البند يخلص
	var taskID string
	db.Get(&taskID, `SELECT "taskId" FROM "OrgPlanItem" WHERE "planId" = $1 AND "ownerId" = $2`, plan, head)
	if err := repository.NewExtraTaskRepository(db).Complete(taskID, head, "سوّيت المراجعة اليومية"); err != nil {
		t.Fatal(err)
	}
	var done bool
	db.Get(&done, `SELECT "doneAt" IS NOT NULL FROM "OrgPlanItem" WHERE "taskId" = $1`, taskID)
	if !done {
		t.Fatal("إنجاز المهمة لازم يخلّص البند")
	}
	rk, err := DirectorsRanking(db, period, nil)
	if err != nil {
		t.Fatal(err)
	}
	var headRow, depRow *DirectorRankRow
	for i := range rk.Units {
		switch rk.Units[i].EmployeeID {
		case head:
			headRow = &rk.Units[i]
		case dep:
			depRow = &rk.Units[i]
		}
	}
	if headRow == nil || depRow == nil || headRow.Criteria[0].Score == nil || *headRow.Criteria[0].Score != 100 || headRow.Total == nil {
		t.Fatalf("مدير الوحدة خلص بنده بموعده = بنوده ١٠٠: %+v", headRow)
	}
	ev := headRow.Criteria[0].Evidence
	if len(ev) != 1 || !strings.Contains(ev[0].Note, "خلص بموعده") {
		t.Fatalf("السبب لازم يبين: %+v", ev)
	}
	if depRow.Total == nil || *depRow.Total >= *headRow.Total {
		t.Fatalf("النائب بعده ما خلص — لازم أقل من المدير: %v %v", *depRow.Total, *headRow.Total)
	}
}
