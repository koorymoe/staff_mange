package handler

import (
	"net/http"

	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type ProcurementHandler struct {
	service *service.ProcurementService
}

func NewProcurementHandler(s *service.ProcurementService) *ProcurementHandler {
	return &ProcurementHandler{service: s}
}

// GET /api/v1/procurement
func (h *ProcurementHandler) List(w http.ResponseWriter, r *http.Request) {
	requests, err := h.service.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب طلبات الشراء")
		return
	}
	WriteJSON(w, http.StatusOK, requests)
}

// GET /api/v1/procurement/stats
func (h *ProcurementHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.Stats()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الإحصائيات")
		return
	}
	WriteJSON(w, http.StatusOK, stats)
}

// POST /api/v1/procurement
func (h *ProcurementHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateProcurementRequestRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	request, err := h.service.Create(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, request)
}

// PUT /api/v1/procurement/{id}/status
func (h *ProcurementHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateProcurementStatusRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	request, err := h.service.UpdateStatus(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, request)
}

// PUT /api/v1/procurement/{id}/fulfill
func (h *ProcurementHandler) Fulfill(w http.ResponseWriter, r *http.Request) {
	var req model.FulfillProcurementRequestRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	request, err := h.service.Fulfill(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, request)
}
