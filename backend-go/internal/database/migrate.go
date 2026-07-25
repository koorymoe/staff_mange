// Package database يحتوي منطق تهيئة/ترحيل قاعدة البيانات كلها — البنية الأساسية
// (schema_base.go)، التعديلات التراكمية (schema_migrations.go)، وبيانات البذر/
// المساعدات (seed_*.go). Migrate هو نقطة الدخول الوحيدة المستدعاة من الخارج.
package database

import (
	"github.com/jmoiron/sqlx"
)

func Migrate(db *sqlx.DB, ownerUsername, ownerPassword string) error {
	for _, stmt := range baseSchema {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	for _, stmt := range migrations {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	if err := migrateGpsEngineersToSkill(db); err != nil {
		return err
	}
	if err := seedEngineeringSkills(db); err != nil {
		return err
	}
	if err := seedLegacySkills(db); err != nil {
		return err
	}
	if err := seedDefaultSkillForServices(db); err != nil {
		return err
	}
	if err := grantGpsSystemToMonitors(db); err != nil {
		return err
	}
	if err := grantRolePermission(db, "PROCUREMENT_ADMIN", "procurement", "المشتريات"); err != nil {
		return err
	}
	if err := grantRolePermission(db, "PROCUREMENT_ADMIN", "inventory", "جرد الأدوات"); err != nil {
		return err
	}
	if err := seedOwnerAccount(db, ownerUsername, ownerPassword); err != nil {
		return err
	}
	if err := seedKpiCriteria(db); err != nil {
		return err
	}
	return dropAttendanceDailyUniqueConstraint(db)
}
