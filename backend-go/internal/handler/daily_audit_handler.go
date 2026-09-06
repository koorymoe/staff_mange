package handler

import (
	"log"
	"net/http"
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
func (h *DailyAuditHandler) Day(w http.ResponseWriter, r *http.Request) {
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
