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
		SELECT * FROM "Attendance"
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
			SUM(
				EXTRACT(EPOCH FROM (COALESCE("checkOut", now()) - "checkIn")) / 60
			)::int AS "totalMinutes"
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
			SUM(
				EXTRACT(EPOCH FROM (COALESCE("checkOut", now()) - "checkIn")) / 60
			)::int AS "totalMinutes"
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
		SELECT * FROM "Attendance"
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
