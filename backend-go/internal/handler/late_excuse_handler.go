package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/service"
	"staffmange-api/internal/timeutil"
)

// LateExcuseHandler أعذار التأخير والانضباط بالدوام (دازاي 11-11).
type LateExcuseHandler struct{ svc *service.LateExcuseService }

func NewLateExcuseHandler(s *service.LateExcuseService) *LateExcuseHandler {
	return &LateExcuseHandler{svc: s}
}

// POST /api/late-excuses {date, reason, minutes}
func (h *LateExcuseHandler) Submit(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Date    string `json:"date"`
		Reason  string `json:"reason"`
		Minutes *int   `json:"minutes"`
	}
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	x, err := h.svc.Submit(middleware.EmployeeIDFromContext(r), in.Date, in.Reason, in.Minutes)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, x)
}

// GET /api/late-excuses/mine — أعذاري + انضباطي هالشهر
func (h *LateExcuseHandler) Mine(w http.ResponseWriter, r *http.Request) {
	me := middleware.EmployeeIDFromContext(r)
	rows, err := h.svc.Mine(me)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	month := r.URL.Query().Get("month")
	from, to, ok := timeutil.MonthStart(month)
	if !ok {
		from, to, _ = timeutil.MonthStart(timeutil.ThisMonth())
	}
	WriteJSON(w, http.StatusOK, map[string]any{"excuses": rows, "discipline": h.svc.DisciplineFor(me, from, to),
		"startsOn": service.DisciplineStart.Format("2006-01-02")})
}

// GET /api/late-excuses/inbox — الي أگدر أبت بيها
func (h *LateExcuseHandler) Inbox(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.Inbox(middleware.EmployeeIDFromContext(r), middleware.RoleFromContext(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// POST /api/late-excuses/{id}/decide {approve, note}
func (h *LateExcuseHandler) Decide(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Approve bool   `json:"approve"`
		Note    string `json:"note"`
	}
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	x, err := h.svc.Decide(middleware.EmployeeIDFromContext(r), middleware.RoleFromContext(r), r.PathValue("id"), in.Approve, in.Note)
	if err == service.ErrExcuseHidden {
		WriteError(w, http.StatusNotFound, "مو موجود")
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, x)
}

// GET /api/late-excuses/audit?month= — للمالك: مين وافق على شنو
func (h *LateExcuseHandler) Audit(w http.ResponseWriter, r *http.Request) {
	from, to, ok := timeutil.MonthStart(r.URL.Query().Get("month"))
	if !ok {
		from, to, _ = timeutil.MonthStart(timeutil.ThisMonth())
	}
	appr, rows, err := h.svc.Audit(from, to)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"approvers": appr, "excuses": rows})
}
