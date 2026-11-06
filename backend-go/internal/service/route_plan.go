package service

import (
	"fmt"
	"math"
	"strings"
)

// ═══ ترتيب مسار اليوم للكادر (دازاي 11-06، فكرة ٢ من البحث) ═══
// أقرب جار ثم تحسين 2-opt على مسافة خط مستقيم (haversine) — بلا خدمة خارجية ولا بيانات تطلع برا.
// اقتراح بس: الموعد المتفق ويا الزبون يبقى أهم من المسافة، والليدر يقرر.

type RoutePoint struct {
	ID    string  `db:"id" json:"id"`
	Label string  `db:"label" json:"label"`
	Lat   float64 `db:"lat" json:"lat"`
	Lng   float64 `db:"lng" json:"lng"`
}

type RoutePlan struct {
	Stops      []RoutePoint `json:"stops"`
	KM         float64      `json:"km"`         // المسار المقترح (من الشركة وبين التوقفات)
	OriginalKM float64      `json:"originalKm"` // بترتيب المواعيد
	SavedKM    float64      `json:"savedKm"`
	MapsURL    string       `json:"mapsUrl"`
}

func haversineKM(a, b RoutePoint) float64 {
	const r = 6371.0
	rad := math.Pi / 180
	dLat, dLng := (b.Lat-a.Lat)*rad, (b.Lng-a.Lng)*rad
	h := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(a.Lat*rad)*math.Cos(b.Lat*rad)*math.Sin(dLng/2)*math.Sin(dLng/2)
	return 2 * r * math.Asin(math.Sqrt(h))
}

// pathKM طول المسار من البداية مروراً بالتوقفات (بلا رجعة).
func pathKM(start RoutePoint, stops []RoutePoint) float64 {
	d, prev := 0.0, start
	for _, s := range stops {
		d += haversineKM(prev, s)
		prev = s
	}
	return d
}

// PlanRoute يرتب التوقفات من نقطة البداية. stops بترتيب المواعيد (للمقارنة).
func PlanRoute(start RoutePoint, stops []RoutePoint) RoutePlan {
	n := len(stops)
	order := make([]RoutePoint, 0, n)
	used := make([]bool, n)
	cur := start
	for len(order) < n { // أقرب جار
		best, bestD := -1, math.MaxFloat64
		for i, s := range stops {
			if !used[i] {
				if d := haversineKM(cur, s); d < bestD {
					best, bestD = i, d
				}
			}
		}
		used[best] = true
		order = append(order, stops[best])
		cur = stops[best]
	}
	for improved := true; improved; { // 2-opt: اعكس أي مقطع يقصّر المسار
		improved = false
		for i := 0; i < n-1; i++ {
			for j := i + 1; j < n; j++ {
				cand := append(append(append([]RoutePoint{}, order[:i]...), reversed(order[i:j+1])...), order[j+1:]...)
				if pathKM(start, cand) < pathKM(start, order)-1e-9 {
					order, improved = cand, true
				}
			}
		}
	}
	p := RoutePlan{Stops: order, KM: round1(pathKM(start, order)), OriginalKM: round1(pathKM(start, stops))}
	p.SavedKM = round1(math.Max(0, p.OriginalKM-p.KM))
	p.MapsURL = mapsDirURL(start, order)
	return p
}

func reversed(s []RoutePoint) []RoutePoint {
	out := make([]RoutePoint, len(s))
	for i := range s {
		out[i] = s[len(s)-1-i]
	}
	return out
}

// mapsDirURL رابط خرائط Google بكل التوقفات (الحد ٩ توقفات وسطية).
func mapsDirURL(start RoutePoint, stops []RoutePoint) string {
	if len(stops) == 0 {
		return ""
	}
	ll := func(p RoutePoint) string { return fmt.Sprintf("%.6f,%.6f", p.Lat, p.Lng) }
	u := "https://www.google.com/maps/dir/?api=1&travelmode=driving&origin=" + ll(start) + "&destination=" + ll(stops[len(stops)-1])
	mid := stops[:len(stops)-1]
	if len(mid) > 9 {
		mid = mid[:9]
	}
	if len(mid) > 0 {
		w := make([]string, len(mid))
		for i, p := range mid {
			w[i] = ll(p)
		}
		u += "&waypoints=" + strings.Join(w, "%7C")
	}
	return u
}
