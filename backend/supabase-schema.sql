-- CreateEnum
CREATE TYPE "ProjectPriority" AS ENUM ('NORMAL', 'URGENT');

-- CreateEnum
CREATE TYPE "BookingStatus" AS ENUM ('IN_PROGRESS', 'PENDING', 'CONFIRMED', 'COMPLETED', 'CANCELLED');

-- CreateEnum
CREATE TYPE "EquipmentStatus" AS ENUM ('NOT_READY', 'READY');

-- CreateEnum
CREATE TYPE "Shift" AS ENUM ('MORNING', 'EVENING');

-- CreateEnum
CREATE TYPE "TechnicianRole" AS ENUM ('TECH_1', 'TECH_2', 'TECH_3');

-- CreateEnum
CREATE TYPE "EmployeeStatus" AS ENUM ('ACTIVE', 'INACTIVE');

-- CreateEnum
CREATE TYPE "EmployeeRole" AS ENUM ('ADMIN', 'SALES', 'HR_COORDINATOR', 'TECHNICIAN', 'PROJECT_MANAGER', 'MONITOR', 'FINANCE', 'GPS_ADMIN', 'GPS_ENGINEER', 'QUALITY_ENGINEER');

-- CreateEnum
CREATE TYPE "ExpenseStatus" AS ENUM ('PENDING', 'APPROVED', 'REJECTED');

-- CreateEnum
CREATE TYPE "ToolStatus" AS ENUM ('AVAILABLE', 'CHECKED_OUT', 'DAMAGED');

-- CreateEnum
CREATE TYPE "ToolRequestStatus" AS ENUM ('PENDING', 'APPROVED', 'REJECTED', 'RETURNED');

-- CreateEnum
CREATE TYPE "ComplaintStatus" AS ENUM ('NEW', 'IN_PROGRESS', 'RESOLVED', 'CLOSED');

-- CreateEnum
CREATE TYPE "SimOperator" AS ENUM ('ZAIN', 'ASIACELL', 'KOREK', 'OTHER');

-- CreateEnum
CREATE TYPE "SimStatus" AS ENUM ('AVAILABLE', 'IN_USE');

-- CreateEnum
CREATE TYPE "GpsRequestStatus" AS ENUM ('PENDING', 'APPROVED', 'REJECTED', 'DELIVERED');

-- CreateEnum
CREATE TYPE "GpsPurchaseType" AS ENUM ('DEVICE_SIM', 'DEVICE_ONLY');

-- CreateEnum
CREATE TYPE "GpsSubscriptionType" AS ENUM ('THREE_MONTHS', 'SIX_MONTHS', 'YEARLY');

-- CreateEnum
CREATE TYPE "GpsSubscriptionStatus" AS ENUM ('ACTIVE', 'EXPIRED', 'CANCELLED');

-- CreateEnum
CREATE TYPE "GpsMaintenanceStatus" AS ENUM ('PENDING', 'IN_PROGRESS', 'COMPLETED');

-- CreateEnum
CREATE TYPE "QuotationStatus" AS ENUM ('NEW', 'SENT', 'ACCEPTED', 'REJECTED');

-- CreateEnum
CREATE TYPE "ProcurementRequestStatus" AS ENUM ('PENDING', 'IN_PROGRESS', 'FULFILLED', 'REJECTED');

-- CreateEnum
CREATE TYPE "BookingType" AS ENUM ('REGULAR', 'MAINTENANCE');

-- CreateEnum
CREATE TYPE "BookingUrgency" AS ENUM ('ASAP', 'BY_PRIORITY', 'SPECIFIC_DATE');

-- CreateEnum
CREATE TYPE "MaintenanceType" AS ENUM ('EXECUTION_ERROR', 'DEVICE_ISSUE');

-- CreateEnum
CREATE TYPE "WorkStatus" AS ENUM ('COMPLETED', 'STOPPED');

-- CreateEnum
CREATE TYPE "ProjectSpeed" AS ENUM ('URGENT', 'INTERNAL', 'REGULAR', 'DIRECTOR_REQ');

-- CreateEnum
CREATE TYPE "WorkType" AS ENUM ('FREE_MAINTENANCE', 'PAID');

-- CreateEnum
CREATE TYPE "MissionStage" AS ENUM ('ASSIGNED', 'MATERIALS_PREP', 'MATERIALS_READY', 'EN_ROUTE', 'ARRIVED', 'WORK_STARTED', 'COMPLETED', 'STOPPED');

