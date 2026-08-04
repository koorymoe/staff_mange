package handler

import (
	"log"
	"net/http"
	"strings"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type BookingAuditHandler struct {
	repo     *repository.BookingAuditRepository
	bookings *repository.BookingRepository
	notify   *repository.NotificationRepository
}

func NewBookingAuditHandler(r *repository.BookingAuditRepository, b *repository.BookingRepository, n *repository.NotificationRepository) *BookingAuditHandler {
	return &BookingAuditHandler{repo: r, bookings: b, notify: n}
}

// PUT /api/bookings/{id}/audit — قرار المحاسب.
//
// إما «مطابق» (ويأشر مدقق، ويشترط مبلغ)، أو بلاغ خطأ ينوجّه للمعني.
func (h *BookingAuditHandler) Audit(w http.ResponseWriter, r *http.Request) {
	var req model.AuditBookingRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	bookingID := r.PathValue("id")
	empID := middleware.EmployeeIDFromContext(r)

	switch req.Action {
	case model.AuditVerify:
		if err := h.repo.Verify(bookingID, req.AmountCollected, req.AdvancePaid); err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

	case model.AuditMismatch, model.AuditPriceError:
		// المبلغ الي كتبه المحاسب ينحفظ حتى لو أشّر خطأ — هو الرقم
		// الصحيح من الفاتورة، والفرق يبين للي راح يتابع.
		if req.AmountCollected != nil || req.AdvancePaid != nil {
			_ = h.repo.SetAmount(bookingID, req.AmountCollected, req.AdvancePaid)
		}
		b, err := h.bookings.FindByID(bookingID)
		if err != nil || b == nil {
			WriteError(w, http.StatusNotFound, "الحجز غير موجود")
			return
		}
		issue, err := h.repo.RaiseIssue(bookingID, req.Action, req.Note, req.AmountCollected, b.AmountCollected, empID)
		if err != nil {
			log.Printf("raise audit issue: %v", err)
			WriteError(w, http.StatusBadRequest, "تعذر تسجيل البلاغ")
			return
		}
		if h.notify != nil {
			msg := "💸 " + issue.KindLabel + " بالحجز " + issue.BookingCode +
				" — سجّله " + issue.RaisedByName
			if issue.Note != nil && *issue.Note != "" {
				msg += ": " + *issue.Note
			}
			for _, role := range model.AuditRoutedRoles(req.Action) {
				_ = h.notify.CreateForRole(role, "audit_issue", msg)
			}
			_ = h.notify.CreateForRole("OWNER", "audit_issue", msg)
		}
		WriteJSON(w, http.StatusCreated, issue)
		return

	default:
		WriteError(w, http.StatusBadRequest, "نوع القرار غير معروف")
		return
	}

	updated, err := h.bookings.FindByID(bookingID)
	if err != nil || updated == nil {
		WriteError(w, http.StatusNotFound, "الحجز غير موجود")
		return
	}
	WriteJSON(w, http.StatusOK, updated)
}

// GET /api/audit-issues?status=OPEN — بلاغات الأخطاء.
//
// كل واحد يشوف الي يخصه: الجودة تشوف «غير مطابق»، والإداري يشوف
// «خطأ سعر»، والرقابة والمالك يشوفون الاثنين.
func (h *BookingAuditHandler) ListIssues(w http.ResponseWriter, r *http.Request) {
	role := middleware.RoleFromContext(r)
	var kinds []string
	switch role {
	case "QUALITY_ENGINEER":
		kinds = []string{model.AuditMismatch}
	case "HR_COORDINATOR":
		kinds = []string{model.AuditPriceError}
	}
	rows, err := h.repo.List(strings.TrimSpace(r.URL.Query().Get("status")), kinds)
	if err != nil {
		log.Printf("list audit issues: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب البلاغات")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// PUT /api/audit-issues/{id}/resolve
func (h *BookingAuditHandler) ResolveIssue(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.Resolve(r.PathValue("id")); err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر إغلاق البلاغ")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
