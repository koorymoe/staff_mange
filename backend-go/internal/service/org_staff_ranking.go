package service

import (
	"math"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/errlog"
)

// ═══ تصنيف الموظفين بنفس طريقة المدراء والنواب (دازاي 11-13) ═══
// «كل موظف يتقيم بنفس الطريقه مال المدراء والنواب — مخرجات ومدخلات، بس ماعده خطه
// يطلعله اتباع الخطة او التعليمات… كل موظف يتبع المدير مالته».
// نفس السبع معايير ونفس الأوزان، بس على شغل الموظف نفسه:
//   المخرجات ٣٠ ← تقييمه الشهري (ماتركس)
//   اتباع الخطة والتعليمات ٢٠ ← بنود الخطة الي برقبته + المهام الي انطاها إياه مديره/الإدارة
//   السرعة ١٥ ← من محطات شغله الي خلصت: كم وحدة بوقتها
//   المشاكل والشغل الفايت ١٠ ← محطاته الي فاتت وما صارت
//   شكد انحل منها ١٥ ← الفايت الي رجع سوّاه بعدين
//   الشكاوي والتأخير ١٠ ← الشكاوي عليه + مهامه المتأخرة
//   الانضباط بالدوام ١٠ ← انضباطه هو
// المدراء والنواب (إدارة/وحدة) بقوائمهم؛ الأدمن والمالك ما ينقيّمون.

// followItemSQL درجة مهمة (ExtraTask): بموعدها ١٠٠، متأخرة −١٠ لكل يوم (أقل شي ٤٠)،
// فات موعدها وما خلصت −١٠ لكل يوم من ٦٠. الي بعد موعدها ما اجى وما خلصت ما تنحسب.
const followTaskScoreSQL = `CASE
	WHEN t."doneAt" IS NOT NULL THEN GREATEST(40, 100 - 10 * GREATEST(0, baghdad_date(t."doneAt") - baghdad_date(t."dueAt")))
	ELSE GREATEST(0, 60 - 10 * (baghdad_today() - baghdad_date(t."dueAt"))) END`

func staffRanking(db *sqlx.DB, period string, personal PersonalScoreFn) []DirectorRankRow {
	from, to, err := periodRange(period)
	if err != nil {
		return nil
	}
	people := []struct {
		EmployeeID   string `db:"employeeId"`
		EmployeeName string `db:"employeeName"`
		DivisionName string `db:"divisionName"`
		Manager      string `db:"manager"`
	}{}
	// كل موظف فعّال مو مدير/نائب (إدارة أو وحدة)، ويا إدارته ومديره.
	errlog.Err(db.Select(&people, `
		SELECT e.id AS "employeeId", e.name AS "employeeName",
			COALESCE(d.name, 'بلا إدارة بالهيكل') AS "divisionName", COALESCE(m.name, '') AS manager
		FROM "Employee" e
		LEFT JOIN LATERAL (SELECT a."divisionId" FROM "OrgAssignment" a WHERE a."employeeId" = e.id AND a."endedAt" IS NULL
			ORDER BY a."startedAt" LIMIT 1) ma ON true
		LEFT JOIN "OrgDivision" d ON d.id = ma."divisionId"
		LEFT JOIN LATERAL (SELECT me.name FROM "OrgAssignment" x JOIN "Employee" me ON me.id = x."employeeId"
			WHERE x."divisionId" = ma."divisionId" AND x."unitId" IS NULL AND x.position = 'DIRECTOR' AND x."endedAt" IS NULL
			ORDER BY x.acting, x."startedAt" LIMIT 1) m ON true
		WHERE e.status = 'ACTIVE' AND e.role::text NOT IN ('ADMIN', 'OWNER')
		  AND NOT EXISTS (SELECT 1 FROM "OrgAssignment" a WHERE a."employeeId" = e.id AND a."endedAt" IS NULL AND a.position IN ('DIRECTOR', 'DEPUTY'))
		ORDER BY e.name`))
	out := make([]DirectorRankRow, len(people))
	sem := make(chan struct{}, 8)
	var wg sync.WaitGroup
	for i, p := range people {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, id, name, div, mgr string) {
			defer wg.Done()
			defer func() { <-sem }()
			out[i] = staffRow(db, period, from, to, id, name, div, mgr, personal)
		}(i, p.EmployeeID, p.EmployeeName, p.DivisionName, p.Manager)
	}
	wg.Wait()
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

