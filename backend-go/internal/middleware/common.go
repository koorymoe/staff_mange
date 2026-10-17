package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// CORS يسمح فقط للأصول المذكورة بـ allowedOrigins (مطابق لقيد CORS_ORIGIN
// بالباك إند القديم) — بدون قائمة، ما يرد أي رأس Access-Control-Allow-Origin
// (يعني مرفوض من كل المتصفحات cross-origin بشكل افتراضي آمن).
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[o] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if allowed[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			w.Header().Set("Vary", "Origin")
			// ⚠️ `PATCH` مو موجودة چانت، فأي مسار PATCH ينفشل من المتصفح
			// بـ«Failed to fetch» **قبل ما يوصل السيرفر أصلاً** — الطلب
			// التمهيدي (preflight) يرفضه المتصفح لحاله. وينجح بـcurl
			// (ماكو preflight)، فالفرق بين «يشتغل بالفحص» و«ما يشتغل
			// بالواجهة» يضيّع ساعات. انكشف بأول مسار PATCH بالنظام.
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "حدث خطأ غير متوقع بالسيرفر"})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Chain يطبق قائمة middlewares بالترتيب على الراوتر النهائي
func Chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// SecurityHeaders يضيف ترويسات الحماية القياسية لكل رد:
//   - X-Content-Type-Options: يمنع المتصفح من "تخمين" نوع المحتوى (هجوم
//     MIME sniffing يخلي ملف مرفوع ينفّذ كأنه HTML/JS).
//   - X-Frame-Options: يمنع وضع النظام داخل iframe بموقع خبيث (Clickjacking).
//   - Referrer-Policy: ما يسرّب مسارات النظام الداخلية للمواقع الخارجية.
//   - Strict-Transport-Security: يجبر المتصفح يستخدم HTTPS دائماً.
//   - Permissions-Policy: يقفل صلاحيات المتصفح الي ما نحتاجها.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Permissions-Policy", "camera=(), microphone=(), payment=(), usb=()")
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		// HSTS ينضاف بس لما الطلب فعلاً على HTTPS (وراء nginx)
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		next.ServeHTTP(w, r)
	})
}
