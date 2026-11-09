package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/repository"
)

// ItAssetHandler جرد أجهزة تقنية المعلومات وسجل صيانتها وإحصائياتها.
type ItAssetHandler struct {
	repo *repository.ItAssetRepository
}

func NewItAssetHandler(repo *repository.ItAssetRepository) *ItAssetHandler {
	return &ItAssetHandler{repo: repo}
}

// GET /api/it/assets
func (h *ItAssetHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.repo.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الأجهزة")
		return
	}
	WriteJSON(w, http.StatusOK, list)
}

// POST /api/it/assets
func (h *ItAssetHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in repository.ItAssetInput
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	a, err := h.repo.Create(in, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, a)
}

// PUT /api/it/assets/{id}
func (h *ItAssetHandler) Update(w http.ResponseWriter, r *http.Request) {
	var in repository.ItAssetInput
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	a, err := h.repo.Update(r.PathValue("id"), in, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, a)
}

// DELETE /api/it/assets/{id}
func (h *ItAssetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.Delete(r.PathValue("id")); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر حذف الجهاز")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GET /api/it/assets/{id}/logs
func (h *ItAssetHandler) Logs(w http.ResponseWriter, r *http.Request) {
	logs, err := h.repo.Logs(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب سجل الجهاز")
		return
	}
	WriteJSON(w, http.StatusOK, logs)
}

// POST /api/it/assets/{id}/logs — {kind, note, cost?}
func (h *ItAssetHandler) AddLog(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Kind string   `json:"kind"`
		Note string   `json:"note"`
		Cost *float64 `json:"cost"`
	}
	if err := DecodeJSON(r, &body); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	l, err := h.repo.AddLog(r.PathValue("id"), body.Kind, body.Note, body.Cost, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, l)
}

// GET /api/it/stats
func (h *ItAssetHandler) Stats(w http.ResponseWriter, r *http.Request) {
	s, err := h.repo.Stats()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر حساب الإحصائيات")
		return
	}
	WriteJSON(w, http.StatusOK, s)
}
