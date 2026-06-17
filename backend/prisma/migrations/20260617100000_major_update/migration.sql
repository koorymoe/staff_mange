-- AlterEnum: Add new values to "EmployeeRole"
ALTER TYPE "EmployeeRole" ADD VALUE 'GPS_ADMIN';
ALTER TYPE "EmployeeRole" ADD VALUE 'GPS_ENGINEER';
ALTER TYPE "EmployeeRole" ADD VALUE 'QUALITY_ENGINEER';

-- CreateEnum
CREATE TYPE "ToolStatus" AS ENUM ('AVAILABLE', 'CHECKED_OUT', 'DAMAGED');
CREATE TYPE "ToolRequestStatus" AS ENUM ('PENDING', 'APPROVED', 'REJECTED', 'RETURNED');
CREATE TYPE "ComplaintStatus" AS ENUM ('NEW', 'IN_PROGRESS', 'RESOLVED', 'CLOSED');
CREATE TYPE "SimOperator" AS ENUM ('ZAIN', 'ASIACELL', 'KOREK', 'OTHER');
CREATE TYPE "SimStatus" AS ENUM ('AVAILABLE', 'IN_USE');
CREATE TYPE "GpsRequestStatus" AS ENUM ('PENDING', 'APPROVED', 'REJECTED', 'DELIVERED');
CREATE TYPE "GpsPurchaseType" AS ENUM ('DEVICE_SIM', 'DEVICE_ONLY');
CREATE TYPE "GpsSubscriptionType" AS ENUM ('THREE_MONTHS', 'SIX_MONTHS', 'YEARLY');
CREATE TYPE "GpsSubscriptionStatus" AS ENUM ('ACTIVE', 'EXPIRED', 'CANCELLED');
CREATE TYPE "GpsMaintenanceStatus" AS ENUM ('PENDING', 'IN_PROGRESS', 'COMPLETED');
CREATE TYPE "QuotationStatus" AS ENUM ('NEW', 'SENT', 'ACCEPTED', 'REJECTED');

-- CreateTable: Permission
CREATE TABLE "Permission" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "label" TEXT NOT NULL,

    CONSTRAINT "Permission_pkey" PRIMARY KEY ("id")
);

-- CreateTable: EmployeePermission
CREATE TABLE "EmployeePermission" (
    "id" TEXT NOT NULL,
    "employeeId" TEXT NOT NULL,
    "permissionId" TEXT NOT NULL,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "EmployeePermission_pkey" PRIMARY KEY ("id")
);

-- CreateTable: KpiEvaluation
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

-- CreateTable: CartItem
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

-- CreateTable: PersonalTool
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

-- CreateTable: VehicleTool
CREATE TABLE "VehicleTool" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "barcode" TEXT NOT NULL,
    "vehicleId" TEXT NOT NULL,
    "status" "ToolStatus" NOT NULL DEFAULT 'AVAILABLE',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "VehicleTool_pkey" PRIMARY KEY ("id")
);

-- CreateTable: OnDemandTool
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

-- CreateTable: ToolRequest
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

-- CreateTable: Complaint
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

-- CreateTable: GpsCustomer
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

-- CreateTable: SimCard
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

-- CreateTable: GpsDeviceRequest
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

-- CreateTable: GpsRenewalRequest
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

-- CreateTable: GpsMaintenanceRequest
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

-- CreateTable: GpsSubscriptionPrice
CREATE TABLE "GpsSubscriptionPrice" (
    "id" TEXT NOT NULL,
    "subscriptionType" "GpsSubscriptionType" NOT NULL,
    "price" DOUBLE PRECISION NOT NULL,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "GpsSubscriptionPrice_pkey" PRIMARY KEY ("id")
);

-- CreateTable: Product
CREATE TABLE "Product" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "unit" TEXT,
    "defaultPrice" DOUBLE PRECISION,
    "imageBase64" TEXT,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "Product_pkey" PRIMARY KEY ("id")
);

-- CreateTable: Quotation
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

-- CreateTable: QuotationItem
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

