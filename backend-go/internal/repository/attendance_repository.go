package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type AttendanceRepository struct {
	db *sqlx.DB
}

func NewAttendanceRepository(db *sqlx.DB) *AttendanceRepository {
	return &AttendanceRepository{db: db}
}

func (r *AttendanceRepository) loadEmployeeBrief(id string) *model.EmployeeBrief {
	var brief model.EmployeeBrief
	if err := r.db.Get(&brief, `SELECT id, name FROM "Employee" WHERE id = $1`, id); err != nil {
		return nil
	}
	return &brief
}

func (r *AttendanceRepository) hydrate(a *model.Attendance) {
	a.Employee = r.loadEmployeeBrief(a.EmployeeID)
}

func (r *AttendanceRepository) FindToday(employeeID string) (*model.Attendance, error) {
	var a model.Attendance
	err := r.db.Get(&a, `
		SELECT *, att_last_activity("employeeId", "checkIn", "checkOut") AS "lastActivity" FROM "Attendance"
		WHERE "employeeId" = $1 AND date = baghdad_today()
	`, employeeID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	r.hydrate(&a)
	return &a, nil
}

func (r *AttendanceRepository) FindOpenSession(employeeID string) (*model.Attendance, error) {
	var a model.Attendance
	err := r.db.Get(&a, `
		SELECT * FROM "Attendance"
		WHERE "employeeId" = $1 AND "checkOut" IS NULL
		ORDER BY "checkIn" DESC LIMIT 1
	`, employeeID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	r.hydrate(&a)
	return &a, nil
}

// TodaySessions ترجع كل جلسات حضور الموظف باليوم الحالي، مرتبة بوقت الدخول.
func (r *AttendanceRepository) TodaySessions(employeeID string) ([]model.Attendance, error) {
	records := []model.Attendance{}
	if err := r.db.Select(&records, `
		SELECT * FROM "Attendance"
		WHERE "employeeId" = $1 AND date = baghdad_today()
		ORDER BY "checkIn" ASC
	`, employeeID); err != nil {
		return nil, err
	}
	for i := range records {
		r.hydrate(&records[i])
	}
	return records, nil
}

func (r *AttendanceRepository) CheckIn(employeeID string) (*model.Attendance, error) {
	open, err := r.FindOpenSession(employeeID)
	if err != nil {
		return nil, err
	}
	if open != nil {
		return nil, errors.New("عندك تسجيل حضور مفتوح، سجل انصراف أول")
	}

	var a model.Attendance
	err = r.db.Get(&a, `
		INSERT INTO "Attendance" (id, "employeeId", "checkIn", date)
		VALUES (gen_random_uuid()::text, $1, now(), baghdad_today())
		RETURNING *
	`, employeeID)
	if err != nil {
		return nil, err
	}
	r.hydrate(&a)
	return &a, nil
}

func (r *AttendanceRepository) CheckOut(employeeID string) (*model.Attendance, error) {
	var a model.Attendance
	err := r.db.Get(&a, `
		UPDATE "Attendance" SET "checkOut" = now()
		WHERE id = (
			SELECT id FROM "Attendance"
			WHERE "employeeId" = $1 AND "checkOut" IS NULL
			ORDER BY "checkIn" DESC LIMIT 1
		)
		RETURNING *
	`, employeeID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("لا يوجد تسجيل حضور مفتوح اليوم")
	}
	if err != nil {
		return nil, err
	}
	r.hydrate(&a)
	return &a, nil
}

func (r *AttendanceRepository) Today() ([]model.Attendance, error) {
	records := []model.Attendance{}
	if err := r.db.Select(&records, `
		SELECT * FROM "Attendance" WHERE date = baghdad_today() ORDER BY "checkIn" ASC
	`); err != nil {
		return nil, err
	}
	for i := range records {
		r.hydrate(&records[i])
	}
	return records, nil
}

// TodaySummary ترجع ملخص حضور كل موظف عنده جلسة (أو أكثر) باليوم الحالي —
// مجمّعة بـ GROUP BY لتفادي N+1، وتستخدم بجدول المراقب.
func (r *AttendanceRepository) TodaySummary() ([]model.EmployeeDailyAttendanceSummary, error) {
	return r.daySummary("baghdad_today()")
}

// DaySummary نفس TodaySummary لكن بتاريخ محدد (لدعم ?date= بتصدير الإكسل).
func (r *AttendanceRepository) DaySummary(date string) ([]model.EmployeeDailyAttendanceSummary, error) {
	rows := []model.EmployeeDailyAttendanceSummary{}
	if err := r.db.Select(&rows, `
		SELECT
			"employeeId",
			COUNT(*)::int AS "sessionsCount",
			MIN("checkIn") AS "firstCheckIn",
			CASE WHEN bool_or("checkOut" IS NULL) THEN NULL ELSE MAX("checkOut") END AS "lastCheckOut",
			bool_or("checkOut" IS NULL) AS "currentlyActive",
			SUM(att_counted_minutes("employeeId", "checkIn", "checkOut"))::int AS "totalMinutes"
		FROM "Attendance"
		WHERE date = $1::date
		GROUP BY "employeeId"
		ORDER BY MIN("checkIn") ASC
	`, date); err != nil {
		return nil, err
	}
	for i := range rows {
		rows[i].Employee = r.loadEmployeeBrief(rows[i].EmployeeID)
	}
	return rows, nil
}

func (r *AttendanceRepository) daySummary(dateExpr string) ([]model.EmployeeDailyAttendanceSummary, error) {
	rows := []model.EmployeeDailyAttendanceSummary{}
	if err := r.db.Select(&rows, `
		SELECT
			"employeeId",
			COUNT(*)::int AS "sessionsCount",
			MIN("checkIn") AS "firstCheckIn",
			CASE WHEN bool_or("checkOut" IS NULL) THEN NULL ELSE MAX("checkOut") END AS "lastCheckOut",
			bool_or("checkOut" IS NULL) AS "currentlyActive",
			SUM(att_counted_minutes("employeeId", "checkIn", "checkOut"))::int AS "totalMinutes"
		FROM "Attendance"
		WHERE date = `+dateExpr+`
		GROUP BY "employeeId"
		ORDER BY MIN("checkIn") ASC
	`); err != nil {
		return nil, err
	}
	for i := range rows {
		rows[i].Employee = r.loadEmployeeBrief(rows[i].EmployeeID)
	}
	return rows, nil
}

func (r *AttendanceRepository) ForEmployeeInRange(employeeID string, from, to string) ([]model.Attendance, error) {
	records := []model.Attendance{}
	if err := r.db.Select(&records, `
		SELECT *, att_last_activity("employeeId", "checkIn", "checkOut") AS "lastActivity" FROM "Attendance"
		WHERE "employeeId" = $1 AND date >= $2::date AND date < $3::date
		ORDER BY date ASC
	`, employeeID, from, to); err != nil {
		return nil, err
	}
	for i := range records {
		r.hydrate(&records[i])
	}
	return records, nil
}

func (r *AttendanceRepository) Correct(id string, checkIn, checkOut *time.Time) (*model.Attendance, error) {
	var a model.Attendance
	err := r.db.Get(&a, `
		UPDATE "Attendance" SET
			"checkIn" = COALESCE($2, "checkIn"),
			"checkOut" = COALESCE($3, "checkOut")
		WHERE id = $1
		RETURNING *
	`, id, checkIn, checkOut)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("سجل الحضور غير موجود")
	}
	if err != nil {
		return nil, err
	}
	r.hydrate(&a)
	return &a, nil
}

// ═══ الحضور الإجباري + الانصراف التلقائي ═══

// GateInfo شفت الموظف وحالته اليوم.
type GateInfo struct {
	Role       string  `db:"role"`
	Shift      *string `db:"shift"`
	ShiftStart *string `db:"shiftStart"`
	ShiftEnd   *string `db:"shiftEnd"`
	HasOpen    bool    `db:"hasOpen"`
	HadToday   bool    `db:"hadToday"`
	OnLeave    bool    `db:"onLeave"`
}

func (r *AttendanceRepository) Gate(employeeID string) (*GateInfo, error) {
	var g GateInfo
	err := r.db.Get(&g, `
		SELECT e.role::text AS role, e.shift::text AS shift, e."shiftStart", e."shiftEnd",
		       EXISTS (SELECT 1 FROM "Attendance" a WHERE a."employeeId" = e.id AND a."checkOut" IS NULL) AS "hasOpen",
		       EXISTS (SELECT 1 FROM "Attendance" a WHERE a."employeeId" = e.id AND a.date = baghdad_today()) AS "hadToday",
		       EXISTS (SELECT 1 FROM "LeaveRequest" l WHERE l."employeeId" = e.id AND l.status = 'APPROVED'
		               AND baghdad_today()::date BETWEEN l."startDate" AND l."endDate") AS "onLeave"
		FROM "Employee" e WHERE e.id = $1`, employeeID)
	return &g, err
}

// OpenWithShift الجلسات المفتوحة ويا شفت أصحابها وآخر حجز خلّصوه بعد الدخول.
type OpenWithShift struct {
	ID           string     `db:"id"`
	EmployeeID   string     `db:"employeeId"`
	Name         string     `db:"name"`
	CheckIn      time.Time  `db:"checkIn"`
	Shift        *string    `db:"shift"`
	ShiftStart   *string    `db:"shiftStart"`
	ShiftEnd     *string    `db:"shiftEnd"`
	LastActivity *time.Time `db:"lastActivity"`
	Busy         bool       `db:"busy"`
}

func (r *AttendanceRepository) OpenWithShift() ([]OpenWithShift, error) {
	rows := []OpenWithShift{}
	err := r.db.Select(&rows, `
		SELECT a.id, a."employeeId", e.name, a."checkIn", e.shift::text AS shift, e."shiftStart", e."shiftEnd",
		       att_last_activity(a."employeeId", a."checkIn", NULL) AS "lastActivity",
		       EXISTS (SELECT 1 FROM "BookingAssignment" ba JOIN "Booking" b ON b.id = ba."bookingId"
		        WHERE ba."employeeId" = a."employeeId" AND b.status = 'IN_PROGRESS'
		          AND b."updatedAt" >= a."checkIn") AS busy
		FROM "Attendance" a JOIN "Employee" e ON e.id = a."employeeId"
		WHERE a."checkOut" IS NULL`)
	return rows, err
}

// AutoClose يسكّر جلسة بوقت محدد ويعلّمها «انصراف تلقائي».
func (r *AttendanceRepository) AutoClose(id, employeeID string, at time.Time, reason string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	res, err := tx.Exec(`UPDATE "Attendance" SET "checkOut" = ($2::timestamptz AT TIME ZONE 'UTC') WHERE id = $1 AND "checkOut" IS NULL`, id, at)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil
	}
	if _, err := tx.Exec(`INSERT INTO "AttendanceAuto" ("attendanceId", "employeeId", reason) VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING`, id, employeeID, reason); err != nil {
		return err
	}
	return tx.Commit()
}

