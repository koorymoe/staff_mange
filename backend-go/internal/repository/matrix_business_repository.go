package repository

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// MatrixBusinessRepository أرقام الشركة: إيراد، توقع، زبون فعلي مقابل مستفسر.
type MatrixBusinessRepository struct{ db *sqlx.DB }

func NewMatrixBusinessRepository(db *sqlx.DB) *MatrixBusinessRepository {
	return &MatrixBusinessRepository{db: db}
}

type MonthRevenue struct {
	Month    string  `db:"month" json:"month"`
	Bookings int     `db:"bookings" json:"bookings"`
	Revenue  float64 `db:"revenue" json:"revenue"`
	Invoiced int     `db:"invoiced" json:"invoiced"`
}

// completedCountable حجوزات منجزة تنحسب (مو داخلية، مو مؤرشفة).
const completedCountable = `b.status = 'COMPLETED' AND b."completedAt" IS NOT NULL AND b."archivedAt" IS NULL AND b."bookingType" IS DISTINCT FROM 'INTERNAL'`

// Monthly آخر n أشهر: حجوزات منجزة، وصافي فواتيرها (آخر فاتورة لكل حجز).
func (r *MatrixBusinessRepository) Monthly(n int) ([]MonthRevenue, error) {
	rows := []MonthRevenue{}
	err := r.db.Select(&rows, `
		SELECT to_char(date_trunc('month', baghdad_date(b."completedAt")), 'YYYY-MM') AS month,
		       COUNT(*) AS bookings,
		       COALESCE(SUM(li."netTotal"), 0) AS revenue,
		       COUNT(li.id) AS invoiced
		FROM "Booking" b
		LEFT JOIN LATERAL (SELECT id, "netTotal" FROM "LeaderInvoice" WHERE "bookingId" = b.id ORDER BY "createdAt" DESC LIMIT 1) li ON true
		WHERE `+completedCountable+`
		  AND baghdad_date(b."completedAt") >= (date_trunc('month', baghdad_today()) - make_interval(months => $1 - 1))::date
		GROUP BY 1 ORDER BY 1`, n)
	return rows, err
}

type MonthToDate struct {
	Bookings     int     `db:"bookings" json:"bookings"`
	Revenue      float64 `db:"revenue" json:"revenue"`
	LastBookings int     `db:"lastBookings" json:"lastBookings"`
	LastRevenue  float64 `db:"lastRevenue" json:"lastRevenue"`
}

// MTD هالشهر لحد اليوم مقابل نفس الفترة الشهر الماضي.
func (r *MatrixBusinessRepository) MTD() (*MonthToDate, error) {
	var m MonthToDate
	err := r.db.Get(&m, `
		WITH x AS (
		  SELECT baghdad_date(b."completedAt") AS d, COALESCE(li."netTotal",0) AS net
		  FROM "Booking" b
		  LEFT JOIN LATERAL (SELECT "netTotal" FROM "LeaderInvoice" WHERE "bookingId" = b.id ORDER BY "createdAt" DESC LIMIT 1) li ON true
		  WHERE `+completedCountable+`)
		SELECT
		  COUNT(*) FILTER (WHERE d >= date_trunc('month', baghdad_today())::date) AS bookings,
		  COALESCE(SUM(net) FILTER (WHERE d >= date_trunc('month', baghdad_today())::date),0) AS revenue,
		  COUNT(*) FILTER (WHERE d >= (date_trunc('month', baghdad_today()) - interval '1 month')::date
		                     AND d <= (baghdad_today() - interval '1 month')::date) AS "lastBookings",
		  COALESCE(SUM(net) FILTER (WHERE d >= (date_trunc('month', baghdad_today()) - interval '1 month')::date
		                     AND d <= (baghdad_today() - interval '1 month')::date),0) AS "lastRevenue"
		FROM x`)
	return &m, err
}

type ForecastBase struct {
	PendingThisMonth int             `db:"pendingThisMonth" json:"pendingThisMonth"`
	CompletionRate   sql.NullFloat64 `db:"completionRate" json:"-"`
	AvgInvoice       sql.NullFloat64 `db:"avgInvoice" json:"-"`
	Samples          int             `db:"samples" json:"samples"`
}

