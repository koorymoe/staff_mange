package handler

import (
	"net/http"

	"staffmange-api/internal/service"
)

// MoneyAuditHandler 🔎 فحص الأرقام (دازاي 11-26) — للمالك بس، قراءة بس.
type MoneyAuditHandler struct{ svc *service.MoneyAuditService }

func NewMoneyAuditHandler(s *service.MoneyAuditService) *MoneyAuditHandler {
	return &MoneyAuditHandler{svc: s}
}

// GET /api/owner/money-audit?period=YYYY-MM
func (h *MoneyAuditHandler) Get(w http.ResponseWriter, r *http.Request) {
	a, err := h.svc.Run(r.URL.Query().Get("period"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الفحص")
		return
	}
	WriteJSON(w, http.StatusOK, a)
}
