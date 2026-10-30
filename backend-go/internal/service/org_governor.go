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
	// orgFineWindow الغرامة بس إذا لحظة استحقاقها قريبة — حتى ما تنزل الغرامات المتراكمة
	// دفعة وحدة بعد ما المفتاح ينطفي ويرجع، أو النوع يتوقف ويرجع.
	orgFineWindow = 72 * time.Hour
	// orgMinMatrixPoints أقل بيانات للحكم: نقطة فاشلة وحدة ما تسوّي ٠٪ — لازم ١٠ نقاط
	// ماتركس أو تقييم بشري واحد على الأقل.
	orgMinMatrixPoints = 10
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

// ── التراجع ──
// يرجع خطأ إذا ما گدر يرجّع — فالقرار يرجع DONE (MatrixAutopilotService.Undo) وينعاد.

func (g *OrgGovernor) RegisterUndo(onUndo func(kind string, f func(model.AiAction) error)) {
	refund := func(a model.AiAction) error {
		var d struct {
			Points int `json:"appliedPoints"`
		}
		_ = json.Unmarshal(a.Details, &d)
		if a.TargetEmployeeID == nil || d.Points <= 0 {
			return nil
		}
		_, back, err := g.discipline.AdjustKind(*a.TargetEmployeeID, d.Points, model.DisciplineMatrixFineUndo, "تراجع عن غرامة ماتركس: "+a.Summary, "")
		if err != nil {
			return err
		}
		g.say(*a.TargetEmployeeID, "ORG_FINE_UNDO", fmt.Sprintf("↩️ انلغت غرامة ماتركس ورجعتلك %d نقطة (%s د.ع).", back, formatIQD(back*model.DisciplineDinarPerPoint)))
		return nil
	}
	onUndo(model.AiActionOrgFineCustomer, refund)
	onUndo(model.AiActionOrgFineOps, refund)
	onUndo(model.AiActionOrgWarn, func(a model.AiAction) error {
		// التنبيه انلغى ← السلسلة تنكسر (الشهر الجاي ما يسحب على هالتنبيه)
		_, err := g.db.Exec(`UPDATE "OrgEvaluation" SET decision = 'OK' WHERE "employeeId" = $1 AND period = $2 AND decision = 'WARN'`, a.TargetEmployeeID, a.Period)
		return err
	})
	onUndo(model.AiActionOrgRevoke, func(a model.AiAction) error {
		var d struct {
			IDs []string `json:"assignmentIds"`
		}
		_ = json.Unmarshal(a.Details, &d)
		for _, id := range d.IDs {
			var ended bool
			if err := g.db.Get(&ended, `SELECT "endedAt" IS NOT NULL FROM "OrgAssignment" WHERE id = $1`, id); err != nil {
				return err
			}
			if !ended {
				continue // رجع من محاولة سابقة
			}
			if err := g.org.Restore("", id); err != nil {
				if err == ErrOrgSeatTaken {
					return fmt.Errorf("ما گدرت أرجّع المنصب — مكانه انشغل. أنهِ منصب البديل أول وعيد التراجع")
				}
				return err
			}
		}
		if _, err := g.db.Exec(`UPDATE "OrgEvaluation" SET decision = 'WARN' WHERE "employeeId" = $1 AND decision = 'REVOKE' AND period = $2`, a.TargetEmployeeID, a.Period); err != nil {
			return err
		}
		if a.TargetEmployeeID != nil {
			g.say(*a.TargetEmployeeID, "ORG_RESTORE", "↩️ رجع منصبك ومخصصه — انلغى قرار ماتركس بالسحب.")
		}
		return nil
	})
}

// protected المالك ومدير النظام ما ينحكمون (والي ما نعرف دوره ما نحكم عليه).
func (g *OrgGovernor) protected(empID string) bool {
	role := g.org.TargetRole(empID)
	return role == "" || role == "OWNER" || role == "ADMIN"
}

// ── الغرامات (كل ساعة) ──

// orgComplaintPerson الشخص الي الشكوى عليه: المسجّل بالشكوى أول، وإلا ليدر الحجز
// (نفس قاعدة complaintLeaderExpr بـai_repository).
const orgComplaintPerson = `COALESCE(c."relatedEmployeeId", CASE WHEN c."bookingId" IS NULL THEN NULL ELSE booking_leader_id(c."bookingId") END)`

