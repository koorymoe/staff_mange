package model

import "testing"

// TestCalculateLeaderCommission يطابق المثال المحسوب يدوياً من المالك:
// executionCost=20000, بند مادة واحد qty=2 sellPrice=11000 wholesalePrice=5000
// => profitPerUnit=6000 => totalProfit=12000 => salesCommission=6000،
// executionCommission=8000، leaderTotal=14000.
func TestCalculateLeaderCommission(t *testing.T) {
	executionCost := 20000.0
	sellPrice := 11000.0
	wholesalePrice := 5000.0
	profitPerUnit := sellPrice - wholesalePrice // 6000
	materials := []LeaderInvoiceMaterialItem{
		{Quantity: 2, UnitPrice: sellPrice, ProfitPerUnit: profitPerUnit},
	}

	executionCommission, salesCommission, totalProfit := CalculateLeaderCommission(executionCost, materials)

	if profitPerUnit != 6000 {
		t.Fatalf("profitPerUnit = %.2f, want 6000", profitPerUnit)
	}
	if totalProfit != 12000 {
		t.Fatalf("totalProfit = %.2f, want 12000", totalProfit)
	}
	if executionCommission != 8000 {
		t.Fatalf("executionCommission = %.2f, want 8000", executionCommission)
	}
	if salesCommission != 6000 {
		t.Fatalf("salesCommission = %.2f, want 6000", salesCommission)
	}
	leaderTotal := executionCommission + salesCommission
	if leaderTotal != 14000 {
		t.Fatalf("leaderTotal = %.2f, want 14000", leaderTotal)
	}
}

// TestCalculateTechnicianCommission يطابق مثال المالك: فني على نفس الحجز يأخذ
// executionCost * 0.3 كاملة (بدون تقسيم بين الفنيين المتعددين).
func TestCalculateTechnicianCommission(t *testing.T) {
	executionCost := 20000.0
	got := CalculateTechnicianCommission(executionCost)
	if got != 6000 {
		t.Fatalf("technicianCommission = %.2f, want 6000", got)
	}
}

// TestCalculateTechnicianCommission_MultipleTechniciansEachGetFull يتأكد إن
// كل فني ياخذ نفس المبلغ الكامل ولا ينقسم بينهم — مؤكد صراحة من المالك.
func TestCalculateTechnicianCommission_MultipleTechniciansEachGetFull(t *testing.T) {
	executionCost := 20000.0
	tech1 := CalculateTechnicianCommission(executionCost)
	tech2 := CalculateTechnicianCommission(executionCost)
	if tech1 != 6000 || tech2 != 6000 {
		t.Fatalf("each technician commission = %.2f / %.2f, want 6000 / 6000 (not divided)", tech1, tech2)
	}
}