-- CreateIndex: Unique constraints
CREATE UNIQUE INDEX "Permission_name_key" ON "Permission"("name");
CREATE UNIQUE INDEX "EmployeePermission_employeeId_permissionId_key" ON "EmployeePermission"("employeeId", "permissionId");
CREATE UNIQUE INDEX "PersonalTool_barcode_key" ON "PersonalTool"("barcode");
CREATE UNIQUE INDEX "VehicleTool_barcode_key" ON "VehicleTool"("barcode");
CREATE UNIQUE INDEX "OnDemandTool_barcode_key" ON "OnDemandTool"("barcode");
CREATE UNIQUE INDEX "SimCard_simNumber_key" ON "SimCard"("simNumber");
CREATE UNIQUE INDEX "GpsSubscriptionPrice_subscriptionType_key" ON "GpsSubscriptionPrice"("subscriptionType");
CREATE UNIQUE INDEX "Product_name_key" ON "Product"("name");
CREATE UNIQUE INDEX "Quotation_quotationNumber_key" ON "Quotation"("quotationNumber");

-- CreateIndex: Regular indexes
CREATE INDEX "KpiEvaluation_employeeId_idx" ON "KpiEvaluation"("employeeId");
CREATE INDEX "CartItem_bookingId_idx" ON "CartItem"("bookingId");
CREATE INDEX "PersonalTool_employeeId_idx" ON "PersonalTool"("employeeId");
CREATE INDEX "VehicleTool_vehicleId_idx" ON "VehicleTool"("vehicleId");
CREATE INDEX "ToolRequest_employeeId_idx" ON "ToolRequest"("employeeId");
CREATE INDEX "ToolRequest_toolId_idx" ON "ToolRequest"("toolId");
CREATE INDEX "Complaint_customerId_idx" ON "Complaint"("customerId");
CREATE INDEX "Complaint_status_idx" ON "Complaint"("status");
CREATE INDEX "SimCard_status_idx" ON "SimCard"("status");
CREATE INDEX "GpsDeviceRequest_customerId_idx" ON "GpsDeviceRequest"("customerId");
CREATE INDEX "GpsDeviceRequest_status_idx" ON "GpsDeviceRequest"("status");
CREATE INDEX "GpsRenewalRequest_customerId_idx" ON "GpsRenewalRequest"("customerId");
CREATE INDEX "GpsMaintenanceRequest_customerId_idx" ON "GpsMaintenanceRequest"("customerId");
CREATE INDEX "Quotation_status_idx" ON "Quotation"("status");
CREATE INDEX "QuotationItem_quotationId_idx" ON "QuotationItem"("quotationId");

