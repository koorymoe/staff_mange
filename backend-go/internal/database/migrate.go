// Package database يحتوي منطق تهيئة/ترحيل قاعدة البيانات كلها — البنية الأساسية
// (schema_base.go)، التعديلات التراكمية (schema_migrations.go)، وبيانات البذر/
// المساعدات (seed_*.go). Migrate هو نقطة الدخول الوحيدة المستدعاة من الخارج.
package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
)

// runVersionedMigrations ينفّذ كل ترحيل مو مطبّق بعد (مو موجود بجدول "SchemaMigration")
// مرة واحدة فقط، كل واحد جوة معاملة (transaction) خاصة فيه: لو نجح SQL الترحيل يسجَّل
// رقمه بنفس المعاملة ثم Commit، ولو فشل يصير Rollback ويتوقف كل الترحيل فوراً (ما نكمل
// بعد أول فشل). الترحيلات المطبّقة أصلاً تُتخطى تماماً بدون أي تنفيذ SQL — هذا هو الفرق
// الجوهري عن النظام القديم اللي كان يعيد تنفيذ كل الترحيلات (~216 عبارة) بكل إقلاع.
//
// أول تشغيل لهذا الكود على قاعدة بيانات فيها البنية الكاملة أصلاً (بدون جدول
// "SchemaMigration" سابقاً) هو حالة آمنة تماماً: كل عبارات baseSchema/migrations
// عبارة عن IF NOT EXISTS/ON CONFLICT أصلاً، فإعادة تنفيذها هنا "لأول مرة تسجيل" ما
// تغيّر شي فعلياً — بس تسجّلها كمطبّقة حتى لا تُعاد بعدها.
func runVersionedMigrations(db *sqlx.DB) error {
	// جدول "SchemaMigration" نفسه جزء من migration "0001_initial_schema" (baseSchema)،
	// لكن لازم نتأكد من وجوده هنا أول شي عشان نقدر نستعلم عنه حتى قبل أي commit.
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS "SchemaMigration" (
		version TEXT PRIMARY KEY,
		"appliedAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return fmt.Errorf("create SchemaMigration table: %w", err)
	}

	for _, m := range versionedMigrations() {
		var alreadyApplied int
		err := db.Get(&alreadyApplied, `SELECT 1 FROM "SchemaMigration" WHERE version = $1`, m.Version)
		if err == nil {
			continue // مطبّق أصلاً — تخطّي كامل بدون تنفيذ أي SQL
		}
		if err != sql.ErrNoRows {
			return fmt.Errorf("check migration %s: %w", m.Version, err)
		}

		tx, err := db.Beginx()
		if err != nil {
			return fmt.Errorf("begin tx for migration %s: %w", m.Version, err)
		}
		if _, err := tx.Exec(m.SQL); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", m.Version, err)
		}
		if _, err := tx.Exec(`INSERT INTO "SchemaMigration" (version) VALUES ($1)`, m.Version); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %s: %w", m.Version, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", m.Version, err)
		}
		log.Printf("[migrate] applied %s", m.Version)
	}
	return nil
}

func Migrate(db *sqlx.DB, ownerUsername, ownerPassword string) error {
	if err := runVersionedMigrations(db); err != nil {
		return err
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
	if err := seedDecorSkills(db); err != nil {
		return err
	}
	if err := grantGpsSystemToMonitors(db); err != nil {
		return err
	}
	if err := grantLeaderBasketToLeaders(db); err != nil {
		return err
	}
	if err := grantRolePermission(db, "PROCUREMENT_ADMIN", "procurement", "المشتريات"); err != nil {
		return err
	}
	if err := grantRolePermission(db, "PROCUREMENT_ADMIN", "inventory", "جرد الأدوات"); err != nil {
		return err
	}
	// الفني هو الي يكتشف المادة الناقصة وهو بموقع الشغل — فزر «اطلب مادة
	// ناقصة» بشاشة مهامه لازم يشتغل من غير ما أحد ينطيه صلاحية يدوياً.
	// قبل هيچي الزر كان يوصله لشاشة يرفضها السيرفر: «لا تملك صلاحية
	// تقديم هذا النوع من الطلبات». تنبقى صلاحية عادية تنسحب لأي موظف
	// بالإرادة، بس الافتراضي إنها موجودة.
	if err := grantRolePermission(db, "TECHNICIAN", "procurement", "المشتريات"); err != nil {
		return err
	}
	if err := grantRolePermission(db, "TECHNICIAN", "procurement_customer", "طلب منتج للزبون"); err != nil {
		return err
	}
	// الطلب اليدوي: الإداري والمحاسب والليدر يكدرون يطلبون مادة بالتفصيل.
	// (الليدرات ينمنحون بالاستعلام الي بعده لأن الليدر صفة مو دور.)
	for _, role := range []string{"ADMIN", "OWNER", "FINANCE", "HR_COORDINATOR", "PROJECT_MANAGER"} {
		if err := grantRolePermission(db, role, "procurement_manual", "طلب مادة يدوي (إداري/ليدر)"); err != nil {
			return err
		}
	}
	if _, err := db.Exec(`
		INSERT INTO "EmployeePermission" (id, "employeeId", "permissionId")
		SELECT gen_random_uuid()::text, e.id, p.id
		FROM "Employee" e CROSS JOIN "Permission" p
		WHERE p.name = 'procurement_manual' AND e."isLeader" = true
		  AND NOT EXISTS (
			SELECT 1 FROM "EmployeePermission" ep
			WHERE ep."employeeId" = e.id AND ep."permissionId" = p.id
		  )`); err != nil {
		return err
	}
	// التخريج وتعويض الدوار — للمحاسب.
	if err := grantRolePermission(db, "FINANCE", "fund_discharge", "تخريج المواد وتعويض الدوار"); err != nil {
		return err
	}
	// الاتصال بالزبون على الشكوى — مهندس الجودة والمراقب افتراضياً،
	// والمدير يكدر ينطيها لأي موظف ثاني من شاشة الصلاحيات.
	for _, role := range []string{"QUALITY_ENGINEER", "MONITOR"} {
		if err := grantRolePermission(db, role, "complaint_contact", "الاتصال بالزبون ومتابعة الشكوى"); err != nil {
			return err
		}
	}
	if err := seedOwnerAccount(db, ownerUsername, ownerPassword); err != nil {
		return err
	}
	if err := seedKpiCriteria(db); err != nil {
		return err
	}
	if err := seedSystemPriceCatalog(db); err != nil {
		return err
	}
	if err := seedMaterialArchive(db); err != nil {
		return err
	}
	return dropAttendanceDailyUniqueConstraint(db)
}
