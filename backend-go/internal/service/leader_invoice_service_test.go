package service

import (
	"testing"
	"time"

	"staffmange-api/internal/model"
)

func testCatalog() []model.SystemPriceCatalog {
	return []model.SystemPriceCatalog{
		{SystemName: "كاميرات انلوك", ItemName: "ربط متحكم كي في ام", Category: "install", Value: 20000},
		{SystemName: "كاميرات انلوك", ItemName: "تسليك كيبل كاميرا انلوك", Category: "wiring", Value: 1},
		{SystemName: "كاميرات انلوك", ItemName: "كيبل كهرباء", Category: "wiring", Value: 0.9},
		{SystemName: "كاميرات انلوك", ItemName: "برمجة المنظومة مع ربطها بالتطبيق اضافة هاتف", Category: "programming", Value: 2500},
	}
}

// TestCalculateExecutionCost_SingleItemNoWiring يطابق مثال الشيت الحقيقي المُتحقق
// منه يدوياً: تكاليف المشروع!A5 = CEILING(SUM(B59:BB59),1000) حسبت 20000 لبند
// وحيد "ربط متحكم كي في ام" بسعر تركيب 20000 (ارتفاع <=4م، بدون تسليك ولا برمجة).
func TestCalculateExecutionCost_SingleItemNoWiring(t *testing.T) {
	items := []model.ExecutionCostItem{
		{SystemName: "كاميرات انلوك", ItemName: "ربط متحكم كي في ام", Count: 1, HeightMeters: 4},
	}
	got, err := CalculateExecutionCost(items, testCatalog(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 20000 {
		t.Fatalf("expected 20000, got %d", got)
	}
}

// TestCalculateExecutionCost_ZeroCountItemIgnored يتحقق من أن IF(count>0,1,0) —
// أي بند بعدد 0 لا يُحتسب إطلاقاً حتى لو له سعر تركيب.
func TestCalculateExecutionCost_ZeroCountItemIgnored(t *testing.T) {
	items := []model.ExecutionCostItem{
		{SystemName: "كاميرات انلوك", ItemName: "ربط متحكم كي في ام", Count: 0, HeightMeters: 4},
	}
	got, err := CalculateExecutionCost(items, testCatalog(), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

// TestCalculateExecutionCost_HeightWeightMultiplier يتحقق من مضاعف الارتفاع:
// نفس البند بارتفاع 8 أمتار (×1.7) = 34000 مقرَّبة لأعلى لـ1000.
func TestCalculateExecutionCost_HeightWeightMultiplier(t *testing.T) {
	items := []model.ExecutionCostItem{
		{SystemName: "كاميرات انلوك", ItemName: "ربط متحكم كي في ام", Count: 1, HeightMeters: 8},
	}
	got, err := CalculateExecutionCost(items, testCatalog(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 20000 * 1.7 = 34000 -> already multiple of 1000
	if got != 34000 {
		t.Fatalf("expected 34000, got %d", got)
	}
}

// TestCalculateExecutionCost_WiringUsesMaxOfTwoTables يتحقق من أن تكلفة التسليك
// تاخذ أكبر قيمة بين جدول العدد الكلي وجدول طول الكيبل، ناقص installation_total
// لتفادي الاحتساب المزدوج.
func TestCalculateExecutionCost_WiringUsesMaxOfTwoTables(t *testing.T) {
	items := []model.ExecutionCostItem{
		{
			SystemName:        "كاميرات انلوك",
			ItemName:          "ربط متحكم كي في ام",
			Count:             1,
			HeightMeters:      4,
			WiringItemName:    "تسليك كيبل كاميرا انلوك", // multiplier = 1
			CableLengthMeters: 40,                        // >=30 -> 710/meter => 28400
		},
	}
	// totalDeviceCount = 1 -> device table price = 10200, * mult(1) - installTotal(20000) = negative -> clamp 0
	// length-based: 710 * 40 = 28400
	// total = installTotal(20000) + max(0, 28400) = 48400 -> ceil to 1000 = 49000
	got, err := CalculateExecutionCost(items, testCatalog(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 49000 {
		t.Fatalf("expected 49000, got %d", got)
	}
}

// TestCalculateExecutionCost_ProgrammingAdded يتحقق من إضافة سعر البرمجة الثابت.
func TestCalculateExecutionCost_ProgrammingAdded(t *testing.T) {
	items := []model.ExecutionCostItem{
		{
			SystemName:      "كاميرات انلوك",
			ItemName:        "ربط متحكم كي في ام",
			Count:           1,
			HeightMeters:    4,
			ProgrammingItem: "برمجة المنظومة مع ربطها بالتطبيق اضافة هاتف",
		},
	}
	// 20000 + 2500 = 22500 -> ceil 1000 = 23000
	got, err := CalculateExecutionCost(items, testCatalog(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 23000 {
		t.Fatalf("expected 23000, got %d", got)
	}
}

func TestCalculateExecutionCost_EmptyItemsError(t *testing.T) {
	_, err := CalculateExecutionCost(nil, testCatalog(), 0)
	if err == nil {
		t.Fatal("expected error for empty items")
	}
}

func TestGenerateAccountingCode(t *testing.T) {
	ts, _ := time.Parse("2006-01-02", "2026-07-26")
	code := GenerateAccountingCode("abcdef1234567890", ts)
	if code != "LDR-20260726-567890" {
		t.Fatalf("unexpected code: %s", code)
	}
}
