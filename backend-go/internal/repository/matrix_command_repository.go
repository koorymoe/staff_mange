package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// ═══ مركز قيادة ماتركس — البث المباشر، والاتجاه اليومي، وتركيز التأخير ═══
// كله قراءة من الجداول الموجودة؛ الشي الوحيد الي ينكتب هو سجل «اسأل ماتركس».

type MatrixCommandRepository struct{ db *sqlx.DB }

func NewMatrixCommandRepository(db *sqlx.DB) *MatrixCommandRepository {
	return &MatrixCommandRepository{db: db}
}

type FeedRow struct {
	Kind  string    `db:"kind" json:"kind"` // BOOKING | INVOICE | ACTION | ESCALATED | PROPOSAL
	Title string    `db:"title" json:"title"`
	Sub   string    `db:"sub" json:"sub"`
	At    time.Time `db:"at" json:"at"`
	Ref   string    `db:"ref" json:"ref"`
}

// Feed آخر الأحداث المهمة (٣ أيام).
func (r *MatrixCommandRepository) Feed(limit int) ([]FeedRow, error) {
	rows := []FeedRow{}
	err := r.db.Select(&rows, `
		SELECT * FROM (
			SELECT 'BOOKING' AS kind, 'تم تسجيل حجز جديد — ' || COALESCE(c.name, 'زبون') AS title,
			       'رمز ' || b.code || COALESCE(' · ' || s.name, '') AS sub, b."createdAt" AS at, b.id AS ref
			FROM "Booking" b LEFT JOIN "Customer" c ON c.id = b."customerId" LEFT JOIN "Service" s ON s.id = b."serviceId"
			WHERE b."createdAt" >= now() - interval '3 days' AND b."archivedAt" IS NULL
			  AND b."bookingType" IS DISTINCT FROM 'INTERNAL'
			UNION ALL
			SELECT 'INVOICE', 'فاتورة جديدة — ' || COALESCE(li."customerName", ''),
			       'قيمة ' || to_char(COALESCE(li."netTotal", 0), 'FM999,999,999') || ' د.ع', li."createdAt", li.id
			FROM "LeaderInvoice" li WHERE li."createdAt" >= now() - interval '3 days'
			UNION ALL
			SELECT CASE WHEN a."escalatedAt" IS NOT NULL THEN 'ESCALATED' ELSE 'ACTION' END,
			       a.summary, COALESCE(a."targetLabel", ''), COALESCE(a."escalatedAt", a."createdAt"), a.id
			FROM "AiAction" a WHERE a."createdAt" >= now() - interval '3 days' AND a.status <> 'UNDONE'
			UNION ALL
			SELECT 'PROPOSAL', 'اقتراح جديد من ماتركس', p.title, p."createdAt", p.id
			FROM "MatrixProposal" p WHERE p.status = 'PENDING' AND p."createdAt" >= now() - interval '3 days'
		) x ORDER BY at DESC LIMIT $1`, limit)
	return rows, err
}

type TrendDay struct {
	Day       string  `db:"day" json:"day"`
	Completed int     `db:"completed" json:"completed"`
	Revenue   float64 `db:"revenue" json:"revenue"`
	OnTime    int     `db:"onTime" json:"onTime"` // انجز بنفس يوم موعده وبلا جزئي
	Scheduled int     `db:"scheduled" json:"scheduled"`
}

// DailyTrend آخر n يوم: المنجز، والإيراد، والمنجز بوقته من المجدول.
func (r *MatrixCommandRepository) DailyTrend(n int) ([]TrendDay, error) {
	rows := []TrendDay{}
	err := r.db.Select(&rows, `
		WITH days AS (SELECT generate_series(baghdad_today() - ($1 - 1), baghdad_today(), interval '1 day')::date AS d)
		SELECT to_char(d, 'MM-DD') AS day,
		  (SELECT COUNT(*) FROM "Booking" b WHERE b.status = 'COMPLETED' AND baghdad_date(b."completedAt") = d
		     AND b."archivedAt" IS NULL AND b."bookingType" IS DISTINCT FROM 'INTERNAL') AS completed,
		  (SELECT COALESCE(SUM(li."netTotal"), 0) FROM "LeaderInvoice" li JOIN "Booking" b ON b.id = li."bookingId"
		     WHERE baghdad_date(b."completedAt") = d AND b.status = 'COMPLETED') AS revenue,
		  (SELECT COUNT(*) FROM "Booking" b WHERE baghdad_date(b."scheduledAt") = d AND b.status = 'COMPLETED'
		     AND baghdad_date(b."completedAt") = d AND b."partialCount" = 0 AND b."archivedAt" IS NULL) AS "onTime",
		  (SELECT COUNT(*) FROM "Booking" b WHERE baghdad_date(b."scheduledAt") = d AND b.status::text <> 'CANCELLED'
		     AND b."archivedAt" IS NULL AND b."bookingType" IS DISTINCT FROM 'INTERNAL') AS scheduled
		FROM days ORDER BY d`, n)
	return rows, err
}

type LateWindow struct {
	Total     int `db:"total" json:"total"`
	Late      int `db:"late" json:"late"`
	Unstaffed int `db:"unstaffed" json:"unstaffed"`
	Partial   int `db:"partial" json:"partial"`
	OpenNow   int `db:"openNow" json:"openNow"`
}

