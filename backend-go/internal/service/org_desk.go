package service

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"staffmange-api/internal/model"
)

// ═══ 🏛️ مكتب المدير (دازاي 10-29) ═══
// «كل مدير يقدّم تقرير المشاكل ومضارها على العمل، ويقدّم خطة تحسين بنسبة إنجاز
// ويا ماتركس — لازم يكون تقدّم مرئي». النائب يشارك الرأي (ملاحظات).
// البند المربوط بمقياس: ماتركس يحسب نسبته من أرقام النظام، مو بالكلام.

// ── المقاييس ──

type OrgMetric struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	LowerBetter bool   `json:"lowerBetter"`
	Unit        string `json:"unit"`
}

var OrgMetrics = []OrgMetric{
	{Key: "members_score", Label: "معدّل تقييم ماتركس لكادر الإدارة", Unit: "%"},
	{Key: "late_stations", Label: "نسبة المحطات المتأخرة أو الفايتة لكادر الإدارة", LowerBetter: true, Unit: "%"},
	{Key: "complaints_open", Label: "شكاوي مفتوحة على كادر الإدارة", LowerBetter: true, Unit: "شكوى"},
	{Key: "issues_open", Label: "مشاكل مفتوحة بتقارير الإدارة", LowerBetter: true, Unit: "مشكلة"},
}

func orgMetric(key string) (OrgMetric, bool) {
	for _, m := range OrgMetrics {
		if m.Key == key {
			return m, true
		}
	}
	return OrgMetric{}, false
}

var baghdad = time.FixedZone("Asia/Baghdad", 3*3600)

// OrgPeriod الشهر الحالي بتوقيت بغداد YYYY-MM.
func OrgPeriod(t time.Time) string { return t.In(baghdad).Format("2006-01") }

func periodRange(period string) (time.Time, time.Time, error) {
	start, err := time.ParseInLocation("2006-01", period, baghdad)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("الفترة لازم بشكل YYYY-MM")
	}
	return start, start.AddDate(0, 1, 0), nil
}

// كادر الإدارة بالفترة: كل من چان بيها بأي وكت من الشهر (مو بس الحاليين)
const orgMembersSQL = `SELECT DISTINCT "employeeId" FROM "OrgAssignment" WHERE "divisionId" = $1
	AND "startedAt" < $3 AND ("endedAt" IS NULL OR "endedAt" >= $2)`

// orgSnapshot مقاييس لحظية (بلا تاريخ) — تنحسب للشهر الحالي بس.
var orgSnapshot = map[string]bool{"complaints_open": true, "issues_open": true}

// orgCumulative مقاييس تتجمع خلال الشهر — خط أساسها الشهر الي فات كامل.
var orgCumulative = map[string]bool{"members_score": true, "late_stations": true}

var ErrOrgNoData = errors.New("ماكو بيانات كافية لهالشهر")

// MetricValue قيمة مقياس لإدارة بفترة. ErrOrgNoData إذا ماكو أرقام (فالنسبة ترجع يدوية).
func (s *OrgService) MetricValue(key, divisionID, period string) (float64, error) {
	from, to, err := periodRange(period)
	if err != nil {
		return 0, err
	}
	if orgSnapshot[key] && period != OrgPeriod(time.Now()) {
		return 0, ErrOrgNoData
	}
	var v float64
	switch key {
	case "members_score":
		var r struct {
			P float64 `db:"p"`
			M float64 `db:"m"`
		}
		err = s.db.Get(&r, `SELECT COALESCE(sum(points), 0) AS p, COALESCE(sum("maxPoints"), 0) AS m FROM "MatrixScore"
			WHERE "employeeId" IN (`+orgMembersSQL+`) AND "cancelledAt" IS NULL AND "at" >= $2 AND "at" < $3`, divisionID, from, to)
		if err == nil && r.M == 0 {
			return 0, ErrOrgNoData
		}
		if r.M > 0 {
			v = math.Round(r.P/r.M*1000) / 10
		}
	case "late_stations":
		// نسبة مو عدد — العدد يصعد طول الشهر حتى لو الشغل تحسّن
		var r struct {
			Late float64 `db:"late"`
			All  float64 `db:"total"`
		}
		err = s.db.Get(&r, `SELECT count(*) FILTER (WHERE points < "maxPoints") AS late, count(*) AS total FROM "MatrixScore"
			WHERE "employeeId" IN (`+orgMembersSQL+`) AND "cancelledAt" IS NULL AND "at" >= $2 AND "at" < $3`, divisionID, from, to)
		if err == nil && r.All == 0 {
			return 0, ErrOrgNoData
		}
		if r.All > 0 {
			v = math.Round(r.Late/r.All*1000) / 10
		}
	case "complaints_open":
		err = s.db.Get(&v, `SELECT count(*) FROM "Complaint" WHERE status IN ('NEW', 'IN_PROGRESS')
			AND "assignedToEmployeeId" IN (`+orgMembersSQL+`)`, divisionID, from, to)
	case "issues_open":
		err = s.db.Get(&v, `SELECT count(*) FROM "OrgIssueReport" WHERE "divisionId" = $1 AND status <> 'RESOLVED'`, divisionID)
	default:
		return 0, errors.New("مقياس غير معروف")
	}
	return v, err
}

