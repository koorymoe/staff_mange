package service

import (
	"testing"

	"staffmange-api/internal/model"
)

// هذي الاختبارات تثبّت الشروط الي كانت ناقصة من الاكسل — أي تعديل مستقبلي
// يكسر واحد منها يطلع فشل فوراً بدل ما تنكسر الأسعار بصمت.

func condCatalog() []model.SystemPriceCatalog {
	return []model.SystemPriceCatalog{
		{SystemName: "س", ItemName: "جهاز", Category: "install", Value: 10000},
		{SystemName: "س", ItemName: "كيبل", Category: "wiring", Value: 1.0},
		{SystemName: "س", ItemName: "برمجة", Category: "programming", Value: 2500},
	}
}

// شرط 1: اختيار نوع تسليك بالسطر يصفّر أجور التركيب (G = IF(H="", C*F*E, 0)).
func TestWiringZeroesInstallation(t *testing.T) {
	base := model.ExecutionCostItem{SystemName: "س", ItemName: "جهاز", Count: 20, HeightMeters: 4}
	_, withoutWiring, _, _ := CalculateExecutionCostDetailed(
		[]model.ExecutionCostItem{base}, condCatalog(), 20)
	if withoutWiring[0].InstallTotal != 200000 {
		t.Fatalf("بدون تسليك لازم التركيب = 200000، طلع %v", withoutWiring[0].InstallTotal)
	}

	wired := base
	wired.WiringItemName = "كيبل"
	wired.CableLengthMeters = 10
	_, withWiring, _, _ := CalculateExecutionCostDetailed(
		[]model.ExecutionCostItem{wired}, condCatalog(), 20)
	if withWiring[0].InstallTotal != 0 {
		t.Fatalf("مع التسليك لازم التركيب يتصفّر، طلع %v", withWiring[0].InstallTotal)
	}
}

// شرط 2: وزن ارتفاع التسليك ثنائي — O = IF(الارتفاع >= 5, 2, 1).
func TestWiringHeightWeightIsBinary(t *testing.T) {
	cases := map[int]float64{0: 1, 4: 1, 5: 2, 9: 2}
	for h, want := range cases {
		if got := wiringHeightWeight(h); got != want {
			t.Fatalf("ارتفاع تسليك %d: توقعنا %v وطلع %v", h, want, got)
		}
	}
	item := model.ExecutionCostItem{
		SystemName: "س", ItemName: "جهاز", Count: 1, HeightMeters: 4,
		WiringItemName: "كيبل", CableLengthMeters: 10,
	}
	item.WiringHeightMeters = 4
	_, low, _, _ := CalculateExecutionCostDetailed([]model.ExecutionCostItem{item}, condCatalog(), 1)
	item.WiringHeightMeters = 5
	_, high, _, _ := CalculateExecutionCostDetailed([]model.ExecutionCostItem{item}, condCatalog(), 1)
	if high[0].WiringByCableLength != low[0].WiringByCableLength*2 {
		t.Fatalf("ارتفاع 5 متر لازم يضاعف تسليك الطول: %v مقابل %v",
			high[0].WiringByCableLength, low[0].WiringByCableLength)
	}
}

// شرط 3: الحد الأدنى لأجور التركيب حسب عدد الأجهزة (G59).
func TestInstallMinimumApplies(t *testing.T) {
	// جهاز واحد بسعر تركيب 10000 — أقل من الحد الأدنى 14000، فلازم يترفع.
	items := []model.ExecutionCostItem{{SystemName: "س", ItemName: "جهاز", Count: 1, HeightMeters: 4}}
	total, _, mins, _ := CalculateExecutionCostDetailed(items, condCatalog(), 1)
	if !mins[0].InstallFloorUsed {
		t.Fatal("لازم ينطبق الحد الأدنى للتركيب")
	}
	if mins[0].InstallApplied != 14000 {
		t.Fatalf("الحد الأدنى لجهاز واحد = 14000، طلع %v", mins[0].InstallApplied)
	}
	if total != 14000 {
		t.Fatalf("المجموع لازم 14000، طلع %d", total)
	}
}

