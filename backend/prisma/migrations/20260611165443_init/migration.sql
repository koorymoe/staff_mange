-- CreateEnum
CREATE TYPE "ProjectPriority" AS ENUM ('NORMAL', 'URGENT');

-- CreateEnum
CREATE TYPE "BookingStatus" AS ENUM ('PENDING', 'CONFIRMED', 'CANCELLED');

-- CreateEnum
CREATE TYPE "EquipmentStatus" AS ENUM ('NOT_READY', 'READY');

-- CreateEnum
CREATE TYPE "Shift" AS ENUM ('MORNING', 'EVENING');

-- CreateEnum
CREATE TYPE "TechnicianRole" AS ENUM ('TECH_1', 'TECH_2', 'TECH_3');

-- CreateEnum
CREATE TYPE "EmployeeStatus" AS ENUM ('ACTIVE', 'INACTIVE');

-- CreateTable
CREATE TABLE "Service" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "category" TEXT,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "Service_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "Customer" (
    "id" TEXT NOT NULL,
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
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "Employee_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "EmployeeSkill" (
    "id" TEXT NOT NULL,
    "employeeId" TEXT NOT NULL,
    "serviceId" TEXT NOT NULL,
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
    "adminNotes" TEXT,
    "equipmentStatus" "EquipmentStatus" NOT NULL DEFAULT 'NOT_READY',
    "shift" "Shift",
    "deviceCount" INTEGER,
    "inspectionSupervisorId" TEXT,
    "projectSupervisorId" TEXT,
    "projectCar" TEXT,
    "crewNotes" TEXT,
    "mapLocation" TEXT,
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

-- CreateIndex
CREATE UNIQUE INDEX "Service_name_key" ON "Service"("name");

-- CreateIndex
CREATE INDEX "EmployeeSkill_serviceId_idx" ON "EmployeeSkill"("serviceId");

-- CreateIndex
CREATE UNIQUE INDEX "EmployeeSkill_employeeId_serviceId_key" ON "EmployeeSkill"("employeeId", "serviceId");

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

-- AddForeignKey
ALTER TABLE "EmployeeSkill" ADD CONSTRAINT "EmployeeSkill_employeeId_fkey" FOREIGN KEY ("employeeId") REFERENCES "Employee"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "EmployeeSkill" ADD CONSTRAINT "EmployeeSkill_serviceId_fkey" FOREIGN KEY ("serviceId") REFERENCES "Service"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "Booking" ADD CONSTRAINT "Booking_customerId_fkey" FOREIGN KEY ("customerId") REFERENCES "Customer"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "Booking" ADD CONSTRAINT "Booking_serviceId_fkey" FOREIGN KEY ("serviceId") REFERENCES "Service"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "Booking" ADD CONSTRAINT "Booking_transferEmployeeId_fkey" FOREIGN KEY ("transferEmployeeId") REFERENCES "Employee"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "Booking" ADD CONSTRAINT "Booking_inspectionSupervisorId_fkey" FOREIGN KEY ("inspectionSupervisorId") REFERENCES "Employee"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "Booking" ADD CONSTRAINT "Booking_projectSupervisorId_fkey" FOREIGN KEY ("projectSupervisorId") REFERENCES "Employee"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "BookingAssignment" ADD CONSTRAINT "BookingAssignment_bookingId_fkey" FOREIGN KEY ("bookingId") REFERENCES "Booking"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "BookingAssignment" ADD CONSTRAINT "BookingAssignment_employeeId_fkey" FOREIGN KEY ("employeeId") REFERENCES "Employee"("id") ON DELETE RESTRICT ON UPDATE CASCADE;
