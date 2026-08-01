package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
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
	bookings, err := h.service.List(r.URL.Query().Get("status"), r.URL.Query().Get("customerId"), r.URL.Query().Get("date"))
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
	_ = DecodeJSON(r, &req) // الملاحظة اختيارية
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
	bookings, err := h.service.List("PENDING", "", "")
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
