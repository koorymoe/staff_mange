#!/bin/bash
# استيراد زبائن الجي بي اس من نظام التتبع القديم — ملف عام بدون أي بيانات
# زبائن (آمن ينرفع لـ Git). يقرأ ملف CSV واحد ويدخل الزبائن لجدول "Customer"
# الموحّد (بنفس الكود الموحّد CUST-xxxxx لكل الزبائن)، ويوسمهم بخدمة "GPS"،
# ويحفظ معلوماتهم الإضافية (رقم الجهاز، تاريخ انتهاء الاشتراك) بجدول منفصل.
#
# لو الزبون موجود أصلاً بجدول الزبائن العام (نفس رقم الهاتف)، ما ينعمل له
# زبون جديد — بس ينضاف له وسم "GPS" ومعلوماته الإضافية، فيصير عنده كود واحد
# بس يستخدم بكل الخدمات.
#
# الاستخدام:
#   ./import-gps-customers.sh gps_customers_clean.csv
#
# صيغة ملف الـ CSV المطلوبة (UTF-8، سطر عناوين أول):
#   name,phone,gps_number,device_id,subscription_end
#   (subscription_end بصيغة YYYY-MM-DD أو فارغ)
#
# بعد الاستيراد احذف الملف يدوياً من السيرفر.
set -e

CSV="$1"
DB_CONTAINER="${DB_CONTAINER:-staff_mange-db-1}"

if [ -z "$CSV" ]; then
  echo "الاستخدام: $0 <gps_customers.csv>"
  exit 1
fi
[ -f "$CSV" ] || { echo "==> الملف غير موجود: $CSV"; exit 1; }

echo "==> نسخ الملف داخل حاوية قاعدة البيانات..."
docker cp "$CSV" "$DB_CONTAINER:/tmp/gps_customers_import.csv"

echo "==> استيراد زبائن الجي بي اس (يربط تلقائياً بالزبائن الموجودين عن طريق الهاتف، وينشئ زبون جديد لو ماكو)..."
docker exec -i "$DB_CONTAINER" psql -U staffmange -d staffmange <<'SQL'
BEGIN;
CREATE TEMP TABLE gps_customer_import (
  name TEXT, phone TEXT, gps_number TEXT, device_id TEXT, subscription_end TEXT
);
\copy gps_customer_import FROM '/tmp/gps_customers_import.csv' WITH (FORMAT csv, HEADER true)

-- إنشاء أي زبون غير موجود أصلاً بجدول الزبائن العام (نفس الكود الموحّد للكل)
INSERT INTO "Customer" (id, name, phone)
SELECT gen_random_uuid()::text, btrim(name), phone
FROM gps_customer_import
WHERE phone IS NOT NULL AND btrim(phone) <> ''
ON CONFLICT (phone) DO NOTHING;

-- وسم كل زبون بخدمة "GPS" — بدون كود منفصل، نفس كود الزبون الموحّد
INSERT INTO "CustomerServiceTag" (id, "customerId", service)
SELECT gen_random_uuid()::text, c.id, 'GPS'
FROM gps_customer_import gi
JOIN "Customer" c ON c.phone = gi.phone
ON CONFLICT ("customerId", service) DO NOTHING;

-- معلومات الجي بي اس الإضافية (رقم الجهاز، تاريخ انتهاء الاشتراك)
INSERT INTO "CustomerGpsInfo" (id, "customerId", "gpsNumber", "deviceId", "subscriptionEnd")
SELECT
  gen_random_uuid()::text,
  c.id,
  NULLIF(btrim(gi.gps_number), ''),
  NULLIF(btrim(gi.device_id), ''),
  NULLIF(btrim(gi.subscription_end), '')::timestamp
FROM gps_customer_import gi
JOIN "Customer" c ON c.phone = gi.phone
ON CONFLICT ("customerId") DO UPDATE SET
  "gpsNumber" = EXCLUDED."gpsNumber",
  "deviceId" = EXCLUDED."deviceId",
  "subscriptionEnd" = EXCLUDED."subscriptionEnd";
COMMIT;
SQL

echo "==> حذف الملف المؤقت من داخل الحاوية..."
docker exec "$DB_CONTAINER" rm -f /tmp/gps_customers_import.csv

echo "==> خلص الاستيراد. تحقق من الأعداد:"
docker exec "$DB_CONTAINER" psql -U staffmange -d staffmange -c '
SELECT
  (SELECT COUNT(*) FROM "CustomerServiceTag" WHERE service = '"'"'GPS'"'"') AS gps_tagged_customers,
  (SELECT COUNT(*) FROM "CustomerGpsInfo") AS gps_info_rows;
'
