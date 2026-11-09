package service

import (
	"math"
	"sort"
	"strconv"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/errlog"
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
	// (دازاي 11-04) «اريد اشوف على شنو اعتمد»: المعادلة، والسطور الي انحسب منها الرقم.
	Formula  string         `json:"formula"`
	Evidence []EvidenceLine `json:"evidence"`
}

// EvidenceLine سطر دليل: شنو (Label)، وقيمته (Value)، وملاحظة (Note).
type EvidenceLine struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Note  string `json:"note,omitempty"`
}

func f1(v float64) string { return strconv.FormatFloat(math.Round(v*10)/10, 'f', -1, 64) }

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
	// للنواب: نص رقم الإدارة ونص شغله هو (دازاي 11-04 «خليط»)
	DivisionTotal *float64 `json:"divisionTotal,omitempty"`
	OwnTotal      *float64 `json:"ownTotal,omitempty"`
}

type DirectorRanking struct {
	Period   string            `json:"period"`
	Rows     []DirectorRankRow `json:"rows"`
	Deputies []DirectorRankRow `json:"deputies"`
	// (دازاي 11-06) مدراء ونواب الوحدات: على خطة تحسين وحدتهم
	Units []DirectorRankRow `json:"units"`
}

// PersonalScoreFn التقييم الشهري للموظف نفسه (MatrixScoreService.Detail): النهائي + أكبر الخسارات.
type PersonalScoreFn func(employeeID, period string) (final *float64, lines []EvidenceLine)

func clamp100(v float64) float64 { return math.Max(0, math.Min(100, math.Round(v*10)/10)) }
func rankPtr(v float64) *float64 { c := clamp100(v); return &c }

