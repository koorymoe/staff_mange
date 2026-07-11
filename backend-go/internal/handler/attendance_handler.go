package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type AttendanceHandler struct {
	service *service.AttendanceService
}

func NewAttendanceHandler(s *service.AttendanceService) *AttendanceHandler {
	return &AttendanceHandler{service: s}
}

// POST /api/attendance/checkin
func (h *AttendanceHandler) CheckIn(w http.ResponseWriter, r *http.Request) {
	employeeID := middleware.EmployeeIDFromContext(r)
	rec, err := h.service.CheckIn(employeeID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر تسجيل الحضور")
		return
	}
	WriteJSON(w, http.StatusOK, rec)
}

// POST /api/attendance/checkout
func (h *AttendanceHandler) CheckOut(w http.ResponseWriter, r *http.Request) {
	employeeID := middleware.EmployeeIDFromContext(r)
	rec, err := h.service.CheckOut(employeeID)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, rec)
}

// GET /api/attendance/mine
func (h *AttendanceHandler) Mine(w http.ResponseWriter, r *http.Request) {
	employeeID := middleware.EmployeeIDFromContext(r)
	rec, err := h.service.Mine(employeeID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب سجل الحضور")
		return
	}
	WriteJSON(w, http.StatusOK, rec)
}

// GET /api/attendance/today
func (h *AttendanceHandler) Today(w http.ResponseWriter, r *http.Request) {
	records, err := h.service.Today()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب حضور اليوم")
		return
	}
	WriteJSON(w, http.StatusOK, records)
}

// GET /api/attendance/employee/{id}?month=YYYY-MM
func (h *AttendanceHandler) MonthlyReport(w http.ResponseWriter, r *http.Request) {
	report, err := h.service.MonthlyReport(r.PathValue("id"), r.URL.Query().Get("month"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, report)
}

// PUT /api/attendance/{id}
func (h *AttendanceHandler) Correct(w http.ResponseWriter, r *http.Request) {
	var req model.SetAttendanceCorrectionRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	rec, err := h.service.Correct(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, rec)
}
