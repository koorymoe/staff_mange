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
