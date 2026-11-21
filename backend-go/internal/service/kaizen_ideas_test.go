package service

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/database"
)

// (دازاي 11-21) أفكار ١ و٣ و٤: أثرك اليوم، سهم التحسّن، عندي فكرة.

func TestDailyImpactMessageGoodFirst(t *testing.T) {
	w := "حجز B1 · التثبيت: متأخرة"
	m := DailyImpactMessage(DailyImpact{Full: 8, Partial: 1, Lost: 1, Earned: 17, Max: 20, Worst: &w})
	if !strings.HasPrefix(m, "📊 أثرك اليوم: 85٪ — ✅ 8") || !strings.Contains(m, "❌ 1 فاتت") || !strings.Contains(m, "اعتذر") {
		t.Fatalf("%s", m)
	}
	if m := DailyImpactMessage(DailyImpact{Full: 5, Earned: 10, Max: 10}); !strings.Contains(m, "عاشت إيدك") {
		t.Fatalf("%s", m)
	}
}

func TestTrendAndTopImprover(t *testing.T) {
	f := func(v float64) *float64 { return &v }
	a := StaffScore{Name: "أ", Final: f(80)}
	applyTrend(&a, 60+70+65, 3) // معدّل ٦٥ → +١٥
	b := StaffScore{Name: "ب", Final: f(90)}
	applyTrend(&b, 88, 1) // +٢
	c := StaffScore{Name: "ج", Final: f(50)}
	applyTrend(&c, 0, 0) // ماكو سابق
	if a.Trend == nil || *a.Trend != 15 || c.Trend != nil {
		t.Fatalf("trend a=%v c=%v", a.Trend, c.Trend)
	}
	if n, d := TopImprover([]StaffScore{a, b, c}); n != "أ" || d != 15 {
		t.Fatalf("%s %v", n, d)
	}
	if n, _ := TopImprover([]StaffScore{b}); n != "" {
		t.Fatalf("+٢ ما يعتبر تحسّن: %s", n)
	}
}

func TestIdeaAcceptedIsExtraWork(t *testing.T) {
	if RuleFitsGroup("IDEA", GroupTechs) {
		t.Fatal("الفكرة لازم تنحسب شغل إضافي مو بند دور")
	}
}

type ideaNotif struct{ got []string }

func (n *ideaNotif) Create(id, typ, msg string) error { n.got = append(n.got, id+"|"+msg); return nil }

func TestIdeasLive(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	emp, boss := uuid.NewString(), uuid.NewString()
	for _, id := range []string{emp, boss} {
		db.MustExec(`INSERT INTO "Employee" (id, name, username, password, role, phone) VALUES ($1, $2, $2, 'x', 'TECHNICIAN', $3)`,
			id, "idea_"+id[:6], "079"+id[:8])
	}
	t.Cleanup(func() {
		db.Exec(`DELETE FROM "MatrixScore" WHERE "employeeId" = $1`, emp) //nolint:errcheck
		db.Exec(`DELETE FROM "Employee" WHERE id IN ($1, $2)`, emp, boss) //nolint:errcheck
	})
	n := &ideaNotif{}
	s := NewIdeaService(db, n)
	if err := s.Submit(emp, "قصيرة", "TIME"); err != ErrIdeaInvalid {
		t.Fatalf("القصيرة انقبلت: %v", err)
	}
	if err := s.Submit(emp, "نخلي عدّة الكاميرات بصندوق بكل سيارة", "TIME"); err != nil {
		t.Fatal(err)
	}
	mine, _ := s.Mine(emp)
	if len(mine) != 1 || mine[0].Status != "NEW" {
		t.Fatalf("%+v", mine)
	}
	id := mine[0].ID
	// موظف عادي ما بالهيكل = غريب → 404
	if err := s.Decide(boss, "TECHNICIAN", id, "ACCEPTED", ""); err != ErrReviewHidden {
		t.Fatalf("الغريب رد: %v", err)
	}
	if err := s.Decide(boss, "ADMIN", id, "REJECTED", ""); err == nil {
		t.Fatal("«لا» بلا سبب انقبلت")
	}
	if err := s.Decide(boss, "ADMIN", id, "ACCEPTED", "ممتازة"); err != nil {
		t.Fatal(err)
	}
	var pts int
	if err := db.Get(&pts, `SELECT points FROM "MatrixScore" WHERE "employeeId" = $1 AND source = 'IDEA' AND "sourceId" = $2`, emp, id); err != nil || pts != 2 {
		t.Fatalf("نقاط الفكرة: %d %v", pts, err)
	}
	if len(n.got) != 1 || !strings.Contains(n.got[0], "نطبقها") {
		t.Fatalf("الإشعار: %v", n.got)
	}
	// الفكرة الي انرد عليها ما تنرد مرة ثانية
	if err := s.Decide(boss, "ADMIN", id, "REJECTED", "لا"); err != ErrReviewHidden {
		t.Fatalf("انردت مرتين: %v", err)
	}
	// أثرك اليوم يشوف نقطة الفكرة
	rows, err := DailyImpacts(db, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, r := range rows {
		if r.EmployeeID == emp && r.Full >= 1 {
			found = true
		}
	}
	if !found {
		t.Fatal("أثرك اليوم ما شاف نقطة اليوم")
	}
}
