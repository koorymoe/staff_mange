package service

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/repository"
)

// (ل) 11-12 دازاي: الماتشينغ — مشكلة بحجز (تقييم سيئ / «ما يعرف» / «يحتاج تدريب على…») = إلزامي
// ما ينلغي ويوصل لمسؤول التدريب؛ نقص مهارة = اقتراح ينلغي بسبب؛ التقرير يوصل للمالك؛ المالك ما ينحسب.
func TestTrainingMatching_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	emps := []string{"tm-lead", "tm-t1", "tm-t2", "tm-t3", "tm-own", "tm-head"}
	clean := func() {
		db.Exec(`DELETE FROM "TrainingMatch" WHERE "employeeId" LIKE 'tm-%'`)
		db.Exec(`DELETE FROM "CrewRating" WHERE "bookingId" LIKE 'tm-%'`)
		db.Exec(`DELETE FROM "BookingAssignment" WHERE "bookingId" LIKE 'tm-%'`)
		db.Exec(`DELETE FROM "EmployeeSkill" WHERE "employeeId" LIKE 'tm-%'`)
		db.Exec(`DELETE FROM "Booking" WHERE id LIKE 'tm-%'`)
		db.Exec(`DELETE FROM "TrainingCurriculum" WHERE id LIKE 'tm-c%' AND "parentId" IS NOT NULL`)
		db.Exec(`DELETE FROM "TrainingCurriculum" WHERE id LIKE 'tm-c%'`)
		db.Exec(`DELETE FROM "Skill" WHERE id LIKE 'tm-%'`)
		db.Exec(`DELETE FROM "Service" WHERE id LIKE 'tm-%'`)
		db.Exec(`DELETE FROM "Customer" WHERE id LIKE 'tm-%'`)
		for _, id := range emps {
			db.Exec(`DELETE FROM "Notification" WHERE "employeeId" = $1`, id)
			db.Exec(`DELETE FROM "EmployeePermission" WHERE "employeeId" = $1`, id)
			db.Exec(`DELETE FROM "Employee" WHERE id = $1`, id)
		}
	}
	clean()
	defer clean()
	for _, id := range emps {
		role := "TECHNICIAN"
		if id == "tm-own" {
			role = "OWNER"
		}
		db.MustExec(`INSERT INTO "Employee"(id, name, role) VALUES ($1, $1, $2)`, id, role)
	}
	db.MustExec(`INSERT INTO "Permission"(id, name, label) VALUES (gen_random_uuid()::text, 'training_manage', 'x') ON CONFLICT (name) DO NOTHING`)
	db.MustExec(`INSERT INTO "EmployeePermission"(id, "employeeId", "permissionId") SELECT gen_random_uuid()::text, 'tm-head', id FROM "Permission" WHERE name = 'training_manage'`)
	db.MustExec(`INSERT INTO "Customer"(id, name, phone) VALUES ('tm-cu', 'زبون', '07700000001')`)
	db.MustExec(`INSERT INTO "Service"(id, name) VALUES ('tm-dash', 'داش كام اختبار')`)
	db.MustExec(`INSERT INTO "Skill"(id, name, "serviceId") VALUES ('tm-sk1', 'تنصيب قفل باب', 'tm-dash')`)
	// المنهج: فني ← داش كام ← تنصيب قفل باب ← الحفر
	db.MustExec(`INSERT INTO "TrainingCurriculum"(id, role, title, "serviceId") VALUES ('tm-c1', 'TECHNICIAN', 'داش كام', 'tm-dash')`)
	db.MustExec(`INSERT INTO "TrainingCurriculum"(id, role, title, "parentId") VALUES ('tm-c2', 'TECHNICIAN', 'تنصيب قفل باب', 'tm-c1')`)
	db.MustExec(`INSERT INTO "TrainingCurriculum"(id, role, title, "parentId") VALUES ('tm-c3', 'TECHNICIAN', 'الحفر', 'tm-c2')`)
	db.MustExec(`INSERT INTO "Booking"(id, code, "customerId", status, "serviceId", "completedAt") VALUES ('tm-b1', 'TM-1', 'tm-cu', 'COMPLETED', 'tm-dash', now())`)
	db.MustExec(`INSERT INTO "Booking"(id, code, "customerId", status, "serviceId", "completedAt") VALUES ('tm-b2', 'TM-2', 'tm-cu', 'COMPLETED', 'tm-dash', now())`)
	for i, e := range []string{"tm-t1", "tm-t2", "tm-t3"} {
		db.MustExec(`INSERT INTO "BookingAssignment"(id, "bookingId", "employeeId", role) VALUES (gen_random_uuid()::text, 'tm-b1', $1, $2)`, e, []string{"TECH_1", "TECH_2", "TECH_3"}[i])
	}
	db.MustExec(`INSERT INTO "BookingAssignment"(id, "bookingId", "employeeId", role) VALUES (gen_random_uuid()::text, 'tm-b2', 'tm-own', 'TECH_1')`)
	db.MustExec(`INSERT INTO "EmployeeSkill"(id, "employeeId", "skillId", "canPerform") VALUES (gen_random_uuid()::text, 'tm-t3', 'tm-sk1', true)`)
	chain := repository.NewMatrixChainRepository(db)
	note := "ما يعرف يحفر"
	want := "الحفر"
	if err := chain.SaveRating("tm-b1", "tm-lead", "tm-t1", 4, &note); err != nil { // نجوم عالية بس كاتب «ما يعرف»
		t.Fatal(err)
	}
	if err := chain.SaveRating("tm-b1", "tm-lead", "tm-t2", 5, nil, &want); err != nil { // حدد «يحتاج تدريب على»
		t.Fatal(err)
	}
	good := "ممتاز"
	if err := chain.SaveRating("tm-b1", "tm-lead", "tm-t3", 5, &good); err != nil {
		t.Fatal(err)
	}
	if err := chain.SaveRating("tm-b1", "tm-lead", "tm-own", 1, nil); err != nil {
		t.Fatal(err)
	}
	m := NewTrainingMatching(db, repository.NewNotificationRepository(db))
	if _, _, err := m.Generate(); err != nil {
		t.Fatal(err)
	}
	if _, _, err := m.Generate(); err != nil { // ما يكرر
		t.Fatal(err)
	}
	rows, err := m.List("")
	if err != nil {
		t.Fatal(err)
	}
	by := map[string][]TrainingNeedRow{}
	for _, r := range rows {
		if len(r.EmployeeID) > 3 && r.EmployeeID[:3] == "tm-" {
			by[r.EmployeeID+"|"+r.Kind] = append(by[r.EmployeeID+"|"+r.Kind], r)
		}
	}
	t1 := by["tm-t1|BOOKING_ISSUE"]
	if len(t1) != 1 || t1[0].DueAt == nil || t1[0].CurriculumID == nil || *t1[0].CurriculumID != "tm-c1" {
		t.Fatalf("«ما يعرف» = إلزامي على مسار الخدمة: %+v", t1)
	}
	t2 := by["tm-t2|BOOKING_ISSUE"]
	if len(t2) != 1 || t2[0].Topic != "الحفر" || t2[0].CurriculumID == nil || *t2[0].CurriculumID != "tm-c3" || t2[0].Path == nil || *t2[0].Path != "داش كام ← تنصيب قفل باب ← الحفر" {
		t.Fatalf("«يحتاج تدريب على الحفر» يطابق المفردة بالمنهج: %+v", t2)
	}
	if len(by["tm-t3|BOOKING_ISSUE"]) != 0 || len(by["tm-t3|SKILL_GAP"]) != 0 {
		t.Fatal("تقييم زين ومهارته موجودة = ماكو شي")
	}
	if len(by["tm-own|BOOKING_ISSUE"])+len(by["tm-own|SKILL_GAP"]) != 0 {
		t.Fatal("المالك ما ينحسب")
	}
	gap := by["tm-t1|SKILL_GAP"]
	if len(gap) != 1 || gap[0].Topic != "تنصيب قفل باب" {
		t.Fatalf("نقص مهارة = اقتراح: %+v", gap)
	}
	var n int
	db.Get(&n, `SELECT count(*) FROM "Notification" WHERE "employeeId" = 'tm-head' AND type = 'TRAINING_NEED'`)
	if n != 2 {
		t.Fatalf("مسؤول التدريب يوصله الإلزامي بس (٢): %d", n)
	}
	db.Get(&n, `SELECT count(*) FROM "Notification" WHERE "employeeId" = 'tm-lead'`)
	if n != 0 {
		t.Fatal("الليدر ما يوصله شي من الماتشينغ")
	}

	// (11-13) النص يطابق الي صار: ٤/٥ مع «ما يعرف» ما ينكتب «قيّمه ٤/٥».
	if strings.Contains(*t1[0].Reason, "قيّمه") || !strings.Contains(*t2[0].Reason, "طلب يتدرب على «الحفر»") {
		t.Fatalf("نص السبب: %q / %q", *t1[0].Reason, *t2[0].Reason)
	}
	if t1[0].DueAt.Before(time.Now().Add(6 * 24 * time.Hour)) {
		t.Fatal("المهلة من هسه")
	}
	// (11-13) نفس الموظف ونفس الموضوع بحجز ثاني وبعده مفتوح = ما يتكرر ولا تنبيه ثاني.
	db.MustExec(`INSERT INTO "Booking"(id, code, "customerId", status, "serviceId", "completedAt") VALUES ('tm-b3', 'TM-3', 'tm-cu', 'COMPLETED', 'tm-dash', now())`)
	db.MustExec(`INSERT INTO "BookingAssignment"(id, "bookingId", "employeeId", role) VALUES (gen_random_uuid()::text, 'tm-b3', 'tm-t2', 'TECH_1')`)
	if err := chain.SaveRating("tm-b3", "tm-lead", "tm-t2", 5, nil, &want); err != nil {
		t.Fatal(err)
	}
	if _, _, err := m.Generate(); err != nil {
		t.Fatal(err)
	}
	db.Get(&n, `SELECT count(*) FROM "TrainingMatch" WHERE "employeeId" = 'tm-t2' AND kind = 'BOOKING_ISSUE'`)
	if n != 1 {
		t.Fatalf("نفس الموضوع ما يتكرر: %d", n)
	}
	db.Get(&n, `SELECT count(*) FROM "Notification" WHERE "employeeId" = 'tm-head' AND type = 'TRAINING_NEED'`)
	if n != 2 {
		t.Fatalf("ما يوصل تنبيه ثاني: %d", n)
	}

	if _, err := m.Dismiss(t1[0].ID, "سبب"); err != ErrNeedMandatory {
		t.Fatalf("الإلزامي ما ينلغي: %v", err)
	}
	if _, err := m.Dismiss(gap[0].ID, ""); err != ErrNeedState {
		t.Fatal("الاقتراح يحتاج سبب")
	}
	if x, err := m.Dismiss(gap[0].ID, "عنده خبرة من شغله القديم"); err != nil || x.Status != "DISMISSED" {
		t.Fatalf("الاقتراح ينلغي: %v", err)
	}
	if x, err := m.Start(t1[0].ID); err != nil || x.Status != "IN_TRAINING" {
		t.Fatalf("بداية التدريب: %v", err)
	}
	if _, err := m.Complete("tm-head", t1[0].ID, "", 4); err != ErrNeedReport {
		t.Fatal("بلا تقرير ما يخلص")
	}
	x, err := m.Complete("tm-head", t1[0].ID, "تدرب على الحفر يومين وطبّق بحجزين", 4)
	if err != nil || x.Status != "DONE" || x.ReportScore == nil || *x.ReportScore != 4 || x.ReportedBy == nil {
		t.Fatalf("التقرير: %v %+v", err, x)
	}
	db.Get(&n, `SELECT count(*) FROM "Notification" WHERE "employeeId" = 'tm-own' AND type = 'TRAINING_REPORT'`)
	if n != 1 {
		t.Fatalf("التقرير يوصل للمالك: %d", n)
	}
	done, _ := m.List("done")
	found := false
	for _, d := range done {
		found = found || d.ID == t1[0].ID
	}
	if !found {
		t.Fatal("يطلع بتقارير التدريب")
	}

	// المتأخر: مرة وحدة للمالك ومسؤول التدريب
	db.MustExec(`UPDATE "TrainingMatch" SET "dueAt" = now() - interval '1 day' WHERE id = $1`, t2[0].ID)
	if k, _ := m.NotifyOverdue(); k < 1 {
		t.Fatal("المتأخر ينبّه")
	}
	if k, _ := m.NotifyOverdue(); k != 0 {
		t.Fatal("مرة وحدة بس")
	}
	if _, err := m.Add("tm-head", "tm-own", "شي", nil, ""); err != ErrNeedNotFound {
		t.Fatal("ما ينضاف تدريب للمالك")
	}
}
