package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type ChecklistHandler struct {
	service *service.ChecklistService
}

func NewChecklistHandler(s *service.ChecklistService) *ChecklistHandler {
	return &ChecklistHandler{service: s}
}

func (h *ChecklistHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الكشوفات")
		return
	}
	WriteJSON(w, http.StatusOK, items)
}

func (h *ChecklistHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateChecklistRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	item, err := h.service.Create(req, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, item)
}

func (h *ChecklistHandler) AddPhotos(w http.ResponseWriter, r *http.Request) {
	var req model.AddChecklistPhotosRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	item, err := h.service.AddPhotos(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, item)
}
