package service

import (
	"testing"
	"time"

	"staffmange-api/internal/repository"
)

// (دازاي 11-19) البيّاع: تسجيل 30 · تحويل 20 · مراقب 15 · حضور 10 · مهام 3 · زملاء 2 + ماتركس 20
func TestSalesScore(t *testing.T) {
	rows := []StaffScore{{ID: "s", Group: GroupSales,
		RelParts:   []ReliabilityPart{{Key: "ATTENDANCE", Pct: 100}, {Key: "PEERS", Pct: 100}},
		ruleTotals: map[string][2]int{"CREATE": {4, 4}}}}
	hum := []repository.HumanRow{{RateeID: "s", Stage: "MONITOR_DAILY", Score: 4}}
	applyRoleScores(rows, hum, nil, map[string]repository.CoordFact{"s": {Completed: 1, Cancelled: 3}})
	// تسجيل 30 + تحويل 25%×20=5 + مراقب 80%×15=12 + حضور 10 + زملاء 2 = 59 من 77 ⇒ ×80/77 = 61.30
	// ماتركس: التحويل تحت ٧٠ (−٢٥) ⇒ 75 ⇒ 15 ⇒ 76.30
	if f := rows[0].Final; f == nil || !near(*f, 76.30) {
		t.Fatalf("final = %v", f)
	}
}

func TestEveryScoredGroupHasWeights(t *testing.T) {
	for _, g := range ScoreGroupOrder {
		if g == GroupOffSystem || g == GroupDecor {
			continue // برّا النظام ينقيّم على تقريره بس، والديكور بالأخير
		}
		if _, ok := roleWeightsOf[g]; !ok {
			t.Errorf("المجموعة %s بلا أوزان جديدة", g)
		}
	}
	if targetOf(GroupSales) != 300 || targetOf(GroupTechnical) != 130 || targetOf(GroupFinance) != 100 {
		t.Fatal("أهداف الفلوس")
	}
}

func TestSalesRevenue300(t *testing.T) {
	from := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)
	f := 80.0
	rows := []StaffScore{{ID: "s", Group: GroupSales, Final: &f}}
	ApplyRevenue(rows, []repository.RevenueRow{{EmployeeID: "s", Kind: "BOOKED", Amount: 3_000_000}}, map[string]float64{"s": 1_000_000}, from, to, to.Add(-time.Hour))
	if *rows[0].Final < 80 { // ٣ أضعاف = وصل هدفه، ما ينزل
		t.Fatalf("sales dropped at target: %v", *rows[0].Final)
	}
}
