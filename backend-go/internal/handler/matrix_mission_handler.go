package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/service"
)

// ═══ 🎯 مهام ماتركس «راقب فلان» — المالك والمدير (قرار (ع) 10-11) ═══
type MatrixMissionHandler struct{ svc *service.MatrixMissionService }

func NewMatrixMissionHandler(svc *service.MatrixMissionService) *MatrixMissionHandler {
	return &MatrixMissionHandler{svc: svc}
}

// GET /api/matrix/missions
func (h *MatrixMissionHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.List(middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب المهام")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// GET /api/matrix/missions/{id}
func (h *MatrixMissionHandler) Get(w http.ResponseWriter, r *http.Request) {
	m, logs, err := h.svc.Get(r.PathValue("id"), middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusNotFound, "المهمة غير موجودة")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"mission": m, "logs": logs})
}

// POST /api/matrix/missions {employeeId, note, days}
func (h *MatrixMissionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in struct {
		EmployeeID string `json:"employeeId"`
		Note       string `json:"note"`
		Days       int    `json:"days"`
	}
	if err := DecodeJSON(r, &in); err != nil || in.EmployeeID == "" {
		WriteError(w, http.StatusBadRequest, "اختار الموظف")
		return
	}
	m, err := h.svc.Create(middleware.EmployeeIDFromContext(r), in.EmployeeID, in.Note, in.Days)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, m)
}

// DELETE /api/matrix/missions/{id} — يوقفها (ما تنحذف)
func (h *MatrixMissionHandler) Stop(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Stop(r.PathValue("id"), middleware.EmployeeIDFromContext(r)); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
