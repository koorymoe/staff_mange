package handler

import "time"

// ═══ خوارزمية المسار — قرار (ع) 10-14 ═══
//
// النظام يحسب بالخفية، بجنب ضغطات الموظف، من نقاط الـGPS:
//   - الانطلاق: أول نقطة تبعد أكثر من 300 م عن نقطة البداية (الشركة أو الحجز السابق).
//   - الوصول: أول توقف (حركة ضمن 100 م) يدوم 30 دقيقة فأكثر وقريب من الزبون
//     (ضمن 1.5 كم، لأن موقع الزبون المسجّل ممكن يغلط كيلو). وقت الوصول =
//     **بداية** التوقف — هنا يوكف وقت الطريق ويبدي وقت العمل.
//   - الإنجاز: نهاية ذاك التوقف (لما يتحرك ويبعد عنه).
//   - التوقفات بالطريق: أي توقف 3 دقايق فأكثر قبل الوصول، بمكانه ومدته.

type trackPoint struct {
	Lat, Lng float64
	At       time.Time
}

type trackStop struct {
	Lat, Lng   float64
	Start, End time.Time
}

func (s trackStop) Minutes() int { return int(s.End.Sub(s.Start).Minutes()) }

type trackResult struct {
	Departed  *time.Time
	Arrived   *time.Time
	Completed *time.Time
	ArrivalAt *trackStop
	RoadStops []trackStop
}

const (
	stopRadiusM   = 100.0
	departRadiusM = 300.0
	arrivalMin    = 30 * time.Minute
	roadStopMin   = 3 * time.Minute
	customerNearM = 1500.0
)

// clusterStops يقسم النقاط لتوقفات: نقاط متتالية ضمن 100 م من أول نقطة بالمجموعة.
func clusterStops(pts []trackPoint) []trackStop {
	var out []trackStop
	i := 0
	for i < len(pts) {
		j := i
		sumLat, sumLng := 0.0, 0.0
		for j < len(pts) && haversineM(pts[i].Lat, pts[i].Lng, pts[j].Lat, pts[j].Lng) <= stopRadiusM {
			sumLat += pts[j].Lat
			sumLng += pts[j].Lng
			j++
		}
		n := float64(j - i)
		out = append(out, trackStop{Lat: sumLat / n, Lng: sumLng / n, Start: pts[i].At, End: pts[j-1].At})
		i = j
	}
	return out
}

// analyzeTrack: origin = نقطة البداية (nil = أول نقطة)، customer = موقع الزبون (nil = غير معروف).
func analyzeTrack(pts []trackPoint, origin, customer *[2]float64) trackResult {
	var res trackResult
	if len(pts) == 0 {
		return res
	}
	o := [2]float64{pts[0].Lat, pts[0].Lng}
	if origin != nil {
		o = *origin
	}
	start := 0
	for k, p := range pts {
		if haversineM(o[0], o[1], p.Lat, p.Lng) > departRadiusM {
			t := p.At
			res.Departed = &t
			start = k
			break
		}
	}
	if res.Departed == nil {
		return res
	}
	stops := clusterStops(pts[start:])
	for idx, s := range stops {
		dur := s.End.Sub(s.Start)
		near := customer == nil || haversineM(customer[0], customer[1], s.Lat, s.Lng) <= customerNearM
		if dur >= arrivalMin && near {
			st := s
			res.ArrivalAt = &st
			a := s.Start
			res.Arrived = &a
			// الإنجاز بس إذا أكو حركة بعد التوقف (يعني طلع من عند الزبون)
			if idx < len(stops)-1 {
				c := s.End
				res.Completed = &c
			}
			break
		}
		if dur >= roadStopMin {
			res.RoadStops = append(res.RoadStops, s)
		}
	}
	return res
}
