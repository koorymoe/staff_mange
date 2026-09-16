package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type VehicleBookingHandler struct {
	service *service.VehicleBookingService
}

func NewVehicleBookingHandler(s *service.VehicleBookingService) *VehicleBookingHandler {
	return &VehicleBookingHandler{service: s}
}

func (h *VehicleBookingHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateVehicleBookingRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	actorID := middleware.EmployeeIDFromContext(r)
	booking, err := h.service.CreateBooking(actorID, req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, booking)
}

func (h *VehicleBookingHandler) Decide(w http.ResponseWriter, r *http.Request) {
	var req model.DecideVehicleBookingRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	actorID := middleware.EmployeeIDFromContext(r)
	booking, err := h.service.DecideBooking(r.PathValue("id"), actorID, req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, booking)
}

func (h *VehicleBookingHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.EmployeeIDFromContext(r)
	isAdmin := isAdminOrMonitor(middleware.RoleFromContext(r))
	booking, err := h.service.CancelBooking(r.PathValue("id"), actorID, isAdmin)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, booking)
}

func (h *VehicleBookingHandler) List(w http.ResponseWriter, r *http.Request) {
	filters := model.VehicleBookingFilters{}
	q := r.URL.Query()
	if v := q.Get("vehicleId"); v != "" {
		filters.VehicleID = &v
	}
	if v := q.Get("requestedById"); v != "" {
		filters.RequestedByID = &v
	}
	if v := q.Get("status"); v != "" {
		filters.Status = &v
	}
	if v := q.Get("from"); v != "" {
		filters.From = &v
	}
	if v := q.Get("to"); v != "" {
		filters.To = &v
	}
	bookings, err := h.service.List(filters)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب قائمة الحجوزات")
		return
	}
	WriteJSON(w, http.StatusOK, bookings)
}