// ── التقارير والخطة ──

type OrgIssue struct {
	ID         string     `db:"id" json:"id"`
	DivisionID string     `db:"divisionId" json:"divisionId"`
	UnitID     *string    `db:"unitId" json:"unitId"`
	AuthorID   *string    `db:"authorId" json:"authorId"`
	AuthorName *string    `db:"authorName" json:"authorName"`
	Title      string     `db:"title" json:"title"`
	Problem    string     `db:"problem" json:"problem"`
	Impact     string     `db:"impact" json:"impact"`
	Scope      string     `db:"scope" json:"scope"`
	Principle  *int       `db:"principle" json:"principle"`
	Severity   string     `db:"severity" json:"severity"`
	Status     string     `db:"status" json:"status"`
	Resolution *string    `db:"resolution" json:"resolution"`
	CreatedAt  time.Time  `db:"createdAt" json:"createdAt"`
	ResolvedAt *time.Time `db:"resolvedAt" json:"resolvedAt"`
}

type OrgPlanItem struct {
	ID             string     `db:"id" json:"id"`
	PlanID         string     `db:"planId" json:"planId"`
	Kind           string     `db:"kind" json:"kind"`
	Title          string     `db:"title" json:"title"`
	IssueID        *string    `db:"issueId" json:"issueId"`
	Metric         *string    `db:"metric" json:"metric"`
	Baseline       *float64   `db:"baseline" json:"baseline"`
	Target         *float64   `db:"target" json:"target"`
	ManualProgress int        `db:"manualProgress" json:"manualProgress"`
	DoneAt         *time.Time `db:"doneAt" json:"doneAt"`
	// محسوبة
	Current  *float64 `db:"-" json:"current"`
	Progress int      `db:"-" json:"progress"`
	ByMatrix bool     `db:"-" json:"byMatrix"`
}

type OrgPlan struct {
	ID         string        `db:"id" json:"id"`
	DivisionID string        `db:"divisionId" json:"divisionId"`
	Period     string        `db:"period" json:"period"`
	Title      string        `db:"title" json:"title"`
	Items      []OrgPlanItem `db:"-" json:"items"`
	Progress   int           `db:"-" json:"progress"`
}

type OrgComment struct {
	ID         string    `db:"id" json:"id"`
	TargetType string    `db:"targetType" json:"targetType"`
	TargetID   string    `db:"targetId" json:"targetId"`
	AuthorName *string   `db:"authorName" json:"authorName"`
	Text       string    `db:"text" json:"text"`
	CreatedAt  time.Time `db:"createdAt" json:"createdAt"`
}

type OrgAuthorityRow struct {
	ID       string   `db:"id" json:"id"`
	Decision string   `db:"decision" json:"decision"`
	Mode     string   `db:"mode" json:"mode"`
	LimitIQD *float64 `db:"limitIQD" json:"limitIQD"`
	Sort     int      `db:"sort" json:"sort"`
}

