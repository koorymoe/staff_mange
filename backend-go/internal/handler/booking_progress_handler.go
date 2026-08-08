package handler

import (
	"net/http"
	"strings"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/timeutil"
)

// BookingProgressHandler الإنجاز الجزئي — الحجز الي ياخذ أكثر من يوم.
type BookingProgressHandler struct {
	repo     *repository.BookingProgressRepository
	bookings *repository.BookingRepository
	notify   *repository.NotificationRepository
}

func NewBookingProgressHandler(
	repo *repository.BookingProgressRepository,
	bookings *repository.BookingRepository,
	notify *repository.NotificationRepository,
) *BookingProgressHandler {
	return &BookingProgressHandler{repo: repo, bookings: bookings, notify: notify}
}

// POST /api/bookings/{id}/partial-complete — «خلصنا جزء والباقي باچر»
func (h *BookingProgressHandler) PartialComplete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req model.PartialCompleteRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صالحة")
		return
	}
	// الشرط الوحيد المتشدد بهاي الميزة، وهو سبب وجودها كله: بلا «شنو
	// انخلص» و«شنو باقي» التقرير ما ينفع الكادر الجاي بشي.
	if TextLen(req.WorkDone) < 3 {
		WriteError(w, http.StatusBadRequest, "لازم تكتب شنو انخلص اليوم")
		return
	}
	if TextLen(req.RemainingWork) < 3 {
		WriteError(w, http.StatusBadRequest, "لازم تكتب شنو باقي — هذا الي يعتمد عليه الكادر الجاي")
		return
	}
	req.WorkDone = strings.TrimSpace(req.WorkDone)
	req.RemainingWork = strings.TrimSpace(req.RemainingWork)

	report, err := h.repo.PartialComplete(id, middleware.EmployeeIDFromContext(r), req)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر تسجيل الإنجاز الجزئي")
		return
	}

	// الإداري لازم يعرف فوراً — الحجز رجعله وينتظر تنسيق ليوم جديد.
	// بدون هذا الإشعار الحجز يقعد ساكت بقائمة ومحد ينتبهله.
	if h.notify != nil {
		code := id
		if b, err := h.bookings.FindByID(id); err == nil && b != nil {
			code = b.Code
		}
		_ = h.notify.CreateForRole("HR_COORDINATOR", "booking_partial",
			"🔄 الحجز "+code+" انجز جزئياً ("+itoa(report.PercentDone)+"٪) ويحتاج تنسيق ليوم جديد")
		_ = h.notify.CreateForRole("ADMIN", "booking_partial",
			"🔄 الحجز "+code+" انجز جزئياً ("+itoa(report.PercentDone)+"٪) ويحتاج تنسيق ليوم جديد")
	}

	WriteJSON(w, http.StatusOK, report)
}

// POST /api/bookings/{id}/schedule-continuation — الإداري يحدد موعد
// الإكمال بعد ما يتفق مع الزبون.
func (h *BookingProgressHandler) ScheduleContinuation(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ScheduledAt string `json:"scheduledAt"`
	}
	if err := DecodeJSON(r, &req); err != nil || strings.TrimSpace(req.ScheduledAt) == "" {
		WriteError(w, http.StatusBadRequest, "لازم تحدد موعد الإكمال")
		return
	}
	// الوقت جاي من حقل datetime-local يعني توقيت بغداد — لازم يتحول
	// للعالمي متل باقي مواعيد النظام، وإلا يطلع متقدم ٣ ساعات.
	when := timeutil.NormalizeCompanyLocal(strings.TrimSpace(req.ScheduledAt))
	if err := h.repo.ScheduleContinuation(r.PathValue("id"), when, middleware.EmployeeIDFromContext(r)); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	b, err := h.bookings.FindByID(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "انجدول بس تعذر جلب الحجز")
		return
	}
	WriteJSON(w, http.StatusOK, b)
}

// GET /api/bookings/{id}/progress — تقارير كل الأيام. يقراها الكادر
// الجاي قبل ما يطلع، والإداري قبل ما ينسّق.
func (h *BookingProgressHandler) Reports(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.Reports(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب تقارير الإنجاز")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// GET /api/bookings/{id}/suggested-crew — الكادر الي اشتغل بالأيام
// الفائتة. اقتراح بس: الإداري إله الحق الكامل يبدّل.
func (h *BookingProgressHandler) SuggestedCrew(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.SuggestedCrew(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الكادر المقترح")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	if neg {
		s = "-" + s
	}
	return s
}
