package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
)

// نموذج وهمي بسكربت: كل نداء ياخذ الرد الجاي (نص أو أداة).
type waScript struct {
	steps []string // "text:…" أو "tool:name:{json}"
	calls []string
	i     int
}

func (s *waScript) New(_ context.Context, p anthropic.MessageNewParams) (*anthropic.Message, error) {
	raw, _ := json.Marshal(p)
	s.calls = append(s.calls, string(raw))
	step := "text:تمام"
	if s.i < len(s.steps) {
		step = s.steps[s.i]
	}
	s.i++
	var js string
	if strings.HasPrefix(step, "tool:") {
		parts := strings.SplitN(strings.TrimPrefix(step, "tool:"), ":", 2)
		js = `{"id":"m","type":"message","role":"assistant","model":"x","stop_reason":"tool_use","content":[{"type":"tool_use","id":"t` + itoa(s.i) + `","name":"` + parts[0] + `","input":` + parts[1] + `}],"usage":{"input_tokens":1,"output_tokens":1}}`
	} else {
		txt, _ := json.Marshal(strings.TrimPrefix(step, "text:"))
		js = `{"id":"m","type":"message","role":"assistant","model":"x","stop_reason":"end_turn","content":[{"type":"text","text":` + string(txt) + `}],"usage":{"input_tokens":1,"output_tokens":1}}`
	}
	var m anthropic.Message
	err := json.Unmarshal([]byte(js), &m)
	return &m, err
}

func itoa(n int) string { b, _ := json.Marshal(n); return string(b) }

// أدوات ناروتو الوهمية: تسجّل شنو انادى.
type waFakeTools struct{ ran []string }

func (f *waFakeTools) set(tc *WaTurnCtx) []agentTool {
	mk := func(name string) agentTool {
		return agentTool{Name: name, Run: func(in map[string]any, _ *nameMask) (string, error) {
			f.ran = append(f.ran, name)
			if name == "handoff" {
				tc.Handoff, tc.HandoffReason = true, strArg(in, "reason")
			}
			return "تم", nil
		}}
	}
	return []agentTool{mk("get_followup"), mk("record_contact"), mk("create_booking"), mk("suggest_to_matrix"),
		mk("get_booking"), mk("add_booking_details"), mk("confirm_booking"), mk("handoff")}
}

func newTestQuality(s *waScript, f *waFakeTools) *QualityAgent {
	q := NewQualityAgent(nil, "x", f.set, 0)
	q.llm = s
	return q
}

func turn(text string, hist ...WaHistory) WaAgentTurn {
	return WaAgentTurn{ConversationID: "c1", CustomerName: "زبون تجريبي", AgentName: "ابو الجودة",
		Facts: "حجز تجريبي: منظومة شمسية، انجز أمس", History: hist, Text: text}
}

func TestQualityOpeningAndPolicyOrder(t *testing.T) {
	op := QualityOpening("ابو الجودة")
	for _, w := range []string{"مرحبا زبوننا الكريم", "ابو الجودة", "شركة الأماني", "مساعد", "خلص زين"} {
		if !strings.Contains(op, w) {
			t.Fatalf("opening missing %q: %s", w, op)
		}
	}
	s := &waScript{steps: []string{"text:الحمدلله! شكد تقيّم الشغل من ٥؟"}}
	f := &waFakeTools{}
	res := newTestQuality(s, f).Reply(context.Background(), turn("اي خلص زين", WaHistory{"agent", op}))
	if res.Handoff || !strings.Contains(res.Reply, "من ٥") {
		t.Fatalf("res = %+v", res)
	}
	sys := s.calls[0]
	i1, i2, i3 := strings.Index(sys, "تقييمك للشغل"), strings.Index(sys, "تقييمك للكادر"), strings.Index(sys, "اكو أي مشكلة")
	if i1 < 0 || i2 < i1 || i3 < i2 {
		t.Fatalf("policy order wrong")
	}
	if !strings.Contains(sys, "سؤال واحد بالمرة") || !strings.Contains(sys, "زبون تجريبي") {
		t.Fatalf("system missing one-question rule or customer name")
	}
	// القالب (رسالة الوكيل الأولى) ينشال حتى المحادثة تبدي بالزبون
	if strings.Contains(s.calls[0], `"role":"assistant","content":[{"text":"مرحبا`) {
		t.Fatalf("leading agent template should be dropped")
	}
}

