package handler

import (
	"strings"
	"testing"
)

func TestSafeErrorMessage(t *testing.T) {
	leaks := []string{
		`تعذر: pq: duplicate key value violates unique constraint "Booking_code_key"`,
		`sql: no rows in result set`,
		`pq: column "foo" does not exist`,
		`dial tcp 127.0.0.1:5432: connect: connection refused`,
	}
	for _, m := range leaks {
		got := SafeErrorMessage(500, m)
		if strings.Contains(got, "pq") || strings.Contains(got, "sql") || strings.Contains(got, "5432") || !strings.Contains(got, "الرمز") {
			t.Errorf("تسرّب: %q → %q", m, got)
		}
	}
	for _, m := range []string{"اكتب مبلغ الدفعة", "هذا الحجز عليه فاتورة اليوم أصلاً (LI-12)", "مو موجود"} {
		if got := SafeErrorMessage(400, m); got != m {
			t.Errorf("غيّر رسالة مقصودة: %q → %q", m, got)
		}
	}
}
