package repository

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// MatrixReportRepository مادة تقرير الموظف اليومي وملف سلوكه.
type MatrixReportRepository struct{ db *sqlx.DB }

func NewMatrixReportRepository(db *sqlx.DB) *MatrixReportRepository {
	return &MatrixReportRepository{db: db}
}

type ReportEmployee struct {
	ID            string         `db:"id"`
	Name          string         `db:"name"`
	Role          string         `db:"role"`
	IsLeader      bool           `db:"isLeader"`
	Shift         sql.NullString `db:"shift"`
	ShiftStart    sql.NullString `db:"shiftStart"`
	ShiftEnd      sql.NullString `db:"shiftEnd"`
	MonthlyLeaves int            `db:"monthlyLeaves"`
}

func (r *MatrixReportRepository) Employee(id string) (*ReportEmployee, error) {
	var e ReportEmployee
	err := r.db.Get(&e, `SELECT id, name, role::text AS role, "isLeader", shift::text AS shift, "shiftStart", "shiftEnd", COALESCE("monthlyLeaves",2) AS "monthlyLeaves" FROM "Employee" WHERE id = $1`, id)
	return &e, err
}

type DayAttendance struct {
	CheckIn  time.Time    `db:"checkIn"`
	CheckOut sql.NullTime `db:"checkOut"`
}

func (r *MatrixReportRepository) Attendance(id, day string) ([]DayAttendance, error) {
	rows := []DayAttendance{}
	err := r.db.Select(&rows, `SELECT "checkIn", "checkOut" FROM "Attendance" WHERE "employeeId" = $1 AND baghdad_date("checkIn") = $2::date ORDER BY "checkIn"`, id, day)
	return rows, err
}

// AttendanceRange حضور فترة — لملف السلوك.
func (r *MatrixReportRepository) AttendanceRange(id string, days int) ([]DayAttendance, error) {
	rows := []DayAttendance{}
	err := r.db.Select(&rows, `SELECT "checkIn", "checkOut" FROM "Attendance" WHERE "employeeId" = $1 AND "checkIn" > now() - make_interval(days => $2) ORDER BY "checkIn"`, id, days)
	return rows, err
}

// DayJob حجز اشتغل بي الموظف بيوم، بكل مراحله.
type DayJob struct {
	ID               string         `db:"id"`
	Code             string         `db:"code"`
	Status           string         `db:"status"`
	Service          sql.NullString `db:"service"`
	ScheduledAt      sql.NullTime   `db:"scheduledAt"`
	AssignedAt       sql.NullTime   `db:"assignedAt"`
	MaterialsReadyAt sql.NullTime   `db:"materialsReadyAt"`
	MaterialsBy      sql.NullString `db:"materialsBy"`
	DepartedAt       sql.NullTime   `db:"departedAt"`
	ArrivedAt        sql.NullTime   `db:"arrivedAt"`
	StartedAt        sql.NullTime   `db:"startedAt"`
	CompletedAt      sql.NullTime   `db:"completedAt"`
	LeaderID         sql.NullString `db:"leaderId"`
	LeaderName       sql.NullString `db:"leaderName"`
}

func (r *MatrixReportRepository) Jobs(id, day string) ([]DayJob, error) {
	rows := []DayJob{}
	err := r.db.Select(&rows, `
		SELECT b.id, b.code, b.status::text AS status, s.name AS service, b."scheduledAt",
		       m."assignedAt",
		       COALESCE(m."materialsReadyAt", b."materialsReadyAt") AS "materialsReadyAt",
		       me.name AS "materialsBy",
		       m."departedAt",
		       COALESCE(m."arrivedAt", b."arrivedAt") AS "arrivedAt",
		       COALESCE(m."workStartedAt", b."startedAt") AS "startedAt",
		       COALESCE(m."completedAt", b."completedAt") AS "completedAt",
		       m."leaderId", le.name AS "leaderName"
		FROM "Booking" b
		LEFT JOIN "Service" s ON s.id = b."serviceId"
		LEFT JOIN LATERAL (SELECT * FROM "Mission" WHERE "bookingId" = b.id ORDER BY "assignedAt" DESC NULLS LAST LIMIT 1) m ON true
		LEFT JOIN "Employee" me ON me.id = b."materialsReadyById"
		LEFT JOIN "Employee" le ON le.id = m."leaderId"
		WHERE (baghdad_date(b."scheduledAt") = $2::date OR baghdad_date(b."completedAt") = $2::date)
		  AND (EXISTS (SELECT 1 FROM "BookingAssignment" a WHERE a."bookingId" = b.id AND a."employeeId" = $1)
		       OR m."leaderId" = $1 OR $1 = ANY(m."memberIds"))
		ORDER BY b."scheduledAt" NULLS LAST`, id, day)
	return rows, err
}