// AutoCountSince كم مرة انسكّر انصرافه تلقائياً من تاريخ.
func (r *AttendanceRepository) AutoCountSince(since time.Time) (map[string]int, error) {
	type row struct {
		EmployeeID string `db:"employeeId"`
		N          int    `db:"n"`
	}
	rows := []row{}
	err := r.db.Select(&rows, `SELECT "employeeId", count(*)::int AS n FROM "AttendanceAuto" WHERE "createdAt" >= $1 GROUP BY 1`, since)
	out := map[string]int{}
	for _, x := range rows {
		out[x.EmployeeID] = x.N
	}
	return out, err
}

// ═══ ما بعد الانصراف التلقائي: سؤال الموظف والدليل ═══

// AutoClosedOpen آخر جلسة انسكّرت تلقائياً خلال ٢٠ ساعة وماكو جواب عليها.
type AutoClosedSession struct {
	ID       string    `db:"id"`
	CheckIn  time.Time `db:"checkIn"`
	CheckOut time.Time `db:"checkOut"`
}

func (r *AttendanceRepository) AutoClosedUnanswered(employeeID string) (*AutoClosedSession, error) {
	var s AutoClosedSession
	err := r.db.Get(&s, `SELECT a.id, a."checkIn", a."checkOut" FROM "Attendance" a
		JOIN "AttendanceAuto" au ON au."attendanceId" = a.id
		WHERE a."employeeId" = $1 AND au."createdAt" > now() - interval '20 hours'
		  AND NOT EXISTS (SELECT 1 FROM "AttendanceClaim" c WHERE c."attendanceId" = a.id)
		ORDER BY a."checkOut" DESC LIMIT 1`, employeeID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &s, err
}

// WorkEvidence آخر دليل شغل للموظف بين وقتين: حجز خلّصه، فاتورة، تقرير، مهمة.
type WorkEvidence struct {
	At    time.Time `db:"at"`
	Label string    `db:"label"`
}

func (r *AttendanceRepository) LastEvidence(employeeID string, from, to time.Time) (*WorkEvidence, error) {
	var e WorkEvidence
	err := r.db.Get(&e, `SELECT at, label FROM (
		SELECT (b."completedAt" AT TIME ZONE 'UTC') AS at, 'خلّص حجز ' || b.code AS label
		  FROM "Booking" b JOIN "BookingAssignment" ba ON ba."bookingId" = b.id
		  WHERE ba."employeeId" = $1 AND b."completedAt" IS NOT NULL
		UNION ALL
		SELECT (li."createdAt" AT TIME ZONE 'UTC'), 'رفع فاتورة حجز ' || b.code
		  FROM "LeaderInvoice" li JOIN "Booking" b ON b.id = li."bookingId" WHERE li."employeeId" = $1
		UNION ALL
		SELECT (w."createdAt" AT TIME ZONE 'UTC'), 'كتب تقرير حجز ' || b.code
		  FROM "WorkReport" w JOIN "Booking" b ON b.id = w."bookingId" WHERE w."employeeId" = $1
		UNION ALL
		SELECT t."doneAt", 'خلّص مهمة «' || t.title || '»' FROM "ExtraTask" t WHERE t."assignedToId" = $1 AND t."doneAt" IS NOT NULL
	) x WHERE at > $2 AND at <= $3 ORDER BY at DESC LIMIT 1`, employeeID, from, to)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &e, err
}

type AttendanceClaimIn struct {
	AttendanceID string
	EmployeeID   string
	Kind         string
	AutoAt       time.Time
	ClaimedUntil *time.Time
	EvidenceAt   *time.Time
	Evidence     *string
	Note         *string
	Status       string
}

func (r *AttendanceRepository) SaveClaim(c AttendanceClaimIn) error {
	_, err := r.db.Exec(`INSERT INTO "AttendanceClaim" ("attendanceId", "employeeId", kind, "autoAt", "claimedUntil", "evidenceAt", evidence, note, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''), $9) ON CONFLICT ("attendanceId") DO NOTHING`,
		c.AttendanceID, c.EmployeeID, c.Kind, c.AutoAt, c.ClaimedUntil, c.EvidenceAt, c.Evidence, c.Note, c.Status)
	return err
}

// SetCheckOut يصحّح وقت الانصراف (بدليل أو بقرار المراقب).
func (r *AttendanceRepository) SetCheckOut(id string, at time.Time) error {
	_, err := r.db.Exec(`UPDATE "Attendance" SET "checkOut" = ($2::timestamptz AT TIME ZONE 'UTC') WHERE id = $1`, id, at)
	return err
}

type AttendanceClaimRow struct {
	ID           string     `db:"id" json:"id"`
	AttendanceID string     `db:"attendanceId" json:"attendanceId"`
	EmployeeID   string     `db:"employeeId" json:"employeeId"`
	Name         string     `db:"name" json:"name"`
	Kind         string     `db:"kind" json:"kind"`
	AutoAt       time.Time  `db:"autoAt" json:"autoAt"`
	ClaimedUntil *time.Time `db:"claimedUntil" json:"claimedUntil"`
	EvidenceAt   *time.Time `db:"evidenceAt" json:"evidenceAt"`
	Evidence     *string    `db:"evidence" json:"evidence"`
	Note         *string    `db:"note" json:"note"`
	Status       string     `db:"status" json:"status"`
	CreatedAt    time.Time  `db:"createdAt" json:"createdAt"`
}

func (r *AttendanceRepository) Claims(status string) ([]AttendanceClaimRow, error) {
	rows := []AttendanceClaimRow{}
	err := r.db.Select(&rows, `SELECT c.id, c."attendanceId", c."employeeId", e.name, c.kind, c."autoAt", c."claimedUntil",
		c."evidenceAt", c.evidence, c.note, c.status, c."createdAt"
		FROM "AttendanceClaim" c JOIN "Employee" e ON e.id = c."employeeId"
		WHERE ($1 = '' OR c.status = $1) ORDER BY c."createdAt" DESC LIMIT 200`, status)
	return rows, err
}

func (r *AttendanceRepository) Claim(id string) (*AttendanceClaimRow, error) {
	var c AttendanceClaimRow
	err := r.db.Get(&c, `SELECT c.id, c."attendanceId", c."employeeId", e.name, c.kind, c."autoAt", c."claimedUntil",
		c."evidenceAt", c.evidence, c.note, c.status, c."createdAt"
		FROM "AttendanceClaim" c JOIN "Employee" e ON e.id = c."employeeId" WHERE c.id = $1`, id)
	return &c, err
}

func (r *AttendanceRepository) DecideClaim(id, byID, status string) error {
	_, err := r.db.Exec(`UPDATE "AttendanceClaim" SET status = $3, "decidedById" = $2, "decidedAt" = now()
		WHERE id = $1 AND status = 'PENDING'`, id, byID, status)
	return err
}

// ExtendableAuto جلسات انسكّرت تلقائياً بآخر ٢٠ ساعة وماكو جواب — حتى إذا
// خلّص شغل بعدها، الانصراف يتمدّد لحاله لحد الدليل.
func (r *AttendanceRepository) ExtendableAuto() ([]struct {
	ID         string    `db:"id"`
	EmployeeID string    `db:"employeeId"`
	CheckOut   time.Time `db:"checkOut"`
}, error) {
	rows := []struct {
		ID         string    `db:"id"`
		EmployeeID string    `db:"employeeId"`
		CheckOut   time.Time `db:"checkOut"`
	}{}
	err := r.db.Select(&rows, `SELECT a.id, a."employeeId", (a."checkOut" AT TIME ZONE 'UTC') AS "checkOut" FROM "Attendance" a
		JOIN "AttendanceAuto" au ON au."attendanceId" = a.id
		WHERE au."createdAt" > now() - interval '20 hours'
		  AND NOT EXISTS (SELECT 1 FROM "AttendanceClaim" c WHERE c."attendanceId" = a.id)
		  AND NOT EXISTS (SELECT 1 FROM "Attendance" o WHERE o."employeeId" = a."employeeId" AND o."checkOut" IS NULL)`)
	return rows, err
}

func (r *AttendanceRepository) NoteAuto(id, reason string) {
	_, _ = r.db.Exec(`UPDATE "AttendanceAuto" SET reason = $2 WHERE "attendanceId" = $1`, id, reason)
}