type OrgDesk struct {
	Division  OrgChartDivision  `json:"division"`
	Period    string            `json:"period"`
	CanEdit   bool              `json:"canEdit"`
	Issues    []OrgIssue        `json:"issues"`
	Plan      *OrgPlan          `json:"plan"`
	Comments  []OrgComment      `json:"comments"`
	Authority []OrgAuthorityRow `json:"authority"`
	Metrics   []OrgMetric       `json:"metrics"`
}

// ── الوصول ──
// يشوف المكتب: المالك ومدير النظام، مدير/نائب الإدارة، مدير/نائب أي وحدة بيها،
// ومدير/نائب «الإدارة» (يراقب الكل). يعدّل الخطة والصلاحيات: مدير/نائب الإدارة + المالك/مدير النظام.
// الي ما يحق له: نرجّع «مو موجود» (مو 403) — حتى ما تنسجل مخالفة على فضول.

func (s *OrgService) Access(empID, role, divisionID string) (see, edit bool) {
	if role == "OWNER" || role == "ADMIN" {
		return true, true
	}
	rows := []struct {
		Div  string  `db:"divisionId"`
		Unit *string `db:"unitId"`
	}{}
	_ = s.db.Select(&rows, `SELECT "divisionId", "unitId" FROM "OrgAssignment" WHERE "employeeId" = $1
		AND "endedAt" IS NULL AND position IN ('DIRECTOR', 'DEPUTY')`, empID)
	for _, r := range rows {
		if r.Div == "admin" && r.Unit == nil {
			see = true
		}
		if r.Div == divisionID {
			see = true
			if r.Unit == nil {
				edit = true
			}
		}
	}
	return see, edit
}

var ErrOrgHidden = errors.New("مو موجود")

func (s *OrgService) Desk(empID, role, divisionID, period string) (*OrgDesk, error) {
	see, edit := s.Access(empID, role, divisionID)
	if !see {
		return nil, ErrOrgHidden
	}
	if period == "" {
		period = OrgPeriod(time.Now())
	}
	if _, _, err := periodRange(period); err != nil {
		return nil, err
	}
	chart, err := s.Chart()
	if err != nil {
		return nil, err
	}
	desk := &OrgDesk{Period: period, CanEdit: edit, Metrics: OrgMetrics, Issues: []OrgIssue{}, Comments: []OrgComment{}, Authority: []OrgAuthorityRow{}}
	found := false
	for _, d := range chart {
		if d.ID == divisionID {
			desk.Division, found = d, true
		}
	}
	if !found {
		return nil, ErrOrgHidden
	}
	if err := s.db.Select(&desk.Issues, `SELECT i.id, i."divisionId", i."unitId", i."authorId", e.name AS "authorName", i.title, i.problem,
		i.impact, i.scope, i.principle, i.severity, i.status, i.resolution, i."createdAt", i."resolvedAt"
		FROM "OrgIssueReport" i LEFT JOIN "Employee" e ON e.id = i."authorId"
		WHERE i."divisionId" = $1 ORDER BY (i.status = 'RESOLVED'), CASE i.severity WHEN 'HIGH' THEN 0 WHEN 'MEDIUM' THEN 1 ELSE 2 END, i."createdAt" DESC
		LIMIT 200`, divisionID); err != nil {
		return nil, err
	}
	if desk.Plan, err = s.plan(divisionID, period); err != nil {
		return nil, err
	}
	if err := s.db.Select(&desk.Comments, `SELECT c.id, c."targetType", c."targetId", e.name AS "authorName", c.text, c."createdAt"
		FROM "OrgComment" c LEFT JOIN "Employee" e ON e.id = c."authorId"
		WHERE (c."targetType" = 'ISSUE' AND c."targetId" IN (SELECT id FROM "OrgIssueReport" WHERE "divisionId" = $1))
		   OR (c."targetType" = 'PLAN_ITEM' AND c."targetId" IN (SELECT i.id FROM "OrgPlanItem" i JOIN "OrgPlan" p ON p.id = i."planId" WHERE p."divisionId" = $1))
		ORDER BY c."createdAt"`, divisionID); err != nil {
		return nil, err
	}
	if err := s.db.Select(&desk.Authority, `SELECT id, decision, mode, "limitIQD", sort FROM "OrgAuthority" WHERE "divisionId" = $1 ORDER BY sort, decision`, divisionID); err != nil {
		return nil, err
	}
	return desk, nil
}

