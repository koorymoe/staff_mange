package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/service"
)

// 🖼️ مكتبة صور أنواع العمل (ل) 11-10.
type WorkTypeImageHandler struct{ svc *service.WorkTypeImageService }

func NewWorkTypeImageHandler(s *service.WorkTypeImageService) *WorkTypeImageHandler {
	return &WorkTypeImageHandler{svc: s}
}

// GET /api/work-type-images?workTypeId= — بلا نوع: عدد الصور لكل نوع
func (h *WorkTypeImageHandler) List(w http.ResponseWriter, r *http.Request) {
	if id := r.URL.Query().Get("workTypeId"); id != "" {
		rows, err := h.svc.List(id)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "تعذر")
			return
		}
		WriteJSON(w, http.StatusOK, rows)
		return
	}
	st, err := h.svc.Stats()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	WriteJSON(w, http.StatusOK, st)
}

// POST /api/work-type-images {workTypeId, fileKeys[]}
func (h *WorkTypeImageHandler) Add(w http.ResponseWriter, r *http.Request) {
	var in struct {
		WorkTypeID string   `json:"workTypeId"`
		FileKeys   []string `json:"fileKeys"`
	}
	if err := DecodeJSON(r, &in); err != nil || in.WorkTypeID == "" || len(in.FileKeys) == 0 {
		WriteError(w, http.StatusBadRequest, "اختار نوع العمل والصور")
		return
	}
	n, err := h.svc.Add(middleware.EmployeeIDFromContext(r), in.WorkTypeID, in.FileKeys)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]int{"added": n})
}

// DELETE /api/work-type-images/{id} — يشيلها من المكتبة (تبقى بالقاعدة)
func (h *WorkTypeImageHandler) Remove(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Remove(r.PathValue("id")); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// POST /api/quotations/{id}/images/pick {chapters[], again}
func (h *WorkTypeImageHandler) Pick(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Chapters []string `json:"chapters"`
		Again    bool     `json:"again"`
	}
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	out, err := h.svc.Pick(r.PathValue("id"), in.Chapters, in.Again)
	if err != nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, out)
}
