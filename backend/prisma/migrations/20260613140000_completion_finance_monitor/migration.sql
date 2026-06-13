-- AlterEnum
ALTER TYPE "BookingStatus" ADD VALUE 'COMPLETED';

-- AlterEnum
-- This migration adds more than one value to an enum.
-- With PostgreSQL versions 11 and earlier, this is not possible
-- in a single migration. This can be worked around by creating
-- multiple migrations, each migration adding only one value to
-- the enum.


ALTER TYPE "EmployeeRole" ADD VALUE 'MONITOR';
ALTER TYPE "EmployeeRole" ADD VALUE 'FINANCE';

-- AlterTable
ALTER TABLE "Booking" ADD COLUMN     "amountCollected" DOUBLE PRECISION,
ADD COLUMN     "amountVerified" BOOLEAN NOT NULL DEFAULT false,
ADD COLUMN     "assignedVehicle" TEXT,
ADD COLUMN     "completedAt" TIMESTAMP(3),
ADD COLUMN     "completionNotes" TEXT;

