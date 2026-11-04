package service

import (
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/repository"
)

// يحتاج قاعدة حية: كل استعلامات الواجب تشتغل بلا خطأ لكل موظف فعّال (ما تنبلع بصمت).
func TestRatingDuty_QueriesRun(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	defer db.Close()
	d := NewRatingDuty(db)
	ids := []string{}
	_ = db.Select(&ids, `SELECT id FROM "Employee" WHERE status = 'ACTIVE'`)
	for _, id := range ids {
		for off := -3; off <= 0; off++ {
			if _, err := d.Owed(id, companyDay(off)); err != nil {
				t.Fatalf("Owed(%s): %v", id, err)
			}
		}
	}
	if _, err := d.Sweep(); err != nil {
		t.Fatalf("Sweep: %v", err)
	}
}

// سجل الإهمال يصير نقاط RATING_DUTY: عوّض = ١، ما عوّض = ٠ (من ٢).
func TestRatingDuty_Score(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	defer db.Close()
	s := &MatrixScoreService{repo: repository.NewMatrixScoreRepository(db)}
	var open, done int
	_ = db.Get(&open, `SELECT count(*) FROM "EmployeeNegligence" WHERE "compensatedAt" IS NULL`)
	_ = db.Get(&done, `SELECT count(*) FROM "EmployeeNegligence" WHERE "compensatedAt" IS NOT NULL`)
	got0, got1 := 0, 0
	s.scoreRatingDuty(time.Now(), func(r repository.MatrixScoreRow) {
		if r.Rule != "RATING_DUTY" || r.MaxPoints != 2 {
			t.Fatalf("row %+v", r)
		}
		if r.Points == 0 {
			got0++
		} else if r.Points == 1 {
			got1++
		}
	})
	if got0 != open || got1 != done {
		t.Fatalf("0=%d/%d 1=%d/%d", got0, open, got1, done)
	}
	t.Logf("ما عوّض %d · عوّض %d", got0, got1)
}
