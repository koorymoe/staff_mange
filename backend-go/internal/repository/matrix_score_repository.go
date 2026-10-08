package repository

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// ═══ تقييم ماتركس + تقييمات البشر ═══
// ⚠️ درجة تقييم بس — ما يلمس الفلوس ولا نقاط الانضباط ولا KpiEvaluation.

type MatrixScoreRepository struct{ db *sqlx.DB }

func NewMatrixScoreRepository(db *sqlx.DB) *MatrixScoreRepository {
	return &MatrixScoreRepository{db: db}
}

type MatrixScoreRow struct {
	ID          string     `db:"id" json:"id"`
	EmployeeID  string     `db:"employeeId" json:"employeeId"`
	Source      string     `db:"source" json:"source"`
	SourceID    string     `db:"sourceId" json:"sourceId"`
	SourceLabel *string    `db:"sourceLabel" json:"sourceLabel"`
	Rule        string     `db:"rule" json:"rule"`
	Points      int        `db:"points" json:"points"`
	MaxPoints   int        `db:"maxPoints" json:"maxPoints"`
	Reason      string     `db:"reason" json:"reason"`
	At          time.Time  `db:"at" json:"at"`
	CancelledAt *time.Time `db:"cancelledAt" json:"cancelledAt"`
	CancelNote  *string    `db:"cancelNote" json:"cancelNote"`
}

