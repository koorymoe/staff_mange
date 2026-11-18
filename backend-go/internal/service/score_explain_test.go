package service

import (
	"strings"
	"testing"

	"staffmange-api/internal/repository"
)

// (دازاي 11-18) «على شنو ينقيّم»: من نفس جداول الحساب.
func TestExplainOne(t *testing.T) {
	tech := ExplainOne(repository.Scorable{ID: "t", Name: "فني", Role: "TECHNICIAN"})
	if tech.Group != GroupTechs || tech.Revenue == "" || !strings.Contains(strings.Join(tech.Raters, "،"), "ليدره") {
		t.Fatalf("الفني: %+v", tech)
	}
	coord := ExplainOne(repository.Scorable{ID: "c", Role: "HR_COORDINATOR"})
	if coord.Revenue != "" {
		t.Fatalf("الإداري ما تنحسب عليه فلوس: %+v", coord)
	}
	ai := ExplainOne(repository.Scorable{ID: "a", Role: "IT_SUPPORT", AIOnly: true})
	if len(ai.Raters) != 1 || !strings.Contains(ai.Raters[0], "ماتركس بس") {
		t.Fatalf("ماتركس بس: %+v", ai.Raters)
	}
	off := ExplainOne(repository.Scorable{ID: "o", Role: "IT_SUPPORT", OffSystem: true})
	for _, rl := range off.Rules {
		if strings.Contains(rl, "المراحل") {
			t.Fatalf("شغل برّا النظام ما ينقيّم على المراحل: %v", off.Rules)
		}
	}
	if len(off.Rules) == 0 || off.Why == "" {
		t.Fatalf("لازم قواعد وسبب: %+v", off)
	}
}
