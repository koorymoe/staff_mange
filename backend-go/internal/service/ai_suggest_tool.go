package service

import (
	"fmt"
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
			_ = db.Get(&today, `SELECT count(*) FROM "AiSuggestion" WHERE "createdAt" > now() - interval '24 hours'`)
			if today >= aiSuggestDailyMax {
				return "وصلت حد الاقتراحات اليوم — احتفظ بيه لباچر إذا بعده مهم.", nil
			}
			_ = db.Get(&dup, `SELECT count(*) FROM "AiSuggestion" WHERE lower(title) = lower($1) AND "createdAt" > now() - interval '30 days'`, title)
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
}
