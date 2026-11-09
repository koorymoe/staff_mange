package service

import (
	"context"
	"sync"

	"github.com/anthropics/anthropic-sdk-go"
)

// ═══ الربط: عقل لوفي (QualityAgent/BookingAgent بحراساته) ← أدوات ناروتو (WaBrain) ═══
// عقل لوفي ياخذ الأدوات بـtoolsFor ساعة البناء؛ أدوات ناروتو تنبني كل دور (مربوطة
// بالمحادثة). فالجسر يحط أدوات الدور الحالي قبل Reply — دور واحد بالمرة.

type luffyWaBrain struct {
	mu    sync.Mutex
	cur   []agentTool
	reply func(context.Context, WaAgentTurn) *WaAgentResult
}

// NewLuffyWaBrain عقل الوكيل. client nil = كل رسالة تتحول لإنسان (حراسة لوفي).
func NewLuffyWaBrain(agent string, client *anthropic.Client, model string, dailyCap int) WaBrain {
	b := &luffyWaBrain{}
	toolsFor := func(*WaTurnCtx) []agentTool { return b.cur }
	if agent == WaAgentQuality {
		b.reply = NewQualityAgent(client, model, toolsFor, dailyCap).Reply
	} else {
		b.reply = NewBookingAgent(client, model, toolsFor, dailyCap).Reply
	}
	return b
}

func (b *luffyWaBrain) Reply(ctx context.Context, t WaTurn) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.cur = t.Tools
	hist := make([]WaHistory, 0, len(t.History))
	for _, h := range t.History {
		role := "customer"
		if h.Role == "ASSISTANT" {
			role = "agent"
		}
		hist = append(hist, WaHistory{Role: role, Text: h.Text})
	}
	// ⚠️ بلا CustomerName — لحد قرار الخصوصية المكتوب.
	res := b.reply(ctx, WaAgentTurn{AgentName: t.AgentName, Facts: t.Context, History: hist, Text: t.Message})
	if res == nil {
		return "", nil
	}
	// تحويل قرره لوفي بلا أداة (طلب إنسان، عصبية، مطفي، خطأ) — نسجّله ونحوّل المحادثة.
	if res.Handoff && !contains(res.Actions, "handoff") {
		for _, tool := range t.Tools {
			if tool.Name == "handoff" {
				_, _ = tool.Run(map[string]any{"reason": res.HandoffReason}, nil)
			}
		}
	}
	return res.Reply, nil
}

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}
