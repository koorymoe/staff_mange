package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type PerformanceReviewHandler struct {
	service *service.PerformanceReviewService
}

func NewPerformanceReviewHandler(s *service.PerformanceReviewService) *PerformanceReviewHandler {
	return &PerformanceReviewHandler{service: s}
}

func (h *PerformanceReviewHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreatePerformanceReviewRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	review, err := h.service.Create(middleware.EmployeeIDFromContext(r), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, review)
}

// GET /api/performance-reviews/ratable — الموظفين الي المستخدم الحالي يقدر
// يقيّمهم (تيم ليدر: زملاء حجوزاته الفعليين بس، أدمن/إداري كوادر: كل التيم ليدرات).
func (h *PerformanceReviewHandler) Ratable(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.RatableEmployees(middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, list)
}

func (h *PerformanceReviewHandler) ListForEmployee(w http.ResponseWriter, r *http.Request) {
	if !requireSelfOrSupervisor(w, r, r.PathValue("employeeId")) {
		return
	}
	reviews, err := h.service.ListForEmployee(r.PathValue("employeeId"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب التقييمات")
		return
	}
	WriteJSON(w, http.StatusOK, reviews)
}

func (h *PerformanceReviewHandler) List(w http.ResponseWriter, r *http.Request) {
	reviews, err := h.service.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب التقييمات")
		return
	}
	WriteJSON(w, http.StatusOK, reviews)
}

// GET /api/performance-reviews/my-bookings
//
// حجوزات الليدر المنجزة وكادر كل وحدة وحالة تقييمهم — الليدر ما
// يحتاج يدور على موظفيه بقائمة، النظام يگله «هذني شغلاتك ومنو طلع
// وياك بكل وحدة».
func (h *PerformanceReviewHandler) MyBookings(w http.ResponseWriter, r *http.Request) {
	// نطاق تاريخ الإنجاز يجي من مرشّح الشاشة. چان الترشيح يصير
	// بالواجهة بعد ما الخادم قصّ على ٣٠ يوم، فاختيار تاريخ قديم
	// ما چان يجيب ولا شي.
	rows, err := h.service.BookingsAwaitingReview(
		middleware.EmployeeIDFromContext(r),
		r.URL.Query().Get("from"),
		r.URL.Query().Get("to"),
	)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الحجوزات")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// GET /api/performance-reviews/evaluator-leaderboard — «تقييم بين الإداريين»:
// ترتيب ADMIN/OWNER/MONITOR/HR_COORDINATOR حسب عدد الحجوزات الي راجعوها.
func (h *PerformanceReviewHandler) EvaluatorLeaderboard(w http.ResponseWriter, r *http.Request) {
	board, err := h.service.EvaluatorLeaderboard()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب ترتيب المراجعين")
		return
	}
	WriteJSON(w, http.StatusOK, board)
}