-- AddForeignKey
ALTER TABLE "EmployeePermission" ADD CONSTRAINT "EmployeePermission_employeeId_fkey" FOREIGN KEY ("employeeId") REFERENCES "Employee"("id") ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE "EmployeePermission" ADD CONSTRAINT "EmployeePermission_permissionId_fkey" FOREIGN KEY ("permissionId") REFERENCES "Permission"("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "KpiEvaluation" ADD CONSTRAINT "KpiEvaluation_employeeId_fkey" FOREIGN KEY ("employeeId") REFERENCES "Employee"("id") ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE "KpiEvaluation" ADD CONSTRAINT "KpiEvaluation_evaluatorId_fkey" FOREIGN KEY ("evaluatorId") REFERENCES "Employee"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

ALTER TABLE "CartItem" ADD CONSTRAINT "CartItem_bookingId_fkey" FOREIGN KEY ("bookingId") REFERENCES "Booking"("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "PersonalTool" ADD CONSTRAINT "PersonalTool_employeeId_fkey" FOREIGN KEY ("employeeId") REFERENCES "Employee"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

ALTER TABLE "ToolRequest" ADD CONSTRAINT "ToolRequest_employeeId_fkey" FOREIGN KEY ("employeeId") REFERENCES "Employee"("id") ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE "ToolRequest" ADD CONSTRAINT "ToolRequest_toolId_fkey" FOREIGN KEY ("toolId") REFERENCES "OnDemandTool"("id") ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE "ToolRequest" ADD CONSTRAINT "ToolRequest_approvedById_fkey" FOREIGN KEY ("approvedById") REFERENCES "Employee"("id") ON DELETE SET NULL ON UPDATE CASCADE;

ALTER TABLE "Complaint" ADD CONSTRAINT "Complaint_customerId_fkey" FOREIGN KEY ("customerId") REFERENCES "Customer"("id") ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE "Complaint" ADD CONSTRAINT "Complaint_bookingId_fkey" FOREIGN KEY ("bookingId") REFERENCES "Booking"("id") ON DELETE SET NULL ON UPDATE CASCADE;
ALTER TABLE "Complaint" ADD CONSTRAINT "Complaint_createdByEmployeeId_fkey" FOREIGN KEY ("createdByEmployeeId") REFERENCES "Employee"("id") ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE "Complaint" ADD CONSTRAINT "Complaint_assignedToEmployeeId_fkey" FOREIGN KEY ("assignedToEmployeeId") REFERENCES "Employee"("id") ON DELETE SET NULL ON UPDATE CASCADE;

ALTER TABLE "SimCard" ADD CONSTRAINT "SimCard_customerId_fkey" FOREIGN KEY ("customerId") REFERENCES "GpsCustomer"("id") ON DELETE SET NULL ON UPDATE CASCADE;

ALTER TABLE "GpsDeviceRequest" ADD CONSTRAINT "GpsDeviceRequest_customerId_fkey" FOREIGN KEY ("customerId") REFERENCES "GpsCustomer"("id") ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE "GpsDeviceRequest" ADD CONSTRAINT "GpsDeviceRequest_employeeId_fkey" FOREIGN KEY ("employeeId") REFERENCES "Employee"("id") ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE "GpsDeviceRequest" ADD CONSTRAINT "GpsDeviceRequest_adminId_fkey" FOREIGN KEY ("adminId") REFERENCES "Employee"("id") ON DELETE SET NULL ON UPDATE CASCADE;
ALTER TABLE "GpsDeviceRequest" ADD CONSTRAINT "GpsDeviceRequest_simCardId_fkey" FOREIGN KEY ("simCardId") REFERENCES "SimCard"("id") ON DELETE SET NULL ON UPDATE CASCADE;

ALTER TABLE "GpsRenewalRequest" ADD CONSTRAINT "GpsRenewalRequest_customerId_fkey" FOREIGN KEY ("customerId") REFERENCES "GpsCustomer"("id") ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE "GpsRenewalRequest" ADD CONSTRAINT "GpsRenewalRequest_deviceRequestId_fkey" FOREIGN KEY ("deviceRequestId") REFERENCES "GpsDeviceRequest"("id") ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE "GpsRenewalRequest" ADD CONSTRAINT "GpsRenewalRequest_employeeId_fkey" FOREIGN KEY ("employeeId") REFERENCES "Employee"("id") ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE "GpsRenewalRequest" ADD CONSTRAINT "GpsRenewalRequest_adminId_fkey" FOREIGN KEY ("adminId") REFERENCES "Employee"("id") ON DELETE SET NULL ON UPDATE CASCADE;

ALTER TABLE "GpsMaintenanceRequest" ADD CONSTRAINT "GpsMaintenanceRequest_customerId_fkey" FOREIGN KEY ("customerId") REFERENCES "GpsCustomer"("id") ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE "GpsMaintenanceRequest" ADD CONSTRAINT "GpsMaintenanceRequest_employeeId_fkey" FOREIGN KEY ("employeeId") REFERENCES "Employee"("id") ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE "GpsMaintenanceRequest" ADD CONSTRAINT "GpsMaintenanceRequest_adminId_fkey" FOREIGN KEY ("adminId") REFERENCES "Employee"("id") ON DELETE SET NULL ON UPDATE CASCADE;

ALTER TABLE "Quotation" ADD CONSTRAINT "Quotation_createdByEmployeeId_fkey" FOREIGN KEY ("createdByEmployeeId") REFERENCES "Employee"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

ALTER TABLE "QuotationItem" ADD CONSTRAINT "QuotationItem_quotationId_fkey" FOREIGN KEY ("quotationId") REFERENCES "Quotation"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- Seed: Permission table
INSERT INTO "Permission" ("id", "name", "label") VALUES
  ('perm_staff_mgmt', 'staff_management', 'إدارة الكوادر'),
  ('perm_gps_system', 'gps_system', 'نظام GPS'),
  ('perm_quotation', 'quotation_system', 'نظام عروض الأسعار'),
  ('perm_complaints', 'complaints', 'نظام الشكاوى'),
  ('perm_inventory', 'inventory', 'نظام الجرد'),
  ('perm_kpi_mgmt', 'kpi_management', 'نظام تقييم الأداء'),
  ('perm_booking', 'booking_system', 'نظام الحجوزات'),
  ('perm_finance', 'finance', 'النظام المالي'),
  ('perm_sales', 'sales', 'المبيعات');
