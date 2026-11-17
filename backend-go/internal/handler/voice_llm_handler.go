package handler

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"staffmange-api/internal/service"
)

// ═══ 📞 عقل المكالمة الحية — نقطة «Custom LLM» لمنصة الصوت (دازاي 11-15) ═══
// POST /api/voice/v1/chat/completions بصيغة OpenAI Chat Completions (الي تدعمها منصات الصوت:
// ElevenLabs Agents وغيرها). المنصة تدز المحادثة كلها بكل دور؛ ناخذ آخر جملة للزبون ونرجّع رد
// الوكيل (SSE إذا stream). التفاصيل بـservice/wa_voice.go.
//
// بلا دخول موظف — الحماية بسر مشترك (Authorization: Bearer VOICE_LLM_SECRET). بلا سر محطوط = 404.
// المكالمة ورقم المتصل يجن بسطر بتعليمات الوكيل عند المنصة (هي تعبّي المتغيرات):
//
//	CALL id={{system__conversation_id}} from={{system__caller_id}}

type VoiceLLMHandler struct {
	svc    *service.WaAgentService
	secret string
	// TestPhone (دازاي 11-16) مكالمة بلا رقم (زر التجربة بلوحة المنصة) تنحسب من هالرقم: الحجز
	// ينكتب صدگ حتى المالك يشوفه بالنظام. فاضي = وضع تجربة، ولا شي ينكتب.
	TestPhone string
}

func NewVoiceLLMHandler(svc *service.WaAgentService, secret string) *VoiceLLMHandler {
	return &VoiceLLMHandler{svc: svc, secret: strings.TrimSpace(secret)}
}

var voiceCallRe = regexp.MustCompile(`CALL\s+id=(\S*)\s+from=(\S*)`)

type voiceMsg struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

// text محتوى الرسالة: نص، أو أجزاء [{type:text,text}].
func (m voiceMsg) text() string {
	var s string
	if json.Unmarshal(m.Content, &s) == nil {
		return s
	}
	var parts []struct {
		Text string `json:"text"`
	}
	if json.Unmarshal(m.Content, &parts) == nil {
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			out = append(out, p.Text)
		}
		return strings.Join(out, " ")
	}
	return ""
}

// voiceVar قيمة متغير المنصة — الي ما انعبّى («{{…}}» أو null) يعتبر فاضي.
func voiceVar(v string) string {
	v = strings.TrimSpace(v)
	if strings.Contains(v, "{{") || strings.EqualFold(v, "null") || strings.EqualFold(v, "none") || strings.EqualFold(v, "undefined") {
		return ""
	}
	return v
}

