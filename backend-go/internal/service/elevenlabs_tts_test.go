package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// (دازاي 11-17) كمال يحچي: النص ينظّف، والطلب يروح بالمفتاح والصوت، والصوت يتخزن.
func TestElevenLabsTTS(t *testing.T) {
	if NewElevenLabsTTS("", "v", "") != nil || NewElevenLabsTTS("k", "", "") != nil {
		t.Fatal("بلا مفتاح أو صوت = مطفي")
	}
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("xi-api-key") != "k" || !strings.Contains(r.URL.Path, "/text-to-speech/voice1") {
			w.WriteHeader(401)
			return
		}
		var in map[string]any
		_ = json.NewDecoder(r.Body).Decode(&in)
		if txt, _ := in["text"].(string); strings.ContainsAny(txt, "*#🙂") {
			w.WriteHeader(400)
			return
		}
		_, _ = w.Write([]byte("MP3"))
	}))
	defer srv.Close()
	tts := NewElevenLabsTTS("k", "voice1", "")
	tts.base = srv.URL
	for i := 0; i < 2; i++ {
		b, err := tts.Speak(context.Background(), "**هلا** بيك 🙂 # تدلل")
		if err != nil || string(b) != "MP3" {
			t.Fatalf("Speak: %q %v", b, err)
		}
	}
	if calls != 1 {
		t.Fatalf("نفس النص لازم ينحفظ — الطلبات %d", calls)
	}
	var nilTTS *ElevenLabsTTS
	if _, err := nilTTS.Speak(context.Background(), "x"); err != ErrTTSOff {
		t.Fatal("المطفي لازم يرجّع ErrTTSOff")
	}
}
