#!/bin/bash
# ══════════════════════════════════════════════════════════════════════════
# تصحيح فعلي — الدفعة الثانية من دفتر تدقيق الحسابات (نسخة ٢)
# ══════════════════════════════════════════════════════════════════════════
#
# ⚠️ هذا سكربت **كتابة** (بوضع --apply). بدون --apply يشتغل تجربة بس
# (ROLLBACK دائماً، ولا حرف ينحفظ).
#
# 🔴 الفرق عن النسخة الأولى (كومت df9244e) — بطلب صريح من صاحب النظام:
#
# ١. **المطابقة تُرخى ثلاث خطوات** (بدل التطابق الصارم بالهاتف والعدد):
#    - الهاتف: نجرّب `phone` وإذا ما طابق `phone2` (خلية فيها رقمين)
#    - إذا ما طابق أي هاتف: نطابق بالاسم — **بس لو الاسم يحدد زبون
#      واحد بالضبط بالنظام** (تجنّباً لخلط زبونين بنفس الاسم)
#    - عدد الزيارات: بدل اشتراط تطابق العدد بالضبط، نرتّب الطرفين
#      بالتاريخ ونطابق أقل عدد مشترك (`LEAST(excel_n, system_n)`) —
#      الزائد (لو بالدفتر زيارات أكثر من حجوزات النظام) يبقى للمراجعة
#      اليدوية بس ما يوقف مطابقة الباقي
#    - صفوف "FALSE" بعمود التدقيق **تدخل المطابقة أيضاً** (صاحب النظام
#      أكّد: العمود هذا ماله معنى عندهم، كل المبالغ حقيقية ومؤكدة)
#
# ٢. 🔴 **تنشئ فاتورة ليدر (`LeaderInvoice`) لو مو موجودة** — هذا كان
#    الفجوة الحقيقية. تصحيح `amountVerified` بس ما يطلّع الحجز من
#    خانة «منجزة وناقصها ورق» (`PendingPaperworkForEmployee` بالكود) —
#    تلك الخانة تتأكد من **وجود فاتورة فعلية**، مو من علم `amountVerified`.
#
#    ⚠️ **تصحيح مهم**: أول نسخة حاولت تحدد "الليدر الحقيقي" لكل حجز
#    (كادر الحجز، وإلا اسم عمود "اسم المجهز" بالدفتر). طلعت غلط —
#    "اسم المجهز" هو **المحاسب** الي سجّل الدفتر (مرتضى عباس بـ٩٥٪ من
#    الصفوف)، مو الفني الي سوى الشغل بالميدان. وكادر الحجز نفسه ماكو
#    إله تسجيل لأغلب هذي الحجوزات (شغل صار خارج تتبع التطبيق). صاحب
#    النظام صحّحني وطلب صراحة: «ماريد ليدر اريد فقط الفاتورة».
#
#    فالفاتورة تُنشأ باسم **المالك** (`OWNER`) — **نفس حل الدفعة الأولى**
#    بالضبط (`import-legacy-invoices.sh`): تسوية محاسبية لشغل ما نگدر
#    نوثّق منو نفّذه بالضبط، وحطّها باسم موظف تقني يزوّر إحصائيات أدائه
#    وأرباحه.
#
# ⚠️ وبصراحة لازم تعرفها: خانة «منجزة وناقصها ورق» تشترط **فاتورة
# وتقرير عمل معاً** (`WorkReport`) لغير حجوزات `OLD-`. هذا السكربت
# يضمن الفاتورة بس — إذا الحجز أصلاً بلا تقرير عمل مسجّل بالنظام
# (نادر لحجز مكتمل عادي، بس ممكن)، يبقى ظاهراً بالخانة لحد ما يُسوّى
# له تقرير من طرف ثاني. السكربت **يطبع** عدد هذي الحالات صراحة —
# ما يخفيها.
#
# الاستخدام:
#   ./apply-audit-ledger-batch2.sh <ملف.csv>            # تجربة (ما يحفظ)
#   ./apply-audit-ledger-batch2.sh <ملف.csv> --apply     # حفظ فعلي
#
# أعمدة الـ CSV:
#   project_date,customer_name,phone,phone2,work_type,accounting_code,
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
  phone2           TEXT,
  location         TEXT,
  work_type        TEXT,
  accounting_code  TEXT,
  customer_code    TEXT,
  amount_received  NUMERIC,
  invoice_number   TEXT,
  leader_name      TEXT,
  notes            TEXT
);
\copy audit_import FROM '/tmp/audit_ledger_batch2.csv' WITH (FORMAT csv, HEADER true)

