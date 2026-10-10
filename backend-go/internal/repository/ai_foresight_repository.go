package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"

	"staffmange-api/internal/model"
)

// ═══ ماتركس — التنبؤ بالتأخير + مطابقة الفاتورة + فرص البيع ═══
//
// كلها قراءات من بيانات منظّمة موجودة (أعمدة وأرقام)، بلا تحليل نص
// حر، وبلا أي نموذج خارجي.

// bookingLeaderLateral الليدر (isLeader) الأول بكادر الحجز b.
const bookingLeaderLateral = `LEFT JOIN LATERAL (
	SELECT ba."employeeId" AS eid FROM "BookingAssignment" ba
	JOIN "Employee" e ON e.id = ba."employeeId"
	WHERE ba."bookingId" = b.id AND e."isLeader" = true
	ORDER BY ba."createdAt" LIMIT 1) ldr ON true`

// DelayCandidate حجز مجدول لسه ما بدا.
type DelayCandidate struct {
	ID          string    `db:"id"`
	Code        string    `db:"code"`
	ScheduledAt time.Time `db:"scheduledAt"`
	ServiceID   string    `db:"serviceId"`
	LeaderID    string    `db:"leaderId"`
}

// UpcomingDelayCandidates حجوزات معلّقة/مثبّتة بموعد (يوم بغداد) بين
// from و to، ما بدا تنفيذها، ولها خدمة.
func (r *AiRepository) UpcomingDelayCandidates(from, to string) ([]DelayCandidate, error) {
	rows := []DelayCandidate{}
	err := r.db.Select(&rows, `
		SELECT b.id, b.code, b."scheduledAt", b."serviceId", COALESCE(ldr.eid, '') AS "leaderId"
		FROM "Booking" b
		`+bookingLeaderLateral+`
		WHERE b.status IN ('PENDING', 'CONFIRMED')
		  AND b."archivedAt" IS NULL AND b."startedAt" IS NULL
		  AND b."scheduledAt" IS NOT NULL AND b."serviceId" IS NOT NULL
		  AND baghdad_date(b."scheduledAt") BETWEEN $1::date AND $2::date
		ORDER BY b."scheduledAt"`, from, to)
	return rows, err
}

// DurationStat وسيط مدة التنفيذ الفعلية (بدء → إنجاز) بالدقائق.
// LeaderID فارغ = إحصائية الخدمة كلها.
type DurationStat struct {
	ServiceID     string  `db:"serviceId"`
	LeaderID      string  `db:"leaderId"`
	MedianMinutes float64 `db:"medianMinutes"`
	Samples       int     `db:"samples"`
}

// CompletedDurationStats وسيط (startedAt → completedAt) لحجوزات مكتملة
// بآخر ٣٦٥ يوم لهذي الخدمات — مرة لكل خدمة ومرة لكل (خدمة، ليدر).
// نستبعد: التسويات الإدارية القديمة، الإنجاز الجزئي (يوم ثاني يخربط
// المدة)، والمدد غير المعقولة (أقل من ١٥ دقيقة أو أكثر من ١٦ ساعة).
func (r *AiRepository) CompletedDurationStats(serviceIDs []string) ([]DurationStat, error) {
	rows := []DurationStat{}
	if len(serviceIDs) == 0 {
		return rows, nil
	}
	err := r.db.Select(&rows, `
		WITH d AS (
			SELECT b."serviceId" AS sid, COALESCE(ldr.eid, '') AS lid,
			       EXTRACT(EPOCH FROM (b."completedAt" - b."startedAt")) / 60.0 AS m
			FROM "Booking" b
			`+bookingLeaderLateral+`
			WHERE b.status = 'COMPLETED'
			  AND b."startedAt" IS NOT NULL AND b."completedAt" IS NOT NULL
			  AND b."settledLegacyAt" IS NULL AND b."partialCount" = 0
			  AND b."completedAt" >= now() - interval '365 days'
			  AND b."serviceId" = ANY($1)
		)
		SELECT sid AS "serviceId",
		       CASE WHEN GROUPING(lid) = 1 THEN '' ELSE lid END AS "leaderId",
		       percentile_cont(0.5) WITHIN GROUP (ORDER BY m)::float8 AS "medianMinutes",
		       COUNT(*)::int AS samples
		FROM d
		WHERE m BETWEEN 15 AND 960
		GROUP BY GROUPING SETS ((sid), (sid, lid))`, pq.Array(serviceIDs))
	return rows, err
}

// ═══ مطابقة الفاتورة مع الشغل ═══

// InvoiceWorkFacts الأرقام المنظّمة لفاتورة ليدر وحجزها.
type InvoiceWorkFacts struct {
	InvoiceID          string   `db:"invoiceId"`
	AccountingCode     string   `db:"accountingCode"`
	BookingCode        string   `db:"bookingCode"`
	InvoiceDeviceCount int      `db:"invoiceDeviceCount"`
	BookedDeviceCount  *int     `db:"bookedDeviceCount"`
	InvoiceNetTotal    float64  `db:"invoiceNetTotal"`
	QuotedPrice        *float64 `db:"quotedPrice"`
	IsFree             bool     `db:"isFree"`
}

