package handler

import (
	"net/http"
	"time"

	"staffmange-api/internal/repository"
)

// EmployeeLedgerHandler «شنو سوّى هذا الموظف» — اليوم أو الشهر، بالتفصيل (ع 10-16).
type EmployeeLedgerHandler struct {
	repo *repository.EmployeeLedgerRepository
}

func NewEmployeeLedgerHandler(r *repository.EmployeeLedgerRepository) *EmployeeLedgerHandler {
	return &EmployeeLedgerHandler{repo: r}
}

var ledgerLoc = func() *time.Location {
	if l, err := time.LoadLocation("Asia/Baghdad"); err == nil {
		return l
	}
	return time.FixedZone("AST", 3*3600)
}()

// GET /api/ai/employee-ledger?employeeId=&period=day|month&day=YYYY-MM-DD
func (h *EmployeeLedgerHandler) Get(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	id := q.Get("employeeId")
	if id == "" {
		WriteError(w, http.StatusBadRequest, "employeeId مطلوب")
		return
	}
	day, err := time.ParseInLocation("2006-01-02", q.Get("day"), ledgerLoc)
	if err != nil {
		n := time.Now().In(ledgerLoc)
		day = time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, ledgerLoc)
	}
	from, to := day, day.AddDate(0, 0, 1)
	period := "day"
	if q.Get("period") == "month" {
		period = "month"
		from = time.Date(day.Year(), day.Month(), 1, 0, 0, 0, 0, ledgerLoc)
		to = from.AddDate(0, 1, 0)
	}
	// أعمدة القاعدة بلا منطقة = UTC
	f, t := from.UTC(), to.UTC()
	sections := h.repo.Ledger(id, time.Date(f.Year(), f.Month(), f.Day(), f.Hour(), 0, 0, 0, time.UTC), time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, time.UTC))
	total := 0
	for _, s := range sections {
		total += s.Count
	}
	WriteJSON(w, http.StatusOK, map[string]any{
		"period": period, "from": from.Format("2006-01-02"), "to": to.AddDate(0, 0, -1).Format("2006-01-02"),
		"total": total, "sections": sections,
	})
}
