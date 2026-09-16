#!/bin/bash
# ══════════════════════════════════════════════════════════════════════════
# استيراد فواتير تاريخية لحجوزات الاستيراد القديم (OLD-…)
# ══════════════════════════════════════════════════════════════════════════
#
# الحجوزات المستوردة من النظام القديم وصلت كلها بحالة COMPLETED بلا فاتورة
# ولا تقرير، فبقت معلّقة بخانة «منجزة وناقصها ورق». فواتيرها موجودة بدفتر
# تدقيق الحسابات (إكسل) — هذا السكربت يدخّلها للنظام.
#
# الربط: كل صف بالإكسل انربط بحجزه عن طريق «رقم الهاتف + التاريخ»، وكود
# الحجز (OLD-xxxxxxxxxxxx) محسوب بنفس معادلة import-history.sh:
#     'OLD-' || substr(md5(phone || created_at || employee_name), 1, 12)
# يعني الربط حتمي — ما بيه تخمين ولا بحث تقريبي بقاعدة البيانات.
#
# ⚠️ الفواتير تنكتب باسم **المالك** مو باسم موظف: هاي تسوية إدارية لشغل صار
# قبل النظام، وحطّها باسم موظف يزوّر إحصائيات أدائه وأرباحه.
#
# الاستخدام:
#   ./import-legacy-invoices.sh فواتير-تاريخية-للاستيراد.csv          # تجربة (ما يحفظ)
#   ./import-legacy-invoices.sh فواتير-تاريخية-للاستيراد.csv --apply  # حفظ فعلي
#
# أعمدة الـ CSV (UTF-8، سطر عناوين أول):
#   accounting_code,booking_code,customer_name,customer_phone,
#   customer_address,net_total,device_count,external_number,invoice_date
#
# إعادة التشغيل آمنة: accountingCode فريد بقاعدة البيانات، والإدخال
# ON CONFLICT DO NOTHING — فالتشغيل الثاني ما يكرر ولا يغيّر شي.
# ══════════════════════════════════════════════════════════════════════════
set -euo pipefail

CSV="${1:-}"
MODE="${2:-}"
DB_CONTAINER="${DB_CONTAINER:-staff_mange-db-1}"

if [ -z "$CSV" ]; then
  echo "الاستخدام: $0 <فواتير-تاريخية-للاستيراد.csv> [--apply]"
  exit 1
fi
[ -f "$CSV" ] || { echo "==> الملف غير موجود: $CSV"; exit 1; }

if [ "$MODE" = "--apply" ]; then
  FINISH="COMMIT;"
  echo "═══ وضع الحفظ الفعلي (--apply) — التغييرات راح تنحفظ ═══"
else
  FINISH="ROLLBACK;"
  echo "═══ وضع التجربة — ولا تغيير راح ينحفظ (زيد --apply للحفظ) ═══"
fi
echo

echo "==> نسخ الملف داخل حاوية قاعدة البيانات..."
docker cp "$CSV" "$DB_CONTAINER:/tmp/legacy_invoices.csv"

docker exec -i "$DB_CONTAINER" psql -U staffmange -d staffmange -v ON_ERROR_STOP=1 <<SQL
BEGIN;

CREATE TEMP TABLE inv_import (
  accounting_code  TEXT,
  booking_code     TEXT,
  customer_name    TEXT,
  customer_phone   TEXT,
  customer_address TEXT,
  net_total        NUMERIC,
  device_count     INTEGER,
  external_number  TEXT,
  invoice_date     TEXT
);
\copy inv_import FROM '/tmp/legacy_invoices.csv' WITH (FORMAT csv, HEADER true)

-- ═══ المالك: صاحب الفواتير التاريخية ═══
-- لو ماكو حساب مالك واحد بالضبط نوقف، لأن الفواتير لازم تنكتب على حساب
-- معروف — مو على أول حساب يطلع بالصدفة.
CREATE TEMP TABLE owner_pick AS
  SELECT id, name FROM "Employee" WHERE role = 'OWNER' ORDER BY "createdAt" LIMIT 1;

