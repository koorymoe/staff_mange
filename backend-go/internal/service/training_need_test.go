package service

import (
	"os"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
)

// ٩.٥ — الدراسة تشتغل لكل موظف فعّال بلا أخطاء وبلا راتب، والمتابعة تشتغل.
func TestTrainingStudyLive(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	defer db.Close()
	ids := []string{}
	if err := db.Select(&ids, `SELECT id FROM "Employee" WHERE status = 'ACTIVE'`); err != nil {
		t.Fatal(err)
	}
	for _, id := range ids {
		n, err := StudyTraining(db, id)
		if err != nil {
			t.Fatalf("%s: %v", id, err)
		}
		if n.Verdict == "" || strings.Contains(n.Evidence, "راتب") {
			t.Fatalf("bad %+v", n)
		}
	}
	if _, _, err := FollowUpTraining(db); err != nil {
		t.Fatal(err)
	}
}
