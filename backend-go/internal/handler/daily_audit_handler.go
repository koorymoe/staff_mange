package handler

import (
	"log"
	"net/http"
	"strings"
	"time"

	"staffmange-api/internal/repository"
)

type DailyAuditHandler struct {
	repo *repository.DailyAuditRepository
}

func NewDailyAuditHandler(r *repository.DailyAuditRepository) *DailyAuditHandler {
	return &DailyAuditHandler{repo: r}
}

// GET /api/finance/daily-audit?date=2026-08-04 — اليوم افتراضياً
//
// ومع `?q=...` يصير **بحثاً بلا تاريخ** عبر كل الأيام (رقم الحجز أو
// اسم الزبون أو هاتفه). التاريخ ينتجاهل وقتها — المحاسب يدوّر على حجز
// ما يعرف تاريخه، وهذا كل المطلوب.
func (h *DailyAuditHandler) Day(w http.ResponseWriter, r *http.Request) {
	if q := strings.TrimSpace(r.URL.Query().Get("q")); q != "" {
		rep, err := h.repo.Search(q)
		if err != nil {
			log.Printf("daily audit search: %v", err)
			WriteError(w, http.StatusInternalServerError, "تعذر البحث بحجوزات التدقيق")
			return
		}
		WriteJSON(w, http.StatusOK, rep)
		return
	}
	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	rep, err := h.repo.Day(date)
	if err != nil {
		log.Printf("daily audit: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب تدقيق اليوم")
		return
	}
	WriteJSON(w, http.StatusOK, rep)
}
