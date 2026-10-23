package service

import "testing"

// بوابة الدقة: ما ينفّذ إلا بعد ٢٠ اقتراح محسوم وقبول ٨٠٪ فأكثر.
func TestAutonomyGate(t *testing.T) {
	cases := []struct {
		decided, accepted int
		want              bool
	}{
		{19, 19, false}, // عيّنة صغيرة حتى لو كلها مقبولة
		{20, 15, false}, // ٧٥٪
		{20, 16, true},  // ٨٠٪ بالضبط
		{20, 17, true},  // ٨٥٪
		{0, 0, false},
	}
	for _, c := range cases {
		if got, pct := AutonomyGate(c.decided, c.accepted); got != c.want {
			t.Errorf("AutonomyGate(%d,%d)=%v (%d٪) want %v", c.decided, c.accepted, got, pct, c.want)
		}
	}
}
