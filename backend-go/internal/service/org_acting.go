package service

import (
	"fmt"
	"log"
	"time"
)

// ═══ 🏛️ المناصب المؤقتة بالإجازات (المبدأ ٢٠ — دازاي 10-30) ═══
// «عند الإجازات تتغير المناصب مؤقتاً وفقاً لمصلحة العمل وحسب رؤية المدير».
// كل ساعة: مدير (إدارة أو وحدة) بإجازة موافق عليها تشمل اليوم ← نائبه نفس المكان ياخذ
// «مدير بالوكالة» (acting) ويا صلاحياته لحد آخر يوم بالإجازة. الوكالة تنتهي لحالها من
// تخلص الإجازة أو تنلغى. ماكو نائب ← تنبيه لمدير الإدارة والمالك يعيّنون بديل.
// ما نلمس منطق الإجازات نفسه — نقرا الطلبات الموافق عليها بس.

func (s *OrgService) RunActing(notify func(empID, msg string)) {
	today := time.Now().In(baghdad).Format("2006-01-02")
	type absent struct {
		AssignmentID string  `db:"id"`
		Employee     string  `db:"employeeId"`
		Name         string  `db:"name"`
		DivisionID   string  `db:"divisionId"`
		UnitID       *string `db:"unitId"`
		Until        string  `db:"until"`
		Deputy       *string `db:"deputyId"`
	}
	rows := []absent{}
	if err := s.db.Select(&rows, `
		SELECT a.id, a."employeeId", e.name, a."divisionId", a."unitId", max(l."endDate")::text AS until,
		  (SELECT d."employeeId" FROM "OrgAssignment" d WHERE d."divisionId" = a."divisionId"
		     AND d."unitId" IS NOT DISTINCT FROM a."unitId" AND d.position = 'DEPUTY' AND d."endedAt" IS NULL AND NOT d.acting
		     AND NOT EXISTS (SELECT 1 FROM "LeaveRequest" l2 WHERE l2."employeeId" = d."employeeId" AND l2.status = 'APPROVED'
		                     AND l2."startDate" <= $1::date AND l2."endDate" >= $1::date)
		   LIMIT 1) AS "deputyId"
		FROM "OrgAssignment" a
		JOIN "Employee" e ON e.id = a."employeeId"
		JOIN "LeaveRequest" l ON l."employeeId" = a."employeeId" AND l.status = 'APPROVED'
		  AND l."startDate" <= $1::date AND l."endDate" >= $1::date
		WHERE a.position = 'DIRECTOR' AND a."endedAt" IS NULL AND NOT a.acting
		GROUP BY a.id, a."employeeId", e.name, a."divisionId", a."unitId"`, today); err != nil {
		log.Printf("[org-acting] %v", err)
		return
	}
	for _, r := range rows {
		var exists int
		_ = s.db.Get(&exists, `SELECT count(*) FROM "OrgAssignment" WHERE "replacesId" = $1 AND acting AND "endedAt" IS NULL`, r.AssignmentID)
		if exists > 0 {
			continue
		}
		place := s.placeName(r.DivisionID, r.UnitID)
		if r.Deputy == nil {
			// ماكو نائب متوفر — مرة وحدة باليوم لكل مدير
			if ok := s.claimDaily("ACTING_NONE:"+r.AssignmentID, today); ok && notify != nil {
				msg := fmt.Sprintf("🏛️ %s (مدير %s) بإجازة لحد %s وماكو نائب يمسك مكانه — عيّن بديل مؤقت.", r.Name, place, r.Until)
				for _, id := range s.supervisors(r.Employee) {
					notify(id, msg)
				}
			}
			continue
		}
		until, _ := time.ParseInLocation("2006-01-02", r.Until, baghdad)
		tx, err := s.db.Beginx()
		if err != nil {
			continue
		}
		var id string
		err = tx.Get(&id, `INSERT INTO "OrgAssignment" ("employeeId", "divisionId", "unitId", position, acting, "actingUntil", "replacesId", note)
			VALUES ($1, $2, $3, 'DIRECTOR', true, $4, $5, $6) RETURNING id`,
			*r.Deputy, r.DivisionID, r.UnitID, until.AddDate(0, 0, 1), r.AssignmentID, "وكالة — "+r.Name+" بإجازة (المبدأ ٢٠)")
		if err == nil {
			err = syncOrgPermissions(tx, *r.Deputy, "", "وكالة مدير بالإجازة")
		}
		if err != nil {
			tx.Rollback()
			log.Printf("[org-acting] وكالة %s: %v", r.AssignmentID, err)
			continue
		}
		if tx.Commit() == nil && notify != nil {
			notify(*r.Deputy, fmt.Sprintf("🏛️ صرت مدير %s بالوكالة لحد %s — %s بإجازة. صلاحيات المدير انفتحتلك مؤقتاً.", place, r.Until, r.Name))
			notify(r.Employee, fmt.Sprintf("🏛️ نائبك يمسك مكانك بالوكالة لحد %s. إجازة سعيدة.", r.Until))
		}
	}
	// إنهاء الوكالات: خلص وقتها، أو الإجازة انلغت، أو المدير الأصلي ما عاد بمنصبه
	ending := []struct {
		ID   string `db:"id"`
		Emp  string `db:"employeeId"`
		Name string `db:"name"`
	}{}
	_ = s.db.Select(&ending, `SELECT a.id, a."employeeId", e.name FROM "OrgAssignment" a
		JOIN "OrgAssignment" o ON o.id = a."replacesId" JOIN "Employee" e ON e.id = o."employeeId"
		WHERE a.acting AND a."endedAt" IS NULL AND (
		  a."actingUntil" <= now() OR o."endedAt" IS NOT NULL
		  OR NOT EXISTS (SELECT 1 FROM "LeaveRequest" l WHERE l."employeeId" = o."employeeId" AND l.status = 'APPROVED'
		                 AND l."startDate" <= $1::date AND l."endDate" >= $1::date))`, today)
	for _, a := range ending {
		if _, err := s.End("", a.ID, "انتهت الوكالة — "+a.Name+" رجع من الإجازة"); err == nil && notify != nil {
			notify(a.Emp, "🏛️ انتهت وكالتك — "+a.Name+" رجع لمكانه. شكراً على الإدارة.")
		}
	}
}

