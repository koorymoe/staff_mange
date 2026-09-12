package service

import (
	"testing"

	"staffmange-api/internal/model"
)

// كيف الاكسل يعد البرمجة: S = IF(R>0,1,0) لكل *سطر* اختار خدمة برمجة —
// مو لكل جهاز. فـ3 كاميرات بخدمة برمجة وحدة = خدمة وحدة (S58=1).
func TestProgrammingCountedPerSelectionNotPerDevice(t *testing.T) {
	cat := []model.SystemPriceCatalog{
		{SystemName: "س", ItemName: "كاميرا", Category: "install", Value: 25000},
		{SystemName: "س", ItemName: "برمجة المنظومة", Category: "programming", Value: 2500},
	}
	// 3 كاميرات + خدمة برمجة وحدة
	one := []model.ExecutionCostItem{{
		SystemName: "س", ItemName: "كاميرا", Count: 3, HeightMeters: 4,
		ProgrammingItem: "برمجة المنظومة",
	}}
	_, _, m1, _ := CalculateExecutionCostDetailed(one, cat, 3)
	if m1[0].ProgrammingCount != 1 || m1[0].ProgrammingApplied != 13500 {
		t.Fatalf("خدمة وحدة: توقعنا عدد=1 ومبلغ=13500، طلع عدد=%d ومبلغ=%v",
			m1[0].ProgrammingCount, m1[0].ProgrammingApplied)
	}
	// نفس الكاميرات بس 3 خدمات برمجة (3 بنود = 3 أسطر بالاكسل)
	three := []model.ExecutionCostItem{
		{SystemName: "س", ItemName: "كاميرا", Count: 1, HeightMeters: 4, ProgrammingItem: "برمجة المنظومة"},
		{SystemName: "س", ItemName: "كاميرا", Count: 1, HeightMeters: 4, ProgrammingItem: "برمجة المنظومة"},
		{SystemName: "س", ItemName: "كاميرا", Count: 1, HeightMeters: 4, ProgrammingItem: "برمجة المنظومة"},
	}
	_, _, m3, _ := CalculateExecutionCostDetailed(three, cat, 3)
	if m3[0].ProgrammingCount != 3 || m3[0].ProgrammingApplied != 32500 {
		t.Fatalf("3 خدمات: توقعنا عدد=3 ومبلغ=32500، طلع عدد=%d ومبلغ=%v",
			m3[0].ProgrammingCount, m3[0].ProgrammingApplied)
	}
	t.Log("خدمة وحدة -> 13,500 | 3 خدمات -> 32,500 — نفس منطق الاكسل ✔")
}
