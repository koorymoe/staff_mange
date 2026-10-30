package service

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// ═══ 🎯 ربط المهام والـKPI بالمخرجات الرئيسية (دازاي 10-30 — النقطة ٨) ═══
// «تربط المهام وتقييمات KPI مع الأداء للمخرجات الرئيسية». المخرجات = فقرات لائحة العمل.
// لكل فقرة نجمع: نقاط ماتركس (كل قاعدة مربوطة بفقرة)، المهام الإضافية المربوطة (بوقتها/
// متأخرة/فاتت)، خصومات الـKPI بمعاييرها المربوطة، والشكاوي (م٤/ف٣). لموظف أو وحدة أو إدارة.

// matrixRuleClause قواعد نقاط ماتركس ← فقرة اللائحة (principleNo).
var matrixRuleClause = map[string]int{
	"ATTENDANCE": 3,                 // م١/ف٣ مواعيد العمل
	"MONITOR":    4, "STAGE_LOG": 4, // م١/ف٤ المسؤولية والمحاسبة
	"TASK_DUE": 2,                                   // م١/ف٢ الالتزام بالتوجيهات
	"WORK":     7, "COMPLETE": 7, "PROJECT_WEEK": 7, // م٢/ف١ الدقة والسرعة
	"CREW":      9,                                                                    // م٢/ف٣ العمل الجماعي
	"INVENTORY": 11, "TOOL_RESPONSE": 11, "PROC_RESPONSE": 11, "SHORTAGE_RESOLVE": 11, // م٣/ف١ العُدد
	"VEHICLE_DAILY": 12,                                                           // م٣/ف٢ الأصول والمركبات
	"ACCOUNT":       13,                                                           // م٣/ف٣ الضبط المالي
	"PAPER":         15, "CONTACT": 15, "CONFIRM": 15, "ROUTE": 15, "RECEIVE": 15, // م٤/ف٢ المتابعة
	"RATING": 16, "QUALITY": 16, // م٤/ف٣ جودة الخدمة
}

type OutputRow struct {
	Clause       int      `json:"clause"`
	Code         string   `json:"code"`
	Title        string   `json:"title"`
	MatrixEarned int      `json:"matrixEarned"`
	MatrixMax    int      `json:"matrixMax"`
	MatrixPct    *float64 `json:"matrixPct"`
	TasksOnTime  int      `json:"tasksOnTime"`
	TasksLate    int      `json:"tasksLate"`
	TasksOverdue int      `json:"tasksOverdue"`
	KpiDeducts   int      `json:"kpiDeductions"`
	Complaints   int      `json:"complaints"`
	// Score درجة الفقرة (٠-١٠٠): نقاط ماتركس ويا المهام بوقتها، ناقص الخصومات والشكاوي
	Score *float64 `json:"score"`
}

type OutputsCard struct {
	Scope  string      `json:"scope"`
	Label  string      `json:"label"`
	Period string      `json:"period"`
	People int         `json:"people"`
	Rows   []OutputRow `json:"rows"`
	// Weakest أضعف ٣ فقرات (للتطوير)
	Weakest []string `json:"weakest"`
}

