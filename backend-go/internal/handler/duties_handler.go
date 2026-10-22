package handler

import (
	"net/http"
	"strconv"

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

// GET /api/duties/history/{employeeId}?days=30 — لقطات الأيام السابقة (للمراقب).
func (h *DutiesHandler) History(w http.ResponseWriter, r *http.Request) {
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days <= 0 || days > 120 {
		days = 30
	}
	rows, err := h.repo.History(r.PathValue("employeeId"), days)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب سجل الواجبات")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}
