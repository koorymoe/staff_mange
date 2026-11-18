package service

import (
	"math"
	"testing"
	"time"

	"staffmange-api/internal/repository"
)

func near(a, b float64) bool { return math.Abs(a-b) < 0.01 }

// (دازاي 11-18) الفني: ليدر 30 · حضور 20 · عدّة 15 · مهام 10 · زملاء 5 = 80، وماتركس 20.
func TestTechScoreWeights(t *testing.T) {
	rows := []StaffScore{{ID: "t", Group: GroupTechs, RelParts: []ReliabilityPart{
		{Key: "ATTENDANCE", Pct: 100}, {Key: "INVENTORY", Pct: 100}, {Key: "TASK_DUE", Pct: 50}, {Key: "PEERS", Pct: 75},
		{Key: "COMPLAINTS", Pct: 0}, // ما يدخل برقم الفني
	}}}
	hum := []repository.HumanRow{{RateeID: "t", Stage: "LEADER_CREW", Score: 4}, {RateeID: "t", Stage: "LEADER_CREW", Score: 5},
		{RateeID: "t", Stage: "OPS_DAILY", Score: 1}} // إدارة العمليات ما تدخل
	applyTechScores(rows, hum)
	s := rows[0]
	// ليدر 90%×30=27 · حضور 20 · عدّة 15 · مهام 50%×10=5 · زملاء 75%×5=3.75 = 70.75
	// ماتركس: المهام 50 < 80 ← 80 ⇒ 80×20% = 16 ⇒ 86.75
	if s.Final == nil || !near(*s.Final, 86.75) {
		t.Fatalf("final = %v", s.Final)
	}
	if len(s.MatrixNotes) != 1 || !s.TechScore {
		t.Fatalf("notes = %v", s.MatrixNotes)
	}
}

func TestTechScoreRedistributesMissing(t *testing.T) {
	// ماكو ليدر ولا مهام: الحضور 20 والعدّة 15 والزملاء 5 تتوزع على ٨٠
	rows := []StaffScore{{ID: "t", Group: GroupTechs, RelParts: []ReliabilityPart{
		{Key: "ATTENDANCE", Pct: 50}, {Key: "INVENTORY", Pct: 100}, {Key: "PEERS", Pct: 100}}}}
	applyTechScores(rows, nil)
	// حضور 50%×40 = 20 · عدّة 30 · زملاء 10 ⇒ 60 · ماتركس 100−30 (حضور) = 70 ⇒ 14 ⇒ 74
	if f := rows[0].Final; f == nil || !near(*f, 74) {
		t.Fatalf("final = %v", f)
	}
	// غير الفنيين ما يتأثرون
	other := []StaffScore{{ID: "x", Group: GroupLeaders}}
	applyTechScores(other, nil)
	if other[0].TechScore {
		t.Fatal("leader got tech score")
	}
}

func TestRevenueTargetTech130(t *testing.T) {
	from := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)
	now := to.Add(-time.Hour)
	f1, f2 := 80.0, 80.0
	rows := []StaffScore{{ID: "t", Group: GroupTechs, Final: &f1}, {ID: "l", Group: GroupLeaders, Final: &f2}}
	rev := []repository.RevenueRow{{EmployeeID: "t", Kind: "WORK", Amount: 1_000_000}, {EmployeeID: "l", Kind: "WORK", Amount: 1_000_000}}
	sal := map[string]float64{"t": 1_000_000, "l": 1_000_000}
	ApplyRevenue(rows, rev, sal, from, to, now)
	// الفني غطّى ١٠٠٪ من راتبه = ٧٧٪ من هدفه (١٣٠) فينزل؛ الليدر هدفه ١٠٠ فما ينزل
	if *rows[0].Final >= 80 || rows[0].RevenueTarget != 130 {
		t.Fatalf("tech final = %v target %v", *rows[0].Final, rows[0].RevenueTarget)
	}
	if *rows[1].Final < 80 {
		t.Fatalf("leader dropped: %v", *rows[1].Final)
	}
}
