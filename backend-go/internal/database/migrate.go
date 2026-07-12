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

	// وقت تجهيز المواد (يحدده تيم ليدر الفريق) ومدة استجابة الفنيين بعده لحد ما يبدون
	// الشغل فعلياً — حتى نعرف مين ضيّع وقت بدل ما يتحرك مباشرة.
	`ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "materialsReadyAt" TIMESTAMP`,
	`ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "materialsReadyById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL`,
	`ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "responseMinutes" INTEGER`,

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

	// جرد الأدوات اليومي: الفني يؤكد جرد عدته الخاصة قبل ما يطلع للحجز، والإداري يشوف
	// نتائج كل الفنيين بيوم واحد حتى يوفر البديل بحال اكو نقص.
	`CREATE TABLE IF NOT EXISTS "InventoryCheck" (
		id TEXT PRIMARY KEY,
		"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
		complete BOOLEAN NOT NULL,
		"missingItems" TEXT,
		"checkedAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "InventoryCheck_employeeId_idx" ON "InventoryCheck"("employeeId")`,
	`CREATE INDEX IF NOT EXISTS "InventoryCheck_checkedAt_idx" ON "InventoryCheck"("checkedAt")`,

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

	// دور GPS_ENGINEER اتلغى — تركيب GPS صار مهارة عادية (مثل باقي الخدمات) يمنحها
	// الأدمن لأي فني عادي بدل ما يكون دور مستقل. "أبو الجي بي اس" (GPS_ADMIN) هو
	// الدور الوحيد الخاص بـGPS اللي بقى، ويرتب موعد الزبون لطلبات GPS الجديدة.
	`ALTER TABLE "GpsDeviceRequest" ADD COLUMN IF NOT EXISTS "scheduledAt" TIMESTAMP`,
	`ALTER TABLE "GpsDeviceRequest" ADD COLUMN IF NOT EXISTS "assignedTechnicianId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL`,
}

func Migrate(db *sqlx.DB) error {
	for _, stmt := range migrations {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return migrateGpsEngineersToSkill(db)
}

// migrateGpsEngineersToSkill يحول أي موظف لسه بدور GPS_ENGINEER (القديم) إلى TECHNICIAN
// عادي، ويمنحه كل مهارات تركيب GPS تلقائياً حتى يضل يقدر يستلم نفس شغله بالضبط.
func migrateGpsEngineersToSkill(db *sqlx.DB) error {
	var ids []string
	if err := db.Select(&ids, `SELECT id FROM "Employee" WHERE role = 'GPS_ENGINEER'`); err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}

	var gpsSkillIDs []string
	if err := db.Select(&gpsSkillIDs, `
		SELECT sk.id FROM "Skill" sk JOIN "Service" sv ON sv.id = sk."serviceId" WHERE sv.name = 'GPS'
	`); err != nil {
		return err
	}

	for _, empID := range ids {
		if _, err := db.Exec(`UPDATE "Employee" SET role = 'TECHNICIAN' WHERE id = $1`, empID); err != nil {
			return err
		}
		for _, skillID := range gpsSkillIDs {
			if _, err := db.Exec(`
				INSERT INTO "EmployeeSkill" (id, "employeeId", "skillId", "canPerform")
				VALUES (gen_random_uuid()::text, $1, $2, true)
				ON CONFLICT ("employeeId", "skillId") DO UPDATE SET "canPerform" = true
			`, empID, skillID); err != nil {
				return err
			}
		}
	}
	return nil
}
