package handler

import (
	"net/http"
	"regexp"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/middleware"
)

// ═══ 🧽 غسل السيارات — شاشة عامل الغسل (قرار (ع) 10-09) ═══
type CarWashHandler struct{ db *sqlx.DB }

func NewCarWashHandler(db *sqlx.DB) *CarWashHandler { return &CarWashHandler{db: db} }

var washDateRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// يوم بغداد (+٣ ثابت) — مو يوم غرينتش
func baghdadToday() string { return time.Now().UTC().Add(3 * time.Hour).Format("2006-01-02") }

type carWashRow struct {
	VehicleID   string     `db:"vehicleId" json:"vehicleId"`
	Name        string     `db:"name" json:"name"`
	PlateNumber string     `db:"plateNumber" json:"plateNumber"`
	WashedAt    *time.Time `db:"washedAt" json:"washedAt"`
	WashedBy    *string    `db:"washedBy" json:"washedBy"`
	WashedByID  *string    `db:"washedById" json:"washedById"`
}

// GET /api/car-wash?date=YYYY-MM-DD — السيارات الفعّالة وغسلها بذاك اليوم
func (h *CarWashHandler) List(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	if !washDateRe.MatchString(date) {
		date = baghdadToday()
	}
	rows := []carWashRow{}
	err := h.db.Select(&rows, `
		SELECT v.id AS "vehicleId", v.name, v."plateNumber", l."washedAt", e.name AS "washedBy", l."employeeId" AS "washedById"
		FROM "Vehicle" v
		LEFT JOIN "VehicleWashLog" l ON l."vehicleId" = v.id AND l."washDate" = $1::date
		LEFT JOIN "Employee" e ON e.id = l."employeeId"
		WHERE v."isActive"
		ORDER BY v."createdAt", v.name`, date)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب السيارات")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"date": date, "vehicles": rows})
}

// POST /api/car-wash/{vehicleId} — «تم الغسل» لليوم
func (h *CarWashHandler) Mark(w http.ResponseWriter, r *http.Request) {
	_, err := h.db.Exec(`
		INSERT INTO "VehicleWashLog" (id, "vehicleId", "employeeId", "washDate")
		SELECT gen_random_uuid()::text, v.id, $2, $3::date FROM "Vehicle" v WHERE v.id = $1 AND v."isActive"
		ON CONFLICT ("vehicleId", "washDate") DO NOTHING`,
		r.PathValue("vehicleId"), middleware.EmployeeIDFromContext(r), baghdadToday())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر التسجيل")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// DELETE /api/car-wash/{vehicleId} — تراجع عن ضغطة غلط، بس لنفس اليوم وبس للي سجّلها
// (أو صاحب إدارة المركبات/المدير)
func (h *CarWashHandler) Unmark(w http.ResponseWriter, r *http.Request) {
	role := middleware.RoleFromContext(r)
	q := `DELETE FROM "VehicleWashLog" WHERE "vehicleId" = $1 AND "washDate" = $2::date AND ("employeeId" = $3 OR $4)`
	if _, err := h.db.Exec(q, r.PathValue("vehicleId"), baghdadToday(), middleware.EmployeeIDFromContext(r), role != "CAR_WASHER"); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر التراجع")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
