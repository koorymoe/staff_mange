package handler

import (
	"net/http"

	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type ServiceHandler struct {
	service *service.ServiceCatalogService
}

func NewServiceHandler(s *service.ServiceCatalogService) *ServiceHandler {
	return &ServiceHandler{service: s}
}

// GET /api/v1/services
func (h *ServiceHandler) List(w http.ResponseWriter, r *http.Request) {
	services, err := h.service.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب قائمة الخدمات")
		return
	}
	WriteJSON(w, http.StatusOK, services)
}

// POST /api/v1/services
func (h *ServiceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateServiceRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}

	svc, err := h.service.Create(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, svc)
}
