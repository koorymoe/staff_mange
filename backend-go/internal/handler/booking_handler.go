package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/service"
)

type BookingHandler struct {
	service     *service.BookingService
	permissions *repository.PermissionRepository
	// reminders اختياري — يُربط من main للتشغيل اليدوي للكنسة
	reminders *service.BookingReminderService
	// timeline اختياري — بدونه المسار يرجّع 503 بدل ما يطيح
	timeline *service.BookingTimelineService
}

// SetTimelineService يربط خدمة الخط الزمني بعد البناء.
func (h *BookingHandler) SetTimelineService(t *service.BookingTimelineService) { h.timeline = t }

func NewBookingHandler(s *service.BookingService, p *repository.PermissionRepository) *BookingHandler {
	return &BookingHandler{service: s, permissions: p}
}

// SetReminderService يربط خدمة التذكير بعد البناء.
func (h *BookingHandler) SetReminderService(r *service.BookingReminderService) { h.reminders = r }

// GET /api/v1/bookings?status=&customerId=&assignedTo=me
//
// الفني يريد مهامه هو — كانت الواجهة تنزّل كل حجوزات الشركة (بكل
// تعييناتها وزبائنها وخدماتها) وتفلترها بالمتصفح حتى تلكه سبع مهام.
// يعني ميغابايتات تمشي بالشبكة كل ما يفتح الصفحة، والتأخير يكبر كل ما
// تتراكم الحجوزات. هسه الفلترة بقاعدة البيانات: JOIN على التعيين
// ويرجع حجوزاته هو بس.
//
// وهو قيد أمان بعد: الفني ما عاد يوصل لبيانات زبائن الشركة كلها من
// مسار عام، حتى لو الواجهة طلبتها.
func (h *BookingHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("assignedTo") == "me" {
		bookings, err := h.service.ListAssignedTo(middleware.EmployeeIDFromContext(r))
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "تعذر جلب مهامك")
			return
		}
		WriteJSON(w, http.StatusOK, bookings)
		return
	}
	// limit اختياري: الشاشة الي تعرض «آخر كذا حجز» تطلب عددها بس، بدل ما
	// يمشي أرشيف الشركة كله بالشبكة كل مرة تنفتح.
	limit := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	bookings, err := h.service.List(r.URL.Query().Get("status"), r.URL.Query().Get("customerId"), r.URL.Query().Get("date"), limit)
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
	// حجز داخل الشركة للإداري فما فوق حصراً — موظف المبيعات ما يقدر
	// يسجّل شغل داخلي، لأنه شغل شركة مو بيع لزبون.
	if req.BookingType != nil && *req.BookingType == "INTERNAL" {
		role := middleware.RoleFromContext(r)
		allowed := role == "ADMIN" || role == "OWNER" || role == "HR_COORDINATOR" || role == "MONITOR"
		if !allowed && h.permissions != nil {
			if ok, err := h.permissions.HasPermission(middleware.EmployeeIDFromContext(r), "booking_internal"); err == nil && ok {
				allowed = true
			}
		}
		if !allowed {
			WriteError(w, http.StatusForbidden, "حجز داخل الشركة يسجّله الإداري فما فوق")
			return
		}
	}
	booking, err := h.service.Create(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	// منو أدخل الحجز — «بانتظار التثبيت» كانت تعرضه بلا ما تگول
	// لمنو ترجع لو المعلومة ناقصة.
	h.service.RecordCreator(booking.ID, middleware.EmployeeIDFromContext(r))
	if fresh, err := h.service.Get(booking.ID); err == nil && fresh != nil {
		booking = fresh
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
	booking, err := h.service.UpdateDetails(r.PathValue("id"), req, middleware.EmployeeIDFromContext(r))
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
		// ChangedByID ما ننقراه من الطلب — كان يخلي أي واحد ينسب تغيير
		// الموعد لموظف ثاني بسجل التعديلات (تزوير هوية). نعتمد التوكن.
		ChangedByID string `json:"-"`
	}
	if err := DecodeJSON(r, &body); err != nil || body.ScheduledAt == "" {
		WriteError(w, http.StatusBadRequest, "scheduledAt is required")
		return
	}
	booking, err := h.service.SetSchedule(r.PathValue("id"), middleware.EmployeeIDFromContext(r), body.ScheduledAt)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, booking)
}

