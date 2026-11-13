package service

import (
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"staffmange-api/internal/repository"
)

// (ل) 11-11: الانضباط بالدوام + أعذار التأخير — المتأخر بلا عذر ينحسب، بعذر لا، الغياب أثقل،
// والي يوافق هو المسؤول المباشر بالهيكل بس (والغريب = مو موجود).
func TestLateExcuseDiscipline_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ids := []string{"ld-dir", "ld-unit", "ld-mem", "ld-out"}
	clean := func() {
		for _, q := range []string{`DELETE FROM "LateExcuse" WHERE "employeeId" = ANY($1)`, `DELETE FROM "MatrixScore" WHERE "employeeId" = ANY($1)`,
			`DELETE FROM "Attendance" WHERE "employeeId" = ANY($1)`, `DELETE FROM "OrgAssignment" WHERE "employeeId" = ANY($1)`,
			`DELETE FROM "Notification" WHERE "employeeId" = ANY($1)`} {
			db.Exec(q, pq.Array(ids))
		}
		db.Exec(`DELETE FROM "OrgUnit" WHERE id = 'ld-u'`)
		db.Exec(`DELETE FROM "OrgDivision" WHERE id = 'ld-d'`)
		db.Exec(`DELETE FROM "Employee" WHERE id = ANY($1)`, pq.Array(ids))
	}
	clean()
	defer clean()
	db.MustExec(`INSERT INTO "Employee"(id, name, "shiftStart", "shiftEnd") VALUES
		('ld-dir','مدير','09:00','17:00'),('ld-unit','رئيس وحدة','09:00','17:00'),('ld-mem','عضو','09:00','17:00'),('ld-out','غريب','09:00','17:00')`)
	db.MustExec(`INSERT INTO "OrgDivision"(id, name) VALUES ('ld-d','إدارة تجربة')`)
	db.MustExec(`INSERT INTO "OrgUnit"(id, "divisionId", name) VALUES ('ld-u','ld-d','وحدة تجربة')`)
	db.MustExec(`INSERT INTO "OrgAssignment"("employeeId","divisionId","unitId",position) VALUES
		('ld-dir','ld-d',NULL,'DIRECTOR'),('ld-unit','ld-d','ld-u','DIRECTOR'),('ld-mem','ld-d','ld-u','MEMBER')`)

	// الأيام: قبل ٥ و٤ و٣ و٢ أيام. التطبيق نخلّيه يبدي قبل أسبوع للاختبار.
	old := DisciplineStart
	n := time.Now().In(debriefLoc)
	today := time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, debriefLoc)
	DisciplineStart = today.AddDate(0, 0, -7)
	defer func() { DisciplineStart = old }()
	day := func(back int) time.Time { return today.AddDate(0, 0, -back) }
	att := func(back, pts int, checkIn string) {
		d := day(back)
		db.MustExec(`INSERT INTO "MatrixScore"("employeeId",source,"sourceId",rule,points,"maxPoints",reason,"at")
			VALUES ('ld-mem','DAY',$1,'ATTENDANCE',$2,2,'تجربة',$3)`, d.Format("2006-01-02"), pts, d.Add(12*time.Hour).UTC())
		if checkIn != "" {
			c, _ := time.ParseInLocation("2006-01-02 15:04", d.Format("2006-01-02")+" "+checkIn, debriefLoc)
			db.MustExec(`INSERT INTO "Attendance"(id,"employeeId","checkIn","checkOut",date) VALUES (gen_random_uuid()::text,'ld-mem',$1,$1::timestamp + interval '8 hours',$2)`, c.UTC(), d)
		}
	}
	att(5, 2, "09:05") // بوقته (ضمن السماح)
	att(4, 2, "09:40") // متأخر ٤٠ دقيقة
	att(3, 2, "10:00") // متأخر — راح يجيب عذر مقبول
	att(2, 0, "")      // غياب

	svc := NewLateExcuseService(db, nil)
	x, err := svc.Submit("ld-mem", day(3).Format("2006-01-02"), "زحمة طريق وتصادم", nil)
	if err != nil {
		t.Fatal(err)
	}
	// الغريب ورئيس الوحدة نفسه ما يگدرون؛ الغريب = مو موجود
	if _, err := svc.Decide("ld-out", "TECHNICIAN", x.ID, true, ""); err != ErrExcuseHidden {
		t.Fatalf("الغريب لازم 404: %v", err)
	}
	if _, err := svc.Decide("ld-mem", "TECHNICIAN", x.ID, true, ""); err != ErrExcuseHidden {
		t.Fatal("ما يوافق على عذره هو")
	}
	if in, _ := svc.Inbox("ld-unit", "TECHNICIAN"); len(in) != 1 {
		t.Fatalf("رئيس الوحدة يشوف عذر عضوه: %d", len(in))
	}
	if in, _ := svc.Inbox("ld-out", "TECHNICIAN"); len(in) != 0 {
		t.Fatal("الغريب ما يشوف")
	}
	if _, err := svc.Decide("ld-unit", "TECHNICIAN", x.ID, true, "مقبول"); err != nil {
		t.Fatal(err)
	}
	// مدير الإدارة يگدر يبت لرئيس الوحدة، ورئيس الوحدة ما يگدر لمديره
	y, _ := svc.Submit("ld-unit", today.Format("2006-01-02"), "مراجعة دائرة", nil)
	if !svc.CanDecide("ld-dir", "TECHNICIAN", y.ID) {
		t.Fatal("المدير يبت لرئيس الوحدة")
	}
	z, _ := svc.Submit("ld-dir", today.Format("2006-01-02"), "مراجعة دائرة", nil)
	if svc.CanDecide("ld-unit", "TECHNICIAN", z.ID) || !svc.CanDecide("owner-x", "OWNER", z.ID) {
		t.Fatal("عذر المدير يروح للمالك بس")
	}

	d := svc.DisciplineFor("ld-mem", today.AddDate(0, -1, 0), today.AddDate(0, 0, 1))
	if d == nil || d.Days != 4 || d.OnTime != 1 || d.Late != 1 || d.LateExcused != 1 || d.Absent != 1 || d.LateMinutes != 40 {
		t.Fatalf("الحساب غلط: %+v", d)
	}
	if *d.Score != 93 { // 100 − 2 − 5
		t.Fatalf("الدرجة: %v", *d.Score)
	}
	// تنبيه ماتركس عند ٣ تأخيرات بلا عذر — مرة وحدة للعتبة (نحتاج الأيام بنفس الشهر)
	if n.Day() >= 8 {
		att(6, 2, "09:45") // (11-13) السماحية نص ساعة
		att(7, 2, "09:50")
		claimed := map[string]bool{}
		claim := func(k, d string) (bool, error) {
			if claimed[k+d] {
				return false, nil
			}
			claimed[k+d] = true
			return true, nil
		}
		notif := NewLateExcuseService(db, repository.NewNotificationRepository(db))
		at := today.Add(15 * time.Hour)
		if _, err := notif.DisciplineAlerts(at, claim); err != nil {
			t.Fatal(err)
		}
		var got int
		db.Get(&got, `SELECT count(*) FROM "Notification" WHERE "employeeId" IN ('ld-mem','ld-unit','ld-dir') AND type = 'DISCIPLINE'`)
		if got != 3 { // الموظف + رئيس وحدته + مدير الإدارة
			t.Fatalf("التنبيه: %d", got)
		}
		notif.DisciplineAlerts(at, claim) //nolint:errcheck
		db.Get(&got, `SELECT count(*) FROM "Notification" WHERE "employeeId" IN ('ld-mem','ld-unit','ld-dir') AND type = 'DISCIPLINE'`)
		if got != 3 {
			t.Fatalf("ما يتكرر: %d", got)
		}
	}
	// قبل تاريخ التطبيق ما ينحسب شي
	DisciplineStart = today.AddDate(0, 0, 1)
	if d := svc.DisciplineFor("ld-mem", today.AddDate(0, -1, 0), today.AddDate(0, 0, 1)); d != nil {
		t.Fatalf("قبل التطبيق لازم فاضي: %+v", d)
	}
	// التقييم النهائي: ٨٥٪ + ١٥٪
	f := 80.0
	rows := []StaffScore{{ID: "a", Final: &f}}
	s := 60.0
	applyDiscipline(rows, map[string]*Discipline{"a": {Score: &s}})
	if *rows[0].Final != 77 {
		t.Fatalf("النهائي: %v", *rows[0].Final)
	}
}
