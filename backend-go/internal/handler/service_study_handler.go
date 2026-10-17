package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type ServiceStudyHandler struct {
	service *service.ServiceStudyService
}

func NewServiceStudyHandler(s *service.ServiceStudyService) *ServiceStudyHandler {
	return &ServiceStudyHandler{service: s}
}

func (h *ServiceStudyHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الخدمات")
		return
	}
	WriteJSON(w, http.StatusOK, items)
}

func (h *ServiceStudyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateServiceStudyRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	item, err := h.service.Create(req.Name, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, item)
}

func (h *ServiceStudyHandler) Assign(w http.ResponseWriter, r *http.Request) {
	var req model.AssignServiceStudyRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	item, err := h.service.Assign(r.PathValue("id"), req.EmployeeIDs)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر حفظ التوكيل")
		return
	}
	WriteJSON(w, http.StatusOK, item)
}

func (h *ServiceStudyHandler) AddReport(w http.ResponseWriter, r *http.Request) {
	var req model.CreateServiceStudyReportRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	item, err := h.service.AddReport(r.PathValue("id"), middleware.EmployeeIDFromContext(r), req.Content)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, item)
}

func (h *ServiceStudyHandler) Archive(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.Archive(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر الأرشفة")
		return
	}
	WriteJSON(w, http.StatusOK, item)
}
