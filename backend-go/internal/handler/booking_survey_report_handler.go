package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type BookingSurveyReportHandler struct {
	service *service.BookingSurveyReportService
}

func NewBookingSurveyReportHandler(s *service.BookingSurveyReportService) *BookingSurveyReportHandler {
	return &BookingSurveyReportHandler{service: s}
}

// POST /api/booking-survey-reports
func (h *BookingSurveyReportHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateBookingSurveyReportRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	report, err := h.service.Create(middleware.EmployeeIDFromContext(r), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, report)
}

// GET /api/booking-survey-reports?bookingId=
func (h *BookingSurveyReportHandler) List(w http.ResponseWriter, r *http.Request) {
	reports, err := h.service.List(r.URL.Query().Get("bookingId"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, reports)
}
