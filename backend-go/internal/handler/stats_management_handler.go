package handler

import (
	"net/http"

	"staffmange-api/internal/service"
)

type StatsManagementHandler struct {
	service *service.StatsManagementService
}

func NewStatsManagementHandler(s *service.StatsManagementService) *StatsManagementHandler {
	return &StatsManagementHandler{service: s}
}

// GET /api/stats-management/daily?date=2026-07-29
func (h *StatsManagementHandler) Daily(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.Daily(r.URL.Query().Get("date"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الإحصائية اليومية")
		return
	}
	WriteJSON(w, http.StatusOK, stats)
}

// GET /api/stats-management/weekly?from=2026-07-01&to=2026-07-07
func (h *StatsManagementHandler) Weekly(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.Weekly(r.URL.Query().Get("from"), r.URL.Query().Get("to"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الإحصائية الأسبوعية")
		return
	}
	WriteJSON(w, http.StatusOK, stats)
}

// GET /api/stats-management/internal-works?month=2026-08
func (h *StatsManagementHandler) InternalWorks(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.InternalWorks(r.URL.Query().Get("month"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب إحصائية الأعمال داخل الشركة")
		return
	}
	WriteJSON(w, http.StatusOK, stats)
}

// GET /api/stats-management/projects
func (h *StatsManagementHandler) ProjectStages(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.ProjectStages()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب إحصائية المشاريع")
		return
	}
	WriteJSON(w, http.StatusOK, stats)
}