// OutputsScorecard لوحة المخرجات لموظف (employeeID) أو وحدة (unitID) أو إدارة (divisionID).
func OutputsScorecard(db *sqlx.DB, employeeID, unitID, divisionID, period string) (*OutputsCard, error) {
	from, to, err := periodRange(period)
	if err != nil {
		return nil, err
	}
	card := &OutputsCard{Period: period, Rows: []OutputRow{}, Weakest: []string{}}
	ids := []string{}
	switch {
	case employeeID != "":
		ids = []string{employeeID}
		card.Scope = "EMPLOYEE"
		_ = db.Get(&card.Label, `SELECT name FROM "Employee" WHERE id = $1`, employeeID)
	case unitID != "":
		card.Scope = "UNIT"
		_ = db.Get(&card.Label, `SELECT d.name || ' / ' || u.name FROM "OrgUnit" u JOIN "OrgDivision" d ON d.id = u."divisionId" WHERE u.id = $1`, unitID)
		_ = db.Select(&ids, `SELECT DISTINCT "employeeId" FROM "OrgAssignment" WHERE "unitId" = $1 AND "startedAt" < $3 AND ("endedAt" IS NULL OR "endedAt" >= $2)`, unitID, from, to)
	case divisionID != "":
		card.Scope = "DIVISION"
		_ = db.Get(&card.Label, `SELECT name FROM "OrgDivision" WHERE id = $1`, divisionID)
		_ = db.Select(&ids, `SELECT DISTINCT "employeeId" FROM "OrgAssignment" WHERE "divisionId" = $1 AND "startedAt" < $3 AND ("endedAt" IS NULL OR "endedAt" >= $2)`, divisionID, from, to)
	default:
		return nil, errors.New("حدد موظف أو وحدة أو إدارة")
	}
	card.People = len(ids)
	clauses := []struct {
		No    int    `db:"no"`
		Code  string `db:"code"`
		Title string `db:"title"`
	}{}
	if err := db.Select(&clauses, `SELECT "principleNo" AS no, split_part(content, ' — ', 1) AS code, COALESCE("clauseTitle", '') AS title
		FROM "PrivacyPolicyPoint" WHERE "isActive" AND "principleNo" IS NOT NULL ORDER BY "principleNo"`); err != nil {
		return nil, err
	}
	rows := map[int]*OutputRow{}
	for _, c := range clauses {
		rows[c.No] = &OutputRow{Clause: c.No, Code: c.Code, Title: c.Title}
	}
	if len(ids) > 0 {
		// نقاط ماتركس حسب القاعدة
		mr := []struct {
			Rule string `db:"rule"`
			P    int    `db:"p"`
			M    int    `db:"m"`
		}{}
		_ = db.Select(&mr, `SELECT rule, sum(points) AS p, sum("maxPoints") AS m FROM "MatrixScore"
			WHERE "employeeId" = ANY($1) AND "cancelledAt" IS NULL AND "at" >= $2 AND "at" < $3 GROUP BY rule`, pq.Array(ids), from, to)
		for _, r := range mr {
			if c, ok := matrixRuleClause[r.Rule]; ok && rows[c] != nil {
				rows[c].MatrixEarned += r.P
				rows[c].MatrixMax += r.M
			}
		}
		// المهام الإضافية المربوطة
		tr := []struct {
			Clause  int `db:"clause"`
			OnTime  int `db:"ontime"`
			Late    int `db:"late"`
			Overdue int `db:"overdue"`
		}{}
		_ = db.Select(&tr, `SELECT clause,
			count(*) FILTER (WHERE "doneAt" IS NOT NULL AND ("dueAt" IS NULL OR "doneAt" <= "dueAt")) AS ontime,
			count(*) FILTER (WHERE "doneAt" IS NOT NULL AND "dueAt" IS NOT NULL AND "doneAt" > "dueAt") AS late,
			count(*) FILTER (WHERE "doneAt" IS NULL AND "cancelledAt" IS NULL AND "dueAt" < now()) AS overdue
			FROM "ExtraTask" WHERE "assignedToId" = ANY($1) AND clause IS NOT NULL
			AND COALESCE("dueAt", "createdAt") >= $2 AND COALESCE("dueAt", "createdAt") < $3 GROUP BY clause`, pq.Array(ids), from, to)
		for _, r := range tr {
			if rows[r.Clause] != nil {
				rows[r.Clause].TasksOnTime, rows[r.Clause].TasksLate, rows[r.Clause].TasksOverdue = r.OnTime, r.Late, r.Overdue
			}
		}
		// خصومات الـKPI (التقييم مربوط بمعياره بالاسم)
		kr := []struct {
			Clause int `db:"clause"`
			N      int `db:"n"`
		}{}
		_ = db.Select(&kr, `SELECT c.clause, count(*) AS n FROM "KpiEvaluation" e JOIN "KpiCriterion" c ON c.label = e.reason
			WHERE e."employeeId" = ANY($1) AND NOT e.cancelled AND c.clause IS NOT NULL AND e."createdAt" >= $2 AND e."createdAt" < $3
			GROUP BY c.clause`, pq.Array(ids), from, to)
		for _, r := range kr {
			if rows[r.Clause] != nil {
				rows[r.Clause].KpiDeducts = r.N
			}
		}
		// الشكاوي ← م٤/ف٣ جودة الخدمة
		if rows[16] != nil {
			_ = db.Get(&rows[16].Complaints, `SELECT count(*) FROM "Complaint" c WHERE `+orgComplaintPerson+` = ANY($1)
				AND c."createdAt" >= $2 AND c."createdAt" < $3`, pq.Array(ids), from, to)
		}
	}
	type weak struct {
		label string
		score float64
	}
	weaks := []weak{}
	for _, c := range clauses {
		r := rows[c.No]
		parts, w := 0.0, 0.0
		if r.MatrixMax > 0 {
			v := math.Round(float64(r.MatrixEarned)/float64(r.MatrixMax)*1000) / 10
			r.MatrixPct = &v
			parts += v * 0.7
			w += 0.7
		}
		if n := r.TasksOnTime + r.TasksLate + r.TasksOverdue; n > 0 {
			parts += float64(r.TasksOnTime) / float64(n) * 100 * 0.3
			w += 0.3
		}
		if w > 0 || r.KpiDeducts > 0 || r.Complaints > 0 {
			base := 100.0
			if w > 0 {
				base = parts / w
			}
			// كل خصم KPI أو شكوى ينزّل ١٠ (حد أدنى صفر)
			v := math.Max(0, math.Round((base-float64(10*(r.KpiDeducts+r.Complaints)))*10)/10)
			r.Score = &v
			weaks = append(weaks, weak{r.Code + " " + r.Title, v})
		}
		card.Rows = append(card.Rows, *r)
	}
	sort.Slice(weaks, func(i, j int) bool { return weaks[i].score < weaks[j].score })
	for i := 0; i < len(weaks) && i < 3; i++ {
		if weaks[i].score < 80 {
			card.Weakest = append(card.Weakest, fmt.Sprintf("%s (%.0f)", weaks[i].label, weaks[i].score))
		}
	}
	return card, nil
}

