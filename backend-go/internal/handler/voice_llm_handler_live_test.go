package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/repository"
	"staffmange-api/internal/service"
	"staffmange-api/internal/whatsapp"
)

type voiceTestBrain struct{ voice bool }

func (b *voiceTestBrain) Reply(_ context.Context, t service.WaTurn) (string, error) {
	b.voice = t.Voice
	return "تمام 🌹 موعدك باچر. نثبّته؟", nil
}

// نقطة منصة الصوت (دازاي 11-15): السر إجباري، والرد بصيغة OpenAI (SSE)، وسد المكالمة بأداة end_call.
func TestVoiceLLMEndpoint_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()
	cleanup := func() {
		db.Exec(`DELETE FROM "WaConversation" WHERE phone LIKE 'voice-test:vh-%' OR phone = '9647709997722'`)
	}
	cleanup()
	defer cleanup()
	svc := service.NewWaAgentService(db, repository.NewWaRepository(db), whatsapp.New("", "", "", ""), nil,
		repository.NewQualityFollowUpRepository(db), repository.NewMatrixProposalRepository(db),
		repository.NewBookingDeleteRequestRepository(db), repository.NewSystemSwitchRepository(db), nil, nil)
	brain := &voiceTestBrain{}
	svc.SetBrain(service.WaAgentReception, brain)
	post := func(h *VoiceLLMHandler, auth, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodPost, "/api/voice/v1/chat/completions", strings.NewReader(body))
		if auth != "" {
			r.Header.Set("Authorization", "Bearer "+auth)
		}
		w := httptest.NewRecorder()
		h.ChatCompletions(w, r)
		return w
	}
	body := `{"model":"m","stream":true,"messages":[
		{"role":"system","content":"انت مساعد.\nCALL id=vh-1 from={{system__caller_id}}"},
		{"role":"assistant","content":"هلا بيك"},
		{"role":"user","content":[{"type":"text","text":"اريد موعد"}]}]}`

	// بلا سر محطوط النقطة مسدودة؛ وبسر غلط مرفوضة.
	if w := post(NewVoiceLLMHandler(svc, ""), "x", body); w.Code != http.StatusNotFound {
		t.Fatalf("no secret configured: %d", w.Code)
	}
	h := NewVoiceLLMHandler(svc, "s3cret-long")
	if w := post(h, "wrong", body); w.Code != http.StatusUnauthorized {
		t.Fatalf("wrong secret: %d", w.Code)
	}
	if w := post(h, "", body); w.Code != http.StatusUnauthorized {
		t.Fatalf("missing secret: %d", w.Code)
	}

	// دور تجربة (الرقم ما انعبّى): SSE صحيح، والنص نظيف للنطق.
	w := post(h, "s3cret-long", body)
	if w.Code != 200 || !strings.HasPrefix(w.Header().Get("Content-Type"), "text/event-stream") {
		t.Fatalf("stream: %d %s", w.Code, w.Header().Get("Content-Type"))
	}
	lines := strings.Split(strings.TrimSpace(w.Body.String()), "\n\n")
	if lines[len(lines)-1] != "data: [DONE]" {
		t.Fatalf("missing [DONE]: %q", w.Body.String())
	}
	text, finish := "", ""
	for _, l := range lines[:len(lines)-1] {
		var c struct {
			Object  string `json:"object"`
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
				Finish *string `json:"finish_reason"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(strings.TrimPrefix(l, "data: ")), &c); err != nil || c.Object != "chat.completion.chunk" || len(c.Choices) != 1 {
			t.Fatalf("bad chunk %q: %v", l, err)
		}
		text += c.Choices[0].Delta.Content
		if c.Choices[0].Finish != nil {
			finish = *c.Choices[0].Finish
		}
	}
	if text != "تمام موعدك باچر. نثبّته؟" || finish != "stop" || !brain.voice {
		t.Fatalf("reply %q finish %q voice %v", text, finish, brain.voice)
	}

	// بلا كلام للزبون بعد: المنصة تطلب الترحيب.
	g := post(h, "s3cret-long", `{"messages":[{"role":"system","content":"CALL id=vh-2 from="}]}`)
	var full struct {
		Choices []struct {
			Message struct {
				Content   string `json:"content"`
				ToolCalls []struct {
					Function struct{ Name, Arguments string } `json:"function"`
				} `json:"tool_calls"`
			} `json:"message"`
			Finish string `json:"finish_reason"`
		} `json:"choices"`
	}
	if json.Unmarshal(g.Body.Bytes(), &full) != nil || len(full.Choices) != 1 || !strings.Contains(full.Choices[0].Message.Content, "مساعد ذكي") {
		t.Fatalf("greeting: %s", g.Body.String())
	}

	// رقم حقيقي والمفتاح مطفي: سد المكالمة بأداة end_call والوداع بيها.
	e := post(h, "s3cret-long", `{"tools":[{"type":"function","function":{"name":"end_call"}}],"messages":[
		{"role":"system","content":"CALL id=vh-3 from=+9647709997722"},{"role":"user","content":"هلو"}]}`)
	full.Choices = nil
	if json.Unmarshal(e.Body.Bytes(), &full) != nil || len(full.Choices) != 1 || full.Choices[0].Finish != "tool_calls" ||
		len(full.Choices[0].Message.ToolCalls) != 1 || full.Choices[0].Message.ToolCalls[0].Function.Name != "end_call" ||
		!strings.Contains(full.Choices[0].Message.ToolCalls[0].Function.Arguments, "مع السلامة") || full.Choices[0].Message.Content != "" {
		t.Fatalf("end_call: %s", e.Body.String())
	}
	// والمنصة الي ما عدها end_call تاخذ الوداع كنص.
	e2 := post(h, "s3cret-long", `{"messages":[{"role":"system","content":"CALL id=vh-4 from=9647709997722"},{"role":"user","content":"هلو"}]}`)
	full.Choices = nil
	if json.Unmarshal(e2.Body.Bytes(), &full) != nil || full.Choices[0].Finish != "stop" || !strings.Contains(full.Choices[0].Message.Content, "مع السلامة") {
		t.Fatalf("farewell text: %s", e2.Body.String())
	}
}
