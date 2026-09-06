package handler

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// CoordinationAlertHandler تنبيهات تقصير الإداري بتثبيت الحجز.
//
// ⚠️ الإعلان ينشر **مرة وحدة بالضبط** — عند العاشر (`== Threshold`
// مو `>=`) حتى ما يتكرر بضغطة حادية عشرة.
type CoordinationAlertHandler struct {
	repo          *repository.CoordinationAlertRepository
	bookings      *repository.BookingRepository
	discipline    *repository.DisciplineRepository
	announcements *repository.AnnouncementRepository
	notify        *repository.NotificationRepository
	employees     *repository.EmployeeRepository
}

func NewCoordinationAlertHandler(
	repo *repository.CoordinationAlertRepository,
	bookings *repository.BookingRepository,
	discipline *repository.DisciplineRepository,
	announcements *repository.AnnouncementRepository,
	notify *repository.NotificationRepository,
	employees *repository.EmployeeRepository,
) *CoordinationAlertHandler {
	return &CoordinationAlertHandler{
		repo: repo, bookings: bookings, discipline: discipline,
		announcements: announcements, notify: notify, employees: employees,
	}
}

// byNameFor يرجّع اسم الموظف المسجَّل بالنداء — يُكتب نصاً بالسجل
// حتى لو انحذف الموظف لاحقاً يبقى مقروءاً (نفس نمط سجل الشكاوى).
func (h *CoordinationAlertHandler) byNameFor(id string) string {
	if h.employees == nil {
		return ""
	}
	e, err := h.employees.FindByID(id)
	if err != nil || e == nil {
		return ""
	}
	return e.Name
}

// POST /api/bookings/{id}/coordination-alerts — «تسجيل تقصير»
func (h *CoordinationAlertHandler) Add(w http.ResponseWriter, r *http.Request) {
	bookingID := r.PathValue("id")
	var req model.AddCoordinationAlertRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}

	booking, err := h.bookings.FindByID(bookingID)
	if err != nil || booking == nil {
		WriteError(w, http.StatusNotFound, "الحجز مو موجود")
		return
	}

	var coordinatorID, coordinatorName *string
	if booking.ConfirmedByEmployee != nil {
		coordinatorID = &booking.ConfirmedByEmployee.ID
		coordinatorName = &booking.ConfirmedByEmployee.Name
	}

	byID := middleware.EmployeeIDFromContext(r)
	byName := h.byNameFor(byID)
	openCount, err := h.repo.Add(bookingID, coordinatorID, coordinatorName, strings.TrimSpace(req.Reason), byID, byName)
	if err != nil {
		log.Printf("add coordination alert: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر تسجيل التقصير")
		return
	}

	// ⚠️ `==` لا `>=` — النشر مرة وحدة بالضبط عند العاشر بالضبط.
	if openCount == model.CoordinationAlertThreshold {
		h.announceThreshold(booking, coordinatorName, openCount)
	} else if h.notify != nil && coordinatorID != nil {
		_ = h.notify.Create(*coordinatorID, "coordination_alert",
			"⚠️ تسجّل عليك تقصير بتثبيت الحجز "+booking.Code+" ("+strconv.Itoa(openCount)+"/"+
				strconv.Itoa(model.CoordinationAlertThreshold)+")")
	}

	WriteJSON(w, http.StatusOK, map[string]any{"openCount": openCount})
}

func (h *CoordinationAlertHandler) announceThreshold(booking *model.Booking, coordinatorName *string, count int) {
	if h.announcements == nil || h.discipline == nil {
		return
	}
	who := "الإداري المسؤول"
	if coordinatorName != nil && *coordinatorName != "" {
		who = *coordinatorName
	}
	author, err := h.discipline.SystemAuthorID()
	if err != nil || author == "" {
		log.Printf("[coordination-alert] ماكو حساب يصلح لنشر الإعلان: %v", err)
		return
	}
	body := "⚠️ تنبيه تنسيق: " + who + " وصلت تنبيهاته على الحجز " + booking.Code +
		" إلى " + strconv.Itoa(count) + "/" + strconv.Itoa(model.CoordinationAlertThreshold) + " بدون تثبيت."
	if _, err := h.announcements.Create(body, author, 7); err != nil {
		log.Printf("[coordination-alert] تعذر نشر الإعلان: %v", err)
	}
}

// PUT /api/bookings/{id}/coordination-alerts/resolve — «تمت المعالجة»
func (h *CoordinationAlertHandler) Resolve(w http.ResponseWriter, r *http.Request) {
	bookingID := r.PathValue("id")
	var req model.ResolveCoordinationAlertsRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	byID := middleware.EmployeeIDFromContext(r)
	byName := h.byNameFor(byID)
	if err := h.repo.Resolve(bookingID, byID, byName, strings.TrimSpace(req.Note)); err != nil {
		log.Printf("resolve coordination alerts: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر تعليم المعالجة")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// GET /api/bookings/{id}/coordination-alerts — «عرض السجل»
func (h *CoordinationAlertHandler) ListForBooking(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.ListForBooking(r.PathValue("id"))
	if err != nil {
		log.Printf("list coordination alerts: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب السجل")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// GET /api/coordination-alerts/summaries — ملخص كل الحجوزات دفعة وحدة
func (h *CoordinationAlertHandler) Summaries(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.Summaries()
	if err != nil {
		log.Printf("coordination alert summaries: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الملخص")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}
