package service

import (
	"context"
	"strings"
)

// SecretaryAgentReplier عقل مؤقت فوق NewSecretaryAgent — لحد ما يوصل
// SecretaryBrain مال لوفي (نفس الواجهة، فالتبديل سطر بـmain.go).
type SecretaryAgentReplier struct{ Deps SecretaryDeps }

const secretaryInterimSystem = `انت «السكرتير» بشركة GPS — تحچي ويا المالك أو مدير النظام بالعراقي.
جاوب بجملتين بالأكثر، والتفاصيل تطلع ببطاقات الأدوات. لا تخمّن أي رقم: كل رقم من أداة.
إذا طلب مهمة لموظف استعمل propose_task — هي مسودة بس والمستخدم يرسلها بنفسه.
ممنوع: الفلوس (أي تعديل)، الحذف، تغيير المناصب، الغرامات — كول بوضوح إنها مو بيدك.`

func (r *SecretaryAgentReplier) Reply(ctx context.Context, t SecretaryTurn) (*SecretaryResult, error) {
	tc := &SecretaryTurnCtx{UserID: t.UserID, Request: t.Text}
	a := NewSecretaryAgent(r.Deps, tc)
	sys := secretaryInterimSystem
	if t.UserName != "" {
		sys += "\nالمستخدم: " + t.UserName
	}
	if len(t.Prefs) > 0 {
		sys += "\nتفضيلاته: " + strings.Join(t.Prefs, " · ")
	}
	hist := make([]AgentTurn, 0, len(t.History))
	for _, h := range t.History {
		role := "USER"
		if h.Role == "assistant" {
			role = "ASSISTANT"
		}
		hist = append(hist, AgentTurn{Role: role, Text: h.Text})
	}
	run, err := a.ChatWith("SECRETARY", sys, hist, t.Text, nil)
	if err != nil {
		return nil, err
	}
	return &SecretaryResult{Reply: run.Answer, Cards: tc.Cards, Proposal: tc.Proposal}, nil
}
