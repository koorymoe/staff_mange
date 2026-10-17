package handler

import (
	"net/http"

	"staffmange-api/internal/service"
)

// MatrixForesightHandler مسارات ماتركس الاستباقية: التنبؤ بالتأخير وفرص البيع.
type MatrixForesightHandler struct {
	delays *service.DelayPredictionService
	sales  *service.SalesOpportunityService
}

func NewMatrixForesightHandler(delays *service.DelayPredictionService, sales *service.SalesOpportunityService) *MatrixForesightHandler {
	return &MatrixForesightHandler{delays: delays, sales: sales}
}

// GET /api/ai/delay-risks — حجوزات اليوم و٦ أيام جاية المتوقع تطوّل أكثر من وقتها.
func (h *MatrixForesightHandler) DelayRisks(w http.ResponseWriter, r *http.Request) {
	rows, err := h.delays.Upcoming()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر حساب توقعات التأخير")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// GET /api/ai/opportunities — فرص البيع.
func (h *MatrixForesightHandler) Opportunities(w http.ResponseWriter, r *http.Request) {
	rows, err := h.sales.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر حساب فرص البيع")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}
