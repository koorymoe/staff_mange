#!/bin/bash
# ═══ فحص المطابقة — ما يكتب ولا حرف ═══
#
# يوريك شكد صف من الإكسل يلگه حجزه بالنظام، وبأي طريقة:
#   ١. هاتف + نفس التاريخ بالضبط
#   ٢. هاتف + تاريخ قريب (±٣ أيام)  — النظام القديم ممكن يسجّل بيوم ثاني
#   ٣. هاتف بس (الزبون موجود، بس ماكو حجز بتاريخ قريب)
#   ٤. ماكو مطابقة إطلاقاً
#
# الاستخدام:  ./match-report.sh بيانات-التدقيق-للاستيراد.csv
set -e
CSV_FILE="$1"
DB_CONTAINER="${DB_CONTAINER:-staff_mange-db-1}"
[ -f "$CSV_FILE" ] || { echo "الاستخدام: $0 <csv-file>"; exit 1; }

echo "==> 🔍 فحص فقط — ماكو أي كتابة بقاعدة البيانات"
docker cp "$CSV_FILE" "$DB_CONTAINER:/tmp/audit_match.csv"

docker exec -i "$DB_CONTAINER" psql -U staffmange -d staffmange <<'SQL'
BEGIN;
CREATE TEMP TABLE ai (
  row_no INT, adate TEXT, customer_name TEXT, phone TEXT, location TEXT,
  service TEXT, accounting_code TEXT, customer_code TEXT, vehicle TEXT,
  amount_collected TEXT, device_count TEXT, invoice_amount TEXT,
  invoice_number TEXT, supplier TEXT, note TEXT, verified TEXT
);
\copy ai FROM '/tmp/audit_match.csv' WITH (FORMAT csv, HEADER true)

CREATE TEMP TABLE m AS
SELECT i.row_no, i.customer_name, i.phone, i.adate, i.accounting_code,
       i.invoice_amount, i.amount_collected,
       (SELECT b.id FROM "Booking" b JOIN "Customer" c ON c.id = b."customerId"
         WHERE c.phone = btrim(i.phone) AND b.code LIKE 'OLD-%'
           AND b."createdAt"::date = i.adate::date LIMIT 1) AS exact_id,
       (SELECT b.id FROM "Booking" b JOIN "Customer" c ON c.id = b."customerId"
         WHERE c.phone = btrim(i.phone) AND b.code LIKE 'OLD-%'
           AND ABS(b."createdAt"::date - i.adate::date) <= 3
         ORDER BY ABS(b."createdAt"::date - i.adate::date) LIMIT 1) AS near_id,
       (SELECT COUNT(*) FROM "Booking" b JOIN "Customer" c ON c.id = b."customerId"
         WHERE c.phone = btrim(i.phone) AND b.code LIKE 'OLD-%') AS old_count,
       EXISTS (SELECT 1 FROM "Customer" c WHERE c.phone = btrim(i.phone)) AS cust_exists
FROM ai i WHERE btrim(i.phone) <> '';

\echo ''
\echo '════════ ملخّص المطابقة ════════'
SELECT
  COUNT(*)                                            AS "صفوف بهاتف صالح",
  COUNT(*) FILTER (WHERE exact_id IS NOT NULL)        AS "١ · تطابق تام (هاتف+تاريخ)",
  COUNT(*) FILTER (WHERE exact_id IS NULL AND near_id IS NOT NULL) AS "٢ · تاريخ قريب ±٣ أيام",
  COUNT(*) FILTER (WHERE near_id IS NULL AND old_count > 0)        AS "٣ · زبون عنده حجوزات بس بعيدة",
  COUNT(*) FILTER (WHERE old_count = 0 AND cust_exists)            AS "٤ · زبون موجود بلا حجز قديم",
  COUNT(*) FILTER (WHERE NOT cust_exists)                          AS "٥ · زبون مو موجود إطلاقاً"
FROM m;

\echo ''
\echo '════════ الصفوف الي ما انطابقت (تحتاج قرارك) ════════'
SELECT row_no AS "الصف", customer_name AS "الزبون", phone AS "الهاتف",
       adate AS "التاريخ", invoice_amount AS "مبلغ الفاتورة",
       CASE WHEN NOT cust_exists THEN 'زبون مو موجود'
            WHEN old_count = 0   THEN 'ماكو حجز قديم'
            ELSE 'حجوزاته بتواريخ بعيدة (' || old_count || ' حجز)' END AS "السبب"
FROM m WHERE near_id IS NULL ORDER BY row_no;

\echo ''
\echo '════════ عيّنة تطابق قريب — راجعها بعينك ════════'
SELECT m.row_no AS "الصف", m.customer_name AS "الزبون",
       m.adate AS "تاريخ الإكسل", b."createdAt"::date AS "تاريخ النظام",
       b.code AS "كود الحجز", m.invoice_amount AS "المبلغ"
FROM m JOIN "Booking" b ON b.id = m.near_id
WHERE m.exact_id IS NULL ORDER BY m.row_no LIMIT 20;

\echo ''
\echo '════════ هل «مرتضى عباس» موظف بالنظام؟ ════════'
SELECT id, name, role, status FROM "Employee" WHERE name LIKE '%مرتضى%';

ROLLBACK;
SQL

docker exec "$DB_CONTAINER" rm -f /tmp/audit_match.csv
echo ""
echo "==> ✅ خلص الفحص — ماكو شي انكتب."
