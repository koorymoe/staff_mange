package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

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

// ListEmployeesWithPermission يرجّع الموظفين النشطين الي عندهم صلاحية معيّنة،
// بالإضافة للأدوار المذكورة بـalsoRoles (لأن الدور يعطي الوصول بغض النظر عن
// جدول الصلاحيات) — يُستخدم لتعبئة قوائم منسدلة "مين المسؤول/مين يسوي الكشف".
func (r *PermissionRepository) ListEmployeesWithPermission(permissionName string, alsoRoles []string) ([]model.EmployeeBrief, error) {
	employees := []model.EmployeeBrief{}
	if alsoRoles == nil {
		alsoRoles = []string{}
	}
	err := r.db.Select(&employees, `
		SELECT DISTINCT e.id, e.name, e.position
		FROM "Employee" e
		LEFT JOIN "EmployeePermission" ep ON ep."employeeId" = e.id
		LEFT JOIN "Permission" p ON p.id = ep."permissionId"
		WHERE e.status = 'ACTIVE' AND (p.name = $1 OR e.role = ANY($2))
		ORDER BY e.name
	`, permissionName, pq.Array(alsoRoles))
	return employees, err
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

// HasPermission يتحقق مباشرة إذا كان الموظف عنده صلاحية معينة بالاسم — يُستخدم
// داخل الهاندلر/السيرفس لفحوصات دقيقة إضافية غير الفحص العام بالميدلوير (مثل
// نوع طلب المشتريات).
func (r *PermissionRepository) HasPermission(employeeID, permissionName string) (bool, error) {
	var count int
	err := r.db.Get(&count, `
		SELECT COUNT(*)
		FROM "EmployeePermission" ep
		JOIN "Permission" p ON p.id = ep."permissionId"
		WHERE ep."employeeId" = $1 AND p.name = $2
	`, employeeID, permissionName)
	if err != nil {
		return false, err
	}
	return count > 0, nil
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
