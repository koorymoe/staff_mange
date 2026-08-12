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
	invoices *repository.LeaderInvoiceRepository
}

func NewBookingAuditHandler(r *repository.BookingAuditRepository, b *repository.BookingRepository, n *repository.NotificationRepository, inv *repository.LeaderInvoiceRepository) *BookingAuditHandler {
	return &BookingAuditHandler{repo: r, bookings: b, notify: n, invoices: inv}
}

// stampInvoiceVerdict ينزّل حكم التدقيق اليومي على فاتورة الحجز.
//
// طلب صاحب العمل: «الفواتير أطابقهن، من أطابقهن يروحون وين؟ فواتير
// بحاجة لاعتماد». قبل هذا، التدقيق اليومي جان يشتغل على الحجز بس
// والفاتورة تضل بلا حكم — فالمحاسب يدقق ٢٠ حجز وقائمة الفواتير ما
// تتحرك ولا خطوة، ويرجع يحكم عليهن وحدة وحدة من جديد.
//
// ⚠️ ما يعدّل فاتورة معتمدة (SetAuditVerdict ترفض) ولا يفشّل التدقيق
// إذا ماكو فاتورة أصلاً — أكو حجوزات تنتدقق بتقدير الإداري بلا فاتورة.
func (h *BookingAuditHandler) stampInvoiceVerdict(bookingID, verdict string, note *string, empID string, amount *float64) {
	if h.invoices == nil {
		return
	}
	rows, err := h.invoices.ListByBooking(bookingID)
	if err != nil || len(rows) == 0 {
		return
	}
	latest := rows[len(rows)-1] // ListByBooking مرتبة تصاعدياً بالإنشاء
	noteStr := ""
	if note != nil {
		noteStr = *note
	}
	if _, err := h.invoices.SetAuditVerdict(latest.ID, verdict, noteStr, empID, amount); err != nil {
		log.Printf("stamp invoice verdict (booking %s): %v", bookingID, err)
	}
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
		// «مطابق» يرحّل الفاتورة لطابور الاعتماد بحكمها مثبّت.
		h.stampInvoiceVerdict(bookingID, model.AuditVerdictMatched, req.Note, empID, req.AmountCollected)

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
		// نفس الشي للأحكام السلبية: الفاتورة تحمل سبب رفضها، فالمراقب
		// يفتحها ويلگه ليش انتأشرت بلا ما يدور بالبلاغات.
		{
			v := model.AuditVerdictMismatch
			if req.Action == model.AuditPriceError {
				v = model.AuditVerdictPriceError
			}
			h.stampInvoiceVerdict(bookingID, v, req.Note, empID, req.AmountCollected)
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
	// ═══ المحاسب مو مراقب ═══
	// «أخطاء الفواتير تظهر إله كمحاسب — هو أرسلهن. بس تظهر للمراقب
	// كتدقيق حتى يتأكد من الليدر ليش عنده أخطاء.»
	// المحاسب يشوف **صادره** هو بس؛ المراقب والمدير يشوفون الكل مع
	// اسم الليدر. بدون هذا الاثنين جانوا يشوفون نفس القائمة بالضبط،
	// فالمحاسب يحسبها شغل عليه وهو أصلاً الي سجّلها.
	raisedBy := ""
	if role == "FINANCE" {
		raisedBy = middleware.EmployeeIDFromContext(r)
	}
	rows, err := h.repo.List(strings.TrimSpace(r.URL.Query().Get("status")), kinds, raisedBy)
	if err != nil {
		log.Printf("list audit issues: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب البلاغات")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// PUT /api/audit-issues/{id}/resolve
func (h *BookingAuditHandler) ResolveIssue(w http.ResponseWriter, r *http.Request) {
	// البلاغ ينغلق من الجهة الي انوجّهله بس. بدون هالفحص أي موظف مسجّل
	// دخول يقدر يغلق البلاغ — يعني نفس الي سبّب خطأ السعر يغلق البلاغ
	// المرفوع ضده، وكل فايدة التوجيه تروح.
	kind, err := h.repo.KindOf(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusNotFound, "البلاغ غير موجود")
		return
	}
	role := middleware.RoleFromContext(r)
	allowed := role == "ADMIN" || role == "OWNER"
	for _, rt := range model.AuditRoutedRoles(kind) {
		if rt == role {
			allowed = true
			break
		}
	}
	if !allowed {
		WriteError(w, http.StatusForbidden, "هذا البلاغ ما ينغلق إلا من "+model.AuditRoutedLabel(kind))
		return
	}
	if err := h.repo.Resolve(r.PathValue("id")); err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر إغلاق البلاغ")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
