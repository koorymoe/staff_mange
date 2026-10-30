package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// ═══ 📣 التوجيهات كسياسة عمل (دازاي 10-30 — النقطة ٧) ═══
// «تؤخذ كل توجيه ومخرج كسياسة عمل». الإصدار حسب المقام:
//   - الشركة كلها: المالك ومدير النظام.
//   - إدارة: مديرها/نائبها (أو الأعلى).
//   - وحدة: رئيسها/نائبها (أو الأعلى).
// كل من بنطاقه يقر «اطّلعت وألتزم» وينحفظ وقته. التوجيهات الفعّالة تدخل موجز ماتركس
// وكمال (OrgBrief) حتى يلتزمون بيها. ما ينحذف — يتوقف.

type OrgDirective struct {
	ID         string     `db:"id" json:"id"`
	Scope      string     `db:"scope" json:"scope"`
	DivisionID *string    `db:"divisionId" json:"divisionId"`
	UnitID     *string    `db:"unitId" json:"unitId"`
	Place      string     `db:"place" json:"place"`
	Title      string     `db:"title" json:"title"`
	Body       string     `db:"body" json:"body"`
	Clause     *int       `db:"clause" json:"clause"`
	ClauseCode *string    `db:"clauseCode" json:"clauseCode"`
	IssuedBy   *string    `db:"issuedBy" json:"issuedBy"`
	IssuedAt   time.Time  `db:"issuedAt" json:"issuedAt"`
	ArchivedAt *time.Time `db:"archivedAt" json:"archivedAt"`
	AckAt      *time.Time `db:"ackAt" json:"ackAt"`
	Acks       int        `db:"acks" json:"acks"`
	Audience   int        `db:"audience" json:"audience"`
}

type DirectiveInput struct {
	Scope      string `json:"scope"`
	DivisionID string `json:"divisionId"`
	UnitID     string `json:"unitId"`
	Title      string `json:"title"`
	Body       string `json:"body"`
	Clause     *int   `json:"clause"`
}

// directiveAudienceSQL الي بنطاق التوجيه: الشركة = كل الفعّالين، الإدارة/الوحدة = كادرها بالهيكل.
const directiveAudienceSQL = `SELECT e.id FROM "Employee" e WHERE e.status = 'ACTIVE' AND (
	d.scope = 'COMPANY'
	OR (d.scope = 'DIVISION' AND EXISTS (SELECT 1 FROM "OrgAssignment" a WHERE a."employeeId" = e.id AND a."endedAt" IS NULL AND a."divisionId" = d."divisionId"))
	OR (d.scope = 'UNIT' AND EXISTS (SELECT 1 FROM "OrgAssignment" a WHERE a."employeeId" = e.id AND a."endedAt" IS NULL
		AND (a."unitId" = d."unitId" OR (a."unitId" IS NULL AND a."divisionId" = d."divisionId" AND a.position IN ('DIRECTOR', 'DEPUTY'))))))`

// IssueDirective يصدر توجيه ويرجّع الي لازم يوصلهم إشعار.
func (s *OrgService) IssueDirective(empID, role string, in DirectiveInput) (string, []string, error) {
	in.Scope = strings.ToUpper(strings.TrimSpace(in.Scope))
	in.Title, in.Body = strings.TrimSpace(in.Title), strings.TrimSpace(in.Body)
	if len([]rune(in.Title)) < 3 || len([]rune(in.Body)) < 10 {
		return "", nil, errors.New("اكتب عنوان التوجيه ونصه (١٠ حروف على الأقل)")
	}
	if in.Clause != nil && (*in.Clause < 1 || *in.Clause > 40) {
		return "", nil, errors.New("رقم الفقرة غير صالح")
	}
	var div, unit any
	switch in.Scope {
	case "COMPANY":
		if role != "OWNER" && role != "ADMIN" {
			return "", nil, ErrOrgHidden
		}
	case "DIVISION":
		if _, edit := s.Access(empID, role, in.DivisionID); !edit {
			return "", nil, ErrOrgHidden
		}
		div = in.DivisionID
	case "UNIT":
		var d string
		if err := s.db.Get(&d, `SELECT "divisionId" FROM "OrgUnit" WHERE id = $1`, in.UnitID); err != nil {
			return "", nil, errors.New("الوحدة مو موجودة")
		}
		if !s.UnitAccess(empID, role, d, in.UnitID) {
			return "", nil, ErrOrgHidden
		}
		div, unit = d, in.UnitID
	default:
		return "", nil, errors.New("النطاق: الشركة أو إدارة أو وحدة")
	}
	var id string
	if err := s.db.Get(&id, `INSERT INTO "OrgDirective" (scope, "divisionId", "unitId", title, body, clause, "issuedById")
		VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, '')) RETURNING id`, in.Scope, div, unit, in.Title, in.Body, in.Clause, empID); err != nil {
		return "", nil, err
	}
	// المُصدِر يُحسب مقر
	if empID != "" {
		_, _ = s.db.Exec(`INSERT INTO "OrgDirectiveAck" ("directiveId", "employeeId") VALUES ($1, $2) ON CONFLICT DO NOTHING`, id, empID)
	}
	to := []string{}
	_ = s.db.Select(&to, `SELECT x.id FROM "OrgDirective" d, LATERAL (`+directiveAudienceSQL+`) x WHERE d.id = $1 AND x.id <> $2`, id, empID)
	return id, to, nil
}

