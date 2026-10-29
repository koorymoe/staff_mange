package service

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ═══ 🏛️ ماتركس الحاكم (دازاي 10-29) ═══
// «ماتركس صاحب القرار — يقرر وينفّذ ويرفع تقرير للمدير». القواعد قرّرها المالك:
//   - سحب المنصب ويا مخصصه: شهرين متتاليين تحت ٥٠٪ (أول شهر = تنبيه رسمي).
//   - غرامة «مشكلة مباشرة ويا الزبون»: كل شكوى على حجز الموظف، ٢٥–١٠٠ ألف يقدّرها ماتركس.
//   - غرامة مدير العمليات: شكوى بقت ٤٨ ساعة بلا حل.
//   - الغرامة بنظام الانضباط (النقطة = ١٠ آلاف) — فالمبالغ تصير ٣٠ لـ١٠٠ ألف.
// الضمانات: كل قرار AiAction بالسبب والأرقام، من بيانات النظام بس (بلا ذكاء خارجي ولا
// أسماء تطلع)، وينرجع من صفحة القرارات (التراجع يرجّع النقاط أو المنصب). مفتاح طوارئ
// org_governor. والمالك ومدير النظام ما ينحكمون.
// ⚠️ غرامات الانضباط التلقائية القديمة (DisciplineAutoPenaltyPaused) تبقى موقوفة —
// هذا مسار منفصل بقرار المالك الجديد.

const (
	orgRevokeBelow     = 50.0
	orgFineDeadline    = 48 * time.Hour
	orgFineMinPoints   = 3  // ٣٠ ألف (أقرب مضاعف ١٠ آلاف فوك ٢٥ ألف)
	orgFineMaxPoints   = 10 // ١٠٠ ألف
	orgAIUsageFullAt   = 30.0
	orgGovernorStarted = "started_at"
)

type OrgGovernor struct {
	db         *sqlx.DB
	org        *OrgService
	scores     *MatrixScoreService
	discipline *repository.DisciplineRepository
	actions    *repository.AiActionRepository
	markers    *repository.AiRepository
	switches   *repository.SystemSwitchRepository
	notify     *repository.NotificationRepository
	now        func() time.Time
}

func NewOrgGovernor(db *sqlx.DB, org *OrgService, scores *MatrixScoreService, d *repository.DisciplineRepository,
	a *repository.AiActionRepository, m *repository.AiRepository, sw *repository.SystemSwitchRepository, n *repository.NotificationRepository) *OrgGovernor {
	return &OrgGovernor{db: db, org: org, scores: scores, discipline: d, actions: a, markers: m, switches: sw, notify: n, now: time.Now}
}

func (g *OrgGovernor) on() bool {
	all, err := g.switches.All()
	return err == nil && all[model.SwitchOrgGovernor]
}

func (g *OrgGovernor) say(empID, kind, msg string) {
	if g.notify != nil && empID != "" {
		_ = g.notify.Create(empID, kind, msg)
	}
}

// startedAt أول تشغيل — الغرامات بس على شكاوي انفتحت بعده.
func (g *OrgGovernor) startedAt() time.Time {
	var v string
	_ = g.db.Get(&v, `INSERT INTO "OrgGovernorState" (key, value) VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET key = EXCLUDED.key RETURNING value`, orgGovernorStarted, g.now().UTC().Format(time.RFC3339))
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return g.now()
	}
	return t
}

// RegisterUndo يربط التراجع من صفحة القرارات: الغرامة ترجع نقاطها، والسحب يرجّع المنصب.
func (g *OrgGovernor) RegisterUndo(onUndo func(kind string, f func(model.AiAction) error)) {
	refund := func(a model.AiAction) error {
		var d struct {
			Points int `json:"appliedPoints"`
		}
		_ = json.Unmarshal(a.Details, &d)
		if a.TargetEmployeeID == nil || d.Points <= 0 {
			return nil
		}
		_, _, err := g.discipline.AdjustKind(*a.TargetEmployeeID, d.Points, model.DisciplineMatrixFineUndo, "تراجع عن غرامة ماتركس: "+a.Summary, "")
		if err == nil {
			g.say(*a.TargetEmployeeID, "ORG_FINE_UNDO", fmt.Sprintf("↩️ انلغت غرامة ماتركس (%s) ورجعتلك %d نقطة.", a.Summary, d.Points))
		}
		return err
	}
	onUndo(model.AiActionOrgFineCustomer, refund)
	onUndo(model.AiActionOrgFineOps, refund)
	onUndo(model.AiActionOrgRevoke, func(a model.AiAction) error {
		var d struct {
			AssignmentID string `json:"assignmentId"`
		}
		_ = json.Unmarshal(a.Details, &d)
		if d.AssignmentID == "" {
			return nil
		}
		if err := g.org.Restore("", d.AssignmentID); err != nil {
			return err
		}
		if a.TargetEmployeeID != nil {
			g.say(*a.TargetEmployeeID, "ORG_RESTORE", "↩️ رجع منصبك ومخصصه — انلغى قرار ماتركس بالسحب.")
		}
		_, _ = g.db.Exec(`UPDATE "OrgEvaluation" SET decision = 'WARN' WHERE "employeeId" = $1 AND decision = 'REVOKE' AND period = $2`, a.TargetEmployeeID, a.Period)
		return nil
	})
}

