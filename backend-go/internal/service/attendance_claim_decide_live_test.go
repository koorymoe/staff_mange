package service

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
	"staffmange-api/internal/repository"
)

// (11-07) طلب «چنت أشتغل»: الطلب المرفوض ما يمدّد الانصراف، والقرار الثاني ينرفض.
// چان التمديد يصير قبل الحسم، فموافقة بعد رفض (أو بنفس اللحظة) تمد الساعات.
func TestAttendanceClaimDecideOnceLive(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	svc := NewAttendanceService(repository.NewAttendanceRepository(db))
	emp, boss, att := uuid.NewString(), uuid.NewString(), uuid.NewString()
	sfx := uuid.NewString()[:6]
	db.MustExec(`INSERT INTO "Employee" (id, name, username, role, status) VALUES ($1, 'فني دوام', $2, 'TECHNICIAN', 'ACTIVE'), ($3, 'مراقب', $4, 'ADMIN', 'ACTIVE')`, emp, "ac_"+sfx, boss, "ab_"+sfx)
	t.Cleanup(func() { db.Exec(`DELETE FROM "Employee" WHERE id IN ($1, $2)`, emp, boss) }) //nolint:errcheck
	db.MustExec(`INSERT INTO "Attendance" (id, "employeeId", "checkIn", "checkOut", date)
		VALUES ($1, $2, now() AT TIME ZONE 'UTC' - interval '9 hours', now() AT TIME ZONE 'UTC' - interval '3 hours', baghdad_today())`, att, emp)
	claim := uuid.NewString()
	db.MustExec(`INSERT INTO "AttendanceClaim" (id, "attendanceId", "employeeId", kind, "autoAt", "claimedUntil", status)
		VALUES ($1, $2, $3, 'WORKED', now() - interval '3 hours', now() - interval '1 hour', 'PENDING')`, claim, att, emp)
	checkout := func() string {
		var s string
		if err := db.Get(&s, `SELECT "checkOut"::text FROM "Attendance" WHERE id = $1`, att); err != nil {
			t.Fatal(err)
		}
		return s
	}
	before := checkout()
	if err := svc.DecideClaim(claim, boss, false); err != nil {
		t.Fatal(err)
	}
	// قرار ثاني (موافقة) بعد الرفض: لازم ينرفض وما يلمس الانصراف
	if err := svc.DecideClaim(claim, boss, true); err == nil {
		t.Fatal("موافقة بعد الرفض انقبلت")
	}
	if got := checkout(); got != before {
		t.Fatalf("الانصراف تغيّر بعد الرفض: %s → %s", before, got)
	}
	// المسار الطبيعي: موافقة على طلب معلّق تمدّد الانصراف
	db.MustExec(`UPDATE "AttendanceClaim" SET status = 'PENDING', "decidedById" = NULL, "decidedAt" = NULL WHERE id = $1`, claim)
	if err := svc.DecideClaim(claim, boss, true); err != nil {
		t.Fatal(err)
	}
	if got := checkout(); got == before {
		t.Fatal("الموافقة ما مدّدت الانصراف")
	}
}
