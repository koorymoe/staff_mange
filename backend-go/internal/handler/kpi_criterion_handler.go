package handler

import (
	"net/http"

	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type KpiCriterionHandler struct {
	service *service.KpiCriterionService
}

func NewKpiCriterionHandler(s *service.KpiCriterionService) *KpiCriterionHandler {
	return &KpiCriterionHandler{service: s}
}

// GET /api/kpi-criteria — أي موظف عنده صلاحية تسجيل تقييم يشوف القائمة
func (h *KpiCriterionHandler) List(w http.ResponseWriter, r *http.Request) {
	criteria, err := h.service.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب نقاط الكي بي اي")
		return
	}
	WriteJSON(w, http.StatusOK, criteria)
}

// POST /api/kpi-criteria — إضافة نقطة جديدة (صلاحية kpi_criteria_management)
func (h *KpiCriterionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateKpiCriterionRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	criterion, err := h.service.Create(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, criterion)
}

// DELETE /api/kpi-criteria/{id} — حذف نقطة قديمة (صلاحية kpi_criteria_management)
func (h *KpiCriterionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Delete(r.PathValue("id")); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"success": true})
}