func (c *OutputsCard) Text() string {
	var b strings.Builder
	fmt.Fprintf(&b, "مخرجات %s — %s (%d شخص):\n", c.Label, c.Period, c.People)
	for _, r := range c.Rows {
		if r.Score == nil {
			continue
		}
		fmt.Fprintf(&b, "- %s %s: %.0f", r.Code, r.Title, *r.Score)
		if r.MatrixPct != nil {
			fmt.Fprintf(&b, " (ماتركس %.0f٪)", *r.MatrixPct)
		}
		if n := r.TasksOnTime + r.TasksLate + r.TasksOverdue; n > 0 {
			fmt.Fprintf(&b, " مهام: %d بوقتها/%d متأخرة/%d فاتت", r.TasksOnTime, r.TasksLate, r.TasksOverdue)
		}
		if r.KpiDeducts > 0 {
			fmt.Fprintf(&b, " خصومات KPI %d", r.KpiDeducts)
		}
		if r.Complaints > 0 {
			fmt.Fprintf(&b, " شكاوي %d", r.Complaints)
		}
		b.WriteString("\n")
	}
	if len(c.Weakest) > 0 {
		b.WriteString("أضعف الفقرات: " + strings.Join(c.Weakest, "؛ "))
	} else {
		b.WriteString("ماكو فقرة ضعيفة واضحة (أو ماكو بيانات كافية).")
	}
	return b.String()
}

