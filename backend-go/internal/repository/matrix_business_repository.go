package repository

import (
	"database/sql"
	"errors"
	"sort"
	"time"

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
	// ProjectRevenue دفعات المشاريع بهالشهر (داخلة بـRevenue).
	ProjectRevenue float64 `db:"-" json:"projectRevenue"`
}

// completedCountable حجز منجز ينحسب بالإيراد — نفس تعريف RevenueBookingSQL.
var completedCountable = RevenueBookingSQL("b") + ` AND b."completedAt" IS NOT NULL`

// projectByMonth دفعات المشاريع لكل شهر (قرار (ع) 10-06: فلوس المشاريع ما چانت تنحسب).
func (r *MatrixBusinessRepository) projectByMonth(n int) (map[string]float64, error) {
	type row struct {
		Month string  `db:"month"`
		Sum   float64 `db:"sum"`
	}
	rows := []row{}
	err := r.db.Select(&rows, `SELECT to_char(date_trunc('month', "paidAt"), 'YYYY-MM') AS month, SUM(amount)::float8 AS sum
		FROM "ProjectPayment" WHERE "cancelledAt" IS NULL
		  AND "paidAt" >= (date_trunc('month', baghdad_today()) - make_interval(months => $1 - 1))::date
		  AND "paidAt" <= baghdad_today()
		GROUP BY 1`, n)
	out := map[string]float64{}
	for _, x := range rows {
		out[x.Month] = x.Sum
	}
	return out, err
}

// Monthly آخر n أشهر: حجوزات منجزة، وصافي فواتيرها (آخر فاتورة لكل حجز).
func (r *MatrixBusinessRepository) Monthly(n int) ([]MonthRevenue, error) {
	rows := []MonthRevenue{}
	err := r.db.Select(&rows, `
		SELECT to_char(date_trunc('month', baghdad_date(b."completedAt")), 'YYYY-MM') AS month,
		       COUNT(*) AS bookings,
		       COALESCE(SUM(`+RevenueAmountSQL("b")+`), 0) AS revenue,
		       COUNT(li.id) AS invoiced
		FROM "Booking" b
		LEFT JOIN LATERAL (SELECT id FROM "LeaderInvoice" WHERE "bookingId" = b.id AND "revokedAt" IS NULL ORDER BY "createdAt" DESC LIMIT 1) li ON true
		WHERE `+completedCountable+`
		  AND baghdad_date(b."completedAt") >= (date_trunc('month', baghdad_today()) - make_interval(months => $1 - 1))::date
		GROUP BY 1 ORDER BY 1`, n)
	if err != nil {
		return rows, err
	}
	proj, err := r.projectByMonth(n)
	if err != nil {
		return rows, err
	}
	seen := map[string]bool{}
	for i := range rows {
		rows[i].ProjectRevenue = proj[rows[i].Month]
		rows[i].Revenue += proj[rows[i].Month]
		seen[rows[i].Month] = true
	}
	for m, v := range proj {
		if !seen[m] {
			rows = append(rows, MonthRevenue{Month: m, Revenue: v, ProjectRevenue: v})
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Month < rows[j].Month })
	return rows, nil
}

type MonthToDate struct {
	Bookings     int     `db:"bookings" json:"bookings"`
	Revenue      float64 `db:"revenue" json:"revenue"`
	LastBookings int     `db:"lastBookings" json:"lastBookings"`
	LastRevenue  float64 `db:"lastRevenue" json:"lastRevenue"`
	// دفعات المشاريع (داخلة بـRevenue/LastRevenue).
	ProjectRevenue     float64 `db:"projectRevenue" json:"projectRevenue"`
	LastProjectRevenue float64 `db:"lastProjectRevenue" json:"lastProjectRevenue"`
}

// MTD هالشهر لحد اليوم مقابل نفس الفترة الشهر الماضي.
func (r *MatrixBusinessRepository) MTD() (*MonthToDate, error) {
	var m MonthToDate
	err := r.db.Get(&m, `
		WITH x AS (
		  SELECT baghdad_date(b."completedAt") AS d, `+RevenueAmountSQL("b")+` AS net
		  FROM "Booking" b
		  WHERE `+completedCountable+`)
		SELECT
		  COUNT(*) FILTER (WHERE d >= date_trunc('month', baghdad_today())::date) AS bookings,
		  COALESCE(SUM(net) FILTER (WHERE d >= date_trunc('month', baghdad_today())::date),0) AS revenue,
		  COUNT(*) FILTER (WHERE d >= (date_trunc('month', baghdad_today()) - interval '1 month')::date
		                     AND d <= (baghdad_today() - interval '1 month')::date) AS "lastBookings",
		  COALESCE(SUM(net) FILTER (WHERE d >= (date_trunc('month', baghdad_today()) - interval '1 month')::date
		                     AND d <= (baghdad_today() - interval '1 month')::date),0) AS "lastRevenue"
		FROM x`)
	if err != nil {
		return &m, err
	}
	err = r.db.Get(&m, `SELECT
		  COALESCE(SUM(amount) FILTER (WHERE "paidAt" >= date_trunc('month', baghdad_today())::date AND "paidAt" <= baghdad_today()), 0)::float8 AS "projectRevenue",
		  COALESCE(SUM(amount) FILTER (WHERE "paidAt" >= (date_trunc('month', baghdad_today()) - interval '1 month')::date
		                     AND "paidAt" <= (baghdad_today() - interval '1 month')::date), 0)::float8 AS "lastProjectRevenue"
		FROM "ProjectPayment" WHERE "cancelledAt" IS NULL`)
	if err != nil {
		return &m, err
	}
	m.Revenue += m.ProjectRevenue
	m.LastRevenue += m.LastProjectRevenue
	return &m, err
}

