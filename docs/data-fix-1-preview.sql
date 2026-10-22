-- تصحيح البيانات ١ — المعاينة (ناروتو · 10-22، بموافقة دازاي). قراية بس، ما يغيّر شي.
-- يعرض بالضبط شنو راح يتغير بملف التنفيذ data-fix-1-apply.sql.
--   git fetch origin avg
--   git show origin/avg:docs/data-fix-1-preview.sql | docker compose exec -T db psql -U staffmange -d staffmange
\pset pager off
BEGIN READ ONLY;

\echo '═══ ١. الفواتير الزايدة الي راح ينسحب اعتمادها (٤) ═══'
SELECT b.code AS "الحجز", left(li.id, 8) AS "الفاتورة", li."netTotal" AS "المبلغ", li.status AS "الحالة",
       COALESCE(li."externalInvoiceNumber", '—') AS "رقم المحاسب",
       to_char(li."createdAt" + interval '3 hours', 'MM-DD HH24:MI') AS "انسوّت"
FROM "LeaderInvoice" li JOIN "Booking" b ON b.id = li."bookingId"
WHERE li."revokedAt" IS NULL AND li.status = 'APPROVED'
  AND ((b.code = 'B309' AND li.id LIKE '273681e5%') OR (b.code = 'B363' AND li.id LIKE '43cbf3d9%')
    OR (b.code = 'B585' AND li.id LIKE 'b15bf94a%') OR (b.code = 'B588' AND li.id LIKE 'b88471db%'))
ORDER BY b.code;

\echo '═══ ٢. جلسات الدوام الي راح تنقص لنهاية الشفت (مو المصححة يدوياً) ═══'
\echo '    الساعات المحسوبة ما تتغير — النظام أصلاً يحسبها لنهاية الشفت. التصحيح ينظّف الوقت الخام بس.'
SELECT count(*) AS "عدد الجلسات",
       round(sum(extract(epoch FROM "checkOut" - "checkIn")) / 3600) AS "الساعات الخام هسه",
       round(sum(extract(epoch FROM att_counted_end("employeeId", "checkIn", "checkOut") - "checkIn")) / 3600) AS "الساعات بعد التصحيح"
FROM "Attendance" WHERE "checkOut" - "checkIn" > interval '16 hours' AND NOT "approvedOut";

SELECT left("employeeId", 8) AS "الموظف", to_char("checkIn" + interval '3 hours', 'YYYY-MM-DD HH24:MI') AS "الدخول",
       to_char("checkOut" + interval '3 hours', 'YYYY-MM-DD HH24:MI') AS "الخروج هسه",
       to_char(att_counted_end("employeeId", "checkIn", "checkOut") + interval '3 hours', 'YYYY-MM-DD HH24:MI') AS "الخروج بعد التصحيح"
FROM "Attendance" WHERE "checkOut" - "checkIn" > interval '16 hours' AND NOT "approvedOut"
ORDER BY "checkOut" - "checkIn" DESC LIMIT 15;

\echo '═══ ٣. الحجوزات السبعة بفاتورتين بنفس اليوم — للقرار بس (ما تتغير) ═══'
SELECT b.code AS "الحجز", left(li."employeeId", 8) AS "الليدر", left(li.id, 8) AS "الفاتورة", li."netTotal" AS "المبلغ",
       li.status AS "الحالة", COALESCE(li."externalInvoiceNumber", '—') AS "رقم المحاسب",
       to_char(li."createdAt" + interval '3 hours', 'MM-DD HH24:MI') AS "انسوّت"
FROM "LeaderInvoice" li JOIN "Booking" b ON b.id = li."bookingId"
WHERE li."revokedAt" IS NULL AND b.code IN ('B384', 'B745', 'B783', 'B804', 'B838', 'B841', 'B851')
ORDER BY b.code, li."createdAt";

ROLLBACK;