type LeaveRow struct {
	StartDate time.Time `db:"startDate"`
	EndDate   time.Time `db:"endDate"`
	Status    string    `db:"status"`
}

// Leaves إجازات الموظف من تاريخ.
func (r *MatrixReportRepository) Leaves(id string, since time.Time) ([]LeaveRow, error) {
	rows := []LeaveRow{}
	err := r.db.Select(&rows, `SELECT "startDate", "endDate", status FROM "LeaveRequest" WHERE "employeeId" = $1 AND "startDate" >= $2::date AND status <> 'REJECTED' ORDER BY "startDate"`, id, since)
	return rows, err
}

type ReminderStats struct {
	Total     int            `db:"total"`
	Resolved  int            `db:"resolved"`
	Escalated int            `db:"escalated"`
	TopKind   sql.NullString `db:"topKind"`
	TopCount  int            `db:"topCount"`
}

func (r *MatrixReportRepository) Reminders(id string) (*ReminderStats, error) {
	var s ReminderStats
	err := r.db.Get(&s, `
		SELECT COUNT(*) AS total,
		       COUNT(*) FILTER (WHERE "resolvedAt" IS NOT NULL) AS resolved,
		       COUNT(*) FILTER (WHERE "escalatedAt" IS NOT NULL) AS escalated,
		       (SELECT kind FROM "AiAction" WHERE "targetEmployeeId" = $1 AND "createdAt" > now() - interval '30 days' GROUP BY kind ORDER BY COUNT(*) DESC LIMIT 1) AS "topKind",
		       COALESCE((SELECT COUNT(*) FROM "AiAction" WHERE "targetEmployeeId" = $1 AND "createdAt" > now() - interval '30 days' GROUP BY kind ORDER BY COUNT(*) DESC LIMIT 1),0) AS "topCount"
		FROM "AiAction" WHERE "targetEmployeeId" = $1 AND "createdAt" > now() - interval '30 days'`, id)
	return &s, err
}

type AchievementStats struct {
	Days        int `db:"days"`
	Total       int `db:"total"`
	Duplicates  int `db:"duplicates"`
	Good        int `db:"good"`
	NeedsReview int `db:"needsReview"`
	AvgLen      int `db:"avgLen"`
}

// Achievements نمط «إنجازاتي اليومية» — أرقام بس، بلا قراءة النص لبرّا.
func (r *MatrixReportRepository) Achievements(id string) (*AchievementStats, error) {
	var s AchievementStats
	err := r.db.Get(&s, `
		SELECT COUNT(DISTINCT baghdad_date("createdAt")) AS days, COUNT(*) AS total,
		       COUNT(*) - COUNT(DISTINCT lower(trim("reportText"))) AS duplicates,
		       COUNT(*) FILTER (WHERE "reviewStatus" = 'GOOD') AS good,
		       COUNT(*) FILTER (WHERE "reviewStatus" = 'NEEDS_REVIEW') AS "needsReview",
		       COALESCE(AVG(length("reportText")),0)::int AS "avgLen"
		FROM "Achievement" WHERE "employeeId" = $1 AND "createdAt" > now() - interval '30 days'`, id)
	return &s, err
}

