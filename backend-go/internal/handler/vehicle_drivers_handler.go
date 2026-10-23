package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/middleware"
)

// ═══ «🪪 السواق والبطاقات» — (ع) 10-17 ═══
// إجازة سوق كل موظف وانتهاؤها، التعهد وتخويل السياقة، وبطاقات البنزين.
type VehicleDriversHandler struct{ db *sqlx.DB }

func NewVehicleDriversHandler(db *sqlx.DB) *VehicleDriversHandler { return &VehicleDriversHandler{db: db} }

type driverRow struct {
	EmployeeID         string     `db:"employeeId" json:"employeeId"`
	Name               string     `db:"name" json:"name"`
	Role               string     `db:"role" json:"role"`
	HasDrivingLicense  bool       `db:"hasDrivingLicense" json:"hasDrivingLicense"`
	LicenseNumber      *string    `db:"licenseNumber" json:"licenseNumber"`
	LicenseExpiry      *time.Time `db:"licenseExpiry" json:"licenseExpiry"`
	Authorized         bool       `db:"authorized" json:"authorized"`
	AuthorizedVehicles *string    `db:"authorizedVehicles" json:"authorizedVehicles"`
	PledgeSignedAt     *time.Time `db:"pledgeSignedAt" json:"pledgeSignedAt"`
	PledgeNote         *string    `db:"pledgeNote" json:"pledgeNote"`
}

// GET /api/vehicle-drivers — كل الموظفين الفعّالين وبيانات سياقتهم.
func (h *VehicleDriversHandler) List(w http.ResponseWriter, r *http.Request) {
	rows := []driverRow{}
	err := h.db.Select(&rows, `
		SELECT e.id AS "employeeId", e.name, e.role::text AS role, e."hasDrivingLicense",
		       d."licenseNumber", d."licenseExpiry", COALESCE(d.authorized, false) AS authorized,
		       d."authorizedVehicles", d."pledgeSignedAt", d."pledgeNote"
		  FROM "Employee" e LEFT JOIN "VehicleDriver" d ON d."employeeId" = e.id
		 WHERE e.status = 'ACTIVE' AND e.role::text NOT IN ('OWNER')
		 ORDER BY COALESCE(d.authorized, false) DESC, e."hasDrivingLicense" DESC, e.name`)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب السواق")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// PUT /api/vehicle-drivers/{employeeId}
func (h *VehicleDriversHandler) Save(w http.ResponseWriter, r *http.Request) {
	var req struct {
		LicenseNumber      string `json:"licenseNumber"`
		LicenseExpiry      string `json:"licenseExpiry"`
		Authorized         bool   `json:"authorized"`
		AuthorizedVehicles string `json:"authorizedVehicles"`
		PledgeSignedAt     string `json:"pledgeSignedAt"`
		PledgeNote         string `json:"pledgeNote"`
	}
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صالحة")
		return
	}
	id := r.PathValue("employeeId")
	_, err := h.db.Exec(`
		INSERT INTO "VehicleDriver" ("employeeId", "licenseNumber", "licenseExpiry", authorized, "authorizedVehicles", "pledgeSignedAt", "pledgeNote", "updatedById", "updatedAt")
		VALUES ($1, NULLIF($2,''), NULLIF($3,'')::date, $4, NULLIF($5,''), NULLIF($6,'')::date, NULLIF($7,''), $8, now())
		ON CONFLICT ("employeeId") DO UPDATE SET "licenseNumber" = EXCLUDED."licenseNumber", "licenseExpiry" = EXCLUDED."licenseExpiry",
			authorized = EXCLUDED.authorized, "authorizedVehicles" = EXCLUDED."authorizedVehicles", "pledgeSignedAt" = EXCLUDED."pledgeSignedAt",
			"pledgeNote" = EXCLUDED."pledgeNote", "updatedById" = EXCLUDED."updatedById", "updatedAt" = now()`,
		id, strings.TrimSpace(req.LicenseNumber), req.LicenseExpiry, req.Authorized, strings.TrimSpace(req.AuthorizedVehicles),
		req.PledgeSignedAt, strings.TrimSpace(req.PledgeNote), middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر الحفظ — تأكد من التواريخ")
		return
	}
	// إجازة السوق المسجّلة تأشّر «عنده إجازة سوق» بملفه
	if strings.TrimSpace(req.LicenseNumber) != "" {
		_, _ = h.db.Exec(`UPDATE "Employee" SET "hasDrivingLicense" = true WHERE id = $1`, id)
	}
	h.List(w, r)
}