// ═══ أساس التوقع (قرار (ع) 10-07: «التوقع قريب من النتائج النهائية») ═══
// قبل: المثبّت × نسبة الإنجاز × معدل **فاتورة الليدر** — والإيراد نفسه **فلوس
// مقبوضة**، فالتوقع چان يقيس شي ثاني. هسه كل شي بنفس عملة الإيراد:

// Pipeline الحجوزات المفتوحة الي موعدها هالشهر (حتى المتأخرة عن موعدها)، كل
// وحدة بقيمتها المتوقعة = نسبة إنجاز خدمتها × معدل المقبوض لخدمتها (آخر ٩٠
// يوم؛ الخدمة الي عيّناتها قليلة تاخذ معدل الشركة).
type Pipeline struct {
	Jobs     int             `db:"jobs"`
	ExpJobs  sql.NullFloat64 `db:"expJobs"`
	Value    sql.NullFloat64 `db:"value"`
	Samples  int             `db:"samples"`
	AvgValue sql.NullFloat64 `db:"avgValue"`
	Rate     sql.NullFloat64 `db:"rate"`
}

func (r *MatrixBusinessRepository) Pipeline() (*Pipeline, error) {
	var p Pipeline
	err := r.db.Get(&p, `
		WITH hist AS (
		  SELECT b."serviceId" AS sid, `+RevenueAmountSQL("b")+` AS v
		  FROM "Booking" b WHERE `+completedCountable+`
		    AND b."completedAt" > (now() AT TIME ZONE 'UTC') - interval '90 days' AND `+RevenueAmountSQL("b")+` > 0
		), hs AS (SELECT sid, AVG(v) AS av, COUNT(*) AS n FROM hist GROUP BY sid),
		hg AS (SELECT AVG(v) AS av, COUNT(*) AS n FROM hist),
		conf AS (
		  SELECT "serviceId" AS sid, (status::text IN ('COMPLETED', 'PARTIAL')) AS ok FROM "Booking"
		  WHERE "confirmedAt" > (now() AT TIME ZONE 'UTC') - interval '90 days'
		    AND "confirmedAt" < (now() AT TIME ZONE 'UTC') - interval '7 days'
		    AND "bookingType" IS DISTINCT FROM 'INTERNAL'
		), rs AS (SELECT sid, AVG(ok::int)::float8 AS r, COUNT(*) AS n FROM conf GROUP BY sid),
		rg AS (SELECT AVG(ok::int)::float8 AS r FROM conf),
		pend AS (
		  SELECT b."serviceId" AS sid FROM "Booking" b
		  WHERE b.status::text IN ('CONFIRMED', 'IN_PROGRESS', 'PENDING') AND b."scheduledAt" IS NOT NULL
		    AND b."bookingType" IS DISTINCT FROM 'INTERNAL'`+BookingCountableAndSQL("b")+`
		    AND baghdad_date(b."scheduledAt") >= date_trunc('month', baghdad_today())::date
		    AND baghdad_date(b."scheduledAt") <= (date_trunc('month', baghdad_today()) + interval '1 month - 1 day')::date
		)
		SELECT COUNT(pend.*)::int AS jobs,
		       SUM(COALESCE(CASE WHEN rs.n >= 8 THEN rs.r END, rg.r)) AS "expJobs",
		       SUM(COALESCE(CASE WHEN rs.n >= 8 THEN rs.r END, rg.r) * COALESCE(CASE WHEN hs.n >= 5 THEN hs.av END, hg.av)) AS value,
		       COALESCE(MAX(hg.n), (SELECT n FROM hg))::int AS samples,
		       COALESCE(MAX(hg.av), (SELECT av FROM hg)) AS "avgValue",
		       COALESCE(MAX(rg.r), (SELECT r FROM rg)) AS rate
		FROM hg CROSS JOIN rg LEFT JOIN pend ON true
		LEFT JOIN hs ON hs.sid = pend.sid LEFT JOIN rs ON rs.sid = pend.sid`)
	return &p, err
}

// PaceMonth شهر سابق: إيراده الكلي، وشكد چان واصل لحد نفس يوم هالشهر.
type PaceMonth struct {
	Month string  `db:"month"`
	Total float64 `db:"total"`
	Upto  float64 `db:"upto"`
}

