package service

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
)

// EmpMetrics أرقام موظف: آخر ١٤ يوم (الحالي) والـ١٤ قبلها (Prev).
type EmpMetrics struct {
	ID       string `db:"id" json:"-"`
	Name     string `db:"name" json:"-"`
	Att      int    `db:"att" json:"att"`
	AttPrev  int    `db:"attPrev" json:"attPrev"`
	Late     int    `db:"late" json:"late"`
	LatePrev int    `db:"latePrev" json:"latePrev"`
	Done     int    `db:"done" json:"done"`
	DonePrev int    `db:"donePrev" json:"donePrev"`
	Leave    int    `db:"leave" json:"leave"`
	Streak   int    `db:"streak" json:"streak"`
}

// EmployeeMetrics أرقام الموظفين النشطين (أو واحد بس) — مشتركة بين المؤشرات والمتابعة.
func EmployeeMetrics(db *sqlx.DB, onlyID string) ([]EmpMetrics, error) {
	rows := []EmpMetrics{}
	err := db.Select(&rows, `
				WITH emp AS (
					SELECT id, name FROM "Employee"
					WHERE status = 'ACTIVE' AND role::text NOT IN ('OWNER', 'ADMIN') AND ($1 = '' OR id = $1)
				), att AS (
					SELECT "employeeId", (date AT TIME ZONE 'UTC')::date AS d,
						bool_or(((("checkIn" AT TIME ZONE 'UTC') + interval '3 hours')::time) > time '09:15') AS late
					FROM "Attendance" WHERE "checkIn" > now() - interval '28 days' GROUP BY 1, 2
				), days AS (
					SELECT "employeeId", d, d - (row_number() OVER (PARTITION BY "employeeId" ORDER BY d))::int AS grp
					FROM att WHERE d > current_date - 30
				), streak AS (
					SELECT "employeeId", max(n) AS streak FROM (
						SELECT "employeeId", grp, count(*) AS n, max(d) AS last FROM days GROUP BY 1, 2) x
					WHERE last >= current_date - 1 GROUP BY 1
				), done AS (
					SELECT x.eid, count(*) FILTER (WHERE x.at > now() - interval '14 days') AS cur,
						count(*) FILTER (WHERE x.at <= now() - interval '14 days') AS prev
					FROM (SELECT DISTINCT b.id, COALESCE(a."employeeId", b."projectSupervisorId") AS eid, b."completedAt" AS at
						FROM "Booking" b LEFT JOIN "BookingAssignment" a ON a."bookingId" = b.id
						WHERE b.status = 'COMPLETED' AND b."completedAt" > now() - interval '28 days') x
					GROUP BY 1
				)
				SELECT e.id, e.name,
					count(t.d) FILTER (WHERE t.d > current_date - 14) AS att,
					count(t.d) FILTER (WHERE t.d <= current_date - 14) AS "attPrev",
					count(t.d) FILTER (WHERE t.d > current_date - 14 AND t.late) AS late,
					count(t.d) FILTER (WHERE t.d <= current_date - 14 AND t.late) AS "latePrev",
					COALESCE(max(dn.cur), 0) AS done, COALESCE(max(dn.prev), 0) AS "donePrev",
					COALESCE((SELECT sum(LEAST(l."endDate", current_date) - GREATEST(l."startDate", current_date - 14) + 1)
						FROM "LeaveRequest" l WHERE l."employeeId" = e.id AND l.status = 'APPROVED'
						  AND l."endDate" >= current_date - 14 AND l."startDate" <= current_date), 0)::int AS leave,
					COALESCE(max(s.streak), 0) AS streak
				FROM emp e
				LEFT JOIN att t ON t."employeeId" = e.id
				LEFT JOIN done dn ON dn.eid = e.id
				LEFT JOIN streak s ON s."employeeId" = e.id
				GROUP BY e.id, e.name`, onlyID)
	return rows, err
}

// ═══ 🩺 مؤشرات الموظف (دازاي 10-29) ═══
// «اريد النظام يقيس ويحلل…» — اتفقنا: مؤشرات من الشغل الحقيقي مو تخمين نفسي.
// نقارن آخر ١٤ يوم بالـ١٤ الي قبلها: الحضور، التأخير، الحجوزات المنجزة، الإجازات،
// وأطول سلسلة أيام دوام ورا بعض. الناتج «إشارة حتى ينتبه»، مو حكم على أحد.

