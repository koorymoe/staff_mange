package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// ═══ 🗂️ رأي كمال التنفيذي بكل شخص (دازاي 10-30 — النقطة ٦) ═══
// «رأي كمال بكل شخص من الناحية التنفيذية». كمال (السكرتير، الأسماء مسموحة إله) ياخذ
// حقائق الموظف من النظام بس — مكانه، تقييمه، التزامه، شكاويه، بنود خططه، قرارات
// ماتركس عليه، وموافقته على اللائحة — ويكتب رأيه بتقدير ١-٥ ونقاط قوة وضعف وتوصية،
// وينحفظ (رأي واحد بالشهر يتحدّث) حتى يرجعله المالك والمدير وماتركس.

type ExecFacts struct {
	EmployeeID   string   `json:"employeeId"`
	Name         string   `json:"name"`
	Positions    []string `json:"positions"`
	ScoreNow     *float64 `json:"scoreNow"`
	ScorePrev    *float64 `json:"scorePrev"`
	Stations     int      `json:"stations"`
	Attendance   int      `json:"attendance14"`
	LateDays     int      `json:"lateDays14"`
	DoneBookings int      `json:"doneBookings14"`
	Complaints   int      `json:"complaints30"`
	ItemsTotal   int      `json:"planItems"`
	ItemsDone    int      `json:"planItemsDone"`
	ItemsLate    int      `json:"planItemsOverdue"`
	Fines        int      `json:"fines90"`
	Warnings     int      `json:"warnings90"`
	Revokes      int      `json:"revokes90"`
	Accepted     bool     `json:"regulationsAccepted"`
	OpenIssues   int      `json:"openIssuesInHisUnits"`
}

func monthPct(db *sqlx.DB, empID string, from, to time.Time) (*float64, int) {
	var r struct {
		P float64 `db:"p"`
		M float64 `db:"m"`
		N int     `db:"n"`
	}
	_ = db.Get(&r, `SELECT COALESCE(sum(points), 0) AS p, COALESCE(sum("maxPoints"), 0) AS m, count(*) AS n FROM "MatrixScore"
		WHERE "employeeId" = $1 AND "cancelledAt" IS NULL AND "at" >= $2 AND "at" < $3`, empID, from, to)
	if r.M == 0 {
		return nil, 0
	}
	v := math.Round(r.P/r.M*1000) / 10
	return &v, r.N
}

// ExecutiveFacts حقائق موظف من النظام (بلا رأي).
func ExecutiveFacts(db *sqlx.DB, empID string) (*ExecFacts, error) {
	f := &ExecFacts{EmployeeID: empID, Positions: []string{}}
	if err := db.Get(&f.Name, `SELECT name FROM "Employee" WHERE id = $1`, empID); err != nil {
		return nil, errors.New("الموظف مو موجود")
	}
	pos := []struct {
		Pos  string  `db:"position"`
		Div  string  `db:"div"`
		Unit *string `db:"unit"`
		Act  bool    `db:"acting"`
	}{}
	_ = db.Select(&pos, `SELECT a.position, d.name AS div, u.name AS unit, a.acting FROM "OrgAssignment" a
		JOIN "OrgDivision" d ON d.id = a."divisionId" LEFT JOIN "OrgUnit" u ON u.id = a."unitId" WHERE a."employeeId" = $1 AND a."endedAt" IS NULL`, empID)
	for _, p := range pos {
		s := orgPosAr[p.Pos]
		if p.Act {
			s += " بالوكالة"
		}
		s += " — " + p.Div
		if p.Unit != nil {
			s += " / " + *p.Unit
		}
		f.Positions = append(f.Positions, s)
	}
	now := time.Now().In(baghdad)
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, baghdad)
	f.ScoreNow, f.Stations = monthPct(db, empID, start, start.AddDate(0, 1, 0))
	f.ScorePrev, _ = monthPct(db, empID, start.AddDate(0, -1, 0), start)
	if m, err := EmployeeMetrics(db, empID); err == nil && len(m) == 1 {
		f.Attendance, f.LateDays, f.DoneBookings = m[0].Att, m[0].Late, m[0].Done
	}
	_ = db.Get(&f.Complaints, `SELECT count(*) FROM "Complaint" c WHERE `+orgComplaintPerson+` = $1 AND c."createdAt" > now() - interval '30 days'`, empID)
	_ = db.QueryRow(`SELECT count(*), count("doneAt"), count(*) FILTER (WHERE "doneAt" IS NULL AND "dueDate" < (now() AT TIME ZONE 'Asia/Baghdad')::date)
		FROM "OrgPlanItem" WHERE "ownerId" = $1`, empID).Scan(&f.ItemsTotal, &f.ItemsDone, &f.ItemsLate)
	_ = db.QueryRow(`SELECT count(*) FILTER (WHERE kind IN ('ORG_FINE_CUSTOMER', 'ORG_FINE_OPS')), count(*) FILTER (WHERE kind = 'ORG_WARN'),
		count(*) FILTER (WHERE kind = 'ORG_REVOKE') FROM "AiAction" WHERE "targetEmployeeId" = $1 AND status = 'DONE' AND "createdAt" > now() - interval '90 days'`,
		empID).Scan(&f.Fines, &f.Warnings, &f.Revokes)
	_ = db.Get(&f.Accepted, `SELECT EXISTS (SELECT 1 FROM "PrivacyPolicyAcceptance" WHERE "employeeId" = $1 AND "pointsHash" IS NOT NULL)`, empID)
	_ = db.Get(&f.OpenIssues, `SELECT count(*) FROM "OrgIssueReport" i WHERE i.status <> 'RESOLVED' AND EXISTS (
		SELECT 1 FROM "OrgAssignment" a WHERE a."employeeId" = $1 AND a."endedAt" IS NULL AND a.position IN ('DIRECTOR', 'DEPUTY')
		AND (a."unitId" = i."unitId" OR (a."unitId" IS NULL AND a."divisionId" = i."divisionId")))`, empID)
	return f, nil
}

