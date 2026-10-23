package repository

import (
	"time"

	"github.com/jmoiron/sqlx"
)

// ═══ سجل الانضباط الوظيفي — طلب (ع) 10-05 ═══
// «سجل بي كل الموظفين ويا الخصومات الي انخصمت منهم من يوم الي نشا النظام…
// كم مرة مخصوم وشنو سبب الخصم والتاريخ وكل التفاصيل». يجمع ثلاث مصادر:
//   • KpiEvaluation — الخصم المالي والنقاط (وإذا انرجع: cancelledAt).
//   • DisciplineEvent — نقاط الانضباط (السالب خصم، الموجب استرجاع).
//   • EmployeeLedgerEntry — تسويات وشطب على عهدة الموظف النقدية.

type DisciplineRecordRepository struct{ db *sqlx.DB }

func NewDisciplineRecordRepository(db *sqlx.DB) *DisciplineRecordRepository {
	return &DisciplineRecordRepository{db: db}
}

type DisciplineEntry struct {
	Source       string     `db:"source" json:"source"` // KPI | POINTS | LEDGER
	ID           string     `db:"id" json:"id"`
	EmployeeID   string     `db:"employeeId" json:"employeeId"`
	EmployeeName string     `db:"employeeName" json:"employeeName"`
	Amount       float64    `db:"amount" json:"amount"`
	Points       int        `db:"points" json:"points"`
	Reason       string     `db:"reason" json:"reason"`
	Kind         *string    `db:"kind" json:"kind"`
	ByName       *string    `db:"byName" json:"byName"`
	BookingID    *string    `db:"bookingId" json:"bookingId"`
	BookingCode  *string    `db:"bookingCode" json:"bookingCode"`
	At           time.Time  `db:"at" json:"at"`
	Returned     bool       `db:"returned" json:"returned"`
	ReturnedAt   *time.Time `db:"returnedAt" json:"returnedAt"`
	ReturnNote   *string    `db:"returnNote" json:"returnNote"`
}

const disciplineRecordSQL = `
	SELECT 'KPI' AS source, k.id, k."employeeId", e.name AS "employeeName", COALESCE(k."deductionAmount", 0)::float8 AS amount,
	       COALESCE(k.points, 0) AS points, COALESCE(k.reason, '') AS reason, NULL::text AS kind, ev.name AS "byName",
	       q."bookingId", b.code AS "bookingCode", k."createdAt" AS at,
	       (COALESCE(k.cancelled, false) OR k."cancelledAt" IS NOT NULL) AS returned, k."cancelledAt" AS "returnedAt", k."cancelNote" AS "returnNote"
	FROM "KpiEvaluation" k JOIN "Employee" e ON e.id = k."employeeId"
	LEFT JOIN "Employee" ev ON ev.id = k."evaluatorId"
	LEFT JOIN "QualityFollowUp" q ON q."kpiEvaluationId" = k.id
	LEFT JOIN "Booking" b ON b.id = q."bookingId"
	UNION ALL
	SELECT 'POINTS', d.id, d."employeeId", e.name, 0, d.delta, COALESCE(d.reason, ''), d.kind, COALESCE(bx.name, 'النظام'),
	       d."bookingId", b.code, d."createdAt", false, NULL, NULL
	FROM "DisciplineEvent" d JOIN "Employee" e ON e.id = d."employeeId"
	LEFT JOIN "Employee" bx ON bx.id = d."byEmployeeId"
	LEFT JOIN "Booking" b ON b.id = d."bookingId"
	UNION ALL
	SELECT 'LEDGER', l.id, l."employeeId", e.name, l.amount::float8, 0, COALESCE(l.note, ''), l.kind, rb.name,
	       l."bookingId", b.code, l."occurredAt",
	       EXISTS (SELECT 1 FROM "EmployeeLedgerEntry" r WHERE r."reversesEntryId" = l.id), NULL, NULL
	FROM "EmployeeLedgerEntry" l JOIN "Employee" e ON e.id = l."employeeId"
	LEFT JOIN "Employee" rb ON rb.id = l."recordedById"
	LEFT JOIN "Booking" b ON b.id = l."bookingId"
	WHERE l.kind IN ('ADJUSTMENT', 'WRITE_OFF') AND l."reversesEntryId" IS NULL`

// All كل السجل (الأحدث أول) — اختيارياً لموظف واحد.
func (r *DisciplineRecordRepository) All(employeeID string) ([]DisciplineEntry, error) {
	rows := []DisciplineEntry{}
	q := `SELECT * FROM (` + disciplineRecordSQL + `) x`
	args := []any{}
	if employeeID != "" {
		q += ` WHERE "employeeId" = $1`
		args = append(args, employeeID)
	}
	err := r.db.Select(&rows, q+` ORDER BY at DESC`, args...)
	return rows, err
}
