package service

import (
	"math"
	"sort"
	"strconv"

	"github.com/jmoiron/sqlx"
)

// 🏆 تصنيف المدراء (دازاي 11-01): مدير إدارة مقابل مدير إدارة، على الهيكل الجديد.
// ستة معايير، كل واحد من ١٠٠. المعيار الي ماله بيانات ما ينحسب (ما نظلم أحد بصفر).
// الأدمن والمالك ما ينقيّمون.

type DirectorCriterion struct {
	Key    string   `json:"key"`
	Label  string   `json:"label"`
	Weight float64  `json:"weight"`
	Score  *float64 `json:"score"`
	Detail string   `json:"detail"`
}

type DirectorRankRow struct {
	Rank         int                 `json:"rank"`
	DivisionID   string              `json:"divisionId"`
	DivisionName string              `json:"divisionName"`
	EmployeeID   string              `json:"employeeId"`
	EmployeeName string              `json:"employeeName"`
	Acting       bool                `json:"acting"`
	People       int                 `json:"people"`
	Total        *float64            `json:"total"`
	Criteria     []DirectorCriterion `json:"criteria"`
}

type DirectorRanking struct {
	Period string            `json:"period"`
	Rows   []DirectorRankRow `json:"rows"`
}

func clamp100(v float64) float64 { return math.Max(0, math.Min(100, math.Round(v*10)/10)) }
func rankPtr(v float64) *float64 { c := clamp100(v); return &c }