type orgComplaint struct {
	ID        string     `db:"id"`
	BookingID *string    `db:"bookingId"`
	Person    *string    `db:"personId"`
	Customer  string     `db:"customerId"`
	Status    string     `db:"status"`
	NoAnswer  bool       `db:"noAnswer"`
	Dropped   bool       `db:"dropped"`
	CreatedAt time.Time  `db:"createdAt"`
	Resolved  *time.Time `db:"resolvedAt"`
}

// fineKey: غرامة الموظف وحدة لكل حجز ولكل شخص (شكاوي مكررة على نفس الحجز ما تتراكم،
// وشخص ثاني مذنب بنفس الحجز ما يفلت). غرامة مدير العمليات وحدة لكل حجز (حادثة وحدة).
func (c orgComplaint) fineKey(kind, empID string) string {
	if c.BookingID == nil {
		return "COMPLAINT:" + c.ID
	}
	if kind == model.AiActionOrgFineCustomer {
		return "BOOKING:" + *c.BookingID + ":" + empID
	}
	return "BOOKING:" + *c.BookingID
}

// FinePoints تقدير ماتركس لحجم المشكلة (بلا ذكاء خارجي — من أرقام النظام):
// أساس ٣ نقاط، +٢ إذا ما انحلت خلال ٤٨ ساعة، +٢ لكل حجز ثاني عليه شكوى على نفس
// الموظف خلال ٣٠ يوم (لحد +٤)، +١ إذا نفس الزبون اشتكى گبل. السقف ١٠ (١٠٠ ألف).
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

// fineFloor أقدم لحظة استحقاق تنحسب: بعد أول تشغيل، وبعد آخر مرة المفتاح انشغّل،
// وبآخر ٧٢ ساعة — فالمتراكم بالتوقف ما ينزل دفعة وحدة.
func (g *OrgGovernor) fineFloor(kind string) time.Time {
	floor := g.startedAt()
	var on *time.Time
	_ = g.db.Get(&on, `SELECT "updatedAt" FROM "SystemSwitch" WHERE key = $1 AND enabled`, model.SwitchOrgGovernor)
	if on != nil && on.After(floor) {
		floor = *on
	}
	// النوع انوقف (٣ رفضات) ورجع — ما ننزل المتراكم بالتوقف
	var resumed *time.Time
	_ = g.db.Get(&resumed, `SELECT "resumedAt" FROM "AiActionKindPause" WHERE kind = $1 AND "resumedAt" IS NOT NULL`, kind)
	if resumed != nil && resumed.After(floor) {
		floor = *resumed
	}
	if w := g.now().Add(-orgFineWindow); w.After(floor) {
		floor = w
	}
	return floor
}

