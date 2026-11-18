package repository

import (
	"os"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/model"
)

// (دازاي 11-18) توضيح البلاغ: بس صاحبه، وبس إذا «يحتاج توضيح»، ويرجع «جديد».
func TestBotReportClarify_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect error: %v", err)
	}
	defer db.Close()
	const a, b = "emp_clarify_a", "emp_clarify_b"
	cleanup := func() {
		db.Exec(`DELETE FROM "BotReport" WHERE "employeeId" IN ($1,$2)`, a, b)
		db.Exec(`DELETE FROM "Employee" WHERE id IN ($1,$2)`, a, b)
	}
	cleanup()
	defer cleanup()
	for i, id := range []string{a, b} {
		if _, err := db.Exec(`INSERT INTO "Employee" (id, name, username, password, role, status) VALUES ($1, 'اختبار توضيح', $2, 'x', 'TECHNICIAN', 'ACTIVE')`, id, "clarify_u"+string(rune('a'+i))); err != nil {
			t.Fatal(err)
		}
	}
	repo := NewBotReportRepository(db)
	rep := &model.BotReport{EmployeeID: a, Body: "الشاشة ما تفتح", Photos: []string{"/api/files/botreports/x.png"}, Source: "system"}
	if err := repo.Create(rep); err != nil {
		t.Fatal(err)
	}
	if ok, _ := repo.AddClarification(rep.ID, a, "توضيح", nil); ok {
		t.Fatal("clarified a NEW report")
	}
	q := "بأي شاشة؟"
	if err := repo.SetStatus(rep.ID, "NEEDS_INFO", &q); err != nil {
		t.Fatal(err)
	}
	if open, _ := repo.OpenNeedsInfo(a); open == nil || open.ID != rep.ID {
		t.Fatal("OpenNeedsInfo missed it")
	}
	if ok, _ := repo.AddClarification(rep.ID, b, "مو إله", nil); ok {
		t.Fatal("another employee clarified")
	}
	if ok, err := repo.AddClarification(rep.ID, a, "شاشة الحجوزات", []string{"/api/files/botreports/y.png"}); !ok || err != nil {
		t.Fatalf("clarify failed: %v", err)
	}
	got, _ := repo.Get(rep.ID)
	if got.Status != "NEW" || !strings.Contains(got.Body, "شاشة الحجوزات") || len(got.Photos) != 2 {
		t.Fatalf("got %+v", got)
	}
	if open, _ := repo.OpenNeedsInfo(a); open != nil {
		t.Fatal("still open")
	}
}
