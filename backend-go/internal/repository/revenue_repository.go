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
	ms := LoadMoneySplit(r.db)
	err := r.db.Select(&rows, `
		WITH inv AS (
			SELECT li.id, li."bookingId", li."employeeId", li."netTotal" AS n FROM "LeaderInvoice" li
			WHERE li."revokedAt" IS NULL AND li.status <> 'DRAFT'
			  AND li."createdAt" >= $1 AND li."createdAt" < $2
		),`+workCrewSQL+`
		SELECT * FROM (
			-- (ل) 11-24 دازاي: الليدر نسبته، والفنيين يتقاسمون نسبتهم؛ بلا ليدر الفنيين ياخذون نسبة «بلا ليدر»
			SELECT c."employeeId", 'WORK' AS kind,
				sum(inv.n * CASE WHEN c.lead THEN $4::float8 WHEN cnt.has_lead THEN $5::float8 / cnt.techs ELSE $6::float8 / cnt.techs END / 100.0)::float8 AS amount,
				count(DISTINCT inv.id)::int AS jobs
			FROM inv JOIN crew2 c ON c.inv = inv.id JOIN cnt ON cnt.inv = inv.id GROUP BY 1
			UNION ALL
			-- فاتورة بلا كادر: الي فوترها ياخذ نسبة الليدر إذا هو ليدر (مسؤول الخدمة ما ياخذ شي)
			SELECT inv."employeeId", 'WORK', sum(inv.n * $4::float8 / 100.0)::float8, count(DISTINCT inv.id)::int
			FROM inv JOIN "Employee" e ON e.id = inv."employeeId"
			WHERE e."isLeader" AND e.role::text <> 'SERVICE_MANAGER' AND NOT EXISTS (SELECT 1 FROM crew2 WHERE crew2.inv = inv.id) GROUP BY 1
			UNION ALL
			-- (ل) 11-24: تصحيح الخطأ الفني (بموافقة المراقب) — ينضاف للي صلّح وينخصم من الكادر الأول
			SELECT ra."employeeId", 'WORK', sum(ra.amount)::float8, 0
			FROM "ReworkAdjustment" ra WHERE ra."createdAt" >= $1 AND ra."createdAt" < $2 GROUP BY 1
			UNION ALL
			SELECT b."createdById", 'BOOKED', sum(inv.n)::float8, count(DISTINCT inv."bookingId")::int
			FROM inv JOIN "Booking" b ON b.id = inv."bookingId" WHERE b."createdById" IS NOT NULL GROUP BY 1
			UNION ALL
			SELECT p."responsibleEmployeeId", 'PROJECT', sum(pp.amount)::float8, count(DISTINCT p.id)::int
			FROM "ProjectPayment" pp JOIN "Project" p ON p.id = pp."projectId"
			WHERE p."responsibleEmployeeId" IS NOT NULL AND pp."cancelledAt" IS NULL AND COALESCE(pp."paidAt", pp."createdAt") >= $1 AND COALESCE(pp."paidAt", pp."createdAt") < $2
			GROUP BY 1
			UNION ALL
			-- (دازاي 11-28) تجديد اشتراكات الـGPS المعتمد — ينحسب للي سجّل التجديد
			SELECT g."employeeId", 'GPS', sum(g.price)::float8, count(*)::int
			FROM "GpsRenewalRequest" g
			WHERE g.status::text = 'APPROVED' AND g.price > 0 AND g."createdAt" >= $1 AND g."createdAt" < $2
			GROUP BY 1
		) x WHERE "employeeId" IS NOT NULL AND ($3 = '' OR "employeeId" = $3)`, from, to, emp,
		ms.StatsLeaderPct, ms.StatsTechPct, ms.StatsNoLeaderTechPct)
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
