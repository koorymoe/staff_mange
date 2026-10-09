package handler

import (
	"testing"

	"staffmange-api/internal/model"
)

func ip(v int) *int { return &v }

// صف ١٧/٠٩ «بنگو الشعبة الهندسية» بملف الإكسل: (1.96 / 3.68) = 53.26%.
func TestDailyScoreMatchesExcel(t *testing.T) {
	r := model.VehicleDailyRating{Wash: ip(0), ExteriorClean: ip(2), ExteriorCondition: ip(3), TireCondition: ip(2), GlassClean: ip(2),
		LightsCondition: ip(3), TechnicalFaults: ip(3), InteriorClean: ip(1), SeatsCondition: ip(2), InteriorDirt: ip(1), Smell: ip(0)}
	s := DailyScore(&r)
	if s == nil || *s < 53.25 || *s > 53.27 {
		t.Fatalf("got %v", s)
	}
}

func TestTrackStatsWashAndGrade(t *testing.T) {
	full := func(w int) model.VehicleDailyRating {
		return model.VehicleDailyRating{Wash: ip(w), ExteriorClean: ip(4), ExteriorCondition: ip(4), TireCondition: ip(4), GlassClean: ip(4),
			LightsCondition: ip(4), TechnicalFaults: ip(4), InteriorClean: ip(4), SeatsCondition: ip(4), InteriorDirt: ip(4), Smell: ip(4)}
	}
	st := BuildTrackStats("v", "س", []model.VehicleDailyRating{full(4), full(0)})
	if st.Completed != 2 || st.WashDone != 1 || st.WashNot != 1 || st.WashPct == nil || *st.WashPct != 50 {
		t.Fatalf("%+v", st)
	}
	// (50×0.08 + 100×0.92) / 1 = 96 → ممتاز
	if st.Weighted == nil || *st.Weighted < 95.99 || *st.Weighted > 96.01 || st.Grade != "ممتاز" {
		t.Fatalf("weighted %v %s", st.Weighted, st.Grade)
	}
}
