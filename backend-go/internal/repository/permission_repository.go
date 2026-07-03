package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type PermissionRepository struct {
	db *sqlx.DB
}

func NewPermissionRepository(db *sqlx.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

// EnsureSeeded يضمن وجود كل الصلاحيات الثابتة بالجدول (upsert بالاسم)
func (r *PermissionRepository) EnsureSeeded() error {
	for _, p := range model.DefaultPermissions {
		_, err := r.db.Exec(`
			INSERT INTO "Permission" (id, name, label)
			VALUES (gen_random_uuid()::text, $1, $2)
			ON CONFLICT (name) DO UPDATE SET label = EXCLUDED.label
		`, p.Name, p.Label)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *PermissionRepository) ListAll() ([]model.Permission, error) {
	perms := []model.Permission{}
	err := r.db.Select(&perms, `SELECT id, name, label FROM "Permission" ORDER BY name ASC`)
	return perms, err
}

func (r *PermissionRepository) ListForEmployee(employeeID string) ([]model.Permission, error) {
	perms := []model.Permission{}
	err := r.db.Select(&perms, `
		SELECT p.id, p.name, p.label
		FROM "EmployeePermission" ep
		JOIN "Permission" p ON p.id = ep."permissionId"
		WHERE ep."employeeId" = $1
	`, employeeID)
	return perms, err
}

func (r *PermissionRepository) ReplaceForEmployee(employeeID string, permissionIDs []string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM "EmployeePermission" WHERE "employeeId" = $1`, employeeID); err != nil {
		return err
	}

	for _, permissionID := range permissionIDs {
		_, err := tx.Exec(`
			INSERT INTO "EmployeePermission" (id, "employeeId", "permissionId")
			VALUES (gen_random_uuid()::text, $1, $2)
		`, employeeID, permissionID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PermissionRepository) AddMissingForEmployee(employeeID string, permissionIDs []string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, permissionID := range permissionIDs {
		_, err := tx.Exec(`
			INSERT INTO "EmployeePermission" (id, "employeeId", "permissionId")
			VALUES (gen_random_uuid()::text, $1, $2)
			ON CONFLICT ("employeeId", "permissionId") DO NOTHING
		`, employeeID, permissionID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
