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

func TestAutoCheckoutAt(t *testing.T) {
	end := time.Date(2026, 10, 6, 16, 0, 0, 0, debriefLoc)
	in := end.Add(-8 * time.Hour)
	if _, ok := AutoCheckoutAt(in, end, nil, end.Add(2*time.Hour)); ok {
		t.Fatal("قبل ٣ ساعات ما يسكّر")
	}
	if at, ok := AutoCheckoutAt(in, end, nil, end.Add(4*time.Hour)); !ok || !at.Equal(end) {
		t.Fatal("يسكّر بنهاية الشفت")
	}
	last := end.Add(90 * time.Minute)
	if at, _ := AutoCheckoutAt(in, end, &last, end.Add(4*time.Hour)); !at.Equal(last) {
		t.Fatal("آخر حجز بعد الشفت = وقت الانصراف")
	}
}
