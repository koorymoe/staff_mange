package repository

import "time"

// ═══ «كل موظف يدخّل فلوس أكثر من راتبه» — دازاي 11-13 ═══
// الفلوس الي دخّلها كل موظف بفترة، من ثلاث مصادر (كل دور ياخذ مصدره بالخدمة):
//   - WORK: حصته من صافي فواتير الحجوزات الي اشتغلها (الفاتورة تنقسم على الكادر).
//   - BOOKED: صافي فواتير الحجوزات الي هو سجّلها (البياع والمنسّق).
//   - PROJECT: دفعات المشاريع الي هو مسؤولها.
// الفاتورة المسودة أو الملغية ما تنحسب.

type RevenueRow struct {
	EmployeeID string  `db:"employeeId"`
	Kind       string  `db:"kind"`
	Amount     float64 `db:"amount"`
	Jobs       int     `db:"jobs"`
}

// Revenue الفلوس الي دخّلها كل موظف بين from وto، مصنّفة بالمصدر.
func (r *MatrixScoreRepository) Revenue(from, to time.Time, emp string) ([]RevenueRow, error) {
	rows := []RevenueRow{}
	err := r.db.Select(&rows, `
		WITH inv AS (
			SELECT li."bookingId", li."employeeId", li."netTotal" AS n FROM "LeaderInvoice" li
			WHERE li."revokedAt" IS NULL AND li.status <> 'DRAFT' AND li."bookingId" IS NOT NULL
			  AND li."createdAt" >= $1 AND li."createdAt" < $2
		), crew AS (
			SELECT DISTINCT a."bookingId", a."employeeId" FROM "BookingAssignment" a WHERE a."bookingId" IN (SELECT "bookingId" FROM inv)
		), cnt AS (SELECT "bookingId", count(*) AS c FROM crew GROUP BY 1)
		SELECT * FROM (
			SELECT crew."employeeId", 'WORK' AS kind, sum(inv.n / cnt.c)::float8 AS amount, count(DISTINCT inv."bookingId")::int AS jobs
			FROM inv JOIN crew ON crew."bookingId" = inv."bookingId" JOIN cnt ON cnt."bookingId" = inv."bookingId" GROUP BY 1
			UNION ALL
			SELECT inv."employeeId", 'WORK', sum(inv.n)::float8, count(DISTINCT inv."bookingId")::int
			FROM inv WHERE NOT EXISTS (SELECT 1 FROM crew WHERE crew."bookingId" = inv."bookingId") GROUP BY 1
			UNION ALL
			SELECT b."createdById", 'BOOKED', sum(inv.n)::float8, count(DISTINCT inv."bookingId")::int
			FROM inv JOIN "Booking" b ON b.id = inv."bookingId" WHERE b."createdById" IS NOT NULL GROUP BY 1
			UNION ALL
			SELECT p."responsibleEmployeeId", 'PROJECT', sum(pp.amount)::float8, count(DISTINCT p.id)::int
			FROM "ProjectPayment" pp JOIN "Project" p ON p.id = pp."projectId"
			WHERE p."responsibleEmployeeId" IS NOT NULL AND COALESCE(pp."paidAt", pp."createdAt") >= $1 AND COALESCE(pp."paidAt", pp."createdAt") < $2
			GROUP BY 1
		) x WHERE "employeeId" IS NOT NULL AND ($3 = '' OR "employeeId" = $3)`, from, to, emp)
	return rows, err
}

// Salaries رواتب الموظفين (الي إلهم راتب مسجّل).
func (r *MatrixScoreRepository) Salaries(emp string) (map[string]float64, error) {
	rows := []struct {
		ID     string  `db:"id"`
		Salary float64 `db:"salary"`
	}{}
	if err := r.db.Select(&rows, `SELECT id, salary::float8 AS salary FROM "Employee" WHERE salary > 0 AND ($1 = '' OR id = $1)`, emp); err != nil {
		return nil, err
	}
	out := make(map[string]float64, len(rows))
	for _, x := range rows {
		out[x.ID] = x.Salary
	}
	return out, nil
}
