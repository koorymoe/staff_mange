package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// clientIPForRateLimit يحدد عنوان IP الحقيقي للطلب (يفضّل X-Forwarded-For
// اللي يضيفه nginx، مطابق لنفس منطق clientIP بحزمة handler، بس منسوخ هنا
// تفادياً لحلقة استيراد بين middleware و handler).
func clientIPForRateLimit(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		parts := strings.Split(fwd, ",")
		return strings.TrimSpace(parts[0])
	}
	if real := r.Header.Get("X-Real-IP"); real != "" {
		return real
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

// RateLimit يمنع أكثر من maxAttempts طلب من نفس عنوان IP خلال المدة window —
// حماية أساسية ضد محاولات تخمين كلمة السر المتكررة (Brute Force)، سواءً من
// مهاجم حقيقي أو حتى اختبار داخلي سريع كان سبب حظر IP السيرفر من Hetzner سابقاً.
func RateLimit(maxAttempts int, window time.Duration) func(http.Handler) http.Handler {
	var mu sync.Mutex
	attempts := make(map[string][]time.Time)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIPForRateLimit(r)
			now := time.Now()
			cutoff := now.Add(-window)

			mu.Lock()
			recent := attempts[ip][:0]
			for _, t := range attempts[ip] {
				if t.After(cutoff) {
					recent = append(recent, t)
				}
			}
			if len(recent) >= maxAttempts {
				mu.Unlock()
				writeError(w, http.StatusTooManyRequests, "محاولات كثيرة جداً بوقت قصير — انتظر دقيقة وجرب مرة ثانية")
				return
			}
			attempts[ip] = append(recent, now)
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}
