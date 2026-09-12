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
// ثغرة كانت هنا: الكود جان ياخذ X-Forwarded-For مباشرةً وهي ترويسة يكتبها
// العميل بنفسه. يعني مهاجم يبعث بكل محاولة X-Forwarded-For بقيمة عشوائية
// فيصير كل طلب "IP جديد" ويتجاوز حد المحاولات بالكامل — تخمين كلمات سر بلا
// سقف. الحل: ما نثق بالترويسة إلا إذا الطلب جاي من بروكسي معروف (nginx على
// نفس الجهاز)، وغيرها نعتمد عنوان الاتصال الحقيقي.
func clientIPForRateLimit(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	if isTrustedProxy(host) {
		if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
			parts := strings.Split(fwd, ",")
			return strings.TrimSpace(parts[0])
		}
		if real := r.Header.Get("X-Real-IP"); real != "" {
			return strings.TrimSpace(real)
		}
	}
	return host
}

// isTrustedProxy: منو نثق بترويسة X-Forwarded-For الجاية منه.
//
// انتباه مهم: النظام يشتغل بـDocker — السلسلة Caddy -> nginx -> الباك اند،
// والباك اند يشوف عنوان حاوية nginx (شبكة خاصة 172.x) *مو* loopback. لو
// حصرناها بـloopback بس، تنرفض الترويسة دائماً ويصير كل المستخدمين محسوبين
// بعنوان واحد — يعني 8 محاولات فاشلة من أي شخص تقفل الدخول على الشركة كلها
// (حرمان خدمة كامل). فنثق بالعناوين الخاصة/المحلية (شبكة Docker الداخلية)،
// وهي أصلاً مو قابلة للوصول من الإنترنت مباشرةً.
func isTrustedProxy(host string) bool {
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast()
}

// RateLimit يمنع أكثر من maxAttempts طلب من نفس عنوان IP خلال المدة window —
// حماية أساسية ضد محاولات تخمين كلمة السر المتكررة (Brute Force)، سواءً من
// مهاجم حقيقي أو حتى اختبار داخلي سريع كان سبب حظر IP السيرفر من Hetzner سابقاً.
// maxTrackedClients أقصى عدد عناوين نتابعها قبل ما ننظّف القديم منها.
const maxTrackedClients = 10000

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

			// تنظيف دوري: بدونه الخريطة تكبر بلا حدود مع كل IP جديد وتصير
			// باب لاستنزاف الذاكرة (DoS). ننظّف كل ما تتجاوز حداً معقولاً.
			if len(attempts) > maxTrackedClients {
				for k, v := range attempts {
					if len(v) == 0 || v[len(v)-1].Before(cutoff) {
						delete(attempts, k)
					}
				}
			}
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}
