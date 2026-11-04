package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGeminiTranscriber(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-goog-api-key") != "k" {
			w.WriteHeader(403)
			return
		}
		var in map[string]any
		_ = json.NewDecoder(r.Body).Decode(&in)
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"«وجّه مهمة لياسر باچر الصبح»"}]}}]}`))
	}))
	defer srv.Close()
	g := NewGeminiTranscriber("k")
	g.url = srv.URL
	got, err := g.Transcribe(context.Background(), []byte("OggS"), "audio/ogg; codecs=opus")
	if err != nil || got != "وجّه مهمة لياسر باچر الصبح" {
		t.Fatalf("got %q %v", got, err)
	}
	g.key = "bad"
	if _, err := g.Transcribe(context.Background(), []byte("x"), "audio/webm"); err == nil {
		t.Fatal("expected error on 403")
	}
	if NewGeminiTranscriber("") != nil {
		t.Fatal("no key should be nil")
	}
}
