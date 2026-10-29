package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// ═══ 👁️ متابعة موظف من «مذكراتي» (دازاي 10-29) ═══
// «اكتب اداء احد الموظفين ضعيف… كمال يقيس يشوفه صح ضعيف؟ بعد فتره من يتحسن يبلغ».
// المذكرة الي بيها اسم موظف ← نتحقق من أرقامه هسه ونحفظها خط أساس، وحلقة تقارن
// بعد أسبوع فما فوك: تحسّن واضح ← إشعار للي كتبها. إشارات من الأرقام، مو حكم.

const watchMaxDays = 45

// WatchCard نتيجة التحقق الي تطلع تحت المذكرة.
type WatchCard struct {
	ID       string `json:"id"`
	Employee string `json:"employee"`
	Summary  string `json:"summary"`
	Verdict  string `json:"verdict"`
}

func metricsLine(m EmpMetrics) string {
	return fmt.Sprintf("آخر ١٤ يوم: حضر %d يوم (%d متأخر)، منجز %d — والـ١٤ قبلها: حضر %d (%d متأخر)، منجز %d.",
		m.Att, m.Late, m.Done, m.AttPrev, m.LatePrev, m.DonePrev)
}

func weakSigns(m EmpMetrics) []string {
	var w []string
	if m.Late-m.LatePrev >= 2 || (m.Att > 0 && m.Late*3 >= m.Att) {
		w = append(w, "تأخير لافت")
	}
	if m.AttPrev-m.Att >= 3 && m.Leave < m.AttPrev-m.Att {
		w = append(w, "حضور نازل")
	}
	if m.DonePrev >= 3 && m.Done*10 <= m.DonePrev*6 {
		w = append(w, "إنجاز نازل")
	}
	if m.Streak >= 12 {
		w = append(w, "دوام متواصل بلا راحة")
	}
	return w
}

// OpenWatchFromText يدوّر على اسم موظف واحد بالنص، ويفتح متابعة بأرقامه.
// يرجّع nil إذا ماكو اسم واضح.
func OpenWatchFromText(db *sqlx.DB, ownerID, noteID, text string) (*WatchCard, error) {
	emps := []VoiceTaskEmployee{}
	if err := db.Select(&emps, `SELECT id AS "id", name AS "name" FROM "Employee"
		WHERE status = 'ACTIVE' AND role::text NOT IN ('OWNER', 'ADMIN')`); err != nil {
		return nil, err
	}
	ids := matchEmployeeBest(text, emps)
	if len(ids) != 1 {
		return nil, nil
	}
	return OpenWatch(db, ownerID, ids[0], noteID, text)
}

