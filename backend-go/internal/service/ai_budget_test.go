package service

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// الحد اليومي: بعد ما يوصل الوكيل حده، الطلب ما يطلع للمزوّد أبداً.
func TestAIBudgetBlocksBeforeNetwork(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	aiBudgetMu.Lock()
	aiBudgetDay = time.Now().In(debriefLoc).Format("2006-01-02")
	aiBudgetSpent = map[string]float64{}
	aiBudgetMu.Unlock()
	SetAIBudget("TEST", 0.01)
	defer SetAIBudget("TEST", 0)

	call := func() error {
		c := NewAIClient("TEST", "sk-test")
		_, err := c.Messages.New(context.Background(), anthropic.MessageNewParams{
			Model: "claude-haiku-5-5", MaxTokens: 10,
			Messages: []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock("هلا"))},
		}, option.WithBaseURL(srv.URL), option.WithMaxRetries(0))
		return err
	}
	if err := call(); err == nil || errors.Is(err, ErrAIBudget) {
		t.Fatalf("under budget should reach the server, got %v", err)
	}
	aiBudgetMu.Lock()
	aiBudgetSpent["TEST"] = 0.02
	aiBudgetMu.Unlock()
	before := atomic.LoadInt32(&hits)
	if err := call(); !errors.Is(err, ErrAIBudget) {
		t.Fatalf("over budget: want ErrAIBudget, got %v", err)
	}
	if atomic.LoadInt32(&hits) != before {
		t.Fatal("request reached the network while over budget")
	}
	if p := priceOf("claude-haiku-5-5"); p.In != 0.10 || p.Out != 0.50 {
		t.Fatalf("haiku 5.5 price wrong: %+v", p)
	}
}
