#!/bin/bash
# ══════════════════════════════════════════════════════════════════════════
# ثلاث قوائم تفصيلية — بعد تشغيل apply-audit-ledger-batch2.sh --apply
# ══════════════════════════════════════════════════════════════════════════
#
# ⚠️ قراءة فقط — ما يكتب ولا حرف (ROLLBACK دائماً).
#
# يطبع ثلاث قوائم بالاسم وكود المحاسبة والمبلغ والتاريخ:
#   ١. مدققة أصلاً — كانت عندها فاتورة قبل ما نلمسها اليوم إطلاقاً
#   ٢. انطابقت وصحّحناها اليوم — إحنا الي ضفنا إلها الفاتورة
#   ٣. رقم فاتورة مكرر — انطابقت بس ما انضافت إلها فاتورة (تحتاج مراجعتك)
#
# الاستخدام: ./list-audit-ledger-batch2.sh <ملف.csv>
# ══════════════════════════════════════════════════════════════════════════
set -euo pipefail

CSV="${1:-}"
DB_CONTAINER="${DB_CONTAINER:-staff_mange-db-1}"

if [ -z "$CSV" ]; then
  echo "الاستخدام: $0 <دفتر-التدقيق.csv>"
  exit 1
fi
[ -f "$CSV" ] || { echo "==> الملف غير موجود: $CSV"; exit 1; }

OUR_NOTE='تسوية محاسبية — دفعة ثانية من دفتر تدقيق الحسابات (مطابقة مرخّاة)'

echo "==> نسخ الملف داخل حاوية قاعدة البيانات..."
docker cp "$CSV" "$DB_CONTAINER:/tmp/audit_ledger_batch2.csv"

docker exec -i "$DB_CONTAINER" psql -U staffmange -d staffmange -v ON_ERROR_STOP=1 <<SQL
BEGIN;

CREATE TEMP TABLE audit_import (
  project_date     DATE,
  customer_name    TEXT,
  phone            TEXT,
  phone2           TEXT,
  work_type        TEXT,
  accounting_code  TEXT,
  customer_code    TEXT,
  amount_received  NUMERIC,
  invoice_number   TEXT,
  leader_name      TEXT,
  notes            TEXT
);
\copy audit_import FROM '/tmp/audit_ledger_batch2.csv' WITH (FORMAT csv, HEADER true)

CREATE TEMP TABLE unique_names AS
  SELECT name, min(id) AS id FROM "Customer" GROUP BY name HAVING count(*) = 1;

CREATE TEMP TABLE resolved AS
SELECT a.*, COALESCE(c1.id, c2.id, cn.id) AS customer_id
FROM audit_import a
LEFT JOIN "Customer" c1 ON c1.phone = a.phone AND NULLIF(btrim(a.phone), '') IS NOT NULL
LEFT JOIN "Customer" c2 ON c2.phone = a.phone2 AND NULLIF(btrim(a.phone2), '') IS NOT NULL AND c1.id IS NULL
LEFT JOIN unique_names un ON un.name = a.customer_name AND c1.id IS NULL AND c2.id IS NULL
LEFT JOIN "Customer" cn ON cn.id = un.id;

CREATE TEMP TABLE excel_counts AS
SELECT customer_id, count(*) AS excel_n, min(project_date) AS d_min, max(project_date) AS d_max
FROM resolved WHERE customer_id IS NOT NULL GROUP BY customer_id;

CREATE TEMP TABLE system_counts AS
SELECT ec.customer_id, count(*) AS system_n
FROM excel_counts ec
JOIN "Booking" b ON b."customerId" = ec.customer_id
WHERE b.status = 'COMPLETED'
  AND COALESCE(b."scheduledAt", b."completedAt", b."createdAt")::date
      BETWEEN ec.d_min - 7 AND ec.d_max + 7
GROUP BY ec.customer_id;

CREATE TEMP TABLE excel_ranked AS
SELECT r.*, ec.excel_n, COALESCE(sc.system_n, 0) AS system_n,
       row_number() OVER (PARTITION BY r.customer_id ORDER BY r.project_date) AS rn
FROM resolved r
JOIN excel_counts ec ON ec.customer_id = r.customer_id
LEFT JOIN system_counts sc ON sc.customer_id = r.customer_id;

CREATE TEMP TABLE booking_ranked AS
SELECT b.id, b."amountVerified", b."customerId",
       row_number() OVER (PARTITION BY b."customerId"
         ORDER BY COALESCE(b."scheduledAt", b."completedAt", b."createdAt")) AS rn
FROM "Booking" b
JOIN excel_counts ec ON ec.customer_id = b."customerId"
WHERE b.status = 'COMPLETED'
  AND COALESCE(b."scheduledAt", b."completedAt", b."createdAt")::date
      BETWEEN ec.d_min - 7 AND ec.d_max + 7;

CREATE TEMP TABLE paired AS
SELECT e.*, br.id AS matched_booking_id
FROM excel_ranked e
JOIN booking_ranked br ON br."customerId" = e.customer_id AND br.rn = e.rn
WHERE e.rn <= LEAST(e.excel_n, e.system_n);

\echo ''
\echo '════════ القائمة ١ — مدققة أصلاً (فاتورتها موجودة من قبل، مو منّا) ════════'
SELECT p.project_date AS "التاريخ", p.customer_name AS "الزبون", p.accounting_code AS "كود المحاسبة",
       p.amount_received AS "المبلغ"
FROM paired p
JOIN "LeaderInvoice" li ON li."bookingId" = p.matched_booking_id
WHERE li."manualPriceNote" IS DISTINCT FROM '$OUR_NOTE'
ORDER BY p.project_date;

\echo ''
\echo '════════ القائمة ٢ — انطابقت وصحّحناها اليوم (إحنا ضفنا فاتورتها) ════════'
SELECT p.project_date AS "التاريخ", p.customer_name AS "الزبون", p.accounting_code AS "كود المحاسبة",
       p.amount_received AS "المبلغ", li."externalInvoiceNumber" AS "رقم الفاتورة الخارجي"
FROM paired p
JOIN "LeaderInvoice" li ON li."bookingId" = p.matched_booking_id
WHERE li."manualPriceNote" = '$OUR_NOTE'
ORDER BY p.project_date;

\echo ''
\echo '════════ القائمة ٣ — انطابقت بس ماكو إلها فاتورة (رقم فاتورة مكرر — تحتاج مراجعتك) ════════'
SELECT p.project_date AS "التاريخ", p.customer_name AS "الزبون", p.accounting_code AS "كود المحاسبة",
       p.amount_received AS "المبلغ", p.invoice_number AS "رقم الفاتورة بالدفتر"
FROM paired p
WHERE NOT EXISTS (SELECT 1 FROM "LeaderInvoice" li WHERE li."bookingId" = p.matched_booking_id)
ORDER BY p.project_date;

ROLLBACK;
SQL

echo
echo "═══ قراءة فقط — ولا شي انحفظ أو اتغيّر ═══"