func DirectorsRanking(db *sqlx.DB, period string) (*DirectorRanking, error) {
	from, to, err := periodRange(period)
	if err != nil {
		return nil, err
	}
	heads := []struct {
		DivisionID   string `db:"divisionId"`
		DivisionName string `db:"divisionName"`
		EmployeeID   string `db:"employeeId"`
		EmployeeName string `db:"employeeName"`
		Acting       bool   `db:"acting"`
	}{}
	// مدير أصلي، وإذا ماكو فالي بالوكالة. الأدمن والمالك برّا.
	if err := db.Select(&heads, `
		SELECT DISTINCT ON (a."divisionId") a."divisionId", d.name AS "divisionName",
		       a."employeeId", e.name AS "employeeName", a.acting
		FROM "OrgAssignment" a
		JOIN "OrgDivision" d ON d.id = a."divisionId"
		JOIN "Employee" e ON e.id = a."employeeId"
		WHERE a."unitId" IS NULL AND a.position = 'DIRECTOR' AND a."endedAt" IS NULL
		  AND e.role::text NOT IN ('ADMIN', 'OWNER')
		ORDER BY a."divisionId", a.acting, a."startedAt"`); err != nil {
		return nil, err
	}
	out := &DirectorRanking{Period: period, Rows: []DirectorRankRow{}}
	for _, h := range heads {
		row := DirectorRankRow{DivisionID: h.DivisionID, DivisionName: h.DivisionName,
			EmployeeID: h.EmployeeID, EmployeeName: h.EmployeeName, Acting: h.Acting}

		// ١. المخرجات + ٦. الشكاوي والتأخير — من لوحة مخرجات الإدارة
		var outputs, complaints *float64
		outD, compD := "ماكو بيانات", "ماكو بيانات"
		if card, err := OutputsScorecard(db, "", "", h.DivisionID, period); err == nil {
			row.People = card.People
			sum, n, bad := 0.0, 0, 0
			for _, r := range card.Rows {
				if r.Score != nil {
					sum += *r.Score
					n++
				}
				bad += r.Complaints + r.TasksLate + r.TasksOverdue
			}
			if n > 0 {
				outputs = rankPtr(sum / float64(n))
				outD = strconv.Itoa(n) + " فقرة عليها بيانات"
			}
			if card.People > 0 {
				per := float64(bad) / float64(card.People)
				complaints = rankPtr(100 - per*20)
				compD = strconv.Itoa(bad) + " شكوى/تأخير على " + strconv.Itoa(card.People) + " موظف"
			}
		}

		// ٢. إنجاز الخطة
		var plan *float64
		planD := "ماكو خطة لهالشهر"
		var pl struct {
			N   int      `db:"n"`
			Avg *float64 `db:"avg"`
		}
		if err := db.Get(&pl, `SELECT count(*) AS n,
				avg(CASE WHEN i."doneAt" IS NOT NULL THEN 100 ELSE i."manualProgress" END)::float8 AS avg
			FROM "OrgPlanItem" i JOIN "OrgPlan" p ON p.id = i."planId"
			WHERE p."divisionId" = $1 AND p.period = $2`, h.DivisionID, period); err == nil && pl.N > 0 && pl.Avg != nil {
			plan = rankPtr(*pl.Avg)
			planD = strconv.Itoa(pl.N) + " بند بالخطة"
		}

		// ٣. السرعة، ٤. قلة المشاكل، ٥. حل المشاكل — من تقارير المشاكل
		var is struct {
			Opened   int      `db:"opened"`
			Weighted float64  `db:"weighted"`
			Resolved int      `db:"resolved"`
			AvgDays  *float64 `db:"avgDays"`
		}
		_ = db.Get(&is, `SELECT
				count(*) FILTER (WHERE "createdAt" >= $2 AND "createdAt" < $3) AS opened,
				COALESCE(sum(CASE severity WHEN 'HIGH' THEN 3 WHEN 'MEDIUM' THEN 2 ELSE 1 END)
					FILTER (WHERE "createdAt" >= $2 AND "createdAt" < $3), 0)::float8 AS weighted,
				count(*) FILTER (WHERE status = 'RESOLVED' AND "createdAt" >= $2 AND "createdAt" < $3) AS resolved,
				(avg(EXTRACT(EPOCH FROM ("resolvedAt" - "createdAt")) / 86400)
					FILTER (WHERE "resolvedAt" >= $2 AND "resolvedAt" < $3))::float8 AS "avgDays"
			FROM "OrgIssueReport" WHERE "divisionId" = $1`, h.DivisionID, from, to)
		var speed, few, solve *float64
		speedD, solveD := "ما انحلت مشكلة هالشهر", "ماكو مشاكل هالشهر"
		if is.AvgDays != nil {
			// ٣ أيام أو أقل = ١٠٠، ٣٠ يوم أو أكثر = صفر
			speed = rankPtr(100 - (*is.AvgDays-3)*100/27)
			speedD = "معدل الحل " + strconv.Itoa(int(math.Round(*is.AvgDays))) + " يوم"
		}
		if row.People > 0 {
			few = rankPtr(100 - is.Weighted/float64(row.People)*25)
		}
		fewD := strconv.Itoa(is.Opened) + " مشكلة انفتحت"
		if is.Opened > 0 {
			solve = rankPtr(float64(is.Resolved) * 100 / float64(is.Opened))
			solveD = strconv.Itoa(is.Resolved) + " من " + strconv.Itoa(is.Opened) + " انحلت"
		}

		row.Criteria = []DirectorCriterion{
			{"outputs", "المخرجات", 30, outputs, outD},
			{"plan", "إنجاز الخطة", 20, plan, planD},
			{"speed", "السرعة", 15, speed, speedD},
			{"few", "قلة المشاكل", 10, few, fewD},
			{"solve", "حل المشاكل", 15, solve, solveD},
			{"complaints", "الشكاوي والتأخير", 10, complaints, compD},
		}
		sum, w := 0.0, 0.0
		for _, c := range row.Criteria {
			if c.Score != nil {
				sum += *c.Score * c.Weight
				w += c.Weight
			}
		}
		if w > 0 {
			row.Total = rankPtr(sum / w)
		}
		out.Rows = append(out.Rows, row)
	}
	sort.SliceStable(out.Rows, func(i, j int) bool {
		a, b := out.Rows[i].Total, out.Rows[j].Total
		if a == nil || b == nil {
			return a != nil
		}
		return *a > *b
	})
	for i := range out.Rows {
		if t := out.Rows[i].Total; t != nil {
			// التعادل بنفس الدرجة = نفس المرتبة
			if i > 0 && out.Rows[i-1].Total != nil && *out.Rows[i-1].Total == *t {
				out.Rows[i].Rank = out.Rows[i-1].Rank
			} else {
				out.Rows[i].Rank = i + 1
			}
		}
	}
	return out, nil
}
