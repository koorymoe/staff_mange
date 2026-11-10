package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"staffmange-api/internal/errlog"
)

// ═══ 🤖 خطة تحسين الوحدة بقلم ماتركس، والبنود تصير مهام (دازاي 11-06) ═══
// «خطه لكل وحدة… قوية ومدروسة، يكتبها ماتركس ويدرسها ويا المدير، ويسويها مهام للمدير
// والنائب، ويتقيمون عليها». ماتركس يكتب المسودة من أرقام الوحدة الحقيقية؛ المدير يعدّل
// ويرفع؛ مدير الإدارة (أو المالك) يعتمد؛ ومن تنعتمد كل بند يصير مهمة بموعده.

type planDraftItem struct {
	Kind   string   `json:"kind"`
	Title  string   `json:"title"`
	Reason string   `json:"reason"`
	Metric string   `json:"metric"`
	Target *float64 `json:"target"`
	Due    string   `json:"dueDate"`
	Owner  string   `json:"owner"` // DIRECTOR | DEPUTY
}

// unitLeaders مدير ونائب الوحدة (أو الإدارة إذا خطة إدارة).
func (s *OrgService) unitLeaders(div string, unit *string) (director, deputy string) {
	rows := []struct {
		Emp string `db:"employeeId"`
		Pos string `db:"position"`
	}{}
	errlog.Err(s.db.Select(&rows, `SELECT "employeeId", position FROM "OrgAssignment"
		WHERE "divisionId" = $1 AND COALESCE("unitId", '') = COALESCE($2, '') AND position IN ('DIRECTOR', 'DEPUTY') AND "endedAt" IS NULL
		ORDER BY acting`, div, unit))
	for _, r := range rows {
		if r.Pos == "DIRECTOR" && director == "" {
			director = r.Emp
		}
		if r.Pos == "DEPUTY" && deputy == "" {
			deputy = r.Emp
		}
	}
	return
}

// planFacts أرقام الوحدة الحقيقية لماتركس (بلا أسماء).
func (s *OrgService) planFacts(sc orgPlanScope) string {
	var b strings.Builder
	var divName, unitName string
	errlog.Err(s.db.Get(&divName, `SELECT name FROM "OrgDivision" WHERE id = $1`, sc.Div))
	if sc.Unit != nil {
		errlog.Err(s.db.Get(&unitName, `SELECT name FROM "OrgUnit" WHERE id = $1`, *sc.Unit))
	}
	fmt.Fprintf(&b, "الإدارة: %s · الوحدة: %s · الفترة: %s\n", divName, unitName, sc.Period)
	var people int
	errlog.Err(s.db.Get(&people, `SELECT count(*) FROM "OrgAssignment" WHERE "divisionId" = $1 AND COALESCE("unitId", '') = COALESCE($2, '') AND "endedAt" IS NULL`, sc.Div, sc.Unit))
	fmt.Fprintf(&b, "عدد كادر الوحدة: %d\n", people)
	b.WriteString("المقاييس المتاحة (ينحسب الوضع الحالي منها تلقائياً):\n")
	for _, m := range OrgMetrics {
		cur := "ماكو بيانات"
		if v, err := s.MetricValue(m.Key, sc.Div, sc.Period); err == nil {
			cur = fmt.Sprintf("%.1f %s", v, m.Unit)
		}
		dir := "أعلى أحسن"
		if m.LowerBetter {
			dir = "أقل أحسن"
		}
		fmt.Fprintf(&b, "- %s (%s): هسه %s، %s\n", m.Key, m.Label, cur, dir)
	}
	issues := []struct {
		Title string  `db:"title"`
		Sev   string  `db:"severity"`
		Sol   *string `db:"solutionPlan"`
	}{}
	errlog.Err(s.db.Select(&issues, `SELECT title, severity, "solutionPlan" FROM "OrgIssueReport"
		WHERE "divisionId" = $1 AND ($2::text IS NULL OR "unitId" = $2) AND status <> 'RESOLVED'
		ORDER BY CASE severity WHEN 'HIGH' THEN 0 WHEN 'MEDIUM' THEN 1 ELSE 2 END LIMIT 15`, sc.Div, sc.Unit))
	b.WriteString("المشاكل المفتوحة:\n")
	if len(issues) == 0 {
		b.WriteString("- ماكو\n")
	}
	for _, i := range issues {
		fmt.Fprintf(&b, "- [%s] %s\n", i.Sev, i.Title)
	}
	past := []struct {
		Title string  `db:"title"`
		Done  bool    `db:"done"`
		Due   *string `db:"due"`
	}{}
	errlog.Err(s.db.Select(&past, `SELECT i.title, i."doneAt" IS NOT NULL AS done, i."dueDate"::text AS due FROM "OrgPlanItem" i JOIN "OrgPlan" p ON p.id = i."planId"
		WHERE p."divisionId" = $1 AND COALESCE(p."unitId", '') = COALESCE($2, '') AND p.period < $3 ORDER BY p.period DESC LIMIT 15`, sc.Div, sc.Unit, sc.Period))
	if len(past) > 0 {
		b.WriteString("بنود الخطط السابقة:\n")
		for _, p := range past {
			st := "ما خلص"
			if p.Done {
				st = "خلص"
			}
			fmt.Fprintf(&b, "- %s (%s)\n", p.Title, st)
		}
	}
	return b.String()
}