// PUT /api/v1/bookings/{id}/assign
// Unassign يشيل موظف من خانة كادر بالحجز.
func (h *BookingHandler) Unassign(w http.ResponseWriter, r *http.Request) {
	booking, err := h.service.Unassign(r.PathValue("id"), r.URL.Query().Get("role"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, booking)
}

func (h *BookingHandler) Assign(w http.ResponseWriter, r *http.Request) {
	var req model.AssignBookingRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	booking, err := h.service.Assign(r.PathValue("id"), req, middleware.EmployeeIDFromContext(r))
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
// يقبل جسم اختياري { missingToolIds: string[] } — لو مُرسل، تُسجَّل لقطة الأدوات
// الشخصية الناقصة عند الموظف بلحظة الاستلام قبل ما نكمل نفس منطق الاستلام العادي.
func (h *BookingHandler) Start(w http.ResponseWriter, r *http.Request) {
	var req model.AcceptBookingRequest
	_ = DecodeJSON(r, &req) // الجسم اختياري بالكامل، تجاهل خطأ فك الترميز لو الطلب بلا جسم
	booking, err := h.service.StartWithToolsCheck(r.PathValue("id"), middleware.EmployeeIDFromContext(r), req.MissingToolIDs)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, booking)
}

// PUT /api/bookings/{id}/confirmation-contacted
// يسجّل "تم" الإداري بعد تواصله فعلياً مع الزبون — خطوة سابقة ومنفصلة عن التثبيت
// (Confirm) نفسه، يستخدمها المراقب (صلاحية crew_management) للتدقيق قبل التثبيت.
// ReturnToCrew يرجّع حجز محوّل لإدارة المشاريع رجعة لكادر الشد.
func (h *BookingHandler) ReturnToCrew(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Note *string `json:"note"`
	}
	_ = DecodeJSON(r, &req)
	// السبب إجباري: إداري الكوادر لازم يعرف ليش رجع له الحجز
	if req.Note == nil || strings.TrimSpace(*req.Note) == "" {
		WriteError(w, http.StatusBadRequest, "لازم تكتب سبب إرجاع الحجز لكادر الشد")
		return
	}
	booking, err := h.service.ReturnToCrew(r.PathValue("id"), req.Note)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر إرجاع الحجز لكادر الشد")
		return
	}
	WriteJSON(w, http.StatusOK, booking)
}

func (h *BookingHandler) MarkConfirmationContacted(w http.ResponseWriter, r *http.Request) {
	booking, err := h.service.MarkConfirmationContacted(r.PathValue("id"), middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, booking)
}

// PUT /api/bookings/{id}/arrived
func (h *BookingHandler) MarkArrived(w http.ResponseWriter, r *http.Request) {
	booking, err := h.service.MarkArrived(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, booking)
}

// PUT /api/bookings/{id}/unverify — يرجّع الحجز للتدقيق (مدير النظام)
func (h *BookingHandler) Unverify(w http.ResponseWriter, r *http.Request) {
	booking, err := h.service.Unverify(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, booking)
}

// PUT /api/bookings/{id}/stop-work
func (h *BookingHandler) StopWork(w http.ResponseWriter, r *http.Request) {
	var req model.StopWorkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "طلب غير صالح")
		return
	}
	booking, err := h.service.StopWork(r.PathValue("id"), req.Reason, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, booking)
}

// PUT /api/bookings/{id}/resume-work
func (h *BookingHandler) ResumeWork(w http.ResponseWriter, r *http.Request) {
	booking, err := h.service.ResumeWork(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, booking)
}

// PUT /api/bookings/{id}/materials-ready
func (h *BookingHandler) SetMaterialsReady(w http.ResponseWriter, r *http.Request) {
	booking, err := h.service.SetMaterialsReady(r.PathValue("id"), middleware.EmployeeIDFromContext(r))
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

// GET /api/bookings/pending-audit
// نفس منطق List بحالة PENDING بالضبط، لكن بمسار مستقل يُحمى بصلاحية crew_management
// (يستخدمه المراقب لتدقيق الحجوزات الموجّهة قبل ما يثبّتها الإداري — يشوف فيها هل
// الإداري ضغط "تم" (تواصل مع الزبون) قبل التثبيت الفعلي أو لا، عبر حقل
// confirmationContactedAt).
func (h *BookingHandler) PendingAudit(w http.ResponseWriter, r *http.Request) {
	bookings, err := h.service.List("PENDING", "", "", 0)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الحجوزات")
		return
	}
	WriteJSON(w, http.StatusOK, bookings)
}

// GET /api/bookings/{id}/tool-checks
// لقطات الأدوات الناقصة المسجّلة عند استلام هذا الحجز (Feature 2) — يستخدمها
// الإداري/المراقب لمراجعة شنو كان ناقص عند كل موظف استلم هذا الحجز.
func (h *BookingHandler) ToolChecks(w http.ResponseWriter, r *http.Request) {
	checks, err := h.service.ListToolChecks(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب لقطات الأدوات")
		return
	}
	WriteJSON(w, http.StatusOK, checks)
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

// ═══ الأرشيف والتأجيل والانتظار ═══

// GET /api/bookings/archived
func (h *BookingHandler) ListArchived(w http.ResponseWriter, r *http.Request) {
	limit := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	rows, err := h.service.ListArchived(limit)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب أرشيف الحجوزات")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// DELETE /api/bookings/{id} — حذف من الشاشات، حفظ بالأرشيف
func (h *BookingHandler) ArchiveBooking(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Reason string `json:"reason"`
	}
	_ = DecodeJSON(r, &body)
	b, err := h.service.Archive(r.PathValue("id"), middleware.EmployeeIDFromContext(r), body.Reason)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, b)
}

// PUT /api/bookings/{id}/restore
func (h *BookingHandler) RestoreBooking(w http.ResponseWriter, r *http.Request) {
	b, err := h.service.Restore(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, b)
}

