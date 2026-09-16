#!/bin/bash
# استيراد شرائح (سيمات) الجي بي اس القديمة داخل جدول "SimCard" — يربط كل
# شريحة بالزبون تلقائياً لو رقمها يطابق "رقم الجي بي اس" المستورد سابقاً
# بجدول طلبات الأجهزة (GpsDeviceRequest).
#
# ملف عام بدون أي بيانات زبائن (آمن ينرفع لـ Git).
#
# ⚠️ ملاحظة عن الحالة: جدول SimCard بالنظام عنده حالتين بس (متوفرة/مستخدمة)،
# بينما البيانات القديمة فيها 4 حالات (مفعل، تم الحرق، منتهي، غير موجود).
# "مفعل" ← مستخدمة (IN_USE)، الباقي ← متوفرة (AVAILABLE) بس نحفظ الحالة
# الحقيقية بالعربي داخل حقل الملاحظات حتى ما نضيع المعلومة.
#
# الاستخدام:
#   ./import-gps-sim-cards.sh sim_cards.csv
#
# صيغة ملف الـ CSV المطلوبة (UTF-8، سطر عناوين أول):
#   simNumber,iccid,operator,statusArabic
#
# بعد الاستيراد احذف الملف يدوياً من السيرفر.
set -e

CSV="$1"
DB_CONTAINER="${DB_CONTAINER:-staff_mange-db-1}"

if [ -z "$CSV" ]; then
  echo "الاستخدام: $0 <sim_cards.csv>"
  exit 1
fi
[ -f "$CSV" ] || { echo "==> الملف غير موجود: $CSV"; exit 1; }

echo "==> نسخ الملف داخل حاوية قاعدة البيانات..."
docker cp "$CSV" "$DB_CONTAINER:/tmp/sim_cards_import.csv"

echo "==> استيراد الشرائح (يربط بالزبون تلقائياً لو رقم الشريحة يطابق رقم جي بي اس مسجّل)..."
docker exec -i "$DB_CONTAINER" psql -U staffmange -d staffmange <<'SQL'
BEGIN;
CREATE TEMP TABLE sim_import (simNumber TEXT, iccid TEXT, operator TEXT, statusArabic TEXT);
\copy sim_import FROM '/tmp/sim_cards_import.csv' WITH (FORMAT csv, HEADER true)

INSERT INTO "SimCard" (id, "simNumber", iccid, operator, status, "customerId", notes)
SELECT
  gen_random_uuid()::text,
  si.simnumber,
  NULLIF(btrim(si.iccid), ''),
  si.operator::"SimOperator",
  (CASE WHEN si.statusarabic = 'مفعل' THEN 'IN_USE' ELSE 'AVAILABLE' END)::"SimStatus",
  (SELECT gdr."customerId" FROM "GpsDeviceRequest" gdr WHERE gdr."gpsNumber" = si.simnumber LIMIT 1),
  'استيراد تاريخي — الحالة الأصلية بالنظام القديم: ' || COALESCE(NULLIF(btrim(si.statusarabic), ''), 'غير معروفة')
FROM sim_import si
WHERE si.simnumber IS NOT NULL AND btrim(si.simnumber) <> ''
  AND NOT EXISTS (SELECT 1 FROM "SimCard" existing WHERE existing."simNumber" = si.simnumber);
COMMIT;
SQL

echo "==> حذف الملف المؤقت من داخل الحاوية..."
docker exec "$DB_CONTAINER" rm -f /tmp/sim_cards_import.csv

echo "==> خلص الاستيراد. تحقق من العدد:"
docker exec "$DB_CONTAINER" psql -U staffmange -d staffmange -c '
SELECT
  (SELECT COUNT(*) FROM "SimCard") AS total_sims,
  (SELECT COUNT(*) FROM "SimCard" WHERE "customerId" IS NOT NULL) AS linked_to_customer;
'
