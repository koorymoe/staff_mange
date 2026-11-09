package service

import (
	"testing"

	"staffmange-api/internal/model"
)

// يحرس جداول تسعير فاتورة الليدر: الحدود ما تطيّح (index out of range)، والشرائح صحيحة.
func TestLeaderInvoicePricingTables(t *testing.T) {
	for _, n := range []int{-5, 0, 1, 40, 80, 81, 500} {
		_ = deviceCountWiringPrice(n) // لازم ما يطيح
	}
	if deviceCountWiringPrice(500) != deviceCountWiringPrice(80) {
		t.Error("فوق ٨٠ جهاز لازم ينحسب مثل ٨٠")
	}
	if deviceCountWiringPrice(-1) != deviceCountWiringPrice(0) {
		t.Error("سالب لازم ينحسب صفر")
	}
	for m := -1; m <= 100; m++ {
		_ = cableLengthWiringPricePerMeter(m)
	}
	if cableLengthWiringPricePerMeter(0) != 0 || cableLengthWiringPricePerMeter(31) != cableLengthFlatRateAbove30 || cableLengthWiringPricePerMeter(30) != cableLengthFlatRateAbove30 {
		t.Error("شرائح طول الكيبل")
	}
}

func TestLeaderInvoiceMultipliers(t *testing.T) {
	for h, want := range map[int]float64{0: 1, 4: 1, 5: 1.15, 6: 1.3, 7: 1.5, 8: 1.7, 9: 2, 30: 2} {
		if got := heightWeightMultiplier(h); got != want {
			t.Errorf("ارتفاع %d: %v، المتوقع %v", h, got, want)
		}
	}
	for n, want := range map[int]float64{0: 0, 1: 14000, 4: 14000, 5: 12500, 8: 12500, 9: 11500, 16: 11500, 17: 10000, 100: 10000} {
		if got := installMinimumPerDevice(n); got != want {
			t.Errorf("أجهزة %d: %v، المتوقع %v", n, got, want)
		}
	}
}

func TestLeaderInvoiceCatalogLookup(t *testing.T) {
	cat := []model.SystemPriceCatalog{
		{Category: "wiring", SystemName: "CCTV", ItemName: "مخفي", Value: 1.5},
		{Category: "install", SystemName: "CCTV", ItemName: "كاميرا", Value: 15000},
		{Category: "programming", SystemName: "CCTV", ItemName: "NVR", Value: 25000},
		{Category: "install", SystemName: "ALARM", ItemName: "كاميرا", Value: 99},
	}
	if wiringMultiplierFor(cat, "CCTV", "") != 1 || wiringMultiplierFor(cat, "CCTV", "مخفي") != 1.5 || wiringMultiplierFor(cat, "CCTV", "غير") != 1 {
		t.Error("معامل التمديد")
	}
	if installPriceFor(cat, "CCTV", "كاميرا") != 15000 || installPriceFor(cat, "X", "كاميرا") != 0 {
		t.Error("سعر التنصيب لازم يطابق النظام نفسه")
	}
	if programmingPriceFor(cat, "CCTV", "") != 0 || programmingPriceFor(cat, "CCTV", "NVR") != 25000 {
		t.Error("سعر البرمجة")
	}
}
