package repository

import (
	"time"

	"github.com/jmoiron/sqlx"
)

// سجل النشاط + الأفعال المسجّلة أصلاً بجداول النظام (منو اعتمد، منو قرر...)
type EmployeeActivityRepository struct{ db *sqlx.DB }

func NewEmployeeActivityRepository(db *sqlx.DB) *EmployeeActivityRepository {
	return &EmployeeActivityRepository{db: db}
}

func (r *EmployeeActivityRepository) Record(employeeID, method, pattern, path string, status int) {
	_, _ = r.db.Exec(`INSERT INTO "EmployeeActivity" ("employeeId", method, pattern, path, status) VALUES ($1,$2,$3,$4,$5)`,
		employeeID, method, pattern, path, status)
}

type ActivityRow struct {
	Method  string    `db:"method"`
	Pattern string    `db:"pattern"`
	Path    string    `db:"path"`
	At      time.Time `db:"createdAt"`
}

func (r *EmployeeActivityRepository) Logged(employeeID, day string) ([]ActivityRow, error) {
	rows := []ActivityRow{}
	err := r.db.Select(&rows, `SELECT method, pattern, path, "createdAt" FROM "EmployeeActivity"
		WHERE "employeeId" = $1 AND baghdad_date("createdAt") = $2::date ORDER BY "createdAt"`, employeeID, day)
	return rows, err
}

type KnownAction struct {
	Kind string    `db:"kind"`
	Ref  string    `db:"ref"`
	At   time.Time `db:"at"`
}

// Known أفعال اليوم من الجداول نفسها — تشتغل حتى قبل ما يبدي سجل النشاط.
func (r *EmployeeActivityRepository) Known(employeeID, day string) ([]KnownAction, error) {
	rows := []KnownAction{}
	err := r.db.Select(&rows, `
		SELECT * FROM (
			SELECT 'INV_APPROVE' AS kind, COALESCE(b.code, li."customerName", '') AS ref, li."approvedAt" AS at
			  FROM "LeaderInvoice" li LEFT JOIN "Booking" b ON b.id = li."bookingId"
			  WHERE li."approvedByEmployeeId" = $1 AND baghdad_date(li."approvedAt") = $2::date
			UNION ALL SELECT 'INV_AUDIT', COALESCE(b.code, li."customerName", ''), li."auditedAt"
			  FROM "LeaderInvoice" li LEFT JOIN "Booking" b ON b.id = li."bookingId"
			  WHERE li."auditedById" = $1 AND baghdad_date(li."auditedAt") = $2::date
			UNION ALL SELECT 'INV_MONITOR', COALESCE(b.code, li."customerName", ''), li."monitorDecidedAt"
			  FROM "LeaderInvoice" li LEFT JOIN "Booking" b ON b.id = li."bookingId"
			  WHERE li."monitorDecidedById" = $1 AND baghdad_date(li."monitorDecidedAt") = $2::date
			UNION ALL SELECT 'INV_RETURN', COALESCE(b.code, li."customerName", ''), li."returnedAt"
			  FROM "LeaderInvoice" li LEFT JOIN "Booking" b ON b.id = li."bookingId"
			  WHERE li."returnedById" = $1 AND baghdad_date(li."returnedAt") = $2::date
			UNION ALL SELECT 'INV_CREATE', COALESCE(b.code, li."customerName", ''), li."createdAt"
			  FROM "LeaderInvoice" li LEFT JOIN "Booking" b ON b.id = li."bookingId"
			  WHERE li."employeeId" = $1 AND baghdad_date(li."createdAt") = $2::date
			UNION ALL SELECT 'MONITOR_REVIEW', '', mr."reviewedAt" FROM "MonitorReview" mr
			  WHERE mr."reviewedById" = $1 AND baghdad_date(mr."reviewedAt") = $2::date
			UNION ALL SELECT 'DELETE_DECIDE', COALESCE(b.code, ''), d."decidedAt"
			  FROM "BookingDeleteRequest" d LEFT JOIN "Booking" b ON b.id = d."bookingId"
			  WHERE d."decidedById" = $1 AND baghdad_date(d."decidedAt") = $2::date
			UNION ALL SELECT 'DELETE_REQUEST', COALESCE(b.code, ''), d."createdAt"
			  FROM "BookingDeleteRequest" d LEFT JOIN "Booking" b ON b.id = d."bookingId"
			  WHERE d."requestedById" = $1 AND baghdad_date(d."createdAt") = $2::date
			UNION ALL SELECT 'LEAVE_DECIDE', COALESCE(e.name, ''), l."decidedAt"
			  FROM "LeaveRequest" l LEFT JOIN "Employee" e ON e.id = l."employeeId"
			  WHERE l."decidedById" = $1 AND baghdad_date(l."decidedAt") = $2::date
			UNION ALL SELECT 'BOOKING_CREATE', b.code, b."createdAt" FROM "Booking" b
			  WHERE b."createdById" = $1 AND baghdad_date(b."createdAt") = $2::date
			UNION ALL SELECT 'BOOKING_CONFIRM', b.code, b."confirmedAt" FROM "Booking" b
			  WHERE b."confirmedByEmployeeId" = $1 AND baghdad_date(b."confirmedAt") = $2::date
			UNION ALL SELECT 'MATERIALS_READY', b.code, b."materialsReadyAt" FROM "Booking" b
			  WHERE b."materialsReadyById" = $1 AND baghdad_date(b."materialsReadyAt") = $2::date
			UNION ALL SELECT 'AUDIT_ISSUE', COALESCE(b.code, ''), i."createdAt"
			  FROM "BookingAuditIssue" i LEFT JOIN "Booking" b ON b.id = i."bookingId"
			  WHERE i."raisedById" = $1 AND baghdad_date(i."createdAt") = $2::date
			UNION ALL SELECT 'QUALITY_CONTACT', '', q."contactedAt" FROM "QualityFollowUp" q
			  WHERE q."contactedByEmployeeId" = $1 AND baghdad_date(q."contactedAt") = $2::date
		) x WHERE at IS NOT NULL ORDER BY at`, employeeID, day)
	return rows, err
}

// Name / Code لتسمية الشي الي انعدّل («عدّل ملف فلان»، «حجز B-12»).
func (r *EmployeeActivityRepository) Lookup(table, col, id string) string {
	var v string
	_ = r.db.Get(&v, `SELECT COALESCE(`+col+`::text, '') FROM "`+table+`" WHERE id = $1`, id)
	return v
}