// protected المالك ومدير النظام ما ينحكمون.
func (g *OrgGovernor) protected(empID string) bool {
	role := g.org.TargetRole(empID)
	return role == "" || role == "OWNER" || role == "ADMIN"
}

// ── الغرامات (كل ساعة) ──

type orgComplaint struct {
	ID        string     `db:"id"`
	BookingID *string    `db:"bookingId"`
	Leader    *string    `db:"leaderId"`
	Customer  string     `db:"customerId"`
	Status    string     `db:"status"`
	CreatedAt time.Time  `db:"createdAt"`
	Resolved  *time.Time `db:"resolvedAt"`
}

// FinePoints تقدير ماتركس لحجم المشكلة (بلا ذكاء خارجي — من أرقام النظام):
// أساس ٣ نقاط، +٢ إذا ما انحلت خلال ٤٨ ساعة، +٢ لكل شكوى ثانية على نفس الموظف
// خلال ٣٠ يوم (لحد +٤)، +١ إذا نفس الزبون اشتكى گبل. السقف ١٠ (١٠٠ ألف).
func FinePoints(unresolved48 bool, priorOnEmployee, priorFromCustomer int) (int, []string) {
	pts := orgFineMinPoints
	why := []string{"شكوى زبون على حجزه"}
	if unresolved48 {
		pts += 2
		why = append(why, "ما انحلت خلال ٤٨ ساعة")
	}
	if priorOnEmployee > 0 {
		add := int(math.Min(float64(priorOnEmployee*2), 4))
		pts += add
		why = append(why, fmt.Sprintf("%d شكوى ثانية عليه خلال ٣٠ يوم", priorOnEmployee))
	}
	if priorFromCustomer > 0 {
		pts++
		why = append(why, "نفس الزبون اشتكى گبل")
	}
	if pts > orgFineMaxPoints {
		pts = orgFineMaxPoints
	}
	return pts, why
}

func (g *OrgGovernor) RunFines() {
	if !g.on() {
		return
	}
	since := g.startedAt()
	now := g.now()
	rows := []orgComplaint{}
	if err := g.db.Select(&rows, `SELECT c.id, c."bookingId", CASE WHEN c."bookingId" IS NULL THEN NULL ELSE booking_leader_id(c."bookingId") END AS "leaderId",
		c."customerId", c.status::text AS status, c."createdAt", c."resolvedAt"
		FROM "Complaint" c WHERE c."createdAt" >= $1 ORDER BY c."createdAt"`, since.UTC()); err != nil {
		log.Printf("[org-governor] الشكاوي: %v", err)
		return
	}
	var opsDirector string
	_ = g.db.Get(&opsDirector, `SELECT "employeeId" FROM "OrgAssignment" WHERE "divisionId" = 'ops' AND "unitId" IS NULL
		AND position = 'DIRECTOR' AND "endedAt" IS NULL AND NOT acting LIMIT 1`)
	for _, c := range rows {
		open := c.Status == "NEW" || c.Status == "IN_PROGRESS"
		age := now.Sub(c.CreatedAt)
		late := (open && age >= orgFineDeadline) || (c.Resolved != nil && c.Resolved.Sub(c.CreatedAt) >= orgFineDeadline)
		// (١) غرامة الموظف: نحكم بعد ما تنحل أو تفوت ٤٨ ساعة — حتى نعرف حجمها
		if c.Leader != nil && (!open || age >= orgFineDeadline) && !g.protected(*c.Leader) {
			var priorEmp, priorCust int
			_ = g.db.Get(&priorEmp, `SELECT count(*) FROM "Complaint" WHERE id <> $1 AND "bookingId" IS NOT NULL
				AND booking_leader_id("bookingId") = $2 AND "createdAt" BETWEEN $3::timestamp - interval '30 days' AND $3`, c.ID, *c.Leader, c.CreatedAt)
			_ = g.db.Get(&priorCust, `SELECT count(*) FROM "Complaint" WHERE id <> $1 AND "customerId" = $2 AND "createdAt" < $3`, c.ID, c.Customer, c.CreatedAt)
			pts, why := FinePoints(late, priorEmp, priorCust)
			g.fine(model.AiActionOrgFineCustomer, c.ID, *c.Leader, pts, why, c.BookingID)
		}
		// (٢) غرامة مدير العمليات: شكوى بقت ٤٨ ساعة بلا حل
		if late && opsDirector != "" && !g.protected(opsDirector) {
			over := age
			if c.Resolved != nil {
				over = c.Resolved.Sub(c.CreatedAt)
			}
			extraDays := int((over - orgFineDeadline).Hours() / 24)
			pts := orgFineMinPoints + extraDays
			if pts > orgFineMaxPoints {
				pts = orgFineMaxPoints
			}
			why := []string{fmt.Sprintf("شكوى بقت %d ساعة بلا حل (المهلة ٤٨)", int(over.Hours()))}
			g.fine(model.AiActionOrgFineOps, c.ID, opsDirector, pts, why, c.BookingID)
		}
	}
}

