package service

import (
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/repository"
)

// يحتاج قاعدة حية: TEST_DATABASE_URL — بند تدقيق مو حجز حكم عليه مراقب لازم ينحسبله.
func TestMonitorAudit_Live(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	defer db.Close()
	s := &MatrixScoreService{repo: repository.NewMatrixScoreRepository(db)}
	got := []repository.MatrixScoreRow{}
	s.scoreMonitorAudit(time.Now(), func(r repository.MatrixScoreRow) { got = append(got, r) })
	for _, r := range got {
		t.Logf("%s %s %d/%d %s", r.EmployeeID, r.Rule, r.Points, r.MaxPoints, r.Reason)
	}
	if want := os.Getenv("TEST_AUDIT_ID"); want != "" {
		for _, r := range got {
			if r.SourceID == want {
				return
			}
		}
		t.Fatalf("%s ما انحسب", want)
	}
}
