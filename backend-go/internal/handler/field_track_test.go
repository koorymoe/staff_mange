package handler

import (
	"testing"
	"time"
)

func TestAnalyzeTrack(t *testing.T) {
	base := time.Date(2026, 10, 14, 9, 0, 0, 0, time.UTC)
	company := [2]float64{32.6160, 44.0249}
	customer := [2]float64{32.6500, 44.0600}
	var pts []trackPoint
	add := func(min int, lat, lng float64) {
		pts = append(pts, trackPoint{lat, lng, base.Add(time.Duration(min) * time.Minute)})
	}
	add(0, company[0], company[1])
	add(5, company[0], company[1])
	add(10, 32.6200, 44.0300)   // طلع (أكثر من 300 م)
	for m := 12; m <= 17; m++ { // توقف بالطريق 5 دقايق
		add(m, 32.6300, 44.0400)
	}
	add(20, 32.6400, 44.0500)
	for m := 25; m <= 70; m += 5 { // عند الزبون 45 دقيقة (على بعد ~700 م من المسجّل)
		add(m, 32.6560, 44.0620)
	}
	add(75, 32.6400, 44.0500) // رجع
	r := analyzeTrack(pts, &company, &customer)
	if r.Departed == nil || !r.Departed.Equal(base.Add(10*time.Minute)) {
		t.Fatalf("departed %v", r.Departed)
	}
	if r.Arrived == nil || !r.Arrived.Equal(base.Add(25*time.Minute)) {
		t.Fatalf("arrived %v", r.Arrived)
	}
	if r.Completed == nil || !r.Completed.Equal(base.Add(70*time.Minute)) {
		t.Fatalf("completed %v", r.Completed)
	}
	if len(r.RoadStops) != 1 || r.RoadStops[0].Minutes() != 5 {
		t.Fatalf("stops %+v", r.RoadStops)
	}
	// توقف 20 دقيقة بس → ما ينحسب وصول
	r2 := analyzeTrack(pts[:9], &company, &customer)
	if r2.Arrived != nil {
		t.Fatalf("should not arrive")
	}
}

func TestCurrentStop(t *testing.T) {
	base := time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC)
	pts := []trackPoint{{32.60, 44.00, base}, {32.62, 44.03, base.Add(5 * time.Minute)}}
	for m := 10; m <= 22; m += 2 {
		pts = append(pts, trackPoint{32.65, 44.06, base.Add(time.Duration(m) * time.Minute)})
	}
	s := currentStop(pts)
	if s == nil || s.Minutes() != 12 {
		t.Fatalf("stop %+v", s)
	}
}
