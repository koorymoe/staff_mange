package repository

import (
	"math"
	"testing"

	"staffmange-api/internal/model"
)

// computeWeightedScore: الغسل ما يدخل بنسبة اليوم، البنود الفاضية تنشال من المقام، وكلشي ٤ = ١٠٠٪.
func TestVehicleWeightedScore(t *testing.T) {
	i := func(v int) *int { return &v }
	if computeWeightedScore(&model.VehicleDailyRating{Wash: i(4)}) != nil {
		t.Fatal("الغسل وحده ما يسوي نسبة يوم")
	}
	full := &model.VehicleDailyRating{Wash: i(0), ExteriorClean: i(4), ExteriorCondition: i(4), TireCondition: i(4), GlassClean: i(4),
		LightsCondition: i(4), TechnicalFaults: i(4), InteriorClean: i(4), SeatsCondition: i(4), InteriorDirt: i(4), Smell: i(4)}
	if s := computeWeightedScore(full); s == nil || math.Abs(*s-1) > 1e-9 {
		t.Fatalf("كلشي ٤ لازم ١: %v", s)
	}
	// بند واحد بنص الدرجة = ٠٫٥ مهما كان وزنه
	if s := computeWeightedScore(&model.VehicleDailyRating{TechnicalFaults: i(2)}); s == nil || math.Abs(*s-0.5) > 1e-9 {
		t.Fatalf("%v", s)
	}
	// الأعطال الفنية (وزن ٠٫٢٥) تأثر أكثر من الزجاج (٠٫٠٥)
	a := computeWeightedScore(&model.VehicleDailyRating{TechnicalFaults: i(0), GlassClean: i(4)})
	b := computeWeightedScore(&model.VehicleDailyRating{TechnicalFaults: i(4), GlassClean: i(0)})
	if !(*a < *b) {
		t.Fatalf("الوزن ما يشتغل: %v %v", *a, *b)
	}
}
