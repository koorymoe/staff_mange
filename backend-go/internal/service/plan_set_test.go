package service

import (
	"strings"
	"testing"
)

// (ك) 11-10: الأقسام العشرة — كل نوع كوده يبدي برقم قسمه، وماكو كود مكرر.
func TestPlanCatalog(t *testing.T) {
	if len(PlanCatalog) != 10 {
		t.Fatalf("لازم ١٠ أقسام، طلع %d", len(PlanCatalog))
	}
	seen := map[string]bool{}
	for i, g := range PlanCatalog {
		if g.No != i+1 || len(g.Types) == 0 {
			t.Errorf("القسم %d مو بترتيبه أو فارغ", g.No)
		}
		for _, ty := range g.Types {
			if seen[ty.Code] {
				t.Errorf("كود مكرر %s", ty.Code)
			}
			seen[ty.Code] = true
			if !strings.HasPrefix(ty.Code, strings.Split(ty.Code, ".")[0]+".") || strings.Split(ty.Code, ".")[0] != itoa(g.No) {
				t.Errorf("النوع %s مو تبع قسمه %d", ty.Code, g.No)
			}
		}
	}
	if planTypeCount() != len(seen) || !validPlanType("2.2") || validPlanType("11.1") {
		t.Error("عدّ الأنواع أو التحقق منها غلط")
	}
}

func itoa(n int) string {
	if n == 10 {
		return "10"
	}
	return string(rune('0' + n))
}
