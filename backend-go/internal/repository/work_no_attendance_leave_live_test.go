package repository

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// بلاغ #49 (11-17): موظف مجاز بإجازة معتمدة ما ينسجّل «شغل بلا حضور» حتى لو حجز باسمه بدا ذاك اليوم.
func TestWorkNoAttendanceSkipsLeave_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()
	const emp, cu, bk, lv = "wna-emp", "wna-cu", "wna-bk", "wna-lv"
	clean := func() {
		db.Exec(`DELETE FROM "LeaveRequest" WHERE id = $1`, lv)
		db.Exec(`DELETE FROM "BookingAssignment" WHERE "bookingId" = $1`, bk)
		db.Exec(`DELETE FROM "Mission" WHERE "bookingId" = $1`, bk)
		db.Exec(`DELETE FROM "Booking" WHERE id = $1`, bk)
		db.Exec(`DELETE FROM "Customer" WHERE id = $1`, cu)
		db.Exec(`DELETE FROM "Employee" WHERE id = $1`, emp)
	}
	clean()
	t.Cleanup(clean)
	db.MustExec(`INSERT INTO "Employee"(id, name, role) VALUES ($1, 'فني إجازة', 'TECHNICIAN')`, emp)
	db.MustExec(`INSERT INTO "Customer"(id, name, phone) VALUES ($1, 'x', '07700004949')`, cu)
	db.MustExec(`INSERT INTO "Booking"(id, code, "customerId", status, "startedAt") VALUES ($1, 'WNA-1', $2, 'IN_PROGRESS', now())`, bk, cu)
	db.MustExec(`INSERT INTO "BookingAssignment"(id, "bookingId", "employeeId", role) VALUES ('wna-ba', $1, $2, 'TECH_1')`, bk, emp)
	var day string
	db.Get(&day, `SELECT baghdad_today()::text`)
	r := NewAiRepository(db)

	has := func() bool {
		rows, err := r.AttendanceWorkGaps(day, 360, emp)
		if err != nil {
			t.Fatal(err)
		}
		for _, x := range rows {
			if x.Kind == "WORK_NO_ATTENDANCE" {
				return true
			}
		}
		return false
	}
	if !has() {
		t.Fatal("بلا إجازة لازم ينسجّل شغل بلا حضور")
	}
	db.MustExec(`INSERT INTO "LeaveRequest"(id, "employeeId", "startDate", "endDate", status, route, reason, "createdAt")
		VALUES ($1, $2, baghdad_today(), baghdad_today(), 'APPROVED', 'ADMIN', 'اختبار', now())`, lv, emp)
	if has() {
		t.Fatal("المجاز ما لازم ينسجّل شغل بلا حضور")
	}
}
