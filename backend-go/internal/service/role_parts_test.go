package service

import (
	"testing"
	"time"

	"staffmange-api/internal/repository"
)

// (دازاي 11-21) البيّاع: تسجيل 30 · تحويل 25 · تسجيل الشكاوى 10 · مراقب 8 · حضور 5 · مهام 1 · زملاء 1 + ماتركس 20
func TestSalesScore(t *testing.T) {
	rows := []StaffScore{{ID: "s", Group: GroupSales,
		RelParts:   []ReliabilityPart{{Key: "ATTENDANCE", Pct: 100}, {Key: "PEERS", Pct: 100}},
		ruleTotals: map[string][2]int{"CREATE": {4, 4}}}}
	hum := []repository.HumanRow{{RateeID: "s", Stage: "MONITOR_DAILY", Score: 4}}
	applyRoleScores(rows, hum, nil, map[string]repository.CoordFact{"s": {Completed: 1, Cancelled: 3}})
	// تسجيل 30 + تحويل 25%×25=6.25 + مراقب 80%×8=6.4 + حضور 5 + زملاء 1 = 48.65 من 69 ⇒ ×80/69 = 56.41
	// ماتركس: التحويل تحت ٧٠ (−٣١.٢٥) ⇒ 68.75 ⇒ 13.75 ⇒ 70.16
	if f := rows[0].Final; f == nil || !near(*f, 70.16) {
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
