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

// الحد الأدنى محسوب حرفياً مثل الاكسل: العدد × سعر الشريحة. نثبّت القيم
// عند حدود الشرائح تحديداً — ومنها النزول المقصود عند 17 (تسعيرة جملة).
func TestInstallMinimumMatchesExcel(t *testing.T) {
	cases := map[int]float64{
		1: 14000, 4: 56000, // 1-4 × 14000
		5: 62500, 8: 100000, // 5-8 × 12500
		9: 103500, 16: 184000, // 9-16 × 11500
		17: 170000, 40: 400000, // 17+ × 10000 (ينزل عند 17 — نفس الاكسل)
	}
	for count, want := range cases {
		if got := installMinimumTotal(count); got != want {
			t.Fatalf("عدد %d: الاكسل يعطي %v والنظام أعطى %v", count, want, got)
		}
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

// ── فحص مطابقة كاملة مع الاكسل ──
// سيناريو محسوب باليد خطوة بخطوة من صيغ الشيت الأصلية (بما بيها صيغتَي
// المصفوفة K و L الي انكشفن متأخر)، حتى نتأكد إن النظام يعطي نفس الرقم.
//
// الكتالوج: تركيب=10000، مضاعف نوع التسليك N=1، برمجة=2500
//
// سطر أ: 4 أجهزة، ارتفاع تركيب 6م (F=1.3)، بدون تسليك
//
//	G = 10000 × 1.3 × 4 = 52,000
//
// سطر ب: 2 جهاز، تسليك، طول 10م، ارتفاع تسليك 5م (O=2)
//
//	G = 0                                   (يتصفّر لأنه اكو تسليك)
//	E58 = 4 + 2 = 6  ->  جدول العدد الكلي عند 6 = 11,200
//	K = 11,200 × E   (بدون مضاعف النوع)     -> للوحدة 11,200
//	L = 800 × 10 = 8,000                    (سعر المتر × الأمتار)
//	M = O × N × L = 2 × 1 × 8,000 = 16,000  -> للوحدة 16,000
//	P = MAX(16,000, 11,200) = 16,000  × 2 جهاز = 32,000
//
// G58 = 52,000 + 32,000 = 84,000
// الحد الأدنى: 6 أجهزة -> شريحة 5-8 -> 6 × 12,500 = 75,000
// G59 = MAX(84,000، 75,000) = 84,000      (ما ينطبق الحد الأدنى)
// R59 = MAX(2,500، 13,500) = 13,500
// المجموع = 97,500  ->  CEILING لأقرب 1000 = 98,000
func TestExcelParityFullScenario(t *testing.T) {
	items := []model.ExecutionCostItem{
		{SystemName: "س", ItemName: "جهاز", Count: 4, HeightMeters: 6,
			ProgrammingItem: "برمجة"},
		{SystemName: "س", ItemName: "جهاز", Count: 2, HeightMeters: 4,
			WiringItemName: "كيبل", CableLengthMeters: 10, WiringHeightMeters: 5},
	}
	total, lines, mins, err := CalculateExecutionCostDetailed(items, condCatalog(), 6)
	if err != nil {
		t.Fatal(err)
	}
	check := func(label string, want, got float64) {
		t.Helper()
		if want != got {
			t.Fatalf("%s: الاكسل %v، النظام %v", label, want, got)
		}
	}
	check("G سطر أ", 52000, lines[0].InstallTotal)
	check("G سطر ب (لازم يتصفّر)", 0, lines[1].InstallTotal)
	check("K", 22400, lines[1].WiringByDeviceCount)
	check("M", 32000, lines[1].WiringByCableLength)
	check("P", 32000, lines[1].WiringTotal)

	m := mins[0]
	check("G58", 84000, m.InstallWiringCalculated)
	if m.DeviceCount != 6 {
		t.Fatalf("E58: الاكسل 6، النظام %d", m.DeviceCount)
	}
	check("الحد الأدنى للتركيب", 75000, m.InstallMinimumTotal)
	check("G59", 84000, m.InstallApplied)
	check("R59", 13500, m.ProgrammingApplied)
	if total != 98000 {
		t.Fatalf("المجموع النهائي: الاكسل 98000، النظام %d", total)
	}
}

// K ما ينضرب بمضاعف نوع التسليك — المضاعف يدخل بطرف الطول (M) بس.
// هذا كان غلط بالنظام: كنا نضرب K بالمضاعف فيطلع رقم أعلى من الاكسل.
func TestDeviceBasedWiringIgnoresMultiplier(t *testing.T) {
	cat := []model.SystemPriceCatalog{
		{SystemName: "س", ItemName: "جهاز", Category: "install", Value: 10000},
		{SystemName: "س", ItemName: "كيبل غالي", Category: "wiring", Value: 3.0},
	}
	items := []model.ExecutionCostItem{{
		SystemName: "س", ItemName: "جهاز", Count: 1, HeightMeters: 4,
		WiringItemName: "كيبل غالي", CableLengthMeters: 1,
	}}
	_, lines, _, _ := CalculateExecutionCostDetailed(items, cat, 6)
	// جدول العدد الكلي عند 6 = 11,200 — بدون أي ضرب بـ3
	if lines[0].WiringByDeviceCount != 11200 {
		t.Fatalf("K: الاكسل 11200 (بلا مضاعف)، النظام %v", lines[0].WiringByDeviceCount)
	}
	// أما الطول: 1000 × 1 متر × مضاعف 3 × وزن ارتفاع 1 = 3,000
	if lines[0].WiringByCableLength != 3000 {
		t.Fatalf("M: المتوقع 3000، النظام %v", lines[0].WiringByCableLength)
	}
}