func (s *OrgService) plan(divisionID, period string) (*OrgPlan, error) {
	var p OrgPlan
	if err := s.db.Get(&p, `SELECT id, "divisionId", period, title FROM "OrgPlan" WHERE "divisionId" = $1 AND period = $2`, divisionID, period); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // ماكو خطة لهالشهر بعد
		}
		return nil, err
	}
	p.Items = []OrgPlanItem{}
	if err := s.db.Select(&p.Items, `SELECT id, "planId", kind, title, "issueId", metric, baseline, target, "manualProgress", "doneAt"
		FROM "OrgPlanItem" WHERE "planId" = $1 ORDER BY "createdAt"`, p.ID); err != nil {
		return nil, err
	}
	sum := 0
	for i := range p.Items {
		it := &p.Items[i]
		it.Progress = it.ManualProgress
		if it.Metric != nil && *it.Metric != "" {
			if cur, err := s.MetricValue(*it.Metric, divisionID, period); err == nil {
				it.Current = &cur
				if it.Baseline != nil && it.Target != nil && *it.Target != *it.Baseline {
					it.ByMatrix = true
					pr := (cur - *it.Baseline) / (*it.Target - *it.Baseline) * 100
					it.Progress = int(math.Round(math.Max(0, math.Min(100, pr))))
				}
			}
		}
		if it.DoneAt != nil {
			it.Progress = 100
		}
		sum += it.Progress
	}
	if len(p.Items) > 0 {
		p.Progress = int(math.Round(float64(sum) / float64(len(p.Items))))
	}
	return &p, nil
}

// ── الكتابة ──

type OrgIssueInput struct {
	DivisionID string `json:"divisionId"`
	UnitID     string `json:"unitId"`
	Title      string `json:"title"`
	Problem    string `json:"problem"`
	Impact     string `json:"impact"`
	Scope      string `json:"scope"`
	Principle  *int   `json:"principle"`
	Severity   string `json:"severity"`
}

func (s *OrgService) AddIssue(empID, role string, in OrgIssueInput) (string, error) {
	if see, _ := s.Access(empID, role, in.DivisionID); !see {
		return "", ErrOrgHidden
	}
	in.Title, in.Problem = strings.TrimSpace(in.Title), strings.TrimSpace(in.Problem)
	if len([]rune(in.Title)) < 3 || len([]rune(in.Problem)) < 10 {
		return "", errors.New("اكتب عنوان المشكلة وشرحها (١٠ حروف على الأقل)")
	}
	if in.Scope == "" {
		in.Scope = "SPECIFIC"
	}
	if in.Severity == "" {
		in.Severity = "MEDIUM"
	}
	if in.Principle != nil && (*in.Principle < 1 || *in.Principle > 20) {
		return "", errors.New("رقم المبدأ من ١ لـ٢٠")
	}
	var unit any
	if in.UnitID != "" {
		var div string
		if err := s.db.Get(&div, `SELECT "divisionId" FROM "OrgUnit" WHERE id = $1`, in.UnitID); err != nil || div != in.DivisionID {
			return "", errors.New("الوحدة مو تابعة لهالإدارة")
		}
		unit = in.UnitID
	}
	var id string
	err := s.db.Get(&id, `INSERT INTO "OrgIssueReport" ("divisionId", "unitId", "authorId", title, problem, impact, scope, principle, severity)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`,
		in.DivisionID, unit, empID, in.Title, in.Problem, strings.TrimSpace(in.Impact), in.Scope, in.Principle, in.Severity)
	if err != nil && strings.Contains(err.Error(), "check constraint") {
		return "", errors.New("قيمة غير صالحة بالنطاق أو الخطورة")
	}
	return id, err
}

func (s *OrgService) issueDivision(id string) string {
	var d string
	_ = s.db.Get(&d, `SELECT "divisionId" FROM "OrgIssueReport" WHERE id = $1`, id)
	return d
}

