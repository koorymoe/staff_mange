package handler

import (
	"testing"
	"time"
)

// (11-17) «ذكّر المدير»: بس المعلّق، وبعد يومين من الطلب، ومرة باليوم بالأكثر.
func TestRemindCheck(t *testing.T) {
	now := time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)
	old := now.Add(-72 * time.Hour)
	recent := now.Add(-3 * time.Hour)
	cases := []struct {
		status   string
		created  time.Time
		reminded *time.Time
		want     error
	}{
		{"PENDING", old, nil, nil},
		{"IN_PROGRESS", old, nil, nil},
		{"FULFILLED", old, nil, ErrRemindNotOpen},
		{"PENDING", now.Add(-24 * time.Hour), nil, ErrRemindTooNew},
		{"PENDING", old, &recent, ErrRemindTooSoon},
	}
	for _, c := range cases {
		if got := RemindCheck(c.status, c.created, c.reminded, now); got != c.want {
			t.Fatalf("%s: %v، المفروض %v", c.status, got, c.want)
		}
	}
}
