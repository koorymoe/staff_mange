package service

import (
	"testing"
	"time"
)

func TestResetDue(t *testing.T) {
	at := func(d, h int) time.Time { return time.Date(2026, 10, d, h, 0, 0, 0, debriefLoc) }
	cases := []struct {
		t    time.Time
		want bool
	}{{at(24, 23), false}, {at(25, 22), false}, {at(25, 23), true}, {at(26, 1), true}, {at(3, 23), false}}
	for _, c := range cases {
		if got := resetDue(c.t); got != c.want {
			t.Errorf("%v: got %v want %v", c.t, got, c.want)
		}
	}
}

// (11-27) فترة التقييم ٢٦ ← ٢٥، وتسميتها بشهر نهايتها.
func TestScorePeriod(t *testing.T) {
	at := func(m time.Month, d, h int) time.Time { return time.Date(2026, m, d, h, 0, 0, 0, debriefLoc) }
	key, from, to := MonthRange("", at(11, 27, 10))
	if key != "2026-12" || from.Format("2006-01-02") != "2026-11-26" || to.Format("2006-01-02") != "2026-12-26" {
		t.Fatalf("27/11 → %s %v %v", key, from, to)
	}
	if key, _, _ := MonthRange("", at(11, 25, 23)); key != "2026-11" {
		t.Fatalf("25/11 → %s", key)
	}
	if _, from, _ := MonthRange("2027-01", time.Now()); from.Format("2006-01-02") != "2026-12-26" {
		t.Fatalf("2027-01 from %v", from)
	}
}
