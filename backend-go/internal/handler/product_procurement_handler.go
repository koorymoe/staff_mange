package handler

import (
	"log"
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type ProductProcurementHandler struct {
	service *service.ProductProcurementService
}

func NewProductProcurementHandler(s *service.ProductProcurementService) *ProductProcurementHandler {
	return &ProductProcurementHandler{service: s}
}

// GET /api/product-procurements?status=PENDING
func (h *ProductProcurementHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.service.List(r.URL.Query().Get("status"))
	if err != nil {
		log.Printf("list product procurements: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب تجهيزات المنتجات")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// POST /api/product-requests/{id}/fulfill — أبو الحسابات يجهّز الطلب من الدوار
func (h *ProductProcurementHandler) Fulfill(w http.ResponseWriter, r *http.Request) {
	var req model.FulfillProductRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	p, err := h.service.Fulfill(r.PathValue("id"), req, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, p)
}

// PUT /api/product-procurements/{id}/settle — المحاسب يرجّع المبلغ للدوار
func (h *ProductProcurementHandler) Settle(w http.ResponseWriter, r *http.Request) {
	var req model.SettleProcurementRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	p, err := h.service.Settle(r.PathValue("id"), req, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, p)
}
