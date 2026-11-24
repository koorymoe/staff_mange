package handler

import (
	"errors"
	"net/http"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/timeutil"
)

// ═══ توزيع الفلوس — المالك بس (ل) 11-24 دازاي ═══

type MoneySplitHandler struct{ db *sqlx.DB }

func NewMoneySplitHandler(db *sqlx.DB) *MoneySplitHandler { return &MoneySplitHandler{db: db} }

// GET /api/owner/money-split?month=YYYY-MM — النسب + صافي ربح الشركة للشهر
func (h *MoneySplitHandler) Get(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month")
	if month == "" {
		month = timeutil.ThisMonth()
	}
	from, to, ok := timeutil.MonthStart(month)
	if !ok {
		WriteError(w, http.StatusBadRequest, "الشهر غلط")
		return
	}
	profit, err := repository.NewMatrixScoreRepository(h.db).CompanyProfit(from.UTC(), to.UTC())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر حساب الربح")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"split": repository.LoadMoneySplit(h.db), "month": month, "profit": profit})
}

func validMoneySplit(m repository.MoneySplit) error {
	for _, v := range []float64{m.StatsLeaderPct, m.StatsTechPct, m.StatsNoLeaderTechPct, m.RevenueLeaderPct, m.RevenueTechPct} {
		if v < 0 || v > 100 {
			return errors.New("كل نسبة لازم بين ٠ و١٠٠")
		}
	}
	if m.RevenueLeaderPct+m.RevenueTechPct > 100 {
		return errors.New("نسبة الليدر والفني بالإيراد أكثر من ١٠٠٪ — ما يبقى للشركة شي")
	}
	return nil
}

// PUT /api/owner/money-split
func (h *MoneySplitHandler) Put(w http.ResponseWriter, r *http.Request) {
	var m repository.MoneySplit
	if err := DecodeJSON(r, &m); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	if err := validMoneySplit(m); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := repository.SaveMoneySplit(h.db, m, middleware.EmployeeIDFromContext(r)); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	WriteJSON(w, http.StatusOK, repository.LoadMoneySplit(h.db))
}
