package service

import (
	"testing"
	"time"
)

// يحرس حدود احتساب الدوام: ساعة الحد حسب وقت الحضور، والتمديد بآخر نشاط، وقراءة الساعة.
func TestParseClock(t *testing.T) {
	s := func(v string) *string { return &v }
	cases := []struct {
		in     *string
		h, m   int
		wantOK bool
	}{
		{nil, 0, 0, false}, {s("08:30"), 8, 30, true}, {s(" 16:00 "), 16, 0, true}, {s("24:00"), 24, 0, true},
		{s("25:00"), 0, 0, false}, {s("10:60"), 0, 0, false}, {s("abc"), 0, 0, false}, {s("-1:00"), 0, 0, false},
	}
	for _, c := range cases {
		h, m, ok := parseClock(c.in)
		if ok != c.wantOK || (ok && (h != c.h || m != c.m)) {
			t.Errorf("parseClock(%v) = %d,%d,%v", c.in, h, m, ok)
		}
	}
}

func TestCountCapAndLimit(t *testing.T) {
	at := func(h, m int) time.Time { return time.Date(2026, 11, 6, h, m, 0, 0, debriefLoc) }
	base := at(0, 0)
	for _, c := range []struct {
		in   time.Time
		want time.Time
	}{
		{at(8, 0), base.Add(16 * time.Hour)}, // صباحي: الحد ٤ العصر
		{at(14, 59), base.Add(16 * time.Hour)},
		{at(15, 0), base.Add(24 * time.Hour)}, // مسائي: الحد نص الليل
		{at(2, 0), base},                      // بعد نص الليل: الحد بداية اليوم
	} {
		if got := CountCap(c.in); !got.Equal(c.want) {
			t.Errorf("CountCap(%v) = %v، المتوقع %v", c.in, got, c.want)
		}
	}
	// آخر نشاط بعد الحد يمدّه
	late := at(18, 30)
	if got := countLimit(at(8, 0), &late); !got.Equal(late) {
		t.Errorf("التمديد بآخر نشاط: %v", got)
	}
	// آخر نشاط قبل الحد ما يقصّره
	early := at(10, 0)
	if got := countLimit(at(8, 0), &early); !got.Equal(at(16, 0)) {
		t.Errorf("نشاط قبل الحد قصّره: %v", got)
	}
	// الحد ما ينزل تحت وقت الحضور
	if got := countLimit(at(2, 0), nil); !got.Equal(at(2, 0)) {
		t.Errorf("الحد نزل تحت الحضور: %v", got)
	}
}

func TestClockLabel(t *testing.T) {
	if got := clockLabel(time.Date(2026, 11, 7, 0, 0, 0, 0, debriefLoc)); got != "١٢ بالليل" {
		t.Errorf("%q", got)
	}
	if got := clockLabel(time.Date(2026, 11, 6, 16, 5, 0, 0, debriefLoc)); got != "16:05" {
		t.Errorf("%q", got)
	}
}
