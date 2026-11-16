package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// ApplyMigrationForTest ينفّذ SQL ترحيل واحد بالاسم — للاختبارات الحية بباقي الحزم، حتى ما
// تنسخ الـSQL. الترحيلات كلها idempotent (IF NOT EXISTS / OR REPLACE).
func ApplyMigrationForTest(db *sqlx.DB, version string) error {
	for _, m := range versionedMigrations() {
		if m.Version == version {
			_, err := db.Exec(m.SQL)
			return err
		}
	}
	return fmt.Errorf("ترحيل %s مو موجود", version)
}
