package handler

import (
	"log"
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/service"
)

type DuplicateCandidateHandler struct {
	service *service.DuplicateCandidateService
}

func NewDuplicateCandidateHandler(s *service.DuplicateCandidateService) *DuplicateCandidateHandler {
	return &DuplicateCandidateHandler{service: s}
}

// GET /api/duplicate-candidates?kind=&status=
func (h *DuplicateCandidateHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	rows, err := h.service.List(q.Get("kind"), q.Get("status"))
	if err != nil {
		log.Printf("list duplicate candidates: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب قائمة التكرار")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// PUT /api/duplicate-candidates/{id}/dismiss
func (h *DuplicateCandidateHandler) Dismiss(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Dismiss(r.PathValue("id"), middleware.EmployeeIDFromContext(r)); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
