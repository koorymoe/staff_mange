package repository

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"

	"staffmange-api/internal/model"
)

// ═══ ماتركس — فحوصات دورية (زبون قرب يزعل، الورق المتأخر، الحضور مقابل
// الشغل) + التسعير الشاذ ═══
//
// 🔴 كل الي يطلع من هنا معرّفات وأكواد وأرقام — بلا أسماء ولا هواتف.

// bookingLeaderExpr: ليدر الحجز — أول تكليف لموظف isLeader، وإلا ليدر
// آخر مهمة. NULL لو ما ينعرف (فما ننسب لأحد).
const bookingLeaderExpr = `COALESCE((
	SELECT ba."employeeId" FROM "BookingAssignment" ba
	JOIN "Employee" e ON e.id = ba."employeeId"
	WHERE ba."bookingId" = b.id AND e."isLeader" = true
	ORDER BY ba."createdAt" LIMIT 1), (
	SELECT m."leaderId" FROM "Mission" m WHERE m."bookingId" = b.id
	ORDER BY m."createdAt" DESC LIMIT 1))`

// ── زبون قرب يزعل ──

// CustomerRiskRow أرقام خام لزبون عنده حجز فعّال أو منجز بآخر ٣٠ يوم.
type CustomerRiskRow struct {
	CustomerID        string         `db:"customerId"`
	CustomerCode      int            `db:"customerCode"`
	LatestBookingCode string         `db:"latestBookingCode"`
	LeaderID          sql.NullString `db:"leaderId"`
	MaxPostpone       int            `db:"maxPostpone"`
	PostponeCode      string         `db:"postponeCode"`
	OverdueDays       int            `db:"overdueDays"`
	OverdueCode       string         `db:"overdueCode"`
	OpenComplaints    int            `db:"openComplaints"`
	LowRatings        int            `db:"lowRatings"`
	MinRating         int            `db:"minRating"`
}

const customerRiskSQL = `
	WITH cand AS (
		SELECT b.* FROM "Booking" b
		WHERE b."archivedAt" IS NULL AND b.code NOT LIKE 'OLD-%'
		  AND (b.status NOT IN ('COMPLETED','CANCELLED')
		       OR (b.status = 'COMPLETED' AND b."completedAt" >= now() - interval '30 days'))
	),
	latest AS (
		SELECT DISTINCT ON (b."customerId") b."customerId", b.code, ` + bookingLeaderExpr + ` AS "leaderId"
		FROM cand b
		ORDER BY b."customerId", COALESCE(b."scheduledAt", b."createdAt") DESC
	),
	pp AS (
		SELECT DISTINCT ON ("customerId") "customerId", "postponeCount", code
		FROM cand ORDER BY "customerId", "postponeCount" DESC
	),
	od AS (
		SELECT DISTINCT ON ("customerId") "customerId", code,
		       FLOOR(EXTRACT(EPOCH FROM (now() AT TIME ZONE 'UTC' - "scheduledAt")) / 86400)::int AS days
		FROM cand
		WHERE status NOT IN ('COMPLETED','CANCELLED') AND "scheduledAt" IS NOT NULL
		  AND "scheduledAt" < (now() AT TIME ZONE 'UTC') - interval '1 day'
		ORDER BY "customerId", "scheduledAt"
	),
	cp AS (
		SELECT c."customerId",
		       COUNT(*) FILTER (WHERE c.status IN ('NEW','IN_PROGRESS')
		                         AND c."createdAt" < (now() AT TIME ZONE 'UTC') - interval '3 days') AS open,
		       COUNT(*) FILTER (WHERE c."customerRating" <= 2
		                         AND c."createdAt" >= (now() AT TIME ZONE 'UTC') - interval '60 days') AS low,
		       MIN(c."customerRating") FILTER (WHERE c."customerRating" <= 2
		                         AND c."createdAt" >= (now() AT TIME ZONE 'UTC') - interval '60 days') AS minr
		FROM "Complaint" c WHERE c."customerId" IN (SELECT "customerId" FROM latest)
		GROUP BY c."customerId"
	)
	SELECT l."customerId" AS "customerId", cu."customerCode" AS "customerCode",
	       l.code AS "latestBookingCode", l."leaderId" AS "leaderId",
	       COALESCE(pp."postponeCount", 0) AS "maxPostpone", COALESCE(pp.code, '') AS "postponeCode",
	       COALESCE(od.days, 0) AS "overdueDays", COALESCE(od.code, '') AS "overdueCode",
	       COALESCE(cp.open, 0)::int AS "openComplaints", COALESCE(cp.low, 0)::int AS "lowRatings",
	       COALESCE(cp.minr, 0)::int AS "minRating"
	FROM latest l
	JOIN "Customer" cu ON cu.id = l."customerId"
	LEFT JOIN pp ON pp."customerId" = l."customerId"
	LEFT JOIN od ON od."customerId" = l."customerId"
	LEFT JOIN cp ON cp."customerId" = l."customerId"`

