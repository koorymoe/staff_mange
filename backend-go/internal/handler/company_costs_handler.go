package handler

import (
	"net/http"
	"time"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/service"
	"staffmange-api/internal/timeutil"
)

// ═══ تكاليف الشركة وأهداف الموظفين — المالك بس (ل) 11-24 ═══

type CompanyCostsHandler struct{ svc *service.CompanyCosts }

func NewCompanyCostsHandler(s *service.CompanyCosts) *CompanyCostsHandler {
	return &CompanyCostsHandler{svc: s}
}

func (h *CompanyCostsHandler) report(w http.ResponseWriter, month string) {
	if month == "" {
		month = timeutil.ThisMonth()
	}
	rep, err := h.svc.Report(month, time.Now())
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, rep)
}

// GET /api/owner/costs?month=YYYY-MM
func (h *CompanyCostsHandler) Get(w http.ResponseWriter, r *http.Request) {
	h.report(w, r.URL.Query().Get("month"))
}

func (h *CompanyCostsHandler) done(w http.ResponseWriter, r *http.Request, err error) {
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.report(w, r.URL.Query().Get("month"))
}

// PUT /api/owner/costs/employees/{id}
func (h *CompanyCostsHandler) PutEmployee(w http.ResponseWriter, r *http.Request) {
	var e service.CostEmployee
	if err := DecodeJSON(r, &e); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	e.EmployeeID = r.PathValue("id")
	h.done(w, r, h.svc.SaveEmployee(e))
}

// DELETE /api/owner/costs/employees/{id}
func (h *CompanyCostsHandler) DeleteEmployee(w http.ResponseWriter, r *http.Request) {
	h.done(w, r, h.svc.RemoveEmployee(r.PathValue("id")))
}

// POST /api/owner/costs/expenses
func (h *CompanyCostsHandler) PostExpense(w http.ResponseWriter, r *http.Request) {
	var x service.CostExpense
	if err := DecodeJSON(r, &x); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	h.done(w, r, h.svc.AddExpense(x, middleware.EmployeeIDFromContext(r)))
}

// DELETE /api/owner/costs/expenses/{id}
func (h *CompanyCostsHandler) DeleteExpense(w http.ResponseWriter, r *http.Request) {
	h.done(w, r, h.svc.RemoveExpense(r.PathValue("id")))
}

// POST /api/owner/costs/downtime
func (h *CompanyCostsHandler) PostDowntime(w http.ResponseWriter, r *http.Request) {
	var d service.DowntimePeriod
	if err := DecodeJSON(r, &d); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	h.done(w, r, h.svc.AddDowntime(d, middleware.EmployeeIDFromContext(r)))
}

// DELETE /api/owner/costs/downtime/{id}
func (h *CompanyCostsHandler) DeleteDowntime(w http.ResponseWriter, r *http.Request) {
	h.done(w, r, h.svc.RemoveDowntime(r.PathValue("id")))
}

// PUT /api/owner/costs/margin {marginPct}
func (h *CompanyCostsHandler) PutMargin(w http.ResponseWriter, r *http.Request) {
	var in struct {
		MarginPct float64 `json:"marginPct"`
	}
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	h.done(w, r, h.svc.SetMargin(in.MarginPct))
}
