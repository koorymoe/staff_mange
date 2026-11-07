package repository

import (
	"time"

	"github.com/jmoiron/sqlx"
)

// ═══ 🔁 الرجعات — (ع) 11-07 ═══
// «رجعة» = حجز ثاني لنفس الزبون ونفس الخدمة ينفتح خلال RevisitWindowDays من إنجاز
// الأول. أهم مقياس لجودة الشغل: الشغل الي انعمل صح ما يرجع. تنحسب على ليدر
// الحجز الأول. ⚠️ رقم للمتابعة والتدريب بس — مو غرامة (الغرامات بالقواعد حصراً).
// المستثنى: الاستيراد القديم (OLD-)، والشغل الداخلي، والملغي، والمحذوف/المؤرشف.

const RevisitWindowDays = 30

type RevisitLeaderRow struct {
	LeaderID   string  `db:"leaderId" json:"leaderId"`
	LeaderName string  `db:"leaderName" json:"leaderName"`
	Jobs       int     `db:"jobs" json:"jobs"`
	Revisits   int     `db:"revisits" json:"revisits"`
	Rate       float64 `db:"rate" json:"rate"`
}

type RevisitItem struct {
	FirstID         string    `db:"firstId" json:"firstId"`
	FirstCode       string    `db:"firstCode" json:"firstCode"`
	FirstDoneAt     time.Time `db:"firstDoneAt" json:"firstDoneAt"`
	SecondID        string    `db:"secondId" json:"secondId"`
	SecondCode      string    `db:"secondCode" json:"secondCode"`
	SecondCreatedAt time.Time `db:"secondCreatedAt" json:"secondCreatedAt"`
	GapDays         int       `db:"gapDays" json:"gapDays"`
	ServiceName     *string   `db:"serviceName" json:"serviceName"`
	CustomerName    *string   `db:"customerName" json:"customerName"`
	LeaderName      *string   `db:"leaderName" json:"leaderName"`
}

type RevisitReport struct {
	WindowDays int                `json:"windowDays"`
	Days       int                `json:"days"`
	Jobs       int                `json:"jobs"`
	Revisits   int                `json:"revisits"`
	Leaders    []RevisitLeaderRow `json:"leaders"`
	Items      []RevisitItem      `json:"items"`
}

type RevisitRepository struct{ db *sqlx.DB }

func NewRevisitRepository(db *sqlx.DB) *RevisitRepository { return &RevisitRepository{db: db} }

// الحجز الأول: منجز خلال آخر $1 يوم، مو قديم ولا داخلي.
func revisitBaseSQL() string {
	return `WITH firsts AS (
		SELECT b.id, b.code, b."customerId", b."serviceId", b."completedAt", booking_leader_id(b.id) AS "leaderId"
		FROM "Booking" b
		WHERE b.status = 'COMPLETED' AND b."completedAt" IS NOT NULL AND ` + BookingCountableSQL("b") + `
		  AND b.code NOT LIKE 'OLD-%' AND COALESCE(b."bookingType"::text, '') <> 'INTERNAL'
		  AND b."serviceId" IS NOT NULL
		  AND b."completedAt" >= now() - make_interval(days => $1)
	), pairs AS (
		SELECT DISTINCT ON (f.id) f.id AS "firstId", f.code AS "firstCode", f."completedAt" AS "firstDoneAt",
		       f."customerId", f."serviceId", f."leaderId",
		       n.id AS "secondId", n.code AS "secondCode", n."createdAt" AS "secondCreatedAt"
		FROM firsts f
		JOIN "Booking" n ON n."customerId" = f."customerId" AND n."serviceId" = f."serviceId" AND n.id <> f.id
		 AND n."createdAt" > f."completedAt"
		 AND n."createdAt" <= f."completedAt" + interval '` + "30" + ` days'
		 AND n.status <> 'CANCELLED' AND ` + BookingCountableSQL("n") + `
		 AND COALESCE(n."bookingType"::text, '') <> 'INTERNAL'
		ORDER BY f.id, n."createdAt"
	)`
}

// Report الرجعات لآخر days يوم (حجوزات أولى انجزت بهالمدة).
func (r *RevisitRepository) Report(days int) (*RevisitReport, error) {
	if days <= 0 || days > 365 {
		days = 90
	}
	rep := &RevisitReport{WindowDays: RevisitWindowDays, Days: days, Leaders: []RevisitLeaderRow{}, Items: []RevisitItem{}}
	if err := r.db.Select(&rep.Leaders, revisitBaseSQL()+`
		SELECT f."leaderId", COALESCE(e.name, '—') AS "leaderName",
		       count(*)::int AS jobs, count(p."firstId")::int AS revisits,
		       round(100.0 * count(p."firstId") / count(*), 1)::float8 AS rate
		FROM firsts f
		LEFT JOIN pairs p ON p."firstId" = f.id
		LEFT JOIN "Employee" e ON e.id = f."leaderId"
		WHERE f."leaderId" IS NOT NULL
		GROUP BY f."leaderId", e.name
		ORDER BY rate DESC, revisits DESC, jobs DESC`, days); err != nil {
		return nil, err
	}
	if err := r.db.Get(&rep.Jobs, revisitBaseSQL()+` SELECT count(*)::int FROM firsts`, days); err != nil {
		return nil, err
	}
	if err := r.db.Select(&rep.Items, revisitBaseSQL()+`
		SELECT p."firstId", p."firstCode", p."firstDoneAt", p."secondId", p."secondCode", p."secondCreatedAt",
		       GREATEST(0, (baghdad_date(p."secondCreatedAt") - baghdad_date(p."firstDoneAt")))::int AS "gapDays",
		       s.name AS "serviceName", c.name AS "customerName", e.name AS "leaderName"
		FROM pairs p
		LEFT JOIN "Service" s ON s.id = p."serviceId"
		LEFT JOIN "Customer" c ON c.id = p."customerId"
		LEFT JOIN "Employee" e ON e.id = p."leaderId"
		ORDER BY p."secondCreatedAt" DESC LIMIT 300`, days); err != nil {
		return nil, err
	}
	rep.Revisits = len(rep.Items)
	return rep, nil
}
