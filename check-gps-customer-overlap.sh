#!/bin/bash
# ═══ فحص التداخل بين زبائن الجي بي اس وزبائن الحجوزات — ما يكتب ولا حرف ═══
#
# النظام عنده هويتان منفصلتان للزبون:
#   ١. "Customer" — زبون الحجوزات والفواتير والشكاوى (٥٥٤+ سجل، رقم
#      هاتف فريد، مربوط بكل شي: حجز، فاتورة، مشروع).
#   ٢. "GpsCustomer" — زبون تتبع الجي بي اس (الشرائح، الأجهزة،
#      الاشتراكات)، جدول مستقل تماماً، مستورد من ملف إكسل قديم
#      (cmd/importgps)، بلا أي عمود يربطه بجدول Customer.
#
# يعني نفس الشخص الحقيقي ممكن يصير عنده سجلين منفصلين — وحده يظهر
# بشاشة الحجوزات وحده ثاني بشاشة تتبع الجي بي اس، بلا ما النظام
# يعرف إنهما نفس الزبون.
#
# ⚠️ أكو محاولة سابقة غير مكتملة (CustomerServiceTag + CustomerGpsInfo
# على جدول Customer نفسه) بس **ماكو أي مسار API يستعملها** — كود ميت.
# فالربط الفعلي لسا ما صار.
#
# هذا السكربت يوريك **كم** من زبائن الجي بي اس عندهم نفس رقم الهاتف
# موجود أصلاً بجدول الزبائن — قبل أي قرار عن شلون نوحّدهم.
#
# الاستخدام:  ./check-gps-customer-overlap.sh
set -e
DB_CONTAINER="${DB_CONTAINER:-staff_mange-db-1}"

echo "==> 🔍 فحص فقط — ماكو أي كتابة بقاعدة البيانات"

docker exec -i "$DB_CONTAINER" psql -U staffmange -d staffmange <<'SQL'
-- إجمالي زبائن الجي بي اس
SELECT count(*) AS "إجمالي زبائن الجي بي اس (GpsCustomer)" FROM "GpsCustomer";

-- إجمالي زبائن الحجوزات
SELECT count(*) AS "إجمالي زبائن الحجوزات (Customer)" FROM "Customer";

-- كم من زبائن الجي بي اس عنده نفس رقم هاتف بجدول الحجوزات — يعني نفس
-- الشخص محتمل عنده سجلان
SELECT count(*) AS "زبائن جي بي اس عندهم تطابق هاتف بجدول الحجوزات"
FROM "GpsCustomer" g
WHERE btrim(g.phone) <> ''
  AND EXISTS (SELECT 1 FROM "Customer" c WHERE c.phone = btrim(g.phone));

-- كم زبون جي بي اس بلا هاتف أصلاً (ما نقدر نطابقه آلياً)
SELECT count(*) AS "زبائن جي بي اس بلا رقم هاتف مسجَّل"
FROM "GpsCustomer" WHERE COALESCE(btrim(phone), '') = '';

-- عيّنة من التطابقات — للتأكد الأسماء فعلاً نفس الشخص مو صدفة رقم
SELECT g."fullName" AS "الاسم بجي بي اس", c.name AS "الاسم بالحجوزات",
       g.phone AS "الهاتف"
FROM "GpsCustomer" g
JOIN "Customer" c ON c.phone = btrim(g.phone)
WHERE btrim(g.phone) <> ''
LIMIT 20;
SQL
