package repository

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// ═══ موردين التقنيين ═══

type TechSupplierRepository struct{ db *sqlx.DB }

func NewTechSupplierRepository(db *sqlx.DB) *TechSupplierRepository {
	return &TechSupplierRepository{db: db}
}

type TechSupplier struct {
	ID          string         `db:"id" json:"id"`
	CompanyName string         `db:"companyName" json:"companyName"`
	OwnerName   *string        `db:"ownerName" json:"ownerName"`
	Phone       string         `db:"phone" json:"phone"`
	Address     *string        `db:"address" json:"address"`
	LocationURL *string        `db:"locationUrl" json:"locationUrl"`
	Specialty   *string        `db:"specialty" json:"specialty"`
	Notes       *string        `db:"notes" json:"notes"`
	CreatedBy   *string        `db:"createdBy" json:"createdBy"`
	CreatedAt   time.Time      `db:"createdAt" json:"createdAt"`
	AssignedTo  pq.StringArray `db:"assignedTo" json:"assignedTo"`
}

type TechSupplierIn struct {
	CompanyName string  `json:"companyName"`
	OwnerName   *string `json:"ownerName"`
	Phone       string  `json:"phone"`
	Address     *string `json:"address"`
	LocationURL *string `json:"locationUrl"`
	Specialty   *string `json:"specialty"`
	Notes       *string `json:"notes"`
}

const techSupplierSelect = `SELECT s.id, s."companyName", s."ownerName", s.phone, s.address, s."locationUrl", s.specialty, s.notes,
	ce.name AS "createdBy", s."createdAt",
	COALESCE((SELECT array_agg(a."employeeId") FROM "TechSupplierAccess" a WHERE a."supplierId" = s.id), '{}') AS "assignedTo"
	FROM "TechSupplier" s LEFT JOIN "Employee" ce ON ce.id = s."createdById"`

// CanUse تقني (مهندس/تقني/مسؤول خدمة بالدور أو بجدول ServiceManager) أو عنده الصلاحية.
func (r *TechSupplierRepository) CanUse(employeeID string) bool {
	var ok bool
	_ = r.db.Get(&ok, `SELECT EXISTS (SELECT 1 FROM "Employee" e WHERE e.id = $1 AND (
		e.role::text IN ('ENGINEER', 'TECHNICAL', 'SERVICE_MANAGER', 'ADMIN', 'OWNER')
		OR EXISTS (SELECT 1 FROM "ServiceManager" sm WHERE sm."employeeId" = e.id)
		OR EXISTS (SELECT 1 FROM "EmployeePermission" ep JOIN "Permission" p ON p.id = ep."permissionId"
		           WHERE ep."employeeId" = e.id AND p.name = 'tech_suppliers')))`, employeeID)
	return ok
}

// Mine الموردين الي المدير اختارهم لهالموظف بس.
func (r *TechSupplierRepository) Mine(employeeID string) ([]TechSupplier, error) {
	rows := []TechSupplier{}
	err := r.db.Select(&rows, techSupplierSelect+` WHERE EXISTS (SELECT 1 FROM "TechSupplierAccess" a
		WHERE a."supplierId" = s.id AND a."employeeId" = $1) ORDER BY s."companyName"`, employeeID)
	return rows, err
}

func (r *TechSupplierRepository) All() ([]TechSupplier, error) {
	rows := []TechSupplier{}
	err := r.db.Select(&rows, techSupplierSelect+` ORDER BY s."createdAt" DESC`)
	return rows, err
}

func (r *TechSupplierRepository) Create(in TechSupplierIn, byID string) (string, error) {
	var id string
	err := r.db.Get(&id, `INSERT INTO "TechSupplier" ("companyName", "ownerName", phone, address, "locationUrl", specialty, notes, "createdById")
		VALUES ($1, NULLIF($2, ''), $3, NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''), $8) RETURNING id`,
		in.CompanyName, in.OwnerName, in.Phone, in.Address, in.LocationURL, in.Specialty, in.Notes, byID)
	return id, err
}

func (r *TechSupplierRepository) Update(id string, in TechSupplierIn) error {
	_, err := r.db.Exec(`UPDATE "TechSupplier" SET "companyName" = $2, "ownerName" = NULLIF($3, ''), phone = $4,
		address = NULLIF($5, ''), "locationUrl" = NULLIF($6, ''), specialty = NULLIF($7, ''), notes = NULLIF($8, ''), "updatedAt" = now()
		WHERE id = $1`, id, in.CompanyName, in.OwnerName, in.Phone, in.Address, in.LocationURL, in.Specialty, in.Notes)
	return err
}

func (r *TechSupplierRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM "TechSupplier" WHERE id = $1`, id)
	return err
}

// SetForEmployee يحدد الموردين الي يطلعون لهالتقني (يستبدل الاختيار القديم).
func (r *TechSupplierRepository) SetForEmployee(employeeID string, supplierIDs []string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.Exec(`DELETE FROM "TechSupplierAccess" WHERE "employeeId" = $1`, employeeID); err != nil {
		return err
	}
	if len(supplierIDs) > 0 {
		if _, err := tx.Exec(`INSERT INTO "TechSupplierAccess" ("supplierId", "employeeId")
			SELECT unnest($2::text[]), $1 ON CONFLICT DO NOTHING`, employeeID, pq.Array(supplierIDs)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// Eligible الموظفين الي ممكن يتختارلهم موردين.
type TechPerson struct {
	ID   string `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
	Role string `db:"role" json:"role"`
	Kind string `db:"kind" json:"kind"` // تقني | مسؤول خدمة | بالصلاحية
}

func (r *TechSupplierRepository) Eligible() ([]TechPerson, error) {
	rows := []TechPerson{}
	err := r.db.Select(&rows, `SELECT e.id, e.name, e.role::text AS role,
		CASE WHEN EXISTS (SELECT 1 FROM "ServiceManager" sm WHERE sm."employeeId" = e.id) OR e.role::text = 'SERVICE_MANAGER' THEN 'مسؤول خدمة'
		     WHEN e.role::text IN ('ENGINEER', 'TECHNICAL') THEN 'تقني' ELSE 'بالصلاحية' END AS kind
		FROM "Employee" e WHERE e.status = 'ACTIVE' AND e.role::text NOT IN ('ADMIN', 'OWNER') AND (
		  e.role::text IN ('ENGINEER', 'TECHNICAL', 'SERVICE_MANAGER')
		  OR EXISTS (SELECT 1 FROM "ServiceManager" sm WHERE sm."employeeId" = e.id)
		  OR EXISTS (SELECT 1 FROM "EmployeePermission" ep JOIN "Permission" p ON p.id = ep."permissionId"
		             WHERE ep."employeeId" = e.id AND p.name = 'tech_suppliers'))
		ORDER BY e.name`)
	return rows, err
}
