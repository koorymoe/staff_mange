package handler

import (
	"context"
	"errors"
	"net"
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

// allowedMapHosts نطاقات الخرائط المسموح بفتحها من السيرفر.
//
// ثغرة كانت هنا (SSRF): الهاندلر جان يفتح *أي* رابط http/https يرسله الموظف
// ويتبع التحويلات — يعني يقدر يخلي السيرفر يطلب خدمات داخلية ما إله وصول
// إلها (قاعدة البيانات، واجهات إدارة، بيانات وصفية للسحابة على
// 169.254.169.254...). الحل طبقتين: قائمة نطاقات مسموحة + منع أي عنوان
// داخلي/خاص حتى لو النطاق مسموح وحوّل لعنوان داخلي (DNS rebinding).
var allowedMapHosts = map[string]bool{
	"maps.app.goo.gl": true, "goo.gl": true, "maps.google.com": true,
	"www.google.com": true, "google.com": true, "maps.googleapis.com": true,
	"openstreetmap.org": true, "www.openstreetmap.org": true, "osm.org": true,
	"maps.apple.com": true, "waze.com": true, "www.waze.com": true, "ul.waze.com": true,
}

func isAllowedMapHost(host string) bool {
	h := strings.ToLower(host)
	if i := strings.IndexByte(h, ':'); i >= 0 {
		h = h[:i]
	}
	return allowedMapHosts[h]
}

// isBlockedIP يمنع كل ما هو داخلي: loopback، شبكات خاصة، link-local (ومنها
// عنوان بيانات السحابة 169.254.169.254)، multicast، وغير المحدد.
func isBlockedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsInterfaceLocalMulticast() ||
		ip.IsMulticast() || ip.IsUnspecified()
}

// safeDialContext يفحص العنوان الفعلي بعد ترجمة الـDNS وقبل فتح الاتصال —
// هذا يقفل حتى نطاقاً مسموحاً "يترجم" لعنوان داخلي.
func safeDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	for _, a := range ips {
		if isBlockedIP(a.IP) {
			return nil, errors.New("العنوان غير مسموح")
		}
	}
	d := &net.Dialer{Timeout: 5 * time.Second}
	return d.DialContext(ctx, network, net.JoinHostPort(ips[0].IP.String(), port))
}

func NewMapLinkHandler() *MapLinkHandler {
	return &MapLinkHandler{
		client: &http.Client{
			Timeout:   10 * time.Second,
			Transport: &http.Transport{DialContext: safeDialContext},
			// نوقف عند 5 تحويلات، ونفحص كل تحويل: لازم يبقى ضمن النطاقات
			// المسموحة (التحويل هو الطريق الكلاسيكي لتجاوز قائمة النطاقات).
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 5 {
					return http.ErrUseLastResponse
				}
				if !isAllowedMapHost(req.URL.Host) {
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
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		WriteError(w, http.StatusBadRequest, "الرابط غير صحيح — لازم يكون رابط https")
		return
	}
	if !isAllowedMapHost(parsed.Host) {
		WriteError(w, http.StatusBadRequest, "هذا الرابط مو من مواقع الخرائط المعروفة — الصق رابط من كوكل ماب")
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
