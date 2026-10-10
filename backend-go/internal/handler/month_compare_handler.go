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

type MonthStats struct {
	Month             string  `db:"month" json:"month"`
	BookingsCreated   int     `db:"bookingsCreated" json:"bookingsCreated"`
	BookingsCompleted int     `db:"bookingsCompleted" json:"bookingsCompleted"`
	BookingsCancelled int     `db:"bookingsCancelled" json:"bookingsCancelled"`
	InternalWorks     int     `db:"internalWorks" json:"internalWorks"`
	ProjectsCreated   int     `db:"projectsCreated" json:"projectsCreated"`
	NewCustomers      int     `db:"newCustomers" json:"newCustomers"`
	Complaints        int     `db:"complaints" json:"complaints"`
	BookingsRevenue   float64 `db:"bookingsRevenue" json:"bookingsRevenue"`
	ProjectsRevenue   float64 `db:"projectsRevenue" json:"projectsRevenue"`
	Expenses          float64 `db:"expenses" json:"expenses"`
}

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
	out := make([]MonthStats, 0, len(months))
	for _, m := range months {
		var st MonthStats
		err := h.db.Get(&st, `
			SELECT $1::text AS month,
			  (SELECT COUNT(*) FROM "Booking" b WHERE `+repository.BookingCountableSQL("b")+`
			     AND to_char(baghdad_date(b."createdAt"), 'YYYY-MM') = $1) AS "bookingsCreated",
			  (SELECT COUNT(*) FROM "Booking" b WHERE `+repository.BookingCountableSQL("b")+`
			     AND b.status::text IN ('COMPLETED','PARTIAL') AND b."bookingType" IS DISTINCT FROM 'INTERNAL'
			     AND to_char(baghdad_date(b."completedAt"), 'YYYY-MM') = $1) AS "bookingsCompleted",
			  (SELECT COUNT(*) FROM "Booking" b WHERE `+repository.BookingCountableSQL("b")+`
			     AND b.status::text = 'CANCELLED' AND to_char(baghdad_date(b."createdAt"), 'YYYY-MM') = $1) AS "bookingsCancelled",
			  (SELECT COUNT(*) FROM "Booking" b WHERE `+repository.BookingCountableSQL("b")+`
			     AND b."bookingType" = 'INTERNAL' AND b.status::text = 'COMPLETED'
			     AND to_char(baghdad_date(b."completedAt"), 'YYYY-MM') = $1) AS "internalWorks",
			  (SELECT COUNT(*) FROM "Project" p WHERE to_char(baghdad_date(p."createdAt"), 'YYYY-MM') = $1) AS "projectsCreated",
			  (SELECT COUNT(*) FROM "Customer" c WHERE to_char(baghdad_date(c."createdAt"), 'YYYY-MM') = $1) AS "newCustomers",
			  (SELECT COUNT(*) FROM "Complaint" q WHERE to_char(baghdad_date(q."createdAt"), 'YYYY-MM') = $1) AS "complaints",
			  (SELECT COALESCE(SUM(`+repository.RevenueAmountSQL("b")+`), 0) FROM "Booking" b WHERE `+repository.BookingCountableSQL("b")+`
			     AND `+repository.RevenueBookingSQL("b")+` AND to_char(baghdad_date(b."completedAt"), 'YYYY-MM') = $1) AS "bookingsRevenue",
			  (SELECT COALESCE(SUM(pp.amount), 0) FROM "ProjectPayment" pp WHERE pp."cancelledAt" IS NULL
			     AND to_char(baghdad_date(pp."paidAt"), 'YYYY-MM') = $1) AS "projectsRevenue",
			  (SELECT COALESCE(SUM(e.amount), 0) FROM "Expense" e WHERE e.status = 'APPROVED'
			     AND to_char(baghdad_date(e."createdAt"), 'YYYY-MM') = $1) AS "expenses"`, m)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "تعذر حساب الشهر "+m)
			return
		}
		out = append(out, st)
	}
	WriteJSON(w, http.StatusOK, out)
}
