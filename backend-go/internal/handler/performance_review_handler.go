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

func (h *PerformanceReviewHandler) ListForEmployee(w http.ResponseWriter, r *http.Request) {
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
