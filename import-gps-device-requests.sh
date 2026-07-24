#!/bin/bash
# استيراد سجل أجهزة الجي بي اس القديم (رقم الجهاز، رقم الجي بي اس، تاريخ
# انتهاء الاشتراك) كـ"طلبات جهاز" تاريخية داخل وحدة الجي بي اس الفعلية —
# حتى يطلع سجل كل زبون كامل (جهازه، اشتراكه) لما يفتحون ملفه من الوحدة.
#
# ملف عام بدون أي بيانات زبائن (آمن ينرفع لـ Git). لازم تشغّل
# import-gps-module-customers.sh قبله (لازم الزبون يكون موجود بجدول
# GpsCustomer أول).
#
# ⚠️ افتراضات لازم تعرفها (البيانات القديمة ما فيها هذي التفاصيل، فحطينا
# قيم افتراضية معقولة — عدّلها لاحقاً يدوياً لو تحتاج دقة أكثر):
#   - نوع الشراء: "جهاز + سيم" (DEVICE_SIM) لكل السجلات
#   - نوع الاشتراك: "سنوي" (YEARLY) لكل السجلات (غير معروف من البيانات القديمة)
#   - حالة الاشتراك: "منتهي" لو تاريخ الانتهاء بالماضي، "فعّال" لو بالمستقبل
#   - حالة الطلب: "تم التسليم" (DELIVERED) — لأنها كانت شغالة فعلياً
#   - الموظف المسؤول عن الطلب: حساب المالك (كـ"مسجّل تاريخي" بدون موظف حقيقي)
#
# الاستخدام:
#   ./import-gps-device-requests.sh gps_device_requests.csv
#
# صيغة ملف الـ CSV المطلوبة (UTF-8، سطر عناوين أول):
#   fullName,phone,gpsNumber,deviceId,subscriptionEnd
#
# بعد الاستيراد احذف الملف يدوياً من السيرفر.
set -e

CSV="$1"
DB_CONTAINER="${DB_CONTAINER:-staff_mange-db-1}"

if [ -z "$CSV" ]; then
  echo "الاستخدام: $0 <gps_device_requests.csv>"
  exit 1
fi
[ -f "$CSV" ] || { echo "==> الملف غير موجود: $CSV"; exit 1; }

echo "==> نسخ الملف داخل حاوية قاعدة البيانات..."
docker cp "$CSV" "$DB_CONTAINER:/tmp/gps_device_requests_import.csv"

echo "==> استيراد طلبات الأجهزة التاريخية (يربط بالزبون عن طريق الهاتف)..."
docker exec -i "$DB_CONTAINER" psql -U staffmange -d staffmange <<'SQL'
BEGIN;
CREATE TEMP TABLE gps_req_import (fullName TEXT, phone TEXT, gpsNumber TEXT, deviceId TEXT, subscriptionEnd TEXT);
\copy gps_req_import FROM '/tmp/gps_device_requests_import.csv' WITH (FORMAT csv, HEADER true)

INSERT INTO "GpsDeviceRequest" (
  id, "customerId", "employeeId", "purchaseType", "subscriptionType",
  "subscriptionEnd", "subscriptionStatus", status, "gpsNumber", notes,
  "isChecked", "isActivated", "isDelivered", "deliveredAt", "createdAt"
)
SELECT
  gen_random_uuid()::text,
  c.id,
  (SELECT id FROM "Employee" WHERE role = 'OWNER' LIMIT 1),
  'DEVICE_SIM',
  'YEARLY',
  NULLIF(btrim(gi.subscriptionend), '')::timestamp,
  CASE
    WHEN NULLIF(btrim(gi.subscriptionend), '')::timestamp IS NULL THEN 'ACTIVE'
    WHEN NULLIF(btrim(gi.subscriptionend), '')::timestamp < now() THEN 'EXPIRED'
    ELSE 'ACTIVE'
  END,
  'DELIVERED',
  NULLIF(btrim(gi.gpsnumber), ''),
  'استيراد تاريخي من النظام القديم — رقم الجهاز: ' || COALESCE(NULLIF(btrim(gi.deviceid), ''), 'غير معروف'),
  true, true, true,
  NULLIF(btrim(gi.subscriptionend), '')::timestamp,
  now()
FROM gps_req_import gi
JOIN "GpsCustomer" c ON c.phone = gi.phone
WHERE NOT EXISTS (
  SELECT 1 FROM "GpsDeviceRequest" existing
  WHERE existing."customerId" = c.id AND existing.notes LIKE 'استيراد تاريخي%'
);
COMMIT;
SQL

echo "==> حذف الملف المؤقت من داخل الحاوية..."
docker exec "$DB_CONTAINER" rm -f /tmp/gps_device_requests_import.csv

echo "==> خلص الاستيراد. تحقق من العدد:"
docker exec "$DB_CONTAINER" psql -U staffmange -d staffmange -c '
SELECT COUNT(*) AS imported_device_requests FROM "GpsDeviceRequest" WHERE notes LIKE '"'"'استيراد تاريخي%'"'"';
'
