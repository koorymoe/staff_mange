package handler

import (
	"log"
	"net/http"
	"strings"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type BookingDeleteRequestHandler struct {
	repo   *repository.BookingDeleteRequestRepository
	notify *repository.NotificationRepository
}

func NewBookingDeleteRequestHandler(r *repository.BookingDeleteRequestRepository, n *repository.NotificationRepository) *BookingDeleteRequestHandler {
	return &BookingDeleteRequestHandler{repo: r, notify: n}
}

// POST /api/bookings/{id}/delete-request — الإداري يطلب حذف حجز
func (h *BookingDeleteRequestHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateBookingDeleteRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	// السبب إجباري: بدونه المراقب يبت على العمياني
	if strings.TrimSpace(req.Reason) == "" {
		WriteError(w, http.StatusBadRequest, "اكتب سبب الحذف — المراقب يحتاجه حتى يقرر")
		return
	}
	out, err := h.repo.Create(r.PathValue("id"), middleware.EmployeeIDFromContext(r), strings.TrimSpace(req.Reason))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if h.notify != nil {
		msg := "🗑️ طلب حذف الحجز " + out.BookingCode + " من " + out.RequestedByName + " — السبب: " + out.Reason
		_ = h.notify.CreateForRole("MONITOR", "booking_delete_request", msg)
		_ = h.notify.CreateForRole("ADMIN", "booking_delete_request", msg)
		_ = h.notify.CreateForRole("OWNER", "booking_delete_request", msg)
	}
	WriteJSON(w, http.StatusCreated, out)
}

// GET /api/booking-delete-requests?status=PENDING
func (h *BookingDeleteRequestHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.List(r.URL.Query().Get("status"))
	if err != nil {
		log.Printf("list booking delete requests: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب طلبات الحذف")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// PUT /api/booking-delete-requests/{id}/decide — المراقب أو المدير يبت
func (h *BookingDeleteRequestHandler) Decide(w http.ResponseWriter, r *http.Request) {
	var req model.DecideBookingDeleteRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	out, err := h.repo.Decide(r.PathValue("id"), req.Approve, req.Note, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, out)
}
