package middleware

import (
	"testing"
	"time"
)

// TestAuthzViolationThreshold يثبّت السلوك الي انكسر بيه موظف حقيقي: محاولة
// وصول مرفوضة وحدة ما تكفي للحظر، والعدّ يعتمد نافذة زمنية مو عدّاد أبدي.
func TestAuthzViolationThreshold(t *testing.T) {
	const emp = "emp-threshold"
	clearAuthzViolations(emp)
	t.Cleanup(func() { clearAuthzViolations(emp) })

	now := time.Now()
	for i := 1; i <= authzLockThreshold; i++ {
		got := registerAuthzViolation(emp, now)
		if got != i {
			t.Fatalf("المحاولة %d: توقعنا العدّاد %d وطلع %d", i, i, got)
		}
		if i < authzLockThreshold && got >= authzLockThreshold {
			t.Fatalf("المحاولة %d وصلت العتبة قبل وقتها — الموظف راح ينحظر بالغلط", i)
		}
	}
}

// TestAuthzViolationWindowExpires محاولات قديمة برّا النافذة ما تنحسب — موظف
// غلط مرة كل شهر ما يجوز يتراكم عليه لحد ما ينحظر.
func TestAuthzViolationWindowExpires(t *testing.T) {
	const emp = "emp-window"
	clearAuthzViolations(emp)
	t.Cleanup(func() { clearAuthzViolations(emp) })

	base := time.Now()
	for i := 0; i < authzLockThreshold-1; i++ {
		registerAuthzViolation(emp, base)
	}
	// محاولة جديدة بعد ما تنتهي النافذة: لازم يرجع العدّ لواحد
	if got := registerAuthzViolation(emp, base.Add(authzLockWindow+time.Minute)); got != 1 {
		t.Fatalf("توقعنا العدّاد يرجع 1 بعد انتهاء النافذة، وطلع %d", got)
	}
}
