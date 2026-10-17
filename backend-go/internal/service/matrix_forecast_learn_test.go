package service

import (
	"testing"

	"staffmange-api/internal/repository"
)

func TestForecastLearning(t *testing.T) {
	if e, v := judgeForecast(105, 100); v != "CORRECT" || e != 5 {
		t.Fatalf("5%% → %v %s", e, v)
	}
	if _, v := judgeForecast(125, 100); v != "WRONG" {
		t.Fatal("25% should be wrong")
	}
	hist := []repository.ForecastOutcome{{Kind: "REVENUE", ErrorPct: 12, Verdict: "WRONG"}}
	if _, ok := calibration(hist, "REVENUE"); ok {
		t.Fatal("needs 2 outcomes")
	}
	if realConfidence(hist, "REVENUE") != nil {
		t.Fatal("needs 2 outcomes")
	}
	hist = append(hist, repository.ForecastOutcome{Kind: "REVENUE", ErrorPct: 60, Verdict: "WRONG"}, repository.ForecastOutcome{Kind: "REVENUE", ErrorPct: 3, Verdict: "CORRECT"})
	e, ok := calibration(hist, "REVENUE")
	if !ok || e != 25 {
		t.Fatalf("calibration %v", e)
	}
	// المنجز 100 ثابت، الباقي 125 ينقسم على 1.25
	if c := correctForecast(225, 100, 25); c != 200 {
		t.Fatalf("corrected %v", c)
	}
	if c := realConfidence(hist, "REVENUE"); c == nil || *c != 33 {
		t.Fatalf("confidence %v", c)
	}
	if e, _ := calibration([]repository.ForecastOutcome{{Kind: "JOBS", ErrorPct: 90}, {Kind: "JOBS", ErrorPct: 80}}, "JOBS"); e != 30 {
		t.Fatalf("clamp %v", e)
	}
}
