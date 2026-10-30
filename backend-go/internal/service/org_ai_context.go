package service

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

// ═══ 🧠 الذكاء يفهم النظام الجديد (دازاي 10-30 — النقطة ٢) ═══
// «فهم كل محاور ومخرجات النظام من قبل الذكاء». ماتركس وكمال ياخذون موجز ثابت بالبرومبت
// (الهرم، لائحة العمل، مكان المستخدم، قواعد الحكم، دورة تطوير الوحدة) — بلا أسماء —
// وثلاث أدوات للتفاصيل: org_structure (الأسماء تمر من القناع)، regulations، unit_status.

// OrgBrief موجز النظام الجديد للبرومبت (بلا أسماء). empID فاضي = بلا «مكانك».
func OrgBrief(db *sqlx.DB, empID string) string {
	if db == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString("\n\n【الهيكل الجديد ولائحة العمل】\n")
	b.WriteString("الهرم: مدير الشركة (المالك) ومدير النظام فوك الكل ← مدير «الإدارة» يراقب كل الإدارات ← باقي الإدارات. كل إدارة ووحدة إلها مدير ونائب وأعضاء.\n")
	divs := []struct {
		ID      string `db:"id"`
		Name    string `db:"name"`
		Summary string `db:"summary"`
		Units   string `db:"units"`
	}{}
	_ = db.Select(&divs, `SELECT d.id, d.name, d.summary, COALESCE(string_agg(u.name, '، ' ORDER BY u.sort), '') AS units
		FROM "OrgDivision" d LEFT JOIN "OrgUnit" u ON u."divisionId" = d.id GROUP BY d.id, d.name, d.summary, d.sort ORDER BY d.sort`)
	for _, d := range divs {
		fmt.Fprintf(&b, "- %s (%s): %s\n", d.Name, d.ID, d.Units)
	}
	clauses := []struct {
		Code  string `db:"code"`
		Title string `db:"title"`
	}{}
	_ = db.Select(&clauses, `SELECT split_part(content, ' — ', 1) AS code, "clauseTitle" AS title FROM "PrivacyPolicyPoint"
		WHERE "isActive" AND "principleNo" IS NOT NULL AND "clauseTitle" IS NOT NULL ORDER BY "principleNo"`)
	if len(clauses) > 0 {
		parts := make([]string, 0, len(clauses))
		for _, c := range clauses {
			parts = append(parts, c.Code+" "+c.Title)
		}
		b.WriteString("لائحة العمل (كل موظف وافق عليها): " + strings.Join(parts, "؛ ") + ".\n")
	}
	b.WriteString("الحكم: ماتركس يقيّم أصحاب المناصب شهرياً (تحت ٥٠٪ = تنبيه رسمي، شهرين متتاليين = سحب المنصب ومخصصه)، ويغرّم على شكوى الزبون (٣٠–١٠٠ ألف بنقاط الانضباط) وعلى مدير العمليات إذا شكوى بقت ٤٨ ساعة بلا حل. بالإجازة النائب يمسك المدير بالوكالة. كل قرار ينرجع.\n")
	b.WriteString("تطوير الوحدة: مدير الوحدة يشخّص مشاكلها ويكتب خطة حل، ويسوّي خطة تطوير ببنود بمواعيد ومسؤولين، ويرفعها لمدير إدارته الي يعتمد أو يرجّع بملاحظة.\n")
	if empID != "" {
		mine := []struct {
			Pos  string  `db:"position"`
			Div  string  `db:"div"`
			Unit *string `db:"unit"`
			Act  bool    `db:"acting"`
		}{}
		_ = db.Select(&mine, `SELECT a.position, d.name AS div, u.name AS unit, a.acting FROM "OrgAssignment" a
			JOIN "OrgDivision" d ON d.id = a."divisionId" LEFT JOIN "OrgUnit" u ON u.id = a."unitId"
			WHERE a."employeeId" = $1 AND a."endedAt" IS NULL`, empID)
		if len(mine) == 0 {
			b.WriteString("مكان المستخدم بالهيكل: ما انوضع بعد.\n")
		}
		for _, m := range mine {
			place := m.Div
			if m.Unit != nil {
				place += " / " + *m.Unit
			}
			act := ""
			if m.Act {
				act = " بالوكالة"
			}
			fmt.Fprintf(&b, "مكان المستخدم: %s%s — %s.\n", orgPosAr[m.Pos], act, place)
		}
	}
	b.WriteString(directivesBrief(db) + "\n")
	b.WriteString("اربط أي مشكلة أو ملاحظة بفقرة من اللائحة (برمزها مثل م٢/ف١). للأسماء والأرقام والتفاصيل استعمل org_structure وregulations وunit_status — لا تخمّن.")
	return b.String()
}

