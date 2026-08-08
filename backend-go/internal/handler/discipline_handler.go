package handler

import (
	"net/http"
	"strconv"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/service"
)

type DisciplineHandler struct {
	service *service.DisciplineService
}

func NewDisciplineHandler(s *service.DisciplineService) *DisciplineHandler {
	return &DisciplineHandler{service: s}
}

// GET /api/discipline — أرصدة نقاط الموظفين (الناقصين أول)
func (h *DisciplineHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, items)
}

// GET /api/discipline/events?employeeId=&limit= — سجل الغرامات والرجوعات
func (h *DisciplineHandler) Events(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := h.service.Events(r.URL.Query().Get("employeeId"), limit)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, items)
}

// POST /api/discipline/adjust — تعديل يدوي على رصيد موظف.
// للمالك ومدير النظام حصراً (مربوط بالراوتر). كل تعديل ينسجّل بالسجل
// باسم الي عدّله وسببه.
func (h *DisciplineHandler) Adjust(w http.ResponseWriter, r *http.Request) {
	var req struct {
		EmployeeID string `json:"employeeId"`
		Delta      int    `json:"delta"`
		Reason     string `json:"reason"`
	}
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صالحة")
		return
	}
	result, err := h.service.Adjust(req.EmployeeID, req.Delta, req.Reason, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, result)
}

// POST /api/discipline/run — تشغيل الفحص فوراً بدل انتظار الدورة.
// للمالك/المدير حصراً — يفيد بعد ما يعدّل شي ويريد يشوف الأثر حالاً.
func (h *DisciplineHandler) Run(w http.ResponseWriter, r *http.Request) {
	h.service.RunPaperworkSweep()
	h.service.RunAuditSweep()
	h.service.RunRestoreSweep()
	items, err := h.service.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, items)
}
