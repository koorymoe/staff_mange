package service

import (
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// TestOrgGovernor_Live ماتركس الحاكم (دازاي 10-29) من أوله لآخره على قاعدة حقيقية:
// غرامة الليدر على شكوى حجزه، غرامة مدير العمليات (٤٨ ساعة)، ما يتكرر، التراجع يرجّع
// النقاط؛ شهر تحت ٥٠٪ = تنبيه، الثاني = سحب المنصب وصلاحياته، والتراجع يرجّعه.
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
	const ldr, dir, own, cust, bk, cmp = "__gov_ldr", "__gov_dir", "__gov_own", "__gov_cust", "__gov_bk", "__gov_cmp"
	emps := []string{ldr, dir, own}
	var oldStart *string
	_ = db.Get(&oldStart, `SELECT value FROM "OrgGovernorState" WHERE key = 'started_at'`)
	cleanup := func() {
		db.Exec(`DELETE FROM "AiAction" WHERE "targetEmployeeId" = ANY($1)`, emps)
		db.Exec(`DELETE FROM "DisciplineEvent" WHERE "employeeId" = ANY($1)`, emps)
		db.Exec(`DELETE FROM "DisciplinePoints" WHERE "employeeId" = ANY($1)`, emps)
		db.Exec(`DELETE FROM "MatrixScore" WHERE "employeeId" = ANY($1)`, emps)
		db.Exec(`DELETE FROM "OrgEvaluation" WHERE "employeeId" = ANY($1)`, emps)
		db.Exec(`DELETE FROM "Complaint" WHERE id = $1`, cmp)
		db.Exec(`DELETE FROM "Booking" WHERE id = $1`, bk)
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
	db.MustExec(`INSERT INTO "Employee"(id, name, status, role, "isLeader") VALUES ($1,'ليدر تجربة','ACTIVE','TECHNICIAN',true),($2,'مدير عمليات تجربة','ACTIVE','TECHNICIAN',false),($3,'مالك تجربة','ACTIVE','OWNER',false)`, ldr, dir, own)
	db.MustExec(`INSERT INTO "Customer"(id, name, phone) VALUES ($1, 'زبون حاكم', '07700000099')`, cust)
	db.MustExec(`INSERT INTO "Booking"(id, code, "customerId", status, "projectSupervisorId", "createdAt", "updatedAt") VALUES ($1, 'GOV-1', $2, 'COMPLETED', $3, now(), now())`, bk, cust, ldr)

	org := NewOrgService(db)
	if err := org.Seed(); err != nil {
		t.Fatal(err)
	}
	if _, err := org.Assign("", OrgAssignInput{EmployeeID: dir, DivisionID: "ops", Position: "DIRECTOR"}); err != nil {
		t.Fatal(err)
	}
	if _, err := org.Assign("", OrgAssignInput{EmployeeID: ldr, UnitID: "ops.fleet", Position: "MEMBER", AllowanceIQD: 100000}); err != nil {
		t.Fatal(err)
	}
	db.MustExec(`UPDATE "OrgAssignment" SET "startedAt" = now() - interval '6 months' WHERE "employeeId" = ANY($1)`, []string{ldr, dir})

	sw := repository.NewSystemSwitchRepository(db)
	scores := NewMatrixScoreService(repository.NewMatrixScoreRepository(db), nil, sw, nil)
	actions := repository.NewAiActionRepository(db)
	gov := NewOrgGovernor(db, org, scores, repository.NewDisciplineRepository(db), actions, repository.NewAiRepository(db), sw, repository.NewNotificationRepository(db))
	hooks := map[string]func(model.AiAction) error{}
	gov.RegisterUndo(func(k string, f func(model.AiAction) error) { hooks[k] = f })
	undo := func(kind, entity string) {
		var a model.AiAction
		if err := db.Get(&a, `SELECT * FROM "AiAction" WHERE kind = $1 AND "entityId" = $2`, kind, entity); err != nil {
			t.Fatalf("ماكو قرار %s: %v", kind, err)
		}
		if err := actions.Undo(a.ID, own); err != nil {
			t.Fatal(err)
		}
		if err := hooks[kind](a); err != nil {
			t.Fatal(err)
		}
	}
	points := func(e string) int {
		p := model.DisciplineStartingPoints
		db.Get(&p, `SELECT points FROM "DisciplinePoints" WHERE "employeeId" = $1`, e)
		return p
	}

	// ── الغرامات: شكوى عمرها ٣ أيام وبعدها مفتوحة ──
	db.MustExec(`INSERT INTO "OrgGovernorState"(key, value) VALUES ('started_at', $1) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value`,
		time.Now().Add(-10*24*time.Hour).UTC().Format(time.RFC3339))
	db.MustExec(`INSERT INTO "Complaint"(id, "customerId", "bookingId", description, status, "createdByEmployeeId", "createdAt")
		VALUES ($1, $2, $3, 'تجربة', 'NEW', $4, now() - interval '72 hours')`, cmp, cust, bk, dir)
	if !gov.on() {
		t.Skip("مفتاح org_governor مطفي بالقاعدة المحلية")
	}
	gov.RunFines()
	if p := points(ldr); p != 95 { // ٣ أساس + ٢ ما انحلت خلال ٤٨ = ٥ نقاط (٥٠ ألف)
		t.Fatalf("غرامة الليدر: رصيده %d والمتوقع ٩٥", p)
	}
	if p := points(dir); p != 96 { // ٣ + يوم زيادة = ٤ (٤٠ ألف)
		t.Fatalf("غرامة مدير العمليات: رصيده %d والمتوقع ٩٦", p)
	}
	gov.RunFines() // ما يتكرر
	if points(ldr) != 95 || points(dir) != 96 {
		t.Fatal("الغرامة تكررت")
	}
	undo(model.AiActionOrgFineCustomer, cmp)
	if p := points(ldr); p != 100 {
		t.Fatalf("التراجع ما رجّع النقاط: %d", p)
	}
	gov.RunFines() // المرفوضة ما ترجع
	if points(ldr) != 100 {
		t.Fatal("الغرامة المرفوضة رجعت")
	}

	// ── التقييم الشهري: شهرين تحت ٥٠٪ ──
	now := time.Now().In(baghdad)
	m2 := time.Date(now.Year(), now.Month(), 1, 12, 0, 0, 0, baghdad).AddDate(0, -1, 0)
	m1 := m2.AddDate(0, -1, 0)
	for i, at := range []time.Time{m1, m1.Add(time.Hour), m2, m2.Add(time.Hour)} {
		db.MustExec(`INSERT INTO "MatrixScore"("employeeId", source, "sourceId", rule, points, "maxPoints", reason, "at")
			VALUES ($1, 'BOOKING', $2, 'PAPER', 0, 2, 'تجربة', $3)`, ldr, "__gov_src_"+string(rune('a'+i)), at)
	}
	if err := gov.EvaluateMonth(m1.Format("2006-01")); err != nil {
		t.Fatal(err)
	}
	var dec string
	db.Get(&dec, `SELECT decision FROM "OrgEvaluation" WHERE "employeeId" = $1 AND period = $2`, ldr, m1.Format("2006-01"))
	if dec != "WARN" {
		t.Fatalf("أول شهر لازم تنبيه: %q", dec)
	}
	if err := gov.EvaluateMonth(m2.Format("2006-01")); err != nil {
		t.Fatal(err)
	}
	db.Get(&dec, `SELECT decision FROM "OrgEvaluation" WHERE "employeeId" = $1 AND period = $2`, ldr, m2.Format("2006-01"))
	if dec != "REVOKE" {
		t.Fatalf("ثاني شهر لازم سحب: %q", dec)
	}
	mine, _ := org.Mine(ldr)
	if len(mine) != 0 {
		t.Fatal("المنصب ما انسحب")
	}
	var unitFleet int
	db.Get(&unitFleet, `SELECT count(*) FROM "EmployeePermission" ep JOIN "Permission" p ON p.id = ep."permissionId" WHERE ep."employeeId" = $1 AND p.name = 'unit_fleet'`, ldr)
	if unitFleet != 0 {
		t.Fatal("صلاحيات المنصب ما انسحبت")
	}
	var aid string
	db.Get(&aid, `SELECT id FROM "OrgAssignment" WHERE "employeeId" = $1`, ldr)
	undo(model.AiActionOrgRevoke, aid)
	if mine, _ := org.Mine(ldr); len(mine) != 1 {
		t.Fatal("التراجع ما رجّع المنصب")
	}
	// نفس الشهر ما ينعاد تقييمه
	if err := gov.EvaluateMonth(m2.Format("2006-01")); err != nil {
		t.Fatal(err)
	}
	if mine, _ := org.Mine(ldr); len(mine) != 1 {
		t.Fatal("إعادة التقييم سحبت مرة ثانية")
	}
}
