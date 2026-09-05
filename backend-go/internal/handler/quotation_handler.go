package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type QuotationHandler struct {
	service *service.QuotationService
}

func NewQuotationHandler(s *service.QuotationService) *QuotationHandler {
	return &QuotationHandler{service: s}
}

// GET /api/quotations?search=...
func (h *QuotationHandler) List(w http.ResponseWriter, r *http.Request) {
	quotations, err := h.service.List(middleware.EmployeeIDFromContext(r), middleware.RoleFromContext(r), r.URL.Query().Get("search"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب عروض الأسعار")
		return
	}
	WriteJSON(w, http.StatusOK, quotations)
}

// GET /api/quotations/{id}
func (h *QuotationHandler) Get(w http.ResponseWriter, r *http.Request) {
	quotation, err := h.service.Get(r.PathValue("id"), middleware.EmployeeIDFromContext(r), middleware.RoleFromContext(r))
	if err != nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, quotation)
}

// POST /api/quotations
func (h *QuotationHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateQuotationRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	quotation, err := h.service.Create(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, quotation)
}

// PUT /api/quotations/{id}
func (h *QuotationHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateQuotationRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	quotation, err := h.service.Update(r.PathValue("id"), middleware.EmployeeIDFromContext(r), middleware.RoleFromContext(r), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, quotation)
}

// DELETE /api/quotations/{id}
func (h *QuotationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Delete(r.PathValue("id")); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"success": true})
}
