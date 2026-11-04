package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type MatrixGuideHandler struct{ svc *service.MatrixGuideService }

func NewMatrixGuideHandler(svc *service.MatrixGuideService) *MatrixGuideHandler {
	return &MatrixGuideHandler{svc: svc}
}

// GET /api/ai/guide?path=&label= — توجيه ماتركس لإجراء (كل موظف لنفسه).
func (h *MatrixGuideHandler) Guide(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	res, err := h.svc.Guide(middleware.EmployeeIDFromContext(r), q.Get("path"), q.Get("label"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب التوجيه")
		return
	}
	WriteJSON(w, http.StatusOK, res)
}

// GET /api/ai/guide-rules (ADMIN/OWNER)
func (h *MatrixGuideHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.Rules()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب التعليمات")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"rules": rows, "modelEnabled": h.svc.ModelEnabled()})
}

func (h *MatrixGuideHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in model.MatrixGuideRule
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	out, err := h.svc.CreateRule(in)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, out)
}

func (h *MatrixGuideHandler) Update(w http.ResponseWriter, r *http.Request) {
	var in model.MatrixGuideRule
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	out, err := h.svc.UpdateRule(r.PathValue("id"), in)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, out)
}

func (h *MatrixGuideHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteRule(r.PathValue("id")); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحذف")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
