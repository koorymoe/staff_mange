package model

import "testing"

// (دازاي 10-29) «50 بدل 50000» ينرفض، والصفر والمبالغ الطبيعية تمشي.
func TestCheckIQD(t *testing.T) {
	for v, bad := range map[float64]bool{0: false, 50: true, 999: true, 1000: false, 50000: false, 250000: false} {
		if err := CheckIQD("المبلغ", v); (err != nil) != bad {
			t.Errorf("%v: err=%v", v, err)
		}
	}
	if err := CheckIQD("المبلغ", 50); err == nil || !contains(err.Error(), "50000") {
		t.Errorf("الرسالة لازم تقترح 50000: %v", err)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