// Complaints شكاوى مرتبطة بشغله (٣٠ يوم).
func (r *MatrixReportRepository) Complaints(id string) int {
	var n int
	_ = r.db.Get(&n, `SELECT COUNT(*) FROM "Complaint" WHERE "relatedEmployeeId" = $1 AND "createdAt" > now() - interval '30 days'`, id)
	return n
}

// AdjustedInvoices فواتيره الي انعدّلت (٣٠ يوم).
func (r *MatrixReportRepository) AdjustedInvoices(id string) int {
	var n int
	_ = r.db.Get(&n, `SELECT COUNT(*) FROM "LeaderInvoice" WHERE "employeeId" = $1 AND "adjustedAt" IS NOT NULL AND "createdAt" > now() - interval '30 days'`, id)
	return n
}

// ResponseMinutes متوسط دقائق من التكليف لتجهيز المواد (٣٠ يوم، الليدر).
func (r *MatrixReportRepository) ResponseMinutes(id string) (avg float64, n int) {
	var row struct {
		Avg sql.NullFloat64 `db:"avg"`
		N   int             `db:"n"`
	}
	_ = r.db.Get(&row, `SELECT AVG(EXTRACT(EPOCH FROM ("materialsReadyAt" - "assignedAt"))/60) AS avg, COUNT(*) AS n
		FROM "Mission" WHERE "leaderId" = $1 AND "assignedAt" > now() - interval '30 days'
		  AND "materialsReadyAt" IS NOT NULL AND "materialsReadyAt" > "assignedAt"`, id)
	return row.Avg.Float64, row.N
}

var _ = pq.Array

type PerfJob struct {
	ID       string `db:"id"`
	Code     string `db:"code"`
	Service  string `db:"service"`
	Actual   int    `db:"actual"`
	Expected int    `db:"expected"`
}

type PerfStats struct {
	Total, Completed, Partial, Timed, Outliers int
	// Unstarted مفتوحة وما بدا بيها أحد (ما انثبتت، تأجلت، انتقلت) — تنسيق، مو تقصير الكادر.
	Unstarted              int
	AvgActual, AvgExpected float64
	GroupPartialRate       float64 // -1 = ما معروف
	Slow                   []PerfJob
}

const perfJobsCTE = `
	WITH mine AS (
		SELECT DISTINCT b.id, b.code, b.status::text AS status, b."partialCount", b."serviceId",
		       COALESCE(s.name, '') AS service, (b."startedAt" IS NOT NULL) AS started,
		       EXTRACT(EPOCH FROM (b."completedAt" - b."startedAt")) / 60 AS actual
		FROM "Booking" b
		LEFT JOIN "Service" s ON s.id = b."serviceId"
		LEFT JOIN "Mission" m ON m."bookingId" = b.id
		WHERE b."scheduledAt" >= now() - make_interval(days => $2 + 30)
		  AND b."scheduledAt" <  now() - make_interval(days => $2)
		  AND b.status::text <> 'CANCELLED'
		  AND (EXISTS (SELECT 1 FROM "BookingAssignment" a WHERE a."bookingId" = b.id AND a."employeeId" = $1)
		       OR m."leaderId" = $1 OR $1 = ANY(m."memberIds"))
	), med AS (
		SELECT b."serviceId", percentile_cont(0.5) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM (b."completedAt" - b."startedAt")) / 60) AS expected
		FROM "Booking" b
		WHERE b."completedAt" IS NOT NULL AND b."startedAt" IS NOT NULL
		  AND b."completedAt" - b."startedAt" >= interval '5 minutes' AND b."completedAt" - b."startedAt" < interval '14 hours'
		  AND b."completedAt" >= now() - interval '180 days'
		GROUP BY b."serviceId"
		-- ⚠️ خدمة وسيطها دقائق قليلة (تسجيل بدء/إنجاز بنفس اللحظة) چانت تطلّع
		-- سرعة ×900 لأي حجز عادي — المتوقع لازم يكون ١٥ دقيقة فأكثر حتى ينحسب.
		HAVING COUNT(*) >= 3 AND percentile_cont(0.5) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM (b."completedAt" - b."startedAt")) / 60) >= 15
	)`

