package handler

import (
	"errors"
	"log"
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/service"
)

// 📐 سيت المخططات (ك) 11-10 — الصلاحيات بالخدمة نفسها (نفس صيت التصاميم).
type PlanSetHandler struct {
	svc   *service.PlanSetService
	perms *repository.PermissionRepository
}

func NewPlanSetHandler(s *service.PlanSetService, p *repository.PermissionRepository) *PlanSetHandler {
	return &PlanSetHandler{svc: s, perms: p}
}

func (h *PlanSetHandler) who(r *http.Request) (string, string, []string) {
	me := middleware.EmployeeIDFromContext(r)
	var perms []string
	if h.perms != nil {
		perms, _ = h.perms.Names(me)
	}
	return me, middleware.RoleFromContext(r), perms
}

func (h *PlanSetHandler) fail(w http.ResponseWriter, err error) {
	if errors.Is(err, service.ErrDesignHidden) {
		WriteError(w, http.StatusNotFound, "مو موجود")
		return
	}
	WriteError(w, http.StatusBadRequest, err.Error())
}

// GET /api/plan-sets/meta — الأقسام والأنواع، والأرقام، والمشاريع للي يدير.
func (h *PlanSetHandler) Meta(w http.ResponseWriter, r *http.Request) {
	me, role, perms := h.who(r)
	if !h.svc.CanSee(me, role, perms) {
		h.fail(w, service.ErrDesignHidden)
		return
	}
	stats, err := h.svc.Stats()
	if err != nil {
		log.Printf("plan stats: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	ov, err := h.svc.Overview()
	if err != nil {
		log.Printf("plan overview: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	manage := h.svc.CanManage(me, role, perms)
	out := map[string]any{"groups": service.PlanCatalog, "stats": stats, "overview": ov, "canManage": manage, "projects": []service.ProjectOption{}}
	if manage {
		out["projects"] = h.svc.Projects()
	}
	WriteJSON(w, http.StatusOK, out)
}

func (h *PlanSetHandler) List(w http.ResponseWriter, r *http.Request) {
	me, role, perms := h.who(r)
	rows, err := h.svc.List(me, role, perms)
	if err != nil {
		if !errors.Is(err, service.ErrDesignHidden) {
			log.Printf("plan sets: %v", err)
		}
		h.fail(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

func (h *PlanSetHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in service.PlanSetInput
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

func (h *PlanSetHandler) Get(w http.ResponseWriter, r *http.Request) {
	me, role, perms := h.who(r)
	d, err := h.svc.Detail(me, role, perms, r.PathValue("id"))
	if err != nil {
		h.fail(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, d)
}

func (h *PlanSetHandler) Archive(w http.ResponseWriter, r *http.Request) {
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

func (h *PlanSetHandler) AddFile(w http.ResponseWriter, r *http.Request) {
	var in service.PlanFileInput
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

func (h *PlanSetHandler) RemoveFile(w http.ResponseWriter, r *http.Request) {
	me, role, perms := h.who(r)
	if err := h.svc.RemoveFile(me, role, perms, r.PathValue("id")); err != nil {
		h.fail(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// PUT /api/plan-sets/{id} — الاسم/المشروع/نموذج جاهز/عرض السعر المرتبط
func (h *PlanSetHandler) Update(w http.ResponseWriter, r *http.Request) {
	var in service.PlanSetUpdate
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	me, role, perms := h.who(r)
	if err := h.svc.Update(me, role, perms, r.PathValue("id"), in); err != nil {
		h.fail(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// POST /api/plan-sets/{id}/copy — ينسخ أقسام أو مخططات من سيت ثاني (نموذج جاهز)
func (h *PlanSetHandler) Copy(w http.ResponseWriter, r *http.Request) {
	var in service.PlanCopyInput
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	me, role, perms := h.who(r)
	n, err := h.svc.Copy(me, role, perms, r.PathValue("id"), in)
	if err != nil {
		h.fail(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]int{"copied": n})
}
