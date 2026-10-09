package handler

import (
	"fmt"
	"testing"
	"time"

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

func TestTrackingWorkbook(t *testing.T) {
	i := func(v int) *int { return &v }
	d, _ := time.Parse("2006-01-02", "2026-10-07")
	r := model.VehicleDailyRating{VehicleID: "a", RatedDate: d, Wash: i(4), ExteriorClean: i(4), TechnicalFaults: i(2)}
	f, err := BuildTrackingWorkbook("2026-09-17", "2026-10-16", []trackVehicle{{ID: "a", Name: "ستاركس"}}, []model.VehicleDailyRating{r})
	if err != nil {
		t.Fatal(err)
	}
	if got := f.GetSheetList(); len(got) != 2 || got[0] != "النظام الشهري" || got[1] != "الإحصائيات" {
		t.Fatalf("sheets = %v", got)
	}
	rows, _ := f.GetRows("النظام الشهري")
	if len(rows) != 2 || rows[1][2] != "ستاركس" || rows[1][3] != "تم" {
		t.Fatalf("daily rows = %v", rows)
	}
	want := fmt.Sprintf("%.1f%%", *DailyScore(&r))
	if last := rows[1][len(rows[1])-1]; last != want {
		t.Fatalf("score cell = %q want %q", last, want)
	}
}
