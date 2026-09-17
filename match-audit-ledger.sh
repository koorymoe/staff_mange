#!/bin/bash
# ══════════════════════════════════════════════════════════════════════════
# فحص بس — مطابقة دفتر تدقيق حسابات (كاميرات/بصمة) مقابل الحجوزات الحقيقية
# ══════════════════════════════════════════════════════════════════════════
#
# ⚠️⚠️ هذا السكربت **قراءة فقط** — ما يكتب ولا حرف بقاعدة البيانات مهما
# صار (يشتغل كله داخل معاملة تنتهي بـROLLBACK دائماً، بلا وضع --apply).
# شغله جمع الأدلة بس: منو من صفوف الدفتر يطابق حجزاً موجوداً، ومنو لا.
#
# ⚠️ الفرق عن import-legacy-invoices.sh: هذا الدفتر تواريخه توصل لقبل
# شهرين بس (يوليو ٢٠٢٦) — يعني أغلب صفوفه على الأغلب حجوزات **حقيقية
# موجودة أصلاً بالنظام بكود عادي** (مو OLD-)، فالمطابقة تصير برقم
# الهاتف + قرب التاريخ، مو بمعادلة كود ثابتة.
#
# التصنيف لكل صف:
#   أ. حجز موجود ومبلغه مدقق أصلاً (amountVerified=true)  → ولا لمسة
#   ب. حجز موجود وغير مدقق                                 → مرشّح للتصحيح
#   ج. ماكو حجز يطابق (هاتف فاضي أو ماكو حجز بنفس الهاتف)   → قرار بشري
#   د. أكثر من حجز يطابق (نفس الهاتف، تواريخ قريبة)         → قرار بشري
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

-- ⚠️ نافذة ±٧ أيام: تاريخ المشروع بالدفتر مو بالضرورة نفس تاريخ
-- الجدولة أو الإنجاز بالنظام (فرق أيام بسبب التنسيق أو التأجيل)،
-- فنافذة ضيقة جداً تفوّت مطابقات حقيقية، وواسعة جداً تخلق مطابقات
-- وهمية. سبعة أيام حل وسط — والمطابقات المتعددة تنكشف بعمود العدد
-- تحت مو تُخفى.
CREATE TEMP TABLE candidates AS
SELECT
  a.ctid AS row_id,
  b.id AS booking_id
FROM audit_import a
JOIN "Customer" c ON c.phone = a.phone
JOIN "Booking" b ON b."customerId" = c.id
WHERE a.phone IS NOT NULL AND btrim(a.phone) <> ''
  AND COALESCE(b."scheduledAt", b."completedAt", b."createdAt")::date
      BETWEEN a.project_date - 7 AND a.project_date + 7;

CREATE TEMP TABLE classified AS
SELECT
  a.*,
  (SELECT count(*) FROM candidates cand WHERE cand.row_id = a.ctid) AS candidate_count,
  (SELECT b.id FROM candidates cand JOIN "Booking" b ON b.id = cand.booking_id
     WHERE cand.row_id = a.ctid LIMIT 1) AS matched_booking_id
FROM audit_import a;

\echo ''
\echo '───────────── التصنيف ─────────────'
\echo ''
\echo '» أ. حجز موجود ومبلغه مدقق أصلاً — ولا لمسة:'
SELECT count(*) AS "العدد"
FROM classified cl JOIN "Booking" b ON b.id = cl.matched_booking_id
WHERE cl.candidate_count = 1 AND b."amountVerified" = true;

\echo ''
\echo '» ب. حجز موجود وغير مدقق — مرشّح للتصحيح (هذا الي راح يتصلح):'
SELECT count(*) AS "العدد", COALESCE(sum(cl.amount_received),0) AS "المبلغ الي راح يزاد للإيرادات"
FROM classified cl JOIN "Booking" b ON b.id = cl.matched_booking_id
WHERE cl.candidate_count = 1 AND b."amountVerified" = false;

\echo ''
\echo '» ج. ماكو حجز يطابق أبداً — يحتاج قرارك:'
SELECT count(*) AS "العدد", COALESCE(sum(cl.amount_received),0) AS "المبلغ اليتيم"
FROM classified cl WHERE cl.candidate_count = 0;

\echo ''
\echo '» تفاصيل الفئة (ج) — أول ٣٠ صف بلا مطابقة (مرتبة بالمبلغ):'
SELECT project_date AS "التاريخ", customer_name AS "الزبون", phone AS "الهاتف",
       accounting_code AS "كود المحاسبة", amount_received AS "المبلغ"
FROM classified WHERE candidate_count = 0
ORDER BY amount_received DESC LIMIT 30;

\echo ''
\echo '» د. أكثر من حجز يطابق — يحتاج قرارك (ما نخمّن):'
SELECT count(*) AS "العدد", COALESCE(sum(amount_received),0) AS "المبلغ"
FROM classified WHERE candidate_count > 1;

\echo ''
\echo '» تفاصيل الفئة (د):'
SELECT project_date AS "التاريخ", customer_name AS "الزبون", phone AS "الهاتف",
       candidate_count AS "عدد الحجوزات المطابقة", amount_received AS "المبلغ"
FROM classified WHERE candidate_count > 1
ORDER BY candidate_count DESC LIMIT 30;

\echo ''
\echo '» صفوف فيها ملاحظة "راجع" مسبقة (كتبتها أثناء التحويل من النص — تدقيق يدوي مطلوب):'
SELECT project_date AS "التاريخ", customer_name AS "الزبون", accounting_code AS "كود المحاسبة", notes AS "الملاحظة"
FROM audit_import WHERE notes ILIKE '%راجع%';

ROLLBACK;
SQL

echo
echo "═══ فحص بس — ما انحفظ ولا شي. هذا تقرير للقراءة حصراً. ═══"
