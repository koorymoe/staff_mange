package service

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// ═══ صيانة وقائية (دازاي 11-06، فكرة ٧) ═══
// زبائن آخر شغل منجز إلهم فات عليه ٦ أشهر (ولحد ٣ سنين — الأقدم غالباً راح)، أو عندهم
// شكوتين فما فوق بسنة. ما عندهم حجز مفتوح هسه. قراءة بس — الجودة تقرر تتصل لو لا.

type MaintenanceDue struct {
	CustomerID string    `db:"customerId" json:"customerId"`
	Name       string    `db:"name" json:"name"`
	Phone      *string   `db:"phone" json:"phone,omitempty"`
	LastDone   time.Time `db:"lastDone" json:"lastDone"`
	Service    *string   `db:"service" json:"service"`
	Months     int       `db:"months" json:"months"`
	Complaints int       `db:"complaints" json:"complaints"`
	Reason     string    `db:"reason" json:"reason"`
}

func MaintenanceDueList(db *sqlx.DB, limit int) ([]MaintenanceDue, error) {
	rows := []MaintenanceDue{}
	err := db.Select(&rows, `
		WITH last AS (
			SELECT DISTINCT ON (b."customerId") b."customerId", b."completedAt" AS "lastDone", s.name AS service
			FROM "Booking" b LEFT JOIN "Service" s ON s.id = b."serviceId"
			WHERE b.status = 'COMPLETED' AND b."completedAt" IS NOT NULL AND b."bookingType" IS DISTINCT FROM 'INTERNAL'
			ORDER BY b."customerId", b."completedAt" DESC
		), comp AS (
			SELECT "customerId", count(*) AS n FROM "Complaint" WHERE "createdAt" > now() - interval '365 days' GROUP BY 1
		)
		SELECT c.id AS "customerId", c.name, c.phone, l."lastDone", l.service,
		       (extract(year FROM age(now(), l."lastDone")) * 12 + extract(month FROM age(now(), l."lastDone")))::int AS months,
		       COALESCE(comp.n, 0)::int AS complaints,
		       CASE WHEN COALESCE(comp.n, 0) >= 2 THEN 'شكاوى متكررة' ELSE 'فات ٦ أشهر على آخر شغل' END AS reason
		FROM last l
		JOIN "Customer" c ON c.id = l."customerId"
		LEFT JOIN comp ON comp."customerId" = l."customerId"
		WHERE (l."lastDone" < now() - interval '6 months' AND l."lastDone" > now() - interval '3 years' OR COALESCE(comp.n, 0) >= 2)
		  AND NOT EXISTS (SELECT 1 FROM "Booking" o WHERE o."customerId" = l."customerId" AND o.status::text NOT IN ('COMPLETED','CANCELLED') AND o."archivedAt" IS NULL) -- (11-07) حجز مؤرشف مو مفتوح
		ORDER BY COALESCE(comp.n, 0) DESC, l."lastDone" ASC
		LIMIT $1`, limit)
	return rows, err
}

// maintenanceSummary ملخص بلا أسماء ولا تلفونات — للذكاء (أسماء الزبائن ما تطلع برا).
func maintenanceSummary(rows []MaintenanceDue) string {
	by := map[string]int{}
	reasons := map[string]int{}
	for _, r := range rows {
		svc := "بلا خدمة"
		if r.Service != nil && *r.Service != "" {
			svc = *r.Service
		}
		by[svc]++
		reasons[r.Reason]++
	}
	keys := make([]string, 0, len(by))
	for k := range by {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return by[keys[i]] > by[keys[j]] })
	var b strings.Builder
	fmt.Fprintf(&b, "%d زبون يستاهل اتصال صيانة (القائمة بالأسماء بمكتب الجودة).\n", len(rows))
	for r, n := range reasons {
		fmt.Fprintf(&b, "- %s: %d\n", r, n)
	}
	b.WriteString("حسب الخدمة:\n")
	for _, k := range keys {
		fmt.Fprintf(&b, "- %s: %d\n", k, by[k])
	}
	return b.String()
}