func (f *ExecFacts) Text() string {
	pct := func(v *float64) string {
		if v == nil {
			return "ماكو بيانات"
		}
		return fmt.Sprintf("%.1f٪", *v)
	}
	pos := "بلا مكان بالهيكل"
	if len(f.Positions) > 0 {
		pos = strings.Join(f.Positions, "؛ ")
	}
	acc := "ما وافق"
	if f.Accepted {
		acc = "وافق"
	}
	return fmt.Sprintf("%s — %s\nتقييم ماتركس: هالشهر %s (%d محطة)، الشهر الي فات %s\nآخر ١٤ يوم: حضور %d يوم، تأخير %d يوم، حجوزات خلصها %d\nشكاوي عليه آخر ٣٠ يوم: %d\nبنود خطط عليه: %d (خلص %d، متأخر %d)\nقرارات ماتركس آخر ٩٠ يوم: %d غرامة، %d تنبيه، %d سحب منصب\nلائحة العمل: %s\nمشاكل مفتوحة بوحداته (إذا رئيس): %d",
		f.Name, pos, pct(f.ScoreNow), f.Stations, pct(f.ScorePrev), f.Attendance, f.LateDays, f.DoneBookings,
		f.Complaints, f.ItemsTotal, f.ItemsDone, f.ItemsLate, f.Fines, f.Warnings, f.Revokes, acc, f.OpenIssues)
}

type ExecOpinion struct {
	Period         string    `db:"period" json:"period"`
	Rating         int       `db:"rating" json:"rating"`
	Strengths      string    `db:"strengths" json:"strengths"`
	Weaknesses     string    `db:"weaknesses" json:"weaknesses"`
	Recommendation string    `db:"recommendation" json:"recommendation"`
	UpdatedAt      time.Time `db:"updatedAt" json:"updatedAt"`
}

// SaveExecutiveOpinion يحفظ رأي كمال للشهر الحالي (يتحدّث إذا انكتب مرة ثانية).
func SaveExecutiveOpinion(db *sqlx.DB, empID, by string, rating int, strengths, weaknesses, rec string) error {
	if rating < 1 || rating > 5 {
		return errors.New("التقدير من ١ لـ٥")
	}
	rec = strings.TrimSpace(rec)
	if len([]rune(rec)) < 10 {
		return errors.New("اكتب توصية واضحة (١٠ حروف على الأقل)")
	}
	f, err := ExecutiveFacts(db, empID)
	if err != nil {
		return err
	}
	facts, _ := json.Marshal(f)
	_, err = db.Exec(`INSERT INTO "OrgExecutiveOpinion" ("employeeId", period, rating, strengths, weaknesses, recommendation, facts, "requestedById")
		VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''))
		ON CONFLICT ("employeeId", period) DO UPDATE SET rating = EXCLUDED.rating, strengths = EXCLUDED.strengths, weaknesses = EXCLUDED.weaknesses,
		recommendation = EXCLUDED.recommendation, facts = EXCLUDED.facts, "requestedById" = EXCLUDED."requestedById", "updatedAt" = now()`,
		empID, OrgPeriod(time.Now()), rating, strings.TrimSpace(strengths), strings.TrimSpace(weaknesses), rec, facts, by)
	return err
}