DO \$\$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM owner_pick) THEN
    RAISE EXCEPTION 'ماكو حساب بصلاحية OWNER — ما أگدر أحدد صاحب الفواتير';
  END IF;
END \$\$;

\echo ''
\echo '───────────── الفحص قبل الإدخال ─────────────'
\echo ''
\echo '» حساب المالك الي راح تنكتب الفواتير باسمه:'
SELECT name AS "الاسم", id AS "المعرّف" FROM owner_pick;

\echo ''
\echo '» صفوف الملف مقابل الواقع بقاعدة البيانات:'
SELECT
  (SELECT count(*) FROM inv_import)                                       AS "صفوف الملف",
  (SELECT count(*) FROM inv_import i
     JOIN "Booking" b ON b.code = i.booking_code)                         AS "حجزها موجود",
  (SELECT count(*) FROM inv_import i
     LEFT JOIN "Booking" b ON b.code = i.booking_code
    WHERE b.id IS NULL)                                                   AS "حجزها مفقود",
  (SELECT count(*) FROM inv_import i
     JOIN "LeaderInvoice" li ON li."accountingCode" = i.accounting_code)  AS "الرقم المحاسبي موجود مسبقاً",
  (SELECT count(*) FROM inv_import i
     JOIN "Booking" b ON b.code = i.booking_code
     JOIN "LeaderInvoice" li ON li."bookingId" = b.id)                    AS "الحجز عنده فاتورة أصلاً";

\echo ''
\echo '» رقم الفاتورة الخارجي (المعتمدة لازم إلها رقم — قيد leader_invoice_approved_needs_number):'
SELECT
  (SELECT count(*) FROM inv_import WHERE NULLIF(btrim(external_number),'') IS NOT NULL)  AS "عدها رقم ← تنكتب APPROVED",
  (SELECT count(*) FROM inv_import WHERE NULLIF(btrim(external_number),'') IS NULL)      AS "بلا رقم ← تنكتب SUBMITTED",
  (SELECT count(*) FROM inv_import i JOIN "LeaderInvoice" li
      ON lower(btrim(li."externalInvoiceNumber")) = lower(btrim(i.external_number))
   WHERE NULLIF(btrim(i.external_number),'') IS NOT NULL)                                AS "رقمها مستعمل أصلاً ← تنتخطى";

\echo ''
\echo '» أول ١٠ حجوزات مفقودة (إن وجدت):'
SELECT i.booking_code AS "كود الحجز", i.customer_name AS "الزبون", i.accounting_code AS "الرقم المحاسبي"
  FROM inv_import i LEFT JOIN "Booking" b ON b.code = i.booking_code
 WHERE b.id IS NULL LIMIT 10;

