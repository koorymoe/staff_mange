package handler

import (
	"errors"
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/service"
)

// (دازاي 11-21 — كايزن) 💡 عندي فكرة
type IdeaHandler struct{ svc *service.IdeaService }

func NewIdeaHandler(svc *service.IdeaService) *IdeaHandler { return &IdeaHandler{svc: svc} }

// POST /api/me/ideas {text, category}
func (h *IdeaHandler) Submit(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Text     string `json:"text"`
		Category string `json:"category"`
	}
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	if err := h.svc.Submit(middleware.EmployeeIDFromContext(r), in.Text, in.Category); err != nil {
		if errors.Is(err, service.ErrIdeaInvalid) {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// GET /api/me/ideas
func (h *IdeaHandler) Mine(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.Mine(middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الجلب")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// GET /api/ideas — الي تنتظر ردي
func (h *IdeaHandler) Queue(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.Queue(middleware.EmployeeIDFromContext(r), middleware.RoleFromContext(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الجلب")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// POST /api/ideas/{id}/decide {status, reply}
func (h *IdeaHandler) Decide(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Status string `json:"status"`
		Reply  string `json:"reply"`
	}
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	err := h.svc.Decide(middleware.EmployeeIDFromContext(r), middleware.RoleFromContext(r), r.PathValue("id"), in.Status, in.Reply)
	if errors.Is(err, service.ErrReviewHidden) {
		WriteError(w, http.StatusNotFound, "مو موجود")
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
