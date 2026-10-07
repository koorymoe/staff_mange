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

func TestAutoCheckoutRequiresEightHours(t *testing.T) {
	day := time.Date(2026, 10, 7, 0, 0, 0, 0, debriefLoc)
	_, end := ShiftWindow(nil, nil, nil, day)
	late := day.Add(10 * time.Hour)
	if at, ok := AutoCheckoutAt(late, end, nil, end.Add(6*time.Hour)); !ok || !at.Equal(late.Add(8*time.Hour)) {
		t.Fatalf("late check-in should end at +8h, got %v", at)
	}
	if _, ok := AutoCheckoutAt(late, end, nil, end.Add(4*time.Hour)); ok {
		t.Fatal("grace must count from the 8h end")
	}
	early := day.Add(7 * time.Hour)
	if at, _ := AutoCheckoutAt(early, end, nil, end.Add(4*time.Hour)); !at.Equal(end) {
		t.Fatalf("early check-in should end at shift end, got %v", at)
	}
	last := day.Add(19*time.Hour + 30*time.Minute)
	if at, _ := AutoCheckoutAt(late, end, &last, end.Add(8*time.Hour)); !at.Equal(last) {
		t.Fatalf("booking finished later wins, got %v", at)
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
	f, to = ShiftWindow(nil, sp("04:00"), sp("23:00"), day)
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