// SetTaskClause / SetCriterionClause ربط مهمة أو معيار KPI بفقرة (nil = فك الربط).
func SetTaskClause(db *sqlx.DB, taskID string, clause *int) error {
	if clause != nil && (*clause < 1 || *clause > 40) {
		return errors.New("رقم الفقرة غير صالح")
	}
	res, err := db.Exec(`UPDATE "ExtraTask" SET clause = $2 WHERE id = $1`, taskID, clause)
	if n, _ := res.RowsAffected(); err == nil && n == 0 {
		return errors.New("المهمة مو موجودة")
	}
	return err
}

func SetCriterionClause(db *sqlx.DB, criterionID string, clause *int) error {
	if clause != nil && (*clause < 1 || *clause > 40) {
		return errors.New("رقم الفقرة غير صالح")
	}
	res, err := db.Exec(`UPDATE "KpiCriterion" SET clause = $2 WHERE id = $1`, criterionID, clause)
	if n, _ := res.RowsAffected(); err == nil && n == 0 {
		return errors.New("المعيار مو موجود")
	}
	return err
}

func (a *MatrixAgent) registerOutputsTool() {
	a.AddTool(agentTool{Name: "output_scorecard",
		Desc: "لوحة المخرجات: أداء موظف أو وحدة أو إدارة لكل فقرة من لائحة العمل بالشهر — نقاط ماتركس المربوطة، المهام الإضافية (بوقتها/متأخرة/فاتت)، خصومات الـKPI، والشكاوي، وأضعف الفقرات. who = اسم موظف، أو unit = رمز وحدة، أو division = رمز إدارة. month اختياري YYYY-MM.",
		Props: map[string]any{
			"who": map[string]any{"type": "string"}, "unit": map[string]any{"type": "string"},
			"division": map[string]any{"type": "string"}, "month": map[string]any{"type": "string"},
		},
		Run: func(in map[string]any, mask *nameMask) (string, error) {
			period := strings.TrimSpace(strArg(in, "month"))
			if period == "" {
				period = OrgPeriod(time.Now())
			}
			emp := ""
			if w := strings.TrimSpace(strArg(in, "who")); w != "" {
				id, _ := mask.Resolve(w)
				if id == "" {
					return "", errors.New("ما لگيت الموظف")
				}
				emp = id
			}
			card, err := OutputsScorecard(a.db, emp, strings.TrimSpace(strArg(in, "unit")), strings.TrimSpace(strArg(in, "division")), period)
			if err != nil {
				return "", err
			}
			return mask.Hide(card.Text()), nil
		}})
}

// CanSeeOutputs منو يشوف لوحة مخرجات منو.
func (s *OrgService) CanSeeOutputs(empID, role, employeeID, unitID, divisionID string) bool {
	if role == "OWNER" || role == "ADMIN" {
		return true
	}
	switch {
	case employeeID != "":
		if employeeID == empID {
			return true
		}
		// رئيس يشوف كادر وحداته/إدارته
		var n int
		_ = s.db.Get(&n, `SELECT count(*) FROM "OrgAssignment" h JOIN "OrgAssignment" m ON m."endedAt" IS NULL AND m."employeeId" = $2
			AND (h."unitId" = m."unitId" OR (h."unitId" IS NULL AND h."divisionId" = m."divisionId"))
			WHERE h."employeeId" = $1 AND h."endedAt" IS NULL AND h.position IN ('DIRECTOR', 'DEPUTY')`, empID, employeeID)
		return n > 0
	case unitID != "":
		var div string
		if s.db.Get(&div, `SELECT "divisionId" FROM "OrgUnit" WHERE id = $1`, unitID) != nil {
			return false
		}
		return s.UnitAccess(empID, role, div, unitID)
	case divisionID != "":
		see, _ := s.Access(empID, role, divisionID)
		return see
	}
	return false
}
