package handler

import (
	"net/http"

	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type ServiceHandler struct {
	service *service.ServiceCatalogService
}

func NewServiceHandler(s *service.ServiceCatalogService) *ServiceHandler {
	return &ServiceHandler{service: s}
}

// GET /api/v1/services
func (h *ServiceHandler) List(w http.ResponseWriter, r *http.Request) {
	services, err := h.service.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب قائمة الخدمات")
		return
	}
	WriteJSON(w, http.StatusOK, services)
}

// POST /api/v1/services
func (h *ServiceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateServiceRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}

	svc, err := h.service.Create(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, svc)
}

// POST /api/v1/services/{id}/skills
func (h *ServiceHandler) CreateSkill(w http.ResponseWriter, r *http.Request) {
	var req model.CreateSkillRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}

	skill, err := h.service.CreateSkill(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, skill)
}

// DELETE /api/services/{id}
// PUT /api/services/{id}/manager-paperwork — الورق على مسؤول الخدمة.
func (h *ServiceHandler) SetManagerPaperwork(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	if err := h.service.SetManagerHandlesPaperwork(r.PathValue("id"), req.Enabled); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر حفظ الإعداد")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"enabled": req.Enabled})
}

func (h *ServiceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Delete(r.PathValue("id")); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GET /api/skills — كل المهارات بقائمة مسطّحة (لبرامج التدريب)
func (h *ServiceHandler) ListSkills(w http.ResponseWriter, r *http.Request) {
	rows, err := h.service.AllSkills()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب المهارات")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}
