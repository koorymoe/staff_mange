-- Remove existing employee skill rows (they referenced services directly; will be re-created per skill)
DELETE FROM "EmployeeSkill";

-- DropForeignKey
ALTER TABLE "EmployeeSkill" DROP CONSTRAINT "EmployeeSkill_serviceId_fkey";

-- DropIndex
DROP INDEX "EmployeeSkill_employeeId_serviceId_key";

-- DropIndex
DROP INDEX "EmployeeSkill_serviceId_idx";

-- AlterTable
ALTER TABLE "Employee" ADD COLUMN     "hasDrivingLicense" BOOLEAN NOT NULL DEFAULT false,
ADD COLUMN     "hasSafetyCertificate" BOOLEAN NOT NULL DEFAULT false;

-- AlterTable
ALTER TABLE "EmployeeSkill" DROP COLUMN "serviceId",
ADD COLUMN     "skillId" TEXT NOT NULL;

-- CreateTable
CREATE TABLE "Skill" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "serviceId" TEXT NOT NULL,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "Skill_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE INDEX "Skill_serviceId_idx" ON "Skill"("serviceId");

-- CreateIndex
CREATE UNIQUE INDEX "Skill_serviceId_name_key" ON "Skill"("serviceId", "name");

-- CreateIndex
CREATE INDEX "EmployeeSkill_skillId_idx" ON "EmployeeSkill"("skillId");

-- CreateIndex
CREATE UNIQUE INDEX "EmployeeSkill_employeeId_skillId_key" ON "EmployeeSkill"("employeeId", "skillId");

-- AddForeignKey
ALTER TABLE "Skill" ADD CONSTRAINT "Skill_serviceId_fkey" FOREIGN KEY ("serviceId") REFERENCES "Service"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "EmployeeSkill" ADD CONSTRAINT "EmployeeSkill_skillId_fkey" FOREIGN KEY ("skillId") REFERENCES "Skill"("id") ON DELETE CASCADE ON UPDATE CASCADE;
