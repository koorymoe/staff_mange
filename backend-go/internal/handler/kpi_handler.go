package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type KpiHandler struct {
	service *service.KpiService
}

func NewKpiHandler(s *service.KpiService) *KpiHandler {
	return &KpiHandler{service: s}
}

// GET /api/v1/kpi
func (h *KpiHandler) List(w http.ResponseWriter, r *http.Request) {
	evals, err := h.service.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب التقييمات")
		return
	}
	WriteJSON(w, http.StatusOK, evals)
}

// GET /api/v1/kpi/employee/{employeeId}
func (h *KpiHandler) ListForEmployee(w http.ResponseWriter, r *http.Request) {
	evals, err := h.service.ListForEmployee(r.PathValue("employeeId"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب التقييمات")
		return
	}
	WriteJSON(w, http.StatusOK, evals)
}

// POST /api/v1/kpi
func (h *KpiHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateKpiEvaluationRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	eval, err := h.service.Create(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, eval)
}

// POST /api/v1/employees/{id}/complete-training
func (h *KpiHandler) CompleteTraining(w http.ResponseWriter, r *http.Request) {
	eval, err := h.service.CompleteTraining(r.PathValue("id"), middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, eval)
}

// DELETE /api/v1/kpi/{id}
func (h *KpiHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Delete(r.PathValue("id")); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"success": true})
}
