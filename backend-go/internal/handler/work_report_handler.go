package handler

import (
	"errors"
	"net/http"
	"strings"

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
	report, err := h.service.Create(middleware.EmployeeIDFromContext(r), middleware.RoleFromContext(r), req)
	if err != nil {
		// ⚠️ الرفض بسبب الصلاحية يطلع ٤٠٣ مو ٤٠٠: الواجهة تفرّق
		// بينهما، و«بيانات الطلب غير صحيحة» على منع صلاحية تخلّي
		// الفني يعيد كتابة تقريره عشر مرات بلا فايدة.
		if errors.Is(err, service.ErrPaperworkNotYours) || strings.Contains(err.Error(), "مسؤول الخدمة") {
			WriteError(w, http.StatusForbidden, err.Error())
			return
		}
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
