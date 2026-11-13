package service

import (
	"testing"
	"time"
)

// (دازاي 11-13) موظف بدون جدول: من يسجّل حضور تنحسبله ٨ ساعات وينسكر انصرافه
// تلقائياً؛ وإذا سجّل انصراف بنفسه ينحسب مثل ما هو، بلا تأخير ولا قص.
func TestFlexibleHours(t *testing.T) {
	in := time.Date(2026, 11, 13, 11, 37, 0, 0, debriefLoc) // حضر ١١:٣٧ — ماكو «تأخير»
	if _, ok := FlexibleAutoCheckoutAt(in, nil, in.Add(7*time.Hour+59*time.Minute)); ok {
		t.Fatal("قبل ٨ ساعات ما ينسكر")
	}
	at, ok := FlexibleAutoCheckoutAt(in, nil, in.Add(8*time.Hour+5*time.Minute))
	if !ok || !at.Equal(in.Add(8*time.Hour)) {
		t.Fatalf("بعد ٨ ساعات ينسكر بوقت الثمانية: %v %v", at, ok)
	}
	job := in.Add(9 * time.Hour)
	if _, ok := FlexibleAutoCheckoutAt(in, &job, in.Add(8*time.Hour+30*time.Minute)); ok {
		t.Fatal("خلّص حجز بعد الثمانية = يتمدد لحد الحجز")
	}
	out := in.Add(10 * time.Hour)
	if end, assumed, _ := FlexibleCountedEnd(in, &out, out); !end.Equal(out) || assumed {
		t.Fatal("انصرافه بنفسه بعد ١٠ ساعات ينحسب كامل")
	}
	if end, assumed, _ := FlexibleCountedEnd(in, nil, in.Add(20*time.Hour)); !assumed || !end.Equal(in.Add(8*time.Hour)) {
		t.Fatal("جلسة مفتوحة ٢٠ ساعة = ٨ بس")
	}
}
