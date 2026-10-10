package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
)

// نموذج وهمي: أول مرة يطلب أداة على «موظف#1»، وبعدها يجاوب.
type fakeLLM struct {
	calls []string
	step  int
}

func (f *fakeLLM) New(_ context.Context, p anthropic.MessageNewParams) (*anthropic.Message, error) {
	raw, _ := json.Marshal(p)
	f.calls = append(f.calls, string(raw))
	f.step++
	var js string
	if f.step == 1 {
		js = `{"id":"m1","type":"message","role":"assistant","model":"x","stop_reason":"tool_use","content":[{"type":"tool_use","id":"t1","name":"employee_profile","input":{"who":"موظف#1"}}],"usage":{"input_tokens":1,"output_tokens":1}}`
	} else {
		js = `{"id":"m2","type":"message","role":"assistant","model":"x","stop_reason":"end_turn","content":[{"type":"text","text":"موظف#1 متأخر يومين."}],"usage":{"input_tokens":1,"output_tokens":1}}`
	}
	var m anthropic.Message
	if err := json.Unmarshal([]byte(js), &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func TestAgentHidesNamesAndRunsTools(t *testing.T) {
	f := &fakeLLM{}
	a := &MatrixAgent{llm: f, model: "x", testNames: map[string]string{"علي حسين الليدر": "e1"}}
	got := ""
	a.AddTool(agentTool{Name: "employee_profile", Run: func(in map[string]any, m *nameMask) (string, error) {
		id, _ := m.Resolve(strArg(in, "who"))
		got = id
		return "علي حسين الليدر تأخر يومين بالحجوزات", nil
	}})
	run, err := a.Run("TEST", "sys", "شوفلي علي حسين الليدر")
	if err != nil {
		t.Fatal(err)
	}
	if got != "e1" {
		t.Fatalf("tool resolved %q, want e1", got)
	}
	for i, c := range f.calls {
		if strings.Contains(c, "علي حسين") {
			t.Fatalf("call %d leaked a real name: %s", i, c)
		}
	}
	if run.Answer != "علي حسين الليدر متأخر يومين." {
		t.Fatalf("answer = %q", run.Answer)
	}
	if len(run.Steps) != 1 || run.Steps[0] != "employee_profile" {
		t.Fatalf("steps = %v", run.Steps)
	}
}

func TestAgentOffWithoutModel(t *testing.T) {
	a := &MatrixAgent{}
	if _, err := a.Run("TEST", "s", "q"); err != ErrAgentOff {
		t.Fatalf("err = %v", err)
	}
}

func TestMaskShortNamesWholeWord(t *testing.T) {
	a := &MatrixAgent{testNames: map[string]string{"Hr": "h", "علي": "a"}}
	m := a.newMask()
	got := m.Hide("ما حضر: Hr، علي. وكلمة Hrx وعليوي تبقى")
	if strings.Contains(got, "ما حضر: Hr،") || strings.Contains(got, " علي.") || !strings.Contains(got, "Hrx") || !strings.Contains(got, "عليوي") {
		t.Fatalf("got %q", got)
	}
}

type textLLM struct {
	reply string
	seen  string
}

func (f *textLLM) New(_ context.Context, p anthropic.MessageNewParams) (*anthropic.Message, error) {
	raw, _ := json.Marshal(p)
	f.seen = string(raw)
	js, _ := json.Marshal(map[string]any{"id": "m", "type": "message", "role": "assistant", "model": "x", "stop_reason": "end_turn",
		"content": []map[string]any{{"type": "text", "text": f.reply}}, "usage": map[string]any{"input_tokens": 1, "output_tokens": 1}})
	var m anthropic.Message
	err := json.Unmarshal(js, &m)
	return &m, err
}

func TestVoiceRewriteKeepsFacts(t *testing.T) {
	orig := "🤖 ماتركس — علي حسين الليدر، خلّصت الحجوزات (B753، B740) وبعدك ما قيّمت 3 فنيين."
	good := &textLLM{reply: "هلا موظف#1، بعد B753 و B740 بقى تقييم 3 فنيين — تگدر تخلّصه هسه؟"}
	a := &MatrixAgent{llm: good, model: "x", testNames: map[string]string{"علي حسين الليدر": "e1"}}
	out := a.Rewrite(nil, orig)
	if !strings.Contains(out, "علي حسين الليدر") || !strings.Contains(out, "B753") || strings.Contains(good.seen, "علي حسين") {
		t.Fatalf("good rewrite: %q (seen leaked=%v)", out, strings.Contains(good.seen, "علي حسين"))
	}
	lost := &textLLM{reply: "هلا موظف#1، بعد B753 بقى تقييم فنيين."}
	a.llm = lost
	if got := a.Rewrite(nil, orig); got != orig {
		t.Fatalf("missing fact should fall back, got %q", got)
	}
	bad := &textLLM{reply: "موظف#1 إذا ما تقيّم B753 B740 الـ3 راح تاخذ غرامة"}
	a.llm = bad
	if got := a.Rewrite(nil, orig); got != orig {
		t.Fatalf("banned word should fall back, got %q", got)
	}
}
