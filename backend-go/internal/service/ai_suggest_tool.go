package service

import (
	"fmt"
	"staffmange-api/internal/errlog"
	"strings"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// ═══ 🧠 «اقتراحات الذكاء لتطوير نفسه» (دازاي 10-28) ═══
// الوكيل يكتب اقتراح من يلاحظ نقص بنفسه: سؤال ما عنده أداة إله، سؤال يتكرر، خطأ
// صححه المالك. الي يحتاج برمجة يوصل إشعار للمالك حتى يطلبه من المبرمج.
// ما يكرر نفس العنوان خلال ٣٠ يوم، وحده ٥ باليوم.

const aiSuggestDailyMax = 5

// RegisterSuggestTool يضيف أداة suggest_improvement للوكيل.
func (a *MatrixAgent) RegisterSuggestTool(db *sqlx.DB, agent string, notifyOwner func(msg string)) {
	a.AddTool(agentTool{
		Name: "suggest_improvement",
		Desc: "سجّل اقتراح لتطوير نفسك أو النظام — استعملها من تلاحظ نقص حقيقي: سؤال ما عندك أداة تجاوبه، سؤال يتكرر، معلومة ناقصة، أو خطأ صححه المالك. اكتب عنوان قصير وتفاصيل (شنو الناقص، وشنو تقترح، وليش يفيد). needs_code=true إذا يحتاج برمجة (أداة جديدة، صفحة، بيانات ما تنقرا). لا تستعملها لشي تكدر تسويه هسه.",
		Props: map[string]any{
			"title":      map[string]any{"type": "string", "description": "عنوان قصير (أقل من ٨٠ حرف)"},
			"details":    map[string]any{"type": "string", "description": "التفاصيل: الناقص، الاقتراح، الفايدة"},
			"needs_code": map[string]any{"type": "boolean", "description": "يحتاج برمجة؟"},
		},
		Req: []string{"title", "details", "needs_code"},
		Run: func(in map[string]any, names *nameMask) (string, error) {
			title := strings.TrimSpace(fmt.Sprint(in["title"]))
			details := strings.TrimSpace(fmt.Sprint(in["details"]))
			needs, _ := in["needs_code"].(bool)
			if names != nil {
				title, details = names.Show(title), names.Show(details)
			}
			if r := []rune(title); len(r) > 120 {
				title = string(r[:120])
			}
			if title == "" || title == "<nil>" {
				return "", fmt.Errorf("العنوان فارغ")
			}
			var today, dup int
			errlog.Err(db.Get(&today, `SELECT count(*) FROM "AiSuggestion" WHERE "createdAt" > now() - interval '24 hours'`))
			if today >= aiSuggestDailyMax {
				return "وصلت حد الاقتراحات اليوم — احتفظ بيه لباچر إذا بعده مهم.", nil
			}
			errlog.Err(db.Get(&dup, `SELECT count(*) FROM "AiSuggestion" WHERE lower(title) = lower($1) AND "createdAt" > now() - interval '30 days'`, title))
			if dup > 0 {
				return "هذا الاقتراح مسجّل قبل — ما يحتاج تكرره.", nil
			}
			if _, err := db.Exec(`INSERT INTO "AiSuggestion" (agent, title, details, "needsCode") VALUES ($1, $2, $3, $4)`, agent, title, details, needs); err != nil {
				return "", err
			}
			if needs && notifyOwner != nil {
				notifyOwner("🧠 اقتراح من " + model.AgentName(agent) + " يحتاج برمجة: «" + title + "» — التفاصيل بشاشة «اقتراحات الذكاء».")
			}
			return "انسجل الاقتراح. كمّل جوابك عادي.", nil
		},
	})
	a.registerVerifySuggestionTool(db, agent, notifyOwner)
}

// ═══ ✅ «الوكيل يتأكد بعد تم» (دازاي 11-06، بلاغ #10) ═══
// من المالك يضغط «تم» الاقتراح يصير VERIFYING. الوكيل صاحبه يشوفه بأداة verify_suggestion،
// يتأكد بأدوات القراءة إن الشي صدگ انضاف، ويسجّل: FOUND → DONE، NOT_FOUND → يرجع ACCEPTED
// ويوصل إشعار للمالك. آخر النتائج ترجع ويا القائمة حتى الوكيل يتعلم شنو ينفّذ فعلاً.

type pendingVerify struct {
	ID      string `db:"id"`
	Title   string `db:"title"`
	Details string `db:"details"`
}

type verifyOutcome struct {
	Title  string  `db:"title"`
	Result string  `db:"verifyResult"`
	Note   *string `db:"verifyNote"`
}

func (a *MatrixAgent) registerVerifySuggestionTool(db *sqlx.DB, agent string, notifyOwner func(msg string)) {
	a.AddTool(agentTool{
		Name: "verify_suggestion",
		Desc: "اقتراحاتك الي المالك گال عليها «تم» تنتظر تأكدك إنها صدگ انضافت. action=list يجيبها ويا آخر نتائج تأكدك (تعلّم منها: شنو ينفّذ فعلاً). بعدها تأكد بأدوات القراءة الي عندك (هل الأداة/الشاشة/البيانات موجودة وتشتغل؟) وسجّل action=record ويا id وfound (true/false) وnote (شنو شفت بالضبط). لا تگول found=true إلا إذا شفته بعينك بأداة.",
		Props: map[string]any{
			"action": map[string]any{"type": "string", "enum": []string{"list", "record"}},
			"id":     map[string]any{"type": "string", "description": "رقم الاقتراح (لـ record)"},
			"found":  map[string]any{"type": "boolean", "description": "لگيته منضاف ويشتغل؟ (لـ record)"},
			"note":   map[string]any{"type": "string", "description": "شنو شفت بالضبط (لـ record)"},
		},
		Req: []string{"action"},
		Run: func(in map[string]any, names *nameMask) (string, error) {
			if fmt.Sprint(in["action"]) == "record" {
				return recordSuggestionVerify(db, agent, in, names, notifyOwner)
			}
			return listSuggestionVerify(db, agent)
		},
	})
}

func listSuggestionVerify(db *sqlx.DB, agent string) (string, error) {
	var pend []pendingVerify
	if err := db.Select(&pend, `SELECT id, title, details FROM "AiSuggestion" WHERE agent = $1 AND status = 'VERIFYING' ORDER BY "decidedAt" LIMIT 10`, agent); err != nil {
		return "", err
	}
	var past []verifyOutcome
	errlog.Err(db.Select(&past, `SELECT title, "verifyResult", "verifyNote" FROM "AiSuggestion"
		WHERE agent = $1 AND "verifyResult" IS NOT NULL AND status <> 'VERIFYING' ORDER BY "verifiedAt" DESC LIMIT 5`, agent))
	var b strings.Builder
	if len(pend) == 0 {
		b.WriteString("ماكو اقتراحات تنتظر تأكدك.\n")
	} else {
		b.WriteString("تنتظر تأكدك:\n")
		for _, p := range pend {
			fmt.Fprintf(&b, "- id=%s · %s — %s\n", p.ID, p.Title, p.Details)
		}
	}
	if len(past) > 0 {
		b.WriteString("آخر نتائج تأكدك (تعلّم منها):\n")
		for _, o := range past {
			note := ""
			if o.Note != nil {
				note = " — " + *o.Note
			}
			fmt.Fprintf(&b, "- %s: %s%s\n", o.Title, o.Result, note)
		}
	}
	return b.String(), nil
}

func recordSuggestionVerify(db *sqlx.DB, agent string, in map[string]any, names *nameMask, notifyOwner func(string)) (string, error) {
	id := strings.TrimSpace(fmt.Sprint(in["id"]))
	found, ok := in["found"].(bool)
	if id == "" || id == "<nil>" || !ok {
		return "", fmt.Errorf("لازم id وfound")
	}
	note := strings.TrimSpace(fmt.Sprint(in["note"]))
	if note == "<nil>" {
		note = ""
	}
	if names != nil {
		note = names.Show(note)
	}
	if r := []rune(note); len(r) > 500 {
		note = string(r[:500])
	}
	result, status := "NOT_FOUND", "ACCEPTED"
	if found {
		result, status = "FOUND", "DONE"
	}
	var title string
	err := db.Get(&title, `UPDATE "AiSuggestion" SET status = $3, "verifyResult" = $4, "verifyNote" = $5, "verifiedAt" = now()
		WHERE id = $1 AND agent = $2 AND status = 'VERIFYING' RETURNING title`, id, agent, status, result, note)
	if err != nil {
		return "هذا الاقتراح مو من اقتراحاتك الي تنتظر التأكد.", nil
	}
	if !found && notifyOwner != nil {
		notifyOwner("⚠️ " + model.AgentName(agent) + " ما لگى «" + title + "» منضاف رغم إنه انگال تم: " + note + " — رجع لحالة «مقبول» بشاشة «اقتراحات الذكاء».")
	}
	return "انسجلت النتيجة.", nil
}

// VerifySuggestionForTest يشغّل أداة verify_suggestion بلا وكيل كامل (للاختبارات).
func VerifySuggestionForTest(db *sqlx.DB, agent string, notifyOwner func(string)) func(map[string]any) (string, error) {
	a := &MatrixAgent{}
	a.registerVerifySuggestionTool(db, agent, notifyOwner)
	t := a.tools[len(a.tools)-1]
	return func(in map[string]any) (string, error) { return t.Run(in, nil) }
}