// UpdateIssue يغيّر الحالة/الحل — مدير/نائب الإدارة أو المالك/مدير النظام.
func (s *OrgService) UpdateIssue(empID, role, id, status, resolution string) error {
	if _, edit := s.Access(empID, role, s.issueDivision(id)); !edit {
		return ErrOrgHidden
	}
	if status != "OPEN" && status != "IN_PROGRESS" && status != "RESOLVED" {
		return errors.New("الحالة غير صالحة")
	}
	resolution = strings.TrimSpace(resolution)
	if status == "RESOLVED" && len([]rune(resolution)) < 5 {
		return errors.New("اكتب شلون انحلت المشكلة")
	}
	_, err := s.db.Exec(`UPDATE "OrgIssueReport" SET status = $2, resolution = NULLIF($3, ''),
		"resolvedAt" = CASE WHEN $2 = 'RESOLVED' THEN COALESCE("resolvedAt", now()) ELSE NULL END WHERE id = $1`, id, status, resolution)
	return err
}

func (s *OrgService) CreatePlan(empID, role, divisionID, period, title string) (string, error) {
	if _, edit := s.Access(empID, role, divisionID); !edit {
		return "", ErrOrgHidden
	}
	if period == "" {
		period = OrgPeriod(time.Now())
	}
	if _, _, err := periodRange(period); err != nil {
		return "", err
	}
	var id string
	err := s.db.Get(&id, `INSERT INTO "OrgPlan" ("divisionId", period, title, "createdById") VALUES ($1, $2, $3, $4)
		ON CONFLICT ("divisionId", period) DO UPDATE SET title = CASE WHEN EXCLUDED.title <> '' THEN EXCLUDED.title ELSE "OrgPlan".title END
		RETURNING id`, divisionID, period, strings.TrimSpace(title), empID)
	return id, err
}

func (s *OrgService) planDivision(planID string) (div, period string) {
	var r struct {
		D string `db:"divisionId"`
		P string `db:"period"`
	}
	_ = s.db.Get(&r, `SELECT "divisionId", period FROM "OrgPlan" WHERE id = $1`, planID)
	return r.D, r.P
}

type OrgPlanItemInput struct {
	Kind    string   `json:"kind"`
	Title   string   `json:"title"`
	IssueID string   `json:"issueId"`
	Metric  string   `json:"metric"`
	Target  *float64 `json:"target"`
}

// AddPlanItem — إذا مربوط بمقياس، خط الأساس ينحسب هسه من النظام (ماتركس) مو من الإدخال.
func (s *OrgService) AddPlanItem(empID, role, planID string, in OrgPlanItemInput) (string, error) {
	div, period := s.planDivision(planID)
	if _, edit := s.Access(empID, role, div); !edit || div == "" {
		return "", ErrOrgHidden
	}
	in.Title = strings.TrimSpace(in.Title)
	if len([]rune(in.Title)) < 3 {
		return "", errors.New("اكتب البند")
	}
	if in.Kind != "PROBLEM" && in.Kind != "DEVELOPMENT" {
		return "", errors.New("نوع البند: حل مشكلة أو تطوير")
	}
	var metric, baseline, target any
	if in.Metric != "" {
		m, ok := orgMetric(in.Metric)
		if !ok {
			return "", errors.New("مقياس غير معروف")
		}
		if in.Target == nil {
			return "", errors.New("حدد الهدف للمقياس")
		}
		if *in.Target < 0 || (m.Unit == "%" && *in.Target > 100) {
			return "", errors.New("هدف غير منطقي")
		}
		if period > OrgPeriod(time.Now()) {
			return "", errors.New("البند المربوط بمقياس ينضاف من يبدي شهره")
		}
		// خط الأساس: المقاييس المتجمّعة من الشهر الي فات كامل، واللحظية من هسه
		basePeriod := period
		if orgCumulative[in.Metric] {
			start, _, _ := periodRange(period)
			basePeriod = start.AddDate(0, -1, 0).Format("2006-01")
		}
		base, err := s.MetricValue(in.Metric, div, basePeriod)
		if errors.Is(err, ErrOrgNoData) {
			return "", errors.New("ماكو أرقام سابقة لهالمقياس بهالإدارة — خلي البند بنسبة يدوية")
		}
		if err != nil {
			return "", err
		}
		if (m.LowerBetter && *in.Target >= base) || (!m.LowerBetter && *in.Target <= base) {
			return "", fmt.Errorf("الهدف لازم يكون أحسن من الرقم الحالي (%v)", base)
		}
		metric, baseline, target = in.Metric, base, *in.Target
	}
	var issue any
	if in.IssueID != "" {
		if s.issueDivision(in.IssueID) != div {
			return "", errors.New("المشكلة مو من هالإدارة")
		}
		issue = in.IssueID
	}
	var id string
	err := s.db.Get(&id, `INSERT INTO "OrgPlanItem" ("planId", kind, title, "issueId", metric, baseline, target, "createdById")
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`, planID, in.Kind, in.Title, issue, metric, baseline, target, empID)
	return id, err
}

