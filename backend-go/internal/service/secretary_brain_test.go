package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
)

// نموذج وهمي للسكرتير: يطلب find_booking، بعدها propose_task، بعدها يجاوب.
type secFakeLLM struct {
	calls []string
	step  int
}

func (f *secFakeLLM) New(_ context.Context, p anthropic.MessageNewParams) (*anthropic.Message, error) {
	raw, _ := json.Marshal(p)
	f.calls = append(f.calls, string(raw))
	f.step++
	var js string
	switch f.step {
	case 1:
		js = `{"id":"m1","type":"message","role":"assistant","model":"x","stop_reason":"tool_use","content":[{"type":"tool_use","id":"t1","name":"find_booking","input":{"q":"كرار جعفر"}}],"usage":{"input_tokens":1,"output_tokens":1}}`
	case 2:
		js = `{"id":"m2","type":"message","role":"assistant","model":"x","stop_reason":"tool_use","content":[{"type":"tool_use","id":"t2","name":"propose_task","input":{"who":"أحمد خليل"}}],"usage":{"input_tokens":1,"output_tokens":1}}`
	default:
		js = `{"id":"m3","type":"message","role":"assistant","model":"x","stop_reason":"end_turn","content":[{"type":"text","text":"حجز كرار جعفر قيد العمل، وحضّرت مهمة لأحمد خليل."}],"usage":{"input_tokens":1,"output_tokens":1}}`
	}
	var m anthropic.Message
	if err := json.Unmarshal([]byte(js), &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func secTestTools() []agentTool {
	card := `{"facts":"حجز 8850 لكرار جعفر قيد العمل","cards":[{"kind":"booking","title":"حجز 8850 — كرار جعفر","status":"قيد العمل","lines":["الليدر بالموقع"]},{"kind":"booking","title":"حجز 8850 — كرار جعفر","status":"مكرر"}]}`
	prop := `{"proposal":{"actionId":"a1","kind":"task","summary":"مهمة لأحمد خليل — متابعة الحجز 8850","fields":{"employeeId":"e7","employeeName":"أحمد خليل","priority":"URGENT"}}}`
	return []agentTool{
		{Name: "find_booking", Run: func(map[string]any, *nameMask) (string, error) { return card, nil }},
		{Name: "propose_task", Run: func(map[string]any, *nameMask) (string, error) { return prop, nil }},
	}
}

func TestSecretaryCollectsCardsAndProposalWithNames(t *testing.T) {
	f := &secFakeLLM{}
	b := &SecretaryBrain{llm: f, model: "x", tools: secTestTools()}
	res, err := b.Reply(context.Background(), SecretaryTurn{UserName: "دازاي", Text: "شوفلي حجز كرار جعفر وودّي مهمة لأحمد خليل",
		History: []SecretaryHistory{{Role: "assistant", Text: "تحية قديمة"}, {Role: "user", Text: "هلا"}, {Role: "assistant", Text: "هلا بيك"}},
		Prefs:   []string{"يحب الجواب مختصر"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Cards) != 1 || res.Cards[0].Status != "قيد العمل" {
		t.Fatalf("cards = %+v (لازم وحدة بلا تكرار)", res.Cards)
	}
	if res.Proposal == nil || res.Proposal.Fields.EmployeeID != "e7" {
		t.Fatalf("proposal = %+v", res.Proposal)
	}
	if !strings.Contains(res.Reply, "أحمد خليل") {
		t.Fatalf("reply = %q", res.Reply)
	}
	// الأسماء مفتوحة للسكرتير: الاسم يوصل النموذج مثل ما هو، بلا رموز.
	if !strings.Contains(f.calls[0], "كرار جعفر") || strings.Contains(f.calls[0], "موظف#") {
		t.Fatalf("first call should carry the real name unmasked")
	}
	// التفضيلات والتحية بالتعليمات، والتاريخ يبدي بالمستخدم.
	if !strings.Contains(f.calls[0], "يحب الجواب مختصر") || !strings.Contains(f.calls[0], "دازاي") {
		t.Fatalf("system missing prefs/name")
	}
	if strings.Contains(f.calls[0], "تحية قديمة") {
		t.Fatalf("leading assistant turn should be dropped")
	}
	// النموذج يشوف facts، مو JSON البطاقات.
	if !strings.Contains(f.calls[1], "حجز 8850 لكرار جعفر قيد العمل") || strings.Contains(f.calls[1], `\"cards\"`) {
		t.Fatalf("tool result should be the facts text")
	}
}

func TestSecretaryCapAndOff(t *testing.T) {
	b := &SecretaryBrain{llm: &secFakeLLM{}, model: "x", tools: secTestTools(), dailyCap: 2}
	if _, err := b.Reply(context.Background(), SecretaryTurn{Text: "شوفلي"}); !errors.Is(err, ErrSecretaryCap) {
		t.Fatalf("want cap error after 2 calls, got %v", err)
	}
	off := &SecretaryBrain{model: "x"}
	if _, err := off.Reply(context.Background(), SecretaryTurn{Text: "هلا"}); !errors.Is(err, ErrSecretaryOff) {
		t.Fatalf("want off, got %v", err)
	}
	if _, err := off.Summarize(context.Background(), "", "x"); !errors.Is(err, ErrSecretaryOff) {
		t.Fatalf("summarize want off, got %v", err)
	}
}

type secSummaryLLM struct{}

func (secSummaryLLM) New(_ context.Context, _ anthropic.MessageNewParams) (*anthropic.Message, error) {
	var m anthropic.Message
	err := json.Unmarshal([]byte(`{"id":"s","type":"message","role":"assistant","model":"x","stop_reason":"end_turn","content":[{"type":"text","text":"- ٢ حجز متأخر\n\n١. ٣ شكاوى مفتوحة\n• ٥ موظفين مداومين"}],"usage":{"input_tokens":1,"output_tokens":1}}`), &m)
	return &m, err
}

func TestSecretarySummarizeLines(t *testing.T) {
	b := &SecretaryBrain{llm: secSummaryLLM{}, model: "x"}
	lines, err := b.Summarize(context.Background(), "دازاي", "late=2 complaints=3 present=5")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"٢ حجز متأخر", "٣ شكاوى مفتوحة", "٥ موظفين مداومين"}
	if strings.Join(lines, "|") != strings.Join(want, "|") {
		t.Fatalf("lines = %q", lines)
	}
}
