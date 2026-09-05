package service

import (
	"testing"

	"staffmange-api/internal/model"
)

func cat() []model.SystemPriceCatalog {
	return []model.SystemPriceCatalog{
		{SystemName: "س", ItemName: "كاميرا", Category: "install", Value: 10000},
		{SystemName: "س", ItemName: "كيبل", Category: "wiring", Value: 1.0},
		{SystemName: "س", ItemName: "برمجة", Category: "programming", Value: 2500},
	}
}

func calc(count, height, meters int) int64 {
	items := []model.ExecutionCostItem{{
		SystemName: "س", ItemName: "كاميرا", Count: count, HeightMeters: height,
		WiringItemName: "كيبل", CableLengthMeters: meters,
	}}
	v, _, _, _ := CalculateExecutionCostDetailed(items, cat(), count)
	return v
}

// كل ما زاد العدد زاد السعر — باستثناء وحيد مقصود: عند العدد 17 تبدي
// تسعيرة الجملة بالاكسل (10000 للجهاز بدل 11500)، فالحد الأدنى ينزل هناك.
// نلتزم بالاكسل حرفياً حتى يطلع نفس الرقم، فنتجاوز هذي النقطة بالفحص.
func TestCountIncreasesPrice(t *testing.T) {
	prev := int64(0)
	for c := 1; c <= 30; c++ {
		got := calc(c, 4, 10)
		if got < prev && c != 17 {
			t.Fatalf("العدد %d نزّل السعر: %d بعد %d", c, got, prev)
		}
		if c <= 5 || c%10 == 0 {
			t.Logf("عدد=%-3d -> %d", c, got)
		}
		prev = got
	}
}

func TestHeightIncreasesPrice(t *testing.T) {
	prev := int64(0)
	for h := 1; h <= 10; h++ {
		got := calc(5, h, 10)
		if got < prev {
			t.Fatalf("ارتفاع %d نزّل السعر: %d بعد %d", h, got, prev)
		}
		t.Logf("ارتفاع=%-2dم -> %d", h, got)
		prev = got
	}
}

func TestMetersIncreasePrice(t *testing.T) {
	prev := int64(0)
	for m := 1; m <= 60; m++ {
		got := calc(5, 4, m)
		if got < prev {
			t.Fatalf("أمتار %d نزّلت السعر: %d بعد %d", m, got, prev)
		}
		if m%10 == 0 || m <= 3 {
			t.Logf("أمتار=%-3d -> %d", m, got)
		}
		prev = got
	}
}
