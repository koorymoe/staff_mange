package handler

import (
	"net/http"

	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type MissionHandler struct {
	service *service.MissionService
}

func NewMissionHandler(s *service.MissionService) *MissionHandler {
	return &MissionHandler{service: s}
}

// GET /api/v1/missions?stage=&leaderId=&employeeId=
func (h *MissionHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	missions, err := h.service.List(q.Get("stage"), q.Get("leaderId"), q.Get("employeeId"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب المهام")
		return
	}
	WriteJSON(w, http.StatusOK, missions)
}

// GET /api/v1/missions/{id}
func (h *MissionHandler) Get(w http.ResponseWriter, r *http.Request) {
	mission, err := h.service.Get(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, mission)
}

// POST /api/v1/missions
func (h *MissionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateMissionRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	mission, err := h.service.Create(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, mission)
}

// PUT /api/v1/missions/{id}/stage
func (h *MissionHandler) UpdateStage(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateMissionStageRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	mission, err := h.service.UpdateStage(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, mission)
}

// GET /api/v1/missions/my/{employeeId}
func (h *MissionHandler) ListForEmployee(w http.ResponseWriter, r *http.Request) {
	if !requireSelfOrSupervisor(w, r, r.PathValue("employeeId")) {
		return
	}
	missions, err := h.service.ListForEmployee(r.PathValue("employeeId"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب المهام")
		return
	}
	WriteJSON(w, http.StatusOK, missions)
}

// GET /api/v1/missions/monitor/live
func (h *MissionHandler) MonitorLive(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.MonitorLive()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب لوحة المتابعة")
		return
	}
	WriteJSON(w, http.StatusOK, result)
}

// GET /api/v1/missions/reports/performance?from=&to=
func (h *MissionHandler) PerformanceReport(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var from, to *string
	if v := q.Get("from"); v != "" {
		from = &v
	}
	if v := q.Get("to"); v != "" {
		to = &v
	}
	report, err := h.service.PerformanceReport(from, to)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر إنشاء التقرير")
		return
	}
	WriteJSON(w, http.StatusOK, report)
}
