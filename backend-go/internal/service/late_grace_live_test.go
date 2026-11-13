package service

import (
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/database"
)

// (دازاي 11-13) «اريد ماتركس ينطي سماحية بالتأخير نص ساعة»: دوام ٨، حضور ٨:٢٥ = بوقته،
// وحضور ٨:٣٥ = متأخر.
func TestLateGraceHalfHour_Live(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	emps := []string{"lg-ok", "lg-late"}
	clean := func() {
		for _, e := range emps {
			db.Exec(`DELETE FROM "MatrixScore" WHERE "employeeId" = $1`, e)
			db.Exec(`DELETE FROM "Attendance" WHERE "employeeId" = $1`, e)
			db.Exec(`DELETE FROM "Employee" WHERE id = $1`, e)
		}
	}
	clean()
	t.Cleanup(clean)
	day := time.Now().In(debriefLoc).AddDate(0, 0, -1)
	day = time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, debriefLoc)
	old := DisciplineStart
	DisciplineStart = day.AddDate(0, 0, -7)
	t.Cleanup(func() { DisciplineStart = old })
	for i, e := range emps {
		db.MustExec(`INSERT INTO "Employee"(id, name, role, status, "shiftStart", "shiftEnd") VALUES ($1, $1, 'TECHNICIAN', 'ACTIVE', '08:00', '16:00')`, e)
		in := day.Add(8*time.Hour + []time.Duration{25 * time.Minute, 35 * time.Minute}[i])
		db.MustExec(`INSERT INTO "Attendance"(id, "employeeId", "checkIn", date) VALUES (gen_random_uuid()::text, $1, ($2::timestamptz AT TIME ZONE 'UTC'), $3::date)`, e, in, day.Format("2006-01-02"))
		db.MustExec(`INSERT INTO "MatrixScore"("employeeId", source, "sourceId", rule, points, "maxPoints", "at", reason) VALUES ($1, 'DAY', $2, 'ATTENDANCE', 2, 2, $3, 'x')`, e, day.Format("2006-01-02"), day.Add(12*time.Hour))
	}
	d, err := Disciplines(db, emps, day, day.AddDate(0, 0, 1), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if d["lg-ok"] == nil || d["lg-ok"].Late != 0 {
		t.Fatalf("٨:٢٥ ضمن السماحية: %+v", d["lg-ok"])
	}
	if d["lg-late"] == nil || d["lg-late"].Late != 1 {
		t.Fatalf("٨:٣٥ متأخر: %+v", d["lg-late"])
	}
}
