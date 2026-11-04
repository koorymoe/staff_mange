package repository

import (
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// ═══ تسليم شغل موظف لموظف — (ع) 10-17 ═══
// «من يطلع أحد الموظفين، ما نسوي حساب جديد ونكعد ننطي صلاحيات — نختار موظف
// ويترحّل كلشي إله». المنصب والصلاحيات والشغل المفتوح ينتقلون للمستلم،
// والقديم يتوقّف حسابه (ما ينحذف — تاريخه وفواتيره تبقى باسمه).

type HandoverSummary struct {
	Position    string `json:"position"`
	Permissions int64  `json:"permissions"`
	Bookings    int64  `json:"bookings"`
	Missions    int64  `json:"missions"`
	Projects    int64  `json:"projects"`
	Complaints  int64  `json:"complaints"`
	ExtraTasks  int64  `json:"extraTasks"`
	DeptHead    int64  `json:"deptHead"`
}

type EmployeeHandoverRepository struct{ db *sqlx.DB }

func NewEmployeeHandoverRepository(db *sqlx.DB) *EmployeeHandoverRepository {
	return &EmployeeHandoverRepository{db: db}
}

const openBooking = `(SELECT id FROM "Booking" WHERE status::text NOT IN ('COMPLETED','CANCELLED'))`

// Preview شنو راح ينتقل — بلا تغيير (لشاشة التأكيد).
func (r *EmployeeHandoverRepository) Preview(fromID string) (HandoverSummary, error) {
	var s HandoverSummary
	err := r.db.Get(&s.Position, `SELECT role::text FROM "Employee" WHERE id = $1`, fromID)
	if err != nil {
		return s, errors.New("الموظف مو موجود")
	}
	q := func(dst *int64, sql string) { _ = r.db.Get(dst, sql, fromID) }
	q(&s.Permissions, `SELECT count(*) FROM "EmployeePermission" WHERE "employeeId" = $1`)
	q(&s.Bookings, `SELECT count(*) FROM "BookingAssignment" WHERE "employeeId" = $1 AND "bookingId" IN `+openBooking)
	q(&s.Missions, `SELECT count(*) FROM "Mission" WHERE "completedAt" IS NULL AND ("leaderId" = $1 OR $1 = ANY("memberIds"))`)
	q(&s.Projects, `SELECT count(*) FROM "Booking" WHERE "projectSupervisorId" = $1 AND id IN `+openBooking)
	q(&s.Complaints, `SELECT count(*) FROM "Complaint" WHERE "assignedToEmployeeId" = $1 AND status::text IN ('NEW','IN_PROGRESS')`)
	q(&s.ExtraTasks, `SELECT count(*) FROM "ExtraTask" WHERE "assignedToId" = $1 AND status NOT IN ('DONE','CANCELLED')`)
	q(&s.DeptHead, `SELECT count(*) FROM "DepartmentHead" WHERE "employeeId" = $1 AND active`)
	return s, nil
}

// Handover ينقل كلشي من fromID لـtoID بمعاملة وحدة، ويوقف حساب fromID.
func (r *EmployeeHandoverRepository) Handover(fromID, toID, byID string) (HandoverSummary, error) {
	var s HandoverSummary
	if fromID == toID {
		return s, errors.New("اختار موظف ثاني يستلم")
	}
	tx, err := r.db.Beginx()
	if err != nil {
		return s, err
	}
	defer func() { _ = tx.Rollback() }()

	var role string
	if err := tx.Get(&role, `SELECT role::text FROM "Employee" WHERE id = $1 AND role::text NOT IN ('ADMIN','OWNER')`, fromID); err != nil {
		return s, errors.New("الموظف مو موجود، أو حساب المدير/المالك ما يتسلّم")
	}
	s.Position = role
	var ok bool
	if err := tx.Get(&ok, `SELECT true FROM "Employee" WHERE id = $1 AND status = 'ACTIVE' AND role::text NOT IN ('ADMIN','OWNER')`, toID); err != nil {
		return s, errors.New("المستلم لازم يكون موظف فعّال")
	}
	exec := func(dst *int64, sql string, args ...any) error {
		res, err := tx.Exec(sql, args...)
		if err != nil {
			return err
		}
		if dst != nil {
			*dst, _ = res.RowsAffected()
		}
		return nil
	}
	steps := []struct {
		dst *int64
		sql string
	}{
		// ١) المنصب كامل
		{nil, `UPDATE "Employee" t SET role = f.role, "secondaryRoles" = f."secondaryRoles", "isLeader" = f."isLeader",
			"leaderSkillLevel" = f."leaderSkillLevel", division = f.division, department = f.department, position = f.position,
			"jobTitle" = f."jobTitle", "jobLevel" = f."jobLevel", shift = f.shift, "shiftStart" = f."shiftStart", "shiftEnd" = f."shiftEnd"
			FROM "Employee" f WHERE t.id = $2 AND f.id = $1`},
		// ٢) الصلاحيات ونطاق الإدارة
		{&s.Permissions, `INSERT INTO "EmployeePermission" (id, "employeeId", "permissionId")
			SELECT gen_random_uuid()::text, $2, "permissionId" FROM "EmployeePermission" WHERE "employeeId" = $1
			ON CONFLICT ("employeeId", "permissionId") DO NOTHING`},
		{nil, `INSERT INTO "EmployeeAdminScope" ("employeeId", roles, "updatedById", "updatedAt")
			SELECT $2, roles, NULL, now() FROM "EmployeeAdminScope" WHERE "employeeId" = $1
			ON CONFLICT ("employeeId") DO UPDATE SET roles = EXCLUDED.roles, "updatedAt" = now()`},
		// ٣) الشغل المفتوح
		{&s.Bookings, `UPDATE "BookingAssignment" a SET "employeeId" = $2 WHERE a."employeeId" = $1 AND a."bookingId" IN ` + openBooking + `
			AND NOT EXISTS (SELECT 1 FROM "BookingAssignment" x WHERE x."bookingId" = a."bookingId" AND x."employeeId" = $2)`},
		{&s.Missions, `UPDATE "Mission" SET "leaderId" = CASE WHEN "leaderId" = $1 THEN $2 ELSE "leaderId" END,
			"memberIds" = array_replace("memberIds", $1, $2), "updatedAt" = now()
			WHERE "completedAt" IS NULL AND ("leaderId" = $1 OR $1 = ANY("memberIds"))`},
		{&s.Projects, `UPDATE "Booking" SET "projectSupervisorId" = $2 WHERE "projectSupervisorId" = $1 AND id IN ` + openBooking},
		{&s.Complaints, `UPDATE "Complaint" SET "assignedToEmployeeId" = $2 WHERE "assignedToEmployeeId" = $1 AND status::text IN ('NEW','IN_PROGRESS')`},
		{&s.ExtraTasks, `UPDATE "ExtraTask" SET "assignedToId" = $2, "updatedAt" = now() WHERE "assignedToId" = $1 AND status NOT IN ('DONE','CANCELLED')`},
		{&s.DeptHead, `UPDATE "DepartmentHead" SET "employeeId" = $2 WHERE "employeeId" = $1 AND active`},
	}
	for _, st := range steps {
		if err := exec(st.dst, st.sql, fromID, toID); err != nil {
			return s, err
		}
	}
	// ٤) القديم يتوقّف — ما ينحذف
	if err := exec(nil, `UPDATE "Employee" SET status = 'ARCHIVED', "onDuty" = false, "suspendedReason" = 'سلّم شغله لموظف جديد',
		"suspendedAt" = now(), "suspendedById" = NULLIF($2, ''), "sessionsInvalidatedAt" = now() WHERE id = $1`, fromID, byID); err != nil {
		return s, err
	}
	// ٥) السجل
	sum, _ := json.Marshal(s)
	if err := exec(nil, `INSERT INTO "EmployeeHandover" (id, "fromId", "toId", "byId", summary) VALUES ($1, $2, $3, NULLIF($4, ''), $5)`,
		uuid.NewString(), fromID, toID, byID, sum); err != nil {
		return s, err
	}
	return s, tx.Commit()
}
