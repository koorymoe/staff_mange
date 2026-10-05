package repository

import (
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// ═══ ماتركس يقترح — البيانات الخام ═══

type MatrixSuggestRepository struct{ db *sqlx.DB }

func NewMatrixSuggestRepository(db *sqlx.DB) *MatrixSuggestRepository {
	return &MatrixSuggestRepository{db: db}
}

type SuggestBooking struct {
	ID           string         `db:"id"`
	Code         string         `db:"code"`
	Status       string         `db:"status"`
	ServiceID    *string        `db:"serviceId"`
	ServiceName  *string        `db:"serviceName"`
	ScheduledAt  *time.Time     `db:"scheduledAt"`
	SupervisorID *string        `db:"supervisorId"`
	TechIDs      pq.StringArray `db:"techIds"`
	Solo         bool           `db:"solo"`
}

func (r *MatrixSuggestRepository) Booking(id string) (*SuggestBooking, error) {
	var b SuggestBooking
	err := r.db.Get(&b, `SELECT b.id, b.code, b.status::text AS status, b."serviceId", s.name AS "serviceName", b."scheduledAt",
	       b."projectSupervisorId" AS "supervisorId", COALESCE(s."managerHandlesPaperwork", false) AS solo,
	       COALESCE((SELECT array_agg(a."employeeId" ORDER BY a.role) FROM "BookingAssignment" a WHERE a."bookingId" = b.id), '{}') AS "techIds"
	  FROM "Booking" b LEFT JOIN "Service" s ON s.id = b."serviceId" WHERE b.id = $1`, id)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// ServiceDurations الوسيط الحقيقي لمدة الشغل (بدأ → خلص) لكل خدمة، آخر ١٨٠ يوم.
func (r *MatrixSuggestRepository) ServiceDurations() (map[string]int, int) {
	rows := []struct {
		ServiceID string  `db:"serviceId"`
		Minutes   float64 `db:"minutes"`
		N         int     `db:"n"`
	}{}
	_ = r.db.Select(&rows, `SELECT "serviceId", percentile_cont(0.5) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM ("completedAt" - "startedAt")) / 60) AS minutes, count(*)::int AS n
		FROM "Booking" WHERE "serviceId" IS NOT NULL AND "startedAt" IS NOT NULL AND "completedAt" > "startedAt"
		  AND "completedAt" > now() - interval '180 days' AND "completedAt" - "startedAt" < interval '12 hours'
		GROUP BY "serviceId"`)
	out := map[string]int{}
	for _, x := range rows {
		if x.N >= 3 {
			out[x.ServiceID] = int(x.Minutes)
		}
	}
	var all float64
	_ = r.db.Get(&all, `SELECT COALESCE(percentile_cont(0.5) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM ("completedAt" - "startedAt")) / 60), 120)
		FROM "Booking" WHERE "startedAt" IS NOT NULL AND "completedAt" > "startedAt" AND "completedAt" - "startedAt" < interval '12 hours'
		  AND "completedAt" > now() - interval '180 days'`)
	return out, int(all)
}

// WorkHours ساعات بدء المواعيد المعتادة بالشركة (بتوقيت بغداد): المئين ١٠ و٩٠،
// وأيام الأسبوع الي ما يشتغلون بيها (أقل من ٣٪ من المواعيد).
func (r *MatrixSuggestRepository) WorkHours() (from, to int, offDays []int) {
	from, to = 9, 16
	var row struct {
		P10 *float64 `db:"p10"`
		P90 *float64 `db:"p90"`
		N   int      `db:"n"`
	}
	_ = r.db.Get(&row, `SELECT percentile_cont(0.1) WITHIN GROUP (ORDER BY h) AS p10, percentile_cont(0.9) WITHIN GROUP (ORDER BY h) AS p90, count(*)::int AS n
		FROM (SELECT EXTRACT(HOUR FROM "scheduledAt" AT TIME ZONE 'Asia/Baghdad') AS h FROM "Booking"
		      WHERE "scheduledAt" > now() - interval '180 days' AND "scheduledAt" < now()) x`)
	if row.N >= 20 && row.P10 != nil && row.P90 != nil && *row.P90 > *row.P10 {
		from, to = int(*row.P10), int(*row.P90)
	}
	days := []struct {
		D int `db:"d"`
		N int `db:"n"`
	}{}
	_ = r.db.Select(&days, `SELECT EXTRACT(DOW FROM "scheduledAt" AT TIME ZONE 'Asia/Baghdad')::int AS d, count(*)::int AS n
		FROM "Booking" WHERE "scheduledAt" > now() - interval '180 days' AND "scheduledAt" < now() GROUP BY 1`)
	total := 0
	seen := map[int]int{}
	for _, d := range days {
		total += d.N
		seen[d.D] = d.N
	}
	if total >= 50 {
		for d := 0; d < 7; d++ {
			if seen[d]*100 < total*3 {
				offDays = append(offDays, d)
			}
		}
	} else {
		offDays = []int{5} // الجمعة — لحد ما تتجمع بيانات كافية
	}
	return
}

type Busy struct {
	EmployeeID  string    `db:"employeeId"`
	BookingID   string    `db:"bookingId"`
	ServiceID   *string   `db:"serviceId"`
	ScheduledAt time.Time `db:"scheduledAt"`
}

// BusyBetween منو مشغول بحجز (ليدر أو فني) بين وقتين — الحجوزات الحية بس.
func (r *MatrixSuggestRepository) BusyBetween(from, to time.Time, exceptBooking string) ([]Busy, error) {
	rows := []Busy{}
	err := r.db.Select(&rows, `
		SELECT x."employeeId", b.id AS "bookingId", b."serviceId", b."scheduledAt"
		FROM "Booking" b
		JOIN LATERAL (SELECT b."projectSupervisorId" AS "employeeId" WHERE b."projectSupervisorId" IS NOT NULL
		              UNION SELECT a."employeeId" FROM "BookingAssignment" a WHERE a."bookingId" = b.id) x ON true
		WHERE b."scheduledAt" BETWEEN $1 AND $2 AND b.id <> $3
		  AND b.status::text IN ('PENDING', 'CONFIRMED', 'IN_PROGRESS', 'WAITING')`, from, to, exceptBooking)
	return rows, err
}

type SuggestPerson struct {
	ID        string   `db:"id" json:"id"`
	Name      string   `db:"name" json:"name"`
	IsLeader  bool     `db:"isLeader" json:"isLeader"`
	SameDone  int      `db:"sameDone" json:"sameDone"`   // حجوزات منجزة بنفس الخدمة (١٨٠ يوم)
	WorkMed   *float64 `db:"workMed" json:"workMed"`     // وسيط مدة شغله بنفس الخدمة
	RatingAvg *float64 `db:"ratingAvg" json:"ratingAvg"` // تقييم الليدرية إله
	RatingN   int      `db:"ratingN" json:"ratingN"`
}

// People الليدرية والفنيين الفعّالين ويا خبرتهم بالخدمة وتقييمهم.
// الغايبين بإجازة بنفس اليوم ينشالون.
func (r *MatrixSuggestRepository) People(serviceID *string, day time.Time) ([]SuggestPerson, error) {
	rows := []SuggestPerson{}
	err := r.db.Select(&rows, `
		SELECT e.id, e.name, e."isLeader",
		  (SELECT count(DISTINCT b.id) FROM "Booking" b LEFT JOIN "BookingAssignment" a ON a."bookingId" = b.id
		    WHERE (b."projectSupervisorId" = e.id OR a."employeeId" = e.id) AND b."serviceId" IS NOT DISTINCT FROM $1
		      AND b."completedAt" > now() - interval '180 days')::int AS "sameDone",
		  (SELECT percentile_cont(0.5) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM (b."completedAt" - b."startedAt")) / 60)
		     FROM "Booking" b LEFT JOIN "BookingAssignment" a ON a."bookingId" = b.id
		    WHERE (b."projectSupervisorId" = e.id OR a."employeeId" = e.id) AND b."serviceId" IS NOT DISTINCT FROM $1
		      AND b."startedAt" IS NOT NULL AND b."completedAt" > b."startedAt" AND b."completedAt" - b."startedAt" < interval '12 hours'
		      AND b."completedAt" > now() - interval '180 days') AS "workMed",
		  (SELECT avg(score) FROM "CrewRating" c WHERE c."technicianId" = e.id AND c."createdAt" > now() - interval '180 days') AS "ratingAvg",
		  (SELECT count(*) FROM "CrewRating" c WHERE c."technicianId" = e.id AND c."createdAt" > now() - interval '180 days')::int AS "ratingN"
		FROM "Employee" e
		WHERE e.status = 'ACTIVE' AND (e."isLeader" OR e.role::text = 'TECHNICIAN')
		  AND NOT EXISTS (SELECT 1 FROM "LeaveRequest" l WHERE l."employeeId" = e.id AND l.status::text = 'APPROVED'
		                  AND $2::date BETWEEN l."startDate"::date AND l."endDate"::date)`, serviceID, day)
	return rows, err
}

// CrewSize عدد الفنيين (بلا الليدر) المعتاد لهالخدمة.
func (r *MatrixSuggestRepository) CrewSize(serviceID *string) int {
	var n *float64
	_ = r.db.Get(&n, `SELECT percentile_disc(0.5) WITHIN GROUP (ORDER BY c) FROM (
		SELECT count(a.id) AS c FROM "Booking" b JOIN "BookingAssignment" a ON a."bookingId" = b.id
		 WHERE b."serviceId" IS NOT DISTINCT FROM $1 AND b."completedAt" > now() - interval '180 days'
		   AND a."employeeId" IS DISTINCT FROM b."projectSupervisorId"
		 GROUP BY b.id) x`, serviceID)
	if n == nil || *n < 1 {
		return 1
	}
	if *n > 3 {
		return 3
	}
	return int(*n)
}

// ═══ سجل الاقتراحات ═══

type SuggestionRow struct {
	ID        string     `db:"id" json:"id"`
	BookingID string     `db:"bookingId" json:"bookingId"`
	Code      string     `db:"code" json:"code"`
	Kind      string     `db:"kind" json:"kind"`
	Suggested []byte     `db:"suggested" json:"-"`
	Actual    []byte     `db:"actual" json:"-"`
	Reason    string     `db:"reason" json:"reason"`
	Outcome   string     `db:"outcome" json:"outcome"`
	CreatedAt time.Time  `db:"createdAt" json:"createdAt"`
	DecidedAt *time.Time `db:"decidedAt" json:"decidedAt"`
}

// Log يحفظ الاقتراح — ويحدّثه إذا بعده ما انحسم (الاقتراح الأخير هو الي ينقاس).
func (r *MatrixSuggestRepository) Log(bookingID, kind string, suggested any, reason string) {
	j, _ := json.Marshal(suggested)
	_, _ = r.db.Exec(`INSERT INTO "MatrixSuggestion" ("bookingId", kind, suggested, reason) VALUES ($1, $2, $3, $4)
		ON CONFLICT ("bookingId", kind) DO UPDATE SET suggested = EXCLUDED.suggested, reason = EXCLUDED.reason, "createdAt" = now()
		WHERE "MatrixSuggestion".outcome = 'PENDING'`, bookingID, kind, j, reason)
}

func (r *MatrixSuggestRepository) Pending() ([]SuggestionRow, error) {
	rows := []SuggestionRow{}
	err := r.db.Select(&rows, `SELECT m.id, m."bookingId", b.code, m.kind, m.suggested, m.actual, m.reason, m.outcome, m."createdAt", m."decidedAt"
		FROM "MatrixSuggestion" m JOIN "Booking" b ON b.id = m."bookingId" WHERE m.outcome = 'PENDING'`)
	return rows, err
}

func (r *MatrixSuggestRepository) Decide(id, outcome string, actual any) {
	j, _ := json.Marshal(actual)
	_, _ = r.db.Exec(`UPDATE "MatrixSuggestion" SET outcome = $2, actual = $3, "decidedAt" = now() WHERE id = $1`, id, outcome, j)
}

// Decided الاقتراحات المحسومة بآخر N يوم ويا نتيجة الحجز بعدين.
type DecidedRow struct {
	SuggestionRow
	StartedAt   *time.Time `db:"startedAt" json:"-"`
	ScheduledAt *time.Time `db:"scheduledAt" json:"-"`
	CompletedAt *time.Time `db:"completedAt" json:"-"`
}

func (r *MatrixSuggestRepository) Decided(days int) ([]DecidedRow, error) {
	rows := []DecidedRow{}
	err := r.db.Select(&rows, `SELECT m.id, m."bookingId", b.code, m.kind, m.suggested, m.actual, m.reason, m.outcome, m."createdAt", m."decidedAt",
	       b."startedAt", b."scheduledAt", b."completedAt"
		FROM "MatrixSuggestion" m JOIN "Booking" b ON b.id = m."bookingId"
		WHERE m.outcome <> 'PENDING' AND m."decidedAt" > now() - make_interval(days => $1)
		ORDER BY m."decidedAt" DESC`, days)
	return rows, err
}

// AssignmentRoles الدور ← الموظف بتكليفات الحجز الحالية.
func (r *MatrixSuggestRepository) AssignmentRoles(bookingID string) map[string]string {
	rows := []struct {
		Role string `db:"role"`
		Emp  string `db:"employeeId"`
	}{}
	_ = r.db.Select(&rows, `SELECT role::text AS role, "employeeId" FROM "BookingAssignment" WHERE "bookingId" = $1`, bookingID)
	out := map[string]string{}
	for _, x := range rows {
		out[x.Role] = x.Emp
	}
	return out
}
