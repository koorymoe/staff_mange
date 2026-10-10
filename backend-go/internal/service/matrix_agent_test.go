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
