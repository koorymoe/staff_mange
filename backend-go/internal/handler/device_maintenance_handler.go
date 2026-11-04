package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type DeviceMaintenanceHandler struct {
	service *service.DeviceMaintenanceService
}

func NewDeviceMaintenanceHandler(s *service.DeviceMaintenanceService) *DeviceMaintenanceHandler {
	return &DeviceMaintenanceHandler{service: s}
}

func (h *DeviceMaintenanceHandler) List(w http.ResponseWriter, r *http.Request) {
	tickets, err := h.service.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب تذاكر الصيانة")
		return
	}
	WriteJSON(w, http.StatusOK, tickets)
}

func (h *DeviceMaintenanceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateDeviceMaintenanceTicketRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	ticket, err := h.service.Create(middleware.EmployeeIDFromContext(r), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, ticket)
}

func (h *DeviceMaintenanceHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateDeviceMaintenanceTicketRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	ticket, err := h.service.Update(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, ticket)
}
