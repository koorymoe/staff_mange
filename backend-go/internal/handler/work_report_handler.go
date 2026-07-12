package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type WorkReportHandler struct {
	service *service.WorkReportService
}

func NewWorkReportHandler(s *service.WorkReportService) *WorkReportHandler {
	return &WorkReportHandler{service: s}
}

func (h *WorkReportHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateWorkReportRequest
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

// GET /api/work-reports?employeeId= — بدون employeeId: كل التقارير (للمراقب/الجودة)
func (h *WorkReportHandler) List(w http.ResponseWriter, r *http.Request) {
	reports, err := h.service.List(r.URL.Query().Get("employeeId"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب تقارير العمل")
		return
	}
	WriteJSON(w, http.StatusOK, reports)
}
