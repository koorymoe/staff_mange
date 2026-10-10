package service

import (
	"database/sql"
	"testing"
	"time"

	"staffmange-api/internal/repository"
)

// الفني العادي مجموعته غير الليدر — وهذا الي يمنع طلب الفواتير منه.
func TestWatchGroupSeparatesTechsFromLeaders(t *testing.T) {
	cases := []struct {
		subj repository.WatchSubject
		want string
	}{
		{repository.WatchSubject{Role: "TECHNICIAN", IsLeader: true}, "LEADERS"},
		{repository.WatchSubject{Role: "TECHNICIAN"}, "TECHS"},
		{repository.WatchSubject{Role: "ENGINEER"}, "TECHS"},
		{repository.WatchSubject{Role: "DESIGNER"}, "DESIGN"},
		{repository.WatchSubject{Role: "MONITOR"}, "MONITORS"},
		{repository.WatchSubject{Role: "OWNER"}, "ADMINS"},
		{repository.WatchSubject{Role: "SALES", Perms: []string{"coordinator"}}, "COORDINATORS"},
	}
	for _, c := range cases {
		if got := WatchGroup(c.subj); got != c.want {
			t.Errorf("%+v: got %s want %s", c.subj, got, c.want)
		}
	}
}

func TestShiftWindowDefaultsAndLateness(t *testing.T) {
	e := &repository.ReportEmployee{Shift: sql.NullString{String: "MORNING", Valid: true}}
	s, en, assumed := shiftWindow(e)
	if s != "08:00" || en != "16:00" || !assumed {
		t.Fatalf("default morning window wrong: %s-%s %v", s, en, assumed)
	}
	e.ShiftStart = sql.NullString{String: "09:00", Valid: true}
	e.ShiftEnd = sql.NullString{String: "17:00", Valid: true}
	if s, _, assumed := shiftWindow(e); s != "09:00" || assumed {
		t.Fatalf("explicit window ignored: %s %v", s, assumed)
	}
	day := time.Date(2026, 10, 4, 0, 0, 0, 0, debriefLoc)
	checkIn := time.Date(2026, 10, 4, 9, 25, 0, 0, debriefLoc)
	if late := mins(checkIn.Sub(atClock(day, "09:00"))); late != 25 {
		t.Fatalf("late minutes = %d, want 25", late)
	}
	if durText(90) != "1 ساعة و30 دقيقة" || durText(45) != "45 دقيقة" {
		t.Fatalf("durText wrong: %s / %s", durText(90), durText(45))
	}
}
