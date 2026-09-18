#!/bin/bash
# ═══ ربط زبائن الجي بي اس بجدول الزبائن الموحّد ═══
#
# ٤٠٦ من ٤٠٧ زبون جي بي اس (GpsCustomer) عندهم أصلاً نفس رقم الهاتف
# بجدول الزبائن العام (Customer) — نفس الشخص عنده سجلان. هذا يوحّدهم:
#
#   ١. أي زبون جي بي اس ماله زبون بجدول Customer (الحالة الوحيدة
#      المتبقية) → ينضاف بصف جديد بنفس الاسم والهاتف.
#   ٢. كل زبون جي بي اس → يتوسم بخدمة "GPS" (CustomerServiceTag) على
#      صف Customer المطابق — هذا نفس الوسم الي تعتمد عليه شاشة
#      «الزبائن» بتبويب «GPS» (Customers.tsx) أصلاً.
#   ٣. رقم الجهاز وتاريخ انتهاء الاشتراك (من أحدث طلب جي بي اس له)
#      يُنسخان لجدول CustomerGpsInfo — يطلعون بنفس الشاشة.
#
# ⚠️ **ما يمس** جدول GpsDeviceRequest ولا SimCard ولا شاشة متابعة
# الاشتراكات (GpsFollowUp.tsx) — تلك تبقى تشتغل بجدول GpsCustomer
# القديم بالضبط متل ما هي. هذا فقط يضيف الظهور بقائمة الزبائن
# الموحّدة، بلا ما يغيّر خلفية تتبع الأجهزة.
#
# ⚠️ آمن للتكرار: كل إدراج بـON CONFLICT DO NOTHING/UPDATE، وتشغيله
# أكثر من مرة ما يكرر ولا وسم ولا زبون.
#
# الاستخدام:  ./link-gps-customers.sh --confirm
set -e
DB_CONTAINER="${DB_CONTAINER:-staff_mange-db-1}"

if [ "$1" != "--confirm" ]; then
  echo "هذا السكربت يكتب على قاعدة البيانات (إدراج وتوسيم بس، ماكو حذف)."
  echo "شغّل check-gps-customer-overlap.sh أولاً وتأكد من الأرقام."
  echo "لو متأكد: ./link-gps-customers.sh --confirm"
  exit 1
fi

echo "==> يربط زبائن الجي بي اس بجدول الزبائن الموحّد..."

docker exec -i "$DB_CONTAINER" psql -U staffmange -d staffmange <<'SQL'
BEGIN;

-- ١. زبون جي بي اس بلا مطابقة بجدول Customer → ينضاف بصف جديد
INSERT INTO "Customer" (id, name, phone)
SELECT gen_random_uuid()::text, btrim(g."fullName"), btrim(g.phone)
FROM "GpsCustomer" g
WHERE btrim(g.phone) <> ''
  AND NOT EXISTS (SELECT 1 FROM "Customer" c WHERE c.phone = btrim(g.phone))
ON CONFLICT (phone) DO NOTHING;

-- ٢. توسيم كل زبون جي بي اس بخدمة "GPS" على صف Customer المطابق
INSERT INTO "CustomerServiceTag" (id, "customerId", service)
SELECT gen_random_uuid()::text, c.id, 'GPS'
FROM "GpsCustomer" g
JOIN "Customer" c ON c.phone = btrim(g.phone)
ON CONFLICT ("customerId", service) DO NOTHING;

-- ٣. رقم الجهاز وتاريخ انتهاء الاشتراك — من أحدث طلب جي بي اس لكل زبون
INSERT INTO "CustomerGpsInfo" (id, "customerId", "gpsNumber", "subscriptionEnd")
SELECT gen_random_uuid()::text, c.id, latest."gpsNumber", latest."subscriptionEnd"
FROM "GpsCustomer" g
JOIN "Customer" c ON c.phone = btrim(g.phone)
JOIN LATERAL (
  SELECT d."gpsNumber", d."subscriptionEnd"
  FROM "GpsDeviceRequest" d
  WHERE d."customerId" = g.id
  ORDER BY d."subscriptionEnd" DESC NULLS LAST, d."createdAt" DESC
  LIMIT 1
) latest ON true
ON CONFLICT ("customerId") DO UPDATE SET
  "gpsNumber" = COALESCE(EXCLUDED."gpsNumber", "CustomerGpsInfo"."gpsNumber"),
  "subscriptionEnd" = COALESCE(EXCLUDED."subscriptionEnd", "CustomerGpsInfo"."subscriptionEnd");

COMMIT;
SQL

echo "==> خلص. الأعداد بعد الربط:"
docker exec "$DB_CONTAINER" psql -U staffmange -d staffmange -c '
SELECT
  (SELECT COUNT(*) FROM "CustomerServiceTag" WHERE service = '"'"'GPS'"'"') AS "زبائن موسومين بـGPS",
  (SELECT COUNT(*) FROM "CustomerGpsInfo") AS "صفوف معلومات الجي بي اس",
  (SELECT COUNT(*) FROM "GpsCustomer") AS "إجمالي جدول الجي بي اس القديم";
'