func staffRow(db *sqlx.DB, period string, from, to time.Time, id, name, div, mgr string, personal PersonalScoreFn) DirectorRankRow {
	label := div
	if mgr != "" {
		label += " · مديره " + mgr
	}
	row := DirectorRankRow{DivisionName: label, EmployeeID: id, EmployeeName: name, People: 1}

	// ١. المخرجات
	var outputs *float64
	outD, outEv := "ماكو تقييم شهري إله", []EvidenceLine{}
	if personal != nil {
		if f, lines := personal(id, period); f != nil {
			outputs = rankPtr(*f)
			outD, outEv = "تقييمه الشهري (ماتركس)", lines
		}
	}

	// ٢. اتباع الخطة والتعليمات: بنود الخطة الي برقبته + المهام الي انطوه إياها
	var follow *float64
	followD := "ماكو خطة ولا تعليمات إله هالشهر"
	followEv := planEvidence(db, `i."ownerId" = $1 AND p.period = $2`, id, period)
	var fl struct {
		N   int     `db:"n"`
		Sum float64 `db:"sum"`
	}
	errlog.Err(db.Get(&fl, `SELECT count(*) AS n, COALESCE(sum(s), 0)::float8 AS sum FROM (
			SELECT (`+planItemScoreSQL+`)::float8 AS s FROM "OrgPlanItem" i JOIN "OrgPlan" p ON p.id = i."planId"
			WHERE i."ownerId" = $1 AND p.period = $2
			UNION ALL
			SELECT (`+followTaskScoreSQL+`)::float8 FROM "ExtraTask" t
			WHERE t."assignedToId" = $1 AND t."cancelledAt" IS NULL AND t."dueAt" >= $3 AND t."dueAt" < $4
			  AND (t."doneAt" IS NOT NULL OR t."dueAt" < now())) x`, id, period, from, to))
	tasks := []struct {
		Title string  `db:"title"`
		Score float64 `db:"score"`
		By    string  `db:"by"`
		Done  *string `db:"done"`
		Due   string  `db:"due"`
	}{}
	errlog.Err(db.Select(&tasks, `SELECT t.title, (`+followTaskScoreSQL+`)::float8 AS score, COALESCE(b.name, '') AS by,
			to_char(baghdad_date(t."doneAt"), 'YYYY-MM-DD') AS done, to_char(baghdad_date(t."dueAt"), 'YYYY-MM-DD') AS due
		FROM "ExtraTask" t LEFT JOIN "Employee" b ON b.id = t."assignedById"
		WHERE t."assignedToId" = $1 AND t."cancelledAt" IS NULL AND t."dueAt" >= $2 AND t."dueAt" < $3
		  AND (t."doneAt" IS NOT NULL OR t."dueAt" < now()) ORDER BY t."dueAt"`, id, from, to))
	for _, t := range tasks {
		note := "موعدها " + t.Due
		if t.Done != nil {
			note += " · خلصت " + *t.Done
		} else {
			note += " · بعدها ما خلصت"
		}
		if t.By != "" {
			note += " · من " + t.By
		}
		followEv = append(followEv, EvidenceLine{Label: "📋 " + t.Title, Value: f1(t.Score), Note: note})
	}
	if fl.N > 0 {
		follow = rankPtr(fl.Sum / float64(fl.N))
		followD = strconv.Itoa(fl.N) + " بند/مهمة (خطة + تعليمات)"
	}

	// ٣–٥. محطات شغله: بوقتها، متأخرة، فاتت
	var st struct {
		OnTime int `db:"ontime"`
		Late   int `db:"late"`
		Missed int `db:"missed"`
	}
	errlog.Err(db.Get(&st, `SELECT
			count(*) FILTER (WHERE s.points >= s."maxPoints" AND s."maxPoints" > 0) AS ontime,
			count(*) FILTER (WHERE s.points > 0 AND s.points < s."maxPoints") AS late,
			count(*) FILTER (WHERE s.points = 0) AS missed
		FROM "MatrixScore" s WHERE s."employeeId" = $1 AND s.source IN ('BOOKING','PROJECT','REPORT')
		  AND s."cancelledAt" IS NULL AND s."at" >= $2 AND s."at" < $3`, id, from, to))
	var speed, few, solve *float64
	speedD, solveD := "ماكو محطات خلصت هالشهر", "ماكو شغل فايت هالشهر"
	if d := st.OnTime + st.Late; d > 0 {
		speed = rankPtr(float64(st.OnTime) * 100 / float64(d))
		speedD = strconv.Itoa(st.OnTime) + " من " + strconv.Itoa(d) + " محطة خلصت بوقتها"
	}
	if st.OnTime+st.Late+st.Missed > 0 {
		few = rankPtr(100 - float64(st.Missed)*25)
	}
	fewD := strconv.Itoa(st.Missed) + " محطة فاتت وما صارت"
	if t := st.Missed + st.Late; t > 0 {
		solve = rankPtr(float64(st.Late) * 100 / float64(t))
		solveD = strconv.Itoa(st.Late) + " من " + strconv.Itoa(t) + " تأخرت ورجع سوّاها (" + strconv.Itoa(st.Missed) + " بعدها ما انسوّت)"
	}
	stEv := []EvidenceLine{
		{Label: "بوقتها", Value: strconv.Itoa(st.OnTime)},
		{Label: "متأخرة بس انسوّت", Value: strconv.Itoa(st.Late)},
		{Label: "فاتت وما انسوّت", Value: strconv.Itoa(st.Missed)},
	}

	// ٦. الشكاوي والتأخير
	var cp struct {
		Complaints int `db:"c"`
		LateTasks  int `db:"t"`
	}
	errlog.Err(db.Get(&cp, `SELECT
		(SELECT count(*) FROM "Complaint" c WHERE c."relatedEmployeeId" = $1 AND c."createdAt" >= $2 AND c."createdAt" < $3
			AND COALESCE(c."auditVerdict", '') <> 'EMPLOYEE_RIGHT') AS c,
		(SELECT count(*) FROM "ExtraTask" t WHERE t."assignedToId" = $1 AND t."cancelledAt" IS NULL AND t."dueAt" >= $2 AND t."dueAt" < $3
			AND ((t."doneAt" IS NOT NULL AND t."doneAt" > t."dueAt") OR (t."doneAt" IS NULL AND t."dueAt" < now()))) AS t`, id, from, to))
	complaints := rankPtr(100 - float64(cp.Complaints+cp.LateTasks)*20)
	compD := strconv.Itoa(cp.Complaints) + " شكوى · " + strconv.Itoa(cp.LateTasks) + " مهمة متأخرة"

	// ٧. الانضباط
	var disc *float64
	discD, discEv := "يبدي "+DisciplineStart.Format("2006-01-02"), []EvidenceLine{}
	if dm, err := Disciplines(db, []string{id}, from, to, time.Now()); err == nil {
		if d := dm[id]; d != nil && d.Score != nil {
			disc = rankPtr(*d.Score)
			discD = strconv.Itoa(d.Late) + " تأخير بلا عذر · " + strconv.Itoa(d.Absent) + " غياب"
			discEv = append(discEv, EvidenceLine{Label: "تأخير بلا عذر", Value: strconv.Itoa(d.Late), Note: strconv.Itoa(d.LateMinutes) + " دقيقة"},
				EvidenceLine{Label: "غياب", Value: strconv.Itoa(d.Absent)}, EvidenceLine{Label: "تأخير بعذر", Value: strconv.Itoa(d.LateExcused)})
		}
	}

	row.Criteria = []DirectorCriterion{
		{"outputs", "المخرجات", 30, outputs, outD, "تقييمه الشهري النهائي من ماتركس (شغل دوره)", outEv},
		{"follow", "اتباع الخطة والتعليمات", 20, follow, followD, "معدل درجات بنود الخطة الي برقبته والمهام الي انطاها إياه مديره: بموعدها = ١٠٠، متأخرة = −١٠ لكل يوم، فات موعدها وما خلصت = ٦٠ − ١٠ لكل يوم", followEv},
		{"speed", "السرعة", 15, speed, speedD, "محطات شغله الي خلصت بوقتها ÷ كل الي خلصت × ١٠٠", stEv},
		{"few", "المشاكل والشغل الفايت", 10, few, fewD, "١٠٠ − ٢٥ لكل محطة فاتت وما صارت", stEv},
		{"solve", "شكد انحل منها", 15, solve, solveD, "المحطات المتأخرة الي رجع سوّاها ÷ (المتأخرة + الي فاتت) × ١٠٠", stEv},
		{"complaints", "الشكاوي والتأخير", 10, complaints, compD, "١٠٠ − (الشكاوي عليه + مهامه المتأخرة) × ٢٠", nil},
		{"discipline", "الانضباط بالدوام", 10, disc, discD, "١٠٠ − ٢ لكل يوم تأخير بلا عذر − ٥ لكل غياب بلا إجازة", discEv},
	}
	sum, w, withData := 0.0, 0.0, 0
	for _, c := range row.Criteria {
		if c.Score != nil {
			sum += *c.Score * c.Weight
			w += c.Weight
			withData++
		}
	}
	if w > 0 {
		row.Total = rankPtr(sum / w)
	}
	row.Coverage = math.Round(float64(withData)*100/float64(len(row.Criteria))) / 100
	row.WeakEvidence = withData*2 < len(row.Criteria)
	return row
}