// Add نقطة وحدة — إذا انحسبت قبل (نفس الموظف/المصدر/القاعدة) ما تتكرر؛ بس
// إذا الحكم تغيّر (محطة «ما صارت» وصارت بعدين متأخرة) تتحدّث. الملغية ما تنلمس.
func (r *MatrixScoreRepository) Add(s MatrixScoreRow) (bool, error) {
	res, err := r.db.Exec(`
		INSERT INTO "MatrixScore" ("employeeId", source, "sourceId", "sourceLabel", rule, points, "maxPoints", reason, "at")
		SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9
		WHERE EXISTS (SELECT 1 FROM "Employee" e WHERE e.id = $1 AND e.role::text NOT IN ('ADMIN', 'OWNER'))
		ON CONFLICT ("employeeId", source, "sourceId", rule) DO UPDATE
		SET points = EXCLUDED.points, reason = EXCLUDED.reason, "at" = EXCLUDED."at"
		WHERE "MatrixScore"."cancelledAt" IS NULL AND "MatrixScore".points <> EXCLUDED.points`,
		s.EmployeeID, s.Source, s.SourceID, s.SourceLabel, s.Rule, s.Points, s.MaxPoints, s.Reason, s.At)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// Scorable الموظفين الي ينقيّمون — الكل عدا المالك ومدير النظام.
type Scorable struct {
	ID               string    `db:"id" json:"id"`
	Name             string    `db:"name" json:"name"`
	Role             string    `db:"role" json:"role"`
	CreatedAt        time.Time `db:"createdAt" json:"-"`
	IsLeader         bool      `db:"isLeader" json:"isLeader"`
	IsServiceManager bool      `db:"isServiceManager" json:"isServiceManager"`
}

func (r *MatrixScoreRepository) Scorables() ([]Scorable, error) {
	rows := []Scorable{}
	err := r.db.Select(&rows, `SELECT id, name, role::text AS role, "createdAt", COALESCE("isLeader", false) AS "isLeader",
		EXISTS (SELECT 1 FROM "ServiceManager" sm WHERE sm."employeeId" = "Employee".id) AS "isServiceManager"
		FROM "Employee" WHERE status = 'ACTIVE' AND role::text NOT IN ('ADMIN', 'OWNER') ORDER BY name`)
	return rows, err
}

// DayFacts حضور يوم + الإجازات المعتمدة بيه.
func (r *MatrixScoreRepository) DayFacts(day time.Time) (present map[string]time.Time, onLeave map[string]bool, err error) {
	present, onLeave = map[string]time.Time{}, map[string]bool{}
	type att struct {
		EmployeeID string    `db:"employeeId"`
		CheckIn    time.Time `db:"checkIn"`
	}
	rows := []att{}
	if err = r.db.Select(&rows, `SELECT "employeeId", min("checkIn") AS "checkIn" FROM "Attendance"
		WHERE date::date = $1::date GROUP BY 1`, day.Format("2006-01-02")); err != nil {
		return
	}
	for _, a := range rows {
		present[a.EmployeeID] = a.CheckIn
	}
	ids := []string{}
	if err = r.db.Select(&ids, `SELECT DISTINCT "employeeId" FROM "LeaveRequest"
		WHERE status = 'APPROVED' AND $1::date BETWEEN "startDate" AND "endDate"`, day.Format("2006-01-02")); err != nil {
		return
	}
	for _, id := range ids {
		onLeave[id] = true
	}
	return
}

// TaskDue مهمة مضافة موعدها فات أو خلصت.
type TaskDue struct {
	ID           string     `db:"id"`
	Title        string     `db:"title"`
	AssignedToID string     `db:"assignedToId"`
	DueAt        time.Time  `db:"dueAt"`
	DoneAt       *time.Time `db:"doneAt"`
}

// TasksDue المهام الي موعدها بين from وto (ما ملغية).
func (r *MatrixScoreRepository) TasksDue(from, to time.Time) ([]TaskDue, error) {
	rows := []TaskDue{}
	err := r.db.Select(&rows, `SELECT id, title, "assignedToId", "dueAt", "doneAt" FROM "ExtraTask"
		WHERE "dueAt" >= $1 AND "dueAt" < $2 AND status <> 'CANCELLED'`, from, to)
	return rows, err
}

// ═══ تقييمات البشر ═══

type StaffRatingInput struct {
	RaterID   string
	RateeID   string
	Stage     string
	BookingID string
	Period    string
	Score     int
	Note      *string
}

func (r *MatrixScoreRepository) SaveRating(in StaffRatingInput) error {
	_, err := r.db.Exec(`
		INSERT INTO "StaffRating" ("raterId", "rateeId", stage, "bookingId", period, score, note)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT ("raterId", "rateeId", stage, "bookingId", period)
		DO UPDATE SET score = EXCLUDED.score, note = EXCLUDED.note, "createdAt" = now()`,
		in.RaterID, in.RateeID, in.Stage, in.BookingID, in.Period, in.Score, in.Note)
	return err
}

// HumanRow تقييم بشري واحد (من StaffRating أو CrewRating).
type HumanRow struct {
	RateeID   string    `db:"rateeId" json:"-"`
	RaterName *string   `db:"raterName" json:"raterName"`
	Stage     string    `db:"stage" json:"stage"`
	Code      *string   `db:"code" json:"code"`
	Score     int       `db:"score" json:"score"`
	Note      *string   `db:"note" json:"note"`
	At        time.Time `db:"at" json:"at"`
}

const humanSQL = `
	SELECT * FROM (
		SELECT s."rateeId", e.name AS "raterName", s.stage, b.code, s.score, s.note, s."createdAt" AS "at"
		FROM "StaffRating" s JOIN "Employee" e ON e.id = s."raterId"
		LEFT JOIN "Booking" b ON b.id = NULLIF(s."bookingId", '')
		UNION ALL
		SELECT c."technicianId", e.name, 'LEADER_CREW', b.code, c.score, c.note, c."createdAt"
		FROM "CrewRating" c JOIN "Employee" e ON e.id = c."leaderId" JOIN "Booking" b ON b.id = c."bookingId"
	) h`

func (r *MatrixScoreRepository) Human(from, to time.Time, employeeID string) ([]HumanRow, error) {
	rows := []HumanRow{}
	q := humanSQL + ` WHERE h."at" >= $1 AND h."at" < $2`
	args := []any{from, to}
	if employeeID != "" {
		q += ` AND h."rateeId" = $3`
		args = append(args, employeeID)
	}
	err := r.db.Select(&rows, q+` ORDER BY h."at" DESC`, args...)
	return rows, err
}

func (r *MatrixScoreRepository) Scores(from, to time.Time, employeeID string) ([]MatrixScoreRow, error) {
	rows := []MatrixScoreRow{}
	q := `SELECT id, "employeeId", source, "sourceId", "sourceLabel", rule, points, "maxPoints", reason, "at", "cancelledAt", "cancelNote"
		FROM "MatrixScore" WHERE "at" >= $1 AND "at" < $2`
	args := []any{from, to}
	if employeeID != "" {
		q += ` AND "employeeId" = $3`
		args = append(args, employeeID)
	}
	err := r.db.Select(&rows, q+` ORDER BY "at" DESC`, args...)
	return rows, err
}

// DropStation يشيل نقاط محطة صار حكمها «ما تنطبق» (مثلاً الحجز تحوّل
// لإدارة المشاريع بعد ما انحسبت). الملغية بيد المدير تبقى بالسجل.
func (r *MatrixScoreRepository) DropStation(bookingID, rule string) error {
	_, err := r.db.Exec(`DELETE FROM "MatrixScore" WHERE source = 'BOOKING' AND "sourceId" = $1 AND rule = $2 AND "cancelledAt" IS NULL`,
		bookingID, rule)
	return err
}

// DropStations نفس DropStation بس دفعة وحدة (حجز+محطة) — للمحطات الي صارت «ما تنطبق».
func (r *MatrixScoreRepository) DropStations(bookingIDs, rules []string) error {
	if len(bookingIDs) == 0 {
		return nil
	}
	_, err := r.db.Exec(`DELETE FROM "MatrixScore" m USING unnest($1::text[], $2::text[]) AS x(b, rule)
		WHERE m.source = 'BOOKING' AND m."sourceId" = x.b AND m.rule = x.rule AND m."cancelledAt" IS NULL`,
		pq.Array(bookingIDs), pq.Array(rules))
	return err
}

// KeepOnlyStations يمسح نقاط الحجوزات الي ما عادت بالحساب الحالي: محطة صارت
// NA/تنتظر، أو انشالت (حجز مرحّل للتقني)، أو صاحبها تغيّر. keep = "حجز|قاعدة|موظف".
// الملغية يدوياً (cancelledAt) تبقى.
func (r *MatrixScoreRepository) KeepOnlyStations(bookingIDs, keep []string) error {
	if len(bookingIDs) == 0 {
		return nil
	}
	_, err := r.db.Exec(`DELETE FROM "MatrixScore" m WHERE m.source = 'BOOKING' AND m."cancelledAt" IS NULL
		AND m."sourceId" = ANY($1) AND NOT (m."sourceId" || '|' || m.rule || '|' || m."employeeId" = ANY($2))`,
		pq.Array(bookingIDs), pq.Array(keep))
	return err
}

// DropGoneBookings نقاط حجوزات انلغت أو تأرشفت أو انطلب حذفها — ما تنحسب على أحد.
func (r *MatrixScoreRepository) DropGoneBookings() error {
	_, err := r.db.Exec(`DELETE FROM "MatrixScore" m WHERE m.source = 'BOOKING' AND m."cancelledAt" IS NULL
		AND EXISTS (SELECT 1 FROM "Booking" b WHERE b.id = m."sourceId"
		            AND (b.status::text = 'CANCELLED' OR NOT (` + BookingCountableSQL("b") + `)))`)
	if err != nil {
		return err
	}
	// والمهام الملغية أو المحذوفة بعد ما انحسبت.
	_, err = r.db.Exec(`DELETE FROM "MatrixScore" m WHERE m.source = 'TASK' AND m."cancelledAt" IS NULL
		AND NOT EXISTS (SELECT 1 FROM "ExtraTask" t WHERE t.id = m."sourceId" AND t.status <> 'CANCELLED')`)
	return err
}

// SoleHolder الموظف الوحيد الفعّال بهالدور ("" إذا ماكو أو أكثر من واحد).
// المحطة الي ما اشتغل عليها أحد تنحسب على صاحب الدور إذا هو وحده المسؤول.
func (r *MatrixScoreRepository) SoleHolder(role, perm string) string {
	ids := []string{}
	_ = r.db.Select(&ids, `SELECT DISTINCT e.id FROM "Employee" e WHERE e.status = 'ACTIVE' AND e.role::text NOT IN ('ADMIN', 'OWNER')
		AND (e.role::text = $1 OR EXISTS (SELECT 1 FROM "EmployeePermission" ep JOIN "Permission" p ON p.id = ep."permissionId"
		     WHERE ep."employeeId" = e.id AND p.name = $2)) LIMIT 2`, role, perm)
	if len(ids) == 1 {
		return ids[0]
	}
	return ""
}

func (r *MatrixScoreRepository) Cancel(id, byID, note string) error {
	_, err := r.db.Exec(`UPDATE "MatrixScore" SET "cancelledAt" = now(), "cancelledById" = $2, "cancelNote" = $3
		WHERE id = $1 AND "cancelledAt" IS NULL`, id, byID, note)
	return err
}

// ═══ تقييمات تنتظر ═══

// PendingStaffRating حجز ينتظر تقييم من هالشخص بهالمحطة.
type PendingStaffRating struct {
	BookingID string `db:"bookingId" json:"bookingId"`
	Code      string `db:"code" json:"code"`
	RateeID   string `db:"rateeId" json:"rateeId"`
	RateeName string `db:"rateeName" json:"rateeName"`
	Role      string `db:"role" json:"role"` // LEADER | COORDINATOR
}

// PendingCoordLeader حجوزات رجعت منجزة بآخر ٧ أيام والإداري (الي ثبّتها) ما قيّم ليدرها.
func (r *MatrixScoreRepository) PendingCoordLeader(raterID string) ([]PendingStaffRating, error) {
	rows := []PendingStaffRating{}
	err := r.db.Select(&rows, `
		SELECT b.id AS "bookingId", b.code, l.id AS "rateeId", l.name AS "rateeName", 'LEADER' AS role
		FROM "Booking" b
		JOIN "Employee" l ON l.id = COALESCE(b."projectSupervisorId",
			(SELECT ba."employeeId" FROM "BookingAssignment" ba JOIN "Employee" x ON x.id = ba."employeeId"
			 WHERE ba."bookingId" = b.id AND x."isLeader" ORDER BY ba."createdAt" LIMIT 1))
		WHERE b.status::text IN ('COMPLETED', 'PARTIAL') AND l.role::text NOT IN ('ADMIN', 'OWNER')
		  AND b."updatedAt" >= now() - interval '7 days'
		  AND (b."confirmedByEmployeeId" = $1 OR b."createdById" = $1)
		  AND l.id <> $1
		  AND NOT EXISTS (SELECT 1 FROM "StaffRating" s WHERE s."raterId" = $1 AND s."bookingId" = b.id AND s.stage = 'COORD_LEADER')
		ORDER BY b."updatedAt" DESC LIMIT 20`, raterID)
	return rows, err
}

// BookingRateState الحجز خالص؟ والمقيّم هو مثبّته أو مسجّله؟
func (r *MatrixScoreRepository) BookingRateState(bookingID, raterID string) (done, handled bool, err error) {
	var x struct {
		Done    bool `db:"done"`
		Handled bool `db:"handled"`
	}
	err = r.db.Get(&x, `SELECT b.status::text IN ('COMPLETED', 'PARTIAL') AS done,
		(b."confirmedByEmployeeId" = $2 OR b."createdById" = $2) IS TRUE AS handled
		FROM "Booking" b WHERE b.id = $1`, bookingID, raterID)
	return x.Done, x.Handled, err
}

// BookingParties الليدر والإداري لحجز — للتقييم بالتدقيق واتصال الجودة.
func (r *MatrixScoreRepository) BookingParties(bookingID string) ([]PendingStaffRating, error) {
	rows := []PendingStaffRating{}
	err := r.db.Select(&rows, `
		SELECT b.id AS "bookingId", b.code, p.id AS "rateeId", p.name AS "rateeName", p.role
		FROM "Booking" b
		JOIN LATERAL (
			SELECT e.id, e.name, 'LEADER' AS role FROM "Employee" e
			WHERE e.id = COALESCE(b."projectSupervisorId",
				(SELECT ba."employeeId" FROM "BookingAssignment" ba JOIN "Employee" x ON x.id = ba."employeeId"
				 WHERE ba."bookingId" = b.id AND x."isLeader" ORDER BY ba."createdAt" LIMIT 1)) AND e.role::text NOT IN ('ADMIN', 'OWNER')
			UNION
			SELECT e.id, e.name, 'COORDINATOR' FROM "Employee" e WHERE e.id = COALESCE(b."confirmedByEmployeeId", b."createdById") AND e.role::text NOT IN ('ADMIN', 'OWNER')
		) p ON true
		WHERE b.id = $1`, bookingID)
	return rows, err
}

// RatedOn التقييمات الي سواها هالشخص على حجز بمحطة.
func (r *MatrixScoreRepository) RatedOn(raterID, bookingID, stage string) (map[string]int, error) {
	type row struct {
		RateeID string `db:"rateeId"`
		Score   int    `db:"score"`
	}
	rows := []row{}
	err := r.db.Select(&rows, `SELECT "rateeId", score FROM "StaffRating" WHERE "raterId" = $1 AND "bookingId" = $2 AND stage = $3`,
		raterID, bookingID, stage)
	out := map[string]int{}
	for _, x := range rows {
		out[x.RateeID] = x.Score
	}
	return out, err
}

// PeriodicRated تقييمات المراقب الدورية لهالفترة.
func (r *MatrixScoreRepository) PeriodicRated(raterID, period string) (map[string]int, error) {
	return r.StageRated(raterID, "MONITOR_PERIODIC", period)
}

// StageRated تقييمات دورية لهالمقيّم بمحطة وفترة.
func (r *MatrixScoreRepository) StageRated(raterID, stage, period string) (map[string]int, error) {
	type row struct {
		RateeID string `db:"rateeId"`
		Score   int    `db:"score"`
	}
	rows := []row{}
	err := r.db.Select(&rows, `SELECT "rateeId", score FROM "StaffRating" WHERE "raterId" = $1 AND stage = $2 AND period = $3`,
		raterID, stage, period)
	out := map[string]int{}
	for _, x := range rows {
		out[x.RateeID] = x.Score
	}
	return out, err
}

// PendingAudit حجوزات خلصت بآخر ٧ أيام والمراقب بعد ما قيّم ليدرها وإداريها.
func (r *MatrixScoreRepository) PendingAudit(raterID string) ([]PendingStaffRating, error) {
	rows := []PendingStaffRating{}
	err := r.db.Select(&rows, `
		SELECT b.id AS "bookingId", b.code, p.id AS "rateeId", p.name AS "rateeName", p.role
		FROM "Booking" b
		JOIN LATERAL (
			SELECT e.id, e.name, 'LEADER' AS role FROM "Employee" e
			WHERE e.id = COALESCE(b."projectSupervisorId",
				(SELECT ba."employeeId" FROM "BookingAssignment" ba JOIN "Employee" x ON x.id = ba."employeeId"
				 WHERE ba."bookingId" = b.id AND x."isLeader" ORDER BY ba."createdAt" LIMIT 1)) AND e.role::text NOT IN ('ADMIN', 'OWNER')
			UNION
			SELECT e.id, e.name, 'COORDINATOR' FROM "Employee" e WHERE e.id = COALESCE(b."confirmedByEmployeeId", b."createdById") AND e.role::text NOT IN ('ADMIN', 'OWNER')
		) p ON true
		WHERE b.status::text IN ('COMPLETED', 'PARTIAL')
		  AND b."updatedAt" >= now() - interval '7 days'
		  AND upper(b.code) NOT LIKE 'OLD%'
		  AND p.id <> $1
		  AND NOT EXISTS (SELECT 1 FROM "StaffRating" s WHERE s."raterId" = $1 AND s."bookingId" = b.id
		                  AND s.stage = 'AUDIT' AND s."rateeId" = p.id)
		ORDER BY b."updatedAt" DESC, p.role DESC LIMIT 60`, raterID)
	return rows, err
}

// ReliabilityFacts وقائع الاعتمادية بفترة: جلسات الحضور، الانصراف التلقائي، الشكاوى.
type ReliabilityFact struct {
	Sessions   int `db:"sessions"`
	Auto       int `db:"auto"`
	Complaints int `db:"complaints"`
	// قرار (ع) 10-08 — «مؤشرات أداء الموظف»: الإجازات، المشاكل ويا الزملاء،
	// العدّة (ضياع/تلف)، خصومات KPI. ومعلومات تنعرض بس (ما تدخل بالنسبة):
	// المهارات، إجازة السوق.
	LeaveDays   int  `db:"leaveDays"`
	Issues      int  `db:"issues"`
	ToolLosses  int  `db:"toolLosses"`
	KpiHits     int  `db:"kpiHits"`
	Skills      int  `db:"skills"`
	HasLicense  bool `db:"hasLicense"`
}

func (r *MatrixScoreRepository) ReliabilityFacts(from, to time.Time) (map[string]ReliabilityFact, error) {
	type row struct {
		ID string `db:"id"`
		ReliabilityFact
	}
	rows := []row{}
	err := r.db.Select(&rows, `SELECT e.id,
		(SELECT count(*) FROM "Attendance" a WHERE a."employeeId" = e.id AND a."checkIn" >= ($1::timestamptz AT TIME ZONE 'UTC') AND a."checkIn" < ($2::timestamptz AT TIME ZONE 'UTC') AND a."checkOut" IS NOT NULL)::int AS sessions,
		(SELECT count(*) FROM "AttendanceAuto" au JOIN "Attendance" a ON a.id = au."attendanceId"
		  WHERE au."employeeId" = e.id AND a."checkIn" >= ($1::timestamptz AT TIME ZONE 'UTC') AND a."checkIn" < ($2::timestamptz AT TIME ZONE 'UTC')
		    AND NOT EXISTS (SELECT 1 FROM "AttendanceClaim" c WHERE c."attendanceId" = a.id AND c.status IN ('OK','APPROVED') AND c.kind = 'WORKED'))::int AS auto,
		(SELECT count(*) FROM "Complaint" c WHERE c."relatedEmployeeId" = e.id AND c."createdAt" >= ($1::timestamptz AT TIME ZONE 'UTC') AND c."createdAt" < ($2::timestamptz AT TIME ZONE 'UTC'))::int AS complaints,
		(SELECT COALESCE(sum(GREATEST(0, LEAST(l."endDate", ($2::timestamptz AT TIME ZONE 'Asia/Baghdad')::date - 1) - GREATEST(l."startDate", ($1::timestamptz AT TIME ZONE 'Asia/Baghdad')::date) + 1)), 0)
		   FROM "LeaveRequest" l WHERE l."employeeId" = e.id AND l.status::text = 'APPROVED')::int AS "leaveDays",
		(SELECT count(*) FROM "WorkplaceIssue" wi WHERE (wi."partyAId" = e.id OR wi."partyBId" = e.id)
		   AND wi."createdAt" >= $1::timestamptz AND wi."createdAt" < $2::timestamptz)::int AS issues,
		(SELECT count(*) FROM "PersonalToolEvent" pe WHERE pe."employeeId" = e.id AND pe."toStatus" IN ('LOST','DAMAGED')
		   AND pe."createdAt" >= ($1::timestamptz AT TIME ZONE 'UTC') AND pe."createdAt" < ($2::timestamptz AT TIME ZONE 'UTC'))::int AS "toolLosses",
		(SELECT count(*) FROM "KpiEvaluation" k WHERE k."employeeId" = e.id AND k.points < 0
		   AND k."createdAt" >= ($1::timestamptz AT TIME ZONE 'UTC') AND k."createdAt" < ($2::timestamptz AT TIME ZONE 'UTC'))::int AS "kpiHits",
		(SELECT count(*) FROM "EmployeeSkill" es WHERE es."employeeId" = e.id AND es."canPerform")::int AS skills,
		e."hasDrivingLicense" AS "hasLicense"
		FROM "Employee" e WHERE e.status = 'ACTIVE'`, from, to)
	out := map[string]ReliabilityFact{}
	for _, x := range rows {
		out[x.ID] = x.ReliabilityFact
	}
	return out, err
}