// POST /api/voice/v1/chat/completions
func (h *VoiceLLMHandler) ChatCompletions(w http.ResponseWriter, r *http.Request) {
	if h.secret == "" {
		http.NotFound(w, r)
		return
	}
	got := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if subtle.ConstantTimeCompare([]byte(strings.TrimSpace(got)), []byte(h.secret)) != 1 {
		http.Error(w, `{"error":{"message":"unauthorized"}}`, http.StatusUnauthorized)
		return
	}
	var in struct {
		Model    string     `json:"model"`
		Stream   bool       `json:"stream"`
		Messages []voiceMsg `json:"messages"`
		Tools    []struct {
			Function struct {
				Name string `json:"name"`
			} `json:"function"`
		} `json:"tools"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2<<20)).Decode(&in); err != nil {
		http.Error(w, `{"error":{"message":"bad request"}}`, http.StatusBadRequest)
		return
	}
	callID, phone, greeting, last, users := "", "", "", "", 0
	for _, m := range in.Messages {
		t := m.text()
		switch m.Role {
		case "system", "developer":
			if mm := voiceCallRe.FindStringSubmatch(t); mm != nil {
				callID, phone = voiceVar(mm[1]), voiceVar(mm[2])
			}
		case "assistant":
			if users == 0 && greeting == "" {
				greeting = t // ترحيب المنصة قبل أول كلام للزبون
			}
		case "user":
			users++
			last = t
		}
	}
	phone = voiceCaller(phone, h.TestPhone)
	canEnd := false
	for _, t := range in.Tools {
		if t.Function.Name == "end_call" {
			canEnd = true
		}
	}
	sw := newVoiceStream(w, in.Model, in.Stream)
	ctx := r.Context()
	if in.Stream {
		// (دازاي 11-17 «التسريع») جملة الانتظار تنبث فوراً لمن الوكيل يحتاج أداة.
		ctx = service.WithVoiceProgress(ctx, func(s string) { sw.text(s + " ") })
	}
	var reply service.VoiceReply
	if users == 0 {
		reply = service.VoiceReply{Text: h.svc.VoiceGreeting()} // المنصة تطلب أول جملة
	} else {
		reply = h.svc.VoiceTurn(ctx, callID, phone, users, last, greeting)
	}
	h.write(sw, reply, reply.End && canEnd)
}

// voiceStream يكتب رد OpenAI: SSE يبدي أول ما يجي أول نص (جملة الانتظار)، أو JSON واحد بلا بث.
type voiceStream struct {
	w       http.ResponseWriter
	fl      http.Flusher
	id      string
	now     int64
	model   string
	stream  bool
	started bool
}

func newVoiceStream(w http.ResponseWriter, model string, stream bool) *voiceStream {
	if model == "" {
		model = "amani-voice"
	}
	fl, _ := w.(http.Flusher)
	return &voiceStream{w: w, fl: fl, id: "chatcmpl-" + uuid.NewString(), now: time.Now().Unix(), model: model, stream: stream}
}

func (s *voiceStream) chunk(delta map[string]any, fin any) {
	if !s.started {
		s.started = true
		s.w.Header().Set("Content-Type", "text/event-stream")
		s.w.Header().Set("Cache-Control", "no-cache")
		s.w.Header().Set("X-Accel-Buffering", "no") // nginx/caddy: لا تجمّع الرد
		s.chunk(map[string]any{"role": "assistant", "content": ""}, nil)
	}
	raw, _ := json.Marshal(map[string]any{"id": s.id, "object": "chat.completion.chunk", "created": s.now, "model": s.model,
		"choices": []map[string]any{{"index": 0, "delta": delta, "finish_reason": fin}}})
	fmt.Fprintf(s.w, "data: %s\n\n", raw)
	if s.fl != nil {
		s.fl.Flush()
	}
}

func (s *voiceStream) text(t string) { s.chunk(map[string]any{"content": t}, nil) }

func (h *VoiceLLMHandler) write(sw *voiceStream, reply service.VoiceReply, endCall bool) {
	w, model, stream, id, now := sw.w, sw.model, sw.stream, sw.id, sw.now
	// سد المكالمة: أداة end_call مال المنصة، والوداع بيها (message) — مو نص وأداة سوه حتى ما ينگال مرتين.
	var toolCalls []map[string]any
	content := reply.Text
	if endCall {
		args, _ := json.Marshal(map[string]string{"reason": "انتهت المكالمة — موظف يكمل ويا الزبون", "message": reply.Text})
		toolCalls = []map[string]any{{"index": 0, "id": "call_" + uuid.NewString(), "type": "function",
			"function": map[string]any{"name": "end_call", "arguments": string(args)}}}
		content = ""
	}
	finish := "stop"
	if endCall {
		finish = "tool_calls"
	}
	if !stream {
		msg := map[string]any{"role": "assistant", "content": content}
		if endCall {
			msg["tool_calls"] = toolCalls
		}
		WriteJSON(w, http.StatusOK, map[string]any{"id": id, "object": "chat.completion", "created": now, "model": model,
			"choices": []map[string]any{{"index": 0, "message": msg, "finish_reason": finish}},
			"usage":   map[string]int{"prompt_tokens": 0, "completion_tokens": 0, "total_tokens": 0}})
		return
	}
	chunk := sw.chunk
	if endCall {
		chunk(map[string]any{"tool_calls": toolCalls}, nil)
	} else {
		for _, part := range voiceSentences(content) {
			chunk(map[string]any{"content": part}, nil)
		}
	}
	chunk(map[string]any{}, finish)
	fmt.Fprint(w, "data: [DONE]\n\n")
	if sw.fl != nil {
		sw.fl.Flush()
	}
}

// voiceSentences يقسّم الرد جمل — محرك النطق يبدي بأول جملة بلا ما ينتظر الباقي.
func voiceSentences(s string) []string {
	var out []string
	start := 0
	rs := []rune(s)
	for i, r := range rs {
		if r == '.' || r == '؟' || r == '!' || r == '،' || r == '?' {
			out = append(out, string(rs[start:i+1]))
			start = i + 1
		}
	}
	if start < len(rs) {
		out = append(out, string(rs[start:]))
	}
	return out
}

// voiceCaller رقم المتصل، وإذا ماكو (زر التجربة بالمنصة) فرقم التجربة إذا المالك حطه.
func voiceCaller(phone, testPhone string) string {
	if phone == "" {
		return strings.TrimSpace(testPhone)
	}
	return phone
}