-- ═══ الإدخال ═══
-- status='APPROVED' لأن الشغل خالص والمبلغ مستلم — ما بقى شي ينتظر اعتماد.
-- createdAt = تاريخ المشروع بالدفتر، حتى تقارير الإيراد اليومية تحط المبلغ
-- بيومه الحقيقي مو بيوم الاستيراد.
-- pricingMode='MANUAL' لأن ماكو بنود تنفيذ — السعر مكتوب بالإيد أصلاً
-- (نفس منطق الترحيل 0274_leader_invoice_manual_backfill).
CREATE TEMP TABLE inserted_ids AS
WITH ins AS (
INSERT INTO "LeaderInvoice" (
  id, "bookingId", "employeeId", "customerName", "customerPhone", "customerAddress",
  systems, items, "totalDeviceCount", "executionCost", "materialsTotal",
  "discountValue", "netTotal", "accountingCode", status, "pricingMode",
  "manualPriceNote", "approvedByEmployeeId", "approvedAt",
  "externalInvoiceNumber", "externalInvoiceAt", "createdAt"
)
SELECT
  gen_random_uuid()::text,
  b.id,
  o.id,
  NULLIF(btrim(i.customer_name), ''),
  NULLIF(btrim(i.customer_phone), ''),
  NULLIF(btrim(i.customer_address), ''),
  '[]'::jsonb,
  '[]'::jsonb,
  COALESCE(i.device_count, 0),
  COALESCE(i.net_total, 0),
  0,
  0,
  COALESCE(i.net_total, 0),
  btrim(i.accounting_code),
  -- ⚠️ الفاتورة المعتمدة لازم إلها رقم فاتورة خارجي — قيد
  -- leader_invoice_approved_needs_number، وهو ضابط مالي انحط بعد
  -- حادثة كلّفت فاتورة. فالصف الي ما إله رقم بالدفتر يدخل SUBMITTED
  -- مو APPROVED: مبلغه ينحسب ويطلع الحجز من طابور الورق، ونقص الدليل
  -- يبقى **مكشوف** بطابور التدقيق بدل ما ينخفي وراء ختم اعتماد كاذب.
  CASE WHEN NULLIF(btrim(i.external_number), '') IS NOT NULL
       THEN 'APPROVED' ELSE 'SUBMITTED' END,
  'MANUAL',
  'تسوية محاسبية لشغل قبل النظام — المصدر: دفتر تدقيق الحسابات',
  CASE WHEN NULLIF(btrim(i.external_number), '') IS NOT NULL THEN o.id END,
  CASE WHEN NULLIF(btrim(i.external_number), '') IS NOT NULL
       THEN COALESCE(NULLIF(i.invoice_date, '')::timestamptz, now()) END,
  NULLIF(btrim(i.external_number), ''),
  CASE WHEN NULLIF(btrim(i.external_number), '') IS NOT NULL
       THEN COALESCE(NULLIF(i.invoice_date, '')::timestamptz, now()) END,
  COALESCE(NULLIF(i.invoice_date, '')::timestamptz, now())
FROM inv_import i
JOIN "Booking" b   ON b.code = i.booking_code
CROSS JOIN owner_pick o
WHERE NOT EXISTS (
        SELECT 1 FROM "LeaderInvoice" li WHERE li."bookingId" = b.id
      )
  -- ⚠️ ورقم الفاتورة الخارجي فريد (leader_invoice_external_number_unique):
  -- الصف الي رقمه مستعمل أصلاً ينتخطى بهدوء بدل ما يفشل الاستيراد كله.
  AND NOT EXISTS (
        SELECT 1 FROM "LeaderInvoice" li2
        WHERE NULLIF(btrim(i.external_number), '') IS NOT NULL
          AND lower(btrim(li2."externalInvoiceNumber")) = lower(btrim(i.external_number))
      )
ON CONFLICT ("accountingCode") DO NOTHING
RETURNING id, "netTotal", status
)
SELECT * FROM ins;

\echo ''
\echo '───────────── نتيجة الإدخال ─────────────'
\echo ''
SELECT count(*) AS "فواتير انضافت",
       COALESCE(sum("netTotal"),0) AS "مجموع المبالغ",
       count(*) FILTER (WHERE status = 'APPROVED')  AS "معتمدة",
       count(*) FILTER (WHERE status = 'SUBMITTED') AS "بانتظار التدقيق (بلا رقم فاتورة)"
  FROM inserted_ids;

\echo ''
\echo '» أثر التغيير على خانة «منجزة وناقصها ورق»:'
SELECT
  count(*) FILTER (WHERE b.status = 'COMPLETED')                   AS "منجزة (OLD-)",
  count(*) FILTER (WHERE b.status = 'COMPLETED'
                     AND EXISTS (SELECT 1 FROM "LeaderInvoice" li
                                  WHERE li."bookingId" = b.id))    AS "صار عدها فاتورة",
  count(*) FILTER (WHERE b.status = 'COMPLETED'
                     AND NOT EXISTS (SELECT 1 FROM "LeaderInvoice" li
                                      WHERE li."bookingId" = b.id)) AS "باقية بلا فاتورة"
  FROM "Booking" b WHERE b.code LIKE 'OLD-%';

$FINISH
SQL

echo
if [ "$MODE" = "--apply" ]; then
  echo "═══ انحفظ ═══"
else
  echo "═══ تجربة بس — ولا شي انحفظ. للحفظ: $0 $CSV --apply ═══"
fi
echo "==> احذف الملف من السيرفر بعد الانتهاء: rm -f $CSV"
