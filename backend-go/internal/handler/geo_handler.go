package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// GeoHandler بحث المناطق.
//
// قبل، الواجهة جانت تنادي خدمة الخرائط (Nominatim) مباشرة من المتصفح.
// وهاي الخدمة تحدد الطلبات حسب عنوان الإنترنت وتطلب تعريف بالتطبيق —
// فكل موظفينا يطلعون بنفس العنوان، والخدمة تحجبهم ويطلع «تعذر البحث».
//
// هسه الطلب يمر من سيرفرنا: نعرّف بنفسنا مثل ما تطلب الخدمة، ونخزّن
// نتيجة كل كلمة نصف ساعة — فالبحوث المتكررة ما توصل للخدمة أصلاً.
type GeoHandler struct {
	client *http.Client
	mu     sync.Mutex
	cache  map[string]geoCacheEntry
}

type geoCacheEntry struct {
	body []byte
	at   time.Time
}

const geoCacheTTL = 30 * time.Minute

func NewGeoHandler() *GeoHandler {
	return &GeoHandler{
		client: &http.Client{Timeout: 10 * time.Second},
		cache:  map[string]geoCacheEntry{},
	}
}

// GET /api/geo/search?q=...&limit=5&viewbox=...&bounded=1
func (h *GeoHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		WriteJSON(w, http.StatusOK, []any{})
		return
	}

	params := url.Values{}
	params.Set("format", "json")
	params.Set("countrycodes", "iq")
	params.Set("accept-language", "ar")
	params.Set("q", q)
	if v := r.URL.Query().Get("limit"); v != "" {
		params.Set("limit", v)
	} else {
		params.Set("limit", "5")
	}
	// الحصر بمنطقة الموظف — يجي من الواجهة وقت البحث الفوري
	if v := r.URL.Query().Get("viewbox"); v != "" {
		params.Set("viewbox", v)
		if r.URL.Query().Get("bounded") == "1" {
			params.Set("bounded", "1")
		}
	}
	key := params.Encode()

	h.mu.Lock()
	entry, ok := h.cache[key]
	h.mu.Unlock()
	if ok && time.Since(entry.at) < geoCacheTTL {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write(entry.body)
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet,
		"https://nominatim.openstreetmap.org/search?"+key, nil)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر البحث عن المنطقة")
		return
	}
	// الخدمة تطلب تعريف واضح بالتطبيق، وبدونه تحجب الطلب
	req.Header.Set("User-Agent", "StaffMange/1.0 (+https://staffmange.cc)")
	req.Header.Set("Accept-Language", "ar")

	resp, err := h.client.Do(req)
	if err != nil {
		WriteError(w, http.StatusBadGateway, "تعذر الوصول لخدمة الخرائط، حدد الموقع من الخريطة")
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil || resp.StatusCode != http.StatusOK {
		WriteError(w, http.StatusBadGateway, "خدمة الخرائط ما استجابت، حدد الموقع من الخريطة")
		return
	}
	// نتأكد إن الرد فعلاً قائمة نتائج مو صفحة خطأ
	var probe []map[string]any
	if json.Unmarshal(body, &probe) != nil {
		WriteError(w, http.StatusBadGateway, "خدمة الخرائط رجعت رد غير مفهوم، حدد الموقع من الخريطة")
		return
	}

	h.mu.Lock()
	if len(h.cache) > 500 {
		h.cache = map[string]geoCacheEntry{}
	}
	h.cache[key] = geoCacheEntry{body: body, at: time.Now()}
	h.mu.Unlock()

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write(body)
}
