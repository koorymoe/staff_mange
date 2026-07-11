package database

import "github.com/jmoiron/sqlx"

// migrations هي تعديلات بسيطة وآمنة على البنية (ADD COLUMN IF NOT EXISTS فقط) تُطبَّق
// تلقائياً كل مرة يشتغل فيها السيرفر، حتى ما يحتاج أي شخص يشغّل ملفات SQL يدوياً بعد
// سحب تحديث جديد — بس git pull وتشغيل السيرفر كافي.
var migrations = []string{
	`ALTER TABLE "Customer" ADD COLUMN IF NOT EXISTS "mapLatitude" DOUBLE PRECISION`,
	`ALTER TABLE "Customer" ADD COLUMN IF NOT EXISTS "mapLongitude" DOUBLE PRECISION`,
	`ALTER TABLE "Employee" ADD COLUMN IF NOT EXISTS "isTrainee" BOOLEAN NOT NULL DEFAULT false`,
	`ALTER TABLE "Employee" ADD COLUMN IF NOT EXISTS "shiftStart" TEXT`,
	`ALTER TABLE "Employee" ADD COLUMN IF NOT EXISTS "shiftEnd" TEXT`,

	// جدول متابعة المركبات: وقود، تنظيف، تبديل زيت — سجل واحد لكل حدث، مع موعد الاستحقاق
	// الجاي (nextDueAt) حتى نقدر نبني جدول مواعيد متكرر من فوقه.
	`CREATE TABLE IF NOT EXISTS "VehicleLog" (
		id TEXT PRIMARY KEY,
		"vehicleId" TEXT NOT NULL REFERENCES "Vehicle"(id) ON DELETE CASCADE,
		type TEXT NOT NULL,
		"performedAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"nextDueAt" TIMESTAMP,
		odometer INTEGER,
		cost DOUBLE PRECISION,
		notes TEXT,
		"recordedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "VehicleLog_vehicleId_idx" ON "VehicleLog"("vehicleId")`,

	// أعطال وأضرار (صدمات) لكل سيارة، مع تحديد المسبب والتكلفة وحالة المعالجة.
	`CREATE TABLE IF NOT EXISTS "VehicleIncident" (
		id TEXT PRIMARY KEY,
		"vehicleId" TEXT NOT NULL REFERENCES "Vehicle"(id) ON DELETE CASCADE,
		type TEXT NOT NULL,
		description TEXT NOT NULL,
		"responsibleEmployeeId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
		cost DOUBLE PRECISION,
		status TEXT NOT NULL DEFAULT 'OPEN',
		"reportedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"resolvedAt" TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "VehicleIncident_vehicleId_idx" ON "VehicleIncident"("vehicleId")`,

	// تقرير حالة شهري لكل سيارة (فيها مشكلة هذا الشهر؟ انعالجت لو لا؟)
	`CREATE TABLE IF NOT EXISTS "VehicleMonthlyStatus" (
		id TEXT PRIMARY KEY,
		"vehicleId" TEXT NOT NULL REFERENCES "Vehicle"(id) ON DELETE CASCADE,
		month TEXT NOT NULL,
		"hasIssue" BOOLEAN NOT NULL DEFAULT false,
		"issueDescription" TEXT,
		resolved BOOLEAN NOT NULL DEFAULT false,
		notes TEXT,
		"recordedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE ("vehicleId", month)
	)`,

	// وحدة الجودة: مشاكل تنفيذية ميدانية + مشاكل رقابية/إدارية، مع تحديد المسؤول.
	`CREATE TABLE IF NOT EXISTS "QualityIssue" (
		id TEXT PRIMARY KEY,
		category TEXT NOT NULL,
		title TEXT NOT NULL,
		description TEXT,
		"responsibleEmployeeId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
		"reportedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
		"bookingId" TEXT REFERENCES "Booking"(id) ON DELETE SET NULL,
		status TEXT NOT NULL DEFAULT 'OPEN',
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"resolvedAt" TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "QualityIssue_status_idx" ON "QualityIssue"(status)`,
}

func Migrate(db *sqlx.DB) error {
	for _, stmt := range migrations {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}
