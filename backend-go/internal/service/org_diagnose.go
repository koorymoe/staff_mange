package service

import (
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ═══ 🔎 التشخيص الذكي (دازاي 10-30 — النقطة ٤) ═══
// «طريقة تشخيص المشاكل وتحديدها حسب كمال وماتركس استناداً على الوحدات ومسؤوليتها».
// ماتركس يقرا أرقام كل وحدة (من بيانات النظام بس — بلا ذكاء خارجي ولا أسماء)، وإذا لگى
// إشارة يفتح «مشكلة» بمكتب الإدارة مربوطة بفقرة اللائحة، ويبلّغ رئيس الوحدة يكتب خطة
// الحل. إذا الإشارة بقت يحدّث أرقامها، وإذا انحلت ٣ أيام يسكّرها لحاله.
// المشاكل الي يكتبها البشر ما ينلمسن. المفتاح org_diagnosis.

const orgDiagnoseResolveAfter = 3 * 24 * time.Hour

// orgSignal إشارة لوحدة: count فوك الحد = مشكلة.
type orgSignal struct {
	Key, Unit, Title, Impact string
	Clause                   int // رقم الفقرة (principleNo)
	Count                    int
	Detail                   string
	High                     bool
}

// orgActionSignals تذكيرات ماتركس المفتوحة (آخر ٧ أيام) ← وحدة مسؤولة وفقرة.
var orgActionSignals = []struct {
	Unit, Title, Impact string
	Clause, Min         int
	Kinds               []string
}{
	{"supply.procurement", "طلبات مواد وعُدد معلّقة ونواقص جرد ما انحلت", "الكوادر تشتغل بنقص عُدد ومواد، والحجوزات تتأخر", 11, 3,
		[]string{model.AiActionProcPending, model.AiActionShortageOpen, model.AiActionLowStock}},
	{"ops.fleet", "صيانة ووثائق مركبات فاتت مواعيدها وعُدد ما رجعت", "مركبات وعُدد مو جاهزة، وخطر على الشغل والسلامة", 12, 3,
		[]string{model.AiActionFleetOverdue, model.AiActionVehiclesUnrated, model.AiActionToolReturn}},
	{"eng.projects", "مشاريع متجاوزة حدود مراحلها أو واقفة", "تأخير تسليم المشاريع وزعل الزبائن", 7, 3,
		[]string{model.AiActionProjectOverLimit, model.AiActionProjectIdleToday}},
	{"admin.accounts", "دفعات وفواتير تنتظر تدقيق أو اعتماد", "فلوس ما متأكدة وأعمال بلا فواتير", 13, 3,
		[]string{model.AiActionPaymentsUnverified, model.AiActionInternalNoInvoice, model.AiActionInvoiceApproval}},
	{"ops.bookings", "حجوزات بلا كادر أو ورقها متأخر", "تأخير الحجوزات وضياع متابعة الزبون", 15, 3,
		[]string{model.AiActionUnstaffedAlert, model.AiActionPaperworkReminder, model.AiActionDelayWarning}},
	{"ops.staff", "غياب بلا تسجيل ومهام إضافية فاتت مواعيدها", "كادر غير ملتزم بالدوام والتكليفات", 3, 3,
		[]string{model.AiActionAttendanceNudge, model.AiActionExtraTaskOverdue}},
	{"quality.followup", "زبائن بحاجة متابعة قريبين يزعلون", "خطر خسارة زبائن وسمعة الشركة", 16, 2,
		[]string{model.AiActionCustomerFollowUp}},
}

func (s *OrgService) collectSignals() []orgSignal {
	out := []orgSignal{}
	// (١) تذكيرات ماتركس المفتوحة
	for _, a := range orgActionSignals {
		var n int
		if err := s.db.Get(&n, `SELECT count(*) FROM "AiAction" WHERE kind = ANY($1) AND status = 'DONE' AND "resolvedAt" IS NULL
			AND "createdAt" > now() - interval '7 days'`, pq.Array(a.Kinds)); err != nil || n < a.Min {
			continue
		}
		out = append(out, orgSignal{Key: "reminders:" + a.Unit, Unit: a.Unit, Title: a.Title, Impact: a.Impact, Clause: a.Clause,
			Count: n, Detail: fmt.Sprintf("%d تذكير مفتوح من ماتركس آخر ٧ أيام", n), High: n >= a.Min*3})
	}
	// (٢) شكاوي فاتت ٤٨ ساعة بلا حل (والزبون ينوصل)
	var late int
	if err := s.db.Get(&late, `SELECT count(*) FROM "Complaint" c WHERE c.status IN ('NEW', 'IN_PROGRESS')
		AND c."createdAt" < now() - interval '48 hours' AND c."droppedNoAnswerAt" IS NULL
		AND NOT (c."nextCallAt" IS NOT NULL AND c."nextCallAt" >= c."createdAt" + interval '48 hours')`); err == nil && late > 0 {
		out = append(out, orgSignal{Key: "complaints_late", Unit: "quality.followup", Title: "شكاوي زبائن فاتت ٤٨ ساعة بلا حل",
			Impact: "زبائن زعلانين، وغرامات على مدير العمليات", Clause: 16, Count: late,
			Detail: fmt.Sprintf("%d شكوى مفتوحة من أكثر من ٤٨ ساعة", late), High: late >= 3})
	}
	// (٣) صندوق المراقب متراكم
	var backlog int
	if err := s.db.Get(&backlog, `SELECT count(*) FROM "MonitorReview" WHERE status = 'PENDING' AND "createdAt" < now() - interval '48 hours'`); err == nil && backlog >= 10 {
		out = append(out, orgSignal{Key: "monitor_backlog", Unit: "admin.monitoring", Title: "صندوق المراقب متراكم بلا قرار",
			Impact: "أخطاء تمر بلا تدقيق ولا محاسبة", Clause: 4, Count: backlog,
			Detail: fmt.Sprintf("%d بند بالصندوق ينتظر من أكثر من ٤٨ ساعة", backlog), High: backlog >= 30})
	}
	// (٤) لكل وحدة: نسبة تأخير كادرها آخر ١٤ يوم، وبنود خطتها المتأخرة
	units := []struct {
		ID string `db:"id"`
	}{}
	_ = s.db.Select(&units, `SELECT id FROM "OrgUnit" ORDER BY "divisionId", sort`)
	for _, u := range units {
		var r struct {
			Late  int `db:"late"`
			Total int `db:"total"`
		}
		if err := s.db.Get(&r, `SELECT count(*) FILTER (WHERE points < "maxPoints") AS late, count(*) AS total FROM "MatrixScore"
			WHERE "cancelledAt" IS NULL AND "at" > now() - interval '14 days'
			AND "employeeId" IN (SELECT "employeeId" FROM "OrgAssignment" WHERE "unitId" = $1 AND "endedAt" IS NULL)`, u.ID); err == nil && r.Total >= 10 {
			rate := float64(r.Late) / float64(r.Total) * 100
			if rate > 25 {
				out = append(out, orgSignal{Key: "late_rate:" + u.ID, Unit: u.ID, Title: "تأخير عالي بمحطات شغل كادر الوحدة",
					Impact: "الشغل ما ينجز بوقته والزبون ينتظر", Clause: 7, Count: r.Late,
					Detail: fmt.Sprintf("%.0f٪ من محطات الكادر متأخرة أو فايتة آخر ١٤ يوم (%d من %d)", rate, r.Late, r.Total), High: rate > 40})
			}
		}
		var overdue int
		if err := s.db.Get(&overdue, `SELECT count(*) FROM "OrgPlanItem" i JOIN "OrgPlan" p ON p.id = i."planId"
			WHERE p."unitId" = $1 AND i."doneAt" IS NULL AND i."dueDate" < (now() AT TIME ZONE 'Asia/Baghdad')::date`, u.ID); err == nil && overdue >= 2 {
			out = append(out, orgSignal{Key: "plan_overdue:" + u.ID, Unit: u.ID, Title: "بنود خطة التطوير فاتت مواعيدها",
				Impact: "خطة التطوير ما تمشي والمشاكل تبقى", Clause: 4, Count: overdue,
				Detail: fmt.Sprintf("%d بند بخطة الوحدة فات موعده وما خلص", overdue), High: overdue >= 5})
		}
	}
	return out
}

// RunDiagnosis جولة تشخيص: يفتح/يحدّث/يسكّر مشاكل ماتركس. notify لرؤساء الوحدة.
func (s *OrgService) RunDiagnosis(switches *repository.SystemSwitchRepository, notify func(empID, msg string)) (opened, closed int) {
	if switches != nil {
		if all, err := switches.All(); err != nil || !all[model.SwitchOrgDiagnosis] {
			return 0, 0
		}
	}
	seen := map[string]bool{}
	for _, sig := range s.collectSignals() {
		seen[sig.Key] = true
		var div string
		if err := s.db.Get(&div, `SELECT "divisionId" FROM "OrgUnit" WHERE id = $1`, sig.Unit); err != nil {
			continue
		}
		sev := "MEDIUM"
		if sig.High {
			sev = "HIGH"
		}
		problem := "🤖 ماتركس شخّصها من أرقام النظام: " + sig.Detail + "."
		res, err := s.db.Exec(`UPDATE "OrgIssueReport" SET problem = $2, severity = $3, "lastSeenAt" = now()
			WHERE signal = $1 AND status <> 'RESOLVED'`, sig.Key, problem, sev)
		if err != nil {
			continue
		}
		if n, _ := res.RowsAffected(); n > 0 {
			continue
		}
		var id string
		if err := s.db.Get(&id, `INSERT INTO "OrgIssueReport" ("divisionId", "unitId", title, problem, impact, scope, principle, severity, source, signal, "lastSeenAt")
			VALUES ($1, $2, $3, $4, $5, 'SPECIFIC', $6, $7, 'MATRIX', $8, now()) ON CONFLICT DO NOTHING RETURNING id`,
			div, sig.Unit, sig.Title, problem, sig.Impact, sig.Clause, sev, sig.Key); err != nil || id == "" {
			continue
		}
		opened++
		if notify != nil {
			code := ""
			_ = s.db.Get(&code, `SELECT split_part(content, ' — ', 1) FROM "PrivacyPolicyPoint" WHERE "principleNo" = $1`, sig.Clause)
			for _, h := range s.unitHeads(div, sig.Unit) {
				notify(h, fmt.Sprintf("🔎 ماتركس شخّص مشكلة بوحدتك: «%s» (%s) — %s. اكتب خطة الحل بمكتب الإدارة.", sig.Title, code, sig.Detail))
			}
		}
	}
	// الإشارة اختفت ٣ أيام ← انحلت
	stale := []struct {
		ID     string `db:"id"`
		Signal string `db:"signal"`
		Div    string `db:"divisionId"`
		Unit   string `db:"unitId"`
		Title  string `db:"title"`
	}{}
	_ = s.db.Select(&stale, `SELECT id, signal, "divisionId", COALESCE("unitId", '') AS "unitId", title FROM "OrgIssueReport"
		WHERE source = 'MATRIX' AND status <> 'RESOLVED' AND signal IS NOT NULL AND "lastSeenAt" < $1`, time.Now().Add(-orgDiagnoseResolveAfter))
	for _, st := range stale {
		if seen[st.Signal] {
			continue
		}
		if _, err := s.db.Exec(`UPDATE "OrgIssueReport" SET status = 'RESOLVED', "resolvedAt" = now(),
			resolution = 'انحلت حسب أرقام النظام — الإشارة اختفت ٣ أيام (ماتركس)' WHERE id = $1`, st.ID); err != nil {
			log.Printf("[org-diagnose] تسكير %s: %v", st.ID, err)
			continue
		}
		closed++
		if notify != nil && st.Unit != "" {
			for _, h := range s.unitHeads(st.Div, st.Unit) {
				notify(h, "✅ ماتركس: المشكلة «"+st.Title+"» انحلت حسب الأرقام وتسكّرت.")
			}
		}
	}
	return opened, closed
}

// unitHeads مدير ونائب الوحدة، وإذا ماكو فمدير ونائب الإدارة.
func (s *OrgService) unitHeads(div, unit string) []string {
	ids := []string{}
	_ = s.db.Select(&ids, `SELECT DISTINCT "employeeId" FROM "OrgAssignment" WHERE "unitId" = $1 AND position IN ('DIRECTOR', 'DEPUTY') AND "endedAt" IS NULL`, unit)
	if len(ids) == 0 {
		_ = s.db.Select(&ids, `SELECT DISTINCT "employeeId" FROM "OrgAssignment" WHERE "divisionId" = $1 AND "unitId" IS NULL
			AND position IN ('DIRECTOR', 'DEPUTY') AND "endedAt" IS NULL`, div)
	}
	return ids
}
