package repository

import (
	"github.com/jmoiron/sqlx"
	"staffmange-api/internal/errlog"

	"staffmange-api/internal/model"
)

// OrgEffectiveRoles الأدوار الفعلية لموظف من منصبه بالهيكل (دازاي 11-01).
// فشل القراءة يرجّع فاضي — يضيّق ما يوسّع.
func OrgEffectiveRoles(db *sqlx.DB, employeeID string) []string {
	roles := []string{}
	if employeeID == "" {
		return roles
	}
	errlog.Err(db.Select(&roles, model.OrgRolesOfEmployeeSQL(), employeeID))
	return roles
}

// HasOrgRole هل للموظف دور فعلي من منصبه ضمن roles؟
func HasOrgRole(db *sqlx.DB, employeeID string, roles ...string) bool {
	for _, or := range OrgEffectiveRoles(db, employeeID) {
		for _, r := range roles {
			if or == r {
				return true
			}
		}
	}
	return false
}
