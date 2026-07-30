package handler

import (
	"net/http"

	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type ProjectHandler struct {
	service *service.ProjectService
}

func NewProjectHandler(s *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{service: s}
}

// GET /api/v1/projects
func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب المشاريع")
		return
	}
	WriteJSON(w, http.StatusOK, result)
}

// GET /api/projects/{id} — المشروع كامل بما بيه ملفات العقد (تُطلب عند الحاجة فقط)
func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	project, err := h.service.Get(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusNotFound, "المشروع غير موجود")
		return
	}
	WriteJSON(w, http.StatusOK, project)
}

// POST /api/v1/projects
func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateProjectRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	project, err := h.service.Create(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, project)
}

// PUT /api/v1/projects/{id}
func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateProjectRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	project, err := h.service.Update(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, project)
}

// DELETE /api/projects/{id}/contract?which=plain|signed|both — حذف ملف العقد
// المرفوع. الراوت محمي بـrequireAdmin، يعني مدير النظام بس يقدر يحذف.
func (h *ProjectHandler) DeleteContract(w http.ResponseWriter, r *http.Request) {
	which := r.URL.Query().Get("which")
	req := model.UpdateProjectRequest{}
	switch which {
	case "plain":
		req.ClearContract = true
	case "signed":
		req.ClearSignedContract = true
	case "", "both":
		req.ClearContract = true
		req.ClearSignedContract = true
	default:
		WriteError(w, http.StatusBadRequest, "قيمة غير صحيحة")
		return
	}
	project, err := h.service.Update(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, project)
}

// DELETE /api/v1/projects/{id}
func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Delete(r.PathValue("id")); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