func TestQualityNoBookingWithoutConfirmation(t *testing.T) {
	// الزبون وصف مشكلة بس ما أكد — النموذج يحاول يحجز، والحارس يمنع.
	s := &waScript{steps: []string{`tool:create_booking:{"kind":"MAINTENANCE","details":"العاكسة تطفي"}`, "text:نسويلك حجز صيانة للعاكسة؟ نثبّت؟"}}
	f := &waFakeTools{}
	res := newTestQuality(s, f).Reply(context.Background(), turn("العاكسة تطفي بالليل"))
	if len(f.ran) != 0 {
		t.Fatalf("booking ran without confirmation: %v", f.ran)
	}
	if !strings.Contains(s.calls[1], "بعده ما أكد") {
		t.Fatalf("model should be told the customer did not confirm")
	}
	if res.Handoff || len(res.Actions) != 0 {
		t.Fatalf("res = %+v", res)
	}
	// بعد «اي ثبتها» ينفذ.
	s2 := &waScript{steps: []string{`tool:create_booking:{"kind":"MAINTENANCE","details":"العاكسة تطفي"}`, "text:تم، سوينا الحجز"}}
	f2 := &waFakeTools{}
	res2 := newTestQuality(s2, f2).Reply(context.Background(), turn("اي ثبتها"))
	if len(f2.ran) != 1 || f2.ran[0] != "create_booking" || len(res2.Actions) != 1 {
		t.Fatalf("confirmed booking should run: ran=%v res=%+v", f2.ran, res2)
	}
	// «لا مو هسه» مو موافقة
	if waConfirmed("لا مو هسه") || waConfirmed("ما اريد") || !waConfirmed("تمام ثبّت") {
		t.Fatalf("confirmation detection wrong")
	}
}

func TestWaHandoffOnRequestWithoutModel(t *testing.T) {
	s := &waScript{}
	f := &waFakeTools{}
	for _, msg := range []string{"اريد احچي ويا موظف", "ممكن تحولني لشخص حقيقي", "انتو نصابين"} {
		res := newTestQuality(s, f).Reply(context.Background(), turn(msg))
		if !res.Handoff || res.Reply != waHandoffReply {
			t.Fatalf("%q should hand off: %+v", msg, res)
		}
	}
	if len(s.calls) != 0 {
		t.Fatalf("model must not be called on handoff")
	}
	// ذكر الموظف بالتقييم مو طلب تحويل
	if need, _ := waNeedsHuman("الموظف چان خوش والكادر محترم"); need {
		t.Fatalf("praising the crew is not a handoff")
	}
}

func TestWaHandoffToolAndRatingGuard(t *testing.T) {
	s := &waScript{steps: []string{`tool:handoff:{"reason":"سؤال عن الضمان"}`}}
	f := &waFakeTools{}
	res := newTestQuality(s, f).Reply(context.Background(), turn("شكد الضمان على البطاريات؟"))
	if !res.Handoff || res.HandoffReason != "سؤال عن الضمان" {
		t.Fatalf("res = %+v", res)
	}
	s2 := &waScript{steps: []string{`tool:record_contact:{"customerRating":7,"crewRating":5}`, "text:عفواً، التقييم من ١ إلى ٥ — شكد تنطينا؟"}}
	f2 := &waFakeTools{}
	newTestQuality(s2, f2).Reply(context.Background(), turn("عشرة على عشرة"))
	if len(f2.ran) != 0 || !strings.Contains(s2.calls[1], "من ١ إلى ٥") {
		t.Fatalf("out-of-range rating must be rejected: %v", f2.ran)
	}
}

func TestWaOffAndCapHandOff(t *testing.T) {
	off := NewBookingAgent(nil, "x", nil, 0)
	if r := off.Reply(context.Background(), turn("هلا")); !r.Handoff {
		t.Fatalf("off agent should hand off")
	}
	b := NewBookingAgent(nil, "x", (&waFakeTools{}).set, 1)
	b.llm = &waScript{steps: []string{`tool:get_booking:{}`, "text:هلا"}}
	if r := b.Reply(context.Background(), turn("هلا")); !r.Handoff || !strings.Contains(r.HandoffReason, "حده اليومي") {
		t.Fatalf("cap should hand off: %+v", r)
	}
}

func TestBookingOpeningAndConfirm(t *testing.T) {
	op := BookingOpening("ابو الحجوزات", "تنصيب كاميرات")
	if !strings.Contains(op, "تنصيب كاميرات") || !strings.Contains(op, "مساعد") {
		t.Fatalf("opening = %s", op)
	}
	s := &waScript{steps: []string{`tool:confirm_booking:{"at":"2026-10-25T10:00"}`, "text:نثبّته الخميس الساعة ١٠؟"}}
	f := &waFakeTools{}
	b := NewBookingAgent(nil, "x", f.set, 0)
	b.llm = s
	b.Reply(context.Background(), turn("الخميس يناسبني"))
	if len(f.ran) != 0 {
		t.Fatalf("confirm_booking ran without explicit yes")
	}
	if !strings.Contains(s.calls[0], "لا تذكر سعر") {
		t.Fatalf("booking policy missing price rule")
	}
}
