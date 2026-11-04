package service

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
)

func TestAnalyzeField(t *testing.T) {
	now := time.Date(2026, 11, 2, 10, 0, 0, 0, time.UTC)
	at := func(m int) *time.Time { x := now.Add(-time.Duration(m) * time.Minute); return &x }
	road, work := 30.0, 60.0
	f := FieldTime{ScheduledAt: at(120), DepartedAt: at(180), ArrivedAt: at(90), StartedAt: at(100 - 0), RoadMin: &road, WorkMin: &work}
	f.StartedAt = at(50)
	analyzeField(&f, now)
	got := strings.Join(f.Notes, " | ")
	for _, want := range []string{"الطريق 1س 30د", "وصل متأخر 30د", "انتظر 40د", "يشتغل من 50د"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %s", want, got)
		}
	}
	if !f.Alert {
		t.Fatal("expected alert")
	}
}

func TestFieldTimingLive(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	defer db.Close()
	if _, err := FieldTiming(db, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := FieldTiming(db, "DT9"); err != nil {
		t.Fatal(err)
	}
}
