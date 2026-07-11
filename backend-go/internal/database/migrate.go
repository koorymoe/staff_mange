package database

import "github.com/jmoiron/sqlx"

// migrations هي تعديلات بسيطة وآمنة على البنية (ADD COLUMN IF NOT EXISTS فقط) تُطبَّق
// تلقائياً كل مرة يشتغل فيها السيرفر، حتى ما يحتاج أي شخص يشغّل ملفات SQL يدوياً بعد
// سحب تحديث جديد — بس git pull وتشغيل السيرفر كافي.
var migrations = []string{
	`ALTER TABLE "Customer" ADD COLUMN IF NOT EXISTS "mapLatitude" DOUBLE PRECISION`,
	`ALTER TABLE "Customer" ADD COLUMN IF NOT EXISTS "mapLongitude" DOUBLE PRECISION`,
}

func Migrate(db *sqlx.DB) error {
	for _, stmt := range migrations {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}