func (g *OrgGovernor) fine(kind, complaintID, empID string, pts int, why []string, bookingID *string) {
	name := g.actions.EmployeeName(empID)
	summary := fmt.Sprintf("غرامة %s د.ع — %s", formatIQD(pts*model.DisciplineDinarPerPoint), strings.Join(why, "، "))
	details, _ := json.Marshal(map[string]any{"complaintId": complaintID, "bookingId": bookingID, "points": pts, "reasons": why})
	target := empID
	ok, err := g.actions.Claim(model.AiAction{Kind: kind, EntityType: "COMPLAINT", EntityID: complaintID, Period: "once",
		TargetEmployeeID: &target, TargetLabel: name, Summary: summary, Details: details})
	if err != nil || !ok {
		return
	}
	_, applied, err := g.discipline.AdjustKind(empID, -pts, model.DisciplineMatrixFine, "ماتركس: "+summary, "")
	if err != nil {
		log.Printf("[org-governor] غرامة %s: %v", empID, err)
		return
	}
	// المطبّق فعلاً (الرصيد ما ينزل تحت الصفر) — التراجع يرجّع هذا بالضبط
	details, _ = json.Marshal(map[string]any{"complaintId": complaintID, "bookingId": bookingID, "points": pts, "appliedPoints": -applied, "reasons": why})
	_, _ = g.db.Exec(`UPDATE "AiAction" SET details = $3 WHERE kind = $1 AND "entityId" = $2 AND period = 'once'`, kind, complaintID, details)
	g.say(empID, "ORG_FINE", "🤖 ماتركس: "+summary+". التفاصيل بسجل الانضباط.")
	g.reportToDirector(empID, "🤖 ماتركس غرّم "+name+": "+summary)
}

// reportToDirector «يرفع تقرير بالقرارات للمدير» — مدير إدارة الموظف (ونسخة لمدير الإدارة).
func (g *OrgGovernor) reportToDirector(empID, msg string) {
	dirs := []string{}
	_ = g.db.Select(&dirs, `SELECT DISTINCT d."employeeId" FROM "OrgAssignment" a
		JOIN "OrgAssignment" d ON (d."divisionId" = a."divisionId" OR d."divisionId" = 'admin') AND d."unitId" IS NULL
			AND d.position = 'DIRECTOR' AND d."endedAt" IS NULL
		WHERE a."employeeId" = $1 AND a."endedAt" IS NULL AND d."employeeId" <> $1`, empID)
	for _, d := range dirs {
		g.say(d, "ORG_REPORT", msg)
	}
}

func formatIQD(n int) string {
	s := fmt.Sprint(n)
	out := ""
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out += ","
		}
		out += string(c)
	}
	return out
}

// ── التقييم الشهري (أول ٣ أيام من الشهر، على الشهر الي فات) ──