-- CreateTable
CREATE TABLE "Service" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "category" TEXT,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "Service_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "Skill" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "serviceId" TEXT NOT NULL,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "Skill_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "Customer" (
    "id" TEXT NOT NULL,
    "customerCode" SERIAL NOT NULL,
    "name" TEXT NOT NULL,
    "phone" TEXT NOT NULL,
    "location" TEXT,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "Customer_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "Employee" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "certificate" TEXT,
    "position" TEXT,
    "phone" TEXT,
    "status" "EmployeeStatus" NOT NULL DEFAULT 'ACTIVE',
    "role" "EmployeeRole" NOT NULL DEFAULT 'TECHNICIAN',
    "onDuty" BOOLEAN NOT NULL DEFAULT true,
    "username" TEXT,
    "password" TEXT,
    "hasDrivingLicense" BOOLEAN NOT NULL DEFAULT false,
    "hasSafetyCertificate" BOOLEAN NOT NULL DEFAULT false,
    "salary" DOUBLE PRECISION,
    "shift" "Shift" DEFAULT 'MORNING',
    "monthlyLeaves" INTEGER NOT NULL DEFAULT 2,
    "jobTitle" TEXT,
    "leaderSkillLevel" INTEGER NOT NULL DEFAULT 0,
    "isLeader" BOOLEAN NOT NULL DEFAULT false,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "Employee_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "Expense" (
    "id" TEXT NOT NULL,
    "employeeId" TEXT NOT NULL,
    "amount" DOUBLE PRECISION NOT NULL,
    "description" TEXT,
    "status" "ExpenseStatus" NOT NULL DEFAULT 'PENDING',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "Expense_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "EmployeeSkill" (
    "id" TEXT NOT NULL,
    "employeeId" TEXT NOT NULL,
    "skillId" TEXT NOT NULL,
    "canPerform" BOOLEAN NOT NULL DEFAULT true,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "EmployeeSkill_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "Booking" (
    "id" TEXT NOT NULL,
    "code" TEXT NOT NULL,
    "sequenceNumber" INTEGER,
    "scheduledAt" TIMESTAMP(3),
    "customerId" TEXT NOT NULL,
    "serviceId" TEXT,
    "notes" TEXT,
    "vehicleType" TEXT,
    "priority" "ProjectPriority" NOT NULL DEFAULT 'NORMAL',
    "status" "BookingStatus" NOT NULL DEFAULT 'PENDING',
    "transferEmployeeId" TEXT,
    "transferToProjects" BOOLEAN NOT NULL DEFAULT false,
    "confirmedByName" TEXT,
    "confirmedByEmployeeId" TEXT,
    "adminNotes" TEXT,
    "equipmentStatus" "EquipmentStatus" NOT NULL DEFAULT 'NOT_READY',
    "shift" "Shift",
    "deviceCount" INTEGER,
    "inspectionSupervisorId" TEXT,
    "projectSupervisorId" TEXT,
    "projectCar" TEXT,
    "crewNotes" TEXT,
    "mapLocation" TEXT,
    "expenseResponsibleId" TEXT,
    "assignedVehicle" TEXT,
    "quotedPrice" DOUBLE PRECISION,
    "address" TEXT,
    "completedAt" TIMESTAMP(3),
    "completionNotes" TEXT,
    "amountCollected" DOUBLE PRECISION,
    "advancePaid" DOUBLE PRECISION,
    "amountVerified" BOOLEAN NOT NULL DEFAULT false,
    "bookingType" "BookingType" NOT NULL DEFAULT 'REGULAR',
    "urgency" "BookingUrgency",
    "maintenanceType" "MaintenanceType",
    "remembersExecutionCrew" BOOLEAN NOT NULL DEFAULT false,
    "systemCount" INTEGER,
    "systemType" TEXT,
    "projectSpeed" "ProjectSpeed",
    "workType" "WorkType",
    "addressDescription" TEXT,
    "mapLatitude" DOUBLE PRECISION,
    "mapLongitude" DOUBLE PRECISION,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "Booking_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "BookingAssignment" (
    "id" TEXT NOT NULL,
    "bookingId" TEXT NOT NULL,
    "employeeId" TEXT NOT NULL,
    "role" "TechnicianRole" NOT NULL,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "BookingAssignment_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "ScheduleChangeLog" (
    "id" TEXT NOT NULL,
    "bookingId" TEXT NOT NULL,
    "changedById" TEXT NOT NULL,
    "oldTime" TIMESTAMP(3),
    "newTime" TIMESTAMP(3) NOT NULL,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "ScheduleChangeLog_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "Permission" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "label" TEXT NOT NULL,

    CONSTRAINT "Permission_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "EmployeePermission" (
    "id" TEXT NOT NULL,
    "employeeId" TEXT NOT NULL,
    "permissionId" TEXT NOT NULL,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "EmployeePermission_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "KpiEvaluation" (
    "id" TEXT NOT NULL,
    "employeeId" TEXT NOT NULL,
    "evaluatorId" TEXT NOT NULL,
    "points" INTEGER NOT NULL,
    "reason" TEXT NOT NULL,
    "deductionAmount" DOUBLE PRECISION NOT NULL,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "KpiEvaluation_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "CartItem" (
    "id" TEXT NOT NULL,
    "bookingId" TEXT NOT NULL,
    "productName" TEXT NOT NULL,
    "quantity" INTEGER NOT NULL,
    "unitPrice" DOUBLE PRECISION NOT NULL,
    "totalPrice" DOUBLE PRECISION NOT NULL,
    "notes" TEXT,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "CartItem_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "PersonalTool" (
    "id" TEXT NOT NULL,
    "employeeId" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "barcode" TEXT NOT NULL,
    "status" "ToolStatus" NOT NULL DEFAULT 'AVAILABLE',
    "checkedOut" BOOLEAN NOT NULL DEFAULT false,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "PersonalTool_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "VehicleTool" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "barcode" TEXT NOT NULL,
    "vehicleId" TEXT NOT NULL,
    "status" "ToolStatus" NOT NULL DEFAULT 'AVAILABLE',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "VehicleTool_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "OnDemandTool" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "barcode" TEXT NOT NULL,
    "totalQuantity" INTEGER NOT NULL,
    "availableQuantity" INTEGER NOT NULL,
    "status" "ToolStatus" NOT NULL DEFAULT 'AVAILABLE',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "OnDemandTool_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "ToolRequest" (
    "id" TEXT NOT NULL,
    "employeeId" TEXT NOT NULL,
    "toolId" TEXT NOT NULL,
    "status" "ToolRequestStatus" NOT NULL DEFAULT 'PENDING',
    "approvedById" TEXT,
    "requestedAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "returnedAt" TIMESTAMP(3),

    CONSTRAINT "ToolRequest_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "Complaint" (
    "id" TEXT NOT NULL,
    "customerId" TEXT NOT NULL,
    "bookingId" TEXT,
    "description" TEXT NOT NULL,
    "status" "ComplaintStatus" NOT NULL DEFAULT 'NEW',
    "createdByEmployeeId" TEXT NOT NULL,
    "assignedToEmployeeId" TEXT,
    "resolution" TEXT,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "resolvedAt" TIMESTAMP(3),

    CONSTRAINT "Complaint_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "GpsCustomer" (
    "id" TEXT NOT NULL,
    "fullName" TEXT NOT NULL,
    "fatherName" TEXT,
    "grandfatherName" TEXT,
    "phone" TEXT NOT NULL,
    "address" TEXT,
    "governorate" TEXT,
    "idCardFrontUrl" TEXT,
    "idCardBackUrl" TEXT,
    "residenceCardFrontUrl" TEXT,
    "residenceCardBackUrl" TEXT,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "GpsCustomer_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "SimCard" (
    "id" TEXT NOT NULL,
    "simNumber" TEXT NOT NULL,
    "iccid" TEXT,
    "operator" "SimOperator" NOT NULL,
    "status" "SimStatus" NOT NULL DEFAULT 'AVAILABLE',
    "customerId" TEXT,
    "notes" TEXT,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "SimCard_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "GpsDeviceRequest" (
    "id" TEXT NOT NULL,
    "customerId" TEXT NOT NULL,
    "employeeId" TEXT NOT NULL,
    "adminId" TEXT,
    "purchaseType" "GpsPurchaseType" NOT NULL,
    "subscriptionType" "GpsSubscriptionType" NOT NULL,
    "subscriptionStart" TIMESTAMP(3),
    "subscriptionEnd" TIMESTAMP(3),
    "subscriptionStatus" "GpsSubscriptionStatus" NOT NULL DEFAULT 'ACTIVE',
    "status" "GpsRequestStatus" NOT NULL DEFAULT 'PENDING',
    "simCardId" TEXT,
    "notes" TEXT,
    "isChecked" BOOLEAN NOT NULL DEFAULT false,
    "isActivated" BOOLEAN NOT NULL DEFAULT false,
    "isDelivered" BOOLEAN NOT NULL DEFAULT false,
    "invoicePhotoUrl" TEXT,
    "activationDate" TIMESTAMP(3),
    "deliveredAt" TIMESTAMP(3),
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "GpsDeviceRequest_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "GpsRenewalRequest" (
    "id" TEXT NOT NULL,
    "customerId" TEXT NOT NULL,
    "deviceRequestId" TEXT NOT NULL,
    "employeeId" TEXT NOT NULL,
    "adminId" TEXT,
    "subscriptionType" "GpsSubscriptionType" NOT NULL,
    "newEndDate" TIMESTAMP(3),
    "status" "GpsRequestStatus" NOT NULL DEFAULT 'PENDING',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "GpsRenewalRequest_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "GpsMaintenanceRequest" (
    "id" TEXT NOT NULL,
    "customerId" TEXT NOT NULL,
    "employeeId" TEXT NOT NULL,
    "adminId" TEXT,
    "problemDescription" TEXT NOT NULL,
    "status" "GpsMaintenanceStatus" NOT NULL DEFAULT 'PENDING',
    "adminNotes" TEXT,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "resolvedAt" TIMESTAMP(3),

    CONSTRAINT "GpsMaintenanceRequest_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "GpsSubscriptionPrice" (
    "id" TEXT NOT NULL,
    "subscriptionType" "GpsSubscriptionType" NOT NULL,
    "price" DOUBLE PRECISION NOT NULL,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "GpsSubscriptionPrice_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "Product" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "unit" TEXT,
    "defaultPrice" DOUBLE PRECISION,
    "imageBase64" TEXT,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "Product_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "Quotation" (
    "id" TEXT NOT NULL,
    "quotationNumber" TEXT NOT NULL,
    "customerName" TEXT NOT NULL,
    "customerPhone" TEXT,
    "customerAddress" TEXT,
    "projectName" TEXT,
    "grandTotal" DOUBLE PRECISION NOT NULL DEFAULT 0,
    "discountPercent" DOUBLE PRECISION NOT NULL DEFAULT 0,
    "discountValue" DOUBLE PRECISION NOT NULL DEFAULT 0,
    "netTotal" DOUBLE PRECISION NOT NULL DEFAULT 0,
    "notes" TEXT,
    "duration" TEXT,
    "status" "QuotationStatus" NOT NULL DEFAULT 'NEW',
    "createdByEmployeeId" TEXT NOT NULL,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "Quotation_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "QuotationItem" (
    "id" TEXT NOT NULL,
    "quotationId" TEXT NOT NULL,
    "productName" TEXT NOT NULL,
    "unit" TEXT,
    "quantity" INTEGER NOT NULL,
    "unitPrice" DOUBLE PRECISION NOT NULL,
    "totalPrice" DOUBLE PRECISION NOT NULL,

    CONSTRAINT "QuotationItem_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "Attendance" (
    "id" TEXT NOT NULL,
    "employeeId" TEXT NOT NULL,
    "checkIn" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "checkOut" TIMESTAMP(3),
    "date" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "Attendance_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "WorkReport" (
    "id" TEXT NOT NULL,
    "bookingId" TEXT NOT NULL,
    "employeeId" TEXT NOT NULL,
    "workStatus" "WorkStatus" NOT NULL,
    "events" TEXT,
    "extraRequests" TEXT,
    "cleanedSite" BOOLEAN NOT NULL DEFAULT false,
    "gaveInfo" BOOLEAN NOT NULL DEFAULT false,
    "tookPhotos" BOOLEAN NOT NULL DEFAULT false,
    "stopReason" TEXT,
    "notes" TEXT,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "WorkReport_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "Vehicle" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "plateNumber" TEXT NOT NULL,
    "color" TEXT,
    "type" TEXT,
    "isActive" BOOLEAN NOT NULL DEFAULT true,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "Vehicle_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "SupplierSpecialty" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "order" INTEGER NOT NULL DEFAULT 0,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "SupplierSpecialty_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "Supplier" (
    "id" TEXT NOT NULL,
    "companyName" TEXT NOT NULL,
    "ownerName" TEXT NOT NULL,
    "phone" TEXT NOT NULL,
    "lat" DOUBLE PRECISION,
    "lng" DOUBLE PRECISION,
    "isMaterialSupplier" BOOLEAN NOT NULL DEFAULT false,
    "isContractor" BOOLEAN NOT NULL DEFAULT false,
    "traderTypes" TEXT[],
    "notes" TEXT,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "Supplier_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "SupplierSpecialtyLink" (
    "supplierId" TEXT NOT NULL,
    "specialtyId" TEXT NOT NULL,

    CONSTRAINT "SupplierSpecialtyLink_pkey" PRIMARY KEY ("supplierId","specialtyId")
);

-- CreateTable
CREATE TABLE "SupplierRating" (
    "id" TEXT NOT NULL,
    "supplierId" TEXT NOT NULL,
    "value" INTEGER NOT NULL,
    "note" TEXT,
    "ratedById" TEXT NOT NULL,
    "ratedByName" TEXT NOT NULL,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "SupplierRating_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "Mission" (
    "id" TEXT NOT NULL,
    "code" TEXT NOT NULL,
    "bookingId" TEXT NOT NULL,
    "stage" "MissionStage" NOT NULL DEFAULT 'ASSIGNED',
    "leaderId" TEXT NOT NULL,
    "memberIds" TEXT[],
    "customerLat" DOUBLE PRECISION,
    "customerLng" DOUBLE PRECISION,
    "customerAddress" TEXT,
    "assignedAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "materialsReadyAt" TIMESTAMP(3),
    "departedAt" TIMESTAMP(3),
    "arrivedAt" TIMESTAMP(3),
    "workStartedAt" TIMESTAMP(3),
    "completedAt" TIMESTAMP(3),
    "stoppedAt" TIMESTAMP(3),
    "departureLat" DOUBLE PRECISION,
    "departureLng" DOUBLE PRECISION,
    "arrivalLat" DOUBLE PRECISION,
    "arrivalLng" DOUBLE PRECISION,
    "estimatedMinutes" INTEGER,
    "actualMinutes" INTEGER,
    "distanceKm" DOUBLE PRECISION,
    "stopReason" TEXT,
    "notes" TEXT,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "Mission_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "MissionEvent" (
    "id" TEXT NOT NULL,
    "missionId" TEXT NOT NULL,
    "employeeId" TEXT NOT NULL,
    "action" TEXT NOT NULL,
    "lat" DOUBLE PRECISION,
    "lng" DOUBLE PRECISION,
    "note" TEXT,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "MissionEvent_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "Project" (
    "id" TEXT NOT NULL,
    "code" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "rep" TEXT,
    "phone" TEXT,
    "location" TEXT,
    "workType" TEXT,
    "refPerson" TEXT,
    "stage" TEXT NOT NULL DEFAULT '1. اتصال بالزبون',
    "price" TEXT,
    "staff" TEXT,
    "time" TEXT,
    "task" TEXT,
    "priority" TEXT NOT NULL DEFAULT 'عادي',
    "deliveryDate" TEXT,
    "survey" JSONB,
    "sentToGroup" BOOLEAN NOT NULL DEFAULT false,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "Project_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "ProcurementRequest" (
    "id" TEXT NOT NULL,
    "code" TEXT NOT NULL,
    "requestedById" TEXT NOT NULL,
    "bookingId" TEXT,
    "notes" TEXT,
    "status" "ProcurementRequestStatus" NOT NULL DEFAULT 'PENDING',
    "fulfilledById" TEXT,
    "totalCost" DOUBLE PRECISION,
    "fulfillmentNotes" TEXT,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fulfilledAt" TIMESTAMP(3),

    CONSTRAINT "ProcurementRequest_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "ProcurementItem" (
    "id" TEXT NOT NULL,
    "requestId" TEXT NOT NULL,
    "productName" TEXT NOT NULL,
    "quantity" INTEGER NOT NULL,
    "unitPrice" DOUBLE PRECISION,
    "totalPrice" DOUBLE PRECISION,
    "fulfilled" BOOLEAN NOT NULL DEFAULT false,

    CONSTRAINT "ProcurementItem_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE UNIQUE INDEX "Service_name_key" ON "Service"("name");

-- CreateIndex
CREATE INDEX "Skill_serviceId_idx" ON "Skill"("serviceId");

-- CreateIndex
CREATE UNIQUE INDEX "Skill_serviceId_name_key" ON "Skill"("serviceId", "name");

-- CreateIndex
CREATE UNIQUE INDEX "Customer_customerCode_key" ON "Customer"("customerCode");

-- CreateIndex
CREATE UNIQUE INDEX "Customer_phone_key" ON "Customer"("phone");

-- CreateIndex
CREATE UNIQUE INDEX "Employee_username_key" ON "Employee"("username");

-- CreateIndex
CREATE INDEX "Expense_employeeId_idx" ON "Expense"("employeeId");

-- CreateIndex
CREATE INDEX "EmployeeSkill_skillId_idx" ON "EmployeeSkill"("skillId");

-- CreateIndex
CREATE UNIQUE INDEX "EmployeeSkill_employeeId_skillId_key" ON "EmployeeSkill"("employeeId", "skillId");

-- CreateIndex
CREATE UNIQUE INDEX "Booking_code_key" ON "Booking"("code");

-- CreateIndex
CREATE INDEX "Booking_customerId_idx" ON "Booking"("customerId");

-- CreateIndex
CREATE INDEX "Booking_serviceId_idx" ON "Booking"("serviceId");

-- CreateIndex
CREATE INDEX "Booking_status_idx" ON "Booking"("status");

-- CreateIndex
CREATE INDEX "BookingAssignment_employeeId_idx" ON "BookingAssignment"("employeeId");

-- CreateIndex
CREATE UNIQUE INDEX "BookingAssignment_bookingId_role_key" ON "BookingAssignment"("bookingId", "role");

-- CreateIndex
CREATE INDEX "ScheduleChangeLog_bookingId_idx" ON "ScheduleChangeLog"("bookingId");

-- CreateIndex
CREATE INDEX "ScheduleChangeLog_changedById_idx" ON "ScheduleChangeLog"("changedById");

-- CreateIndex
CREATE UNIQUE INDEX "Permission_name_key" ON "Permission"("name");

-- CreateIndex
CREATE UNIQUE INDEX "EmployeePermission_employeeId_permissionId_key" ON "EmployeePermission"("employeeId", "permissionId");

-- CreateIndex
CREATE INDEX "KpiEvaluation_employeeId_idx" ON "KpiEvaluation"("employeeId");

-- CreateIndex
CREATE INDEX "CartItem_bookingId_idx" ON "CartItem"("bookingId");

-- CreateIndex
CREATE UNIQUE INDEX "PersonalTool_barcode_key" ON "PersonalTool"("barcode");

-- CreateIndex
CREATE INDEX "PersonalTool_employeeId_idx" ON "PersonalTool"("employeeId");

-- CreateIndex
CREATE UNIQUE INDEX "VehicleTool_barcode_key" ON "VehicleTool"("barcode");

-- CreateIndex
CREATE INDEX "VehicleTool_vehicleId_idx" ON "VehicleTool"("vehicleId");

-- CreateIndex
CREATE UNIQUE INDEX "OnDemandTool_barcode_key" ON "OnDemandTool"("barcode");

-- CreateIndex
CREATE INDEX "ToolRequest_employeeId_idx" ON "ToolRequest"("employeeId");

-- CreateIndex
CREATE INDEX "ToolRequest_toolId_idx" ON "ToolRequest"("toolId");

-- CreateIndex
CREATE INDEX "Complaint_customerId_idx" ON "Complaint"("customerId");

-- CreateIndex
CREATE INDEX "Complaint_status_idx" ON "Complaint"("status");

-- CreateIndex
CREATE UNIQUE INDEX "SimCard_simNumber_key" ON "SimCard"("simNumber");

-- CreateIndex
CREATE INDEX "SimCard_status_idx" ON "SimCard"("status");

-- CreateIndex
CREATE INDEX "GpsDeviceRequest_customerId_idx" ON "GpsDeviceRequest"("customerId");

-- CreateIndex
CREATE INDEX "GpsDeviceRequest_status_idx" ON "GpsDeviceRequest"("status");

-- CreateIndex
CREATE INDEX "GpsRenewalRequest_customerId_idx" ON "GpsRenewalRequest"("customerId");

-- CreateIndex
CREATE INDEX "GpsMaintenanceRequest_customerId_idx" ON "GpsMaintenanceRequest"("customerId");

-- CreateIndex
CREATE UNIQUE INDEX "GpsSubscriptionPrice_subscriptionType_key" ON "GpsSubscriptionPrice"("subscriptionType");

-- CreateIndex
CREATE UNIQUE INDEX "Product_name_key" ON "Product"("name");

-- CreateIndex
CREATE UNIQUE INDEX "Quotation_quotationNumber_key" ON "Quotation"("quotationNumber");

-- CreateIndex
CREATE INDEX "Quotation_status_idx" ON "Quotation"("status");

-- CreateIndex
CREATE INDEX "QuotationItem_quotationId_idx" ON "QuotationItem"("quotationId");

-- CreateIndex
CREATE INDEX "Attendance_employeeId_idx" ON "Attendance"("employeeId");

-- CreateIndex
CREATE INDEX "Attendance_date_idx" ON "Attendance"("date");

-- CreateIndex
CREATE UNIQUE INDEX "Attendance_employeeId_date_key" ON "Attendance"("employeeId", "date");

-- CreateIndex
CREATE INDEX "WorkReport_bookingId_idx" ON "WorkReport"("bookingId");

-- CreateIndex
CREATE INDEX "WorkReport_employeeId_idx" ON "WorkReport"("employeeId");

-- CreateIndex
CREATE UNIQUE INDEX "Vehicle_plateNumber_key" ON "Vehicle"("plateNumber");

-- CreateIndex
CREATE UNIQUE INDEX "SupplierSpecialty_name_key" ON "SupplierSpecialty"("name");

-- CreateIndex
CREATE INDEX "SupplierRating_supplierId_idx" ON "SupplierRating"("supplierId");

-- CreateIndex
CREATE UNIQUE INDEX "Mission_code_key" ON "Mission"("code");

-- CreateIndex
CREATE INDEX "Mission_bookingId_idx" ON "Mission"("bookingId");

-- CreateIndex
CREATE INDEX "Mission_stage_idx" ON "Mission"("stage");

-- CreateIndex
CREATE INDEX "Mission_leaderId_idx" ON "Mission"("leaderId");

-- CreateIndex
CREATE INDEX "MissionEvent_missionId_idx" ON "MissionEvent"("missionId");

-- CreateIndex
CREATE UNIQUE INDEX "Project_code_key" ON "Project"("code");

-- CreateIndex
CREATE INDEX "Project_stage_idx" ON "Project"("stage");

-- CreateIndex
CREATE UNIQUE INDEX "ProcurementRequest_code_key" ON "ProcurementRequest"("code");

-- CreateIndex
CREATE INDEX "ProcurementRequest_requestedById_idx" ON "ProcurementRequest"("requestedById");

-- CreateIndex
CREATE INDEX "ProcurementRequest_status_idx" ON "ProcurementRequest"("status");

-- CreateIndex
CREATE INDEX "ProcurementItem_requestId_idx" ON "ProcurementItem"("requestId");

-- AddForeignKey
ALTER TABLE "Skill" ADD CONSTRAINT "Skill_serviceId_fkey" FOREIGN KEY ("serviceId") REFERENCES "Service"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "Expense" ADD CONSTRAINT "Expense_employeeId_fkey" FOREIGN KEY ("employeeId") REFERENCES "Employee"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "EmployeeSkill" ADD CONSTRAINT "EmployeeSkill_employeeId_fkey" FOREIGN KEY ("employeeId") REFERENCES "Employee"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "EmployeeSkill" ADD CONSTRAINT "EmployeeSkill_skillId_fkey" FOREIGN KEY ("skillId") REFERENCES "Skill"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "Booking" ADD CONSTRAINT "Booking_customerId_fkey" FOREIGN KEY ("customerId") REFERENCES "Customer"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "Booking" ADD CONSTRAINT "Booking_serviceId_fkey" FOREIGN KEY ("serviceId") REFERENCES "Service"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "Booking" ADD CONSTRAINT "Booking_transferEmployeeId_fkey" FOREIGN KEY ("transferEmployeeId") REFERENCES "Employee"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "Booking" ADD CONSTRAINT "Booking_confirmedByEmployeeId_fkey" FOREIGN KEY ("confirmedByEmployeeId") REFERENCES "Employee"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "Booking" ADD CONSTRAINT "Booking_inspectionSupervisorId_fkey" FOREIGN KEY ("inspectionSupervisorId") REFERENCES "Employee"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "Booking" ADD CONSTRAINT "Booking_projectSupervisorId_fkey" FOREIGN KEY ("projectSupervisorId") REFERENCES "Employee"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "Booking" ADD CONSTRAINT "Booking_expenseResponsibleId_fkey" FOREIGN KEY ("expenseResponsibleId") REFERENCES "Employee"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "BookingAssignment" ADD CONSTRAINT "BookingAssignment_bookingId_fkey" FOREIGN KEY ("bookingId") REFERENCES "Booking"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "BookingAssignment" ADD CONSTRAINT "BookingAssignment_employeeId_fkey" FOREIGN KEY ("employeeId") REFERENCES "Employee"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "ScheduleChangeLog" ADD CONSTRAINT "ScheduleChangeLog_bookingId_fkey" FOREIGN KEY ("bookingId") REFERENCES "Booking"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "ScheduleChangeLog" ADD CONSTRAINT "ScheduleChangeLog_changedById_fkey" FOREIGN KEY ("changedById") REFERENCES "Employee"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "EmployeePermission" ADD CONSTRAINT "EmployeePermission_employeeId_fkey" FOREIGN KEY ("employeeId") REFERENCES "Employee"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "EmployeePermission" ADD CONSTRAINT "EmployeePermission_permissionId_fkey" FOREIGN KEY ("permissionId") REFERENCES "Permission"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "KpiEvaluation" ADD CONSTRAINT "KpiEvaluation_employeeId_fkey" FOREIGN KEY ("employeeId") REFERENCES "Employee"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "KpiEvaluation" ADD CONSTRAINT "KpiEvaluation_evaluatorId_fkey" FOREIGN KEY ("evaluatorId") REFERENCES "Employee"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "CartItem" ADD CONSTRAINT "CartItem_bookingId_fkey" FOREIGN KEY ("bookingId") REFERENCES "Booking"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "PersonalTool" ADD CONSTRAINT "PersonalTool_employeeId_fkey" FOREIGN KEY ("employeeId") REFERENCES "Employee"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "ToolRequest" ADD CONSTRAINT "ToolRequest_employeeId_fkey" FOREIGN KEY ("employeeId") REFERENCES "Employee"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "ToolRequest" ADD CONSTRAINT "ToolRequest_toolId_fkey" FOREIGN KEY ("toolId") REFERENCES "OnDemandTool"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "ToolRequest" ADD CONSTRAINT "ToolRequest_approvedById_fkey" FOREIGN KEY ("approvedById") REFERENCES "Employee"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "Complaint" ADD CONSTRAINT "Complaint_customerId_fkey" FOREIGN KEY ("customerId") REFERENCES "Customer"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "Complaint" ADD CONSTRAINT "Complaint_bookingId_fkey" FOREIGN KEY ("bookingId") REFERENCES "Booking"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "Complaint" ADD CONSTRAINT "Complaint_createdByEmployeeId_fkey" FOREIGN KEY ("createdByEmployeeId") REFERENCES "Employee"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "Complaint" ADD CONSTRAINT "Complaint_assignedToEmployeeId_fkey" FOREIGN KEY ("assignedToEmployeeId") REFERENCES "Employee"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "SimCard" ADD CONSTRAINT "SimCard_customerId_fkey" FOREIGN KEY ("customerId") REFERENCES "GpsCustomer"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "GpsDeviceRequest" ADD CONSTRAINT "GpsDeviceRequest_customerId_fkey" FOREIGN KEY ("customerId") REFERENCES "GpsCustomer"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "GpsDeviceRequest" ADD CONSTRAINT "GpsDeviceRequest_employeeId_fkey" FOREIGN KEY ("employeeId") REFERENCES "Employee"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "GpsDeviceRequest" ADD CONSTRAINT "GpsDeviceRequest_adminId_fkey" FOREIGN KEY ("adminId") REFERENCES "Employee"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "GpsDeviceRequest" ADD CONSTRAINT "GpsDeviceRequest_simCardId_fkey" FOREIGN KEY ("simCardId") REFERENCES "SimCard"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "GpsRenewalRequest" ADD CONSTRAINT "GpsRenewalRequest_customerId_fkey" FOREIGN KEY ("customerId") REFERENCES "GpsCustomer"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "GpsRenewalRequest" ADD CONSTRAINT "GpsRenewalRequest_deviceRequestId_fkey" FOREIGN KEY ("deviceRequestId") REFERENCES "GpsDeviceRequest"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "GpsRenewalRequest" ADD CONSTRAINT "GpsRenewalRequest_employeeId_fkey" FOREIGN KEY ("employeeId") REFERENCES "Employee"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "GpsRenewalRequest" ADD CONSTRAINT "GpsRenewalRequest_adminId_fkey" FOREIGN KEY ("adminId") REFERENCES "Employee"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "GpsMaintenanceRequest" ADD CONSTRAINT "GpsMaintenanceRequest_customerId_fkey" FOREIGN KEY ("customerId") REFERENCES "GpsCustomer"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "GpsMaintenanceRequest" ADD CONSTRAINT "GpsMaintenanceRequest_employeeId_fkey" FOREIGN KEY ("employeeId") REFERENCES "Employee"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "GpsMaintenanceRequest" ADD CONSTRAINT "GpsMaintenanceRequest_adminId_fkey" FOREIGN KEY ("adminId") REFERENCES "Employee"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "Quotation" ADD CONSTRAINT "Quotation_createdByEmployeeId_fkey" FOREIGN KEY ("createdByEmployeeId") REFERENCES "Employee"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "QuotationItem" ADD CONSTRAINT "QuotationItem_quotationId_fkey" FOREIGN KEY ("quotationId") REFERENCES "Quotation"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "Attendance" ADD CONSTRAINT "Attendance_employeeId_fkey" FOREIGN KEY ("employeeId") REFERENCES "Employee"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "WorkReport" ADD CONSTRAINT "WorkReport_bookingId_fkey" FOREIGN KEY ("bookingId") REFERENCES "Booking"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "WorkReport" ADD CONSTRAINT "WorkReport_employeeId_fkey" FOREIGN KEY ("employeeId") REFERENCES "Employee"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "SupplierSpecialtyLink" ADD CONSTRAINT "SupplierSpecialtyLink_supplierId_fkey" FOREIGN KEY ("supplierId") REFERENCES "Supplier"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "SupplierSpecialtyLink" ADD CONSTRAINT "SupplierSpecialtyLink_specialtyId_fkey" FOREIGN KEY ("specialtyId") REFERENCES "SupplierSpecialty"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "SupplierRating" ADD CONSTRAINT "SupplierRating_supplierId_fkey" FOREIGN KEY ("supplierId") REFERENCES "Supplier"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "SupplierRating" ADD CONSTRAINT "SupplierRating_ratedById_fkey" FOREIGN KEY ("ratedById") REFERENCES "Employee"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "Mission" ADD CONSTRAINT "Mission_bookingId_fkey" FOREIGN KEY ("bookingId") REFERENCES "Booking"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "MissionEvent" ADD CONSTRAINT "MissionEvent_missionId_fkey" FOREIGN KEY ("missionId") REFERENCES "Mission"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "ProcurementRequest" ADD CONSTRAINT "ProcurementRequest_requestedById_fkey" FOREIGN KEY ("requestedById") REFERENCES "Employee"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "ProcurementRequest" ADD CONSTRAINT "ProcurementRequest_bookingId_fkey" FOREIGN KEY ("bookingId") REFERENCES "Booking"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "ProcurementRequest" ADD CONSTRAINT "ProcurementRequest_fulfilledById_fkey" FOREIGN KEY ("fulfilledById") REFERENCES "Employee"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "ProcurementItem" ADD CONSTRAINT "ProcurementItem_requestId_fkey" FOREIGN KEY ("requestId") REFERENCES "ProcurementRequest"("id") ON DELETE CASCADE ON UPDATE CASCADE;