\echo ''
\echo '───────────── صفوف الملف ─────────────'
SELECT count(*) AS "صفوف الملف", COALESCE(sum(amount_received),0) AS "مجموع المبلغ المستلم"
  FROM audit_import;

-- ═══ الاسم يحدد زبون واحد بالضبط — تجنّباً لخلط زبونين بنفس الاسم ═══
CREATE TEMP TABLE unique_names AS
  SELECT name, min(id) AS id FROM "Customer" GROUP BY name HAVING count(*) = 1;

-- ═══ حل هوية الزبون: هاتف١ → هاتف٢ → اسم فريد ═══
CREATE TEMP TABLE resolved AS
SELECT a.*,
       COALESCE(c1.id, c2.id, cn.id)               AS customer_id,
       (CASE WHEN c1.id IS NOT NULL THEN 'phone'
             WHEN c2.id IS NOT NULL THEN 'phone2'
             WHEN cn.id IS NOT NULL THEN 'name'
             ELSE NULL END)                        AS matched_by
FROM audit_import a
LEFT JOIN "Customer" c1 ON c1.phone = a.phone AND NULLIF(btrim(a.phone), '') IS NOT NULL
LEFT JOIN "Customer" c2 ON c2.phone = a.phone2 AND NULLIF(btrim(a.phone2), '') IS NOT NULL AND c1.id IS NULL
LEFT JOIN unique_names un ON un.name = a.customer_name AND c1.id IS NULL AND c2.id IS NULL
LEFT JOIN "Customer" cn ON cn.id = un.id;

\echo ''
\echo '» طريقة تحديد الهوية (هاتف/هاتف بديل/اسم فريد/ما تحدد):'
SELECT COALESCE(matched_by, 'بلا تحديد') AS "الطريقة", count(*) AS "العدد"
  FROM resolved GROUP BY matched_by ORDER BY 2 DESC;

-- ═══ عدد زيارات كل زبون بالدفتر ═══
CREATE TEMP TABLE excel_counts AS
SELECT customer_id, count(*) AS excel_n, min(project_date) AS d_min, max(project_date) AS d_max
FROM resolved
WHERE customer_id IS NOT NULL
GROUP BY customer_id;

-- ═══ عدد حجوزاته المكتملة بنفس الفترة (± أسبوع) بالنظام ═══
CREATE TEMP TABLE system_counts AS
SELECT ec.customer_id, count(*) AS system_n
FROM excel_counts ec
JOIN "Booking" b ON b."customerId" = ec.customer_id
WHERE b.status = 'COMPLETED'
  AND COALESCE(b."scheduledAt", b."completedAt", b."createdAt")::date
      BETWEEN ec.d_min - 7 AND ec.d_max + 7
GROUP BY ec.customer_id;

-- ═══ المطابقة المرخّاة: ترتيب بالتاريخ، وربط أقل عدد مشترك ═══
CREATE TEMP TABLE excel_ranked AS
SELECT r.*, ec.excel_n, COALESCE(sc.system_n, 0) AS system_n,
       row_number() OVER (PARTITION BY r.customer_id ORDER BY r.project_date) AS rn
FROM resolved r
JOIN excel_counts ec ON ec.customer_id = r.customer_id
LEFT JOIN system_counts sc ON sc.customer_id = r.customer_id;

CREATE TEMP TABLE booking_ranked AS
SELECT b.id, b."amountVerified", b."amountCollected", b."customerId",
       row_number() OVER (PARTITION BY b."customerId"
         ORDER BY COALESCE(b."scheduledAt", b."completedAt", b."createdAt")) AS rn
FROM "Booking" b
JOIN excel_counts ec ON ec.customer_id = b."customerId"
WHERE b.status = 'COMPLETED'
  AND COALESCE(b."scheduledAt", b."completedAt", b."createdAt")::date
      BETWEEN ec.d_min - 7 AND ec.d_max + 7;

CREATE TEMP TABLE paired AS
SELECT e.*, br.id AS matched_booking_id, br."amountVerified" AS matched_verified
FROM excel_ranked e
JOIN booking_ranked br ON br."customerId" = e.customer_id AND br.rn = e.rn
WHERE e.rn <= LEAST(e.excel_n, e.system_n);

\echo ''
\echo '───────────── التصنيف قبل التصحيح ─────────────'
\echo ''
\echo '» بلا تحديد هوية إطلاقاً (لا هاتف ولا اسم فريد) — مراجعة يدوية:'
SELECT count(*) AS "العدد", COALESCE(sum(amount_received),0) AS "المبلغ"
FROM resolved WHERE customer_id IS NULL;

