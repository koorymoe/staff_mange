package repository

import (
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// TestMeetingRequest_Live «تذكيراتي»: الطلب ينحفظ، يطلع بالمنتظر، ينحسم مرة وحدة.
func TestMeetingRequest_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة — تخطي اختبار القاعدة الحية")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()
	r := NewMeetingRequestRepository(db)
	const emp = "__meet_emp"
	cleanup := func() {
		db.Exec(`DELETE FROM "MeetingRequest" WHERE "employeeId" = $1`, emp)
		db.Exec(`DELETE FROM "Employee" WHERE id = $1`, emp)
	}
	cleanup()
	defer cleanup()
	db.MustExec(`INSERT INTO "Employee"(id, name, status) VALUES ($1, 'طالب مقابلة', 'ACTIVE')`, emp)

	guest := "ضيف"
	m, err := r.Create(emp, "GUEST", "سبب", &guest, nil)
	if err != nil || m.Who != "طالب مقابلة" || m.Status != "PENDING" {
		t.Fatalf("Create: %v %+v", err, m)
	}
	if n, _ := r.OpenCount(emp); n != 1 {
		t.Fatalf("OpenCount = %d", n)
	}
	p, _ := r.Pending()
	found := false
	for _, x := range p {
		found = found || x.ID == m.ID
	}
	if !found {
		t.Fatal("الطلب ما طلع بالمنتظر")
	}
	at := time.Now().Add(24 * time.Hour)
	if err := r.Decide(m.ID, emp, "SCHEDULED", &at, "بالمكتب"); err != nil {
		t.Fatalf("Decide: %v", err)
	}
	if err := r.Decide(m.ID, emp, "DECLINED", nil, "لا"); err != ErrMeetingRequestDone {
		t.Fatalf("الحسم الثاني لازم يرفض: %v", err)
	}
	if mine, _ := r.Mine(emp); len(mine) != 1 || mine[0].Status != "SCHEDULED" || mine[0].ScheduledAt == nil {
		t.Fatalf("Mine: %+v", mine)
	}
}
