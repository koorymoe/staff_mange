package repository

import (
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// TestDaysSinceLastSignal_Live يحرس النصف الثاني من التصعيد
// بالاتجاهين (التساهل بعد فترة نظيفة، `ai_brain_service.go`): الرقم
// لازم يقيس الفجوة من **آخر** إشارة سابقة، يتجاهل الإشارة الحالية
// نفسها، ويرجع `nil` لو ماكو سجل سابق إطلاقاً — وإلا موظف بلا تاريخ
// يطلع بملاحظة «فترة نظيفة» بلا معنى.
func TestDaysSinceLastSignal_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة — تخطي اختبار القاعدة الحية")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect error: %v", err)
	}
	defer db.Close()

	const empID = "emp_escalation_test"
	cleanup := func() {
		db.Exec(`DELETE FROM "AiSignal" WHERE "employeeId" = $1`, empID)
		db.Exec(`DELETE FROM "Employee" WHERE id = $1`, empID)
	}
	cleanup()
	defer cleanup()

	if _, err := db.Exec(`
		INSERT INTO "Employee" (id, name, username, password, role, status)
		VALUES ($1, 'اختبار التصعيد', 'esc_test_user', 'x', 'TECHNICIAN', 'ACTIVE')
	`, empID); err != nil {
		t.Fatalf("تعذر إنشاء موظف الاختبار: %v", err)
	}

	repo := NewAiRepository(db)
	now := time.Now().UTC()

	// ماكو سجل سابق إطلاقاً — لازم nil.
	days, err := repo.DaysSinceLastSignal("WORK_STOPPED", empID, now)
	if err != nil {
		t.Fatalf("خطأ غير متوقع: %v", err)
	}
	if days != nil {
		t.Fatalf("ماكو سجل سابق — لازم nil، طلع %v", *days)
	}

	// إشارة قبل ٢٠ يوم بالضبط.
	insertSignal := func(id string, occurredAt time.Time) {
		if _, err := db.Exec(`
			INSERT INTO "AiSignal" (id, kind, "entityType", "entityId", "employeeId", payload, status, "occurredAt")
			VALUES ($1, 'WORK_STOPPED', 'BOOKING', 'bk-esc-test', $2, '{}'::jsonb, 'JUDGED', $3)
		`, id, empID, occurredAt); err != nil {
			t.Fatalf("تعذر إدخال إشارة الاختبار: %v", err)
		}
	}
	insertSignal("sig-esc-old", now.AddDate(0, 0, -20))

	days, err = repo.DaysSinceLastSignal("WORK_STOPPED", empID, now)
	if err != nil {
		t.Fatalf("خطأ غير متوقع: %v", err)
	}
	if days == nil || *days < 19 || *days > 21 {
		t.Fatalf("لازم ~٢٠ يوم، طلع %v", days)
	}

	// إشارة أقرب (قبل ٣ أيام) — لازم يرجع الأقرب مو الأقدم.
	insertSignal("sig-esc-recent", now.AddDate(0, 0, -3))
	days, err = repo.DaysSinceLastSignal("WORK_STOPPED", empID, now)
	if err != nil {
		t.Fatalf("خطأ غير متوقع: %v", err)
	}
	if days == nil || *days < 2 || *days > 4 {
		t.Fatalf("لازم يرجع أقرب إشارة (~٣ أيام)، طلع %v", days)
	}

	// إشارة تصير **بعد** نقطة المقارنة — ما تُحسب (نفس الإشارة الحالية
	// أو إشارة مستقبلية بالخطأ).
	future := now.AddDate(0, 0, -3) // نفس وقت المقارنة تماماً — ماكو "قبل"
	daysAtSameInstant, err := repo.DaysSinceLastSignal("WORK_STOPPED", empID, future)
	if err != nil {
		t.Fatalf("خطأ غير متوقع: %v", err)
	}
	// بهذي اللحظة، أقدم إشارة (٢٠ يوم) هي الوحيدة السابقة فعلاً.
	if daysAtSameInstant == nil || *daysAtSameInstant < 16 || *daysAtSameInstant > 18 {
		t.Fatalf("بنقطة أقدم، الإشارة الأحدث (٣ أيام) ما لازم تُحسب لأنها مو سابقة فعلاً — طلع %v", daysAtSameInstant)
	}
}
