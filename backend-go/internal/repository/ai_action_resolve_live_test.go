package repository

import (
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// (دازاي 11-15) حل تذكير واحد ما يخفي التذكيرات الثانية الي انبعثت بنفس الدورة،
// والعدّ لكل مستلم يحسب تذكيراته هو بس.
func TestMarkReadOneAfter_OnlyItsOwnReminder_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()
	const emp = "emp_resolve_key_test"
	cleanup := func() {
		db.Exec(`DELETE FROM "Notification" WHERE "employeeId" = $1`, emp)
		db.Exec(`DELETE FROM "AiAction" WHERE "targetEmployeeId" = $1`, emp)
		db.Exec(`DELETE FROM "Employee" WHERE id = $1`, emp)
	}
	cleanup()
	defer cleanup()
	if _, err := db.Exec(`INSERT INTO "Employee" (id, name, username, password, role, status)
		VALUES ($1, 'اختبار المفتاح', 'resolve_key_test', 'x', 'TECHNICIAN', 'ACTIVE')`, emp); err != nil {
		t.Fatalf("employee: %v", err)
	}
	base := time.Now().Add(-time.Hour).UTC().Truncate(time.Millisecond)
	// ثلث تذكيرات بنفس الدقيقة: مهمة ١، مهمة ٢، حجز.
	for i, msg := range []string{"task A", "task B", "booking B1234"} {
		if _, err := db.Exec(`INSERT INTO "Notification" (id, "employeeId", type, message, "createdAt")
			VALUES (gen_random_uuid()::text, $1, 'AI_AUTOPILOT', $2, $3)`, emp, msg, base.Add(time.Duration(i)*time.Second)); err != nil {
			t.Fatalf("notif: %v", err)
		}
	}
	repo := NewNotificationRepository(db)
	// حل التذكير الثاني (فعله انسجّل قبل إشعاره بلحظة).
	if err := repo.MarkReadOneAfter(emp, "AI_AUTOPILOT", base.Add(500*time.Millisecond), 2*time.Minute); err != nil {
		t.Fatal(err)
	}
	var unread []string
	if err := db.Select(&unread, `SELECT message FROM "Notification" WHERE "employeeId" = $1 AND NOT read ORDER BY "createdAt"`, emp); err != nil {
		t.Fatal(err)
	}
	if len(unread) != 2 || unread[0] != "task A" || unread[1] != "booking B1234" {
		t.Fatalf("resolving task B must leave the others unread, got unread=%v", unread)
	}

	actions := NewAiActionRepository(db)
	for _, id := range []string{"t1", "t2", "t3"} {
		e := emp
		if _, err := db.Exec(`INSERT INTO "AiAction" (id, kind, "entityType", "entityId", period, "targetEmployeeId", "targetLabel", summary)
			VALUES (gen_random_uuid()::text, 'EXTRA_TASK_OVERDUE', 'EXTRA_TASK', $1, 'resolve-test', $2, 'x', 'x')`, "rk-"+id, e); err != nil {
			t.Fatalf("action: %v", err)
		}
	}
	n, err := actions.CountForTargetSince(emp, base)
	if err != nil || n != 3 {
		t.Fatalf("CountForTargetSince = %d, %v; want 3", n, err)
	}
}