const directiveSelect = `SELECT d.id, d.scope, d."divisionId", d."unitId",
	CASE d.scope WHEN 'COMPANY' THEN 'الشركة كلها' WHEN 'DIVISION' THEN dv.name ELSE dv.name || ' / ' || u.name END AS place,
	d.title, d.body, d.clause, split_part(pp.content, ' — ', 1) AS "clauseCode", e.name AS "issuedBy", d."issuedAt", d."archivedAt"`

const directiveJoins = ` FROM "OrgDirective" d LEFT JOIN "OrgDivision" dv ON dv.id = d."divisionId" LEFT JOIN "OrgUnit" u ON u.id = d."unitId"
	LEFT JOIN "PrivacyPolicyPoint" pp ON pp."principleNo" = d.clause LEFT JOIN "Employee" e ON e.id = d."issuedById"`

// MyDirectives التوجيهات الفعّالة الي بنطاق الموظف، ويا وقت إقراره (فاضي = ما أقر).
func (s *OrgService) MyDirectives(empID string) ([]OrgDirective, error) {
	out := []OrgDirective{}
	err := s.db.Select(&out, directiveSelect+`, k."ackAt", 0 AS acks, 0 AS audience`+directiveJoins+`
		LEFT JOIN "OrgDirectiveAck" k ON k."directiveId" = d.id AND k."employeeId" = $1
		WHERE d."archivedAt" IS NULL AND $1 IN (`+directiveAudienceSQL+`)
		ORDER BY (k."ackAt" IS NOT NULL), d."issuedAt" DESC LIMIT 100`, empID)
	return out, err
}

// AckDirective «اطّلعت وألتزم» — بس إذا التوجيه بنطاقه.
func (s *OrgService) AckDirective(empID, id string) error {
	var in bool
	_ = s.db.Get(&in, `SELECT EXISTS (SELECT 1 FROM "OrgDirective" d WHERE d.id = $1 AND d."archivedAt" IS NULL AND $2 IN (`+directiveAudienceSQL+`))`, id, empID)
	if !in {
		return ErrOrgHidden
	}
	_, err := s.db.Exec(`INSERT INTO "OrgDirectiveAck" ("directiveId", "employeeId") VALUES ($1, $2) ON CONFLICT DO NOTHING`, id, empID)
	return err
}

// ManagedDirectives التوجيهات الي المستخدم يگدر يتابعها (الي أصدرها، ونطاقه)، ويا عدد المقرين من الجمهور.
func (s *OrgService) ManagedDirectives(empID, role string) ([]OrgDirective, error) {
	all := []OrgDirective{}
	if err := s.db.Select(&all, directiveSelect+`, NULL::timestamptz AS "ackAt",
		(SELECT count(*) FROM "OrgDirectiveAck" k WHERE k."directiveId" = d.id AND k."employeeId" IN (`+directiveAudienceSQL+`)) AS acks,
		(SELECT count(*) FROM (`+directiveAudienceSQL+`) x) AS audience`+directiveJoins+`
		ORDER BY d."archivedAt" IS NOT NULL, d."issuedAt" DESC LIMIT 200`); err != nil {
		return nil, err
	}
	out := []OrgDirective{}
	for _, d := range all {
		ok := role == "OWNER" || role == "ADMIN"
		if !ok && d.DivisionID != nil {
			if d.UnitID != nil {
				ok = s.UnitAccess(empID, role, *d.DivisionID, *d.UnitID)
			} else {
				_, ok = s.Access(empID, role, *d.DivisionID)
			}
		}
		if ok {
			out = append(out, d)
		}
	}
	return out, nil
}

