package service

import (
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// TestOrgGovernor_Live ماتركس الحاكم (دازاي 10-29/30) على قاعدة حقيقية، ويا حالات المراجعة:
// غرامة وحدة لكل حجز، الشخص المسجّل بالشكوى يغلب الليدر، «الزبون ما رد» ما يغرّم مدير
// العمليات، التراجع يرجّع النقاط؛ شهر گبل التشغيل ما ينقيّم، المدير بلا تقييم = بلا حكم،
// تنبيه + شهر ثاني = سحب، تنبيه ملغى يكسر السلسلة، وتراجع السحب والمكان مشغول يرجع خطأ.
func TestOrgGovernor_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	const ldr, dir, oth, own, cust = "__gov_ldr", "__gov_dir", "__gov_oth", "__gov_own", "__gov_cust"
	const bk, bk2 = "__gov_bk", "__gov_bk2"
	emps := []string{ldr, dir, oth, own}
	var oldStart *string
	_ = db.Get(&oldStart, `SELECT value FROM "OrgGovernorState" WHERE key = 'started_at'`)
	cleanup := func() {
		db.Exec(`DELETE FROM "AiAction" WHERE "targetEmployeeId" = ANY($1)`, emps)
		db.Exec(`DELETE FROM "DisciplineEvent" WHERE "employeeId" = ANY($1)`, emps)
		db.Exec(`DELETE FROM "DisciplinePoints" WHERE "employeeId" = ANY($1)`, emps)
		db.Exec(`DELETE FROM "MatrixScore" WHERE "employeeId" = ANY($1)`, emps)
		db.Exec(`DELETE FROM "OrgEvaluation" WHERE "employeeId" = ANY($1)`, emps)
		db.Exec(`DELETE FROM "Complaint" WHERE "customerId" = $1`, cust)
		db.Exec(`DELETE FROM "Booking" WHERE id IN ($1, $2)`, bk, bk2)
		db.Exec(`DELETE FROM "Customer" WHERE id = $1`, cust)
		db.Exec(`DELETE FROM "OrgAssignment" WHERE "employeeId" = ANY($1)`, emps)
		db.Exec(`DELETE FROM "OrgPermissionGrant" WHERE "employeeId" = ANY($1)`, emps)
		db.Exec(`DELETE FROM "EmployeePermission" WHERE "employeeId" = ANY($1)`, emps)
		db.Exec(`DELETE FROM "PermissionAuditLog" WHERE "employeeId" = ANY($1)`, emps)
		db.Exec(`DELETE FROM "Notification" WHERE "employeeId" = ANY($1)`, emps)
		db.Exec(`DELETE FROM "Employee" WHERE id = ANY($1)`, emps)
		if oldStart != nil {
			db.Exec(`UPDATE "OrgGovernorState" SET value = $1 WHERE key = 'started_at'`, *oldStart)
		} else {
			db.Exec(`DELETE FROM "OrgGovernorState" WHERE key = 'started_at'`)
		}
	}
	cleanup()
	defer cleanup()
	var opsTaken int
	db.Get(&opsTaken, `SELECT count(*) FROM "OrgAssignment" WHERE "divisionId" = 'ops' AND "unitId" IS NULL AND position = 'DIRECTOR' AND "endedAt" IS NULL AND NOT acting`)
	if opsTaken > 0 {
		t.Skip("القاعدة المحلية بيها مدير عمليات — التجربة تحتاج المكان فاضي")
	}
	db.MustExec(`INSERT INTO "Employee"(id, name, status, role, "isLeader") VALUES
		($1,'ليدر تجربة','ACTIVE','TECHNICIAN',true),($2,'مدير عمليات تجربة','ACTIVE','TECHNICIAN',false),
		($3,'موظف ثاني','ACTIVE','TECHNICIAN',false),($4,'مالك تجربة','ACTIVE','OWNER',false)`, ldr, dir, oth, own)
	db.MustExec(`INSERT INTO "Customer"(id, name, phone) VALUES ($1, 'زبون حاكم', '07700000099')`, cust)
	db.MustExec(`INSERT INTO "Booking"(id, code, "customerId", status, "projectSupervisorId", "createdAt", "updatedAt") VALUES
		($1, 'GOV-1', $3, 'COMPLETED', $4, now(), now()), ($2, 'GOV-2', $3, 'COMPLETED', $4, now(), now())`, bk, bk2, cust, ldr)

	org := NewOrgService(db)
	if err := org.Seed(); err != nil {
		t.Fatal(err)
	}
	for _, a := range []OrgAssignInput{
		{EmployeeID: dir, DivisionID: "ops", Position: "DIRECTOR"},
		{EmployeeID: ldr, UnitID: "ops.fleet", Position: "MEMBER", AllowanceIQD: 100000},
		{EmployeeID: oth, UnitID: "ops.audit", Position: "MEMBER"},
	} {
		if _, err := org.Assign("", a); err != nil {
			t.Fatal(err)
		}
	}
	db.MustExec(`UPDATE "OrgAssignment" SET "startedAt" = now() - interval '6 months' WHERE "employeeId" = ANY($1)`, emps)

	sw := repository.NewSystemSwitchRepository(db)
	scores := NewMatrixScoreService(repository.NewMatrixScoreRepository(db), nil, sw, nil)
	actions := repository.NewAiActionRepository(db)
	gov := NewOrgGovernor(db, org, scores, repository.NewDisciplineRepository(db), actions, repository.NewAiRepository(db), sw, repository.NewNotificationRepository(db))
	if !gov.on() {
		t.Skip("مفتاح org_governor مطفي بالقاعدة المحلية")
	}
	hooks := map[string]func(model.AiAction) error{}
	gov.RegisterUndo(func(k string, f func(model.AiAction) error) { hooks[k] = f })
	find := func(kind, entity string) model.AiAction {
		var a model.AiAction
		if err := db.Get(&a, `SELECT * FROM "AiAction" WHERE kind = $1 AND "entityId" = $2 ORDER BY "createdAt" DESC LIMIT 1`, kind, entity); err != nil {
			t.Fatalf("ماكو قرار %s %s: %v", kind, entity, err)
		}
		return a
	}
	// undo مثل MatrixAutopilotService.Undo: يعلّم، يشغّل الأثر، ولو فشل يرجّعه DONE
	undo := func(kind, entity string) error {
		a := find(kind, entity)
		if err := actions.Undo(a.ID, own); err != nil {
			return err
		}
		if err := hooks[kind](a); err != nil {
			_ = actions.Reopen(a.ID)
			return err
		}
		return nil
	}
	points := func(e string) int {
		p := model.DisciplineStartingPoints
		db.Get(&p, `SELECT points FROM "DisciplinePoints" WHERE "employeeId" = $1`, e)
		return p
	}

	// ── الغرامات ──
	now := time.Now().In(baghdad)
	m2 := time.Date(now.Year(), now.Month(), 1, 12, 0, 0, 0, baghdad).AddDate(0, -1, 0)
	m1 := m2.AddDate(0, -1, 0)
	m0 := m1.AddDate(0, -1, 0)
	db.MustExec(`INSERT INTO "OrgGovernorState"(key, value) VALUES ('started_at', $1) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value`,
		time.Date(m1.Year(), m1.Month(), 1, 0, 0, 0, 0, baghdad).Add(-time.Hour).UTC().Format(time.RFC3339))
	db.MustExec(`INSERT INTO "Complaint"(id, "customerId", "bookingId", description, status, "createdByEmployeeId", "createdAt", "relatedEmployeeId", "noAnswerCount") VALUES
		('__gov_c1', $1, $2, 'تجربة', 'NEW', $4, now() - interval '72 hours', NULL, 0),
		('__gov_c2', $1, $2, 'نفس الحجز مكررة', 'NEW', $4, now() - interval '71 hours', NULL, 0),
		('__gov_c3', $1, $3, 'على موظف ثاني والزبون ما يرد', 'NEW', $4, now() - interval '60 hours', $5, 1)`,
		cust, bk, bk2, dir, oth)
	gov.RunFines()
	if p := points(ldr); p != 95 { // ٣ + ٢ (ما انحلت خلال ٤٨) — والمكررة على نفس الحجز ما تزيد
		t.Fatalf("غرامة الليدر: %d والمتوقع ٩٥", p)
	}
	if p := points(oth); p != 96 { // الشخص المسجّل بالشكوى يغلب الليدر: ٣ + ١ (نفس الزبون اشتكى گبل)، والتأخير ما ينحسب لأن الزبون ما يرد
		t.Fatalf("غرامة الموظف المسجّل: %d والمتوقع ٩٦", p)
	}
	if p := points(dir); p != 97 { // مدير العمليات: حجز GOV-1 بس (GOV-2 الزبون ما يرد)
		t.Fatalf("غرامة مدير العمليات: %d والمتوقع ٩٧", p)
	}
	gov.RunFines()
	if points(ldr) != 95 || points(dir) != 97 || points(oth) != 96 {
		t.Fatal("الغرامة تكررت")
	}
	if err := undo(model.AiActionOrgFineCustomer, "BOOKING:"+bk); err != nil {
		t.Fatal(err)
	}
	if p := points(ldr); p != 100 {
		t.Fatalf("التراجع ما رجّع النقاط: %d", p)
	}
	gov.RunFines()
	if points(ldr) != 100 {
		t.Fatal("الغرامة المرفوضة رجعت")
	}

	// ── التقييم الشهري ──
	if err := gov.EvaluateMonth(m0.Format("2006-01")); err != nil {
		t.Fatal(err)
	}
	var n int
	db.Get(&n, `SELECT count(*) FROM "OrgEvaluation" WHERE "employeeId" = ANY($1)`, emps)
	if n != 0 {
		t.Fatal("شهر گبل أول تشغيل انقيّم")
	}
	for i, at := range []time.Time{m1, m1.Add(time.Hour), m2, m2.Add(time.Hour)} {
		for _, e := range []string{ldr, oth} {
			db.MustExec(`INSERT INTO "MatrixScore"("employeeId", source, "sourceId", rule, points, "maxPoints", reason, "at")
				VALUES ($1, 'BOOKING', $2, 'PAPER', 0, 2, 'تجربة', $3)`, e, "__gov_src_"+e+string(rune('a'+i)), at)
		}
	}
	if err := gov.EvaluateMonth(m1.Format("2006-01")); err != nil {
		t.Fatal(err)
	}
	dec := func(e, p string) string {
		var d string
		db.Get(&d, `SELECT decision FROM "OrgEvaluation" WHERE "employeeId" = $1 AND period = $2`, e, p)
		return d
	}
	if d := dec(ldr, m1.Format("2006-01")); d != "WARN" {
		t.Fatalf("أول شهر لازم تنبيه: %q", d)
	}
	if d := dec(dir, m1.Format("2006-01")); d != "NODATA" {
		t.Fatalf("مدير بلا تقييم لازم بلا حكم: %q", d)
	}
	// التنبيه الملغى يكسر السلسلة
	if err := undo(model.AiActionOrgWarn, orgEvalEntity(oth, m1.Format("2006-01"))); err != nil {
		t.Fatal(err)
	}
	if err := gov.EvaluateMonth(m2.Format("2006-01")); err != nil {
		t.Fatal(err)
	}
	if d := dec(oth, m2.Format("2006-01")); d != "WARN" {
		t.Fatalf("تنبيه ملغى + شهر سيء لازم تنبيه جديد مو سحب: %q", d)
	}
	if d := dec(ldr, m2.Format("2006-01")); d != "REVOKE" {
		t.Fatalf("ثاني شهر بعد تنبيه لازم سحب: %q", d)
	}
	if mine, _ := org.Mine(ldr); len(mine) != 0 {
		t.Fatal("المنصب ما انسحب")
	}
	// تراجع السحب والمكان انشغل (انعيّن بنفس الوحدة) — خطأ والقرار يبقى
	re, err := org.Assign("", OrgAssignInput{EmployeeID: ldr, UnitID: "ops.fleet", Position: "MEMBER"})
	if err != nil {
		t.Fatal(err)
	}
	if err := undo(model.AiActionOrgRevoke, orgEvalEntity(ldr, m2.Format("2006-01"))); err == nil {
		t.Fatal("التراجع والمكان مشغول لازم يرجع خطأ")
	}
	if a := find(model.AiActionOrgRevoke, orgEvalEntity(ldr, m2.Format("2006-01"))); a.Status != model.AiActionDone {
		t.Fatalf("القرار لازم يرجع DONE حتى ينعاد: %s", a.Status)
	}
	if _, err := org.End("", re.ID, "تفريغ للتجربة"); err != nil {
		t.Fatal(err)
	}
	if err := undo(model.AiActionOrgRevoke, orgEvalEntity(ldr, m2.Format("2006-01"))); err != nil {
		t.Fatalf("بعد تفريغ المكان التراجع لازم ينجح: %v", err)
	}
	if mine, _ := org.Mine(ldr); len(mine) != 1 {
		t.Fatal("التراجع ما رجّع المنصب")
	}
	if err := gov.EvaluateMonth(m2.Format("2006-01")); err != nil {
		t.Fatal(err)
	}
	if mine, _ := org.Mine(ldr); len(mine) != 1 {
		t.Fatal("إعادة التقييم سحبت مرة ثانية")
	}
}
