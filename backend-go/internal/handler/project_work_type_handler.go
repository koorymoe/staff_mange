package handler

import (
	"net/http"

	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type ProjectWorkTypeHandler struct {
	service *service.ProjectWorkTypeService
}

func NewProjectWorkTypeHandler(s *service.ProjectWorkTypeService) *ProjectWorkTypeHandler {
	return &ProjectWorkTypeHandler{service: s}
}

func (h *ProjectWorkTypeHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب أنواع الأعمال")
		return
	}
	WriteJSON(w, http.StatusOK, items)
}

func (h *ProjectWorkTypeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateProjectWorkTypeRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	item, err := h.service.Create(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, item)
}

func (h *ProjectWorkTypeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Delete(r.PathValue("id")); err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر حذف نوع العمل")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