func ExecutiveOpinions(db *sqlx.DB, empID string) ([]ExecOpinion, error) {
	out := []ExecOpinion{}
	err := db.Select(&out, `SELECT period, rating, strengths, weaknesses, recommendation, "updatedAt" FROM "OrgExecutiveOpinion"
		WHERE "employeeId" = $1 ORDER BY period DESC LIMIT 12`, empID)
	return out, err
}

// registerExecutiveTools أدوات كمال (السكرتير بس — المالك والمدير، والأسماء مسموحة).
func registerExecutiveTools(a *MatrixAgent, tc *SecretaryTurnCtx) {
	if tc == nil {
		return
	}
	a.AddTool(agentTool{Name: "executive_review",
		Desc: "حقائق تنفيذية عن موظف (أو كل كادر وحدة) من النظام: مكانه، تقييم ماتركس، الحضور والتأخير، الشكاوي، بنود خططه، قرارات ماتركس عليه، وموافقته على اللائحة. استعملها قبل ما تعطي رأيك التنفيذي بأي شخص — رأيك يبنى على هالأرقام بس. who = اسم موظف، أو unit = رمز وحدة (مثل ops.fleet) لكل كادرها.",
		Props: map[string]any{
			"who":  map[string]any{"type": "string"},
			"unit": map[string]any{"type": "string"},
		},
		Run: func(in map[string]any, mask *nameMask) (string, error) {
			ids := []string{}
			if u := strings.TrimSpace(strArg(in, "unit")); u != "" {
				_ = a.db.Select(&ids, `SELECT DISTINCT "employeeId" FROM "OrgAssignment" WHERE "unitId" = $1 AND "endedAt" IS NULL`, u)
				if len(ids) == 0 {
					return "ماكو أحد معيّن بهالوحدة.", nil
				}
			} else {
				id, _ := mask.Resolve(strArg(in, "who"))
				if id == "" {
					return "", errors.New("ما لگيت الموظف — اكتب اسمه مثل ما بالنظام")
				}
				ids = append(ids, id)
			}
			var b strings.Builder
			for i, id := range ids {
				if i >= 15 {
					fmt.Fprintf(&b, "\n… و%d غيرهم (اطلب واحد واحد)", len(ids)-15)
					break
				}
				f, err := ExecutiveFacts(a.db, id)
				if err != nil {
					continue
				}
				b.WriteString(f.Text() + "\n\n")
			}
			b.WriteString("اكتب رأيك التنفيذي: تقدير ١-٥، نقاط قوة، نقاط ضعف، وتوصية عملية (مثل: يستحق منصب أعلى / يحتاج متابعة بفقرة كذا من اللائحة / يتدرّب على كذا). اربط الضعف بفقرة اللائحة. وإذا المالك طلب يحفظه استعمل save_executive_opinion.")
			return b.String(), nil
		}})
	a.AddTool(agentTool{Name: "save_executive_opinion",
		Desc: "احفظ رأيك التنفيذي بموظف (للشهر الحالي، يتحدّث إذا انكتب مرة ثانية) — بعد executive_review وبطلب المالك أو المدير.",
		Props: map[string]any{
			"who":            map[string]any{"type": "string"},
			"rating":         map[string]any{"type": "integer", "description": "١-٥"},
			"strengths":      map[string]any{"type": "string"},
			"weaknesses":     map[string]any{"type": "string"},
			"recommendation": map[string]any{"type": "string"},
		},
		Req: []string{"who", "rating", "recommendation"},
		Run: func(in map[string]any, mask *nameMask) (string, error) {
			id, name := mask.Resolve(strArg(in, "who"))
			if id == "" {
				return "", errors.New("ما لگيت الموظف")
			}
			rating := 0
			if v, ok := in["rating"].(float64); ok {
				rating = int(v)
			}
			if err := SaveExecutiveOpinion(a.db, id, tc.UserID, rating, strArg(in, "strengths"), strArg(in, "weaknesses"), strArg(in, "recommendation")); err != nil {
				return "", err
			}
			return "انحفظ رأيك بـ" + name + " لهالشهر. گوله بجملة.", nil
		}})
}
