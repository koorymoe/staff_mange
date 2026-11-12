package service

import (
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/timeutil"
)

// (ل) 11-11 بلاغ #32: المراقب ينحسب عليه بس التقارير الي جت وقت دوامه. تقرير ينكتب بعد ما
// يطلع يصير واجب اليوم الجاي (عادي)، مو «فاتك». والإهمال القديم من هالنوع يتعلّم ملغي ونقطته ترجع.
func TestRatingDutyShiftWindow_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ids := []string{"rd-mon", "rd-early", "rd-late"}
	clean := func() {
		for _, q := range []string{`DELETE FROM "MatrixScore" WHERE "employeeId" = 'rd-mon'`, `DELETE FROM "EmployeeNegligence" WHERE "employeeId" = 'rd-mon'`,
			`DELETE FROM "StaffRating" WHERE "raterId" = 'rd-mon'`, `DELETE FROM "DailyWorkReport" WHERE "employeeId" IN ('rd-early','rd-late')`} {
			db.Exec(q)
		}
		for _, id := range ids {
			db.Exec(`DELETE FROM "Employee" WHERE id = $1`, id)
		}
	}
	clean()
	defer clean()
	// قاعدة جديدة (CI): «الواجب يبدي من» = يوم إنشاء القاعدة، فأيام الاختبار (أمس وقبل ٣ أيام) تطلع
	// قبله وما ينحسب عليها شي. نرجّعه أسبوعين للاختبار ونرجّع القيمة الأصلية بالأخير.
	var since string
	_ = db.Get(&since, `SELECT value FROM "OrgGovernorState" WHERE key = 'rating_duty_since'`)
	db.MustExec(`INSERT INTO "OrgGovernorState"(key, value) VALUES ('rating_duty_since', to_char(current_date - 14, 'YYYY-MM-DD'))
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value`)
	defer func() {
		if since == "" {
			db.Exec(`DELETE FROM "OrgGovernorState" WHERE key = 'rating_duty_since'`)
		} else {
			db.Exec(`UPDATE "OrgGovernorState" SET value = $1 WHERE key = 'rating_duty_since'`, since)
		}
	}()
	db.MustExec(`INSERT INTO "Employee"(id, name, role, "shiftStart", "shiftEnd") VALUES
		('rd-mon','مراقب تجربة','MONITOR','09:00','16:00'),('rd-early','موظف صباحي','SALES','09:00','16:00'),('rd-late','موظف مسائي','SALES','09:00','16:00')`)
	loc := timeutil.CompanyLocation()
	n := time.Now().In(loc)
	yday := time.Date(n.Year(), n.Month(), n.Day()-1, 0, 0, 0, 0, loc)
	at := func(h int) time.Time { return yday.Add(time.Duration(h) * time.Hour) }
	y := yday.Format("2006-01-02")
	db.MustExec(`INSERT INTO "DailyWorkReport"(id, "employeeId", day, body, "createdAt") VALUES
		(gen_random_uuid()::text,'rd-early',$1,'تقرير الظهر: تابعت الطلبات وكلمت الزبائن',$2),(gen_random_uuid()::text,'rd-late',$1,'تقرير الليل: كملت الطلبات المتأخرة وكتبت الملاحظات',$3)`, y, at(14), at(20))
	d := NewRatingDuty(db)
	got := func(day string) map[string]bool {
		items, err := d.Owed("rd-mon", day)
		if err != nil {
			t.Fatal(err)
		}
		m := map[string]bool{}
		for _, it := range items {
			if it.Stage == "MONITOR_DAILY" && (it.RateeID == "rd-early" || it.RateeID == "rd-late") {
				m[it.RateeID] = true
			}
		}
		return m
	}
	if m := got(y); !m["rd-early"] || m["rd-late"] {
		t.Fatalf("أمس: لازم تقرير الظهر بس (الليل وصل بعد دوامه): %v", m)
	}
	if m := got(n.Format("2006-01-02")); !m["rd-late"] || m["rd-early"] {
		t.Fatalf("اليوم: تقرير الليل ينتقل لواجب اليوم: %v", m)
	}

	// إهمال قديم: يوم قبل ٣ أيام والتقرير الوحيد بعد الدوام ← يتلغى ونقطته تنلغي.
	d3 := time.Date(n.Year(), n.Month(), n.Day()-3, 0, 0, 0, 0, loc)
	db.MustExec(`DELETE FROM "DailyWorkReport" WHERE "employeeId" IN ('rd-early','rd-late')`)
	db.MustExec(`INSERT INTO "DailyWorkReport"(id, "employeeId", day, body, "createdAt") VALUES (gen_random_uuid()::text,'rd-late',$1,'تقرير الليل: كملت الطلبات المتأخرة وكتبت الملاحظات',$2)`,
		d3.Format("2006-01-02"), d3.Add(21*time.Hour))
	var negID string
	if err := db.Get(&negID, `INSERT INTO "EmployeeNegligence" ("employeeId", day, kind, detail, reason) VALUES ('rd-mon',$1,'MISSED_RATINGS','فاته 1','') RETURNING id`, d3.Format("2006-01-02")); err != nil {
		t.Fatal(err)
	}
	db.MustExec(`INSERT INTO "MatrixScore"("employeeId",source,"sourceId",rule,points,"maxPoints",reason,"at") VALUES ('rd-mon','DUTY',$1,'RATING_DUTY',0,2,'تجربة',now())`, negID)
	if n, err := d.VoidAfterShift(); err != nil || n < 1 {
		t.Fatalf("لازم يتلغى: %d %v", n, err)
	}
	var voided, cancelled bool
	db.Get(&voided, `SELECT "voidedAt" IS NOT NULL FROM "EmployeeNegligence" WHERE id = $1`, negID)
	db.Get(&cancelled, `SELECT "cancelledAt" IS NOT NULL FROM "MatrixScore" WHERE "sourceId" = $1`, negID)
	if !voided || !cancelled {
		t.Fatalf("الإلغاء: إهمال %v نقطة %v", voided, cancelled)
	}
	// إهمال حقيقي (تقرير وقت الدوام) يبقى
	db.MustExec(`INSERT INTO "DailyWorkReport"(id, "employeeId", day, body, "createdAt") VALUES (gen_random_uuid()::text,'rd-early',$1,'تقرير الظهر: تابعت الطلبات وكلمت الزبائن',$2)`,
		d3.AddDate(0, 0, 1).Format("2006-01-02"), d3.AddDate(0, 0, 1).Add(11*time.Hour))
	var realID string
	db.Get(&realID, `INSERT INTO "EmployeeNegligence" ("employeeId", day, kind, detail, reason) VALUES ('rd-mon',$1,'MISSED_RATINGS','فاته 1','') RETURNING id`, d3.AddDate(0, 0, 1).Format("2006-01-02"))
	d.VoidAfterShift() //nolint:errcheck
	db.Get(&voided, `SELECT "voidedAt" IS NOT NULL FROM "EmployeeNegligence" WHERE id = $1`, realID)
	if voided {
		t.Fatal("الإهمال الحقيقي ما يتلغى")
	}
}