// Forecast أساس التوقع: المثبّت المتبقي هالشهر، ونسبة إنجاز المثبّت (٩٠ يوم)، ومعدل الفاتورة.
func (r *MatrixBusinessRepository) Forecast() (*ForecastBase, error) {
	var f ForecastBase
	err := r.db.Get(&f, `
		SELECT
		  (SELECT COUNT(*) FROM "Booking" b WHERE b.status IN ('CONFIRMED','IN_PROGRESS','PENDING') AND b."archivedAt" IS NULL
		     AND b."bookingType" IS DISTINCT FROM 'INTERNAL' AND b."scheduledAt" IS NOT NULL
		     AND baghdad_date(b."scheduledAt") BETWEEN baghdad_today() AND (date_trunc('month', baghdad_today()) + interval '1 month - 1 day')::date) AS "pendingThisMonth",
		  (SELECT COUNT(*) FILTER (WHERE status = 'COMPLETED')::float / NULLIF(COUNT(*),0)
		     FROM "Booking" WHERE "confirmedAt" > now() - interval '90 days' AND "confirmedAt" < now() - interval '7 days'
		       AND "bookingType" IS DISTINCT FROM 'INTERNAL') AS "completionRate",
		  (SELECT AVG("netTotal") FROM "LeaderInvoice" WHERE "createdAt" > now() - interval '90 days' AND "netTotal" > 0) AS "avgInvoice",
		  (SELECT COUNT(*) FROM "LeaderInvoice" WHERE "createdAt" > now() - interval '90 days' AND "netTotal" > 0) AS samples`)
	return &f, err
}

type InquiryCustomer struct {
	ID       string `db:"id" json:"id"`
	Name     string `db:"name" json:"name"`
	Archived int    `db:"archived" json:"archived"`
	LastAt   string `db:"lastAt" json:"lastAt"`
}

// customerKind: زبون فعلي = عنده حجز منجز. مستفسر = كل حجوزاته مؤرشفة بلا منجز.
const inquiryCTE = `
	WITH c AS (
	  SELECT b."customerId" AS id,
	         COUNT(*) FILTER (WHERE b.status = 'COMPLETED' AND b."archivedAt" IS NULL) AS done,
	         COUNT(*) FILTER (WHERE b."archivedAt" IS NOT NULL) AS archived,
	         COUNT(*) FILTER (WHERE b."archivedAt" IS NULL AND b.status <> 'COMPLETED' AND b.status <> 'CANCELLED') AS open,
	         MAX(b."createdAt") AS "lastAt"
	  FROM "Booking" b WHERE b."customerId" IS NOT NULL AND b."bookingType" IS DISTINCT FROM 'INTERNAL'
	  GROUP BY b."customerId")`

// RepeatInquirers زبائن استفسروا مرتين فأكثر وما نفّذوا ولا شي.
func (r *MatrixBusinessRepository) RepeatInquirers(limit int) ([]InquiryCustomer, error) {
	rows := []InquiryCustomer{}
	err := r.db.Select(&rows, inquiryCTE+`
		SELECT cu.id, cu.name, c.archived, to_char(c."lastAt", 'YYYY-MM-DD') AS "lastAt"
		FROM c JOIN "Customer" cu ON cu.id = c.id
		WHERE c.done = 0 AND c.archived >= 2
		ORDER BY c.archived DESC, c."lastAt" DESC LIMIT $1`, limit)
	return rows, err
}

type CustomerSplit struct {
	Actual   int `db:"actual" json:"actual"`
	Inquiry  int `db:"inquiry" json:"inquiry"`
	Repeat   int `db:"repeat" json:"repeat"`
	OpenOnly int `db:"openOnly" json:"openOnly"`
}

func (r *MatrixBusinessRepository) Split() (*CustomerSplit, error) {
	var s CustomerSplit
	err := r.db.Get(&s, inquiryCTE+`
		SELECT COUNT(*) FILTER (WHERE done > 0) AS actual,
		       COUNT(*) FILTER (WHERE done = 0 AND archived > 0 AND open = 0) AS inquiry,
		       COUNT(*) FILTER (WHERE done = 0 AND archived >= 2) AS repeat,
		       COUNT(*) FILTER (WHERE done = 0 AND archived = 0 AND open > 0) AS "openOnly"
		FROM c`)
	return &s, err
}

// InquiryHistory للزبون: كم مرة استفسر وما نفّذ (لتنبيه الإداري وقت الحجز).
func (r *MatrixBusinessRepository) InquiryHistory(customerID string) (archived, done int) {
	_ = r.db.Get(&archived, `SELECT COUNT(*) FROM "Booking" WHERE "customerId" = $1 AND "archivedAt" IS NOT NULL`, customerID)
	_ = r.db.Get(&done, `SELECT COUNT(*) FROM "Booking" WHERE "customerId" = $1 AND status = 'COMPLETED' AND "archivedAt" IS NULL`, customerID)
	return
}
