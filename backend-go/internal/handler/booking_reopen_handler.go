package handler

import (
	"errors"
	"net/http"
	"strings"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/repository"
)

// PUT /api/bookings/{id}/reopen {reason} — صلاحية booking_reopen (أو المدير).
// حجز انضغط عليه «تم الإنجاز» بالغلط يرجع «قيد العمل» للكادر، حتى يسجّل
// «إنجاز جزئي» من مهامي. فاتورة الليدر (إن وجدت) ما تنحذف — ننبّه بس.
type BookingReopenHandler struct {
	repo  *repository.BookingReopenRepository
	emps  *repository.EmployeeRepository
	notif *repository.NotificationRepository
}

func NewBookingReopenHandler(r *repository.BookingReopenRepository, e *repository.EmployeeRepository, n *repository.NotificationRepository) *BookingReopenHandler {
	return &BookingReopenHandler{repo: r, emps: e, notif: n}
}

func (h *BookingReopenHandler) Reopen(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Reason string `json:"reason"`
	}
	if err := DecodeJSON(r, &b); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	reason := strings.TrimSpace(b.Reason)
	if TextLen(reason) < 3 {
		WriteError(w, http.StatusBadRequest, "اكتب سبب الإرجاع")
		return
	}
	byName := "—"
	if e, err := h.emps.FindByID(middleware.EmployeeIDFromContext(r)); err == nil && e != nil {
		byName = e.Name
	}
	res, err := h.repo.Reopen(r.PathValue("id"), byName, reason)
	if errors.Is(err, repository.ErrNotCompleted) {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	for _, id := range res.CrewIDs {
		_ = h.notif.Create(id, "booking_returned_to_crew", "↩️ الحجز "+res.Code+" رجعلك «قيد العمل» — سجّل عليه «إنجاز جزئي» من مهامي. السبب: "+reason)
	}
	warn := ""
	if res.HasInvoice {
		warn = "⚠️ الحجز عليه فاتورة ليدر — ما انحذفت، بلّغ المحاسب يراجعها."
	}
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "code": res.Code, "notified": len(res.CrewIDs), "warning": warn})
}