// OpenWatch يفتح (أو يحدّث) متابعة موظف ويرجّع بطاقة التحقق.
func OpenWatch(db *sqlx.DB, ownerID, employeeID, noteID, concern string) (*WatchCard, error) {
	rows, err := EmployeeMetrics(db, employeeID)
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	m := rows[0]
	base, _ := json.Marshal(m)
	var note *string
	if noteID != "" {
		note = &noteID
	}
	var id string
	// متابعة مفتوحة لنفس الموظف من نفس الشخص ← تتجدد بدل ما تتكرر
	err = db.Get(&id, `UPDATE "AiWatch" SET concern = $3, baseline = $4, "noteId" = COALESCE($5, "noteId"), "createdAt" = now(), "checkedAt" = NULL, status = 'OPEN'
		WHERE "ownerId" = $1 AND "employeeId" = $2 AND status IN ('OPEN', 'WORSE') RETURNING id`, ownerID, employeeID, concern, base, note)
	if err != nil {
		if err := db.Get(&id, `INSERT INTO "AiWatch" ("ownerId", "employeeId", "noteId", concern, baseline) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
			ownerID, employeeID, note, concern, base); err != nil {
			return nil, err
		}
	}
	card := &WatchCard{ID: id, Employee: m.Name, Summary: metricsLine(m)}
	if m.Att+m.AttPrev+m.Done+m.DonePrev == 0 {
		card.Verdict = "ماكو أرقام حضور أو إنجاز إله بآخر شهر — ما أگدر أتحقق بالأرقام. راح أتابعه وأبلغك من تبين أرقام."
	} else if w := weakSigns(m); len(w) > 0 {
		card.Verdict = "الأرقام تأيد ملاحظتك: " + strings.Join(w, "، ") + ". راح أتابعه وأبلغك إذا تحسّن."
	} else {
		card.Verdict = "أرقامه بالحضور والإنجاز مستقرة — ممكن السبب شي ما ينقاس (طريقة شغل، تعامل). راح أتابعه وأبلغك إذا تغيّر شي."
	}
	return card, nil
}

// SweepWatches يقارن المتابعات (٧ أيام فما فوك، وكل يوم مرة) بخط أساسها.
func SweepWatches(db *sqlx.DB, notify func(ownerID, msg string)) {
	ws := []struct {
		ID       string    `db:"id"`
		Owner    string    `db:"ownerId"`
		Emp      string    `db:"employeeId"`
		Status   string    `db:"status"`
		Baseline []byte    `db:"baseline"`
		Created  time.Time `db:"createdAt"`
	}{}
	if err := db.Select(&ws, `SELECT id, "ownerId", "employeeId", status, baseline, "createdAt" FROM "AiWatch"
		WHERE status IN ('OPEN', 'WORSE') AND "createdAt" < now() - interval '7 days'
		  AND ("checkedAt" IS NULL OR "checkedAt" < now() - interval '1 day') LIMIT 100`); err != nil {
		return
	}
	for _, w := range ws {
		var b EmpMetrics
		_ = json.Unmarshal(w.Baseline, &b)
		rows, err := EmployeeMetrics(db, w.Emp)
		if err != nil || len(rows) == 0 {
			_, _ = db.Exec(`UPDATE "AiWatch" SET status = 'CLOSED', "closedAt" = now() WHERE id = $1`, w.ID)
			continue
		}
		c := rows[0]
		since := w.Created.In(debriefLoc).Format("2006-01-02")
		better, worse := []string{}, []string{}
		if b.Late-c.Late >= 2 {
			better = append(better, fmt.Sprintf("التأخير نزل %d←%d", b.Late, c.Late))
		} else if c.Late-b.Late >= 2 {
			worse = append(worse, fmt.Sprintf("التأخير زاد %d←%d", b.Late, c.Late))
		}
		if c.Att-b.Att >= 3 {
			better = append(better, fmt.Sprintf("الحضور زاد %d←%d", b.Att, c.Att))
		} else if b.Att-c.Att >= 3 {
			worse = append(worse, fmt.Sprintf("الحضور نزل %d←%d", b.Att, c.Att))
		}
		if b.Done >= 1 && c.Done*10 >= b.Done*13 || b.Done == 0 && c.Done >= 3 {
			better = append(better, fmt.Sprintf("الإنجاز زاد %d←%d", b.Done, c.Done))
		} else if b.Done >= 3 && c.Done*10 <= b.Done*6 {
			worse = append(worse, fmt.Sprintf("الإنجاز نزل %d←%d", b.Done, c.Done))
		}
		switch {
		case len(better) > 0 && len(worse) == 0:
			notify(w.Owner, fmt.Sprintf("📈 %s تحسّن من ملاحظتك يوم %s: %s.", c.Name, since, strings.Join(better, "، ")))
			_, _ = db.Exec(`UPDATE "AiWatch" SET status = 'IMPROVED', "checkedAt" = now(), "closedAt" = now() WHERE id = $1`, w.ID)
		case len(worse) > 0 && w.Status != "WORSE":
			notify(w.Owner, fmt.Sprintf("📉 %s تراجع أكثر من ملاحظتك يوم %s: %s. باقي أتابعه.", c.Name, since, strings.Join(worse, "، ")))
			_, _ = db.Exec(`UPDATE "AiWatch" SET status = 'WORSE', "checkedAt" = now() WHERE id = $1`, w.ID)
		case time.Since(w.Created) > watchMaxDays*24*time.Hour:
			notify(w.Owner, fmt.Sprintf("👁️ خلصت متابعة %s (%d يوم من ملاحظتك يوم %s) بلا تغيّر واضح بالأرقام. %s", c.Name, watchMaxDays, since, metricsLine(c)))
			_, _ = db.Exec(`UPDATE "AiWatch" SET status = 'CLOSED', "checkedAt" = now(), "closedAt" = now() WHERE id = $1`, w.ID)
		default:
			_, _ = db.Exec(`UPDATE "AiWatch" SET "checkedAt" = now() WHERE id = $1`, w.ID)
		}
	}
}

func registerWatchTools(a *MatrixAgent, tc *SecretaryTurnCtx) {
	if tc == nil {
		return
	}
	a.AddTool(agentTool{Name: "watch_employee",
		Desc: "تابع موظف (المالك يگول «تابع فلان»، «أداءه ضعيف»، «ملاحظتي على فلان…»): يتحقق من أرقامه هسه، ويراقبه، ويبلغ المالك إذا تحسّن أو تراجع. اكتب الملاحظة كما گالها.",
		Props: map[string]any{
			"who":     map[string]any{"type": "string"},
			"concern": map[string]any{"type": "string", "description": "الملاحظة (مثل: أداءه ضعيف بالتسليم)"},
		},
		Req: []string{"who", "concern"},
		Run: func(in map[string]any, mask *nameMask) (string, error) {
			id, _ := mask.Resolve(strArg(in, "who"))
			if id == "" {
				return "", fmt.Errorf("ما لگيت الموظف")
			}
			card, err := OpenWatch(a.db, tc.UserID, id, "", strings.TrimSpace(strArg(in, "concern")))
			if err != nil || card == nil {
				return "", fmt.Errorf("تعذر فتح المتابعة")
			}
			tc.Cards = append(tc.Cards, SecretaryCard{Kind: "employee", Title: "👁️ متابعة " + card.Employee, Lines: []string{card.Summary, card.Verdict}})
			return card.Summary + "\n" + card.Verdict, nil
		}})
	a.AddTool(agentTool{Name: "list_watches",
		Desc: "متابعات الموظفين المفتوحة وحالها.",
		Run: func(map[string]any, *nameMask) (string, error) {
			rows := []struct {
				Name    string    `db:"name"`
				Concern string    `db:"concern"`
				Status  string    `db:"status"`
				Created time.Time `db:"createdAt"`
			}{}
			if err := a.db.Select(&rows, `SELECT e.name, w.concern, w.status, w."createdAt" FROM "AiWatch" w JOIN "Employee" e ON e.id = w."employeeId"
				WHERE w."ownerId" = $1 AND w."createdAt" > now() - interval '60 days' ORDER BY w."createdAt" DESC LIMIT 30`, tc.UserID); err != nil {
				return "", err
			}
			if len(rows) == 0 {
				return "ماكو متابعات.", nil
			}
			label := map[string]string{"OPEN": "أتابعه", "WORSE": "تراجع — أتابعه", "IMPROVED": "تحسّن ✓", "CLOSED": "انسدت"}
			var b strings.Builder
			for _, r := range rows {
				fmt.Fprintf(&b, "- %s (%s، من %s): %s\n", r.Name, label[r.Status], r.Created.In(debriefLoc).Format("2006-01-02"), r.Concern)
			}
			return b.String(), nil
		}})
}
