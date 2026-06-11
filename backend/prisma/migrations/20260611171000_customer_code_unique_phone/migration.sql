-- AlterTable
ALTER TABLE "Customer" ADD COLUMN     "customerCode" SERIAL NOT NULL;

-- CreateIndex
CREATE UNIQUE INDEX "Customer_customerCode_key" ON "Customer"("customerCode");

-- CreateIndex
CREATE UNIQUE INDEX "Customer_phone_key" ON "Customer"("phone");

