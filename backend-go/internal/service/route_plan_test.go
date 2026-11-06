package service

import (
	"strings"
	"testing"
)

func TestPlanRouteImproves(t *testing.T) {
	start := RoutePoint{ID: "co", Lat: 32.6, Lng: 44.0}
	// بترتيب المواعيد: بعيد، قريب، أبعد، أقرب — متعرج
	stops := []RoutePoint{{ID: "far", Lat: 32.9, Lng: 44.0}, {ID: "near", Lat: 32.65, Lng: 44.0}, {ID: "farther", Lat: 33.0, Lng: 44.0}, {ID: "nearest", Lat: 32.61, Lng: 44.0}}
	p := PlanRoute(start, stops)
	ids := []string{}
	for _, s := range p.Stops {
		ids = append(ids, s.ID)
	}
	if strings.Join(ids, ",") != "nearest,near,far,farther" {
		t.Fatalf("الترتيب %v", ids)
	}
	if !(p.KM < p.OriginalKM) || p.SavedKM <= 0 {
		t.Fatalf("ما وفّر: %+v", p)
	}
	if len(p.Stops) != len(stops) {
		t.Fatal("ضاع توقف")
	}
	if !strings.Contains(p.MapsURL, "origin=32.600000,44.000000") || !strings.Contains(p.MapsURL, "destination=33.000000,44.000000") || strings.Count(p.MapsURL, "%7C") != 2 {
		t.Fatalf("الرابط %s", p.MapsURL)
	}
}

func TestPlanRouteEdges(t *testing.T) {
	start := RoutePoint{Lat: 32.6, Lng: 44}
	if p := PlanRoute(start, nil); len(p.Stops) != 0 || p.MapsURL != "" {
		t.Fatal("بلا توقفات")
	}
	one := PlanRoute(start, []RoutePoint{{ID: "a", Lat: 32.7, Lng: 44}})
	if len(one.Stops) != 1 || one.SavedKM != 0 || strings.Contains(one.MapsURL, "waypoints") {
		t.Fatalf("%+v", one)
	}
	// كربلاء→بغداد تقريباً ٩٠–١٠٠ كم خط مستقيم
	if d := haversineKM(RoutePoint{Lat: 32.616, Lng: 44.025}, RoutePoint{Lat: 33.315, Lng: 44.366}); d < 80 || d > 100 {
		t.Fatalf("haversine %v", d)
	}
	many := make([]RoutePoint, 14)
	for i := range many {
		many[i] = RoutePoint{Lat: 32.6 + float64(i)*0.01, Lng: 44}
	}
	if u := PlanRoute(start, many).MapsURL; strings.Count(u, "%7C") != 8 {
		t.Fatalf("الحد ٩ توقفات وسطية: %s", u)
	}
}
