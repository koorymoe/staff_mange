package handler

import (
	"staffmange-api/internal/middleware"
	"net/http"

	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type TrainingHandler struct {
	service *service.TrainingService
}

func NewTrainingHandler(s *service.TrainingService) *TrainingHandler {
	return &TrainingHandler{service: s}
}

// GET /api/v1/training/materials/mine?employeeId=
func (h *TrainingHandler) MaterialsMine(w http.ResponseWriter, r *http.Request) {
	// ⚠️⚠️ اسمه «مالتي» بس چان ياخذ الرقم **من الرابط** مو من
	// الجلسة — يعني أي موظف يبدّل الرقم ويقرا مواد تدريب زميله.
	// الرقم صار من الجلسة دائماً، والي بالرابط ينتجاهل.
	employeeID := middleware.EmployeeIDFromContext(r)
	if employeeID == "" {
		WriteError(w, http.StatusUnauthorized, "جلسة غير صالحة")
		return
	}
	result, err := h.service.MaterialsMine(employeeID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب مواد التدريب")
		return
	}
	WriteJSON(w, http.StatusOK, result)
}

// GET /api/v1/training/assignments/{employeeId}
func (h *TrainingHandler) Assignments(w http.ResponseWriter, r *http.Request) {
	// تعيينات تدريب موظف = سجل شخصي: صاحبه أو مشرفه.
	if !requireSelfOrSupervisor(w, r, r.PathValue("employeeId")) {
		return
	}
	services, err := h.service.Assignments(r.PathValue("employeeId"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب التعيينات")
		return
	}
	WriteJSON(w, http.StatusOK, services)
}

// PUT /api/v1/training/assignments/{employeeId}
func (h *TrainingHandler) SetAssignments(w http.ResponseWriter, r *http.Request) {
	var req model.SetTrainingAssignmentsRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	services, err := h.service.SetAssignments(r.PathValue("employeeId"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, services)
}

// GET /api/v1/training/materials?serviceId=
func (h *TrainingHandler) ListMaterials(w http.ResponseWriter, r *http.Request) {
	materials, err := h.service.ListMaterials(r.URL.Query().Get("serviceId"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب المواد التدريبية")
		return
	}
	WriteJSON(w, http.StatusOK, materials)
}

// POST /api/v1/training/materials
func (h *TrainingHandler) CreateMaterial(w http.ResponseWriter, r *http.Request) {
	var req model.CreateTrainingMaterialRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	material, err := h.service.CreateMaterial(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, material)
}

// PUT /api/v1/training/materials/{id}
func (h *TrainingHandler) UpdateMaterial(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateTrainingMaterialRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	material, err := h.service.UpdateMaterial(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, material)
}

// DELETE /api/v1/training/materials/{id}
func (h *TrainingHandler) DeleteMaterial(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteMaterial(r.PathValue("id")); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
