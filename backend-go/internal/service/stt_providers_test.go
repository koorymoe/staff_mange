package service

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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

type fakeSTT struct {
	text  string
	err   error
	sleep time.Duration
}

func (f fakeSTT) Transcribe(ctx context.Context, _ []byte, _ string) (string, error) {
	select {
	case <-time.After(f.sleep):
		return f.text, f.err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func TestTranscribeChainFallsBack(t *testing.T) {
	chain := []NamedSTT{
		{Name: "pro", P: fakeSTT{sleep: time.Second}, Timeout: 50 * time.Millisecond},
		{Name: "bad", P: fakeSTT{err: errors.New("400")}},
		{Name: "flash", P: fakeSTT{text: "هلو حبي شلونك"}},
	}
	start := time.Now()
	got, name, err := TranscribeChain(context.Background(), chain, nil, "audio/mpeg")
	if err != nil || name != "flash" || got != "هلو حبي شلونك" {
		t.Fatalf("%q %q %v", got, name, err)
	}
	if time.Since(start) > 500*time.Millisecond {
		t.Fatal("slow provider was not cut by its timeout")
	}
}

func TestTranscribeChainAllFail(t *testing.T) {
	_, _, err := TranscribeChain(context.Background(), []NamedSTT{{Name: "a", P: fakeSTT{err: errors.New("x")}}}, nil, "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCompareSTTSlowProvider(t *testing.T) {
	res := CompareSTT(context.Background(), []NamedSTT{
		{Name: "slow", P: fakeSTT{sleep: time.Second}, Timeout: 50 * time.Millisecond},
		{Name: "ok", P: fakeSTT{text: "زين"}},
	}, nil, "")
	if res[0].Error == "" || res[1].Text != "زين" {
		t.Fatal(res)
	}
}
