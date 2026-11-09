package service

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/repository"
)

// ٩.٤ — المطابقة تشتغل لكل تقرير آخر ٣ أيام بلا أخطاء، وبلا راتب.
func TestReportMatchLive(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	defer db.Close()
	m := NewReportMatcher(db)
	for back := 0; back < 3; back++ {
		day := time.Now().AddDate(0, 0, -back).Format("2006-01-02")
		emps := []string{}
		if err := db.Select(&emps, `SELECT "employeeId" FROM "DailyWorkReport" WHERE day = $1::date`, day); err != nil {
			t.Fatal(err)
		}
		for _, e := range emps {
			r, err := m.Match(e, day)
			if err != nil {
				t.Fatalf("%s %s: %v", e, day, err)
			}
			if r == nil || r.Status == "" || strings.Contains(r.Note, "راتب") {
				t.Fatalf("bad result %+v", r)
			}
		}
	}
}

func TestCodeToken(t *testing.T) {
	got := codeToken.FindAllString("خلصت DT9 و B-12 و SIM1A ورحت للمخزن 2026", -1)
	if strings.Join(got, ",") != "DT9,B-12,SIM1A" {
		t.Fatalf("got %v", got)
	}
}

// المطابقة تصير نقاط REPORT_MATCH (من ٢)، والمبرّر = ٢.
func TestReportMatchScore(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	defer db.Close()
	if _, err := NewReportMatcher(db).SweepDay(time.Now().AddDate(0, 0, -1).Format("2006-01-02")); err != nil {
		t.Fatal(err)
	}
	s := &MatrixScoreService{repo: repository.NewMatrixScoreRepository(db)}
	s.scoreReportMatch(time.Now(), func(r repository.MatrixScoreRow) {
		if r.Rule != "REPORT_MATCH" || r.MaxPoints != 2 || r.Points < 0 || r.Points > 2 {
			t.Fatalf("bad row %+v", r)
		}
		t.Logf("%s %d/2 %s", r.EmployeeID[:6], r.Points, r.Reason)
	})
}
