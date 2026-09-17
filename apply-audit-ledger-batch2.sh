#!/bin/bash
# ══════════════════════════════════════════════════════════════════════════
# تصحيح فعلي — فئة (ب) من دفعة تدقيق الحسابات الثانية (كاميرات/بصمة)
# ══════════════════════════════════════════════════════════════════════════
#
# ⚠️ هذا سكربت **كتابة** (بوضع --apply). بدون --apply يشتغل تجربة بس
# (ROLLBACK دائماً، ولا حرف ينحفظ).
#
# يعتمد **نفس منطق المطابقة بالضبط** من match-audit-ledger.sh (القرائي):
# مطابقة بمستوى الزبون — عدد زيارات كل هاتف بالدفتر (excel_n) مقابل
# عدد حجوزاته المكتملة بالنظام بنفس الفترة (system_n). لو متطابقين،
# نرتب الطرفين بالتاريخ ونطابق وحدة وحدة. لو مختلفين، ما نخمّن — نتركها.
#
# من نتيجة المطابقة، التصحيح يمس فقط:
#   ب. حجز موجود ومطابق بالعدد، وغير مدقق (amountVerified=false)
#
# ولا يمس أبداً:
#   أ. حجز مدقق أصلاً — ولا لمسة
#   ج. زبون بلا حجز مطابق إطلاقاً — تقرير بس، قرار صاحب النظام
#   د. عدد الزيارات مختلف — تقرير بس، ما نخمّن
#   🔴 أي صف عليه ملاحظة "راجع" (أكواد مكررة/مبالغ مشبوهة أثناء
#      التحويل من النص) — يُستثنى من الكتابة حتى لو انطابق بالعدد
#
# الاستخدام:
#   ./apply-audit-ledger-batch2.sh <ملف.csv>            # تجربة (ما يحفظ)
#   ./apply-audit-ledger-batch2.sh <ملف.csv> --apply     # حفظ فعلي
#
# أعمدة الـ CSV (نفس match-audit-ledger.sh):
#   project_date,customer_name,phone,work_type,accounting_code,
#   customer_code,amount_received,invoice_number,leader_name,notes
# ══════════════════════════════════════════════════════════════════════════
set -euo pipefail

CSV="${1:-}"
MODE="${2:-}"
DB_CONTAINER="${DB_CONTAINER:-staff_mange-db-1}"

if [ -z "$CSV" ]; then
  echo "الاستخدام: $0 <دفتر-التدقيق.csv> [--apply]"
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
docker cp "$CSV" "$DB_CONTAINER:/tmp/audit_ledger_batch2.csv"

docker exec -i "$DB_CONTAINER" psql -U staffmange -d staffmange -v ON_ERROR_STOP=1 <<SQL
BEGIN;

CREATE TEMP TABLE audit_import (
  project_date     DATE,
  customer_name    TEXT,
  phone            TEXT,
  work_type        TEXT,
  accounting_code  TEXT,
  customer_code    TEXT,
  amount_received  NUMERIC,
  invoice_number   TEXT,
  leader_name      TEXT,
  notes            TEXT
);
\copy audit_import FROM '/tmp/audit_ledger_batch2.csv' WITH (FORMAT csv, HEADER true)

-- ═══ عدد الزيارات لكل زبون — بالإكسل ═══
CREATE TEMP TABLE excel_counts AS
SELECT phone, count(*) AS excel_n, min(project_date) AS d_min, max(project_date) AS d_max
FROM audit_import
WHERE phone IS NOT NULL AND btrim(phone) <> ''
GROUP BY phone;

-- ═══ عدد الحجوزات المكتملة لنفس الزبون بنفس الفترة — بالنظام ═══
CREATE TEMP TABLE system_counts AS
SELECT c.phone, count(*) AS system_n
FROM excel_counts ec
JOIN "Customer" c ON c.phone = ec.phone
JOIN "Booking" b ON b."customerId" = c.id
WHERE b.status = 'COMPLETED'
  AND COALESCE(b."scheduledAt", b."completedAt", b."createdAt")::date
      BETWEEN ec.d_min - 7 AND ec.d_max + 7
GROUP BY c.phone;

-- ═══ المطابقة الفعلية: ترتيب بالتاريخ لكل طرف، ثم زوج بزوج ═══
CREATE TEMP TABLE excel_ranked AS
SELECT a.*, row_number() OVER (PARTITION BY a.phone ORDER BY a.project_date) AS rn
FROM audit_import a
JOIN excel_counts ec ON ec.phone = a.phone
JOIN system_counts sc ON sc.phone = ec.phone AND sc.system_n = ec.excel_n;

