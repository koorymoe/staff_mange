package service

import (
	"testing"

	"staffmange-api/internal/model"
)

// حالة المستخدم الحقيقية: كاميرات انلوك / تنصيب كاميرا على الستاند
// عدد 3، ارتفاع تركيب 4م، تسليك "كيبل كاميرا انلوك في اي بي" (مضاعف 1.8)،
// ارتفاع تسليك 4م، طول الكيبل 40م.  الاكسل يعطي 154,000.
func TestUserCase154(t *testing.T) {
	catalog := []model.SystemPriceCatalog{
		{SystemName: "كاميرات انلوك", ItemName: "تنصيب كاميرا على الستاند", Category: "install", Value: 5000},
		{SystemName: "كاميرات انلوك", ItemName: "تسليك كيبل كاميرا انلوك في اي بي", Category: "wiring", Value: 1.8},
	}
	items := []model.ExecutionCostItem{{
		SystemName: "كاميرات انلوك", ItemName: "تنصيب كاميرا على الستاند",
		Count: 3, HeightMeters: 4,
		WiringItemName:     "تسليك كيبل كاميرا انلوك في اي بي",
		WiringHeightMeters: 4, CableLengthMeters: 40,
	}}
	total, lines, mins, err := CalculateExecutionCostDetailed(items, catalog, 3)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("K (حسب العدد الكلي) = %v   [الاكسل: 10,600 × 3 = 31,800]", lines[0].WiringByDeviceCount)
	t.Logf("M (حسب الطول)       = %v   [الاكسل: 710×40×1.8 = 51,120 × 3 = 153,360]", lines[0].WiringByCableLength)
	t.Logf("P (المعتمد)         = %v   [الاكسل: 153,360]", lines[0].WiringTotal)
	t.Logf("الحد الأدنى         = %v   [الاكسل: 3 × 14,000 = 42,000]", mins[0].InstallMinimumTotal)
	t.Logf("المجموع النهائي     = %d   [الاكسل: 154,000]", total)
	if total != 154000 {
		t.Fatalf("المتوقع 154000، طلع %d", total)
	}
}