// ArchiveDirective يوقف توجيه (ما ينحذف) — المُصدِر أو صاحب المقام.
func (s *OrgService) ArchiveDirective(empID, role, id string) error {
	mine, err := s.ManagedDirectives(empID, role)
	if err != nil {
		return err
	}
	for _, d := range mine {
		if d.ID == id {
			_, err := s.db.Exec(`UPDATE "OrgDirective" SET "archivedAt" = now() WHERE id = $1 AND "archivedAt" IS NULL`, id)
			return err
		}
	}
	return ErrOrgHidden
}

// directivesBrief التوجيهات الفعّالة لموجز الذكاء (آخر ١٢).
func directivesBrief(db *sqlx.DB) string {
	if db == nil {
		return ""
	}
	rows := []struct {
		Place string  `db:"place"`
		Title string  `db:"title"`
		Code  *string `db:"code"`
	}{}
	_ = db.Select(&rows, `SELECT CASE d.scope WHEN 'COMPANY' THEN 'الشركة' WHEN 'DIVISION' THEN dv.name ELSE dv.name || ' / ' || u.name END AS place,
		d.title, split_part(pp.content, ' — ', 1) AS code
		FROM "OrgDirective" d LEFT JOIN "OrgDivision" dv ON dv.id = d."divisionId" LEFT JOIN "OrgUnit" u ON u.id = d."unitId"
		LEFT JOIN "PrivacyPolicyPoint" pp ON pp."principleNo" = d.clause
		WHERE d."archivedAt" IS NULL ORDER BY d."issuedAt" DESC LIMIT 12`)
	if len(rows) == 0 {
		return ""
	}
	parts := []string{}
	for _, r := range rows {
		p := r.Place + ": " + r.Title
		if r.Code != nil && *r.Code != "" {
			p += " (" + *r.Code + ")"
		}
		parts = append(parts, p)
	}
	return "\nالتوجيهات الفعّالة (سياسة عمل ملزمة — التزم بيها بجوابك): " + strings.Join(parts, "؛ ") + "."
}

var (
	directiveOrg    *OrgService
	directiveNotify func(empID, msg string)
)

// SetDirectiveDeps يربط أداة كمال للتوجيهات بالهيكل والإشعارات (من main).
func SetDirectiveDeps(org *OrgService, notify func(empID, msg string)) {
	directiveOrg, directiveNotify = org, notify
}

// registerDirectiveTool كمال يصدر توجيه بطلب المالك/المدير («عمّم على الكل…»).
func registerDirectiveTool(a *MatrixAgent, tc *SecretaryTurnCtx) {
	org, notify := directiveOrg, directiveNotify
	if tc == nil || org == nil {
		return
	}
	a.AddTool(agentTool{Name: "issue_directive",
		Desc: "أصدر توجيه رسمي (يصير سياسة عمل ملزمة) لما المالك/المدير يگول «عمّم/وجّه الكل/خلي وحدة كذا تلتزم بـ…». scope = COMPANY أو DIVISION (ويا division) أو UNIT (ويا unit)، واربطه بفقرة اللائحة (clause = رقمها ١-٢٠) إذا تنطبق. الكل بنطاقه يوصله إشعار ويقر «اطّلعت وألتزم».",
		Props: map[string]any{
			"scope":    map[string]any{"type": "string", "enum": []string{"COMPANY", "DIVISION", "UNIT"}},
			"division": map[string]any{"type": "string"},
			"unit":     map[string]any{"type": "string"},
			"title":    map[string]any{"type": "string"},
			"body":     map[string]any{"type": "string"},
			"clause":   map[string]any{"type": "integer"},
		},
		Req: []string{"scope", "title", "body"},
		Run: func(in map[string]any, _ *nameMask) (string, error) {
			var clause *int
			if v, ok := in["clause"].(float64); ok && v > 0 {
				c := int(v)
				clause = &c
			}
			var role string
			_ = a.db.Get(&role, `SELECT role::text FROM "Employee" WHERE id = $1`, tc.UserID)
			id, to, err := org.IssueDirective(tc.UserID, role, DirectiveInput{Scope: strArg(in, "scope"), DivisionID: strArg(in, "division"),
				UnitID: strArg(in, "unit"), Title: strArg(in, "title"), Body: strArg(in, "body"), Clause: clause})
			if err != nil {
				if errors.Is(err, ErrOrgHidden) {
					return "", errors.New("مو ضمن صلاحيته يصدر توجيه بهالنطاق")
				}
				return "", err
			}
			if notify != nil {
				for _, e := range to {
					notify(e, "📣 توجيه جديد: «"+strArg(in, "title")+"» — اقراه وأقر «اطّلعت وألتزم».")
				}
			}
			return fmt.Sprintf("انصدر التوجيه (%s) ووصل لـ%d شخص. گوله بجملة.", id[:8], len(to)), nil
		}})
}
