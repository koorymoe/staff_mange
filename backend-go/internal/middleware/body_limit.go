package middleware

import "net/http"

// MaxBodyBytes هو الحد الأقصى المسموح به لحجم أي جسم طلب (request body) بالبايت.
//
// السبب باختيار 25 ميغابايت: أكبر حمولة شرعية بالنظام هي مرفقات مركبات
// (صور/فيديوهات أعطال) تُرسل base64 داخل جسم JSON مباشرة (بدون رفع مباشر
// لـR2) — راجع CreateVehiclePhoto / CreateVehicleIncidentAttachment بالفرونت
// إند. الترميز base64 يضخّم الحجم ~33%، فحمولة JSON بحجم 25 ميغا تسمح بملف
// أصلي (صورة أو مقطع فيديو قصير) حتى ~18-19 ميغابايت تقريباً — يغطي بسخاء
// صور الوثائق/الهوية وأغلب مقاطع الفيديو القصيرة لحوادث المركبات، بدون
// فتح الباب لحمولة عملاقة (مئات الميغابايت) تقدر تُستخدم بهجوم حرمان خدمة.
// لو احتجنا لاحقاً فيديوهات أطول، الحل الصحيح رفعها مباشرة لـR2 (presigned
// upload) مو تضمينها بجسم JSON — مو رفع هذا الحد أكثر.
const MaxBodyBytes = 25 << 20 // 25 MiB

// BodyLimit يحد حجم أي جسم طلب وارد حتى ما يقدر عميل (خبيث أو مسيء) يرسل
// جسم ضخم بشكل تعسفي ويستهلك ذاكرة/نطاق السيرفر (DoS). القراءات اللاحقة
// لـr.Body (json.Decode أو io.ReadAll) ترجع خطأ لما يتجاوز الحد، والهاندلرز
// أصلاً تتعامل مع خطأ فك الترميز بردّ 400 عادي (راجع DecodeJSON بالهاندلرز)
// — مو كراش أو 500.
func BodyLimit(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
