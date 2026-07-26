package service

import (
	"fmt"
	"math"
	"time"

	"staffmange-api/internal/model"
)

// heightWeightMultiplier مضاعف الارتفاع — منسوخ حرفياً من شيت "تكاليف المشروع":
// ارتفاع <=4 متر بدون زيادة، وكل درجة أعلى لها مضاعف ثابت حتى >8 متر = ضعف السعر.
func heightWeightMultiplier(heightMeters int) float64 {
	switch {
	case heightMeters <= 4:
		return 1
	case heightMeters == 5:
		return 1.15
	case heightMeters == 6:
		return 1.3
	case heightMeters == 7:
		return 1.5
	case heightMeters == 8:
		return 1.7
	default: // > 8
		return 2
	}
}

// deviceCountWiringTable جدول التسليك حسب "العدد الكلي لأجهزة المشروع" (0 إلى 80) —
// منسوخ حرفياً من شيت "تكاليف المشروع" (نفس القيم بالضبط).
var deviceCountWiringTable = map[int]float64{
	0: 10000, 1: 10200, 2: 10400, 3: 10600, 4: 10800, 5: 11000, 6: 11200, 7: 11400, 8: 11600, 9: 11800,
	10: 12000, 11: 12200, 12: 12400, 13: 12600, 14: 12800, 15: 13000, 16: 13200, 17: 13400, 18: 13600, 19: 13800,
	20: 14000, 21: 17750, 22: 18000, 23: 18250, 24: 18500, 25: 18750, 26: 19000, 27: 19250, 28: 19500, 29: 19750,
	30: 20000, 31: 20250, 32: 20500, 33: 20750, 34: 21000, 35: 21250, 36: 21500, 37: 21750, 38: 22000, 39: 22250,
	40: 25000, 41: 26000, 42: 27000, 43: 28000, 44: 29000, 45: 30000, 46: 31000, 47: 32000, 48: 33000, 49: 34000,
	50: 35000, 51: 36000, 52: 37000, 53: 38000, 54: 39000, 55: 40000, 56: 41000, 57: 42000, 58: 43000, 59: 44000,
	60: 45000, 61: 46000, 62: 47000, 63: 48000, 64: 49000, 65: 50000, 66: 51000, 67: 52000, 68: 53000, 69: 54000,
	70: 55000, 71: 56000, 72: 57000, 73: 58000, 74: 59000, 75: 60000, 76: 61000, 77: 62000, 78: 63000, 79: 64000,
	80: 65000,
}

// cableLengthWiringTable جدول التسليك حسب طول الكيبل بالمتر (1 إلى 29) — منسوخ
// حرفياً من شيت "تكاليف المشروع". أي طول >=30 متر يستخدم سعر ثابت 710 للمتر.
var cableLengthWiringTable = map[int]float64{
	1: 1000, 2: 1000, 3: 1000, 4: 1000, 5: 800, 6: 800, 7: 800, 8: 800, 9: 800, 10: 800,
	11: 750, 12: 745, 13: 745, 14: 737, 15: 735, 16: 700, 17: 705, 18: 700, 19: 695, 20: 692,
	21: 690, 22: 690, 23: 690, 24: 690, 25: 690, 26: 690, 27: 690, 28: 690, 29: 690,
}

const cableLengthFlatRateAbove30 = 710

// deviceCountWiringPrice يرجّع القيمة من جدول العدد الكلي لأجهزة المشروع، مع
// اقتصاص القيمة على 0/80 لأي عدد خارج المدى المُعرَّف بالشيت.
func deviceCountWiringPrice(totalDeviceCount int) float64 {
	if totalDeviceCount < 0 {
		totalDeviceCount = 0
	}
	if totalDeviceCount > 80 {
		totalDeviceCount = 80
	}
	return deviceCountWiringTable[totalDeviceCount]
}

// cableLengthWiringPricePerMeter يرجّع سعر المتر حسب طول الكيبل الكامل (وليس
// حسب كل متر تراكمياً) — >=30 متر يستخدم السعر الثابت 710/متر.
func cableLengthWiringPricePerMeter(cableLengthMeters int) float64 {
	if cableLengthMeters <= 0 {
		return 0
	}
	if cableLengthMeters >= 30 {
		return cableLengthFlatRateAbove30
	}
	return cableLengthWiringTable[cableLengthMeters]
}

