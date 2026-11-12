package service

import (
	"testing"

	"staffmange-api/internal/repository"
)

// (دازاي 11-11) «ما يتحاسب ع الانصراف، يتحاسب بس ع الحضور»: الانصراف التلقائي
// (نسى يسجّل انصراف) ما ينزّل الاعتمادية — نفس الرقم بانصراف يدوي أو تلقائي.
func TestReliabilityIgnoresCheckout(t *testing.T) {
	rules := map[string][2]int{"ATTENDANCE": {8, 10}}
	_, allSelf := BuildReliability(rules, repository.ReliabilityFact{Sessions: 10, Auto: 0})
	parts, allAuto := BuildReliability(rules, repository.ReliabilityFact{Sessions: 10, Auto: 10})
	if allSelf == nil || allAuto == nil || *allSelf != *allAuto {
		t.Fatalf("الانصراف التلقائي غيّر الاعتمادية: %v ≠ %v", allSelf, allAuto)
	}
	for _, p := range parts {
		if p.Key == "CHECKOUT" {
			t.Fatal("جزء الانصراف لازم ينشال")
		}
	}
}
