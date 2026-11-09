package handler

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/repository"
)

// ═══ متابعة الجرد للمراقب — (ع) 10-16 ═══
// «الجرد فارغ، معقولة الموظفين ما دا يجردون؟» — يجردون، بس من «جردي» و«مهامي»
// (جدول InventoryCheck)، وتبويب المراقب چان يقرا جرد الفريق (TeamInventoryCheck)
// الي نادراً ينستعمل. هنا الجرد الحقيقي، ومنو ما جرد اليوم وعنده شغل.

type InventoryFollowHandler struct{ db *sqlx.DB }

func NewInventoryFollowHandler(db *sqlx.DB) *InventoryFollowHandler {
	return &InventoryFollowHandler{db: db}
}

type followCheck struct {
	ID           string    `db:"id" json:"id"`
	EmployeeID   string    `db:"employeeId" json:"employeeId"`
	EmployeeName string    `db:"employeeName" json:"employeeName"`
	Complete     bool      `db:"complete" json:"complete"`
	MissingItems *string   `db:"missingItems" json:"missingItems"`
	CheckedAt    time.Time `db:"checkedAt" json:"checkedAt"`
	BookingID    *string   `db:"bookingId" json:"bookingId"`
	BookingCode  *string   `db:"bookingCode" json:"bookingCode"`
	Today        bool      `db:"today" json:"today"`
}

type notChecked struct {
	EmployeeID string `db:"employeeId" json:"employeeId"`
	Name       string `db:"name" json:"name"`
	Bookings   string `db:"bookings" json:"bookings"`
}

// GET /api/inventory/checks/follow?days=30
func (h *InventoryFollowHandler) List(w http.ResponseWriter, r *http.Request) {
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days <= 0 || days > 120 {
		days = 30
	}
	checks := []followCheck{}
	if err := h.db.Select(&checks, `
		SELECT c.id, c."employeeId", e.name AS "employeeName", c.complete, c."missingItems", c."checkedAt",
		       c."bookingId", b.code AS "bookingCode",
		       baghdad_date(c."checkedAt") = baghdad_today() AS today
		FROM "InventoryCheck" c
		JOIN "Employee" e ON e.id = c."employeeId"
		LEFT JOIN "Booking" b ON b.id = c."bookingId"
		WHERE baghdad_date(c."checkedAt") >= baghdad_today() - $1::int
		ORDER BY c."checkedAt" DESC
		LIMIT 1000`, days); err != nil {
		log.Printf("inventory follow: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الجرود")
		return
	}
	// موظف ميدان عنده حجز موعده اليوم (بالكادر الموحّد) وما جرد اليوم
	missing := []notChecked{}
	if err := h.db.Select(&missing, `
		SELECT e.id AS "employeeId", e.name,
		       string_agg(DISTINCT b.code, '، ') AS bookings
		FROM "Employee" e
		JOIN "Booking" b ON baghdad_date(b."scheduledAt") = baghdad_today()
		     AND b.status NOT IN ('CANCELLED') AND `+repository.BookingCountableSQL("b")+`
		     AND `+repository.BookingWorkedBySQL("b", "e.id")+`
		WHERE e.status = 'ACTIVE' AND (e.role::text IN ('TECHNICIAN', 'TECHNICAL') OR e."isLeader")
		  AND NOT EXISTS (SELECT 1 FROM "InventoryCheck" c WHERE c."employeeId" = e.id
		                  AND baghdad_date(c."checkedAt") = baghdad_today())
		GROUP BY e.id, e.name ORDER BY e.name`); err != nil {
		log.Printf("inventory follow missing: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الجرود")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"days": days, "checks": checks, "notCheckedToday": missing})
}
