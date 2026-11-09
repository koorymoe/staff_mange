package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// ٦أ — نفس الطلب ٥ مرات بنفس اللحظة = المعالج يشتغل مرة وحدة، والباقي 409 أو نفس الرد.
func TestIdempotentConcurrent(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	defer db.Close()
	var calls int32
	h := Idempotent(db)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		time.Sleep(150 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	emp := "idem_test_" + time.Now().Format("150405.000")
	body := `{"amount":5}`
	var wg sync.WaitGroup
	codes := make([]int, 5)
	for i := range codes {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			req := httptest.NewRequest("POST", "/api/x/pay", strings.NewReader(body))
			req = req.WithContext(context.WithValue(req.Context(), ContextEmployeeID, emp))
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)
			codes[i] = rr.Code
		}(i)
	}
	wg.Wait()
	if calls != 1 {
		t.Fatalf("المعالج اشتغل %d مرة (لازم ١) — codes %v", calls, codes)
	}
	// بعد ما خلص: إعادة = نفس الرد المحفوظ بلا تشغيل
	req := httptest.NewRequest("POST", "/api/x/pay", strings.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), ContextEmployeeID, emp))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if calls != 1 || rr.Code != 200 || rr.Header().Get("Idempotent-Replay") != "true" {
		t.Fatalf("الإعادة ما رجعت الرد المحفوظ: calls=%d code=%d", calls, rr.Code)
	}
	_, _ = db.Exec(`DELETE FROM "IdempotencyKey" WHERE "employeeId" = $1`, emp)
}
