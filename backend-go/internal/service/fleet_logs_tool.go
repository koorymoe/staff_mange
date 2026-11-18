package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// fleetLogsTool (طلب ماتركس 0ef76f3b-5، 11-18): «كم سيارة انغسلت اليوم؟» —
// يقرا سجلات السيارات (غسيل/وقود/دهن/صيانة) ليوم بتوقيت بغداد. قراءة بس.
var fleetLogTypes = map[string]string{"CLEANING": "غسيل", "FUEL": "وقود", "OIL_CHANGE": "تبديل دهن", "MAINTENANCE": "صيانة"}

// fleetDayRange حدود اليوم ببغداد بالـUTC (الأعمدة TIMESTAMP بلا منطقة وتنخزن UTC).
func fleetDayRange(day string, now time.Time) (time.Time, time.Time, error) {
	loc := baghdad
	d := now.In(loc)
	if day != "" {
		t, err := time.ParseInLocation("2006-01-02", day, loc)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("التاريخ لازم YYYY-MM-DD")
		}
		d = t
	}
	from := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, loc).UTC()
	return from, from.Add(24 * time.Hour), nil
}

func fleetLogsTool(db *sqlx.DB) agentTool {
	return agentTool{Name: "fleet_logs_day",
		Desc: "سجلات السيارات ليوم واحد بتوقيت بغداد: كم سيارة انغسلت/تزودت وقود/بدّلت دهن/انصانت، وأسماءها. type فارغ = كل الأنواع. day فارغ = اليوم.",
		Props: map[string]any{
			"type": map[string]any{"type": "string", "enum": []string{"", "CLEANING", "FUEL", "OIL_CHANGE", "MAINTENANCE"}},
			"day":  map[string]any{"type": "string", "description": "YYYY-MM-DD"},
		},
		Run: func(in map[string]any, _ *nameMask) (string, error) {
			typ, _ := in["type"].(string)
			day, _ := in["day"].(string)
			if _, ok := fleetLogTypes[typ]; typ != "" && !ok {
				return "", fmt.Errorf("نوع غير معروف")
			}
			from, to, err := fleetDayRange(day, time.Now())
			if err != nil {
				return "", err
			}
			var rows []struct {
				Type  string `db:"type"`
				Name  string `db:"name"`
				Plate string `db:"plateNumber"`
				N     int    `db:"n"`
			}
			if err := db.Select(&rows, `
				SELECT l.type, v.name, v."plateNumber", COUNT(*) AS n
				FROM "VehicleLog" l JOIN "Vehicle" v ON v.id = l."vehicleId"
				WHERE l."performedAt" >= $1 AND l."performedAt" < $2
				  AND ($3 = '' OR l.type = $3)
				GROUP BY l.type, v.name, v."plateNumber" ORDER BY l.type, v.name`, from, to, typ); err != nil {
				return "", err
			}
			if len(rows) == 0 {
				return "ماكو سجلات بهذا اليوم (ممكن ما انسجلت بالنظام).", nil
			}
			by := map[string][]string{}
			var order []string
			for _, r := range rows {
				if _, ok := by[r.Type]; !ok {
					order = append(order, r.Type)
				}
				by[r.Type] = append(by[r.Type], fmt.Sprintf("%s (%s)", r.Name, r.Plate))
			}
			var b strings.Builder
			for _, t := range order {
				label := fleetLogTypes[t]
				if label == "" {
					label = t
				}
				fmt.Fprintf(&b, "%s: %d سيارة — %s\n", label, len(by[t]), strings.Join(by[t], "، "))
			}
			return strings.TrimSpace(b.String()), nil
		}}
}
