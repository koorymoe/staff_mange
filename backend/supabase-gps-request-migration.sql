-- AlterTable: add device/residence identifiers used on the GPS device request workflow + print form
ALTER TABLE "GpsDeviceRequest" ADD COLUMN "gpsNumber" TEXT;
ALTER TABLE "GpsDeviceRequest" ADD COLUMN "residenceCardNumber" TEXT;
