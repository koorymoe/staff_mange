package middleware

import (
	"net/http"
	"sync/atomic"
	"time"
)

// عداد ضغط بسيط بالذاكرة — يحسب الطلبات الكلية منذ إقلاع السيرفر، وطلبات آخر
// دقيقة تقريباً (نافذة متحركة بسيطة بحاويتين تتبدل كل دقيقة)، أساس شريط
// "الضغط الحي" بلوحة المراقبة الخلفية. ما يحتاج قاعدة بيانات ولا مكتبة خارجية.
var (
	totalRequests  int64
	currentBucket  int64
	previousBucket int64
)

func init() {
	go func() {
		ticker := time.NewTicker(time.Minute)
		for range ticker.C {
			atomic.StoreInt64(&previousBucket, atomic.SwapInt64(&currentBucket, 0))
		}
	}()
}

func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&totalRequests, 1)
		atomic.AddInt64(&currentBucket, 1)
		next.ServeHTTP(w, r)
	})
}

// TotalRequests يرجع عدد الطلبات الكلي منذ إقلاع السيرفر.
func TotalRequests() int64 {
	return atomic.LoadInt64(&totalRequests)
}

// RequestsLastMinute يرجع تقدير عدد الطلبات بآخر دقيقة تقريباً.
func RequestsLastMinute() int64 {
	return atomic.LoadInt64(&currentBucket) + atomic.LoadInt64(&previousBucket)
}