func (a *MatrixAgent) registerSignalsTool() {
	a.AddTool(agentTool{Name: "employee_signals",
		Desc:  "مؤشرات حال الموظفين من شغلهم الحقيقي (آخر ١٤ يوم مقابل الـ١٤ قبلها): الحضور، التأخير، الإنجاز، الإجازات، والدوام المتواصل بلا راحة. بلا who = منو يحتاج انتباه؛ ويا who = موظف واحد. استعملها لأي سؤال عن «حال/نفسية/ضغط/تعب/منو مرتاح» — وگول بوضوح إنها إشارات من الأرقام مو تشخيص نفسي، ولا تحكم على سعادة أو حزن أحد.",
		Props: map[string]any{"who": map[string]any{"type": "string", "description": "اسم أو رمز موظف (اختياري)"}},
		Run: func(in map[string]any, mask *nameMask) (string, error) {
			onlyID := ""
			if w := strings.TrimSpace(strArg(in, "who")); w != "" {
				id, _ := mask.Resolve(w)
				if id == "" {
					return "", fmt.Errorf("ما لگيت الموظف")
				}
				onlyID = id
			}
			rows, err := EmployeeMetrics(a.db, onlyID)
			if err != nil {
				return "", err
			}
			type flagged struct {
				line  string
				score int
			}
			out := []flagged{}
			for _, r := range rows {
				var why []string
				score := 0
				if r.Late-r.LatePrev >= 2 {
					why = append(why, fmt.Sprintf("التأخير زاد (%d ← %d يوم)", r.LatePrev, r.Late))
					score += r.Late - r.LatePrev
				}
				if r.AttPrev-r.Att >= 3 && r.Leave < r.AttPrev-r.Att {
					why = append(why, fmt.Sprintf("الحضور نزل (%d ← %d يوم) بلا إجازة تفسره", r.AttPrev, r.Att))
					score += r.AttPrev - r.Att
				}
				if r.DonePrev >= 3 && r.Done*10 <= r.DonePrev*6 {
					why = append(why, fmt.Sprintf("الإنجاز نزل (%d ← %d حجز)", r.DonePrev, r.Done))
					score += 2
				}
				if r.Streak >= 12 {
					why = append(why, fmt.Sprintf("داوم %d يوم ورا بعض بلا راحة — احتمال تعب", r.Streak))
					score += 3
				}
				if r.Leave >= 5 {
					why = append(why, fmt.Sprintf("إجازات %d يوم بآخر أسبوعين", r.Leave))
				}
				if onlyID != "" {
					line := fmt.Sprintf("%s — آخر ١٤ يوم مقابل الي قبلها: حضور %d←%d، تأخير %d←%d، منجز %d←%d، إجازة %d يوم، أطول دوام متواصل %d يوم.",
						r.Name, r.AttPrev, r.Att, r.LatePrev, r.Late, r.DonePrev, r.Done, r.Leave, r.Streak)
					if len(why) > 0 {
						line += "\nإشارات: " + strings.Join(why, "؛ ")
					} else {
						line += "\nماكو إشارة لافتة — أرقامه مستقرة."
					}
					return line, nil
				}
				if len(why) > 0 {
					out = append(out, flagged{line: "- " + r.Name + ": " + strings.Join(why, "؛ "), score: score})
				}
			}
			if onlyID != "" {
				return "ماكو بيانات لهالموظف.", nil
			}
			if len(out) == 0 {
				return "ماكو إشارات لافتة بآخر أسبوعين — الحضور والإنجاز مستقرين عند الكل.", nil
			}
			sort.SliceStable(out, func(i, j int) bool { return out[i].score > out[j].score })
			if len(out) > 12 {
				out = out[:12]
			}
			var b strings.Builder
			b.WriteString("يستاهلون انتباه (إشارات من الأرقام، مو حكم — الأفضل تسأل عليهم):\n")
			for _, f := range out {
				b.WriteString(f.line + "\n")
			}
			return b.String(), nil
		}})
}
