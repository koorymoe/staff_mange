package service

import (
	"context"
	"testing"
)

// الجسر: تحويل قرره عقل لوفي بلا أداة لازم يشغّل أداة handoff مال ناروتو (حتى المحادثة تتحوّل وتنسجل).
func TestLuffyBridgeHandoff(t *testing.T) {
	called := ""
	b := &luffyWaBrain{reply: func(_ context.Context, turn WaAgentTurn) *WaAgentResult {
		if turn.CustomerName != "" {
			t.Fatal("اسم الزبون ما لازم يوصل للعقل")
		}
		return &WaAgentResult{Reply: waHandoffReply, Handoff: true, HandoffReason: "الزبون طلب يحچي ويا موظف"}
	}}
	tools := []agentTool{{Name: "handoff", Run: func(in map[string]any, _ *nameMask) (string, error) {
		called = strArg(in, "reason")
		return "ok", nil
	}}}
	got, err := b.Reply(context.Background(), WaTurn{Agent: WaAgentQuality, Message: "اريد احچي ويا موظف", Tools: tools,
		History: []AgentTurn{{Role: "ASSISTANT", Text: "مرحبا"}}})
	if err != nil || got != waHandoffReply || called != "الزبون طلب يحچي ويا موظف" {
		t.Fatalf("got %q err %v called %q", got, err, called)
	}
	// بلا مفتاح: عقل لوفي يحوّل، والجسر يسجّل التحويل
	called = ""
	nb := NewLuffyWaBrain(WaAgentBooking, nil, "", 10)
	if r, _ := nb.Reply(context.Background(), WaTurn{Agent: WaAgentBooking, Message: "هلا", Tools: tools}); r == "" || called == "" {
		t.Fatalf("no-key brain should hand off: %q %q", r, called)
	}
}
