package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSonioxFlow(t *testing.T) {
	polls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer k" {
			w.WriteHeader(401)
			return
		}
		switch {
		case r.Method == "POST" && r.URL.Path == "/files":
			w.Write([]byte(`{"id":"f1"}`))
		case r.Method == "POST" && r.URL.Path == "/transcriptions":
			w.Write([]byte(`{"id":"t1","status":"queued"}`))
		case r.Method == "GET" && r.URL.Path == "/transcriptions/t1":
			polls++
			if polls < 2 {
				w.Write([]byte(`{"id":"t1","status":"processing"}`))
			} else {
				w.Write([]byte(`{"id":"t1","status":"completed"}`))
			}
		case r.Method == "GET" && r.URL.Path == "/transcriptions/t1/transcript":
			w.Write([]byte(`{"text":"هلو حبي شلونك"}`))
		case r.Method == "DELETE":
		default:
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()
	s := NewSonioxSTT("k", "")
	s.base = srv.URL
	got, err := s.Transcribe(context.Background(), []byte("x"), "audio/mpeg")
	if err != nil || got != "هلو حبي شلونك" {
		t.Fatalf("%q %v", got, err)
	}
}
