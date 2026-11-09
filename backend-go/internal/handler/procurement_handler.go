package handler

import (
	"errors"
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type ProcurementHandler struct {
	afterFulfill  func(procurementID string)
	beforeFulfill func(procurementID string, req model.FulfillProcurementRequestRequest) error
	service       *service.ProcurementService
}

func NewProcurementHandler(s *service.ProcurementService) *ProcurementHandler {
	return &ProcurementHandler{service: s}
}

// GET /api/v1/procurement
func (h *ProcurementHandler) List(w http.ResponseWriter, r *http.Request) {
	requests, err := h.service.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب طلبات الشراء")
		return
	}
	WriteJSON(w, http.StatusOK, requests)
}

// GET /api/v1/procurement/stats
func (h *ProcurementHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.Stats()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الإحصائيات")
		return
	}
	WriteJSON(w, http.StatusOK, stats)
}

// POST /api/v1/procurement
func (h *ProcurementHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateProcurementRequestRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	employeeID := middleware.EmployeeIDFromContext(r)
	role := middleware.RoleFromContext(r)
	request, err := h.service.Create(employeeID, role, req)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, service.ErrForbiddenRequestType) {
			status = http.StatusForbidden
		}
		WriteError(w, status, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, request)
}

// PUT /api/v1/procurement/{id}/status
func (h *ProcurementHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateProcurementStatusRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	req.DecidedByID = middleware.EmployeeIDFromContext(r)
	request, err := h.service.UpdateStatus(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, request)
}

// PUT /api/v1/procurement/{id}/fulfill
func (h *ProcurementHandler) Fulfill(w http.ResponseWriter, r *http.Request) {
	var req model.FulfillProcurementRequestRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	if h.beforeFulfill != nil {
		if err := h.beforeFulfill(r.PathValue("id"), req); err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	request, err := h.service.Fulfill(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if h.afterFulfill != nil {
		go h.afterFulfill(r.PathValue("id"))
	}
	WriteJSON(w, http.StatusOK, request)
}

// SetAfterFulfill (ع) 10-15: ماتركس يفحص الأسعار بعد كل شراء
func (h *ProcurementHandler) SetAfterFulfill(f func(procurementID string)) { h.afterFulfill = f }

// SetBeforeFulfill (ع) 10-15: شرط العرضين — يرفض الشراء الكبير بلا عرض مورد ثاني
func (h *ProcurementHandler) SetBeforeFulfill(f func(string, model.FulfillProcurementRequestRequest) error) {
	h.beforeFulfill = f
}
