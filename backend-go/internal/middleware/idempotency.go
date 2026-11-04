package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"staffmange-api/internal/errlog"
	"time"

	"github.com/jmoiron/sqlx"
)

// ═══ 🔁 التكرار الآمن (تعديلات ٦أ — 11-04) ═══
// نفس طلب الفلوس إذا انرسل مرتين (ضغطة مزدوجة، نت تقطّع والمتصفح يعيد، زر
// انضغط من جهازين) ينسجل **مرة وحدة**، والثاني ياخذ نفس الرد بلا تسجيل.
//
// المفتاح: ترويسة Idempotency-Key إذا الواجهة أرسلتها، وإلا بصمة الطلب نفسه
// (الموظف + المسار + الجسم) خلال نافذة قصيرة — هيچ يحمي حتى بلا تعديل الواجهة.
// طلب ثاني وصل والأول بعده يشتغل = 409 «نفس الطلب قيد التنفيذ».

const idemWindow = 20 * time.Second

// CleanupIdempotency يمسح المفاتيح الأقدم من يومين. أطول حماية يوم وحد (المفتاح
// الصريح)، فالأقدم ما إلهن فايدة — بلا هذا الجدول يكبر للأبد ويا كل عملية فلوس.
func CleanupIdempotency(db *sqlx.DB) (int64, error) {
	res, err := db.Exec(`DELETE FROM "IdempotencyKey" WHERE "createdAt" < now() - interval '2 days'`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

type idemRecorder struct {
	http.ResponseWriter
	status int
	buf    bytes.Buffer
}

func (r *idemRecorder) WriteHeader(code int) { r.status = code; r.ResponseWriter.WriteHeader(code) }
func (r *idemRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	r.buf.Write(b)
	return r.ResponseWriter.Write(b)
}

// Idempotent لمسارات الفلوس — يُحط بآخر السلسلة (بعد الدخول والحارس).
func Idempotent(db *sqlx.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			emp := EmployeeIDFromContext(r)
			if db == nil || emp == "" {
				next.ServeHTTP(w, r)
				return
			}
			body, err := io.ReadAll(io.LimitReader(r.Body, 8<<20))
			if err != nil {
				http.Error(w, `{"error":"تعذر قراءة الطلب"}`, http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(body))
			key := r.Header.Get("Idempotency-Key")
			explicit := key != ""
			if !explicit {
				h := sha256.Sum256(append([]byte(emp+"|"+r.Method+"|"+r.URL.Path+"|"), body...))
				key = "fp:" + hex.EncodeToString(h[:])
			}
			// نحجز المفتاح. موجود؟ نرجّع الرد المحفوظ أو «قيد التنفيذ».
			res, err := db.Exec(`INSERT INTO "IdempotencyKey" (key, "employeeId", method, path) VALUES ($1, $2, $3, $4)
				ON CONFLICT (key, "employeeId") DO UPDATE SET "createdAt" = now(), status = NULL, response = NULL, method = EXCLUDED.method, path = EXCLUDED.path
				WHERE "IdempotencyKey"."createdAt" < now() - make_interval(secs => $5)`, key, emp, r.Method, r.URL.Path,
				func() float64 {
					if explicit {
						return 86400 // المفتاح الصريح يحمي يوم كامل
					}
					return idemWindow.Seconds()
				}())
			if err != nil {
				log.Printf("[idempotency] %v — نكمل بلا حماية", err)
				next.ServeHTTP(w, r)
				return
			}
			if n, _ := res.RowsAffected(); n == 0 {
				var prev struct {
					Status   *int    `db:"status"`
					Response *string `db:"response"`
				}
				_ = db.Get(&prev, `SELECT status, response FROM "IdempotencyKey" WHERE key = $1 AND "employeeId" = $2`, key, emp)
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Idempotent-Replay", "true")
				if prev.Status == nil {
					w.WriteHeader(http.StatusConflict)
					_, _ = w.Write([]byte(`{"error":"نفس الطلب قيد التنفيذ — لحظة"}`))
					return
				}
				w.WriteHeader(*prev.Status)
				if prev.Response != nil {
					_, _ = w.Write([]byte(*prev.Response))
				}
				log.Printf("[idempotency] تكرار انمنع emp=%s %s %s", emp, r.Method, r.URL.Path)
				return
			}
			rec := &idemRecorder{ResponseWriter: w}
			next.ServeHTTP(rec, r)
			if rec.status >= 500 || rec.status == 0 {
				// فشل بالخادم: نحرر المفتاح حتى إعادة المحاولة تشتغل
				errlog.Exec(db.Exec(`DELETE FROM "IdempotencyKey" WHERE key = $1 AND "employeeId" = $2`, key, emp))
				return
			}
			if rec.status >= 400 {
				// رفض أو خطأ إدخال: ما نحفظه — الموظف يصحح ويعيد بنفس البيانات
				errlog.Exec(db.Exec(`DELETE FROM "IdempotencyKey" WHERE key = $1 AND "employeeId" = $2`, key, emp))
				return
			}
			errlog.Exec(db.Exec(`UPDATE "IdempotencyKey" SET status = $3, response = $4 WHERE key = $1 AND "employeeId" = $2`,
				key, emp, rec.status, rec.buf.String()))
		})
	}
}
