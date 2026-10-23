package handler

import (
	"testing"
	"time"
)

func TestClassifyLate(t *testing.T) {
	loc := time.UTC
	sched := time.Date(2026, 10, 1, 9, 0, 0, 0, loc)
	s := func(v string) *string { return &v }
	at := func(d int) *time.Time { x := sched.AddDate(0, 0, d); return &x }
	cases := []struct {
		r    lateRow
		want string
	}{
		{lateRow{Postpones: 1, Unstaffed: true}, "POSTPONED"},
		{lateRow{Unstaffed: true}, "NO_CREW"},
		{lateRow{Status: "PENDING"}, "NOT_CONFIRMED"},
		{lateRow{Status: "IN_PROGRESS", OpenMaterials: 2}, "MATERIALS"},
		{lateRow{Status: "IN_PROGRESS", WorkStopReason: s("المطر")}, "WORK_STOPPED"},
		{lateRow{Status: "COMPLETED", PartialCount: 2, StartedAt: at(0)}, "MULTI_DAY"},
		{lateRow{Status: "COMPLETED", ArrivedAt: at(2), StartedAt: at(2), CompletedAt: at(2)}, "LATE_START"},
		{lateRow{Status: "CONFIRMED"}, "NOT_STARTED"},
		{lateRow{Status: "COMPLETED", StartedAt: at(0), CompletedAt: at(1)}, "UNKNOWN"},
	}
	for i, c := range cases {
		c.r.ScheduledAt = sched
		c.r.ContactedAt = at(-1)
		if c.want == "NOT_CONFIRMED" {
			c.r.ContactedAt = nil
		}
		classifyLate(&c.r, loc)
		if c.r.Cause != c.want {
			t.Fatalf("case %d: got %s want %s", i, c.r.Cause, c.want)
		}
	}
	// بيانات مشكوكة: الموعد انغيّر بعد الإنجاز، ومنجز بلا مراحل
	r := lateRow{Status: "COMPLETED", ScheduledAt: sched, CompletedAt: at(1), LastChangeAt: at(2)}
	classifyLate(&r, loc)
	if len(r.DataIssues) != 2 {
		t.Fatalf("issues %v", r.DataIssues)
	}
}