// شرط 4: الحد الأدنى لأجور البرمجة (R59) — مبلغ مقطوع حسب عدد الخدمات.
func TestProgrammingMinimumApplies(t *testing.T) {
	for n, want := range map[int]float64{1: 13500, 2: 24500, 3: 32500, 4: 35000, 9: 35000} {
		if got := programmingMinimumTotal(n); got != want {
			t.Fatalf("%d خدمات برمجة: توقعنا %v وطلع %v", n, want, got)
		}
	}
}

// الحدود الدنيا ما تنزل أبداً لما يزيد العدد (غلاف غير متناقص).
func TestInstallMinimumIsMonotonic(t *testing.T) {
	prev := 0.0
	for c := 1; c <= 60; c++ {
		got := installMinimumTotal(c)
		if got < prev {
			t.Fatalf("عدد %d نزّل الحد الأدنى: %v بعد %v", c, got, prev)
		}
		prev = got
	}
}

// ── استمارة كاميرات المراقبة (شيت حساب تكلفة التنفيذ) ──

func TestCameraCostLayers(t *testing.T) {
	req := model.CameraCostRequest{
		PlaceType:  "مدرسة او شركة", // ×1.3
		SystemType: "IP",            // ×1.2
		Rows: []model.CameraCostRow{
			{NormalCableMeters: 15, HeightAbove3m: true}, // شريحة 12000، ارتفاع ×1.1
		},
	}
	res, err := CalculateCameraCost(req)
	if err != nil {
		t.Fatal(err)
	}
	want := 12000 * 1.3 * 1.2 * 1.1 // = 20592
	if res.Rows[0].Total != want {
		t.Fatalf("توقعنا %v وطلع %v", want, res.Rows[0].Total)
	}
	if res.CameraCount != 1 {
		t.Fatalf("عدد الكاميرات لازم 1، طلع %d", res.CameraCount)
	}
}

func TestCameraCostVipCableIs1_2(t *testing.T) {
	normal, _ := CalculateCameraCost(model.CameraCostRequest{
		PlaceType: "منزل سكني", SystemType: "ANLOGE",
		Rows: []model.CameraCostRow{{NormalCableMeters: 15}},
	})
	vip, _ := CalculateCameraCost(model.CameraCostRequest{
		PlaceType: "منزل سكني", SystemType: "ANLOGE",
		Rows: []model.CameraCostRow{{VipCableMeters: 15}},
	})
	if vip.CamerasTotal != normal.CamerasTotal*1.2 {
		t.Fatalf("كيبل VIP لازم ×1.2: %v مقابل %v", vip.CamerasTotal, normal.CamerasTotal)
	}
}

func TestCameraCostExtrasAndDiscount(t *testing.T) {
	res, err := CalculateCameraCost(model.CameraCostRequest{
		PlaceType: "منزل سكني", SystemType: "ANLOGE",
		Rows: []model.CameraCostRow{{NormalCableMeters: 5}}, // شريحة 10000
		Extras: model.CameraCostExtras{
			ScreenLarge43Count: 2,  // 30000
			RackCount:          1,  // 15000
			VipInternetMeters:  10, // 4000
			OtherAmount:        1000,
		},
		Discount: 5000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.ExtrasTotal != 50000 {
		t.Fatalf("الأعمال الإضافية لازم 50000، طلعت %v", res.ExtrasTotal)
	}
	if res.FinalAmount != 10000+50000-5000 {
		t.Fatalf("المبلغ النهائي غلط: %v", res.FinalAmount)
	}
}

func TestCameraCostMorePriceForMoreCable(t *testing.T) {
	prev := 0.0
	for _, m := range []float64{5, 15, 25, 35, 45, 60} {
		res, _ := CalculateCameraCost(model.CameraCostRequest{
			PlaceType: "منزل سكني", SystemType: "ANLOGE",
			Rows: []model.CameraCostRow{{NormalCableMeters: m}},
		})
		if res.CamerasTotal < prev {
			t.Fatalf("%v متر نزّل السعر: %v بعد %v", m, res.CamerasTotal, prev)
		}
		prev = res.CamerasTotal
	}
}
