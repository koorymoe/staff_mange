package handler

import (
	"errors"
	"log"
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/service"
)

// 🏗️ صيت التصاميم (ع) 11-08 — الصلاحيات بالخدمة (المالك/مدير النظام، الهندسة ووحدة
// التصميم، إدارة المشاريع، والمصمم المكلّف لصيته بس).
type DesignSetHandler struct {
	svc   *service.DesignSetService
	perms *repository.PermissionRepository
}

func NewDesignSetHandler(s *service.DesignSetService, p *repository.PermissionRepository) *DesignSetHandler {
	return &DesignSetHandler{svc: s, perms: p}
}

func (h *DesignSetHandler) who(r *http.Request) (string, string, []string) {
	me := middleware.EmployeeIDFromContext(r)
	var perms []string
	if h.perms != nil {
		perms, _ = h.perms.Names(me)
	}
	return me, middleware.RoleFromContext(r), perms
}

func (h *DesignSetHandler) fail(w http.ResponseWriter, err error) {
	if errors.Is(err, service.ErrDesignHidden) {
		WriteError(w, http.StatusNotFound, "مو موجود")
		return
	}
	WriteError(w, http.StatusBadRequest, err.Error())
}

// GET /api/design-sets/meta — التخصصات والأدوات وشنو يكدر يسوي
func (h *DesignSetHandler) Meta(w http.ResponseWriter, r *http.Request) {
	me, role, perms := h.who(r)
	manage := h.svc.CanManage(me, role, perms)
	tools, err := h.svc.Tools(manage)
	if err != nil {
		log.Printf("design tools: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	out := map[string]any{
		"disciplines": service.DesignDisciplines, "tools": tools,
		"canManage": manage, "canReview": h.svc.CanReview(me, role), "projects": []service.ProjectOption{},
	}
	if manage {
		out["projects"] = h.svc.Projects()
	}
	WriteJSON(w, http.StatusOK, out)
}

func (h *DesignSetHandler) List(w http.ResponseWriter, r *http.Request) {
	me, role, perms := h.who(r)
	rows, err := h.svc.List(me, role, perms)
	if err != nil {
		log.Printf("design sets: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

func (h *DesignSetHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in service.DesignSetInput
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	me, role, perms := h.who(r)
	id, err := h.svc.Create(me, role, perms, in)
	if err != nil {
		h.fail(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *DesignSetHandler) Get(w http.ResponseWriter, r *http.Request) {
	me, role, perms := h.who(r)
	d, err := h.svc.Detail(me, role, perms, r.PathValue("id"))
	if err != nil {
		h.fail(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, d)
}

func (h *DesignSetHandler) Archive(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Archived bool `json:"archived"`
	}
	_ = DecodeJSON(r, &in)
	me, role, perms := h.who(r)
	if err := h.svc.Archive(me, role, perms, r.PathValue("id"), in.Archived); err != nil {
		h.fail(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *DesignSetHandler) AddSection(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Discipline string `json:"discipline"`
	}
	_ = DecodeJSON(r, &in)
	me, role, perms := h.who(r)
	if err := h.svc.AddSection(me, role, perms, r.PathValue("id"), in.Discipline); err != nil {
		h.fail(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *DesignSetHandler) Assign(w http.ResponseWriter, r *http.Request) {
	var in struct {
		AssigneeID string `json:"assigneeId"`
		DueDate    string `json:"dueDate"`
	}
	_ = DecodeJSON(r, &in)
	me, role, perms := h.who(r)
	if err := h.svc.Assign(me, role, perms, r.PathValue("id"), in.AssigneeID, in.DueDate); err != nil {
		h.fail(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *DesignSetHandler) AddFile(w http.ResponseWriter, r *http.Request) {
	var in service.DesignFileInput
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	me, role, perms := h.who(r)
	id, err := h.svc.AddFile(me, role, perms, r.PathValue("id"), in)
	if err != nil {
		h.fail(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *DesignSetHandler) RemoveFile(w http.ResponseWriter, r *http.Request) {
	me, role, perms := h.who(r)
	if err := h.svc.RemoveFile(me, role, perms, r.PathValue("id")); err != nil {
		h.fail(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *DesignSetHandler) Submit(w http.ResponseWriter, r *http.Request) {
	me, role, perms := h.who(r)
	if err := h.svc.Submit(me, role, perms, r.PathValue("id")); err != nil {
		h.fail(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *DesignSetHandler) Review(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Approve bool   `json:"approve"`
		Note    string `json:"note"`
	}
	_ = DecodeJSON(r, &in)
	me, role, _ := h.who(r)
	if err := h.svc.Review(me, role, r.PathValue("id"), in.Approve, in.Note); err != nil {
		h.fail(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *DesignSetHandler) SaveTool(w http.ResponseWriter, r *http.Request) {
	var in service.DesignToolInput
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	me, role, perms := h.who(r)
	id, err := h.svc.SaveTool(me, role, perms, r.PathValue("id"), in)
	if err != nil {
		h.fail(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"id": id})
}