// InvoiceWorkFacts يرجّع nil لو الفاتورة مو مربوطة بحجز.
func (r *AiRepository) InvoiceWorkFacts(invoiceID string) (*InvoiceWorkFacts, error) {
	var f InvoiceWorkFacts
	err := r.db.Get(&f, `
		SELECT li.id AS "invoiceId", li."accountingCode" AS "accountingCode", b.code AS "bookingCode",
		       li."totalDeviceCount" AS "invoiceDeviceCount", b."deviceCount" AS "bookedDeviceCount",
		       li."netTotal"::float8 AS "invoiceNetTotal", b."quotedPrice" AS "quotedPrice", li."isFree" AS "isFree"
		FROM "LeaderInvoice" li JOIN "Booking" b ON b.id = li."bookingId"
		WHERE li.id = $1`, invoiceID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

// ═══ فرص البيع ═══

// CustomerFamilyHistory آخر حجز مكتمل لزبون بعائلة خدمات معيّنة.
type CustomerFamilyHistory struct {
	CustomerID     string    `db:"customerId"`
	CustomerCode   int       `db:"customerCode"`
	CustomerName   string    `db:"customerName"`
	CustomerPhone  string    `db:"customerPhone"`
	ServiceID      string    `db:"serviceId"`
	ServiceName    string    `db:"serviceName"`
	ServiceKind    string    `db:"serviceKind"`
	LastCompleted  time.Time `db:"lastCompleted"`
	LaterBookingOf bool      `db:"hasLaterBooking"`
}

// CustomerServiceHistory لكل (زبون، خدمة) آخر إنجاز، وهل عنده حجز
// (غير ملغى) لنفس الخدمة انسوى بعده. الخدمات من Booking.serviceId و
// BookingService كلتيهن. بلا الحجوزات الداخلية والمؤرشفة.
func (r *AiRepository) CustomerServiceHistory() ([]CustomerFamilyHistory, error) {
	rows := []CustomerFamilyHistory{}
	err := r.db.Select(&rows, `
		WITH bs AS (
			SELECT b.id AS bid, b."customerId" AS cid, b.status, b."createdAt",
			       COALESCE(b."completedAt", b."createdAt") AS done_at, s.sid
			FROM "Booking" b
			CROSS JOIN LATERAL (
				SELECT b."serviceId" AS sid WHERE b."serviceId" IS NOT NULL
				UNION SELECT x."serviceId" FROM "BookingService" x WHERE x."bookingId" = b.id
			) s
			WHERE b."archivedAt" IS NULL AND b."bookingType"::text <> 'INTERNAL'
			  AND b.status::text <> 'CANCELLED'
		),
		last_done AS (
			SELECT cid, sid, MAX(done_at) AS last_completed
			FROM bs WHERE status = 'COMPLETED' GROUP BY cid, sid
		)
		SELECT cu.id AS "customerId", cu."customerCode" AS "customerCode",
		       cu.name AS "customerName", COALESCE(cu.phone, '') AS "customerPhone",
		       sv.id AS "serviceId", sv.name AS "serviceName", COALESCE(sv."serviceKind", '') AS "serviceKind",
		       ld.last_completed AS "lastCompleted",
		       EXISTS (SELECT 1 FROM bs x WHERE x.cid = ld.cid AND x.sid = ld.sid
		               AND x."createdAt" > ld.last_completed) AS "hasLaterBooking"
		FROM last_done ld
		JOIN "Customer" cu ON cu.id = ld.cid
		JOIN "Service" sv ON sv.id = ld.sid`)
	return rows, err
}

// CustomerBookedServiceIDs كل الخدمات الي الزبون حجزها بأي حالة غير
// ملغاة — حتى ما نقترح خدمة عنده حجز بيها أصلاً.
func (r *AiRepository) CustomerBookedServiceIDs() (map[string]map[string]bool, error) {
	var rows []struct {
		CID string `db:"cid"`
		SID string `db:"sid"`
	}
	err := r.db.Select(&rows, `
		SELECT DISTINCT b."customerId" AS cid, s.sid
		FROM "Booking" b
		CROSS JOIN LATERAL (
			SELECT b."serviceId" AS sid WHERE b."serviceId" IS NOT NULL
			UNION SELECT x."serviceId" FROM "BookingService" x WHERE x."bookingId" = b.id
		) s
		WHERE b.status::text <> 'CANCELLED'`)
	if err != nil {
		return nil, err
	}
	out := map[string]map[string]bool{}
	for _, r := range rows {
		if out[r.CID] == nil {
			out[r.CID] = map[string]bool{}
		}
		out[r.CID][r.SID] = true
	}
	return out, nil
}

// ActiveServices الخدمات (معرّف، اسم، نوع) — لمطابقة العوائل.
func (r *AiRepository) ActiveServices() ([]model.Service, error) {
	rows := []model.Service{}
	err := r.db.Select(&rows, `SELECT * FROM "Service" ORDER BY name`)
	return rows, err
}
