-- AlterTable
ALTER TABLE "Employee" ADD COLUMN     "password" TEXT,
ADD COLUMN     "username" TEXT;

-- CreateIndex
CREATE UNIQUE INDEX "Employee_username_key" ON "Employee"("username");

