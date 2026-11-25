package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

// ═══ 🧰 عهدة الأدوات (ل) 11-25 دازاي ═══

type ToolCustodyHandler struct {
	svc *service.ToolCustody
	db  *sqlx.DB
}

func NewToolCustodyHandler(s *service.ToolCustody, db *sqlx.DB) *ToolCustodyHandler {
	return &ToolCustodyHandler{svc: s, db: db}
}

func (h *ToolCustodyHandler) isManager(r *http.Request) bool {
	return model.IsAdminRole(middleware.RoleFromContext(r))
}

func (h *ToolCustodyHandler) fail(w http.ResponseWriter, err error) {
	if errors.Is(err, service.ErrDesignHidden) {
		WriteError(w, http.StatusNotFound, "مو موجود")
		return
	}
	WriteError(w, http.StatusBadRequest, err.Error())
}

// GET /api/tool-custody — عهدتي + الأدوات المتوفرة (+ كل العهد للمالك/المدير)
func (h *ToolCustodyHandler) Get(w http.ResponseWriter, r *http.Request) {
	me := middleware.EmployeeIDFromContext(r)
	mine, err := h.svc.Mine(me)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	tools, _ := h.svc.Tools()
	out := map[string]any{"mine": mine, "tools": tools, "canManage": h.isManager(r)}
	if h.isManager(r) {
		all, _ := h.svc.All()
		out["all"] = all
	}
	WriteJSON(w, http.StatusOK, out)
}

// POST /api/tool-custody/take {toolId, takenAt, note, employeeId?} — employeeId للمدير بس
func (h *ToolCustodyHandler) Take(w http.ResponseWriter, r *http.Request) {
	var in struct{ ToolID, TakenAt, Note, EmployeeID string }
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	me := middleware.EmployeeIDFromContext(r)
	emp := me
	if in.EmployeeID != "" && in.EmployeeID != me {
		if !h.isManager(r) {
			WriteError(w, http.StatusNotFound, "مو موجود")
			return
		}
		emp = in.EmployeeID
	}
	if _, err := h.svc.Take(me, emp, in.ToolID, in.TakenAt, in.Note, time.Now()); err != nil {
		h.fail(w, err)
		return
	}
	h.Get(w, r)
}

// POST /api/tool-custody/{id}/return {photo, note} — الموظف يسلّم
func (h *ToolCustodyHandler) Return(w http.ResponseWriter, r *http.Request) {
	var in struct{ Photo, Note string }
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	if err := h.svc.RequestReturn(middleware.EmployeeIDFromContext(r), r.PathValue("id"), in.Photo, in.Note); err != nil {
		h.fail(w, err)
		return
	}
	h.Get(w, r)
}

// POST /api/tool-custody/{id}/verify {ok, note} — المالك/المدير يدقق
func (h *ToolCustodyHandler) Verify(w http.ResponseWriter, r *http.Request) {
	if !h.isManager(r) {
		WriteError(w, http.StatusNotFound, "مو موجود")
		return
	}
	var in struct {
		OK   bool   `json:"ok"`
		Note string `json:"note"`
	}
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	if err := h.svc.Verify(middleware.EmployeeIDFromContext(r), r.PathValue("id"), in.OK, in.Note); err != nil {
		h.fail(w, err)
		return
	}
	h.Get(w, r)
}