// LateWindow الحجوزات المجدولة بين (اليوم-from) و(اليوم-to): كم تأخّر عن يومه.
// المتأخر = خلص بعد يوم موعده، أو بعده مفتوح وموعده فات.
func (r *MatrixCommandRepository) LateWindow(from, to int) (*LateWindow, error) {
	var w LateWindow
	err := r.db.Get(&w, `
		SELECT COUNT(*) AS total,
		  COUNT(*) FILTER (WHERE (b.status = 'COMPLETED' AND baghdad_date(b."completedAt") > baghdad_date(b."scheduledAt"))
		                      OR (b.status::text NOT IN ('COMPLETED','CANCELLED') AND baghdad_date(b."scheduledAt") < baghdad_today())) AS late,
		  COUNT(*) FILTER (WHERE NOT EXISTS (SELECT 1 FROM "BookingAssignment" a WHERE a."bookingId" = b.id)) AS unstaffed,
		  COUNT(*) FILTER (WHERE b."partialCount" > 0) AS partial,
		  COUNT(*) FILTER (WHERE b.status::text NOT IN ('COMPLETED','CANCELLED') AND baghdad_date(b."scheduledAt") < baghdad_today()) AS "openNow"
		FROM "Booking" b
		WHERE baghdad_date(b."scheduledAt") BETWEEN baghdad_today() - $1::int AND baghdad_today() - $2::int
		  AND b.status::text <> 'CANCELLED' AND b."archivedAt" IS NULL AND b."bookingType" IS DISTINCT FROM 'INTERNAL'`, from, to)
	return &w, err
}

func (r *MatrixCommandRepository) AsksToday(employeeID string) int {
	var n int
	_ = r.db.Get(&n, `SELECT COUNT(*) FROM "MatrixChatMessage" WHERE "employeeId" = $1 AND "createdAt" >= baghdad_today()::timestamp`, employeeID)
	return n
}

func (r *MatrixCommandRepository) LogAsk(employeeID, q, a, source string) {
	_, _ = r.db.Exec(`INSERT INTO "MatrixChatMessage" (id, "employeeId", question, answer, source) VALUES ($1,$2,$3,$4,$5)`,
		uuid.NewString(), employeeID, q, a, source)
}

// LateItem حجز متأخر بآخر ٣ أيام — (ع): «من اضغط يوديني للحجوزات المتأخرة صدك».
type LateItem struct {
	ID          string     `db:"id" json:"id"`
	Code        string     `db:"code" json:"code"`
	Customer    *string    `db:"customer" json:"customer"`
	Service     *string    `db:"service" json:"service"`
	Status      string     `db:"status" json:"status"`
	ScheduledAt time.Time  `db:"scheduledAt" json:"scheduledAt"`
	CompletedAt *time.Time `db:"completedAt" json:"completedAt"`
	Leader      *string    `db:"leader" json:"leader"`
	Unstaffed   bool       `db:"unstaffed" json:"unstaffed"`
	Partial     bool       `db:"partial" json:"partial"`
	DaysLate    int        `db:"daysLate" json:"daysLate"`
	Reason      *string    `db:"reason" json:"reason"`
}

// LateItems نفس تعريف «المتأخر» بـLateWindow بالضبط — الأحدث أول.
func (r *MatrixCommandRepository) LateItems(from, to, limit int) ([]LateItem, error) {
	rows := []LateItem{}
	err := r.db.Select(&rows, `
		SELECT b.id, b.code, c.name AS customer, s.name AS service, b.status::text AS status, b."scheduledAt", b."completedAt",
		       (SELECT e.name FROM "Mission" m JOIN "Employee" e ON e.id = m."leaderId" WHERE m."bookingId" = b.id
		         ORDER BY m."assignedAt" DESC NULLS LAST LIMIT 1) AS leader,
		       NOT EXISTS (SELECT 1 FROM "BookingAssignment" a WHERE a."bookingId" = b.id) AS unstaffed,
		       b."partialCount" > 0 AS partial,
		       (COALESCE(baghdad_date(b."completedAt"), baghdad_today()) - baghdad_date(b."scheduledAt"))::int AS "daysLate",
		       COALESCE(
		         (SELECT NULLIF(btrim(me.note), '') FROM "Mission" m JOIN "MissionEvent" me ON me."missionId" = m.id
		           WHERE m."bookingId" = b.id AND me.note IS NOT NULL AND btrim(me.note) <> ''
		           ORDER BY me."createdAt" DESC LIMIT 1),
		         NULLIF(btrim(b."postponeReason"), '')) AS reason
		FROM "Booking" b
		LEFT JOIN "Customer" c ON c.id = b."customerId"
		LEFT JOIN "Service" s ON s.id = b."serviceId"
		WHERE baghdad_date(b."scheduledAt") BETWEEN baghdad_today() - $1::int AND baghdad_today() - $2::int
		  AND b.status::text <> 'CANCELLED' AND b."archivedAt" IS NULL AND b."bookingType" IS DISTINCT FROM 'INTERNAL'
		  AND ((b.status = 'COMPLETED' AND baghdad_date(b."completedAt") > baghdad_date(b."scheduledAt"))
		       OR (b.status::text NOT IN ('COMPLETED','CANCELLED') AND baghdad_date(b."scheduledAt") < baghdad_today()))
		ORDER BY b."scheduledAt" DESC LIMIT $3`, from, to, limit)
	return rows, err
}
