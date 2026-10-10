package service

import (
	"testing"
	"time"
)

func sp(s string) *string { return &s }

func TestShiftWindow(t *testing.T) {
	day := time.Date(2026, 10, 6, 10, 0, 0, 0, debriefLoc)
	f, to := ShiftWindow(nil, nil, nil, day)
	if f.Hour() != 8 || to.Hour() != 16 {
		t.Fatalf("الصباحي الافتراضي ٨–٤: %v %v", f, to)
	}
	f, to = ShiftWindow(sp("EVENING"), nil, nil, day)
	if f.Hour() != 16 || to.Day() != 7 || to.Hour() != 0 {
		t.Fatalf("المسائي ٤–١٢ يعبر نص الليل: %v %v", f, to)
	}
	f, to = ShiftWindow(nil, sp("09:30"), sp("17:00"), day)
	if f.Hour() != 9 || f.Minute() != 30 || to.Hour() != 17 {
		t.Fatalf("وقت الموظف المسجّل: %v %v", f, to)
	}
}

func TestCountedEnd(t *testing.T) {
	day := time.Date(2026, 10, 2, 0, 0, 0, 0, debriefLoc)
	at := func(h, m int) time.Time { return day.Add(time.Duration(h)*time.Hour + time.Duration(m)*time.Minute) }
	now := day.AddDate(0, 0, 3)
	// ٨ الصبح وبلا انصراف لحد بعد يومين → لحد ٤ العصر.
	if e, ok, why := CountedEnd(at(8, 0), nil, nil, now); !ok || !e.Equal(at(16, 0)) || why == "" {
		t.Fatalf("morning: %v %v %q", e, ok, why)
	}
	// ٣:١٥ العصر وانصراف بعد يومين (الـ٥٥ ساعة) → لحد ١٢ بالليل.
	out := at(15, 15).Add(55 * time.Hour)
	if e, ok, _ := CountedEnd(at(15, 15), &out, nil, now); !ok || !e.Equal(at(24, 0)) {
		t.Fatalf("55h: %v", e)
	}
	// ٨:٥٠ بالليل → لحد ١٢ بالليل.
	if e, _, _ := CountedEnd(at(20, 50), nil, nil, now); !e.Equal(at(24, 0)) {
		t.Fatalf("night: %v", e)
	}
	// حجز خلص ٧:٣٠ المسا لحضور صباحي → لحده.
	last := at(19, 30)
	if e, _, _ := CountedEnd(at(8, 0), nil, &last, now); !e.Equal(last) {
		t.Fatalf("booking: %v", e)
	}
	// انصراف فعلي ٢ الظهر → كما هو، بلا تعليم.
	o := at(14, 0)
	if e, ok, _ := CountedEnd(at(8, 0), &o, nil, now); ok || !e.Equal(o) {
		t.Fatalf("real: %v %v", e, ok)
	}
	// بعده بالدوام: ما ينقطع.
	if e, ok, _ := CountedEnd(at(8, 0), nil, nil, at(17, 0)); ok || !e.Equal(at(17, 0)) {
		t.Fatalf("live: %v %v", e, ok)
	}
}

func TestAutoCheckoutAt(t *testing.T) {
	day := time.Date(2026, 10, 6, 0, 0, 0, 0, debriefLoc)
	in := day.Add(8 * time.Hour)
	end := day.Add(16 * time.Hour)
	if _, ok := AutoCheckoutAt(in, nil, end.Add(2*time.Hour)); ok {
		t.Fatal("قبل ٣ ساعات ما يسكّر")
	}
	if at, ok := AutoCheckoutAt(in, nil, end.Add(4*time.Hour)); !ok || !at.Equal(end) {
		t.Fatal("يسكّر ٤ العصر")
	}
	night := day.Add(20*time.Hour + 50*time.Minute)
	if at, ok := AutoCheckoutAt(night, nil, day.Add(28*time.Hour)); !ok || !at.Equal(day.Add(24*time.Hour)) {
		t.Fatalf("حضور ٨:٥٠ بالليل يسكّر ١٢ مو بنفس الدقيقة: %v", at)
	}
}

func TestShiftWindowTwelveHourTypo(t *testing.T) {
	day := time.Date(2026, 10, 7, 0, 0, 0, 0, debriefLoc)
	f, to := ShiftWindow(nil, sp("04:00"), sp("12:00"), day)
	if !f.Equal(day.Add(16*time.Hour)) || !to.Equal(day.Add(24*time.Hour)) {
		t.Fatalf("04:00–12:00 = 16:00–24:00, got %v–%v", f, to)
	}
}

func TestShiftWindowEveningNoonTypo(t *testing.T) {
	day := time.Date(2026, 10, 7, 0, 0, 0, 0, debriefLoc)
	f, to := ShiftWindow(sp("EVENING"), sp("16:00"), sp("12:00"), day)
	if !to.Equal(day.Add(24*time.Hour)) || !f.Equal(day.Add(16*time.Hour)) {
		t.Fatalf("16:00–12:00 should end at midnight, got %v–%v", f, to)
	}
	if !IsEveningShift(f) {
		t.Fatal("should be evening")
	}
	// غلط إدخال يطلّع شفت طويل يرجع للافتراضي.
	f, to = ShiftWindow(nil, sp("07:00"), sp("23:30"), day)
	if !f.Equal(day.Add(8*time.Hour)) || !to.Equal(day.Add(16*time.Hour)) {
		t.Fatalf("19h shift should fall back to default, got %v–%v", f, to)
	}
}

func TestInShift(t *testing.T) {
	day := time.Date(2026, 10, 7, 0, 0, 0, 0, debriefLoc)
	f, to := ShiftWindow(nil, nil, nil, day)
	if InShift(f, to, day.Add(6*time.Hour)) {
		t.Fatal("6am is outside an 8–16 shift")
	}
	if !InShift(f, to, day.Add(7*time.Hour+30*time.Minute)) || !InShift(f, to, day.Add(8*time.Hour+30*time.Minute)) {
		t.Fatal("7:30 and 8:30 are inside")
	}
	if InShift(f, to, day.Add(16*time.Hour)) {
		t.Fatal("16:00 is after the shift")
	}
}
