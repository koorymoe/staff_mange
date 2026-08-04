package handler

import (
	"net/http"
	"strings"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type AnnouncementHandler struct {
	repo *repository.AnnouncementRepository
}

func NewAnnouncementHandler(r *repository.AnnouncementRepository) *AnnouncementHandler {
	return &AnnouncementHandler{repo: r}
}

// GET /api/announcements — كل موظف يشوف الشغالة.
// ?all=1 للإدارة: تشمل المخفية.
func (h *AnnouncementHandler) List(w http.ResponseWriter, r *http.Request) {
	role := middleware.RoleFromContext(r)
	if r.URL.Query().Get("all") == "1" && (role == "ADMIN" || role == "OWNER") {
		rows, err := h.repo.ListAll()
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "تعذر جلب الإعلانات")
			return
		}
		WriteJSON(w, http.StatusOK, rows)
		return
	}
	rows, err := h.repo.ListActive()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الإعلانات")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// POST /api/announcements — المالك ومدير النظام حصراً (الراوت محمي).
func (h *AnnouncementHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateAnnouncementRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	body := strings.TrimSpace(req.Body)
	if body == "" {
		WriteError(w, http.StatusBadRequest, "اكتب نص الإعلان")
		return
	}
	out, err := h.repo.Create(body, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر نشر الإعلان")
		return
	}
	WriteJSON(w, http.StatusCreated, out)
}

// PUT /api/announcements/{id}/active
func (h *AnnouncementHandler) SetActive(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Active bool `json:"active"`
	}
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	if err := h.repo.SetActive(r.PathValue("id"), req.Active); err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر التحديث")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// DELETE /api/announcements/{id}
func (h *AnnouncementHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.Delete(r.PathValue("id")); err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر الحذف")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
