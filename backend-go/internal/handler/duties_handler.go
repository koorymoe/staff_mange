package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/repository"
)

// DutiesHandler «واجباتك اليوم» — كل منصب اله أهدافه اليومية (ع 10-17).
type DutiesHandler struct{ repo *repository.DutiesRepository }

func NewDutiesHandler(r *repository.DutiesRepository) *DutiesHandler { return &DutiesHandler{repo: r} }

// GET /api/duties/me
func (h *DutiesHandler) Me(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, h.repo.Today(middleware.EmployeeIDFromContext(r)))
}

// GET /api/duties/{employeeId} — للمراقب
func (h *DutiesHandler) ForEmployee(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, h.repo.Today(r.PathValue("employeeId")))
}
