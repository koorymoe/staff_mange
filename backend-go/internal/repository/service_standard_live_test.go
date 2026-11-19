package repository

import (
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// (دازاي 11-18) استعلامات الوقت القياسي تشتغل على قاعدة حقيقية، والتعديل يغلب المتعلَّم.
func TestServiceStandardsLive(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db := sqlx.MustConnect("postgres", dsn)
	defer db.Close()
	r := NewMatrixScoreRepository(db)
	if _, err := r.LeaderJobs(time.Now().AddDate(0, -1, 0), time.Now(), ""); err != nil {
		t.Fatal(err)
	}
	rows, err := r.ServiceStandards()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) == 0 {
		t.Skip("ماكو خدمات بالقاعدة المحلية")
	}
	id := rows[0].ServiceID
	t.Cleanup(func() { db.Exec(`DELETE FROM "ServiceStandardTime" WHERE "serviceId" = $1`, id) })
	m := 90
	if err := r.SetServiceStandard(id, &m, ""); err != nil {
		t.Fatal(err)
	}
	exp, _ := r.ExpectedMinutes()
	if exp[id] != 90 {
		t.Fatalf("override ما غلب: %v", exp[id])
	}
	if err := r.SetServiceStandard(id, nil, ""); err != nil {
		t.Fatal(err)
	}
}

func TestCoordFactsLive(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db := sqlx.MustConnect("postgres", dsn)
	defer db.Close()
	from := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	if _, err := NewMatrixScoreRepository(db).CoordFacts(from, from.AddDate(0, 1, 0), "2026-11", `i."manualProgress"`, ""); err != nil {
		t.Fatal(err)
	}
}
