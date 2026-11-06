package service

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/errlog"
)

// ═══ وقت الالتزام ويا الزبون (دازاي 11-06، فكرة ١ — «٤٨ ساعة للكل») ═══
// الموعد = تسجيل الحجز + slaHours الخدمة. ما ينحسب للمنتهي والملغي والداخلي،
// ولا للي بالانتظار (زبون ما رد / الزبون يقرر) — نفس روح قاعدة المتأخرات.

// SlaOpenSQL شرط الحجز «المفتوح والمحسوب» — مكان واحد.
const SlaOpenSQL = `b.status::text NOT IN ('COMPLETED','CANCELLED') AND b."archivedAt" IS NULL
	AND b."waitingSince" IS NULL AND b."bookingType" IS DISTINCT FROM 'INTERNAL'`

type SlaItem struct {
	ID        string    `db:"id" json:"id"`
	Code      string    `db:"code" json:"code"`
	Customer  *string   `db:"customer" json:"customer"`
	Service   *string   `db:"service" json:"service"`
	DueAt     time.Time `db:"dueAt" json:"dueAt"`
	HoursLeft float64   `db:"hoursLeft" json:"hoursLeft"`
}

// SlaBoard الحجوزات الي باقيلها أقل من within ساعة أو فات موعدها — الأقرب أول.
func SlaBoard(db *sqlx.DB, within int) ([]SlaItem, error) {
	rows := []SlaItem{}
	err := db.Select(&rows, `
		SELECT b.id, b.code, c.name AS customer, s.name AS service,
		       b."createdAt" + make_interval(hours => COALESCE(s."slaHours", 48)) AS "dueAt",
		       round((extract(epoch FROM (b."createdAt" + make_interval(hours => COALESCE(s."slaHours", 48)) - now())) / 3600)::numeric, 1)::float8 AS "hoursLeft"
		FROM "Booking" b
		LEFT JOIN "Customer" c ON c.id = b."customerId"
		LEFT JOIN "Service" s ON s.id = b."serviceId"
		WHERE `+SlaOpenSQL+`
		  AND b."createdAt" + make_interval(hours => COALESCE(s."slaHours", 48)) < now() + make_interval(hours => $1)
		ORDER BY 5 DESC LIMIT 200`, within) // الأقرب للموعد أول، وبعدها الي فات توّه، والقديم بالآخر
	return rows, err
}

// SlaAlerter ينبّه مرة وحدة لكل حجز: «قرب» (٦ ساعات) للمنسّقين، و«فات» للإدارة.
type SlaAlerter struct {
	db     *sqlx.DB
	notify func(kind, msg string) // kind: SOON | OVERDUE
}

func NewSlaAlerter(db *sqlx.DB, notify func(kind, msg string)) *SlaAlerter {
	return &SlaAlerter{db: db, notify: notify}
}

func (a *SlaAlerter) Tick() error {
	items, err := SlaBoard(a.db, 6)
	if err != nil {
		return err
	}
	for _, it := range items {
		// فات من أكثر من يوم = قديم (قبل الميزة أو انبلّغ) — يبين باللوحة بس، بلا سيل تنبيهات يوم النشر
		if it.HoursLeft < -24 {
			continue
		}
		kind := "SOON"
		msg := fmt.Sprintf("⏱ الحجز %s باقيله %.0f ساعة على موعد الالتزام ويا الزبون", it.Code, it.HoursLeft)
		if it.HoursLeft < 0 {
			kind = "OVERDUE"
			msg = fmt.Sprintf("🔴 الحجز %s فات موعد الالتزام بـ%.0f ساعة", it.Code, -it.HoursLeft)
		}
		res, err := a.db.Exec(`INSERT INTO "SlaAlert" ("bookingId", kind) VALUES ($1, $2) ON CONFLICT DO NOTHING`, it.ID, kind)
		if err != nil {
			errlog.Err(err)
			continue
		}
		if n, _ := res.RowsAffected(); n == 1 && a.notify != nil {
			a.notify(kind, msg)
		}
	}
	return nil
}
