-- AlterTable: add custom shift time range to Employee
ALTER TABLE "Employee" ADD COLUMN "shiftStart" TEXT;
ALTER TABLE "Employee" ADD COLUMN "shiftEnd" TEXT;
