package database

// migrations تعديلات تراكمية على البنية الأساسية — تُطبَّق تلقائياً كل مرة يشتغل
// فيها السيرفر حتى ما يحتاج أي شخص يشغّل ملفات SQL يدوياً.
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
	// دور "إداري الكميات" (PROCUREMENT_ADMIN) — يستلم طلبات المواد الناقصة من الموظفين
	// ويوفرها ويسجل كلفتها (نظام المشتريات الموجود أصلاً)، بدون صلاحيات إضافية افتراضية —
	// تُمنح له لاحقاً من صفحة الصلاحيات حسب الحاجة.
	`ALTER TYPE "EmployeeRole" ADD VALUE IF NOT EXISTS 'PROCUREMENT_ADMIN'`,
	// دور "مصمم" (DESIGNER) — بدون صلاحيات إضافية افتراضية، تُمنح له لاحقاً من صفحة
	// الصلاحيات حسب الحاجة، نفس نمط "إداري الكميات".
	`ALTER TYPE "EmployeeRole" ADD VALUE IF NOT EXISTS 'DESIGNER'`,
	// دور "مسؤول خدمة" (SERVICE_MANAGER) — المسؤول عن متابعة خدمة معينة (جي بي اس،
	// كاميرات...) وتوجيه كوادرها، بدون صلاحيات إضافية افتراضية — تُمنح له لاحقاً من
	// صفحة الصلاحيات حسب الخدمة الي يديرها.
	`ALTER TYPE "EmployeeRole" ADD VALUE IF NOT EXISTS 'SERVICE_MANAGER'`,
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
	// اسم احتياطي (نص) للموظف المتسبب لما يكون سجل تاريخي مستورد من نظام قديم
	// وما عنده حساب فعلي بالنظام الجديد بعد — نفس فكرة "confirmedByName" بجدول الحجوزات.
	`ALTER TABLE "Complaint" ADD COLUMN IF NOT EXISTS "relatedEmployeeName" TEXT`,

	// فهرس على تاريخ إنشاء الحجز — الصفحة الرئيسية للحجوزات تفرز حسب هذا العمود،
	// وبدون فهرس الفرز يصير بطيء جداً مع زيادة عدد الحجوزات (خصوصاً بعد استيراد البيانات التاريخية).
	`CREATE INDEX IF NOT EXISTS "Booking_createdAt_idx" ON "Booking"("createdAt" DESC)`,
	`CREATE INDEX IF NOT EXISTS "Booking_customerId_idx" ON "Booking"("customerId")`,
	`CREATE INDEX IF NOT EXISTS "Booking_status_idx" ON "Booking"(status)`,

	// وسم الخدمة على الزبون — نفس الزبون (بنفس الكود الموحّد CUST-xxxxx) ممكن يكون
	// عنده أكثر من وسم (جي بي اس، كاميرات، طاقة شمسية...) حسب الخدمات الي طلبها،
	// يستخدم لعرض "زبائن الجي بي اس" وحدهم من ضمن قائمة الزبائن الكلية بدون أي كود منفصل.
	`CREATE TABLE IF NOT EXISTS "CustomerServiceTag" (
		id TEXT PRIMARY KEY,
		"customerId" TEXT NOT NULL REFERENCES "Customer"(id) ON DELETE CASCADE,
		service TEXT NOT NULL,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE ("customerId", service)
	)`,
	`CREATE INDEX IF NOT EXISTS "CustomerServiceTag_customerId_idx" ON "CustomerServiceTag"("customerId")`,
	`CREATE INDEX IF NOT EXISTS "CustomerServiceTag_service_idx" ON "CustomerServiceTag"(service)`,

	// معلومات إضافية خاصة بزبائن الجي بي اس فقط (مستوردة من نظام التتبع القديم) —
	// جدول منفصل حتى ما نثقل جدول Customer العام بحقول ما تخص كل الزبائن.
	`CREATE TABLE IF NOT EXISTS "CustomerGpsInfo" (
		id TEXT PRIMARY KEY,
		"customerId" TEXT NOT NULL UNIQUE REFERENCES "Customer"(id) ON DELETE CASCADE,
		"gpsNumber" TEXT,
		"deviceId" TEXT,
		"subscriptionEnd" TIMESTAMP,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,

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

	// تأشير "تم" على نقص جرد الفني — يسويها الإداري/الأدمن بعد ما يوفر البديل، حتى
	// المراقب يشوف مين وفّر الاحتياج ومتى (بدون صلاحية تعديل، عرض بس).
	`ALTER TABLE "InventoryCheck" ADD COLUMN IF NOT EXISTS "resolved" BOOLEAN NOT NULL DEFAULT false`,
	`ALTER TABLE "InventoryCheck" ADD COLUMN IF NOT EXISTS "resolvedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL`,
	`ALTER TABLE "InventoryCheck" ADD COLUMN IF NOT EXISTS "resolvedAt" TIMESTAMP`,

	// إشعارات الموظفين — تصدر الترتيب، خصم نقاط الكي بي اي، وأي تنبيهات مستقبلية.
	`CREATE TABLE IF NOT EXISTS "Notification" (
		id TEXT PRIMARY KEY,
		"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
		type TEXT NOT NULL,
		message TEXT NOT NULL,
		read BOOLEAN NOT NULL DEFAULT false,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "Notification_employeeId_idx" ON "Notification"("employeeId", "createdAt" DESC)`,

	// تقييم يومي لنظافة/حالة السيارة (11 بند 0-4) + جودة غسيل الفني (0-2) — مقتبس
	// من ملف إكسل الشركة. النتائج تذكير للمراقب بس، بدون أي ربط تلقائي بالراتب.
	`CREATE TABLE IF NOT EXISTS "VehicleDailyRating" (
		id TEXT PRIMARY KEY,
		"vehicleId" TEXT NOT NULL REFERENCES "Vehicle"(id) ON DELETE CASCADE,
		"ratedDate" DATE NOT NULL DEFAULT CURRENT_DATE,
		wash INT,
		"exteriorClean" INT,
		"exteriorCondition" INT,
		"tireCondition" INT,
		"glassClean" INT,
		"lightsCondition" INT,
		"technicalFaults" INT,
		"faultDescription" TEXT,
		"interiorClean" INT,
		"seatsCondition" INT,
		"interiorDirt" INT,
		smell INT,
		notes TEXT,
		"recordedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "VehicleDailyRating_vehicleId_idx" ON "VehicleDailyRating"("vehicleId", "ratedDate" DESC)`,
	`CREATE TABLE IF NOT EXISTS "VehicleWashRating" (
		id TEXT PRIMARY KEY,
		"dailyRatingId" TEXT NOT NULL REFERENCES "VehicleDailyRating"(id) ON DELETE CASCADE,
		"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
		score INT NOT NULL,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "VehicleWashRating_dailyRatingId_idx" ON "VehicleWashRating"("dailyRatingId")`,
	`CREATE INDEX IF NOT EXISTS "VehicleWashRating_employeeId_idx" ON "VehicleWashRating"("employeeId")`,

	// تقسيم طلبات المشتريات لنوعين مستقلين بالصلاحية: احتياجات شخصية للموظف
	// نفسه (PERSONAL_SUPPLY) مقابل طلب منتج للزبون (CUSTOMER_PRODUCT، وهو النوع
	// الأصلي القديم لذلك القيمة الافتراضية تحافظ على الصفوف الموجودة).
	`ALTER TABLE "ProcurementRequest" ADD COLUMN IF NOT EXISTS "requestType" TEXT NOT NULL DEFAULT 'CUSTOMER_PRODUCT'`,

	// ملف السيارة الكامل: موديل، سنة صنع، أرقام شاصي/محرك، نوع وقود، عداد
	// كيلومترات حالي (يتحدث تلقائياً من نظام المهمة، مع إمكانية تعديل يدوي
	// كخيار احتياطي)، وحالة السيارة الحالية.
	`ALTER TABLE "Vehicle" ADD COLUMN IF NOT EXISTS model TEXT`,
	`ALTER TABLE "Vehicle" ADD COLUMN IF NOT EXISTS year INTEGER`,
	`ALTER TABLE "Vehicle" ADD COLUMN IF NOT EXISTS "chassisNumber" TEXT`,
	`ALTER TABLE "Vehicle" ADD COLUMN IF NOT EXISTS "engineNumber" TEXT`,
	`ALTER TABLE "Vehicle" ADD COLUMN IF NOT EXISTS "fuelType" TEXT`,
	`ALTER TABLE "Vehicle" ADD COLUMN IF NOT EXISTS "currentOdometer" INTEGER NOT NULL DEFAULT 0`,
	`ALTER TABLE "Vehicle" ADD COLUMN IF NOT EXISTS condition TEXT`,

	// وثائق السيارة: تأمين، إجازة سنوية، فحص دوري... إلخ.
	`CREATE TABLE IF NOT EXISTS "VehicleDocument" (
		id TEXT PRIMARY KEY,
		"vehicleId" TEXT NOT NULL REFERENCES "Vehicle"(id) ON DELETE CASCADE,
		"documentType" TEXT NOT NULL,
		"documentNumber" TEXT,
		"issueDate" TIMESTAMP,
		"expiryDate" TIMESTAMP,
		"fileUrl" TEXT,
		notes TEXT,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "VehicleDocument_vehicleId_idx" ON "VehicleDocument"("vehicleId")`,

	// صور السيارة (معرض صور عام للسيارة).
	`CREATE TABLE IF NOT EXISTS "VehiclePhoto" (
		id TEXT PRIMARY KEY,
		"vehicleId" TEXT NOT NULL REFERENCES "Vehicle"(id) ON DELETE CASCADE,
		url TEXT NOT NULL,
		caption TEXT,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "VehiclePhoto_vehicleId_idx" ON "VehiclePhoto"("vehicleId")`,

	// نظام المهمة: كل خروج سيارة يصير سجل مهمة (سبب، وجهة، عداد بداية/نهاية،
	// مسافة محسوبة، ركاب مرافقين) بدل سجل خروج/دخول مجرد.
	`CREATE TABLE IF NOT EXISTS "VehicleMission" (
		id TEXT PRIMARY KEY,
		"vehicleId" TEXT NOT NULL REFERENCES "Vehicle"(id),
		"driverId" TEXT NOT NULL REFERENCES "Employee"(id),
		purpose TEXT NOT NULL,
		destination TEXT NOT NULL,
		"startedAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"endedAt" TIMESTAMP,
		"startOdometer" INTEGER NOT NULL,
		"endOdometer" INTEGER,
		"distanceKm" INTEGER,
		notes TEXT,
		status TEXT NOT NULL DEFAULT 'IN_PROGRESS',
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "VehicleMission_vehicleId_idx" ON "VehicleMission"("vehicleId")`,
	`CREATE INDEX IF NOT EXISTS "VehicleMission_driverId_idx" ON "VehicleMission"("driverId")`,
	`CREATE INDEX IF NOT EXISTS "VehicleMission_status_idx" ON "VehicleMission"("status")`,

	// الموظفين المرافقين بالمهمة (يُعلَنون عند بدء المهمة).
	`CREATE TABLE IF NOT EXISTS "VehicleMissionPassenger" (
		id TEXT PRIMARY KEY,
		"missionId" TEXT NOT NULL REFERENCES "VehicleMission"(id) ON DELETE CASCADE,
		"employeeId" TEXT NOT NULL REFERENCES "Employee"(id),
		UNIQUE("missionId", "employeeId")
	)`,

	// المرحلة 2: صيانة عامة (نوع إضافي لسجل السيارة) + استحقاق الصيانة حسب العداد
	// بالإضافة لاستحقاقها حسب التاريخ (الموجود مسبقاً).
	`ALTER TABLE "VehicleLog" ADD COLUMN IF NOT EXISTS "nextDueOdometer" INTEGER`,

	// مرفقات الأعطال/الأضرار (صور/فيديو) — نفس أسلوب تخزين base64 المستخدم بالوثائق والصور.
	`CREATE TABLE IF NOT EXISTS "VehicleIncidentAttachment" (
		id TEXT PRIMARY KEY,
		"incidentId" TEXT NOT NULL REFERENCES "VehicleIncident"(id) ON DELETE CASCADE,
		url TEXT NOT NULL,
		"mediaType" TEXT NOT NULL DEFAULT 'IMAGE',
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "VehicleIncidentAttachment_incidentId_idx" ON "VehicleIncidentAttachment"("incidentId")`,

	// متابعة قطع الاستهلاك (إطارات وبطاريات) — استحقاق التبديل حسب المسافة أو الزمن أيهما أقرب.
	`CREATE TABLE IF NOT EXISTS "VehiclePart" (
		id TEXT PRIMARY KEY,
		"vehicleId" TEXT NOT NULL REFERENCES "Vehicle"(id) ON DELETE CASCADE,
		"partType" TEXT NOT NULL,
		"installedAt" TIMESTAMP NOT NULL,
		"installedOdometer" INTEGER NOT NULL,
		"expectedLifespanKm" INTEGER,
		"expectedLifespanMonths" INTEGER,
		notes TEXT,
		"replacedAt" TIMESTAMP,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "VehiclePart_vehicleId_idx" ON "VehiclePart"("vehicleId")`,

	// المرحلة 3: توثيق كامل للحوادث (نوع ACCIDENT) + حقول تفصيلية اختيارية تظهر فقط لهذا النوع.
	`ALTER TABLE "VehicleIncident" ADD COLUMN IF NOT EXISTS "location" TEXT`,
	`ALTER TABLE "VehicleIncident" ADD COLUMN IF NOT EXISTS "driverId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL`,
	`ALTER TABLE "VehicleIncident" ADD COLUMN IF NOT EXISTS "peoplePresent" TEXT`,
	`ALTER TABLE "VehicleIncident" ADD COLUMN IF NOT EXISTS "policeReportNumber" TEXT`,
	`ALTER TABLE "VehicleIncident" ADD COLUMN IF NOT EXISTS "repairCost" DOUBLE PRECISION`,

	// المرحلة 3: تكلفة استبدال القطعة (إطار/بطارية) — تدخل بحساب مصاريف السيارة.
	`ALTER TABLE "VehiclePart" ADD COLUMN IF NOT EXISTS cost DOUBLE PRECISION`,

	// المرحلة 4-أ: نظام حجز المركبات (مسبق) — منفصل عن بدء المهمة الفعلي (VehicleMission).
	`CREATE TABLE IF NOT EXISTS "VehicleBooking" (
		id TEXT PRIMARY KEY,
		"vehicleId" TEXT NOT NULL REFERENCES "Vehicle"(id),
		"requestedById" TEXT NOT NULL REFERENCES "Employee"(id),
		purpose TEXT NOT NULL,
		"startAt" TIMESTAMP NOT NULL,
		"endAt" TIMESTAMP NOT NULL,
		status TEXT NOT NULL DEFAULT 'PENDING',
		"approvedById" TEXT REFERENCES "Employee"(id),
		"rejectionReason" TEXT,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"decidedAt" TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "VehicleBooking_vehicleId_idx" ON "VehicleBooking"("vehicleId")`,
	`CREATE INDEX IF NOT EXISTS "VehicleBooking_requestedById_idx" ON "VehicleBooking"("requestedById")`,
	`CREATE INDEX IF NOT EXISTS "VehicleBooking_status_idx" ON "VehicleBooking"("status")`,

	// المرحلة 4-ب: تقييم السائق بعد كل مهمة مكتملة.
	`CREATE TABLE IF NOT EXISTS "VehicleMissionRating" (
		id TEXT PRIMARY KEY,
		"missionId" TEXT NOT NULL UNIQUE REFERENCES "VehicleMission"(id) ON DELETE CASCADE,
		"ratedById" TEXT NOT NULL REFERENCES "Employee"(id),
		commitment INTEGER NOT NULL,
		"vehicleCare" INTEGER NOT NULL,
		driving INTEGER NOT NULL,
		cleanliness INTEGER NOT NULL,
		notes TEXT,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "VehicleMissionRating_missionId_idx" ON "VehicleMissionRating"("missionId")`,
	`CREATE TABLE IF NOT EXISTS "AssistantConversation" (
		id TEXT PRIMARY KEY,
		"employeeId" TEXT NOT NULL REFERENCES "Employee"(id),
		message TEXT NOT NULL,
		reply TEXT NOT NULL,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "AssistantConversation_employeeId_createdAt_idx" ON "AssistantConversation"("employeeId", "createdAt")`,

	// معرفة تعلمها المساعد الذكي من محادثاته مع الموظفين (بحث ويب، معلومات
	// علّمه إياها موظف، حقائق عامة مفيدة) — يخزنها ويرجّعها بمحادثات لاحقة
	// حتى يصير "أذكى" مع الاستخدام بدل ما ينسى كل شي بعد كل رد.
	`CREATE TABLE IF NOT EXISTS "AssistantKnowledge" (
		id TEXT PRIMARY KEY,
		topic TEXT NOT NULL,
		content TEXT NOT NULL,
		"learnedFromEmployeeId" TEXT REFERENCES "Employee"(id),
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "AssistantKnowledge_topic_idx" ON "AssistantKnowledge"(topic)`,
	`CREATE INDEX IF NOT EXISTS "AssistantKnowledge_createdAt_idx" ON "AssistantKnowledge"("createdAt")`,
}
