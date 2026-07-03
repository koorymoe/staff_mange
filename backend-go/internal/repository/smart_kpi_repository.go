package repository

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type SmartKpiRepository struct {
	db *sqlx.DB
}

func NewSmartKpiRepository(db *sqlx.DB) *SmartKpiRepository {
	return &SmartKpiRepository{db: db}
}

type employeeRoleRow struct {
	ID   string `db:"id"`
	Name string `db:"name"`
	Role string `db:"role"`
}

func (r *SmartKpiRepository) GetEmployeeForKpi(employeeID string) (*employeeRoleRow, error) {
	var e employeeRoleRow
	if err := r.db.Get(&e, `SELECT id, name, role FROM "Employee" WHERE id = $1`, employeeID); err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *SmartKpiRepository) ListActiveTechnicianIDs() ([]string, error) {
	var ids []string
	err := r.db.Select(&ids, `SELECT id FROM "Employee" WHERE role = 'TECHNICIAN' AND status = 'ACTIVE'`)
	return ids, err
}

func (r *SmartKpiRepository) CompletedBookingTimes(employeeID string, start, end time.Time) (createdAt, completedAt []time.Time, err error) {
	type row struct {
		CreatedAt   time.Time `db:"createdAt"`
		CompletedAt time.Time `db:"completedAt"`
	}
	var rows []row
	err = r.db.Select(&rows, `
		SELECT b."createdAt", b."completedAt"
		FROM "BookingAssignment" a
		JOIN "Booking" b ON b.id = a."bookingId"
		WHERE a."employeeId" = $1 AND b.status = 'COMPLETED' AND b."completedAt" >= $2 AND b."completedAt" < $3
	`, employeeID, start, end)
	if err != nil {
		return nil, nil, err
	}
	for _, rw := range rows {
		createdAt = append(createdAt, rw.CreatedAt)
		completedAt = append(completedAt, rw.CompletedAt)
	}
	return createdAt, completedAt, nil
}

func (r *SmartKpiRepository) MissionTimes(employeeID string, start, end time.Time) (assignedAt, completedAt []time.Time, err error) {
	type row struct {
		AssignedAt  time.Time `db:"assignedAt"`
		CompletedAt time.Time `db:"completedAt"`
	}
	var rows []row
	err = r.db.Select(&rows, `
		SELECT "assignedAt", "completedAt"
		FROM "Mission"
		WHERE "leaderId" = $1 AND stage = 'COMPLETED' AND "completedAt" >= $2 AND "completedAt" < $3
	`, employeeID, start, end)
	if err != nil {
		return nil, nil, err
	}
	for _, rw := range rows {
		assignedAt = append(assignedAt, rw.AssignedAt)
		completedAt = append(completedAt, rw.CompletedAt)
	}
	return assignedAt, completedAt, nil
}

func (r *SmartKpiRepository) WorkReportFlags(employeeID string, start, end time.Time) (full, total int, err error) {
	type row struct {
		CleanedSite bool `db:"cleanedSite"`
		GaveInfo    bool `db:"gaveInfo"`
		TookPhotos  bool `db:"tookPhotos"`
	}
	var rows []row
	err = r.db.Select(&rows, `
		SELECT "cleanedSite", "gaveInfo", "tookPhotos"
		FROM "WorkReport"
		WHERE "employeeId" = $1 AND "createdAt" >= $2 AND "createdAt" < $3
	`, employeeID, start, end)
	if err != nil {
		return 0, 0, err
	}
	for _, rw := range rows {
		if rw.CleanedSite && rw.GaveInfo && rw.TookPhotos {
			full++
		}
	}
	return full, len(rows), nil
}

func (r *SmartKpiRepository) AttendanceDaysPresent(employeeID string, start, end time.Time) (int, error) {
	var count int
	err := r.db.Get(&count, `
		SELECT COUNT(*) FROM "Attendance" WHERE "employeeId" = $1 AND date >= $2 AND date < $3
	`, employeeID, start, end)
	return count, err
}

func (r *SmartKpiRepository) ComplaintsCount(employeeID string, start, end time.Time) (int, error) {
	var count int
	err := r.db.Get(&count, `
		SELECT COUNT(*) FROM "Complaint" WHERE "assignedToEmployeeId" = $1 AND "createdAt" >= $2 AND "createdAt" < $3
	`, employeeID, start, end)
	return count, err
}

func (r *SmartKpiRepository) ManualDeductionPoints(employeeID string, start, end time.Time) (int, int, error) {
	var points []int
	err := r.db.Select(&points, `
		SELECT points FROM "KpiEvaluation" WHERE "employeeId" = $1 AND "createdAt" >= $2 AND "createdAt" < $3
	`, employeeID, start, end)
	if err != nil {
		return 0, 0, err
	}
	sum := 0
	for _, p := range points {
		sum += p
	}
	return sum, len(points), nil
}
