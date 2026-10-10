package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type AttendanceIconRequestHandler struct {
	service *service.AttendanceIconRequestService
}

func NewAttendanceIconRequestHandler(s *service.AttendanceIconRequestService) *AttendanceIconRequestHandler {
	return &AttendanceIconRequestHandler{service: s}
}

func (h *AttendanceIconRequestHandler) ListPending(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListPending()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الطلبات")
		return
	}
	WriteJSON(w, http.StatusOK, items)
}

func (h *AttendanceIconRequestHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateAttendanceIconRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	item, err := h.service.Create(middleware.EmployeeIDFromContext(r), req.RequestedIcon)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, item)
}

func (h *AttendanceIconRequestHandler) Approve(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Approve(r.PathValue("id"), middleware.EmployeeIDFromContext(r)); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AttendanceIconRequestHandler) Reject(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Reject(r.PathValue("id"), middleware.EmployeeIDFromContext(r)); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
