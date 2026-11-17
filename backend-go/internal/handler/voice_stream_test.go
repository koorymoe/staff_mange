package handler

import (
	"net/http/httptest"
	"strings"
	"testing"

	"staffmange-api/internal/service"
)

// (دازاي 11-17 «التسريع») جملة الانتظار تطلع أول — قبل الرد النهائي — ويا رأس SSE مرة وحدة.
func TestVoiceStreamHoldOnFirst(t *testing.T) {
	rec := httptest.NewRecorder()
	sw := newVoiceStream(rec, "", true)
	sw.text("لحظة خلي أشوفلك. ")
	(&VoiceLLMHandler{}).write(sw, service.VoiceReply{Text: "عندنا باچر الساعة عشرة."}, false)
	out := rec.Body.String()
	hold := strings.Index(out, "لحظة خلي أشوفلك")
	final := strings.Index(out, "عندنا باچر")
	if hold < 0 || final < 0 || hold > final {
		t.Fatalf("الترتيب غلط:\n%s", out)
	}
	if strings.Count(out, `"role":"assistant"`) != 1 {
		t.Fatalf("رأس الدور لازم مرة وحدة:\n%s", out)
	}
	if !strings.HasSuffix(strings.TrimSpace(out), "data: [DONE]") {
		t.Fatalf("لازم ينتهي بـDONE:\n%s", out)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("Content-Type: %s", ct)
	}
}
