package service

import (
	"testing"
	"time"

	"staffmange-api/internal/repository"
)

func TestStationPoints(t *testing.T) {
	cases := map[string]int{ChainOK: 2, ChainLate: 1, ChainIssue: 1, ChainMissed: 0}
	for st, want := range cases {
		if got, ok := StationPoints(st); !ok || got != want {
			t.Fatalf("%s → %d %v", st, got, ok)
		}
	}
	for _, st := range []string{ChainWaiting, ChainNA} {
		if _, ok := StationPoints(st); ok {
			t.Fatalf("%s لازم ما تنحسب", st)
		}
	}
}

func TestTaskPoints(t *testing.T) {
	due := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	at := func(h int) *time.Time { x := due.Add(time.Duration(h) * time.Hour); return &x }
	if p, _ := TaskPoints(due, at(-1), due); p != 2 {
		t.Fatal("بموعدها = ٢")
	}
	if p, _ := TaskPoints(due, at(20), due); p != 1 {
		t.Fatal("خلال يوم = ١")
	}
	if p, _ := TaskPoints(due, at(30), due); p != 0 {
		t.Fatal("بعد يوم = ٠")
	}
	if _, ok := TaskPoints(due, nil, due.Add(2*time.Hour)); ok {
		t.Fatal("لسه بمهلة اليوم: ما تنحسب")
	}
	if p, ok := TaskPoints(due, nil, due.Add(30*time.Hour)); !ok || p != 0 {
		t.Fatal("ما خلصت = ٠")
	}
}

func TestFinalScore(t *testing.T) {
	if f, _ := FinalScore(80, true, 60, true); f < 71.99 || f > 72.01 {
		t.Fatalf("0.6*80+0.4*60 = 72, got %v", f)
	}
	if f, _ := FinalScore(80, true, 0, false); f != 80 {
		t.Fatal("بلا بشر = ماتركس")
	}
	if _, ok := FinalScore(0, false, 0, false); ok {
		t.Fatal("بلا شي = ماكو")
	}
}

func TestBuildScores(t *testing.T) {
	now := time.Now()
	people := []repository.Scorable{{ID: "a", Name: "A"}, {ID: "b", Name: "B"}}
	pts := []repository.MatrixScoreRow{
		{EmployeeID: "a", Source: "BOOKING", Rule: "PAPER", Points: 1, MaxPoints: 2},
		{EmployeeID: "a", Source: "DAY", Rule: "ATTENDANCE", Points: 2, MaxPoints: 2},
		{EmployeeID: "a", Source: "DAY", Rule: "ATTENDANCE", Points: 0, MaxPoints: 2, CancelledAt: &now},
	}
	hum := []repository.HumanRow{{RateeID: "a", Score: 4}, {RateeID: "a", Score: 2}}
	out := BuildScores(people, pts, hum)
	a := out[0]
	if a.Earned != 3 || a.Max != 4 || *a.MatrixPct != 75 || *a.HumanPct != 60 {
		t.Fatalf("%+v", a)
	}
	if *a.Final < 68.99 || *a.Final > 69.01 { // 0.6*75 + 0.4*60
		t.Fatalf("final %v", *a.Final)
	}
	if len(a.TopLosses) != 1 || a.TopLosses[0].Rule != "PAPER" {
		t.Fatalf("losses %+v", a.TopLosses)
	}
	if out[1].Final != nil {
		t.Fatal("B بلا شي")
	}
}

func TestPeriodKey(t *testing.T) {
	if k := PeriodKey(time.Date(2026, 10, 3, 12, 0, 0, 0, debriefLoc)); k != "2026-10-A" {
		t.Fatal(k)
	}
	if k := PeriodKey(time.Date(2026, 10, 20, 12, 0, 0, 0, debriefLoc)); k != "2026-10-B" {
		t.Fatal(k)
	}
}
