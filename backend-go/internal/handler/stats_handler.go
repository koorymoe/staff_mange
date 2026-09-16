package handler

import (
	"net/http"

	"staffmange-api/internal/service"
)

type StatsHandler struct {
	service *service.StatsService
}

func NewStatsHandler(s *service.StatsService) *StatsHandler {
	return &StatsHandler{service: s}
}

// GET /api/v1/stats
func (h *StatsHandler) Overview(w http.ResponseWriter, r *http.Request) {
	overview, err := h.service.Overview()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الإحصائيات")
		return
	}
	WriteJSON(w, http.StatusOK, overview)
}