type fuelCard struct {
	ID               string   `db:"id" json:"id"`
	Number           string   `db:"number" json:"number"`
	Provider         *string  `db:"provider" json:"provider"`
	VehicleID        *string  `db:"vehicleId" json:"vehicleId"`
	VehicleName      *string  `db:"vehicleName" json:"vehicleName"`
	HolderEmployeeID *string  `db:"holderEmployeeId" json:"holderEmployeeId"`
	HolderName       *string  `db:"holderName" json:"holderName"`
	MonthlyLimit     *float64 `db:"monthlyLimit" json:"monthlyLimit"`
	Active           bool     `db:"active" json:"active"`
	Notes            *string  `db:"notes" json:"notes"`
}

// GET /api/fuel-cards
func (h *VehicleDriversHandler) Cards(w http.ResponseWriter, r *http.Request) {
	rows := []fuelCard{}
	err := h.db.Select(&rows, `
		SELECT c.id, c.number, c.provider, c."vehicleId", COALESCE(v.name || ' ' || COALESCE(v."plateNumber", ''), NULL) AS "vehicleName",
		       c."holderEmployeeId", e.name AS "holderName", c."monthlyLimit", c.active, c.notes
		  FROM "FuelCard" c LEFT JOIN "Vehicle" v ON v.id = c."vehicleId" LEFT JOIN "Employee" e ON e.id = c."holderEmployeeId"
		 ORDER BY c.active DESC, c."createdAt" DESC`)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب البطاقات")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// PUT /api/fuel-cards — إضافة (بلا id) أو تعديل.
func (h *VehicleDriversHandler) SaveCard(w http.ResponseWriter, r *http.Request) {
	var c struct {
		ID               string   `json:"id"`
		Number           string   `json:"number"`
		Provider         string   `json:"provider"`
		VehicleID        string   `json:"vehicleId"`
		HolderEmployeeID string   `json:"holderEmployeeId"`
		MonthlyLimit     *float64 `json:"monthlyLimit"`
		Active           bool     `json:"active"`
		Notes            string   `json:"notes"`
	}
	if err := DecodeJSON(r, &c); err != nil || strings.TrimSpace(c.Number) == "" {
		WriteError(w, http.StatusBadRequest, "رقم البطاقة مطلوب")
		return
	}
	var err error
	if c.ID == "" {
		_, err = h.db.Exec(`INSERT INTO "FuelCard" (id, number, provider, "vehicleId", "holderEmployeeId", "monthlyLimit", active, notes)
			VALUES ($1, $2, NULLIF($3,''), NULLIF($4,''), NULLIF($5,''), $6, true, NULLIF($7,''))`,
			uuid.NewString(), strings.TrimSpace(c.Number), c.Provider, c.VehicleID, c.HolderEmployeeID, c.MonthlyLimit, c.Notes)
	} else {
		_, err = h.db.Exec(`UPDATE "FuelCard" SET number = $2, provider = NULLIF($3,''), "vehicleId" = NULLIF($4,''), "holderEmployeeId" = NULLIF($5,''),
			"monthlyLimit" = $6, active = $7, notes = NULLIF($8,''), "updatedAt" = now() WHERE id = $1`,
			c.ID, strings.TrimSpace(c.Number), c.Provider, c.VehicleID, c.HolderEmployeeID, c.MonthlyLimit, c.Active, c.Notes)
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر حفظ البطاقة")
		return
	}
	h.Cards(w, r)
}