func (g *OrgGovernor) RunFines() {
	if !g.on() {
		return
	}
	since := g.startedAt()
	floorCust, floorOps := g.fineFloor(model.AiActionOrgFineCustomer), g.fineFloor(model.AiActionOrgFineOps)
	now := g.now()
	rows := []orgComplaint{}
	if err := g.db.Select(&rows, `SELECT c.id, c."bookingId", `+orgComplaintPerson+` AS "personId",
		c."customerId", c.status::text AS status,
		-- معفية بس إذا الزبون ما چان ينوصل وقت المهلة (سقطت، أو المكالمة الجاية بعد الـ٤٨)
		(c."droppedNoAnswerAt" IS NOT NULL OR (c."nextCallAt" IS NOT NULL AND c."nextCallAt" >= c."createdAt" + interval '48 hours')) AS "noAnswer",
		c."droppedNoAnswerAt" IS NOT NULL AS dropped,
		c."createdAt", c."resolvedAt"
		FROM "Complaint" c WHERE c."createdAt" >= $1 ORDER BY c."createdAt"`, since.UTC()); err != nil {
		log.Printf("[org-governor] الشكاوي: %v", err)
		return
	}
	for _, c := range rows {
		open := c.Status == "NEW" || c.Status == "IN_PROGRESS"
		deadline := c.CreatedAt.Add(orgFineDeadline)
		// متأخرة = فاتت ٤٨ ساعة وهي مفتوحة، أو انحلت بعد ٤٨ (مغلقة بلا وقت = ما نعرف، ما نحسبها)
		late := (open && now.After(deadline)) || (c.Resolved != nil && c.Resolved.After(deadline))
		// لحظة الاستحقاق: الإغلاق إذا گبل الموعد، وإلا الموعد نفسه
		due := deadline
		if !open && c.Resolved != nil && c.Resolved.Before(deadline) {
			due = *c.Resolved
		}
		if !open && c.Resolved == nil {
			due = deadline
		}
		if due.After(now) {
			continue
		}
		// (١) غرامة الموظف الي الشكوى عليه — وحدة لكل حجز ولكل شخص. الشكوى الي سقطت لأن
		// الزبون ما رد أبداً ما تنحسب عليه.
		if c.Person != nil && !c.Dropped && !due.Before(floorCust) && !g.protected(*c.Person) {
			// ننتظر لحد ما كل شكاوي نفس الحجز على نفس الشخص تستحق، والتأخير لو أي وحدة تأخرت
			var pending, anyLate bool
			if c.BookingID != nil {
				_ = g.db.Get(&pending, `SELECT EXISTS (SELECT 1 FROM "Complaint" c WHERE c."bookingId" = $1 AND `+orgComplaintPerson+` = $2
					AND c.status IN ('NEW', 'IN_PROGRESS') AND c."createdAt" + interval '48 hours' > $3)`, *c.BookingID, *c.Person, now.UTC())
				_ = g.db.Get(&anyLate, `SELECT EXISTS (SELECT 1 FROM "Complaint" c WHERE c."bookingId" = $1 AND `+orgComplaintPerson+` = $2
					AND c."droppedNoAnswerAt" IS NULL AND NOT (c."nextCallAt" IS NOT NULL AND c."nextCallAt" >= c."createdAt" + interval '48 hours')
					AND ((c.status IN ('NEW', 'IN_PROGRESS') AND c."createdAt" + interval '48 hours' <= $3) OR c."resolvedAt" > c."createdAt" + interval '48 hours'))`,
					*c.BookingID, *c.Person, now.UTC())
			} else {
				anyLate = late && !c.NoAnswer
			}
			if !pending {
				var priorEmp, priorCust int
				_ = g.db.Get(&priorEmp, `SELECT count(DISTINCT COALESCE(c."bookingId", c.id)) FROM "Complaint" c
				WHERE c.id <> $1 AND `+orgComplaintPerson+` = $2 AND c."createdAt" >= $4
				AND c."createdAt" BETWEEN $3::timestamp - interval '30 days' AND $3
				AND ($5::text IS NULL OR c."bookingId" IS DISTINCT FROM $5)
				AND c."droppedNoAnswerAt" IS NULL
				AND NOT EXISTS (SELECT 1 FROM "AiAction" a WHERE a.kind = $6 AND a.status = 'UNDONE'
					AND a."entityId" = COALESCE('BOOKING:' || c."bookingId" || ':' || $2, 'COMPLAINT:' || c.id))`,
					c.ID, *c.Person, c.CreatedAt, since.UTC(), c.BookingID, model.AiActionOrgFineCustomer)
				_ = g.db.Get(&priorCust, `SELECT count(DISTINCT COALESCE("bookingId", id)) FROM "Complaint"
				WHERE id <> $1 AND "customerId" = $2 AND "createdAt" < $3 AND "createdAt" >= $4
				AND ($5::text IS NULL OR "bookingId" IS DISTINCT FROM $5) AND "droppedNoAnswerAt" IS NULL`,
					c.ID, c.Customer, c.CreatedAt, since.UTC(), c.BookingID)
				pts, why := FinePoints(anyLate, priorEmp, priorCust)
				g.fine(model.AiActionOrgFineCustomer, c, *c.Person, pts, why)
			}
		}
		// (٢) مدير العمليات: شكوى بقت ٤٨ ساعة بلا حل — إلا إذا الزبون هو الي ما يرد
		// الماسك وقت الاستحقاق: المدير بالوكالة إذا الأصلي بإجازة (المبدأ ٢٠)، وإلا الأصلي
		var opsDirector string
		if late && !c.NoAnswer && !due.Before(floorOps) {
			_ = g.db.Get(&opsDirector, `SELECT "employeeId" FROM "OrgAssignment" WHERE "divisionId" = 'ops' AND "unitId" IS NULL
				AND position = 'DIRECTOR' AND "startedAt" <= $1 AND ("endedAt" IS NULL OR "endedAt" > $1)
				ORDER BY acting DESC, "startedAt" DESC LIMIT 1`, due.UTC())
		}
		if opsDirector != "" && !g.protected(opsDirector) {
			pts := orgFineMinPoints
			why := []string{"شكوى بقت أكثر من ٤٨ ساعة بلا حل"}
			if c.Resolved != nil {
				extra := int(c.Resolved.Sub(deadline).Hours() / 24)
				if extra > 0 {
					pts += extra
					why = []string{fmt.Sprintf("شكوى انحلت بعد %d ساعة (المهلة ٤٨)", int(c.Resolved.Sub(c.CreatedAt).Hours()))}
				}
			}
			if pts > orgFineMaxPoints {
				pts = orgFineMaxPoints
			}
			g.fine(model.AiActionOrgFineOps, c, opsDirector, pts, why)
		}
	}
}