// CustomerRiskRows كل الزبائن المرشّحين — العوامل تنحسب بالخدمة.
func (r *AiRepository) CustomerRiskRows() ([]CustomerRiskRow, error) {
	rows := []CustomerRiskRow{}
	err := r.db.Select(&rows, customerRiskSQL)
	return rows, err
}

// CustomerRiskRowFor نفس الشي لزبون واحد (لجمع الأدلة) — nil لو ما بقى مرشّح.
func (r *AiRepository) CustomerRiskRowFor(customerID string) (*CustomerRiskRow, error) {
	var row CustomerRiskRow
	err := r.db.Get(&row, `SELECT * FROM (`+customerRiskSQL+`) x WHERE x."customerId" = $1`, customerID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// ── تسعير شاذ ──

// PriceOutlierFacts صافي الفاتورة ووسيط نفس الخدمة.
type PriceOutlierFacts struct {
	InvoiceID      string  `db:"invoiceId"`
	AccountingCode string  `db:"accountingCode"`
	BookingCode    string  `db:"bookingCode"`
	ServiceID      string  `db:"serviceId"`
	NetTotal       float64 `db:"netTotal"`
	IsFree         bool    `db:"isFree"`
	Median         float64 `db:"median"`
	Samples        int     `db:"samples"`
}

// PriceOutlierBaseline — الوسيط من صافي فواتير الليدر لنفس خدمة الحجز
// (آخر ١٨٠ يوم لحد وقت هذي الفاتورة، بلا المسودات والملغاة والمجانية
// والصفرية، وبلا هذي الفاتورة). nil لو الفاتورة بلا حجز أو خدمة.
//
// ⚠️ ليش مو ServicePriceSample: العيّنات تنسجّل من الفواتير اليدوية بس
// (جزء صغير)، والوسيط هنا من كل الفواتير المعتمدة/المرسلة أوسع وأصدق.
func (r *AiRepository) PriceOutlierBaseline(invoiceID string) (*PriceOutlierFacts, error) {
	var f PriceOutlierFacts
	err := r.db.Get(&f, `
		WITH cur AS (
			SELECT li.id, li."createdAt", COALESCE(li."accountingCode", '') AS "accountingCode",
			       b.code, b."serviceId", li."netTotal"::float8 AS net, li."isFree"
			FROM "LeaderInvoice" li JOIN "Booking" b ON b.id = li."bookingId"
			WHERE li.id = $1 AND b."serviceId" IS NOT NULL
		),
		hist AS (
			SELECT li."netTotal"::float8 AS net
			FROM "LeaderInvoice" li JOIN "Booking" b ON b.id = li."bookingId"
			CROSS JOIN cur
			WHERE b."serviceId" = cur."serviceId" AND li.id <> cur.id
			  AND li.status <> 'DRAFT' AND li."revokedAt" IS NULL
			  AND NOT li."isFree" AND li."netTotal" > 0
			  AND li."createdAt" >= cur."createdAt" - interval '180 days'
			  AND li."createdAt" <= cur."createdAt"
		)
		SELECT cur.id AS "invoiceId", cur."accountingCode" AS "accountingCode", cur.code AS "bookingCode",
		       cur."serviceId" AS "serviceId", cur.net AS "netTotal", cur."isFree" AS "isFree",
		       COALESCE((SELECT percentile_cont(0.5) WITHIN GROUP (ORDER BY net) FROM hist), 0)::float8 AS median,
		       (SELECT COUNT(*) FROM hist)::int AS samples
		FROM cur`, invoiceID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

// ── الورق المتأخر ──

// LatePaperworkRow حجز منجز ناقصه ورق بعد ٤٨ ساعة، مع ليدره.
type LatePaperworkRow struct {
	LeaderID string `db:"leaderId"`
	model.LatePaperworkItem
}

// LatePaperworkRows حجوزات آخر ٣٠ يوم المنجزة من أكثر من ٤٨ ساعة وناقصها
// فاتورة أو تقرير — نفس شروط محطات «ناقصها ورق» (بلا الكشف والتاريخي
// والمسوّى إدارياً والمطلوب حذفه). leaderID فارغ = كل الليدرية.
func (r *AiRepository) LatePaperworkRows(leaderID string) ([]LatePaperworkRow, error) {
	rows := []LatePaperworkRow{}
	err := r.db.Select(&rows, `
		SELECT * FROM (
			SELECT `+bookingLeaderExpr+` AS "leaderId", b.code AS "bookingCode", b.id AS "bookingId",
			       NOT `+hasInvoiceSQL+` AS "missingInvoice",
			       NOT `+hasReportSQL+` AS "missingReport",
			       FLOOR(EXTRACT(EPOCH FROM ((now() AT TIME ZONE 'UTC') - b."completedAt")) / 3600)::int AS "hoursSinceDone"
			FROM "Booking" b
			WHERE b.status = 'COMPLETED' AND b."archivedAt" IS NULL
			  AND b."completedAt" >= (now() AT TIME ZONE 'UTC') - interval '30 days'
			  AND b."completedAt" < (now() AT TIME ZONE 'UTC') - interval '48 hours'
			  AND NOT `+isSurveySQL+` AND NOT (`+isLegacyImportSQL+`)
			  AND b."settledLegacyAt" IS NULL`+NotDeletePendingSQL("b")+`
			  AND (NOT `+hasInvoiceSQL+` OR NOT `+hasReportSQL+`)
		) x
		WHERE "leaderId" IS NOT NULL AND ($1 = '' OR "leaderId" = $1)
		ORDER BY "leaderId", "hoursSinceDone" DESC`, leaderID)
	return rows, err
}

// ── الحضور مقابل الشغل ──

// AttendanceGapRow موظف بيوم معيّن: دقائق حضوره المغلقة وأكواد شغله.
type AttendanceGapRow struct {
	EmployeeID      string         `db:"employeeId"`
	Kind            string         `db:"kind"`
	AttendedMinutes int            `db:"attendedMinutes"`
	BookingCodes    pq.StringArray `db:"bookingCodes"`
}

// workOnDaySQL: حجوزات الموظف (تكليف أو مهمة ليدر/عضو) الي انجدولت أو
// بدت أو خلصت بيوم بغداد $1. يحتاج e.id بالنطاق.
const workOnDaySQL = `
	SELECT b.code FROM "Booking" b
	WHERE (baghdad_date(b."scheduledAt") = $1::date OR baghdad_date(b."startedAt") = $1::date
	       OR baghdad_date(b."completedAt") = $1::date)
	  AND (EXISTS (SELECT 1 FROM "BookingAssignment" ba WHERE ba."bookingId" = b.id AND ba."employeeId" = e.id)
	    OR EXISTS (SELECT 1 FROM "Mission" m WHERE m."bookingId" = b.id
	               AND (m."leaderId" = e.id OR e.id = ANY(m."memberIds"))))`

// AttendanceWorkGaps ليوم بغداد day (YYYY-MM-DD). employeeID فارغ = الكل.
//
//	PRESENT_NO_WORK: فني/ليدر حاضر ≥minMinutes (جلسات مغلقة بس) وبلا أي
//	  حجز مجدول/بادي/منجز باسمه بذاك اليوم.
//	WORK_NO_ATTENDANCE: حجز باسمه بدا أو خلص بذاك اليوم وماكو ولا جلسة حضور.
func (r *AiRepository) AttendanceWorkGaps(day string, minMinutes int, employeeID string) ([]AttendanceGapRow, error) {
	rows := []AttendanceGapRow{}
	err := r.db.Select(&rows, `
		SELECT * FROM (
			SELECT e.id AS "employeeId", 'PRESENT_NO_WORK' AS kind,
			       SUM(EXTRACT(EPOCH FROM (a."checkOut" - a."checkIn")) / 60)::int AS "attendedMinutes",
			       '{}'::text[] AS "bookingCodes"
			FROM "Attendance" a JOIN "Employee" e ON e.id = a."employeeId"
			WHERE a.date = $1::date AND a."checkOut" IS NOT NULL
			  AND e.status = 'ACTIVE' AND (e.role = 'TECHNICIAN' OR e."isLeader")
			  AND NOT EXISTS (`+workOnDaySQL+`)
			GROUP BY e.id
			HAVING SUM(EXTRACT(EPOCH FROM (a."checkOut" - a."checkIn")) / 60) >= $2
			UNION ALL
			SELECT e.id, 'WORK_NO_ATTENDANCE', 0,
			       ARRAY(SELECT b.code FROM "Booking" b
			             WHERE (baghdad_date(b."startedAt") = $1::date OR baghdad_date(b."completedAt") = $1::date)
			               AND (EXISTS (SELECT 1 FROM "BookingAssignment" ba WHERE ba."bookingId" = b.id AND ba."employeeId" = e.id)
			                 OR EXISTS (SELECT 1 FROM "Mission" m WHERE m."bookingId" = b.id
			                            AND (m."leaderId" = e.id OR e.id = ANY(m."memberIds"))))
			             ORDER BY b.code)
			FROM "Employee" e
			WHERE e.status = 'ACTIVE'
			  AND NOT EXISTS (SELECT 1 FROM "Attendance" a WHERE a."employeeId" = e.id AND a.date = $1::date)
		) x
		WHERE (kind = 'PRESENT_NO_WORK' OR cardinality("bookingCodes") > 0)
		  AND ($3 = '' OR "employeeId" = $3)
		ORDER BY "employeeId"`, day, minMinutes, employeeID)
	return rows, err
}
