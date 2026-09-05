package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type TechShowcaseHandler struct {
	service *service.TechShowcaseService
}

func NewTechShowcaseHandler(s *service.TechShowcaseService) *TechShowcaseHandler {
	return &TechShowcaseHandler{service: s}
}

func (h *TechShowcaseHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب معرض الأعمال")
		return
	}
	WriteJSON(w, http.StatusOK, items)
}

func (h *TechShowcaseHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateTechShowcaseItemRequest
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

func (h *TechShowcaseHandler) AddMedia(w http.ResponseWriter, r *http.Request) {
	var req model.AddTechShowcaseMediaRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	item, err := h.service.AddMedia(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, item)
}
