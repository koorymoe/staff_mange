-- فحص صدق الأرقام ٣ (ناروتو · 10-22) — قراية بس (READ ONLY + ROLLBACK)، ما يغيّر شي.
-- تفصيل الي طلّعه data-audit.sql حتى نعرف السبب ونقرر التصحيح. بلا أسماء ولا هواتف.
-- التشغيل على السيرفر:
--   git fetch origin avg
--   git show origin/avg:docs/data-audit-2.sql | docker compose exec -T db psql -U staffmange -d staffmange
\pset pager off
BEGIN READ ONLY;

\echo '═══ ١. الجلسات الطويلة (> ١٦ ساعة): منين جت؟ — حسب سنة الخروج ونوعها ═══'
SELECT extract(year FROM "checkOut")::int AS "سنة الخروج",
       CASE WHEN "approvedOut" THEN 'مصحّحة يدوياً/بدليل' ELSE 'بصمة/تلقائي' END AS "المصدر",
       count(*) AS "العدد",
       round(avg(extract(epoch FROM "checkOut" - "checkIn") / 3600)) AS "معدل الساعات"
FROM "Attendance" WHERE "checkOut" - "checkIn" > interval '16 hours'
GROUP BY 1, 2 ORDER BY 1, 2;

\echo '═══ ١ب. أسوأ ٢٠ جلسة: الدخول والخروج كما هم بالقاعدة (توقيت UTC) ═══'
SELECT left("employeeId", 8) AS "الموظف", to_char("checkIn", 'YYYY-MM-DD HH24:MI') AS "الدخول",
       to_char("checkOut", 'YYYY-MM-DD HH24:MI') AS "الخروج",
       round(extract(epoch FROM "checkOut" - "checkIn") / 3600) AS "ساعات", "approvedOut" AS "مصحّحة"
FROM "Attendance" WHERE "checkOut" - "checkIn" > interval '16 hours'
ORDER BY "checkOut" - "checkIn" DESC LIMIT 20;

\echo '═══ ٢. أيام بعدة جلسات (٦٠ يوم): أوقات كل جلسة (بغداد) ═══'
WITH multi AS (
  SELECT "employeeId", "checkIn"::date d FROM "Attendance"
  WHERE "checkIn" >= now() - interval '60 days' GROUP BY 1, 2 HAVING count(*) > 2)
SELECT left(a."employeeId", 8) AS "الموظف", to_char(m.d, 'MM-DD') AS "اليوم",
       string_agg(to_char(a."checkIn" + interval '3 hours', 'HH24:MI') || '→' ||
                  COALESCE(to_char(a."checkOut" + interval '3 hours', 'HH24:MI'), '…'), '  ' ORDER BY a."checkIn") AS "الجلسات"
FROM multi m JOIN "Attendance" a ON a."employeeId" = m."employeeId" AND a."checkIn"::date = m.d
GROUP BY 1, 2 ORDER BY 2 DESC LIMIT 20;

\echo '═══ ٣. الفواتير المكررة (نفس الليدر ونفس المبلغ): أي وحدة الزايدة ═══'
WITH dup AS (
  SELECT li."bookingId" FROM "LeaderInvoice" li WHERE li."revokedAt" IS NULL AND li."bookingId" IS NOT NULL
  GROUP BY 1 HAVING count(*) > 1 AND count(DISTINCT li."employeeId") = 1 AND count(DISTINCT li."netTotal") = 1)
SELECT b.code AS "الحجز", left(li.id, 8) AS "الفاتورة", to_char(li."createdAt" + interval '3 hours', 'YYYY-MM-DD HH24:MI') AS "انسوّت",
       li."netTotal" AS "المبلغ", li.status AS "الحالة", li."accountingCode" AS "الكود المحاسبي",
       CASE WHEN row_number() OVER (PARTITION BY li."bookingId" ORDER BY li."createdAt") = 1 THEN 'الأصل' ELSE '⚠️ الزايدة' END AS "الحكم"
FROM dup JOIN "LeaderInvoice" li ON li."bookingId" = dup."bookingId" AND li."revokedAt" IS NULL
JOIN "Booking" b ON b.id = li."bookingId" ORDER BY b.code, li."createdAt";

\echo '═══ ٤. حجوزات بفاتورتين من ليدرين: نفس اليوم (غلط) لو يوم ثاني (صح) ═══'
WITH multi AS (
  SELECT li."bookingId", count(*) n, count(DISTINCT baghdad_date(li."createdAt")) days
  FROM "LeaderInvoice" li WHERE li."revokedAt" IS NULL AND li."bookingId" IS NOT NULL
  GROUP BY 1 HAVING count(*) > 1 AND count(DISTINCT li."employeeId") > 1)
SELECT CASE WHEN days < n THEN 'بنفس اليوم — غلط حسب القاعدة' ELSE 'كل فاتورة بيوم — صح' END AS "النوع",
       count(*) AS "العدد",
       string_agg(b.code, ' · ' ORDER BY b.code) AS "الحجوزات"
FROM multi JOIN "Booking" b ON b.id = multi."bookingId" GROUP BY 1;

ROLLBACK;
