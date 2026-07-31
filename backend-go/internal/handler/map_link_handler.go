package handler

import (
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// MapLinkHandler يحل روابط الخرائط ويطلع منها إحداثيات.
//
// ليش نحتاج السيرفر أصلاً؟ روابط كوكل ماب المختصرة (maps.app.goo.gl/xxxx و
// goo.gl/maps/xxxx) ما تحتوي أي إحداثيات — هي مجرد مفتاح، والإحداثيات تبين
// بس بعد ما ينفتح الرابط ويتحول للرابط الطويل. المتصفح ما يقدر يتبع هذا
// التحويل بنفسه (CORS يمنعه)، فالسيرفر يتبعه ويرجع النقطة جاهزة.
type MapLinkHandler struct {
	client *http.Client
}

func NewMapLinkHandler() *MapLinkHandler {
	return &MapLinkHandler{
		client: &http.Client{
			Timeout: 10 * time.Second,
			// نوقف عند 10 تحويلات حتى ما ندور بحلقة مفرغة لو الرابط خربان.
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return http.ErrUseLastResponse
				}
				return nil
			},
		},
	}
}

// أنماط الإحداثيات بروابط الخرائط، مرتبة بالأولوية: !3d/!4d هو الأدق (نقطة
// المكان نفسه) بينما @lat,lng ممكن يكون بس مركز الكاميرا.
var mapCoordPatterns = []*regexp.Regexp{
	regexp.MustCompile(`!3d(-?\d+\.\d+)!4d(-?\d+\.\d+)`),
	regexp.MustCompile(`[?&]q=(-?\d+\.\d+),\s*(-?\d+\.\d+)`),
	regexp.MustCompile(`[?&]destination=(-?\d+\.\d+),\s*(-?\d+\.\d+)`),
	regexp.MustCompile(`[?&]ll=(-?\d+\.\d+),\s*(-?\d+\.\d+)`),
	regexp.MustCompile(`[?&]mlat=(-?\d+\.\d+)[^&]*&mlon=(-?\d+\.\d+)`),
	regexp.MustCompile(`@(-?\d+\.\d+),\s*(-?\d+\.\d+)`),
	regexp.MustCompile(`#map=\d+/(-?\d+\.\d+)/(-?\d+\.\d+)`),
	regexp.MustCompile(`/(-?\d+\.\d+),\s*(-?\d+\.\d+)`),
}

func extractCoords(raw string) (float64, float64, bool) {
	// الرابط الملصوق ممكن يكون مُرمّز (%2C بدل الفاصلة) فنفك الترميز أول شي.
	if decoded, err := url.QueryUnescape(raw); err == nil {
		raw = decoded
	}
	for _, re := range mapCoordPatterns {
		if m := re.FindStringSubmatch(raw); m != nil {
			lat, err1 := strconv.ParseFloat(m[1], 64)
			lng, err2 := strconv.ParseFloat(m[2], 64)
			if err1 == nil && err2 == nil && lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180 {
				return lat, lng, true
			}
		}
	}
	return 0, 0, false
}

func (h *MapLinkHandler) Resolve(w http.ResponseWriter, r *http.Request) {
	raw := strings.TrimSpace(r.URL.Query().Get("url"))
	if raw == "" {
		WriteError(w, http.StatusBadRequest, "الرابط مطلوب")
		return
	}
	parsed, err := url.Parse(raw)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		WriteError(w, http.StatusBadRequest, "الرابط غير صحيح")
		return
	}

	// أول محاولة: الإحداثيات موجودة بالرابط نفسه (رابط طويل) — بلا شبكة أصلاً.
	if lat, lng, ok := extractCoords(raw); ok {
		WriteJSON(w, http.StatusOK, map[string]float64{"lat": lat, "lng": lng})
		return
	}

	// ثاني محاولة: نتبع التحويل ونفتش بالرابط النهائي وبجسم الصفحة.
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, raw, nil)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "الرابط غير صحيح")
		return
	}
	// كوكل يرجع صفحة بلا إحداثيات لو حسبنا بوت — نعرّف نفسنا كمتصفح عادي.
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0 Safari/537.36")
	resp, err := h.client.Do(req)
	if err != nil {
		WriteError(w, http.StatusBadGateway, "تعذر فتح الرابط — تأكد من الاتصال أو حدد الموقع على الخريطة")
		return
	}
	defer resp.Body.Close()

	if lat, lng, ok := extractCoords(resp.Request.URL.String()); ok {
		WriteJSON(w, http.StatusOK, map[string]float64{"lat": lat, "lng": lng})
		return
	}

	// آخر محاولة: نقرأ أول جزء من الصفحة (محدود بـ512 كيلوبايت حتى ما نبلع
	// صفحة ضخمة) وندور بيها على الإحداثيات.
	body := make([]byte, 512*1024)
	n, _ := resp.Body.Read(body)
	if lat, lng, ok := extractCoords(string(body[:n])); ok {
		WriteJSON(w, http.StatusOK, map[string]float64{"lat": lat, "lng": lng})
		return
	}

	WriteError(w, http.StatusUnprocessableEntity, "ما كدرنا نطلع الإحداثيات من هذا الرابط — حدد الموقع على الخريطة")
}
