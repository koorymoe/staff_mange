package handler

import (
	"net/http"

	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type ComplaintHandler struct {
	service *service.ComplaintService
}

func NewComplaintHandler(s *service.ComplaintService) *ComplaintHandler {
	return &ComplaintHandler{service: s}
}

// GET /api/v1/complaints
func (h *ComplaintHandler) List(w http.ResponseWriter, r *http.Request) {
	complaints, err := h.service.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الشكاوى")
		return
	}
	WriteJSON(w, http.StatusOK, complaints)
}

// POST /api/v1/complaints
func (h *ComplaintHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateComplaintRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	complaint, err := h.service.Create(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, complaint)
}

// PUT /api/v1/complaints/{id}
func (h *ComplaintHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateComplaintRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	complaint, err := h.service.Update(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, complaint)
}

// PUT /api/v1/complaints/{id}/resolve
func (h *ComplaintHandler) Resolve(w http.ResponseWriter, r *http.Request) {
	var req model.ResolveComplaintRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	complaint, err := h.service.Resolve(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, complaint)
}