// Performance أداء الموظف بنافذة ٣٠ يوم تنتهي قبل offsetDays.
func (r *MatrixReportRepository) Performance(id string, offsetDays int) (*PerfStats, error) {
	var agg struct {
		Total     int             `db:"total"`
		Completed int             `db:"completed"`
		Partial   int             `db:"partial"`
		Timed     int             `db:"timed"`
		Outliers  int             `db:"outliers"`
		Unstarted int             `db:"unstarted"`
		AvgAct    sql.NullFloat64 `db:"avg_act"`
		AvgExp    sql.NullFloat64 `db:"avg_exp"`
	}
	err := r.db.Get(&agg, perfJobsCTE+`
		SELECT COUNT(*) AS total,
		       COUNT(*) FILTER (WHERE status = 'COMPLETED' AND "partialCount" = 0) AS completed,
		       COUNT(*) FILTER (WHERE "partialCount" > 0) AS partial,
		       COUNT(*) FILTER (WHERE actual >= 5 AND actual < 840 AND med.expected IS NOT NULL) AS timed,
		       COUNT(*) FILTER (WHERE actual >= 840) AS outliers,
		       COUNT(*) FILTER (WHERE status <> 'COMPLETED' AND "partialCount" = 0 AND NOT started) AS unstarted,
		       AVG(actual) FILTER (WHERE actual >= 5 AND actual < 840 AND med.expected IS NOT NULL) AS avg_act,
		       AVG(med.expected) FILTER (WHERE actual >= 5 AND actual < 840) AS avg_exp
		FROM mine LEFT JOIN med ON med."serviceId" = mine."serviceId"`, id, offsetDays)
	if err != nil {
		return nil, err
	}
	st := &PerfStats{Total: agg.Total, Completed: agg.Completed, Partial: agg.Partial, Timed: agg.Timed, Outliers: agg.Outliers, Unstarted: agg.Unstarted,
		AvgActual: agg.AvgAct.Float64, AvgExpected: agg.AvgExp.Float64, GroupPartialRate: -1, Slow: []PerfJob{}}
	if offsetDays == 0 {
		_ = r.db.Select(&st.Slow, perfJobsCTE+`
			SELECT mine.id, mine.code, mine.service, ROUND(actual)::int AS actual, ROUND(med.expected)::int AS expected
			FROM mine JOIN med ON med."serviceId" = mine."serviceId"
			WHERE actual >= 5 AND actual < 840 AND actual > med.expected * 1.25
			ORDER BY actual / NULLIF(med.expected, 0) DESC LIMIT 5`, id, offsetDays)
		var rate sql.NullFloat64
		if r.db.Get(&rate, `SELECT AVG(("partialCount" > 0)::int) FROM "Booking"
			WHERE "scheduledAt" >= now() - interval '30 days' AND status::text <> 'CANCELLED'`) == nil && rate.Valid {
			st.GroupPartialRate = rate.Float64
		}
	}
	return st, nil
}

// SpeedMedian وسيط (الفعلي ÷ المتوقع) لكل حجز بنافذة ٣٠ يوم — مقاوم للحجز
// الشاذ. حجز واحد مسجّل بدايته غلط (أيام) چان يخلّي المعدّل ×٤٢ ويطلّع
// «صار يطوّل ٣٢٤٪». نقص كل حجز عند ×١٠ ونأخذ الوسيط مو المعدّل.
func (r *MatrixReportRepository) SpeedMedian(id string, offsetDays int) (float64, int, error) {
	var out struct {
		Med sql.NullFloat64 `db:"med"`
		N   int             `db:"n"`
	}
	err := r.db.Get(&out, perfJobsCTE+`
		SELECT percentile_cont(0.5) WITHIN GROUP (ORDER BY LEAST(actual / med.expected, 10)) AS med, COUNT(*) AS n
		FROM mine JOIN med ON med."serviceId" = mine."serviceId"
		WHERE actual >= 5 AND actual < 840 AND med.expected > 0`, id, offsetDays)
	return out.Med.Float64, out.N, err
}