// Pace آخر ٣ أشهر كاملة — منحنى «شكد يوصل الإيراد لحد اليوم N من الشهر».
func (r *MatrixBusinessRepository) Pace() ([]PaceMonth, error) {
	rows := []PaceMonth{}
	err := r.db.Select(&rows, `
		WITH ev AS (
		  SELECT baghdad_date(b."completedAt") AS d, `+RevenueAmountSQL("b")+` AS v FROM "Booking" b WHERE `+completedCountable+`
		  UNION ALL
		  SELECT "paidAt", amount FROM "ProjectPayment" WHERE "cancelledAt" IS NULL
		), m AS (
		  SELECT k, (date_trunc('month', baghdad_today()) - make_interval(months => k))::date AS ms FROM generate_series(1, 3) k
		)
		SELECT to_char(m.ms, 'YYYY-MM') AS month,
		       COALESCE(SUM(ev.v) FILTER (WHERE ev.d >= m.ms AND ev.d < (m.ms + interval '1 month')::date), 0)::float8 AS total,
		       COALESCE(SUM(ev.v) FILTER (WHERE ev.d >= m.ms AND ev.d < LEAST((m.ms + interval '1 month')::date,
		                 m.ms + extract(day FROM baghdad_today())::int)), 0)::float8 AS upto
		FROM m LEFT JOIN ev ON ev.d >= m.ms AND ev.d < (m.ms + interval '1 month')::date
		GROUP BY m.ms ORDER BY m.ms DESC`)
	return rows, err
}

// LogForecast يحفظ توقع اليوم (مرة باليوم — آخر قيمة تغلب).
func (r *MatrixBusinessRepository) LogForecast(expected, pipeline float64, pace *float64, mtd float64) error {
	_, err := r.db.Exec(`INSERT INTO "MatrixForecastLog" (day, month, expected, pipeline, pace, mtd)
		VALUES (baghdad_today(), to_char(baghdad_today(), 'YYYY-MM'), $1, $2, $3, $4)
		ON CONFLICT (day) DO UPDATE SET expected = EXCLUDED.expected, pipeline = EXCLUDED.pipeline,
		  pace = EXCLUDED.pace, mtd = EXCLUDED.mtd`, expected, pipeline, pace, mtd)
	return err
}

// ForecastCheck دقة توقع الشهر الماضي: التوقع بنص الشهر مقابل النتيجة الفعلية.
type ForecastCheck struct {
	Month    string          `db:"month" json:"month"`
	Day      string          `db:"day" json:"day"`
	Expected float64         `db:"expected" json:"expected"`
	Actual   sql.NullFloat64 `db:"-" json:"-"`
}

func (r *MatrixBusinessRepository) LastMonthForecast() (*ForecastCheck, error) {
	var c ForecastCheck
	err := r.db.Get(&c, `SELECT month, to_char(day, 'YYYY-MM-DD') AS day, expected FROM "MatrixForecastLog"
		WHERE month = to_char(baghdad_today() - interval '1 month', 'YYYY-MM')
		ORDER BY abs(extract(day FROM day) - 15), day LIMIT 1`)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &c, err
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

// UninvoicedItem حجز منجز هالشهر وما عليه فاتورة — إيراد ما انحسب بعد.
type UninvoicedItem struct {
	ID          string    `db:"id" json:"id"`
	Code        string    `db:"code" json:"code"`
	Customer    *string   `db:"customer" json:"customer"`
	Service     *string   `db:"service" json:"service"`
	CompletedAt time.Time `db:"completedAt" json:"completedAt"`
	Leader      *string   `db:"leader" json:"leader"`
}

// Uninvoiced نفس استثناءات «ناقصها ورق»: القديم (OLD) والكشف ما يحتاجون فاتورة.
func (r *MatrixBusinessRepository) Uninvoiced(limit int) ([]UninvoicedItem, error) {
	rows := []UninvoicedItem{}
	err := r.db.Select(&rows, `
		SELECT b.id, b.code, c.name AS customer, s.name AS service, b."completedAt",
		       (SELECT e.name FROM "Mission" m JOIN "Employee" e ON e.id = m."leaderId" WHERE m."bookingId" = b.id
		         ORDER BY m."assignedAt" DESC NULLS LAST LIMIT 1) AS leader
		FROM "Booking" b
		LEFT JOIN "Customer" c ON c.id = b."customerId"
		LEFT JOIN "Service" s ON s.id = b."serviceId"
		WHERE b.status = 'COMPLETED' AND b."completedAt" >= date_trunc('month', now() AT TIME ZONE 'Asia/Baghdad') AT TIME ZONE 'Asia/Baghdad'
		  AND NOT EXISTS (SELECT 1 FROM "LeaderInvoice" li WHERE li."bookingId" = b.id AND li."revokedAt" IS NULL)
		  AND upper(b.code) NOT LIKE 'OLD%' AND b."bookingType" IS DISTINCT FROM 'SURVEY'
		  AND b."archivedAt" IS NULL
		ORDER BY b."completedAt" DESC LIMIT $1`, limit)
	return rows, err
}
