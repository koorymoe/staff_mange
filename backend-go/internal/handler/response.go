package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"staffmange-api/internal/monitor"

	"staffmange-api/internal/model"
)

type errorEnvelope struct {
	Error string `json:"error"`
}

// WriteJSON يرجّع الجسم مباشرة بدون أي غلاف — يطابق شكل استجابات الباك إند
// القديم (TypeScript/Express) بالضبط حتى يقدر الفرونت إند يتعامل مع الاثنين
// بدون أي تعديل.
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, status int, message string) {
	message = SafeErrorMessage(status, message)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorEnvelope{Error: model.Rebrand(message)})
}

// ═══ (11-04 — تعديلات ٧ط) أخطاء داخلية ما تطلع للمستخدم ═══
// ٣٧٤ مكان يمرر err.Error() خام: رسائل قاعدة البيانات وأسماء الجداول والأعمدة
// تطلع للموظف (وتكشف بنية النظام). هنا بمكان واحد: الرسالة الي تشبه خطأ داخلي
// تنسجل كاملة ويا رقم مرجع، والموظف ياخذ جملة عربية ويا نفس الرقم للدعم.
var internalErr = regexp.MustCompile(`(?i)\bpq:|sql:|sqlstate|violates|duplicate key|relation "|column "|syntax error|invalid input syntax|no rows in result|dial tcp|connection refused|context deadline|runtime error|nil pointer|json: cannot|unexpected end of json|deadlock|too many clients|i/o timeout|broken pipe`)

// SafeErrorMessage يرجّع الرسالة كما هي إذا مقصودة للمستخدم، وإلا جملة عامة + رقم مرجع.
func SafeErrorMessage(status int, message string) string {
	if !internalErr.MatchString(message) {
		return message
	}
	b := make([]byte, 3)
	_, _ = rand.Read(b)
	ref := hex.EncodeToString(b)
	log.Printf("[error-ref %s] status=%d %s", ref, status, message)
	monitor.Message(message, ref, status)
	if status >= 500 {
		return "صار خطأ بالسيرفر — جرّب مرة ثانية، وإذا تكرر خبّر الدعم بالرمز " + ref
	}
	return "ما انحفظ — تأكد من البيانات وجرّب مرة ثانية، وإذا تكرر خبّر الدعم بالرمز " + ref
}

func DecodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(dst)
}