// DraftPlanByMatrix ماتركس يكتب بنود المسودة. يرجّع عدد البنود الي انضافت.
func (s *OrgService) DraftPlanByMatrix(empID, role, planID, guidance string) (int, error) {
	if s.planDrafter == nil {
		return 0, errors.New("ماتركس مطفي هسه")
	}
	sc := s.planScope(planID)
	if !s.planEdit(empID, role, sc) {
		return 0, ErrOrgHidden
	}
	if sc.Status != "DRAFT" && sc.Status != "RETURNED" {
		return 0, errors.New("المسودة تنكتب بس للخطة الي بعدها مسودة أو راجعة للتعديل")
	}
	start, end, _ := periodRange(sc.Period)
	system := "إنت ماتركس، مستشار تطوير الوحدات بشركة الأماني. تكتب خطة تحسين قوية ومدروسة لوحدة وحدة، من أرقامها الحقيقية بس، وبنودها قابلة للقياس والتنفيذ خلال الشهر. ممنوع تخترع أرقام."
	task := s.planFacts(sc) + fmt.Sprintf(`
اكتب ٤ إلى ٧ بنود للخطة. كل بند:
- kind: PROBLEM (يحل مشكلة مفتوحة) أو DEVELOPMENT (تطوير)
- title: شنو ينسوّى بالضبط، جملة عملية قصيرة
- reason: ليش هذا البند (مربوط برقم أو مشكلة من فوق)
- metric و target: إذا البند ينقاس بمقياس من القائمة فوق (الهدف أحسن من الوضع الحالي)، وإلا فاضيين
- dueDate: موعد بين %s و %s (YYYY-MM-DD)، والبنود موزعة على الشهر
- owner: DIRECTOR أو DEPUTY — وزّع الشغل بينهم بعدل حسب طبيعة البند
توجيه المدير: %s
جاوب **بمصفوفة JSON بس** بلا أي كلام ثاني.`, start.Format("2006-01-02"), end.AddDate(0, 0, -1).Format("2006-01-02"), strings.TrimSpace(guidance))
	out, err := s.planDrafter(system, task)
	if err != nil {
		return 0, err
	}
	i, j := strings.Index(out, "["), strings.LastIndex(out, "]")
	if i < 0 || j <= i {
		return 0, errors.New("ماتركس ما رجّع خطة مفهومة — جرّب مرة ثانية")
	}
	var items []planDraftItem
	if err := json.Unmarshal([]byte(out[i:j+1]), &items); err != nil {
		return 0, errors.New("ماتركس ما رجّع خطة مفهومة — جرّب مرة ثانية")
	}
	director, deputy := s.unitLeaders(sc.Div, sc.Unit)
	n := 0
	for _, it := range items {
		if n >= 8 {
			break
		}
		owner := director
		if strings.EqualFold(it.Owner, "DEPUTY") && deputy != "" {
			owner = deputy
		}
		if it.Kind != "PROBLEM" {
			it.Kind = "DEVELOPMENT"
		}
		in := OrgPlanItemInput{Kind: it.Kind, Title: it.Title, Metric: it.Metric, Target: it.Target, DueDate: it.Due, OwnerID: owner}
		id, err := s.AddPlanItem(empID, role, planID, in)
		if err != nil && in.Metric != "" { // المقياس ما ينطبق: البند يبقى بنسبة يدوية
			in.Metric, in.Target = "", nil
			id, err = s.AddPlanItem(empID, role, planID, in)
		}
		if err != nil && in.DueDate != "" { // موعد غلط: بلا موعد
			in.DueDate = ""
			id, err = s.AddPlanItem(empID, role, planID, in)
		}
		if err != nil {
			continue
		}
		errlog.Exec(s.db.Exec(`UPDATE "OrgPlanItem" SET "byMatrix" = true, reason = NULLIF($2, '') WHERE id = $1`, id, strings.TrimSpace(it.Reason)))
		n++
	}
	if n == 0 {
		return 0, errors.New("ما انضاف ولا بند — جرّب مرة ثانية")
	}
	return n, nil
}

// planToTasks (بعد الاعتماد) كل بند بمسؤول وموعد بالمستقبل يصير مهمة بمهام النظام.
func (s *OrgService) planToTasks(planID, approverID string) int {
	if s.taskCreator == nil {
		return 0
	}
	sc := s.planScope(planID)
	director, _ := s.unitLeaders(sc.Div, sc.Unit)
	items := []struct {
		ID     string  `db:"id"`
		Title  string  `db:"title"`
		Reason *string `db:"reason"`
		Owner  *string `db:"ownerId"`
		Due    *string `db:"due"`
	}{}
	errlog.Err(s.db.Select(&items, `SELECT id, title, reason, "ownerId", "dueDate"::text AS due FROM "OrgPlanItem"
		WHERE "planId" = $1 AND "taskId" IS NULL AND "doneAt" IS NULL`, planID))
	n := 0
	for _, it := range items {
		to := director
		if it.Owner != nil {
			to = *it.Owner
		}
		if to == "" {
			continue
		}
		due := time.Now().In(baghdad).AddDate(0, 0, 7)
		if it.Due != nil {
			if d, err := time.ParseInLocation("2006-01-02", *it.Due, baghdad); err == nil {
				due = d.Add(23 * time.Hour)
			}
		}
		if !due.After(time.Now()) {
			due = time.Now().Add(48 * time.Hour)
		}
		desc := "بند من خطة تحسين الوحدة المعتمدة (" + sc.Period + ")."
		if it.Reason != nil {
			desc += " ليش: " + *it.Reason
		}
		taskID, err := s.taskCreator("📋 خطة الوحدة: "+it.Title, desc, to, approverID, due)
		if err != nil {
			continue
		}
		errlog.Exec(s.db.Exec(`UPDATE "OrgPlanItem" SET "taskId" = $2, "ownerId" = COALESCE("ownerId", $3) WHERE id = $1`, it.ID, taskID, to))
		n++
	}
	return n
}
