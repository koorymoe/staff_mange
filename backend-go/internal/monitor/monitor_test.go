package monitor

import (
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestScrub(t *testing.T) {
	in := `pq: duplicate key Key (phone)=(07901234567) أحمد علي محمد a.b@x.com 9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d مبلغ 250000`
	out := Scrub(in)
	for _, bad := range []string{"07901234567", "أحمد", "a.b@x.com", "9b1deb4d", "250000"} {
		if strings.Contains(out, bad) {
			t.Fatalf("%q بعده موجود: %s", bad, out)
		}
	}
	if !strings.Contains(out, "duplicate key") {
		t.Fatalf("انمسح الكلام المفيد: %s", out)
	}
}

// الي يطلع فعلاً لـ Sentry (نلزمه بسيرفر وهمي) ما بيه اسم ولا رقم ولا إيميل.
func TestWhatLeavesIsClean(t *testing.T) {
	var mu sync.Mutex
	var got []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rd io.Reader = r.Body
		if r.Header.Get("Content-Encoding") == "gzip" {
			if gz, err := gzip.NewReader(r.Body); err == nil {
				rd = gz
			}
		}
		b, _ := io.ReadAll(rd)
		mu.Lock()
		got = append(got, string(b))
		mu.Unlock()
		w.WriteHeader(200)
	}))
	defer srv.Close()
	Init("http://pubkey@"+strings.TrimPrefix(srv.URL, "http://")+"/1", "test", "")
	defer func() { enabled = false }()
	Error(errors.New("تعذر حفظ فاتورة الزبون حسين كريم 07801112233 h@x.iq"), "x.go:1")
	Message("pq: Key (phone)=(07709998877) exists — سجاد", "abc123", 500)
	Flush()
	mu.Lock()
	defer mu.Unlock()
	all := strings.Join(got, "\n")
	if len(got) == 0 {
		t.Fatal("ما طلع شي لسيرفر Sentry الوهمي")
	}
	for _, bad := range []string{"حسين", "كريم", "07801112233", "h@x.iq", "07709998877", "سجاد"} {
		if strings.Contains(all, bad) {
			t.Fatalf("%q طلع لـ Sentry:\n%s", bad, all)
		}
	}
	if !strings.Contains(all, "abc123") || !strings.Contains(all, "x.go:1") {
		t.Fatal("الرقم المرجعي أو المكان ما طلع")
	}
}
