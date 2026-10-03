#!/bin/bash
# استيراد بيانات الزبائن من ملف CSV مباشرة لقاعدة بيانات النظام (بدون أي بيانات
# زبائن مكتوبة بهذا الملف نفسه — آمن ينرفع لـ Git). يشتغل ضد حاوية قاعدة البيانات
# مباشرة عن طريق docker، يتجاهل أي رقم هاتف مكرر (ON CONFLICT DO NOTHING).
#
# الاستخدام:
#   ./import-customers.sh customers_clean.csv
#
# صيغة ملف CSV المطلوبة (سطر أول عناوين، بترميز UTF-8):
#   name,phone,location
#   محمد أحمد العلي,07801234567,بغداد - الكرادة
#
# بعد الاستيراد احذف ملف الـ CSV من السيرفر يدوياً — هذا السكربت يحذفه بس من
# داخل حاوية قاعدة البيانات المؤقتة، مو من مكان رفعه الأصلي.
set -e

CSV_FILE="$1"
DB_CONTAINER="${DB_CONTAINER:-staff_mange-db-1}"

if [ -z "$CSV_FILE" ]; then
  echo "الاستخدام: $0 <csv-file>"
  exit 1
fi
if [ ! -f "$CSV_FILE" ]; then
  echo "==> الملف غير موجود: $CSV_FILE"
  exit 1
fi

echo "==> نسخ الملف داخل حاوية قاعدة البيانات..."
docker cp "$CSV_FILE" "$DB_CONTAINER:/tmp/customers_import.csv"

echo "==> استيراد البيانات (تجاهل تلقائي لأي رقم هاتف مكرر)..."
docker exec -i "$DB_CONTAINER" psql -U staffmange -d staffmange <<'SQL'
BEGIN;

CREATE TEMP TABLE customer_import (name TEXT, phone TEXT, location TEXT);
\copy customer_import FROM '/tmp/customers_import.csv' WITH (FORMAT csv, HEADER true)

INSERT INTO "Customer" (id, name, phone, location)
SELECT gen_random_uuid()::text, btrim(name), btrim(phone), NULLIF(btrim(location), '')
FROM customer_import
WHERE phone IS NOT NULL AND btrim(phone) <> ''
ON CONFLICT (phone) DO NOTHING;

COMMIT;
SQL

echo "==> حذف الملف المؤقت من داخل الحاوية..."
docker exec "$DB_CONTAINER" rm -f /tmp/customers_import.csv

echo "==> خلص الاستيراد. تحقق من العدد:"
docker exec "$DB_CONTAINER" psql -U staffmange -d staffmange -c 'SELECT COUNT(*) AS total_customers FROM "Customer"'
