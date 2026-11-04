package handler

import (
	"net/http"

	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type SupplierHandler struct {
	service *service.SupplierService
}

func NewSupplierHandler(s *service.SupplierService) *SupplierHandler {
	return &SupplierHandler{service: s}
}

// GET /api/v1/suppliers/specialties
func (h *SupplierHandler) ListSpecialties(w http.ResponseWriter, r *http.Request) {
	specs, err := h.service.ListSpecialties()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب التخصصات")
		return
	}
	WriteJSON(w, http.StatusOK, specs)
}

// POST /api/v1/suppliers/specialties
func (h *SupplierHandler) CreateSpecialty(w http.ResponseWriter, r *http.Request) {
	var req model.CreateSupplierSpecialtyRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	spec, err := h.service.CreateSpecialty(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, spec)
}

// DELETE /api/v1/suppliers/specialties/{id}
func (h *SupplierHandler) DeleteSpecialty(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteSpecialty(r.PathValue("id")); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// GET /api/v1/suppliers
func (h *SupplierHandler) List(w http.ResponseWriter, r *http.Request) {
	suppliers, err := h.service.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الموردين")
		return
	}
	WriteJSON(w, http.StatusOK, suppliers)
}

// POST /api/v1/suppliers
func (h *SupplierHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.UpsertSupplierRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	supplier, err := h.service.Create(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, supplier)
}

// PUT /api/v1/suppliers/{id}
func (h *SupplierHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req model.UpsertSupplierRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	supplier, err := h.service.Update(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, supplier)
}

// DELETE /api/v1/suppliers/{id}
func (h *SupplierHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Delete(r.PathValue("id")); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// POST /api/v1/suppliers/{id}/rate
func (h *SupplierHandler) Rate(w http.ResponseWriter, r *http.Request) {
	var req model.RateSupplierRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	rating, err := h.service.Rate(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, rating)
}
