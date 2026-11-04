package handler

import (
	"net/http"

	"staffmange-api/internal/service"
)

type SmartKpiHandler struct {
	service *service.SmartKpiService
}

func NewSmartKpiHandler(s *service.SmartKpiService) *SmartKpiHandler {
	return &SmartKpiHandler{service: s}
}

// GET /api/v1/smart-kpi/technician/{employeeId}?month=YYYY-MM
func (h *SmartKpiHandler) Technician(w http.ResponseWriter, r *http.Request) {
	if !requireSelfOrSupervisor(w, r, r.PathValue("employeeId")) {
		return
	}
	result, err := h.service.CalculateTechnicianKpi(r.PathValue("employeeId"), r.URL.Query().Get("month"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, result)
}

// GET /api/v1/smart-kpi/leaderboard?month=YYYY-MM
func (h *SmartKpiHandler) Leaderboard(w http.ResponseWriter, r *http.Request) {
	results, err := h.service.Leaderboard(r.URL.Query().Get("month"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, results)
}
