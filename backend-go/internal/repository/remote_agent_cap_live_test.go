package repository

import (
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// (11-17) طلبات وكيل متزامنة بمفاتيح مختلفة ما تعبر حد اليوم (٦ ساعات).
func TestAgentDayCapConcurrent_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()
	const emp = "agent-cap-emp"
	clean := func() {
		db.Exec(`DELETE FROM "RemoteWorkEntry" WHERE "employeeId" = $1`, emp)
		db.Exec(`DELETE FROM "Employee" WHERE id = $1`, emp)
	}
	clean()
	t.Cleanup(clean)
	db.MustExec(`INSERT INTO "Employee"(id, name) VALUES ($1, 'وكيل حد')`, emp)
	r := NewRemoteHoursRepository(db)

	// ٤ طلبات × ساعتين بنفس اللحظة = ٨ ساعات — الحد ٦، فلازم ٣ بس تعبر
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			note := fmt.Sprintf("🤖 Claude [cap-%d]: اختبار", i)
			_, _ = r.CreateFrom(emp, "2026-11-17", 2*3600, &note, "", "AGENT")
		}(i)
	}
	wg.Wait()
	if got := r.AgentDaySeconds(emp, "2026-11-17"); got > AgentDayCap {
		t.Fatalf("عبر الحد: %d ثانية", got)
	} else if got != 3*2*3600 {
		t.Fatalf("المفروض ٦ ساعات بالضبط، طلع %d", got)
	}
}
