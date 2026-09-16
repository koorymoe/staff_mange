#!/bin/bash
# استيراد زبائن الجي بي اس القدامى داخل جدول "GpsCustomer" — نفس الجدول الي
# تستخدمه وحدة الجي بي اس الفعلية (طلبات الأجهزة، السيمات، التجديد، الصيانة)
# حتى يصير الموظف يقدر يختارهم مباشرة عند تسجيل طلب جديد إلهم.
#
# ملف عام بدون أي بيانات زبائن (آمن ينرفع لـ Git). يربط بالاسم ورقم الهاتف
# فقط — باقي الحقول (اسم الأب/الجد، العنوان، صور الهوية) تنضاف لاحقاً يدوياً
# لو احتاج الموظف يكملها وقت طلب جديد.
#
# هذا الاستيراد منفصل تماماً عن import-gps-customers.sh (الي يرفع لجدول
# الزبائن العام الموحّد) — الاثنين آمن تشغلهم، ما يتعارضون.
#
# الاستخدام:
#   ./import-gps-module-customers.sh gps_module_customers.csv
#
# صيغة ملف الـ CSV المطلوبة (UTF-8، سطر عناوين أول):
#   fullName,phone
#
# بعد الاستيراد احذف الملف يدوياً من السيرفر.
set -e

CSV="$1"
DB_CONTAINER="${DB_CONTAINER:-staff_mange-db-1}"

if [ -z "$CSV" ]; then
  echo "الاستخدام: $0 <gps_module_customers.csv>"
  exit 1
fi
[ -f "$CSV" ] || { echo "==> الملف غير موجود: $CSV"; exit 1; }

echo "==> نسخ الملف داخل حاوية قاعدة البيانات..."
docker cp "$CSV" "$DB_CONTAINER:/tmp/gps_module_customers_import.csv"

echo "==> استيراد زبائن الجي بي اس لجدول وحدة الجي بي اس الفعلية..."
docker exec -i "$DB_CONTAINER" psql -U staffmange -d staffmange <<'SQL'
BEGIN;
CREATE TEMP TABLE gps_module_customer_import (fullName TEXT, phone TEXT);
\copy gps_module_customer_import FROM '/tmp/gps_module_customers_import.csv' WITH (FORMAT csv, HEADER true)

INSERT INTO "GpsCustomer" (id, "fullName", phone)
SELECT gen_random_uuid()::text, btrim(gi.fullname), gi.phone
FROM gps_module_customer_import gi
WHERE gi.phone IS NOT NULL AND btrim(gi.phone) <> ''
  AND NOT EXISTS (SELECT 1 FROM "GpsCustomer" existing WHERE existing.phone = gi.phone);
COMMIT;
SQL

echo "==> حذف الملف المؤقت من داخل الحاوية..."
docker exec "$DB_CONTAINER" rm -f /tmp/gps_module_customers_import.csv

echo "==> خلص الاستيراد. تحقق من العدد:"
docker exec "$DB_CONTAINER" psql -U staffmange -d staffmange -c '
SELECT COUNT(*) AS total_gps_module_customers FROM "GpsCustomer";
'
