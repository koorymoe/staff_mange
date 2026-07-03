package handler

import (
	"net/http"

	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type ExpenseHandler struct {
	service *service.ExpenseService
}

func NewExpenseHandler(s *service.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{service: s}
}

// GET /api/v1/expenses?employeeId=
func (h *ExpenseHandler) List(w http.ResponseWriter, r *http.Request) {
	expenses, err := h.service.List(r.URL.Query().Get("employeeId"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب المصاريف")
		return
	}
	WriteJSON(w, http.StatusOK, expenses)
}

// POST /api/v1/expenses
func (h *ExpenseHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateExpenseRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	expense, err := h.service.Create(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, expense)
}

// PUT /api/v1/expenses/{id}/status
func (h *ExpenseHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateExpenseStatusRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	expense, err := h.service.UpdateStatus(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, expense)
}
