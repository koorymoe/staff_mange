package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/repository"
)

// EmployeeHandoverHandler «🔁 تسليم شغله لموظف جديد» — للمدير والمالك (ع 10-17).
type EmployeeHandoverHandler struct{ repo *repository.EmployeeHandoverRepository }

func NewEmployeeHandoverHandler(r *repository.EmployeeHandoverRepository) *EmployeeHandoverHandler {
	return &EmployeeHandoverHandler{repo: r}
}

// GET /api/employees/{id}/handover — شنو راح ينتقل (بلا تغيير).
func (h *EmployeeHandoverHandler) Preview(w http.ResponseWriter, r *http.Request) {
	s, err := h.repo.Preview(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, s)
}

// POST /api/employees/{id}/handover {toEmployeeId}
func (h *EmployeeHandoverHandler) Handover(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ToEmployeeID string `json:"toEmployeeId"`
	}
	if err := DecodeJSON(r, &req); err != nil || req.ToEmployeeID == "" {
		WriteError(w, http.StatusBadRequest, "اختار الموظف المستلم")
		return
	}
	s, err := h.repo.Handover(r.PathValue("id"), req.ToEmployeeID, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, s)
}
