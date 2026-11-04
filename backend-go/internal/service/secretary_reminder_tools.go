package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// ═══ ⏰ منبّه كمال (دازاي 10-29: «سويلي منبه بعد دقيقة… ماريد يكلي ماكدر») ═══
// set_reminder يحفظ التذكير، وحلقة كل ٣٠ ثانية (SendDueReminders) ترسله إشعار
// بوقته. الإشعار يوصل لجرس النظام (يتحدث كل دقيقة).

func registerReminderTools(a *MatrixAgent, tc *SecretaryTurnCtx) {
	if tc == nil {
		return
	}
	a.AddTool(agentTool{Name: "set_reminder",
		Desc: "منبّه/تذكير إله بوقت معيّن: «ذكّرني بعد دقيقة…»، «نبهني باچر ٩ الصبح…». يوصله إشعار بالنظام بالوقت. استعمل in_minutes للمدة النسبية، أو at بتوقيت بغداد YYYY-MM-DDTHH:MM للوقت المحدد.",
		Props: map[string]any{
			"text":       map[string]any{"type": "string", "description": "شنو يذكّره (مثل: لازم تكوم تدخن)"},
			"in_minutes": map[string]any{"type": "integer", "description": "بعد كم دقيقة (اختياري)"},
			"at":         map[string]any{"type": "string", "description": "وقت محدد بتوقيت بغداد YYYY-MM-DDTHH:MM (اختياري)"},
		},
		Req: []string{"text"},
		Run: func(in map[string]any, _ *nameMask) (string, error) {
			text := strings.TrimSpace(strArg(in, "text"))
			if text == "" {
				return "", fmt.Errorf("شنو أذكّره؟")
			}
			var due time.Time
			if m, ok := in["in_minutes"].(float64); ok && m > 0 {
				due = time.Now().Add(time.Duration(m * float64(time.Minute)))
			} else if at := strings.TrimSpace(strArg(in, "at")); at != "" {
				t, err := time.ParseInLocation("2006-01-02T15:04", at, debriefLoc)
				if err != nil {
					return "", fmt.Errorf("الوقت لازم YYYY-MM-DDTHH:MM")
				}
				due = t
			} else {
				return "", fmt.Errorf("حدد متى: in_minutes أو at")
			}
			if due.Before(time.Now().Add(-time.Minute)) {
				return "", fmt.Errorf("الوقت فات — اختار وقت جاي")
			}
			var n int
			_ = a.db.Get(&n, `SELECT count(*) FROM "AiReminder" WHERE "employeeId" = $1 AND "sentAt" IS NULL AND "cancelledAt" IS NULL`, tc.UserID)
			if n >= 50 {
				return "عنده ٥٠ تذكير ينتظر — خلّيه يلغي شي قبل.", nil
			}
			if _, err := a.db.Exec(`INSERT INTO "AiReminder" ("employeeId", text, "dueAt") VALUES ($1, $2, $3)`, tc.UserID, text, due); err != nil {
				return "", err
			}
			return fmt.Sprintf("انحفظ التذكير — يوصله إشعار %s (بغداد). گوله بجملة قصيرة.", due.In(debriefLoc).Format("2006-01-02 15:04")), nil
		}})

	a.AddTool(agentTool{Name: "list_reminders",
		Desc: "تذكيراته الي بعدها تنتظر.",
		Run: func(map[string]any, *nameMask) (string, error) {
			rows := []struct {
				Text string    `db:"text"`
				Due  time.Time `db:"dueAt"`
			}{}
			if err := a.db.Select(&rows, `SELECT text, "dueAt" FROM "AiReminder" WHERE "employeeId" = $1 AND "sentAt" IS NULL AND "cancelledAt" IS NULL ORDER BY "dueAt" LIMIT 30`, tc.UserID); err != nil {
				return "", err
			}
			if len(rows) == 0 {
				return "ماكو تذكيرات تنتظر.", nil
			}
			var b strings.Builder
			for _, r := range rows {
				fmt.Fprintf(&b, "- %s — %s\n", r.Due.In(debriefLoc).Format("2006-01-02 15:04"), r.Text)
			}
			return b.String(), nil
		}})

	a.AddTool(agentTool{Name: "cancel_reminder",
		Desc:  "الغي تذكير ينتظر (اكتب جزء من نصه).",
		Props: map[string]any{"text": map[string]any{"type": "string"}},
		Req:   []string{"text"},
		Run: func(in map[string]any, _ *nameMask) (string, error) {
			res, err := a.db.Exec(`UPDATE "AiReminder" SET "cancelledAt" = now() WHERE "employeeId" = $1 AND "sentAt" IS NULL AND "cancelledAt" IS NULL AND text ILIKE '%' || $2 || '%'`,
				tc.UserID, strings.TrimSpace(strArg(in, "text")))
			if err != nil {
				return "", err
			}
			k, _ := res.RowsAffected()
			return fmt.Sprintf("انلغى %d تذكير.", k), nil
		}})
}

// SendDueReminders يرسل التذكيرات الي حان وقتها كإشعار، ويرجّع عددها.
func SendDueReminders(db *sqlx.DB, notify func(employeeID, msg string)) int {
	rows := []struct {
		ID   string `db:"id"`
		Emp  string `db:"employeeId"`
		Text string `db:"text"`
	}{}
	if err := db.Select(&rows, `UPDATE "AiReminder" SET "sentAt" = now()
		WHERE id IN (SELECT id FROM "AiReminder" WHERE "sentAt" IS NULL AND "cancelledAt" IS NULL AND "dueAt" <= now() ORDER BY "dueAt" LIMIT 100 FOR UPDATE SKIP LOCKED)
		RETURNING id, "employeeId", text`); err != nil {
		return 0
	}
	for _, r := range rows {
		notify(r.Emp, "⏰ تذكير: "+r.Text)
	}
	return len(rows)
}