\echo ''
\echo '» زبون محدد الهوية بس بلا أي حجز مكتمل بالنظام بهذي الفترة (يتيم):'
SELECT count(*) AS "العدد", COALESCE(sum(amount_received),0) AS "المبلغ"
FROM excel_ranked WHERE system_n = 0;

\echo ''
\echo '» صفوف زايدة عن عدد الحجوزات المتوفرة (بالدفتر زيارات أكثر) — تبقى لمراجعة لاحقة:'
SELECT count(*) AS "العدد", COALESCE(sum(amount_received),0) AS "المبلغ"
FROM excel_ranked WHERE system_n > 0 AND rn > system_n;

\echo ''
\echo '» مطابق فعلياً (سيُصحَّح أو موجود مسبقاً):'
SELECT count(*) AS "العدد", COALESCE(sum(amount_received),0) AS "المبلغ"
FROM paired;

\echo ''
\echo '» منها: مدقق ومفوتر أصلاً — ولا لمسة:'
SELECT count(*) AS "العدد"
FROM paired p
WHERE p.matched_verified = true
  AND EXISTS (SELECT 1 FROM "LeaderInvoice" li WHERE li."bookingId" = p.matched_booking_id);

\echo ''
\echo '» منها: مرشّح للتصحيح (تحديث المبلغ و/أو إنشاء فاتورة):'
SELECT count(*) AS "العدد", COALESCE(sum(amount_received),0) AS "المبلغ"
FROM paired p
WHERE NOT (p.matched_verified = true
           AND EXISTS (SELECT 1 FROM "LeaderInvoice" li WHERE li."bookingId" = p.matched_booking_id));

-- ═══ المالك: صاحب الفواتير الي ما نگدر نحدد ليدرها الحقيقي ═══
-- نفس نمط import-legacy-invoices.sh بالضبط — لو ماكو حساب مالك واحد
-- بالضبط نوقف، لأن الفواتير لازم تنكتب على حساب معروف.
CREATE TEMP TABLE owner_pick AS
  SELECT id, name FROM "Employee" WHERE role = 'OWNER' ORDER BY "createdAt" LIMIT 1;

DO \$\$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM owner_pick) THEN
    RAISE EXCEPTION 'ماكو حساب بصلاحية OWNER — ما أگدر أحدد صاحب الفواتير';
  END IF;
END \$\$;

\echo ''
\echo '» حساب المالك الي راح تنكتب الفواتير باسمه:'
SELECT name AS "الاسم", id AS "المعرّف" FROM owner_pick;

-- ═══ رقم فاتورة داخلي مكرر بين صفين بنفس الدفعة — ناخذ أولهم بالتاريخ ═══
-- (فريد بمستوى القاعدة أصلاً `leader_invoice_external_number_unique`؛
-- بدون هذا التصنيف، صفّين بنفس الرقم بنفس دفعة الـINSERT يفشّلوا
-- بعضهم داخل نفس الاستعلام، مو بس ضد فواتير موجودة سابقاً).
CREATE TEMP TABLE invoice_number_ranked AS
SELECT p.*,
       CASE WHEN NULLIF(btrim(p.invoice_number), '') IS NOT NULL THEN
         row_number() OVER (PARTITION BY lower(btrim(p.invoice_number)) ORDER BY p.project_date)
       ELSE 1 END AS inv_rn
FROM paired p;