type OrgEvalComponents struct {
	Staff     *float64 `json:"staff"`
	Plan      *float64 `json:"plan,omitempty"`
	AIUsage   *float64 `json:"aiUsage,omitempty"`
	AIMsgs    int      `json:"aiMessages,omitempty"`
	Director  bool     `json:"director"`
	Principle bool     `json:"principlesAccepted"`
}

// OrgScore الدرجة: للعضو تقييم ماتركس الشهري (٦٠٪ ماتركس + ٤٠٪ بشر). للمدير/النائب:
// ٦٠٪ تقييمه + ٢٥٪ نسبة خطة إدارته + ١٥٪ استعماله للذكاء — والناقص يتوزّع على الموجود.
func OrgScore(c OrgEvalComponents) *float64 {
	type part struct {
		v *float64
		w float64
	}
	parts := []part{{c.Staff, 1}}
	if c.Director {
		parts = []part{{c.Staff, 0.60}, {c.Plan, 0.25}, {c.AIUsage, 0.15}}
	}
	sum, w := 0.0, 0.0
	for _, p := range parts {
		if p.v != nil {
			sum += *p.v * p.w
			w += p.w
		}
	}
	if w == 0 {
		return nil
	}
	v := math.Round(sum/w*10) / 10
	return &v
}

func (g *OrgGovernor) RunMonthlyIfDue() {
	if !g.on() {
		return
	}
	now := g.now().In(baghdad)
	if now.Day() > 3 {
		return
	}
	prev := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, baghdad).AddDate(0, -1, 0)
	period := prev.Format("2006-01")
	if ok, err := g.markers.ClaimDailyMarker("ORG_GOVERN_MONTH", prev.Format("2006-01-02")); err != nil || !ok {
		return
	}
	if err := g.EvaluateMonth(period); err != nil {
		log.Printf("[org-governor] تقييم %s: %v", period, err)
	}
}

// EvaluateMonth يقيّم كل صاحب منصب فعّال بالشهر ويقرر (تنبيه/سحب). ما يكرر نفس الشهر.
func (g *OrgGovernor) EvaluateMonth(period string) error {
	from, to, err := periodRange(period)
	if err != nil {
		return err
	}
	board, err := g.scores.Board(period)
	if err != nil {
		return err
	}
	staff := map[string]*float64{}
	for _, s := range board.Staff {
		staff[s.ID] = s.Final
	}
	type holder struct {
		Emp      string `db:"employeeId"`
		Director bool   `db:"director"`
		Division string `db:"divisionId"`
	}
	holders := []holder{}
	if err := g.db.Select(&holders, `SELECT "employeeId", bool_or(position IN ('DIRECTOR', 'DEPUTY') AND "unitId" IS NULL) AS director,
		COALESCE(min("divisionId") FILTER (WHERE position IN ('DIRECTOR', 'DEPUTY') AND "unitId" IS NULL), '') AS "divisionId"
		FROM "OrgAssignment" WHERE "endedAt" IS NULL AND "startedAt" < $1 GROUP BY "employeeId"`, to); err != nil {
		return err
	}
	decisions := map[string]int{}
	for _, h := range holders {
		if g.protected(h.Emp) {
			continue
		}
		var done int
		_ = g.db.Get(&done, `SELECT count(*) FROM "OrgEvaluation" WHERE "employeeId" = $1 AND period = $2`, h.Emp, period)
		if done > 0 {
			continue
		}
		c := OrgEvalComponents{Staff: staff[h.Emp], Director: h.Director}
		if h.Director && h.Division != "" {
			if p, err := g.org.plan(h.Division, period); err == nil {
				v := 0.0 // ماكو خطة = صفر — الخطة واجب المدير
				if p != nil {
					v = float64(p.Progress)
				}
				c.Plan = &v
			}
			var msgs int
			_ = g.db.Get(&msgs, `SELECT
				(SELECT count(*) FROM "MatrixTalkMessage" m JOIN "MatrixTalk" t ON t.id = m."chatId"
				  WHERE t."employeeId" = $1 AND m.role = 'USER' AND m."createdAt" >= $2 AND m."createdAt" < $3)
				+ (SELECT count(*) FROM "SecretaryMessage" s WHERE s."employeeId" = $1 AND s.role = 'user' AND s."createdAt" >= $2 AND s."createdAt" < $3)`,
				h.Emp, from.UTC(), to.UTC())
			u := math.Min(100, float64(msgs)/orgAIUsageFullAt*100)
			c.AIUsage, c.AIMsgs = &u, msgs
		}
		var accepted bool
		_ = g.db.Get(&accepted, `SELECT EXISTS (SELECT 1 FROM "PrivacyPolicyAcceptance" WHERE "employeeId" = $1 AND "pointsHash" IS NOT NULL)`, h.Emp)
		c.Principle = accepted
		score := OrgScore(c)
		decision := "NODATA"
		if score != nil {
			decision = "OK"
			if *score < orgRevokeBelow {
				decision = "WARN"
				prevPeriod := from.AddDate(0, -1, 0).Format("2006-01")
				var prevBad bool
				_ = g.db.Get(&prevBad, `SELECT EXISTS (SELECT 1 FROM "OrgEvaluation" WHERE "employeeId" = $1 AND period = $2
					AND decision IN ('WARN', 'REVOKE') AND score < $3)`, h.Emp, prevPeriod, orgRevokeBelow)
				if prevBad {
					decision = "REVOKE"
				}
			}
		}
		comp, _ := json.Marshal(c)
		res, err := g.db.Exec(`INSERT INTO "OrgEvaluation" ("employeeId", period, score, components, decision) VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT DO NOTHING`, h.Emp, period, score, comp, decision)
		if err != nil {
			continue
		}
		if n, _ := res.RowsAffected(); n == 0 {
			continue
		}
		decisions[decision]++
		switch decision {
		case "WARN":
			g.warn(h.Emp, period, *score, comp)
		case "REVOKE":
			g.revoke(h.Emp, period, *score, comp)
		}
	}
	if len(decisions) > 0 {
		msg := fmt.Sprintf("🏛️ ماتركس قيّم أصحاب المناصب لشهر %s: %d سليم، %d تنبيه، %d سحب منصب، %d بلا بيانات. التفاصيل بصفحة قرارات ماتركس (وتگدر تتراجع).",
			period, decisions["OK"], decisions["WARN"], decisions["REVOKE"], decisions["NODATA"])
		owners := []string{}
		_ = g.db.Select(&owners, `SELECT id FROM "Employee" WHERE role = 'OWNER' AND status = 'ACTIVE'`)
		for _, o := range owners {
			g.say(o, "ORG_REPORT", msg)
		}
	}
	return nil
}