func (s *OrgService) placeName(div string, unit *string) string {
	var name string
	if unit != nil {
		_ = s.db.Get(&name, `SELECT name FROM "OrgUnit" WHERE id = $1`, *unit)
	} else {
		_ = s.db.Get(&name, `SELECT name FROM "OrgDivision" WHERE id = $1`, div)
	}
	return name
}

// supervisors مدير الإدارة (admin) + المالك — الي يقررون البديل.
func (s *OrgService) supervisors(exclude string) []string {
	ids := []string{}
	_ = s.db.Select(&ids, `SELECT DISTINCT id FROM (
		SELECT "employeeId" AS id FROM "OrgAssignment" WHERE "divisionId" = 'admin' AND "unitId" IS NULL AND position = 'DIRECTOR' AND "endedAt" IS NULL
		UNION SELECT id FROM "Employee" WHERE role = 'OWNER' AND status = 'ACTIVE') x WHERE id <> $1`, exclude)
	return ids
}

// claimDaily علامة مرة وحدة باليوم (بلا جدول جديد — OrgGovernorState).
func (s *OrgService) claimDaily(key, day string) bool {
	res, err := s.db.Exec(`INSERT INTO "OrgGovernorState" (key, value) VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, "updatedAt" = now() WHERE "OrgGovernorState".value <> EXCLUDED.value`, key, day)
	if err != nil {
		return false
	}
	n, _ := res.RowsAffected()
	return n == 1
}
