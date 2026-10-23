-- فحص أوقات حصيلة الموظف (ناروتو · 10-23) — قراية بس، بلا أسماء.
-- السؤال: ليش «تقرير العمل» بالساعة ١:٢٠ الصبح والحجز B877 خلص ٩:٥٦؟
-- وقت المرحلة (pressedAt) يجي من **ساعة التلفون**، والسيرفر يقبله إذا بآخر ٢٤ ساعة.
-- receivedAt = لحظة وصوله للسيرفر. إذا الفرق كبير، ساعة التلفون غلط.
--   git fetch origin avg
--   git show origin/avg:docs/data-check-times.sql | docker compose exec -T db psql -U staffmange -d staffmange
\pset pager off
BEGIN READ ONLY;

\echo '═══ ١. مراحل B877: وقت التلفون مقابل وقت وصوله للسيرفر (بغداد) ═══'
SELECT p.stage AS "المرحلة", left(p."employeeId", 8) AS "الموظف",
       to_char(p."pressedAt" + interval '3 hours', 'MM-DD HH24:MI') AS "وقت التلفون",
       to_char(p."receivedAt" + interval '3 hours', 'MM-DD HH24:MI') AS "وصل للسيرفر",
       p.offline AS "بلا نت"
FROM "BookingStageProof" p JOIN "Booking" b ON b.id = p."bookingId"
WHERE b.code = 'B877' ORDER BY p."pressedAt";

\echo '═══ ٢. تقرير وفاتورة B877 ═══'
SELECT 'تقرير' AS "النوع", to_char(w."createdAt" + interval '3 hours', 'MM-DD HH24:MI') AS "انسوّى", w."workStatus"::text AS "الحالة"
FROM "WorkReport" w JOIN "Booking" b ON b.id = w."bookingId" WHERE b.code = 'B877'
UNION ALL
SELECT 'فاتورة', to_char(li."createdAt" + interval '3 hours', 'MM-DD HH24:MI'), li.status
FROM "LeaderInvoice" li JOIN "Booking" b ON b.id = li."bookingId" WHERE b.code = 'B877';

\echo '═══ ٣. فاتورة LDR-20261023-13031b: على أي حجز؟ ═══'
SELECT li."accountingCode" AS "الكود", COALESCE(b.code, '— بلا حجز') AS "الحجز",
       li."customerName" IS NULL AS "بلا اسم زبون", li."netTotal" AS "الصافي", li."pricingMode" AS "النوع",
       to_char(li."createdAt" + interval '3 hours', 'MM-DD HH24:MI') AS "انسوّت"
FROM "LeaderInvoice" li LEFT JOIN "Booking" b ON b.id = li."bookingId"
WHERE li."accountingCode" = 'LDR-20261023-13031b';

\echo '═══ ٤. كل المراحل بآخر ٧ أيام: كم وحدة ساعة التلفون بيها غلط أكثر من ساعة؟ ═══'
SELECT count(*) AS "مراحل", count(*) FILTER (WHERE abs(extract(epoch FROM p."receivedAt" - p."pressedAt")) > 3600 AND NOT p.offline) AS "فرق +ساعة (بنت)",
       count(DISTINCT p."employeeId") FILTER (WHERE abs(extract(epoch FROM p."receivedAt" - p."pressedAt")) > 3600 AND NOT p.offline) AS "موظفين"
FROM "BookingStageProof" p WHERE p."receivedAt" > (now() AT TIME ZONE 'UTC') - interval '7 days';

ROLLBACK;