CREATE TEMP TABLE booking_ranked AS
SELECT b.id, b."amountVerified", b."amountCollected", c.phone,
       row_number() OVER (PARTITION BY c.phone
         ORDER BY COALESCE(b."scheduledAt", b."completedAt", b."createdAt")) AS rn
FROM "Booking" b
JOIN "Customer" c ON c.id = b."customerId"
JOIN excel_counts ec ON ec.phone = c.phone
JOIN system_counts sc ON sc.phone = ec.phone AND sc.system_n = ec.excel_n
WHERE b.status = 'COMPLETED'
  AND COALESCE(b."scheduledAt", b."completedAt", b."createdAt")::date
      BETWEEN ec.d_min - 7 AND ec.d_max + 7;

CREATE TEMP TABLE paired AS
SELECT e.*, br.id AS matched_booking_id, br."amountVerified" AS matched_verified,
       br."amountCollected" AS matched_amount
FROM excel_ranked e
JOIN booking_ranked br ON br.phone = e.phone AND br.rn = e.rn;

\echo ''
\echo '───────────── قبل التصحيح ─────────────'
\echo ''
\echo '» فئة (ب) — حجز موجود وغير مدقق، ومؤهّل للكتابة (بلا ملاحظة "راجع"):'
SELECT count(*) AS "العدد", COALESCE(sum(amount_received),0) AS "المبلغ الي راح يزاد للإيرادات"
FROM paired
WHERE matched_verified = false
  AND (notes IS NULL OR notes NOT ILIKE '%راجع%');

\echo ''
\echo '» صفوف مستثناة من الكتابة رغم مطابقتها (ملاحظة "راجع" — مراجعة يدوية):'
SELECT project_date AS "التاريخ", customer_name AS "الزبون", accounting_code AS "كود المحاسبة",
       amount_received AS "المبلغ", notes AS "الملاحظة"
FROM paired
WHERE matched_verified = false AND notes ILIKE '%راجع%';

-- ═══ التصحيح — فئة (ب) فقط، بلا صفوف "راجع" ═══
-- ⚠️ RETURNING يمسك بالضبط الصفوف الي هذا التشغيل لمسها.
CREATE TEMP TABLE fixed_rows AS
WITH upd AS (
  UPDATE "Booking" b
  SET "amountCollected" = p.amount_received,
      "amountVerified"  = true
  FROM paired p
  WHERE b.id = p.matched_booking_id
    AND p.matched_verified = false
    AND (p.notes IS NULL OR p.notes NOT ILIKE '%راجع%')
  RETURNING b.id, b."amountCollected"
)
SELECT * FROM upd;

\echo ''
\echo '───────────── بعد التصحيح ─────────────'
\echo ''
\echo '» شكد حجز صار مدقق تلقائياً هالتشغيلة، وشكد زاد إجمالي الإيرادات:'
SELECT count(*) AS "حجوزات انصلحت الآن", COALESCE(sum("amountCollected"), 0) AS "زيادة إجمالي الإيرادات"
  FROM fixed_rows;

\echo ''
\echo '» للعلم بس — فئة (ج): زبون بلا حجوزات مكتملة بالنظام إطلاقاً بهذي الفترة (بلا كتابة):'
SELECT count(*) AS "العدد", COALESCE(sum(a.amount_received),0) AS "المبلغ اليتيم"
FROM audit_import a
JOIN excel_counts ec ON ec.phone = a.phone
LEFT JOIN system_counts sc ON sc.phone = ec.phone
WHERE sc.phone IS NULL;

\echo ''
\echo '» للعلم بس — عدد الزيارات مختلف بين الدفتر والنظام (بلا كتابة، مراجعة يدوية):'
SELECT ec.phone AS "الهاتف", ec.excel_n AS "زياراته بالدفتر",
       COALESCE(sc.system_n, 0) AS "حجوزاته المكتملة بالنظام"
FROM excel_counts ec LEFT JOIN system_counts sc ON sc.phone = ec.phone
WHERE ec.excel_n <> COALESCE(sc.system_n, 0)
ORDER BY ec.excel_n DESC;

$FINISH
SQL

echo
if [ "$MODE" = "--apply" ]; then
  echo "═══ انحفظ ═══"
else
  echo "═══ تجربة بس — ولا شي انحفظ. للحفظ: $0 $CSV --apply ═══"
fi
