package handler

import (
	"net/http"

	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type BookingHandler struct {
	service *service.BookingService
}

func NewBookingHandler(s *service.BookingService) *BookingHandler {
	return &BookingHandler{service: s}
}

// GET /api/v1/bookings?status=&customerId=
func (h *BookingHandler) List(w http.ResponseWriter, r *http.Request) {
	bookings, err := h.service.List(r.URL.Query().Get("status"), r.URL.Query().Get("customerId"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الحجوزات")
		return
	}
	WriteJSON(w, http.StatusOK, bookings)
}

// POST /api/v1/bookings
func (h *BookingHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateBookingRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	booking, err := h.service.Create(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, booking)
}

// PUT /api/v1/bookings/{id}/confirm
func (h *BookingHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	var req model.ConfirmBookingRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	booking, err := h.service.Confirm(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, booking)
}

// PUT /api/v1/bookings/{id}/details
func (h *BookingHandler) UpdateDetails(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateBookingDetailsRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	booking, err := h.service.UpdateDetails(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, booking)
}

// GET /api/v1/bookings/{id}/schedule-log
func (h *BookingHandler) ScheduleLog(w http.ResponseWriter, r *http.Request) {
	logs, err := h.service.ScheduleLog(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب سجل التعديلات")
		return
	}
	WriteJSON(w, http.StatusOK, logs)
}

// PUT /api/v1/bookings/{id}/schedule
func (h *BookingHandler) Schedule(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ScheduledAt string `json:"scheduledAt"`
		ChangedByID string `json:"changedById"`
	}
	if err := DecodeJSON(r, &body); err != nil || body.ScheduledAt == "" {
		WriteError(w, http.StatusBadRequest, "scheduledAt is required")
		return
	}
	booking, err := h.service.SetSchedule(r.PathValue("id"), body.ChangedByID, body.ScheduledAt)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, booking)
}

// PUT /api/v1/bookings/{id}/assign
func (h *BookingHandler) Assign(w http.ResponseWriter, r *http.Request) {
	var req model.AssignBookingRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	booking, err := h.service.Assign(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, booking)
}

// PUT /api/v1/bookings/{id}/supervisor
func (h *BookingHandler) Supervisor(w http.ResponseWriter, r *http.Request) {
	var body struct {
		EmployeeID *string `json:"employeeId"`
	}
	if err := DecodeJSON(r, &body); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	booking, err := h.service.SetSupervisor(r.PathValue("id"), body.EmployeeID)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, booking)
}

// PUT /api/v1/bookings/{id}/start
func (h *BookingHandler) Start(w http.ResponseWriter, r *http.Request) {
	booking, err := h.service.Start(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, booking)
}

// PUT /api/v1/bookings/{id}/complete
func (h *BookingHandler) Complete(w http.ResponseWriter, r *http.Request) {
	var req model.CompleteBookingRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	booking, err := h.service.Complete(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, booking)
}

// PUT /api/v1/bookings/{id}/verify
func (h *BookingHandler) Verify(w http.ResponseWriter, r *http.Request) {
	booking, err := h.service.Verify(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, booking)
}