-- ═══ التصحيح: إنشاء الفاتورة الناقصة باسم المالك ═══
CREATE TEMP TABLE created_invoices AS
WITH ins AS (
INSERT INTO "LeaderInvoice" (
  id, "bookingId", "employeeId", "customerName", "customerPhone",
  systems, items, "totalDeviceCount", "executionCost", "materialsTotal",
  "discountValue", "netTotal", "accountingCode", status, "pricingMode",
  "manualPriceNote", "approvedByEmployeeId", "approvedAt",
  "externalInvoiceNumber", "externalInvoiceAt", "createdAt"
)
SELECT
  gen_random_uuid()::text,
  p.matched_booking_id,
  o.id,
  NULLIF(btrim(p.customer_name), ''),
  NULLIF(btrim(p.phone), ''),
  '[]'::jsonb, '[]'::jsonb,
  0, COALESCE(p.amount_received, 0), 0, 0, COALESCE(p.amount_received, 0),
  btrim(p.accounting_code),
  CASE WHEN NULLIF(btrim(p.invoice_number), '') IS NOT NULL
       THEN 'APPROVED' ELSE 'SUBMITTED' END,
  'MANUAL',
  'تسوية محاسبية — دفعة ثانية من دفتر تدقيق الحسابات (مطابقة مرخّاة)',
  CASE WHEN NULLIF(btrim(p.invoice_number), '') IS NOT NULL THEN o.id END,
  CASE WHEN NULLIF(btrim(p.invoice_number), '') IS NOT NULL
       THEN p.project_date::timestamptz ELSE NULL END,
  NULLIF(btrim(p.invoice_number), ''),
  CASE WHEN NULLIF(btrim(p.invoice_number), '') IS NOT NULL
       THEN p.project_date::timestamptz ELSE NULL END,
  p.project_date::timestamptz
FROM invoice_number_ranked p
CROSS JOIN owner_pick o
WHERE NOT (p.matched_verified = true
           AND EXISTS (SELECT 1 FROM "LeaderInvoice" li WHERE li."bookingId" = p.matched_booking_id))
  AND NOT EXISTS (SELECT 1 FROM "LeaderInvoice" li WHERE li."bookingId" = p.matched_booking_id)
  AND p.inv_rn = 1
  -- ⚠️ رقم الفاتورة الخارجي فريد (leader_invoice_external_number_unique)
  -- بمستوى القاعدة نفسها — بلا هذا الشرط، رقم مستعمل أصلاً بفاتورة
  -- موجودة سابقاً يفشّل الـINSERT بأكمله ويطيح المعاملة كلها. نتخطاه
  -- بهدوء بدل ما نكسر باقي التصحيح.
  AND NOT EXISTS (
        SELECT 1 FROM "LeaderInvoice" li2
        WHERE NULLIF(btrim(p.invoice_number), '') IS NOT NULL
          AND lower(btrim(li2."externalInvoiceNumber")) = lower(btrim(p.invoice_number))
      )
ON CONFLICT ("accountingCode") DO NOTHING
RETURNING id, "bookingId", "netTotal"
)
SELECT * FROM ins;

-- ═══ مزامنة الحجز — نفس ضمان السكربتات السابقة: ما نلمس مبلغاً حقيقياً مدقّق يدوياً ═══
--
-- 🔴 شرط `EXISTS(LeaderInvoice)` هنا **إجباري**: هذا الاستعلام يشتغل
-- بعد الـINSERT فوق مباشرة بنفس المعاملة، فيشوف الفواتير المضافة
-- توّها. صف استُثني من الإنشاء (رقم فاتورة مكرر بالدفعة، مثلاً) ما
-- يصير يتأشّر "مدقق" بلا فاتورة وراه — نفس الغلط الي انصلح قبل مع
-- حالة "بلا ليدر".
CREATE TEMP TABLE fixed_rows AS
WITH upd AS (
  UPDATE "Booking" b
  SET "amountCollected" = p.amount_received,
      "amountVerified"  = true
  FROM paired p
  WHERE b.id = p.matched_booking_id
    AND NOT (p.matched_verified = true
             AND EXISTS (SELECT 1 FROM "LeaderInvoice" li WHERE li."bookingId" = p.matched_booking_id))
    AND EXISTS (SELECT 1 FROM "LeaderInvoice" li WHERE li."bookingId" = p.matched_booking_id)
    AND (b."amountCollected" IS NULL OR b."amountCollected" = 0)
    AND b."amountVerified" = false
  RETURNING b.id, b."amountCollected"
)
SELECT * FROM upd;

\echo ''
\echo '───────────── نتيجة التصحيح ─────────────'
\echo ''
\echo '» فواتير ليدر انضافت الآن:'
SELECT count(*) AS "العدد", COALESCE(sum("netTotal"),0) AS "المبلغ" FROM created_invoices;

\echo ''
\echo '» حجوزات صار مبلغها مدقق تلقائياً الآن (وزيادة إجمالي الإيرادات):'
SELECT count(*) AS "العدد", COALESCE(sum("amountCollected"),0) AS "الزيادة" FROM fixed_rows;

\echo ''
\echo '» ⚠️ من هذي الحجوزات المصحَّحة — كم واحد **بعده بلا تقرير عمل** (يبقى بخانة «منجزة وناقصها ورق» رغم الفاتورة):'
SELECT count(*) AS "العدد"
FROM fixed_rows fr
JOIN "Booking" b ON b.id = fr.id
WHERE NOT EXISTS (SELECT 1 FROM "WorkReport" wr WHERE wr."bookingId" = b.id)
  AND b.code NOT LIKE 'OLD-%';

$FINISH
SQL

echo
if [ "$MODE" = "--apply" ]; then
  echo "═══ انحفظ ═══"
else
  echo "═══ تجربة بس — ولا شي انحفظ. للحفظ: $0 $CSV --apply ═══"
fi