func (g *OrgGovernor) fine(kind string, c orgComplaint, empID string, pts int, why []string) {
	name := g.actions.EmployeeName(empID)
	key := c.fineKey(kind, empID)
	target := empID
	details := func(applied int) []byte {
		b, _ := json.Marshal(map[string]any{"complaintId": c.ID, "bookingId": c.BookingID, "points": pts, "appliedPoints": applied, "reasons": why})
		return b
	}
	summary := func(p int) string {
		return fmt.Sprintf("غرامة %s د.ع — %s", formatIQD(p*model.DisciplineDinarPerPoint), strings.Join(why, "، "))
	}
	ok, err := g.actions.Claim(model.AiAction{Kind: kind, EntityType: "COMPLAINT", EntityID: key, Period: "once",
		TargetEmployeeID: &target, TargetLabel: name, Summary: summary(pts), Details: details(0)})
	if err != nil || !ok {
		return
	}
	_, applied, err := g.discipline.AdjustKind(empID, -pts, model.DisciplineMatrixFine, "ماتركس: "+summary(pts), "")
	if err != nil {
		// الخصم فشل ← نشيل الحجز حتى ينعاد بالمرة الجاية (ما يبقى قرار بلا خصم)
		log.Printf("[org-governor] غرامة %s: %v", empID, err)
		_ = g.actions.DeleteClaim(kind, key, "once")
		return
	}
	got := -applied // المخصوم فعلاً (الرصيد ما ينزل تحت الصفر) — التراجع يرجّع هذا بالضبط
	if _, err := g.db.Exec(`UPDATE "AiAction" SET details = $3, summary = $4 WHERE kind = $1 AND "entityId" = $2 AND period = 'once'`,
		kind, key, details(got), summary(got)); err != nil {
		log.Printf("[org-governor] تسجيل المخصوم %s: %v", empID, err)
	}
	if got <= 0 {
		return
	}
	g.say(empID, "ORG_FINE", "🤖 ماتركس: "+summary(got)+". التفاصيل بسجل الانضباط.")
	g.reportToDirector(empID, "🤖 ماتركس غرّم "+name+": "+summary(got))
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
// ٦٠٪ تقييمه + ٢٥٪ نسبة خطة إدارته + ١٥٪ استعماله للذكاء.
// بلا تقييم ماتركس = ما نحكم (nil) — حتى المدير، ما ننحكم على الخطة والذكاء لحالهم.
func OrgScore(c OrgEvalComponents) *float64 {
	if c.Staff == nil {
		return nil
	}
	v := *c.Staff
	if c.Director {
		if c.Plan == nil || c.AIUsage == nil {
			return nil
		}
		v = *c.Staff*0.60 + *c.Plan*0.25 + *c.AIUsage*0.15
	}
	v = math.Round(v*10) / 10
	return &v
}

// RunMonthlyIfDue كل ساعة أول ٣ أيام من الشهر (بغداد). ما يكرر — كل موظف ينقيّم مرة بالشهر،
// فخطأ بنص التشغيل ينعاد بالساعة الجاية بدل ما يضيع الشهر.
func (g *OrgGovernor) RunMonthlyIfDue() {
	if !g.on() {
		return
	}
	now := g.now().In(baghdad)
	// يوم ٣ بس — حتى درجات آخر يوم بالشهر تكتمل (المقيّم يرجع ٣ أيام)، وبعدها كل ساعة
	// باليوم يعيد للي ما انقيّم
	if now.Day() != 3 {
		return
	}
	prev := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, baghdad).AddDate(0, -1, 0)
	if err := g.EvaluateMonth(prev.Format("2006-01")); err != nil {
		log.Printf("[org-governor] تقييم %s: %v", prev.Format("2006-01"), err)
	}
}