func (g *OrgGovernor) warn(emp, period string, score float64, comp []byte) {
	name := g.actions.EmployeeName(emp)
	summary := fmt.Sprintf("تنبيه رسمي: تقييم %s = %.1f%% (تحت ٥٠٪). شهر ثاني تحت ٥٠٪ يسحب المنصب ومخصصه.", period, score)
	target := emp
	if ok, _ := g.actions.Claim(model.AiAction{Kind: model.AiActionOrgWarn, EntityType: "EMPLOYEE", EntityID: emp, Period: period,
		TargetEmployeeID: &target, TargetLabel: name, Summary: summary, Details: comp}); ok {
		g.say(emp, "ORG_WARN", "⚠️ ماتركس: "+summary)
		g.reportToDirector(emp, "⚠️ ماتركس نبّه "+name+": "+summary)
	}
}

func (g *OrgGovernor) revoke(emp, period string, score float64, comp []byte) {
	name := g.actions.EmployeeName(emp)
	ids := []string{}
	_ = g.db.Select(&ids, `SELECT id FROM "OrgAssignment" WHERE "employeeId" = $1 AND "endedAt" IS NULL`, emp)
	for _, id := range ids {
		reason := fmt.Sprintf("ماتركس: شهرين متتاليين تحت ٥٠٪ (آخرها %s = %.1f%%)", period, score)
		details, _ := json.Marshal(map[string]any{"assignmentId": id, "evaluation": json.RawMessage(comp)})
		target := emp
		ok, err := g.actions.Claim(model.AiAction{Kind: model.AiActionOrgRevoke, EntityType: "ORG_ASSIGNMENT", EntityID: id, Period: period,
			TargetEmployeeID: &target, TargetLabel: name, Summary: "سحب منصب ومخصصه — " + reason, Details: details})
		if err != nil || !ok {
			continue
		}
		if _, err := g.org.End("", id, reason); err != nil {
			log.Printf("[org-governor] سحب %s: %v", id, err)
			continue
		}
		g.say(emp, "ORG_REVOKE", "🏛️ "+reason+" — انسحب منصبك ومخصصه وصلاحياته. تگدر تراجع مديرك.")
		g.reportToDirector(emp, "🏛️ ماتركس سحب منصب "+name+": "+reason)
	}
}
