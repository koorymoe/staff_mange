package handler

import (
	"encoding/json"
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/service"
)

// MatrixLearningHandler اقتراحات ماتركس — المدير يقرر (ADMIN/OWNER).
type MatrixLearningHandler struct {
	svc *service.MatrixLearningService
}

func NewMatrixLearningHandler(svc *service.MatrixLearningService) *MatrixLearningHandler {
	return &MatrixLearningHandler{svc: svc}
}

// GET /api/ai/proposals?status=PENDING
func (h *MatrixLearningHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.List(r.URL.Query().Get("status"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الاقتراحات")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// POST /api/ai/proposals/{id}/approve — {payload?} تعديل المدير قبل الموافقة.
func (h *MatrixLearningHandler) Approve(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Payload json.RawMessage `json:"payload"`
	}
	_ = DecodeJSON(r, &body)
	if err := h.svc.Approve(r.PathValue("id"), middleware.EmployeeIDFromContext(r), body.Payload); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// POST /api/ai/proposals/{id}/reject — {note} السبب يصير درس لماتركس.
func (h *MatrixLearningHandler) Reject(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Note string `json:"note"`
	}
	_ = DecodeJSON(r, &body)
	if err := h.svc.Reject(r.PathValue("id"), middleware.EmployeeIDFromContext(r), body.Note); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
