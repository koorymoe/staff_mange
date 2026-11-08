package service

import (
	"errors"
	"strings"
	"time"

	"staffmange-api/internal/errlog"
	"staffmange-api/internal/model"
)

// ═══ 📋 كل الخطط بمكان واحد (دازاي 11-08) ═══
// «المالك ومدير النظام يشوفون كل الخطط ويرفضوها ويعدلون عليها». قبل، لازم يفتح
// كل إدارة وحدة وحدة. هسه صف لكل إدارة ولكل وحدة — ويّا الي ما سوّت خطة بعد.

type OrgPlanOverviewRow struct {
	DivisionID   string     `db:"divisionId" json:"divisionId"`
	DivisionName string     `db:"divisionName" json:"divisionName"`
	UnitID       *string    `db:"unitId" json:"unitId"`
	UnitName     *string    `db:"unitName" json:"unitName"`
	PlanID       *string    `db:"planId" json:"planId"`
	Title        *string    `db:"title" json:"title"`
	Status       string     `db:"status" json:"status"` // NONE | DRAFT | SUBMITTED | APPROVED | RETURNED
	SubmittedAt  *time.Time `db:"submittedAt" json:"submittedAt"`
	Items        int        `db:"items" json:"items"`
	Done         int        `db:"done" json:"done"`
	Late         int        `db:"late" json:"late"`
	Score        *float64   `db:"score" json:"score"`
}

// PlansOverview للمالك ومدير النظام بس.
func (s *OrgService) PlansOverview(role, period string) ([]OrgPlanOverviewRow, error) {
	if !model.IsAdminRole(role) {
		return nil, ErrOrgHidden
	}
	if period == "" {
		period = OrgPeriod(time.Now())
	}
	if _, _, err := periodRange(period); err != nil {
		return nil, err
	}
	rows := []OrgPlanOverviewRow{}
	err := s.db.Select(&rows, `
		WITH scopes AS (
			SELECT d.id AS "divisionId", d.name AS "divisionName", NULL::text AS "unitId", NULL::text AS "unitName", d.sort AS ds, -1 AS us
			FROM "OrgDivision" d
			UNION ALL
			SELECT d.id, d.name, u.id, u.name, d.sort, u.sort
			FROM "OrgUnit" u JOIN "OrgDivision" d ON d.id = u."divisionId"
		)
		SELECT s."divisionId", s."divisionName", s."unitId", s."unitName",
		       p.id AS "planId", p.title, COALESCE(p.status, 'NONE') AS status, p."submittedAt",
		       (SELECT count(*) FROM "OrgPlanItem" i WHERE i."planId" = p.id)::int AS items,
		       (SELECT count(*) FROM "OrgPlanItem" i WHERE i."planId" = p.id AND i."doneAt" IS NOT NULL)::int AS done,
		       (SELECT count(*) FROM "OrgPlanItem" i WHERE i."planId" = p.id AND i."doneAt" IS NULL
		          AND i."dueDate" IS NOT NULL AND i."dueDate" < baghdad_today())::int AS late,
		       (SELECT round(avg(`+planItemScoreSQL+`)::numeric, 1)::float8 FROM "OrgPlanItem" i WHERE i."planId" = p.id) AS score
		FROM scopes s
		LEFT JOIN "OrgPlan" p ON p."divisionId" = s."divisionId" AND p."unitId" IS NOT DISTINCT FROM s."unitId" AND p.period = $1
		ORDER BY s.ds, s."divisionId", s.us`, period)
	return rows, err
}

// ReopenPlan (دازاي 11-08) المالك ومدير النظام يرجّعون خطة **حتى بعد الاعتماد** للتعديل،
// ويّا ملاحظة. المهام الي انولدت تبقى (شغل صار فعلاً)؛ والبنود الجديدة تصير مهام
// من تنعتمد مرة ثانية. يرجّع صاحب الخطة حتى ينبّه.
func (s *OrgService) ReopenPlan(empID, role, planID, note string) (string, error) {
	if !model.IsAdminRole(role) {
		return "", ErrOrgHidden
	}
	sc := s.planScope(planID)
	if sc.Div == "" {
		return "", ErrOrgHidden
	}
	if sc.Status != "APPROVED" && sc.Status != "SUBMITTED" {
		return "", errors.New("الخطة بعدها بالتعديل أصلاً")
	}
	note = strings.TrimSpace(note)
	if len([]rune(note)) < 5 {
		return "", errors.New("اكتب شنو لازم يتعدل حتى ترجّعها")
	}
	var author string
	errlog.Err(s.db.Get(&author, `SELECT COALESCE("createdById", '') FROM "OrgPlan" WHERE id = $1`, planID))
	_, err := s.db.Exec(`UPDATE "OrgPlan" SET status = 'RETURNED', "reviewedById" = $2, "reviewedAt" = now(), "reviewNote" = $3 WHERE id = $1`,
		planID, empID, note)
	return author, err
}
