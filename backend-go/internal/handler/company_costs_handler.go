package handler

import (
	"net/http"
	"strconv"
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
		month = timeutil.ThisPeriod()
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

// GET /api/me/money-target — الموظف يشوف هدفه هو بس (null إذا مو «يدخّل فلوس»)
func (h *CompanyCostsHandler) Mine(w http.ResponseWriter, r *http.Request) {
	t, err := h.svc.MyTarget(middleware.EmployeeIDFromContext(r), time.Now())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"target": t})
}

// MoneyRankRow ترتيب المدراء — بلا رواتب (الرواتب للمالك بس بصفحة التكاليف).
type MoneyRankRow struct {
	Rank        int     `json:"rank"`
	Name        string  `json:"name"`
	Shift       *string `json:"shift"`
	Target      float64 `json:"target"`
	Daily       float64 `json:"daily"`
	Brought     float64 `json:"brought"`
	Remaining   float64 `json:"remaining"`
	Forecast    float64 `json:"forecast"`
	Pct         float64 `json:"pct"`
	DailyNeeded float64 `json:"dailyNeeded"`
	Errors      int     `json:"errors"`
	Fixes       int     `json:"fixes"`
}

// GET /api/employee-stats/money-ranking?month= — (ل) 11-24 دازاي: الترتيب بصفحة الإحصائيات للمدراء
func (h *CompanyCostsHandler) Ranking(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month")
	if month == "" {
		month = timeutil.ThisPeriod()
	}
	rep, err := h.svc.Report(month, time.Now())
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	rows := make([]MoneyRankRow, 0, len(rep.Targets))
	for _, t := range rep.Targets {
		rows = append(rows, MoneyRankRow{Rank: t.Rank, Name: t.Name, Shift: t.Shift, Target: t.Target, Daily: t.Daily, Brought: t.Brought,
			Remaining: t.Remaining, Forecast: t.Forecast, Pct: t.Pct, DailyNeeded: t.DailyNeeded, Errors: t.Errors, Fixes: t.Fixes})
	}
	WriteJSON(w, http.StatusOK, map[string]any{"month": month, "rows": rows, "downtimeDays": rep.DowntimeDays})
}

// GET /api/employee-stats/fragility?days=90 — (ل) 11-25 دازاي: مؤشر هشاشة النظام
func (h *CompanyCostsHandler) Fragility(w http.ResponseWriter, r *http.Request) {
	days := 90
	if v := r.URL.Query().Get("days"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			days = n
		}
	}
	rep, err := service.Fragility(h.svc.DB(), days, time.Now())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	WriteJSON(w, http.StatusOK, rep)
}
