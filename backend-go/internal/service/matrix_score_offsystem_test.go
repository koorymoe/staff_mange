package service

import (
	"testing"

	"staffmange-api/internal/repository"
)

// (دازاي 11-13) «اني شغلي ماموجود بالنظام… لازم اتقيم ع تقرير الانجاز وع الحضور والمهام»
func TestOffSystemGroup(t *testing.T) {
	g := ScoreGroup(repository.Scorable{Role: "TECHNICIAN", OffSystem: true})
	if g != GroupOffSystem {
		t.Fatalf("المجموعة = %s", g)
	}
	for _, r := range []string{"DAILY_REPORT", "ATTENDANCE", "TASK_DUE"} {
		if !RuleFitsGroup(r, g) {
			t.Fatalf("%s لازم تنحسب", r)
		}
	}
	for _, r := range []string{"REPORT_MATCH", "WORK", "CREATE", "COMPLETE", "PAPER", "STAGE_LOG"} {
		if RuleFitsGroup(r, g) {
			t.Fatalf("%s ما تنحسب على شغل برّا النظام", r)
		}
	}
	if _, ok := revenueKindOf[g]; ok {
		t.Fatal("ماكو مبالغ مدخولة")
	}
}