// EvaluateMonth يقيّم كل صاحب منصب ماسكه من أول الشهر، ويقرر (تنبيه/سحب). ما يكرر.
// الشهر الي بدأ گبل أول تشغيل للحاكم ما ينقيّم (الهيكل ما چان موجود كامل).
func (g *OrgGovernor) EvaluateMonth(period string) error {
	from, to, err := periodRange(period)
	if err != nil {
		return err
	}
	if from.Before(g.startedAt()) {
		return nil
	}
	board, err := g.scores.Board(period)
	if err != nil {
		return err
	}
	staff := map[string]*float64{}
	for _, s := range board.Staff {
		if s.HumanCount >= 1 || s.Max >= orgMinMatrixPoints {
			staff[s.ID] = s.Final
		}
	}
	type holder struct {
		Emp      string `db:"employeeId"`
		Director bool   `db:"director"`
		Division string `db:"divisionId"`
	}
	holders := []holder{}
	if err := g.db.Select(&holders, `SELECT "employeeId", bool_or(position IN ('DIRECTOR', 'DEPUTY') AND "unitId" IS NULL) AS director,
		COALESCE(min("divisionId") FILTER (WHERE position IN ('DIRECTOR', 'DEPUTY') AND "unitId" IS NULL), '') AS "divisionId"
		FROM "OrgAssignment" WHERE "endedAt" IS NULL AND NOT acting AND "startedAt" <= $1 GROUP BY "employeeId"`, from); err != nil {
		return err
	}
	decisions := map[string]int{}
	for _, h := range holders {
		if g.protected(h.Emp) {
			continue
		}
		var done int
		if err := g.db.Get(&done, `SELECT count(*) FROM "OrgEvaluation" WHERE "employeeId" = $1 AND period = $2`, h.Emp, period); err != nil || done > 0 {
			continue
		}
		c := OrgEvalComponents{Staff: staff[h.Emp], Director: h.Director}
		noData := c.Staff == nil
		if h.Director && h.Division != "" && !noData {
			if p, err := g.org.plan(h.Division, period); err != nil {
				noData = true
			} else {
				v := 0.0 // ماكو خطة = صفر — الخطة واجب المدير
				if p != nil {
					v = float64(p.Progress)
				}
				c.Plan = &v
			}
			var msgs int
			if err := g.db.Get(&msgs, `SELECT
				(SELECT count(*) FROM "MatrixTalkMessage" m JOIN "MatrixTalk" t ON t.id = m."chatId"
				  WHERE t."employeeId" = $1 AND m.role = 'USER' AND m."createdAt" >= $2 AND m."createdAt" < $3)
				+ (SELECT count(*) FROM "SecretaryMessage" s WHERE s."employeeId" = $1 AND s.role = 'user' AND s."createdAt" >= $2 AND s."createdAt" < $3)`,
				h.Emp, from.UTC(), to.UTC()); err != nil {
				noData = true
			} else {
				u := math.Min(100, float64(msgs)/orgAIUsageFullAt*100)
				c.AIUsage, c.AIMsgs = &u, msgs
			}
		}
		var accepted bool
		_ = g.db.Get(&accepted, `SELECT EXISTS (SELECT 1 FROM "PrivacyPolicyAcceptance" WHERE "employeeId" = $1 AND "pointsHash" IS NOT NULL)`, h.Emp)
		c.Principle = accepted
		score := OrgScore(c)
		if noData {
			score = nil
		}
		comp, _ := json.Marshal(c)
		if score == nil {
			continue // بلا بيانات كافية: ما نسجّل حكم، يمكن تكتمل بالساعة الجاية
		}
		// القرار ينكتب بعد التنفيذ بالي صار فعلاً
		decision := "NODATA"
		if score != nil {
			decision = "OK"
			if *score < orgRevokeBelow {
				// السحب يحتاج تنبيه وصل فعلاً الشهر الي فات (ما انمنع ولا انلغى)
				prevPeriod := from.AddDate(0, -1, 0).Format("2006-01")
				var warned bool
				_ = g.db.Get(&warned, `SELECT EXISTS (SELECT 1 FROM "AiAction" WHERE kind = $1 AND "entityId" = $2 AND period = $3 AND status = 'DONE')`,
					model.AiActionOrgWarn, orgEvalEntity(h.Emp, prevPeriod), prevPeriod)
				var already bool
				_ = g.db.Get(&already, `SELECT EXISTS (SELECT 1 FROM "AiAction" WHERE kind = $1 AND "entityId" = $2 AND status = 'DONE')`,
					model.AiActionOrgWarn, orgEvalEntity(h.Emp, period))
				switch {
				case warned && g.revoke(h.Emp, period, *score, comp):
					decision = "REVOKE"
				case already || g.warn(h.Emp, period, *score, comp):
					decision = "WARN"
				default:
					continue // التنبيه ما انسجّل (موقوف/خطأ) — ما نسجّل حكم، ونعيد بالساعة الجاية
				}
			}
		}
		res, err := g.db.Exec(`INSERT INTO "OrgEvaluation" ("employeeId", period, score, components, decision) VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT DO NOTHING`, h.Emp, period, score, comp, decision)
		if err != nil {
			log.Printf("[org-governor] تسجيل تقييم %s: %v", h.Emp, err)
			continue
		}
		if n, _ := res.RowsAffected(); n > 0 {
			decisions[decision]++
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

func (g *OrgGovernor) warn(emp, period string, score float64, comp []byte) bool {
	name := g.actions.EmployeeName(emp)
	summary := fmt.Sprintf("تنبيه رسمي: تقييم %s = %.1f%% (تحت ٥٠٪). شهر ثاني تحت ٥٠٪ يسحب المنصب ومخصصه.", period, score)
	target := emp
	ok, err := g.actions.Claim(model.AiAction{Kind: model.AiActionOrgWarn, EntityType: "EMPLOYEE", EntityID: orgEvalEntity(emp, period), Period: period,
		TargetEmployeeID: &target, TargetLabel: name, Summary: summary, Details: comp})
	if err != nil || !ok {
		return false
	}
	g.say(emp, "ORG_WARN", "⚠️ ماتركس: "+summary)
	g.reportToDirector(emp, "⚠️ ماتركس نبّه "+name+": "+summary)
	return true
}

// orgEvalEntity مفتاح قرار التقييم: الشخص + الشهر — رفض قرار شهر ما يمنع قرار شهر ثاني
// (الحجب ٣٠ يوم بـClaim مصمم للتذكيرات، مو لأحكام شهرية).
func orgEvalEntity(emp, period string) string { return "EMP:" + emp + ":" + period }

// revoke قرار سحب واحد للشخص (كل مناصبه) — حتى تراجع شخص واحد ما يوقّف النوع كله.
func (g *OrgGovernor) revoke(emp, period string, score float64, comp []byte) bool {
	name := g.actions.EmployeeName(emp)
	ids := []string{}
	if err := g.db.Select(&ids, `SELECT id FROM "OrgAssignment" WHERE "employeeId" = $1 AND "endedAt" IS NULL ORDER BY "startedAt"`, emp); err != nil || len(ids) == 0 {
		return false
	}
	reason := fmt.Sprintf("ماتركس: شهرين متتاليين تحت ٥٠٪ (آخرها %s = %.1f%%)", period, score)
	details, _ := json.Marshal(map[string]any{"assignmentIds": ids, "evaluation": json.RawMessage(comp)})
	target := emp
	ok, err := g.actions.Claim(model.AiAction{Kind: model.AiActionOrgRevoke, EntityType: "EMPLOYEE", EntityID: orgEvalEntity(emp, period), Period: period,
		TargetEmployeeID: &target, TargetLabel: name, Summary: "سحب منصب ومخصصه — " + reason, Details: details})
	if err != nil || !ok {
		return false
	}
	ended := []string{}
	for _, id := range ids {
		if _, err := g.org.End("", id, reason); err != nil {
			log.Printf("[org-governor] سحب %s: %v", id, err)
			continue
		}
		ended = append(ended, id)
	}
	if len(ended) == 0 {
		_ = g.actions.DeleteClaim(model.AiActionOrgRevoke, orgEvalEntity(emp, period), period)
		return false
	}
	if len(ended) != len(ids) {
		details, _ = json.Marshal(map[string]any{"assignmentIds": ended, "evaluation": json.RawMessage(comp)})
		_, _ = g.db.Exec(`UPDATE "AiAction" SET details = $3 WHERE kind = $1 AND "entityId" = $2 AND period = $4`, model.AiActionOrgRevoke, orgEvalEntity(emp, period), details, period)
	}
	g.say(emp, "ORG_REVOKE", "🏛️ "+reason+" — انسحب منصبك ومخصصه وصلاحياته. تگدر تراجع مديرك.")
	g.reportToDirector(emp, "🏛️ ماتركس سحب منصب "+name+": "+reason)
	return true
}
