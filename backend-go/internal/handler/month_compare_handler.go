package handler

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/repository"
)

// ═══ مقارنة الأشهر (قرار (ع) 10-10) ═══
// «شهر التاسع جان بي ٢٠ مشروع و٢٠٠ حجز والمبالغ ٤٠٠ مليون، وشهر العاشر…».
// كل رقم بنفس تعريفه بباقي النظام: الإيراد من revenue_sql.go، والحجوزات
// الي تنعدّ من BookingCountableSQL.
type MonthCompareHandler struct{ db *sqlx.DB }

func NewMonthCompareHandler(db *sqlx.DB) *MonthCompareHandler { return &MonthCompareHandler{db: db} }

var monthRe = regexp.MustCompile(`^\d{4}-\d{2}$`)

// GET /api/stats-management/month-compare?months=2026-09,2026-10
func (h *MonthCompareHandler) Compare(w http.ResponseWriter, r *http.Request) {
	months := []string{}
	for _, m := range strings.Split(r.URL.Query().Get("months"), ",") {
		m = strings.TrimSpace(m)
		if monthRe.MatchString(m) && len(months) < 12 {
			months = append(months, m)
		}
	}
	if len(months) == 0 {
		WriteError(w, http.StatusBadRequest, "اختار الأشهر")
		return
	}
	out := make([]repository.MonthStats, 0, len(months))
	for _, m := range months {
		st, err := repository.MonthStatsFor(h.db, m)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "تعذر حساب الشهر "+m)
			return
		}
		out = append(out, *st)
	}
	WriteJSON(w, http.StatusOK, out)
}
