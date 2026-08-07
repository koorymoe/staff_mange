package handler

import (
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
)

// DashboardHandler أرقام اللوحة الرئيسية.
//
// قبله كانت اللوحة تسحب كل الموظفين وكل العملاء وكل الحجوزات (ميغابايتات)
// عشان تعرض أربع أرقام. هذا يعني إن أي موظف يفتح الصفحة الرئيسية ينزل
// على جهازه أرشيف الشركة كامل ويقدر يقراه من F12.
//
// هذا المسار يرجّع الأرقام بس — بدون ولا سطر بيانات.
type DashboardHandler struct {
	db *sqlx.DB
}

func NewDashboardHandler(db *sqlx.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

type dashboardSummary struct {
	EmployeeCount  int `db:"employeeCount" json:"employeeCount"`
	CustomerCount  int `db:"customerCount" json:"customerCount"`
	BookingCount   int `db:"bookingCount" json:"bookingCount"`
	GpsDeviceCount int `db:"gpsDeviceCount" json:"gpsDeviceCount"`
}

// GET /api/dashboard/summary
func (h *DashboardHandler) Summary(w http.ResponseWriter, r *http.Request) {
	var s dashboardSummary
	err := h.db.Get(&s, `
		SELECT
			(SELECT COUNT(*) FROM "Employee" WHERE status = 'ACTIVE')  AS "employeeCount",
			(SELECT COUNT(*) FROM "Customer")                          AS "customerCount",
			(SELECT COUNT(*) FROM "Booking")                           AS "bookingCount",
			(SELECT COUNT(*) FROM "GpsDeviceRequest")                  AS "gpsDeviceCount"`)
	if err != nil {
		log.Printf("dashboard summary: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب ملخص اللوحة")
		return
	}
	WriteJSON(w, http.StatusOK, s)
}

// ═══ ملخّص المحاسب والمراقب ═══
//
// لوحة المحاسب ولوحة المراقب جانت تنزّل **كل الحجوزات المنجزة من يوم ما
// انفتح النظام** — بكل زبائنها وخدماتها وتعييناتها وسلة موادها — بس
// حتى تجمع تسع أرقام بالمتصفح. يعني كل ما تخلص شغلة جديدة، صفحتهم
// الرئيسية تصير أثقل، والثقل ما يوكف لأن الأرشيف ما ينقص.
//
// الجمع مكانه قاعدة البيانات. هذا المسار يرجّع نفس الأرقام بالضبط —
// مجاميع كاملة مو مقصوصة — بعشرات البايتات بدل ميغابايتات.
type financeSummary struct {
	CompletedCount    int     `db:"completedCount" json:"completedCount"`
	UnverifiedCount   int     `db:"unverifiedCount" json:"unverifiedCount"`
	VerifiedCount     int     `db:"verifiedCount" json:"verifiedCount"`
	TodayCompleted    int     `db:"todayCompleted" json:"todayCompleted"`
	PendingCount      int     `db:"pendingCount" json:"pendingCount"`
	ConfirmedCount    int     `db:"confirmedCount" json:"confirmedCount"`
	InProgressCount   int     `db:"inProgressCount" json:"inProgressCount"`
	ActiveCrewCount   int     `db:"activeCrewCount" json:"activeCrewCount"`
	TotalCollected    float64 `db:"totalCollected" json:"totalCollected"`
	TotalQuoted       float64 `db:"totalQuoted" json:"totalQuoted"`
	TotalCartValue    float64 `db:"totalCartValue" json:"totalCartValue"`
	PendingExpenses   int     `db:"pendingExpenses" json:"pendingExpenses"`
	ApprovedExpenses  int     `db:"approvedExpenses" json:"approvedExpenses"`
	TotalExpenseValue float64 `db:"totalExpenseValue" json:"totalExpenseValue"`
}

// GET /api/dashboard/finance-summary
func (h *DashboardHandler) FinanceSummary(w http.ResponseWriter, r *http.Request) {
	var s financeSummary
	err := h.db.Get(&s, `
		SELECT
			COUNT(*) FILTER (WHERE b.status = 'COMPLETED')                              AS "completedCount",
			COUNT(*) FILTER (WHERE b.status = 'COMPLETED' AND NOT b."amountVerified")   AS "unverifiedCount",
			COUNT(*) FILTER (WHERE b.status = 'COMPLETED' AND b."amountVerified")       AS "verifiedCount",
			COUNT(*) FILTER (WHERE b.status = 'COMPLETED'
			                 AND b."completedAt"::date = CURRENT_DATE)                  AS "todayCompleted",
			COUNT(*) FILTER (WHERE b.status = 'PENDING')                                AS "pendingCount",
			COUNT(*) FILTER (WHERE b.status = 'CONFIRMED')                              AS "confirmedCount",
			COUNT(*) FILTER (WHERE b.status = 'IN_PROGRESS')                            AS "inProgressCount",
			(SELECT COUNT(DISTINCT ba."employeeId") FROM "BookingAssignment" ba
			  JOIN "Booking" b2 ON b2.id = ba."bookingId"
			  WHERE b2.status = 'IN_PROGRESS')                                          AS "activeCrewCount",
			COALESCE(SUM(COALESCE(b."amountCollected",0) + COALESCE(b."advancePaid",0))
			         FILTER (WHERE b.status = 'COMPLETED'), 0)                          AS "totalCollected",
			COALESCE(SUM(COALESCE(b."quotedPrice",0))
			         FILTER (WHERE b.status = 'COMPLETED'), 0)                          AS "totalQuoted",
			(SELECT COALESCE(SUM(ci."totalPrice"), 0) FROM "CartItem" ci
			  JOIN "Booking" b3 ON b3.id = ci."bookingId"
			  WHERE b3.status = 'COMPLETED')                                            AS "totalCartValue",
			(SELECT COUNT(*) FROM "Expense" WHERE status = 'PENDING')                   AS "pendingExpenses",
			(SELECT COUNT(*) FROM "Expense" WHERE status = 'APPROVED')                  AS "approvedExpenses",
			(SELECT COALESCE(SUM(amount),0) FROM "Expense" WHERE status = 'APPROVED')   AS "totalExpenseValue"
		FROM "Booking" b`)
	if err != nil {
		log.Printf("finance summary: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الملخص المالي")
		return
	}
	WriteJSON(w, http.StatusOK, s)
}