func DirectorsRanking(db *sqlx.DB, period string, personal PersonalScoreFn) (*DirectorRanking, error) {
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
		outEv, compEv := []EvidenceLine{}, []EvidenceLine{}
		if card, err := OutputsScorecard(db, "", "", h.DivisionID, period); err == nil {
			row.People = card.People
			sum, n, bad := 0.0, 0, 0
			for _, r := range card.Rows {
				if r.Score != nil {
					sum += *r.Score
					n++
					note := "ماتركس " + strconv.Itoa(r.MatrixEarned) + "/" + strconv.Itoa(r.MatrixMax) +
						" · مهام بوقتها " + strconv.Itoa(r.TasksOnTime) + " · متأخرة " + strconv.Itoa(r.TasksLate+r.TasksOverdue)
					if r.KpiDeducts > 0 {
						note += " · خصومات " + strconv.Itoa(r.KpiDeducts)
					}
					outEv = append(outEv, EvidenceLine{Label: r.Title, Value: f1(*r.Score), Note: note})
				}
				bad += r.Complaints + r.TasksLate + r.TasksOverdue
				if r.Complaints+r.TasksLate+r.TasksOverdue > 0 {
					compEv = append(compEv, EvidenceLine{Label: r.Title,
						Value: strconv.Itoa(r.Complaints + r.TasksLate + r.TasksOverdue),
						Note:  "شكاوي " + strconv.Itoa(r.Complaints) + " · متأخرة " + strconv.Itoa(r.TasksLate) + " · فاتها الموعد " + strconv.Itoa(r.TasksOverdue)})
				}
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
				avg(`+planItemScoreSQL+`)::float8 AS avg
			FROM "OrgPlanItem" i JOIN "OrgPlan" p ON p.id = i."planId"
			WHERE p."divisionId" = $1 AND p.period = $2`, h.DivisionID, period); err == nil && pl.N > 0 && pl.Avg != nil {
			plan = rankPtr(*pl.Avg)
			planD = strconv.Itoa(pl.N) + " بند بالخطة"
		}
		planEv := planEvidence(db, `p."divisionId" = $1 AND p.period = $2`, h.DivisionID, period)

		// ٣. السرعة، ٤. قلة المشاكل، ٥. حل المشاكل — من تقارير المشاكل
		var is struct {
			Opened   int      `db:"opened"`
			Weighted float64  `db:"weighted"`
			Resolved int      `db:"resolved"`
			AvgDays  *float64 `db:"avgDays"`
		}
		errlog.Err(db.Get(&is, `SELECT
				count(*) FILTER (WHERE "createdAt" >= $2 AND "createdAt" < $3) AS opened,
				COALESCE(sum(CASE severity WHEN 'HIGH' THEN 3 WHEN 'MEDIUM' THEN 2 ELSE 1 END)
					FILTER (WHERE "createdAt" >= $2 AND "createdAt" < $3), 0)::float8 AS weighted,
				count(*) FILTER (WHERE status = 'RESOLVED' AND "createdAt" >= $2 AND "createdAt" < $3) AS resolved,
				(avg(EXTRACT(EPOCH FROM ("resolvedAt" - "createdAt")) / 86400)
					FILTER (WHERE "resolvedAt" >= $2 AND "resolvedAt" < $3))::float8 AS "avgDays"
			FROM "OrgIssueReport" WHERE "divisionId" = $1`, h.DivisionID, from, to))
		issues := issueEvidence(db, h.DivisionID, from, to)
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

		peopleS := strconv.Itoa(row.People)
		row.Criteria = []DirectorCriterion{
			{"outputs", "المخرجات", 30, outputs, outD, "معدل درجات فقرات المخرجات الي عليها بيانات (كل فقرة: نقاط ماتركس + المهام بوقتها − الخصومات والشكاوي)", outEv},
			{"plan", "إنجاز الخطة", 20, plan, planD, "معدل درجات بنود خطة الشهر: خلص بموعده = ١٠٠، متأخر = −١٠ لكل يوم، وبعده بالموعد = نسبته", planEv},
			{"speed", "السرعة", 15, speed, speedD, "١٠٠ − (معدل أيام الحل − ٣) × ١٠٠ ÷ ٢٧ — ٣ أيام أو أقل = ١٠٠، و٣٠ يوم أو أكثر = صفر", issues.resolved},
			{"few", "قلة المشاكل", 10, few, fewD, "١٠٠ − (مجموع أوزان المشاكل ÷ " + peopleS + " موظف) × ٢٥ — الوزن: عالية ٣، متوسطة ٢، واطية ١", issues.opened},
			{"solve", "حل المشاكل", 15, solve, solveD, "المشاكل الي انحلت ÷ المشاكل الي انفتحت هالشهر × ١٠٠", issues.opened},
			{"complaints", "الشكاوي والتأخير", 10, complaints, compD, "١٠٠ − (الشكاوي والمهام المتأخرة ÷ " + peopleS + " موظف) × ٢٠", compEv},
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
	out.Deputies = deputiesRanking(db, period, out.Rows, personal)
	out.Units = unitLeadersRanking(db, period, personal)
	return out, nil
}

var sevAr = map[string]string{"HIGH": "عالية", "MEDIUM": "متوسطة", "LOW": "واطية"}
var issueStatusAr = map[string]string{"RESOLVED": "انحلت", "OPEN": "مفتوحة", "IN_PROGRESS": "قيد الحل"}

type issueEv struct{ opened, resolved []EvidenceLine }

// issueEvidence المشاكل نفسها بأسمائها: الي انفتحت بالشهر، والي انحلت بالشهر وبكم يوم.
func issueEvidence(db *sqlx.DB, divisionID string, from, to any) issueEv {
	rows := []struct {
		Title      string   `db:"title"`
		Severity   string   `db:"severity"`
		Status     string   `db:"status"`
		Opened     bool     `db:"opened"`
		ResolvedIn bool     `db:"resolvedIn"`
		Days       *float64 `db:"days"`
	}{}
	// ⚠️ resolvedAt الفاضي يطلع NULL بالمقارنة — بلا COALESCE القراءة كلها تفشل بصمت
	errlog.Err(db.Select(&rows, `SELECT COALESCE(title, '') AS title, COALESCE(severity::text, '') AS severity, COALESCE(status::text, '') AS status,
			("createdAt" >= $2 AND "createdAt" < $3) AS opened,
			COALESCE("resolvedAt" >= $2 AND "resolvedAt" < $3, false) AS "resolvedIn",
			(EXTRACT(EPOCH FROM ("resolvedAt" - "createdAt")) / 86400)::float8 AS days
		FROM "OrgIssueReport" WHERE "divisionId" = $1
		  AND (("createdAt" >= $2 AND "createdAt" < $3) OR ("resolvedAt" >= $2 AND "resolvedAt" < $3))
		ORDER BY "createdAt"`, divisionID, from, to))
	ev := issueEv{opened: []EvidenceLine{}, resolved: []EvidenceLine{}}
	for _, r := range rows {
		st := issueStatusAr[r.Status]
		if st == "" {
			st = r.Status
		}
		if r.Opened {
			ev.opened = append(ev.opened, EvidenceLine{Label: r.Title, Value: st, Note: "خطورة " + sevAr[r.Severity]})
		}
		if r.ResolvedIn && r.Days != nil {
			ev.resolved = append(ev.resolved, EvidenceLine{Label: r.Title, Value: f1(*r.Days) + " يوم", Note: "خطورة " + sevAr[r.Severity]})
		}
	}
	return ev
}

// planItemScoreSQL (دازاي 11-06) درجة البند بالالتزام بالموعد، مو بس «خلص لو لا»:
// خلص بموعده = ١٠٠؛ خلص متأخر = ١٠٠ − ١٠ لكل يوم (أقل شي ٤٠)؛ فات موعده وما خلص =
// نسبته − ١٠ لكل يوم (أقل شي صفر)؛ بعده بالموعد = نسبته المسجّلة.
const planItemScoreSQL = `CASE
	WHEN i."doneAt" IS NOT NULL THEN GREATEST(40, 100 - 10 * GREATEST(0, COALESCE(baghdad_date(i."doneAt") - i."dueDate", 0)))
	WHEN i."dueDate" IS NOT NULL AND i."dueDate" < baghdad_today() THEN GREATEST(0, i."manualProgress" - 10 * (baghdad_today() - i."dueDate"))
	ELSE i."manualProgress" END`

// planEvidence بنود الخطة: درجة كل بند، والسبب (بموعده/متأخر كم يوم)، ومنو مسؤول عنه.
func planEvidence(db *sqlx.DB, where string, args ...any) []EvidenceLine {
	rows := []struct {
		Title    string  `db:"title"`
		Pct      float64 `db:"pct"`
		Owner    *string `db:"owner"`
		Done     bool    `db:"done"`
		LateDays int     `db:"lateDays"`
		Manual   int     `db:"manualProgress"`
	}{}
	errlog.Err(db.Select(&rows, `SELECT i.title, COALESCE(`+planItemScoreSQL+`, 0)::float8 AS pct, e.name AS owner,
			i."doneAt" IS NOT NULL AS done, i."manualProgress",
			GREATEST(0, COALESCE(COALESCE(baghdad_date(i."doneAt"), baghdad_today()) - i."dueDate", 0))::int AS "lateDays"
		FROM "OrgPlanItem" i JOIN "OrgPlan" p ON p.id = i."planId" LEFT JOIN "Employee" e ON e.id = i."ownerId"
		WHERE `+where+` ORDER BY i."createdAt"`, args...))
	out := []EvidenceLine{}
	for _, r := range rows {
		why := ""
		switch {
		case r.Done && r.LateDays == 0:
			why = "خلص بموعده"
		case r.Done:
			why = "خلص متأخر " + strconv.Itoa(r.LateDays) + " يوم"
		case r.LateDays > 0:
			why = "فات موعده بـ" + strconv.Itoa(r.LateDays) + " يوم وما خلص (نسبته " + strconv.Itoa(r.Manual) + "٪)"
		default:
			why = "بعده بموعده (نسبته " + strconv.Itoa(r.Manual) + "٪)"
		}
		l := EvidenceLine{Label: r.Title, Value: f1(r.Pct) + "٪", Note: why}
		if r.Owner != nil {
			l.Note += " · المسؤول: " + *r.Owner
		}
		out = append(out, l)
	}
	return out
}

// deputiesRanking (دازاي 11-04 «خليط»): النائب = ٥٠٪ رقم إدارته + ٥٠٪ شغله هو
// (تقييمه الشهري، وبنود الخطة الي برقبته). الأدمن والمالك ما ينقيّمون.
func deputiesRanking(db *sqlx.DB, period string, directors []DirectorRankRow, personal PersonalScoreFn) []DirectorRankRow {
	divTotal := map[string]*float64{}
	for _, d := range directors {
		divTotal[d.DivisionID] = d.Total
	}
	deps := []struct {
		DivisionID   string `db:"divisionId"`
		DivisionName string `db:"divisionName"`
		EmployeeID   string `db:"employeeId"`
		EmployeeName string `db:"employeeName"`
	}{}
	errlog.Err(db.Select(&deps, `
		SELECT DISTINCT ON (a."divisionId", a."employeeId") a."divisionId", d.name AS "divisionName", a."employeeId", e.name AS "employeeName"
		FROM "OrgAssignment" a
		JOIN "OrgDivision" d ON d.id = a."divisionId"
		JOIN "Employee" e ON e.id = a."employeeId"
		WHERE a."unitId" IS NULL AND a.position = 'DEPUTY' AND a."endedAt" IS NULL
		  AND e.role::text NOT IN ('ADMIN', 'OWNER')
		ORDER BY a."divisionId", a."employeeId"`))
	out := []DirectorRankRow{}
	for _, d := range deps {
		row := DirectorRankRow{DivisionID: d.DivisionID, DivisionName: d.DivisionName, EmployeeID: d.EmployeeID, EmployeeName: d.EmployeeName}
		// ١. رقم الإدارة (نفس رقم المدير)
		div := divTotal[d.DivisionID]
		divD := "ماكو رقم للإدارة"
		if div != nil {
			divD = "نفس رقم إدارة " + d.DivisionName
		}
		// ٢. تقييمه الشهري
		var pers *float64
		persD, persEv := "ماكو تقييم شهري إله", []EvidenceLine{}
		if personal != nil {
			if f, lines := personal(d.EmployeeID, period); f != nil {
				pers = rankPtr(*f)
				persD = "تقييمه الشهري من ماتركس والتقييم البشري"
				persEv = lines
			}
		}
		// ٣. بنود الخطة الي برقبته
		var own *float64
		ownEv := planEvidence(db, `i."ownerId" = $1 AND p.period = $2`, d.EmployeeID, period)
		ownD := "ماكو بنود خطة برقبته"
		if len(ownEv) > 0 {
			var avg float64
			errlog.Err(db.Get(&avg, `SELECT COALESCE(avg(`+planItemScoreSQL+`), 0)::float8
				FROM "OrgPlanItem" i JOIN "OrgPlan" p ON p.id = i."planId" WHERE i."ownerId" = $1 AND p.period = $2`, d.EmployeeID, period))
			own = rankPtr(avg)
			ownD = strconv.Itoa(len(ownEv)) + " بند برقبته"
		}
		divEv := []EvidenceLine{}
		for _, dr := range directors {
			if dr.DivisionID == d.DivisionID {
				for _, c := range dr.Criteria {
					v := "—"
					if c.Score != nil {
						v = f1(*c.Score)
					}
					divEv = append(divEv, EvidenceLine{Label: c.Label + " (" + f1(c.Weight) + "٪)", Value: v, Note: c.Detail})
				}
			}
		}
		row.Criteria = []DirectorCriterion{
			{"division", "رقم الإدارة", 50, div, divD, "نفس الرقم النهائي لمدير الإدارة (تفاصيله تحت)", divEv},
			{"personal", "تقييمه الشهري", 30, pers, persD, "التقييم الشهري النهائي للنائب نفسه", persEv},
			{"ownplan", "بنوده بالخطة", 20, own, ownD, "معدل درجات بنود الخطة الي برقبته (خلص بموعده = ١٠٠، متأخر = −١٠ لكل يوم)", ownEv},
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
		row.DivisionTotal = div
		ow, owW := 0.0, 0.0
		for _, c := range row.Criteria[1:] {
			if c.Score != nil {
				ow += *c.Score * c.Weight
				owW += c.Weight
			}
		}
		if owW > 0 {
			row.OwnTotal = rankPtr(ow / owW)
		}
		out = append(out, row)
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i].Total, out[j].Total
		if a == nil || b == nil {
			return a != nil
		}
		return *a > *b
	})
	for i := range out {
		if t := out[i].Total; t != nil {
			if i > 0 && out[i-1].Total != nil && *out[i-1].Total == *t {
				out[i].Rank = out[i-1].Rank
			} else {
				out[i].Rank = i + 1
			}
		}
	}
	return out
}

// unitLeadersRanking (دازاي 11-06): مدير ونائب كل وحدة بنسبة مئوية وأسباب:
// ٥٠٪ بنود خطة الوحدة الي برقبته (بالالتزام بالموعد)، ٢٠٪ الخطة كاملة، ٣٠٪ تقييمه الشهري.
func unitLeadersRanking(db *sqlx.DB, period string, personal PersonalScoreFn) []DirectorRankRow {
	people := []struct {
		DivisionName string `db:"divisionName"`
		UnitID       string `db:"unitId"`
		UnitName     string `db:"unitName"`
		EmployeeID   string `db:"employeeId"`
		EmployeeName string `db:"employeeName"`
		Position     string `db:"position"`
		Acting       bool   `db:"acting"`
	}{}
	errlog.Err(db.Select(&people, `
		SELECT d.name AS "divisionName", u.id AS "unitId", u.name AS "unitName", a."employeeId", e.name AS "employeeName", a.position, a.acting
		FROM "OrgAssignment" a JOIN "OrgUnit" u ON u.id = a."unitId" JOIN "OrgDivision" d ON d.id = a."divisionId"
		JOIN "Employee" e ON e.id = a."employeeId"
		WHERE a."unitId" IS NOT NULL AND a.position IN ('DIRECTOR', 'DEPUTY') AND a."endedAt" IS NULL
		  AND e.role::text NOT IN ('ADMIN', 'OWNER')
		ORDER BY d.sort, u.sort, a.position`))
	out := []DirectorRankRow{}
	for _, p := range people {
		pos := "مدير"
		if p.Position == "DEPUTY" {
			pos = "نائب"
		}
		row := DirectorRankRow{DivisionName: p.DivisionName + " · " + p.UnitName + " (" + pos + ")", EmployeeID: p.EmployeeID, EmployeeName: p.EmployeeName, Acting: p.Acting}
		var own, unit, pers *float64
		ownEv := planEvidence(db, `i."ownerId" = $1 AND p.period = $2 AND p."unitId" = $3`, p.EmployeeID, period, p.UnitID)
		ownD := "ماكو بنود برقبته بخطة الوحدة"
		if len(ownEv) > 0 {
			var avg float64
			errlog.Err(db.Get(&avg, `SELECT COALESCE(avg(`+planItemScoreSQL+`), 0)::float8 FROM "OrgPlanItem" i JOIN "OrgPlan" p ON p.id = i."planId"
				WHERE i."ownerId" = $1 AND p.period = $2 AND p."unitId" = $3`, p.EmployeeID, period, p.UnitID))
			own = rankPtr(avg)
			ownD = strconv.Itoa(len(ownEv)) + " بند برقبته"
		}
		unitEv := planEvidence(db, `p."unitId" = $1 AND p.period = $2`, p.UnitID, period)
		unitD := "ماكو خطة للوحدة هالشهر"
		if len(unitEv) > 0 {
			var avg float64
			errlog.Err(db.Get(&avg, `SELECT COALESCE(avg(`+planItemScoreSQL+`), 0)::float8 FROM "OrgPlanItem" i JOIN "OrgPlan" p ON p.id = i."planId"
				WHERE p."unitId" = $1 AND p.period = $2`, p.UnitID, period))
			unit = rankPtr(avg)
			unitD = strconv.Itoa(len(unitEv)) + " بند بخطة الوحدة"
		}
		persD, persEv := "ماكو تقييم شهري إله", []EvidenceLine{}
		if personal != nil {
			if f, lines := personal(p.EmployeeID, period); f != nil {
				pers = rankPtr(*f)
				persD = "تقييمه الشهري"
				persEv = lines
			}
		}
		row.Criteria = []DirectorCriterion{
			{"ownplan", "بنوده بخطة الوحدة", 50, own, ownD, "معدل درجات بنوده: خلص بموعده = ١٠٠، متأخر = −١٠ لكل يوم، فات موعده وما خلص = نسبته − ١٠ لكل يوم", ownEv},
			{"unitplan", "خطة الوحدة كاملة", 20, unit, unitD, "معدل درجات كل بنود خطة الوحدة بنفس القاعدة", unitEv},
			{"personal", "تقييمه الشهري", 30, pers, persD, "التقييم الشهري النهائي إله", persEv},
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
		out = append(out, row)
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i].Total, out[j].Total
		if a == nil || b == nil {
			return a != nil
		}
		return *a > *b
	})
	for i := range out {
		if t := out[i].Total; t != nil {
			if i > 0 && out[i-1].Total != nil && *out[i-1].Total == *t {
				out[i].Rank = out[i-1].Rank
			} else {
				out[i].Rank = i + 1
			}
		}
	}
	return out
}