// PUT /api/bookings/{id}/postpone
func (h *BookingHandler) Postpone(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ScheduledAt string `json:"scheduledAt"`
		Reason      string `json:"reason"`
	}
	if err := DecodeJSON(r, &body); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات التأجيل غير صحيحة")
		return
	}
	b, err := h.service.Postpone(r.PathValue("id"), body.ScheduledAt, body.Reason, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, b)
}

// POST /api/bookings/waiting-reminder-sweep — تشغيل يدوي للكنسة
// (للفحص وللإدارة)، نفس أسلوب كنسة الانضباط.
func (h *BookingHandler) RunWaitingReminderSweep(w http.ResponseWriter, r *http.Request) {
	if h.reminders == nil {
		WriteError(w, http.StatusServiceUnavailable, "خدمة التذكير مو مربوطة")
		return
	}
	n, err := h.reminders.RunWaitingReminderSweep()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]int{"reminded": n})
}

// GET /api/bookings/postponed — المؤجلة بلا موعد، طابور قرارات الإداري
func (h *BookingHandler) ListPostponed(w http.ResponseWriter, r *http.Request) {
	rows, err := h.service.ListPostponed()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الحجوزات المؤجلة")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// PUT /api/bookings/{id}/waiting — الزبون ما رد
func (h *BookingHandler) MarkWaiting(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Note string `json:"note"`
	}
	_ = DecodeJSON(r, &body)
	b, err := h.service.MarkWaiting(r.PathValue("id"), body.Note, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, b)
}

// PUT /api/bookings/{id}/resume — الزبون رد، يرجع للطابور
func (h *BookingHandler) ResumeFromWaiting(w http.ResponseWriter, r *http.Request) {
	b, err := h.service.ResumeFromWaiting(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, b)
}

// PUT /api/bookings/{id}/type — تغيير نوع الحجز.
//
// للمالك ومدير النظام حصراً (مقيّد بالراوتر): نوع الحجز يأثر على
// الإحصاءات والعمولات وحساب الصيانة، فمو قرار إداري يومي.
func (h *BookingHandler) ChangeType(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BookingType string `json:"bookingType"`
	}
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صالحة")
		return
	}
	allowed := map[string]bool{"REGULAR": true, "MAINTENANCE": true, "INTERNAL": true, "SOLAR": true}
	if !allowed[req.BookingType] {
		WriteError(w, http.StatusBadRequest, "نوع حجز غير معروف")
		return
	}
	booking, err := h.service.ChangeType(r.PathValue("id"), req.BookingType, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, booking)
}

// ═══ تتبّع المراحل ═══

// PUT /api/bookings/{id}/crew-notes — ملاحظة الإداري للكادر المنفّذ.
func (h *BookingHandler) SetCrewNotes(w http.ResponseWriter, r *http.Request) {
	h.setNote(w, r, "/crew-notes", h.service.SetCrewNotes)
}

// PUT /api/bookings/{id}/project-notes — ملاحظة الإداري لمدير المشاريع.
func (h *BookingHandler) SetProjectNotes(w http.ResponseWriter, r *http.Request) {
	h.setNote(w, r, "/project-notes", h.service.SetProjectNotes)
}

// setNote الجسم المشترك للملاحظتين — نفس الشكل بالضبط، والفرق بالدالة
// الي تنحفظ بيها. تكرار الجسم مرتين يعني إصلاح أي عيب لازم ينعمل مرتين.
func (h *BookingHandler) setNote(
	w http.ResponseWriter, r *http.Request, suffix string,
	save func(id, note, byEmployeeID string) (*model.Booking, error),
) {
	id := strings.TrimSuffix(extractID(r.URL.Path, "/api/bookings/"), suffix)
	var req struct {
		Note string `json:"note"`
	}
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	booking, err := save(id, req.Note, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, booking)
}

// PUT /api/bookings/{id}/cancel — إلغاء بسبب مكتوب.
func (h *BookingHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSuffix(extractID(r.URL.Path, "/api/bookings/"), "/cancel")
	var req struct {
		Reason string `json:"reason"`
	}
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	booking, err := h.service.Cancel(id, req.Reason, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, booking)
}

// GET /api/bookings/stage-bucket?bucket=...  — حجوزات سلّة وحدة.
func (h *BookingHandler) ListByStageBucket(w http.ResponseWriter, r *http.Request) {
	rows, err := h.service.ListByStageBucket(r.URL.Query().Get("bucket"), 0)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// GET /api/bookings/stage-bucket-counts — أرقام فوق التبويبات.
func (h *BookingHandler) StageBucketCounts(w http.ResponseWriter, r *http.Request) {
	counts, err := h.service.StageBucketCounts()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر حساب الأعداد")
		return
	}
	WriteJSON(w, http.StatusOK, counts)
}

// GET /api/bookings/{id}/timeline — قصة الحجز كاملة + التأخيرات.
func (h *BookingHandler) Timeline(w http.ResponseWriter, r *http.Request) {
	if h.timeline == nil {
		WriteError(w, http.StatusServiceUnavailable, "الخط الزمني مو مربوط")
		return
	}
	tl, err := h.timeline.Build(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, tl)
}
