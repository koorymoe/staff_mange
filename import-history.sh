#!/bin/bash
# استيراد بيانات تاريخية (حجوزات قديمة + شكاوى قديمة + مواد خاصة) من النظام
# القديم — ملف عام بدون أي بيانات زبائن (آمن ينرفع لـ Git). يقرأ 3 ملفات CSV
# ويدخلها لقاعدة البيانات مباشرة، مع ربط تلقائي بالزبائن الموجودين عن طريق رقم
# الهاتف.
#
# الموظفين القدامى (الي ثبتوا الحجوز أو المسؤولين عن الشكاوى) ينحفظ اسمهم
# نصياً بس — بدون إنشاء أي حساب دخول لهم. لو رجعوا للشركة وانسوى لهم حساب
# جديد بنفس الاسم بالضبط، تكدر تربط سجلاتهم القديمة بحسابهم من صفحة "إدارة
# الكوادر" بزر "🔗 ربط السجلات التاريخية".
#
# الاستخدام:
#   ./import-history.sh bookings_clean.csv complaints_clean.csv materials_clean.csv
#
# صيغ ملفات الـ CSV المطلوبة (UTF-8، سطر عناوين أول):
#   bookings_clean.csv:   phone,employee_name,service_name,notes,vehicle_type,priority,created_at
#   complaints_clean.csv: customer_name,phone,location,supervisor_name,problem_type,reason
#   materials_clean.csv:  name,unit
#
# بعد الاستيراد احذف الملفات الثلاثة يدوياً من السيرفر.
set -e

BOOKINGS_CSV="$1"
COMPLAINTS_CSV="$2"
MATERIALS_CSV="$3"
DB_CONTAINER="${DB_CONTAINER:-staff_mange-db-1}"

if [ -z "$BOOKINGS_CSV" ] || [ -z "$COMPLAINTS_CSV" ] || [ -z "$MATERIALS_CSV" ]; then
  echo "الاستخدام: $0 <bookings.csv> <complaints.csv> <materials.csv>"
  exit 1
fi
for f in "$BOOKINGS_CSV" "$COMPLAINTS_CSV" "$MATERIALS_CSV"; do
  [ -f "$f" ] || { echo "==> الملف غير موجود: $f"; exit 1; }
done

echo "==> نسخ الملفات داخل حاوية قاعدة البيانات..."
docker cp "$BOOKINGS_CSV" "$DB_CONTAINER:/tmp/bookings_import.csv"
docker cp "$COMPLAINTS_CSV" "$DB_CONTAINER:/tmp/complaints_import.csv"
docker cp "$MATERIALS_CSV" "$DB_CONTAINER:/tmp/materials_import.csv"

echo "==> استيراد المواد الخاصة..."
docker exec -i "$DB_CONTAINER" psql -U staffmange -d staffmange <<'SQL'
BEGIN;
CREATE TEMP TABLE material_import (name TEXT, unit TEXT);
\copy material_import FROM '/tmp/materials_import.csv' WITH (FORMAT csv, HEADER true)

INSERT INTO "Product" (id, name, unit)
SELECT gen_random_uuid()::text, btrim(name), NULLIF(btrim(unit), '')
FROM material_import
WHERE name IS NOT NULL AND btrim(name) <> ''
ON CONFLICT (name) DO NOTHING;
COMMIT;
SQL

echo "==> استيراد الحجوزات القديمة (تُربط تلقائياً بالزبائن الموجودين عن طريق الهاتف)..."
docker exec -i "$DB_CONTAINER" psql -U staffmange -d staffmange <<'SQL'
BEGIN;
CREATE TEMP TABLE booking_import (phone TEXT, employee_name TEXT, service_name TEXT, notes TEXT, vehicle_type TEXT, priority TEXT, created_at TEXT);
\copy booking_import FROM '/tmp/bookings_import.csv' WITH (FORMAT csv, HEADER true)

-- إنشاء أي خدمة غير موجودة أصلاً بجدول الخدمات
INSERT INTO "Service" (id, name)
SELECT gen_random_uuid()::text, s.name
FROM (SELECT DISTINCT service_name AS name FROM booking_import WHERE service_name IS NOT NULL AND btrim(service_name) <> '') s
ON CONFLICT (name) DO NOTHING;

