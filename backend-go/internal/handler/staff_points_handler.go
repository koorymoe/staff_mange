package handler

import (
	"net/http"
	"regexp"
	"time"

	"staffmange-api/internal/service"
)

// ═══ 🏅 رصيد نقاط الشهر — قرار (ع) 10-11 ═══
type StaffPointsHandler struct{ svc *service.StaffPointsService }

func NewStaffPointsHandler(svc *service.StaffPointsService) *StaffPointsHandler {
	return &StaffPointsHandler{svc: svc}
}

var pointsMonthRe = regexp.MustCompile(`^\d{4}-\d{2}$`)

// GET /api/staff-points?month=YYYY-MM
func (h *StaffPointsHandler) Month(w http.ResponseWriter, r *http.Request) {
	m := r.URL.Query().Get("month")
	if !pointsMonthRe.MatchString(m) {
		loc, _ := time.LoadLocation("Asia/Baghdad")
		if loc == nil {
			loc = time.UTC
		}
		m = time.Now().In(loc).Format("2006-01")
	}
	rows, err := h.svc.Month(m)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر حساب الرصيد")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"month": m, "staff": rows})
}