var orgPosAr = map[string]string{"DIRECTOR": "مدير", "DEPUTY": "نائب", "MEMBER": "عضو"}

func (a *MatrixAgent) registerOrgTools() {
	a.AddTool(agentTool{Name: "org_structure",
		Desc: "الهيكل التنظيمي الحالي: كل إدارة ووحدة، منو المدير والنائب (ومنو بالوكالة)، الكراسي الفاضية، وعدد الأعضاء. استعملها لأي «منو مدير…/منو بوحدة…/وين مكان فلان».",
		Run: func(_ map[string]any, mask *nameMask) (string, error) {
			rows := []struct {
				Div    string  `db:"div"`
				Unit   *string `db:"unit"`
				Pos    string  `db:"position"`
				Name   string  `db:"name"`
				Acting bool    `db:"acting"`
			}{}
			if err := a.db.Select(&rows, `SELECT d.name AS div, u.name AS unit, a.position, e.name, a.acting
				FROM "OrgAssignment" a JOIN "Employee" e ON e.id = a."employeeId" JOIN "OrgDivision" d ON d.id = a."divisionId"
				LEFT JOIN "OrgUnit" u ON u.id = a."unitId" WHERE a."endedAt" IS NULL
				ORDER BY d.sort, u.sort NULLS FIRST, CASE a.position WHEN 'DIRECTOR' THEN 0 WHEN 'DEPUTY' THEN 1 ELSE 2 END, e.name`); err != nil {
				return "", err
			}
			if len(rows) == 0 {
				return "ماكو أحد معيّن بالهيكل بعد — المالك يعيّن من «🏛️ الهيكل التنظيمي».", nil
			}
			var b strings.Builder
			for _, r := range rows {
				place := r.Div
				if r.Unit != nil {
					place += " / " + *r.Unit
				}
				act := ""
				if r.Acting {
					act = " (بالوكالة)"
				}
				fmt.Fprintf(&b, "- %s: %s%s — %s\n", place, orgPosAr[r.Pos], act, r.Name)
			}
			vac := []string{}
			_ = a.db.Select(&vac, `SELECT x.place || ' — بلا ' || CASE x.pos WHEN 'DIRECTOR' THEN 'مدير' ELSE 'نائب' END FROM (
				SELECT d.name AS place, d.id AS div, NULL::text AS unit, d.sort AS s1, -1 AS s2 FROM "OrgDivision" d
				UNION ALL SELECT d.name || ' / ' || u.name, d.id, u.id, d.sort, u.sort FROM "OrgUnit" u JOIN "OrgDivision" d ON d.id = u."divisionId") p
				CROSS JOIN (VALUES ('DIRECTOR'), ('DEPUTY')) AS v(pos)
				CROSS JOIN LATERAL (SELECT p.place, v.pos) x
				WHERE NOT EXISTS (SELECT 1 FROM "OrgAssignment" a WHERE a."divisionId" = p.div AND a."unitId" IS NOT DISTINCT FROM p.unit
					AND a.position = v.pos AND a."endedAt" IS NULL)
				ORDER BY p.s1, p.s2, v.pos`)
			if len(vac) > 0 {
				b.WriteString("الكراسي الفاضية: " + strings.Join(vac, "؛ "))
			}
			return mask.Hide(b.String()), nil
		}})

	a.AddTool(agentTool{Name: "regulations",
		Desc:  "نص لائحة العمل الرسمية (المواد والفقرات) — للاستشهاد بفقرة أو شرحها. article اختياري (١-٥).",
		Props: map[string]any{"article": map[string]any{"type": "integer", "description": "رقم المادة (اختياري)"}},
		Run: func(in map[string]any, _ *nameMask) (string, error) {
			art := 0
			if v, ok := in["article"].(float64); ok {
				art = int(v)
			}
			rows := []string{}
			if err := a.db.Select(&rows, `SELECT p.content FROM "PrivacyPolicyPoint" p
				WHERE p."isActive" AND p."principleNo" IS NOT NULL AND ($1 = 0 OR p."articleNo" = $1) ORDER BY p."principleNo"`, art); err != nil {
				return "", err
			}
			if len(rows) == 0 {
				return "ماكو فقرات فعّالة بهالمادة.", nil
			}
			return strings.Join(rows, "\n"), nil
		}})

	a.AddTool(agentTool{Name: "unit_status",
		Desc:  "حال إدارة (أو كل الإدارات): المشاكل المفتوحة بخطورتها وفقرة اللائحة، خطط الوحدات وحالتها ونسبتها، البنود المتأخرة ومسؤوليها، وقرارات ماتركس الحاكم الأخيرة. division = رمز الإدارة (admin, ops, bizdev, eng, quality, supply, contracts, it) أو فاضي للكل.",
		Props: map[string]any{"division": map[string]any{"type": "string"}},
		Run: func(in map[string]any, mask *nameMask) (string, error) {
			div := strings.TrimSpace(strArg(in, "division"))
			var b strings.Builder
			issues := []struct {
				Div, Title, Sev, Status string
				Unit, Code, Sol         *string
			}{}
			rows, err := a.db.Queryx(`SELECT d.name, i.title, i.severity, i.status, u.name, split_part(pp.content, ' — ', 1), i."solutionPlan"
				FROM "OrgIssueReport" i JOIN "OrgDivision" d ON d.id = i."divisionId" LEFT JOIN "OrgUnit" u ON u.id = i."unitId"
				LEFT JOIN "PrivacyPolicyPoint" pp ON pp."principleNo" = i.principle
				WHERE i.status <> 'RESOLVED' AND ($1 = '' OR i."divisionId" = $1)
				ORDER BY d.sort, CASE i.severity WHEN 'HIGH' THEN 0 WHEN 'MEDIUM' THEN 1 ELSE 2 END LIMIT 40`, div)
			if err != nil {
				return "", err
			}
			for rows.Next() {
				var x struct {
					Div, Title, Sev, Status string
					Unit, Code, Sol         *string
				}
				if err := rows.Scan(&x.Div, &x.Title, &x.Sev, &x.Status, &x.Unit, &x.Code, &x.Sol); err == nil {
					issues = append(issues, x)
				}
			}
			rows.Close()
			sevAr := map[string]string{"HIGH": "عالية", "MEDIUM": "متوسطة", "LOW": "قليلة"}
			b.WriteString(fmt.Sprintf("المشاكل المفتوحة (%d):\n", len(issues)))
			for _, x := range issues {
				place := x.Div
				if x.Unit != nil {
					place += " / " + *x.Unit
				}
				code, sol := "", " — بلا خطة حل"
				if x.Code != nil && *x.Code != "" {
					code = " [" + *x.Code + "]"
				}
				if x.Sol != nil && *x.Sol != "" {
					sol = " — بيها خطة حل"
				}
				fmt.Fprintf(&b, "- %s: %s (%s)%s%s\n", place, x.Title, sevAr[x.Sev], code, sol)
			}
			plans := []struct {
				Div    string  `db:"div"`
				Unit   *string `db:"unit"`
				Period string  `db:"period"`
				Status string  `db:"status"`
				Items  int     `db:"items"`
				Done   int     `db:"done"`
				Late   int     `db:"late"`
			}{}
			_ = a.db.Select(&plans, `SELECT d.name AS div, u.name AS unit, p.period, p.status,
				count(i.id) AS items, count(i."doneAt") AS done,
				count(i.id) FILTER (WHERE i."doneAt" IS NULL AND i."dueDate" < (now() AT TIME ZONE 'Asia/Baghdad')::date) AS late
				FROM "OrgPlan" p JOIN "OrgDivision" d ON d.id = p."divisionId" LEFT JOIN "OrgUnit" u ON u.id = p."unitId"
				LEFT JOIN "OrgPlanItem" i ON i."planId" = p.id
				WHERE ($1 = '' OR p."divisionId" = $1) AND p.period >= to_char((now() AT TIME ZONE 'Asia/Baghdad') - interval '1 month', 'YYYY-MM')
				GROUP BY d.name, d.sort, u.name, u.sort, p.period, p.status ORDER BY d.sort, u.sort NULLS FIRST, p.period DESC`, div)
			stAr := map[string]string{"DRAFT": "مسودة", "SUBMITTED": "مرفوعة للاعتماد", "APPROVED": "معتمدة", "RETURNED": "مرجّعة للتعديل"}
			b.WriteString(fmt.Sprintf("الخطط (الشهر الحالي والي فات، %d):\n", len(plans)))
			for _, p := range plans {
				place := p.Div
				if p.Unit != nil {
					place += " / " + *p.Unit
				}
				fmt.Fprintf(&b, "- %s %s: %s — %d بند، خلص %d، متأخر %d\n", place, p.Period, stAr[p.Status], p.Items, p.Done, p.Late)
			}
			late := []struct {
				Title string  `db:"title"`
				Due   string  `db:"due"`
				Owner *string `db:"owner"`
				Place string  `db:"place"`
			}{}
			_ = a.db.Select(&late, `SELECT i.title, i."dueDate"::text AS due, e.name AS owner, d.name || COALESCE(' / ' || u.name, '') AS place
				FROM "OrgPlanItem" i JOIN "OrgPlan" p ON p.id = i."planId" JOIN "OrgDivision" d ON d.id = p."divisionId"
				LEFT JOIN "OrgUnit" u ON u.id = p."unitId" LEFT JOIN "Employee" e ON e.id = i."ownerId"
				WHERE i."doneAt" IS NULL AND i."dueDate" < (now() AT TIME ZONE 'Asia/Baghdad')::date AND ($1 = '' OR p."divisionId" = $1)
				ORDER BY i."dueDate" LIMIT 20`, div)
			if len(late) > 0 {
				b.WriteString("البنود المتأخرة:\n")
				for _, l := range late {
					owner := "بلا مسؤول"
					if l.Owner != nil {
						owner = *l.Owner
					}
					fmt.Fprintf(&b, "- %s: «%s» موعده %s — %s\n", l.Place, l.Title, l.Due, owner)
				}
			}
			var gov struct {
				Fines   int `db:"fines"`
				Warns   int `db:"warns"`
				Revokes int `db:"revokes"`
			}
			_ = a.db.Get(&gov, `SELECT count(*) FILTER (WHERE kind IN ('ORG_FINE_CUSTOMER', 'ORG_FINE_OPS')) AS fines,
				count(*) FILTER (WHERE kind = 'ORG_WARN') AS warns, count(*) FILTER (WHERE kind = 'ORG_REVOKE') AS revokes
				FROM "AiAction" WHERE status = 'DONE' AND "createdAt" > now() - interval '30 days'`)
			fmt.Fprintf(&b, "قرارات ماتركس الحاكم آخر ٣٠ يوم (كل الشركة): %d غرامة، %d تنبيه، %d سحب منصب.", gov.Fines, gov.Warns, gov.Revokes)
			return mask.Hide(b.String()), nil
		}})
}
