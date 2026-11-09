package handler

import (
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/authz"
	"staffmange-api/internal/middleware"
	"staffmange-api/internal/service"
)

// RoutePlanHandler (دازاي 11-06، فكرة ٢): «رتّب مسار اليوم» — حجوزات الموظف بيوم واحد
// مرتبة بأقصر مسار من موقع الشركة، ويا رابط خرائط Google.
type RoutePlanHandler struct{ db *sqlx.DB }

func NewRoutePlanHandler(db *sqlx.DB) *RoutePlanHandler { return &RoutePlanHandler{db: db} }

type routeRow struct {
	ID       string   `db:"id"`
	Code     string   `db:"code"`
	Customer *string  `db:"customer"`
	Lat      *float64 `db:"lat"`
	Lng      *float64 `db:"lng"`
}

// GET /api/route-plan?date=YYYY-MM-DD&employeeId= — الموظف لنفسه، والمشرف لأي موظف.
func (h *RoutePlanHandler) Plan(w http.ResponseWriter, r *http.Request) {
	emp := r.URL.Query().Get("employeeId")
	if emp == "" {
		emp = middleware.EmployeeIDFromContext(r)
	}
	if !selfOrSupervisor(r, emp) {
		authz.Deny(w, r, authz.EmployeeRecords)
		return
	}
	day := r.URL.Query().Get("date")
	if _, err := time.Parse("2006-01-02", day); err != nil {
		day = time.Now().In(baghdadLoc()).Format("2006-01-02")
	}
	rows := []routeRow{}
	if err := h.db.Select(&rows, `
		SELECT b.id, b.code, c.name AS customer,
		       COALESCE(b."mapLatitude", c."mapLatitude", c.lat) AS lat,
		       COALESCE(b."mapLongitude", c."mapLongitude", c.lng) AS lng
		FROM "Booking" b
		LEFT JOIN "Customer" c ON c.id = b."customerId"
		WHERE baghdad_date(b."scheduledAt") = $2::date
		  AND b.status::text NOT IN ('COMPLETED','CANCELLED') AND b."archivedAt" IS NULL
		  AND $1 IN (SELECT booking_crew_ids(b.id))
		ORDER BY b."scheduledAt"`, emp, day); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب حجوزات اليوم")
		return
	}
	var start service.RoutePoint
	if err := h.db.Get(&start, `SELECT 'company' AS id, 'الشركة' AS label, lat, lng FROM "CompanyLocation" WHERE id = 1`); err != nil {
		WriteError(w, http.StatusConflict, "موقع الشركة مو محدد — حدده من إعدادات الحضور")
		return
	}
	stops, noLoc := []service.RoutePoint{}, []map[string]string{}
	for _, b := range rows {
		label := b.Code
		if b.Customer != nil {
			label += " — " + *b.Customer
		}
		if b.Lat == nil || b.Lng == nil || (*b.Lat == 0 && *b.Lng == 0) {
			noLoc = append(noLoc, map[string]string{"id": b.ID, "label": label})
			continue
		}
		stops = append(stops, service.RoutePoint{ID: b.ID, Label: label, Lat: *b.Lat, Lng: *b.Lng})
	}
	WriteJSON(w, http.StatusOK, map[string]any{"date": day, "plan": service.PlanRoute(start, stops), "noLocation": noLoc})
}