// wiringMultiplierFor يبحث عن مضاعف نوع التسليك (wiring_price_per_meter_multiplier)
// من الكتالوج حسب اسم المنظومة واسم نوع التسليك. لو مو موجود يرجّع 1 (بدون تأثير).
func wiringMultiplierFor(catalog []model.SystemPriceCatalog, systemName, wiringItemName string) float64 {
	if wiringItemName == "" {
		return 1
	}
	for _, row := range catalog {
		if row.Category == "wiring" && row.SystemName == systemName && row.ItemName == wiringItemName {
			return row.Value
		}
	}
	return 1
}

// installPriceFor يبحث عن سعر التركيب الثابت لعنصر معين ضمن منظومة معينة.
func installPriceFor(catalog []model.SystemPriceCatalog, systemName, itemName string) float64 {
	for _, row := range catalog {
		if row.Category == "install" && row.SystemName == systemName && row.ItemName == itemName {
			return row.Value
		}
	}
	return 0
}

// programmingPriceFor يبحث عن سعر خدمة البرمجة/التفعيل الثابت ضمن منظومة معينة.
func programmingPriceFor(catalog []model.SystemPriceCatalog, systemName, programmingItemName string) float64 {
	if programmingItemName == "" {
		return 0
	}
	for _, row := range catalog {
		if row.Category == "programming" && row.SystemName == systemName && row.ItemName == programmingItemName {
			return row.Value
		}
	}
	return 0
}

// CalculateExecutionCost يحسب "تكاليف التنفيذ" لكامل المشروع بنفس منطق شيت
// "تكاليف المشروع" حرفياً:
//  1. لكل بند (منظومة+عنصر) مع count > 0: سعر التركيب × مضاعف الارتفاع.
//  2. تكلفة التسليك = أكبر قيمة بين: (أ) جدول العدد الكلي × مضاعف نوع التسليك،
//     مطروح منه installation_total لتفادي الاحتساب المزدوج (تفسيرنا لطرح "-G7"
//     الغامض بالشيت الأصلي: الشيت يطرح تكلفة التركيب لأنها محسوبة أصلاً بجزء
//     منفصل، فلا نكررها ضمن مجموع التسليك — أي قيمة سالبة الناتجة تُصفّر لأنها
//     لا تعني تكلفة تسليك سالبة)، أو (ب) جدول طول الكيبل × عدد الأمتار.
//  3. سعر البرمجة الثابت لو محدد.
//  4. المجموع الكلي يُقرَّب لأعلى لأقرب 1000 (CEILING).
//
// totalDeviceCount هو العدد الكلي لأجهزة المشروع (وليس عدد بند واحد) — يُستخدم
// كمفتاح موحّد لجدول التسليك حسب العدد الكلي، بنفس ما يفعله الشيت.
func CalculateExecutionCost(items []model.ExecutionCostItem, catalog []model.SystemPriceCatalog, totalDeviceCount int) (int64, error) {
	if len(items) == 0 {
		return 0, fmt.Errorf("لا يمكن حساب تكلفة تنفيذ بدون بنود")
	}

	var sum float64
	deviceCountPrice := deviceCountWiringPrice(totalDeviceCount)

	for _, item := range items {
		if item.Count <= 0 {
			continue // IF(count>0,1,0) — البند غير المُختار لا يُحتسب إطلاقاً
		}

		basePrice := installPriceFor(catalog, item.SystemName, item.ItemName)
		heightWeight := heightWeightMultiplier(item.HeightMeters)
		installationTotal := basePrice * heightWeight
		sum += installationTotal

		if item.WiringItemName != "" {
			mult := wiringMultiplierFor(catalog, item.SystemName, item.WiringItemName)
			deviceBased := deviceCountPrice*mult - installationTotal
			if deviceBased < 0 {
				deviceBased = 0
			}
			lengthBased := cableLengthWiringPricePerMeter(item.CableLengthMeters) * float64(item.CableLengthMeters)
			wiringCost := math.Max(deviceBased, lengthBased)
			sum += wiringCost
		}

		if item.ProgrammingItem != "" {
			sum += programmingPriceFor(catalog, item.SystemName, item.ProgrammingItem)
		}
	}

	// CEILING(sum, 1000) — تقريب لأعلى لأقرب 1000
	rounded := math.Ceil(sum/1000) * 1000
	return int64(rounded), nil
}

// GenerateAccountingCode يبني كوداً محاسبياً فريداً وقابل للتتبع (إعادة تصدير
// من model.GenerateAccountingCode حتى تبقى الدالة قابلة للاستخدام من الخدمة
// كما كانت، بينما الـrepository يستخدم النسخة بـmodel مباشرة لتفادي دائرة
// استيراد service <-> repository).
func GenerateAccountingCode(invoiceID string, createdAt time.Time) string {
	return model.GenerateAccountingCode(invoiceID, createdAt)
}
