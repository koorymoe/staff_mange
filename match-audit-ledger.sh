#!/bin/bash
# ══════════════════════════════════════════════════════════════════════════
# فحص بس — مطابقة دفتر تدقيق حسابات (كاميرات/بصمة) مقابل الحجوزات الحقيقية
# ══════════════════════════════════════════════════════════════════════════
#
# ⚠️⚠️ هذا السكربت **قراءة فقط** — ما يكتب ولا حرف بقاعدة البيانات مهما
# صار (يشتغل كله داخل معاملة تنتهي بـROLLBACK دائماً، بلا وضع --apply).
#
# ⚠️ نسخة ثانية — طلب صاحب العمل: «أني جاي أدقق الحجوزات مو الزبائن.
# الزيارة إذا تكررت، دوّر بالنظام هل متكررة نفس العدد. إذا متكررة
# بنفس العدد، طابقها بالترتيب من الإكسل. وإذا مو متكررة (عدد مختلف)
# خلّي الموجود — يعني ما تخمّن، اعرضها للمراجعة».
#
# يعني المطابقة **بمستوى الزبون** مو الصف: لكل هاتف، نقارن عدد زياراته
# بالدفتر (excel_n) مقابل عدد حجوزاته المكتملة بالنظام بنفس الفترة
# الزمنية (system_n):
#
#   excel_n = system_n  → نرتب الطرفين بالتاريخ ونطابق وحدة وحدة
#   excel_n ≠ system_n  → ما نخمّن، نعرض التفصيل للمراجعة اليدوية
#
# وبعد المطابقة، كل زوج (صف إكسل ↔ حجز) ينقسم لنفس فئتين سابقتين:
#   حجز مدقق أصلاً (ولا لمسة) · حجز غير مدقق (مرشّح للتصحيح)
#
# الاستخدام:
#   ./match-audit-ledger.sh <ملف.csv>
#
# أعمدة الـ CSV (UTF-8، سطر عناوين أول):
#   project_date,customer_name,phone,work_type,accounting_code,
#   customer_code,amount_received,invoice_number,leader_name,notes
# ══════════════════════════════════════════════════════════════════════════
set -euo pipefail

CSV="${1:-}"
DB_CONTAINER="${DB_CONTAINER:-staff_mange-db-1}"

if [ -z "$CSV" ]; then
  echo "الاستخدام: $0 <دفتر-التدقيق.csv>"
  exit 1
fi
[ -f "$CSV" ] || { echo "==> الملف غير موجود: $CSV"; exit 1; }

echo "═══ فحص بس — ما يكتب ولا حرف (ROLLBACK دائماً) ═══"
echo

echo "==> نسخ الملف داخل حاوية قاعدة البيانات..."
docker cp "$CSV" "$DB_CONTAINER:/tmp/audit_ledger.csv"

docker exec -i "$DB_CONTAINER" psql -U staffmange -d staffmange -v ON_ERROR_STOP=1 <<'SQL'
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
\copy audit_import FROM '/tmp/audit_ledger.csv' WITH (FORMAT csv, HEADER true)

\echo ''
\echo '───────────── صفوف الملف ─────────────'
SELECT count(*) AS "صفوف الملف", COALESCE(sum(amount_received),0) AS "مجموع المبلغ المستلم"
  FROM audit_import;

\echo ''
\echo '» صفوف بلا رقم هاتف صالح (ما تنطابق أبداً):'
SELECT count(*) AS "بلا هاتف" FROM audit_import
 WHERE phone IS NULL OR btrim(phone) = '';

-- ═══ عدد الزيارات لكل زبون — بالإكسل ═══
CREATE TEMP TABLE excel_counts AS
SELECT phone, count(*) AS excel_n, min(project_date) AS d_min, max(project_date) AS d_max
FROM audit_import
WHERE phone IS NOT NULL AND btrim(phone) <> ''
GROUP BY phone;

-- ═══ عدد الحجوزات المكتملة لنفس الزبون بنفس الفترة — بالنظام ═══
-- ⚠️ الفترة = مدى تواريخ زياراته بالدفتر ± أسبوع، مو كل تاريخه —
-- زبون قديم عنده عشرات الحجوزات عبر السنين، وحصر الفترة يمنع خلط
-- حجوزات برّا نطاق هذا الدفتر أصلاً.
CREATE TEMP TABLE system_counts AS
SELECT c.phone, count(*) AS system_n
FROM excel_counts ec
JOIN "Customer" c ON c.phone = ec.phone
JOIN "Booking" b ON b."customerId" = c.id
WHERE b.status = 'COMPLETED'
  AND COALESCE(b."scheduledAt", b."completedAt", b."createdAt")::date
      BETWEEN ec.d_min - 7 AND ec.d_max + 7
GROUP BY c.phone;

\echo ''
\echo '───────────── مطابقة العدد لكل زبون ─────────────'
\echo ''
\echo '» زبائن عدد زياراتهم بالدفتر = عدد حجوزاتهم المكتملة بالنظام (يتطابقون تلقائياً):'
SELECT count(DISTINCT ec.phone) AS "عدد الزبائن", sum(ec.excel_n) AS "عدد الصفوف"
FROM excel_counts ec JOIN system_counts sc ON sc.phone = ec.phone AND sc.system_n = ec.excel_n;

\echo ''
\echo '» زبائن العدد عندهم مختلف (يحتاج مراجعتك — ما نخمّن):'
SELECT ec.phone AS "الهاتف", ec.excel_n AS "زياراته بالدفتر",
       COALESCE(sc.system_n, 0) AS "حجوزاته المكتملة بالنظام"
FROM excel_counts ec LEFT JOIN system_counts sc ON sc.phone = ec.phone
WHERE ec.excel_n <> COALESCE(sc.system_n, 0)
ORDER BY ec.excel_n DESC;

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
\echo '───────────── التصنيف النهائي (بعد المطابقة بالترتيب) ─────────────'
\echo ''
\echo '» أ. حجز موجود ومبلغه مدقق أصلاً — ولا لمسة:'
SELECT count(*) AS "العدد" FROM paired WHERE matched_verified = true;

\echo ''
\echo '» ب. حجز موجود وغير مدقق — مرشّح للتصحيح:'
SELECT count(*) AS "العدد", COALESCE(sum(amount_received),0) AS "المبلغ الي راح يزاد للإيرادات"
FROM paired WHERE matched_verified = false;

\echo ''
\echo '» ج. زبون بلا حجوزات مكتملة بالنظام إطلاقاً بهذي الفترة (excel_n موجود، system_n=0):'
SELECT count(*) AS "العدد", COALESCE(sum(a.amount_received),0) AS "المبلغ اليتيم"
FROM audit_import a
JOIN excel_counts ec ON ec.phone = a.phone
LEFT JOIN system_counts sc ON sc.phone = ec.phone
WHERE sc.phone IS NULL;

\echo ''
\echo '» صفوف فيها ملاحظة "راجع" مسبقة (كتبتها أثناء التحويل من النص — تدقيق يدوي مطلوب، تُستثنى من أي تصحيح تلقائي):'
SELECT project_date AS "التاريخ", customer_name AS "الزبون", accounting_code AS "كود المحاسبة", notes AS "الملاحظة"
FROM audit_import WHERE notes ILIKE '%راجع%';

ROLLBACK;
SQL

echo
echo "═══ فحص بس — ما انحفظ ولا شي. هذا تقرير للقراءة حصراً. ═══"
