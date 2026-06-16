-- AlterTable
ALTER TABLE "Booking" ADD COLUMN     "confirmedByEmployeeId" TEXT;

-- AddForeignKey
ALTER TABLE "Booking" ADD CONSTRAINT "Booking_confirmedByEmployeeId_fkey" FOREIGN KEY ("confirmedByEmployeeId") REFERENCES "Employee"("id") ON DELETE SET NULL ON UPDATE CASCADE;