-- إدخال الحجوزات — كود ثابت (deterministic) حتى إعادة تشغيل السكربت ما تكرر البيانات
INSERT INTO "Booking" (
  id, code, "customerId", "serviceId", status, "confirmedByName", "vehicleType",
  notes, priority, "createdAt", "adminNotes"
)
SELECT
  gen_random_uuid()::text,
  'OLD-' || substr(md5(bi.phone || COALESCE(bi.created_at, '') || COALESCE(bi.employee_name, '')), 1, 12),
  c.id,
  s.id,
  'COMPLETED',
  NULLIF(btrim(bi.employee_name), ''),
  NULLIF(btrim(bi.vehicle_type), ''),
  NULLIF(btrim(bi.notes), ''),
  CASE WHEN bi.priority = 'URGENT' THEN 'URGENT'::"ProjectPriority" ELSE 'NORMAL'::"ProjectPriority" END,
  COALESCE(NULLIF(bi.created_at, '')::timestamp, now()),
  'استيراد تاريخي من النظام القديم'
FROM booking_import bi
JOIN "Customer" c ON c.phone = bi.phone
LEFT JOIN "Service" s ON s.name = bi.service_name
ON CONFLICT (code) DO NOTHING;
COMMIT;
SQL

echo "==> استيراد الشكاوى القديمة (تُربط تلقائياً بالزبائن، وتُنشئ زبون جديد لو ماكو أصلاً)..."
docker exec -i "$DB_CONTAINER" psql -U staffmange -d staffmange <<SQL
BEGIN;
CREATE TEMP TABLE complaint_import (customer_name TEXT, phone TEXT, location TEXT, supervisor_name TEXT, problem_type TEXT, reason TEXT);
\copy complaint_import FROM '/tmp/complaints_import.csv' WITH (FORMAT csv, HEADER true)

-- لو رقم الزبون مو موجود أصلاً بجدول الزبائن، أنشئه
INSERT INTO "Customer" (id, name, phone, location)
SELECT gen_random_uuid()::text, btrim(customer_name), btrim(phone), NULLIF(btrim(location), '')
FROM complaint_import
WHERE phone IS NOT NULL AND btrim(phone) <> ''
ON CONFLICT (phone) DO NOTHING;

INSERT INTO "Complaint" (
  id, "customerId", type, description, "relatedEmployeeName", status,
  "createdByEmployeeId", "createdAt"
)
SELECT
  gen_random_uuid()::text,
  c.id,
  'OTHER',
  btrim(concat_ws(' — السبب: ', NULLIF(btrim(ci.problem_type), ''), NULLIF(btrim(ci.reason), ''))),
  NULLIF(btrim(ci.supervisor_name), ''),
  'CLOSED',
  (SELECT id FROM "Employee" WHERE role = 'OWNER' LIMIT 1),
  now()
FROM complaint_import ci
JOIN "Customer" c ON c.phone = ci.phone
WHERE NOT EXISTS (
  SELECT 1 FROM "Complaint" existing
  WHERE existing."customerId" = c.id
    AND existing.description = btrim(concat_ws(' — السبب: ', NULLIF(btrim(ci.problem_type), ''), NULLIF(btrim(ci.reason), '')))
);
COMMIT;
SQL

echo "==> حذف الملفات المؤقتة من داخل الحاوية..."
docker exec "$DB_CONTAINER" rm -f /tmp/bookings_import.csv /tmp/complaints_import.csv /tmp/materials_import.csv

echo "==> خلص الاستيراد. تحقق من الأعداد:"
docker exec "$DB_CONTAINER" psql -U staffmange -d staffmange -c '
SELECT
  (SELECT COUNT(*) FROM "Booking" WHERE code LIKE '"'"'OLD-%'"'"') AS old_bookings,
  (SELECT COUNT(*) FROM "Complaint" WHERE description LIKE '"'"'%السبب%'"'"' OR "relatedEmployeeName" IS NOT NULL) AS old_complaints,
  (SELECT COUNT(*) FROM "Product") AS total_products;
'
