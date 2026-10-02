package handler

import (
	"net/http"

	"staffmange-api/internal/service"
)

type MatrixReportHandler struct {
	reports  *service.MatrixEmployeeReportService
	business *service.MatrixBusinessService
}

func NewMatrixReportHandler(r *service.MatrixEmployeeReportService, b *service.MatrixBusinessService) *MatrixReportHandler {
	return &MatrixReportHandler{reports: r, business: b}
}

// GET /api/ai/employee-report?employeeId=&day= (ADMIN/OWNER)
func (h *MatrixReportHandler) Employee(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("employeeId")
	if id == "" {
		WriteError(w, http.StatusBadRequest, "employeeId مطلوب")
		return
	}
	rep, err := h.reports.Report(id, r.URL.Query().Get("day"))
	if err != nil {
		WriteError(w, http.StatusNotFound, "الموظف مو موجود")
		return
	}
	WriteJSON(w, http.StatusOK, rep)
}

// GET /api/ai/business (ADMIN/OWNER)
func (h *MatrixReportHandler) Business(w http.ResponseWriter, r *http.Request) {
	v, err := h.business.View()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر حساب الأرقام")
		return
	}
	WriteJSON(w, http.StatusOK, v)
}
