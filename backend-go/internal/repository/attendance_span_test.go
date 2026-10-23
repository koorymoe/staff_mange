package repository

import (
	"testing"
	"time"
)

// فحص السيرفر 10-22: جلسات بملايين الساعات من تصحيح بسنة غلط.
func TestValidAttendanceSpan(t *testing.T) {
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	in := time.Date(2026, 10, 22, 5, 0, 0, 0, time.UTC)
	at := func(y int, mo time.Month, d, h int) *time.Time {
		x := time.Date(y, mo, d, h, 0, 0, 0, time.UTC)
		return &x
	}
	cases := []struct {
		name string
		in   time.Time
		out  *time.Time
		ok   bool
	}{
		{"عادي", in, at(2026, 10, 22, 13), true},
		{"بلا انصراف", in, nil, true},
		{"خروج قبل الدخول", in, at(2026, 10, 22, 4), false},
		{"سنة غلط بالخروج", in, at(2626, 10, 22, 13), false},
		{"أطول من ٢٤ ساعة", in, at(2026, 10, 23, 6), false},
		{"دخول بالمستقبل", time.Date(2027, 1, 1, 5, 0, 0, 0, time.UTC), nil, false},
	}
	for _, c := range cases {
		if err := ValidAttendanceSpan(c.in, c.out, now); (err == nil) != c.ok {
			t.Errorf("%s: err=%v", c.name, err)
		}
	}
}
