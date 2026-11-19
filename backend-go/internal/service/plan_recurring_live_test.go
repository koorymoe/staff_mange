package service

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// (دازاي 11-19) الواجب اليومي ما إله «متأخر»: درجته = نقاط ماتركس لشغل صاحبه بشهر الخطة.
// والبند العادي يبقى: خلص بعد موعده بيوم = ٩٠.
func TestPlanItemRecurringScoreLive(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db := sqlx.MustConnect("postgres", dsn)
	cleanup := func() {
		db.Exec(`DELETE FROM "MatrixScore" WHERE "employeeId" = 'pr_emp'`)
		db.Exec(`DELETE FROM "OrgPlanItem" WHERE "planId" = 'pr_plan'`)
		db.Exec(`DELETE FROM "OrgPlan" WHERE id = 'pr_plan'`)
		db.Exec(`DELETE FROM "OrgDivision" WHERE id = 'pr_div'`)
		db.Exec(`DELETE FROM "Employee" WHERE id = 'pr_emp'`)
	}
	cleanup()
	t.Cleanup(func() { cleanup(); db.Close() })
	db.MustExec(`INSERT INTO "Employee" (id, name, username, password, role, status) VALUES ('pr_emp', 'pr', 'pr_emp', 'x', 'PROCUREMENT_ADMIN', 'ACTIVE')`)
	db.MustExec(`INSERT INTO "OrgDivision" (id, name) VALUES ('pr_div', 'اختبار')`)
	db.MustExec(`INSERT INTO "OrgPlan" (id, "divisionId", period) VALUES ('pr_plan', 'pr_div', to_char(baghdad_today(), 'YYYY-MM'))`)
	db.MustExec(`INSERT INTO "OrgPlanItem" (id, "planId", kind, title, "ownerId", "dueDate", "doneAt", recurring) VALUES
		('pr_daily', 'pr_plan', 'DEVELOPMENT', 'متابعة يومية', 'pr_emp', baghdad_today() - 1, now(), true),
		('pr_once', 'pr_plan', 'DEVELOPMENT', 'مرة وحدة', 'pr_emp', baghdad_today() - 1, now(), false)`)
	db.MustExec(`INSERT INTO "MatrixScore" ("employeeId", source, "sourceId", rule, points, "maxPoints", reason, "at") VALUES
		('pr_emp', 'DAY', 'pr1', 'PROC_RESPONSE', 2, 2, 'x', now()), ('pr_emp', 'DAY', 'pr2', 'SHORTAGE_RESOLVE', 1, 2, 'x', now()),
		('pr_emp', 'DAY', 'pr3', 'ATTENDANCE', 0, 2, 'x', now())`) // الحضور ما يدخل بالواجب
	score := func(id string) float64 {
		var v float64
		if err := db.Get(&v, `SELECT (`+planItemScoreSQL+`)::float8 FROM "OrgPlanItem" i WHERE i.id = $1`, id); err != nil {
			t.Fatal(err)
		}
		return v
	}
	if v := score("pr_daily"); v != 75 {
		t.Fatalf("الواجب اليومي = %v نريد 75 (٣ من ٤ نقاط)", v)
	}
	if v := score("pr_once"); v != 90 {
		t.Fatalf("البند العادي = %v نريد 90", v)
	}
}
