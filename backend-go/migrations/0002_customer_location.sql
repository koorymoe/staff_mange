-- يضيف إحداثيات الموقع المحفوظة لكل زبون، حتى يترحل موقعه تلقائياً للحجز الجديد
-- بدل ما يعيد موظف المبيعات تحديد نفس الموقع من الصفر بكل مرة.
--
-- شغّل هذا الملف مرة واحدة على قاعدة البيانات:
--   psql "$DATABASE_URL" -f backend-go/migrations/0002_customer_location.sql

ALTER TABLE "Customer"
  ADD COLUMN IF NOT EXISTS "mapLatitude" DOUBLE PRECISION,
  ADD COLUMN IF NOT EXISTS "mapLongitude" DOUBLE PRECISION;
