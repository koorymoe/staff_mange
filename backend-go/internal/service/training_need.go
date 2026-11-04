package service

import (
	"fmt"
	"staffmange-api/internal/errlog"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// ═══ 🎓 مشاكل الموظفين بدراسة حقيقية (دازاي 11-02 — تعديلات ٩.٥) ═══
// «الموظف الجديد يحتاج تدريب» — ماتركس ما يصدّق ولا يكذّب: يفتح ملف الموظف
// المذكور (مهاراته، تدريبه، منحناه، أضعف شي عنده، سرعته)، يطابق ويا الكلام،
// ويحسم: يحتاج تدريب لو لا، وبأي شي. وبعد ٣٠ يوم يرجع يقيس: تحسّن لو لا.

type TrainingNeed struct {
	ID          string         `db:"id" json:"id"`
	IssueID     *string        `db:"issueId" json:"issueId"`
	EmployeeID  string         `db:"employeeId" json:"employeeId"`
	Verdict     string         `db:"verdict" json:"verdict"` // NEEDS | NO_NEED | UNKNOWN
	Areas       []string       `db:"-" json:"areas"`
	AreasDB     pq.StringArray `db:"areas" json:"-"`
	Rules       pq.StringArray `db:"rules" json:"-"`
	Evidence    string         `db:"evidence" json:"evidence"`
	BaselinePct *float64       `db:"baselinePct" json:"baselinePct"`
	FollowUpAt  time.Time      `db:"followUpAt" json:"followUpAt"`
	Status      string         `db:"status" json:"status"` // OPEN | IMPROVED | NOT_IMPROVED | CLOSED
	CurrentPct  *float64       `db:"currentPct" json:"currentPct"`
	ResultNote  *string        `db:"resultNote" json:"resultNote"`
	CreatedAt   time.Time      `db:"createdAt" json:"createdAt"`
}

// StudyTraining يدرس الموظف من ملفه ويحسم الاحتياج. ما يخزّن.
func StudyTraining(db *sqlx.DB, empID string) (*TrainingNeed, error) {
	d, err := BuildDossier(db, empID)
	if err != nil {
		return nil, err
	}
	n := &TrainingNeed{EmployeeID: empID, Areas: []string{}, Rules: pq.StringArray{}}
	ev := []string{}
	sum, cnt := 0.0, 0
	for _, w := range d.Weak {
		if !trainableRule(w.Rule) {
			continue // انضباط مو مهارة — التدريب ما يصلّحه
		}
		n.Areas = append(n.Areas, w.Title)
		n.Rules = append(n.Rules, w.Rule)
		ev = append(ev, fmt.Sprintf("%s %.0f٪ على %d مرة", w.Title, w.Pct, w.N))
		sum += w.Pct
		cnt++
	}
	if d.Trend == "DOWN" {
		ev = append(ev, d.TrendNote)
	}
	slow := d.SpeedRatio != nil && *d.SpeedRatio >= 1.4
	if slow {
		n.Areas = append(n.Areas, "سرعة التنفيذ")
		ev = append(ev, fmt.Sprintf("أبطأ من المعتاد %.1f مرة على %d حجز", *d.SpeedRatio, d.SpeedN))
	}
	if len(d.Training) > 0 {
		ev = append(ev, "تدريب سابق: "+strings.Join(d.Training, "، "))
	}
	if len(d.Strong) > 0 {
		s := []string{}
		for _, x := range d.Strong {
			s = append(s, x.Title)
		}
		ev = append(ev, "قوي بـ: "+strings.Join(s, "، "))
	}
	switch {
	case cnt > 0 || slow || d.Trend == "DOWN":
		n.Verdict = "NEEDS"
		if len(n.Areas) == 0 {
			n.Areas = append(n.Areas, "الأداء العام (منحناه نازل)")
		}
	case len(d.Months) > 0 && d.Months[len(d.Months)-1].MatrixPct != nil:
		n.Verdict = "NO_NEED"
		ev = append(ev, "ماكو ضعف واضح بالأرقام — الكلام ما يطابق بياناته، يمكن المشكلة بالتعامل مو بالمهارة")
	default:
		n.Verdict = "UNKNOWN"
		ev = append(ev, "ماكو بيانات كافية عنه بعد — يتابَع شهر ويرجع ينحسم")
	}
	if cnt > 0 {
		b := sum / float64(cnt)
		n.BaselinePct = &b
	}
	n.Evidence = strings.Join(ev, " · ")
	n.AreasDB = pq.StringArray(n.Areas)
	return n, nil
}

// SaveTrainingNeed يخزّن الدراسة على المشكلة (واحدة لكل موظف بالمشكلة) — المتابعة بعد ٣٠ يوم.
func SaveTrainingNeed(db *sqlx.DB, issueID string, n *TrainingNeed) error {
	// الي ما يحتاج تدريب ما يتابَع (CLOSED)؛ ولو صار يحتاج بإعادة التحليل يرجع OPEN.
	return db.QueryRow(`INSERT INTO "TrainingNeed" ("issueId", "employeeId", verdict, areas, rules, evidence, "baselinePct", "followUpAt", status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, now() + interval '30 days', CASE WHEN $3 = 'NO_NEED' THEN 'CLOSED' ELSE 'OPEN' END)
		ON CONFLICT ("issueId", "employeeId") DO UPDATE SET verdict = EXCLUDED.verdict, areas = EXCLUDED.areas, rules = EXCLUDED.rules,
		  evidence = EXCLUDED.evidence, "baselinePct" = COALESCE("TrainingNeed"."baselinePct", EXCLUDED."baselinePct"),
		  status = CASE WHEN EXCLUDED.verdict = 'NO_NEED' THEN 'CLOSED' WHEN "TrainingNeed".status = 'CLOSED' THEN 'OPEN' ELSE "TrainingNeed".status END
		RETURNING id, "followUpAt", status`, issueID, n.EmployeeID, n.Verdict, n.AreasDB, n.Rules, n.Evidence, n.BaselinePct).Scan(&n.ID, &n.FollowUpAt, &n.Status)
}

func IssueTrainingNeeds(db *sqlx.DB, issueID string) []TrainingNeed {
	rows := []TrainingNeed{}
	_ = db.Select(&rows, `SELECT id, "issueId", "employeeId", verdict, areas, rules, evidence, "baselinePct", "followUpAt", status, "currentPct", "resultNote", "createdAt"
		FROM "TrainingNeed" WHERE "issueId" = $1 ORDER BY "createdAt"`, issueID)
	for i := range rows {
		rows[i].Areas = []string(rows[i].AreasDB)
	}
	return rows
}

// FollowUpTraining يقيس الي حل موعد متابعتهم: نفس القواعد الضعيفة آخر ٣٠ يوم مقابل الأساس.
// يرجّع الي ما تحسّنوا (للإشعار).
func FollowUpTraining(db *sqlx.DB) (improved, notImproved int, err error) {
	due := []TrainingNeed{}
	if err = db.Select(&due, `SELECT id, "issueId", "employeeId", verdict, areas, rules, evidence, "baselinePct", "followUpAt", status, "currentPct", "resultNote", "createdAt"
		FROM "TrainingNeed" WHERE status = 'OPEN' AND "followUpAt" <= now()`); err != nil {
		return
	}
	for _, n := range due {
		var cur *float64
		if len(n.Rules) > 0 {
			_ = db.Get(&cur, `SELECT round(100.0 * sum(points) / NULLIF(sum("maxPoints"), 0), 1)::float8 FROM "MatrixScore"
				WHERE "employeeId" = $1 AND rule = ANY($2) AND "cancelledAt" IS NULL AND "at" > now() - interval '30 days'`, n.EmployeeID, n.Rules)
		} else {
			_ = db.Get(&cur, `SELECT round(100.0 * sum(points) / NULLIF(sum("maxPoints"), 0), 1)::float8 FROM "MatrixScore"
				WHERE "employeeId" = $1 AND rule NOT IN ('ATTENDANCE','INVENTORY') AND "cancelledAt" IS NULL AND "at" > now() - interval '30 days'`, n.EmployeeID)
		}
		status, note := "", ""
		switch {
		case cur == nil:
			// ماكو شغل ينقاس — نأجّل ١٥ يوم
			_, err = db.Exec(`UPDATE "TrainingNeed" SET "followUpAt" = now() + interval '15 days' WHERE id = $1`, n.ID)
			if err != nil {
				return
			}
			continue
		case *cur >= 80 || (n.BaselinePct != nil && *cur >= *n.BaselinePct+10):
			status, note = "IMPROVED", fmt.Sprintf("تحسّن: %.0f٪ هسه", *cur)
			improved++
		default:
			status, note = "NOT_IMPROVED", fmt.Sprintf("ما تحسّن: %.0f٪ هسه", *cur)
			notImproved++
		}
		if n.BaselinePct != nil {
			note += fmt.Sprintf(" (كان %.0f٪)", *n.BaselinePct)
		}
		if _, err = db.Exec(`UPDATE "TrainingNeed" SET status = $2, "currentPct" = $3, "resultNote" = $4 WHERE id = $1`, n.ID, status, cur, note); err != nil {
			return
		}
		if n.IssueID != nil {
			errlog.Exec(db.Exec(`INSERT INTO "WorkplaceIssueEntry" ("issueId", kind, text) VALUES ($1, 'MATRIX', $2)`, *n.IssueID, "🎓 متابعة ٣٠ يوم: "+note))
		}
	}
	return
}

func trainingVerdictAr(v string) string {
	return map[string]string{"NEEDS": "يحتاج تدريب", "NO_NEED": "ما يحتاج تدريب", "UNKNOWN": "ما ينحسم بعد"}[v]
}

// TrainingNeedText للتحليل (بلا أسماء — «الطرف أ/ب»).
func TrainingNeedText(label string, n *TrainingNeed) string {
	s := fmt.Sprintf("دراسة ملف %s: %s", label, trainingVerdictAr(n.Verdict))
	if n.Verdict == "NEEDS" {
		s += " بـ: " + strings.Join(n.Areas, "، ")
	}
	return s + ". الأدلة: " + n.Evidence
}

// trainableRule (٩.٦) قواعد الانضباط (التقرير، واجب التقييم، الحضور) مو نقص مهارة.
func trainableRule(r string) bool {
	switch r {
	case "DAILY_REPORT", "REPORT_MATCH", "RATING_DUTY", "ATTENDANCE", "INVENTORY":
		return false
	}
	return true
}
