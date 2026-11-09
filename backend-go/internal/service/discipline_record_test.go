package service

import (
	"testing"

	"staffmange-api/internal/repository"
)

// الاسترجاع (نقاط موجبة) مو خصم، والخصم المالي والتسوية خصم دايماً.
func TestDisciplineIsDeduction(t *testing.T) {
	cases := []struct {
		e    repository.DisciplineEntry
		want bool
	}{
		{repository.DisciplineEntry{Source: "POINTS", Points: -5}, true},
		{repository.DisciplineEntry{Source: "POINTS", Points: 10}, false},
		{repository.DisciplineEntry{Source: "KPI", Amount: 25000}, true},
		{repository.DisciplineEntry{Source: "LEDGER", Amount: 5000}, true},
	}
	for _, c := range cases {
		if got := isDeduction(c.e); got != c.want {
			t.Errorf("%+v: got %v want %v", c.e, got, c.want)
		}
	}
	if shortReason("") != "بلا سبب مكتوب" || shortReason("تأخير بالورق — حجز B12") != "تأخير بالورق" {
		t.Errorf("shortReason: %q / %q", shortReason(""), shortReason("تأخير بالورق — حجز B12"))
	}
}
