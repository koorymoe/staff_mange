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

// GET /api/stats-management/weekly
func (h *StatsManagementHandler) Weekly(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.Weekly()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الإحصائية الأسبوعية")
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
