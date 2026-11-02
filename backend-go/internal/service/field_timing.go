package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/repository"
)

// ═══ ⏱️ أوقات الميدان + تحليل ماتركس (دازاي 11-02) ═══
// «الموظفين الي بالميدان: شوكت انطلقوا، شوكت وصلوا، شوكت ضغطوا بدء العمل —
// وماتركس يحلل الوقت». كل رقم يتقارن ويا المعتاد (وسيط آخر ١٨٠ يوم).

type FieldTime struct {
	BookingID   string     `db:"id" json:"bookingId"`
	Code        string     `db:"code" json:"code"`
	ServiceID   *string    `db:"serviceId" json:"-"`
	ScheduledAt *time.Time `db:"scheduledAt" json:"scheduledAt"`
	DepartedAt  *time.Time `db:"departedAt" json:"departedAt"`
	ArrivedAt   *time.Time `db:"arrivedAt" json:"arrivedAt"`
	StartedAt   *time.Time `db:"startedAt" json:"startedAt"`
	RoadMin     *float64   `db:"roadMed" json:"-"`
	WorkMin     *float64   `db:"workMed" json:"-"`
	Alert       bool       `json:"alert"`
	Notes       []string   `json:"notes"`
}

func durAr(d time.Duration) string {
	m := int(d.Minutes() + 0.5)
	if m < 0 {
		m = -m
	}
	if m >= 60 {
		return fmt.Sprintf("%dس %dد", m/60, m%60)
	}
	return fmt.Sprintf("%dد", m)
}

// FieldTiming الحجوزات بالميدان هسه (أو حجز بكوده) ويا التحليل.
func FieldTiming(db *sqlx.DB, code string) ([]FieldTime, error) {
	where := repository.InFieldSQL
	args := []any{}
	if code != "" {
		where, args = `upper(b.code) = upper($1)`, append(args, code)
	}
	rows := []FieldTime{}
	err := db.Select(&rows, `
		WITH road AS (SELECT percentile_cont(0.5) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM (COALESCE("arrivedAt","autoArrivedAt") - "departedAt"))/60) AS m
		              FROM "Booking" WHERE "departedAt" IS NOT NULL AND COALESCE("arrivedAt","autoArrivedAt") > "departedAt"
		                AND COALESCE("arrivedAt","autoArrivedAt") - "departedAt" < interval '4 hours' AND "departedAt" > now() - interval '180 days'),
		     work AS (SELECT "serviceId", percentile_cont(0.5) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM ("completedAt" - "startedAt"))/60) AS m, count(*) AS n
		              FROM "Booking" WHERE status = 'COMPLETED' AND "startedAt" IS NOT NULL AND "completedAt" > "startedAt"
		                AND "completedAt" - "startedAt" < interval '14 hours' AND "completedAt" > now() - interval '180 days' GROUP BY "serviceId")
		SELECT b.id, b.code, b."serviceId", b."scheduledAt", b."departedAt", COALESCE(b."arrivedAt", b."autoArrivedAt") AS "arrivedAt", b."startedAt",
		       (SELECT m FROM road) AS "roadMed", (SELECT m FROM work w WHERE w."serviceId" = b."serviceId" AND w.n >= 3) AS "workMed"
		FROM "Booking" b WHERE `+where+` ORDER BY COALESCE(b."startedAt", b."arrivedAt", b."departedAt") LIMIT 100`, args...)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	for i := range rows {
		analyzeField(&rows[i], now)
	}
	return rows, nil
}

// الأوقات مخزونة UTC بلا منطقة — نقارنها ويا now() UTC.
func analyzeField(f *FieldTime, now time.Time) {
	n := []string{}
	if f.DepartedAt != nil && f.ArrivedAt != nil && f.ArrivedAt.After(*f.DepartedAt) {
		d := f.ArrivedAt.Sub(*f.DepartedAt)
		s := "الطريق " + durAr(d)
		if f.RoadMin != nil && *f.RoadMin > 0 {
			s += " (المعتاد " + durAr(time.Duration(*f.RoadMin*float64(time.Minute))) + ")"
			if d.Minutes() > *f.RoadMin*1.5 && d > 20*time.Minute {
				s += " ⚠️"
				f.Alert = true
			}
		}
		n = append(n, s)
	} else if f.DepartedAt != nil && f.ArrivedAt == nil && f.StartedAt == nil {
		d := now.Sub(*f.DepartedAt)
		s := "بالطريق من " + durAr(d)
		if f.RoadMin != nil && d.Minutes() > *f.RoadMin*2 && d > 30*time.Minute {
			s += " ⚠️ طوّل"
			f.Alert = true
		}
		n = append(n, s)
	}
	if f.ScheduledAt != nil && f.ArrivedAt != nil {
		if late := f.ArrivedAt.Sub(*f.ScheduledAt); late > 15*time.Minute {
			n = append(n, "وصل متأخر "+durAr(late)+" عن الموعد ⚠️")
			f.Alert = true
		} else if late < -30*time.Minute {
			n = append(n, "وصل قبل الموعد بـ "+durAr(late))
		} else {
			n = append(n, "وصل بالموعد ✓")
		}
	}
	if f.ArrivedAt != nil && f.StartedAt != nil {
		if w := f.StartedAt.Sub(*f.ArrivedAt); w > 20*time.Minute {
			n = append(n, "انتظر "+durAr(w)+" قبل ما يبدي ⚠️")
			f.Alert = true
		}
	} else if f.ArrivedAt != nil && f.StartedAt == nil {
		if w := now.Sub(*f.ArrivedAt); w > 20*time.Minute {
			n = append(n, "واصل من "+durAr(w)+" وما ضغط بدء العمل ⚠️")
			f.Alert = true
		}
	}
	if f.StartedAt != nil {
		d := now.Sub(*f.StartedAt)
		s := "يشتغل من " + durAr(d)
		if f.WorkMin != nil && *f.WorkMin > 0 {
			s += " (المعتاد " + durAr(time.Duration(*f.WorkMin*float64(time.Minute))) + ")"
			if d.Minutes() > *f.WorkMin*1.5 {
				s += " ⚠️ طوّل"
				f.Alert = true
			}
		}
		n = append(n, s)
	}
	if f.DepartedAt == nil && (f.ArrivedAt != nil || f.StartedAt != nil) {
		n = append(n, "ما ضغط «انطلقنا»")
	}
	if len(n) == 0 {
		n = append(n, "ماكو أوقات كافية للتحليل")
	}
	f.Notes = n
}

// FieldTimingText للأداة.
func FieldTimingText(rows []FieldTime) string {
	if len(rows) == 0 {
		return "ماكو حجوزات بالميدان هسه."
	}
	t := func(x *time.Time) string {
		if x == nil {
			return "—"
		}
		return x.In(debriefLoc).Format("15:04")
	}
	out := []string{}
	for _, r := range rows {
		out = append(out, fmt.Sprintf("%s: انطلق %s · وصل %s · بدأ العمل %s · الموعد %s — %s", r.Code, t(r.DepartedAt), t(r.ArrivedAt), t(r.StartedAt), t(r.ScheduledAt), strings.Join(r.Notes, "، ")))
	}
	return strings.Join(out, "\n")
}