// UpdatePlanItem نسبة يدوية (للبنود بلا مقياس) أو «خلص».
func (s *OrgService) UpdatePlanItem(empID, role, itemID string, manual *int, done *bool) error {
	var planID string
	_ = s.db.Get(&planID, `SELECT "planId" FROM "OrgPlanItem" WHERE id = $1`, itemID)
	div, _ := s.planDivision(planID)
	if _, edit := s.Access(empID, role, div); !edit || div == "" {
		return ErrOrgHidden
	}
	if manual != nil {
		if *manual < 0 || *manual > 100 {
			return errors.New("النسبة من ٠ لـ١٠٠")
		}
		if _, err := s.db.Exec(`UPDATE "OrgPlanItem" SET "manualProgress" = $2 WHERE id = $1`, itemID, *manual); err != nil {
			return err
		}
	}
	if done != nil {
		if _, err := s.db.Exec(`UPDATE "OrgPlanItem" SET "doneAt" = CASE WHEN $2 THEN COALESCE("doneAt", now()) ELSE NULL END WHERE id = $1`, itemID, *done); err != nil {
			return err
		}
	}
	return nil
}

// AddComment رأي النائب (أو أي واحد يشوف المكتب) على مشكلة أو بند.
func (s *OrgService) AddComment(empID, role, targetType, targetID, text string) error {
	var div string
	switch targetType {
	case "ISSUE":
		div = s.issueDivision(targetID)
	case "PLAN_ITEM":
		var planID string
		_ = s.db.Get(&planID, `SELECT "planId" FROM "OrgPlanItem" WHERE id = $1`, targetID)
		div, _ = s.planDivision(planID)
	default:
		return errors.New("نوع غير صالح")
	}
	if see, _ := s.Access(empID, role, div); !see || div == "" {
		return ErrOrgHidden
	}
	text = strings.TrimSpace(text)
	if len([]rune(text)) < 2 {
		return errors.New("اكتب الملاحظة")
	}
	_, err := s.db.Exec(`INSERT INTO "OrgComment" ("targetType", "targetId", "authorId", text) VALUES ($1, $2, $3, $4)`, targetType, targetID, empID, text)
	return err
}

// SetAuthority يبدّل جدول صلاحيات إدارة (المالك ومدير النظام بس — الحارس بالراوت).
func (s *OrgService) SetAuthority(empID, divisionID string, rows []OrgAuthorityRow) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var n int
	if err := tx.Get(&n, `SELECT count(*) FROM "OrgDivision" WHERE id = $1`, divisionID); err != nil || n == 0 {
		return errors.New("الإدارة مو موجودة")
	}
	if _, err := tx.Exec(`DELETE FROM "OrgAuthority" WHERE "divisionId" = $1`, divisionID); err != nil {
		return err
	}
	for i, r := range rows {
		r.Decision = strings.TrimSpace(r.Decision)
		if r.Decision == "" {
			continue
		}
		if r.Mode != "DECIDE" && r.Mode != "ESCALATE" {
			return errors.New("كل قرار: يقرر لحاله أو يرفعه")
		}
		if r.LimitIQD != nil {
			if *r.LimitIQD < 0 {
				return errors.New("السقف ما يكون سالب")
			}
			if err := model.CheckIQD("سقف «"+r.Decision+"»", *r.LimitIQD); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(`INSERT INTO "OrgAuthority" ("divisionId", decision, mode, "limitIQD", sort, "updatedById") VALUES ($1, $2, $3, $4, $5, $6)`,
			divisionID, r.Decision, r.Mode, r.LimitIQD, i, empID); err != nil {
			return err
		}
	}
	return tx.Commit()
}
