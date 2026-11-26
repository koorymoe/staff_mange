package repository

import (
	"testing"
	"time"
)

func TestRemotePeriod(t *testing.T) {
	d := func(s string) time.Time { x, _ := time.Parse("2006-01-02", s); return x }
	cases := map[string]string{"2026-11-25": "2026-11", "2026-11-26": "2026-12", "2026-12-25": "2026-12", "2026-12-31": "2027-01", "2026-12-01": "2026-12"}
	for day, want := range cases {
		if got := PeriodOf(d(day), 26); got != want {
			t.Errorf("PeriodOf(%s)=%s want %s", day, got, want)
		}
	}
	from, to := PeriodRange("2026-12", 26)
	if from.Format("2006-01-02") != "2026-11-26" || to.Format("2006-01-02") != "2026-12-25" {
		t.Fatalf("range %v %v", from, to)
	}
	from, to = PeriodRange("2027-01", 26)
	if from.Format("2006-01-02") != "2026-12-26" || to.Format("2006-01-02") != "2027-01-25" {
		t.Fatalf("jan range %v %v", from, to)
	}
	if f, l := PeriodRange("2026-02", 1); f.Format("01-02") != "02-01" || l.Format("01-02") != "02-28" {
		t.Fatalf("calendar %v %v", f, l)
	}
	loc, _ := time.LoadLocation("Asia/Baghdad")
	if got := CloseMoment("2026-11", DefaultRemoteHoursConfig, loc).Format("2006-01-02 15"); got != "2026-11-25 23" {
		t.Fatalf("close %s", got)
	}
	c := RemoteHoursConfig{PeriodStartDay: 26, CloseDay: 1, CloseHour: 10}
	if got := CloseMoment("2026-11", c, loc).Format("2006-01-02 15"); got != "2026-12-01 10" {
		t.Fatalf("close after %s", got)
	}
}
