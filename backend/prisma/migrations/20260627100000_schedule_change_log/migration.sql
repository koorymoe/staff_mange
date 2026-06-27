-- AlterTable: remove pendingScheduledAt
ALTER TABLE "Booking" DROP COLUMN IF EXISTS "pendingScheduledAt";

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

-- CreateIndex
CREATE INDEX "ScheduleChangeLog_bookingId_idx" ON "ScheduleChangeLog"("bookingId");
CREATE INDEX "ScheduleChangeLog_changedById_idx" ON "ScheduleChangeLog"("changedById");

-- AddForeignKey
ALTER TABLE "ScheduleChangeLog" ADD CONSTRAINT "ScheduleChangeLog_bookingId_fkey" FOREIGN KEY ("bookingId") REFERENCES "Booking"("id") ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE "ScheduleChangeLog" ADD CONSTRAINT "ScheduleChangeLog_changedById_fkey" FOREIGN KEY ("changedById") REFERENCES "Employee"("id") ON DELETE RESTRICT ON UPDATE CASCADE;
