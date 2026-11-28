package handler

import (
	"errors"
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/service"
)

// ═══ 🔧 الخطأ الفني والتصحيح (ل) 11-24 ═══

type ReworkHandler struct{ svc *service.Rework }

func NewReworkHandler(s *service.Rework) *ReworkHandler { return &ReworkHandler{svc: s} }

// GET /api/rework
func (h *ReworkHandler) List(w http.ResponseWriter, r *http.Request) {
	me := middleware.EmployeeIDFromContext(r)
	rows, err := h.svc.List(me)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"items": rows, "canDecide": h.svc.CanDecide(me), "canReport": h.svc.CanReport(me)})
}

// POST /api/rework {fixCode, origCode, note}
func (h *ReworkHandler) Submit(w http.ResponseWriter, r *http.Request) {
	var in struct{ FixCode, OrigCode, Note string }
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	if err := h.svc.Submit(middleware.EmployeeIDFromContext(r), in.FixCode, in.OrigCode, in.Note); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.List(w, r)
}

// POST /api/rework/{id}/decide {approve, note}
func (h *ReworkHandler) Decide(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Approve bool   `json:"approve"`
		Note    string `json:"note"`
	}
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	if err := h.svc.Decide(middleware.EmployeeIDFromContext(r), r.PathValue("id"), in.Approve, in.Note); err != nil {
		if errors.Is(err, service.ErrDesignHidden) {
			WriteError(w, http.StatusNotFound, "مو موجود")
			return
		}
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.List(w, r)
}

// GET /api/rework/bookings?q= — (11-28) بحث الحجز بالكود أو الزبون، ومعلوماته تطلع لحالها
func (h *ReworkHandler) Search(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.SearchBookings(middleware.EmployeeIDFromContext(r), r.URL.Query().Get("q"))
	if err != nil {
		if errors.Is(err, service.ErrDesignHidden) {
			WriteError(w, http.StatusNotFound, "مو موجود")
			return
		}
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"items": rows})
}

// POST /api/rework/{id}/fix {fixCode} — (11-28) ربط حجز التصحيح بعدين
func (h *ReworkHandler) LinkFix(w http.ResponseWriter, r *http.Request) {
	var in struct {
		FixCode string `json:"fixCode"`
	}
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	if err := h.svc.LinkFix(middleware.EmployeeIDFromContext(r), r.PathValue("id"), in.FixCode); err != nil {
		if errors.Is(err, service.ErrDesignHidden) {
			WriteError(w, http.StatusNotFound, "مو موجود")
			return
		}
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.List(w, r)
}
