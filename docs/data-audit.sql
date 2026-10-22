-- فحص صدق الأرقام (ناروتو · 10-22) — قراية بس (READ ONLY + ROLLBACK)، ما يغيّر شي.
-- بلا أسماء ولا هواتف: رمز الحجز وأول ٨ حروف من معرّف الموظف بس.
-- التشغيل على السيرفر:
--   docker compose exec -T db psql -U staffmange -d staffmange < docs/data-audit.sql
-- كل سطر: الفحص · العدد · أول ١٠ أمثلة.
\pset pager off
BEGIN READ ONLY;

WITH
-- ═══ ١. الفواتير المكررة على نفس الحجز (الـ54) ═══
inv AS (
  SELECT b.code, li."bookingId", count(*) n, count(DISTINCT li."employeeId") leaders,
         count(DISTINCT li."netTotal") amounts,
         count(*) FILTER (WHERE li.status = 'APPROVED') approved,
         round(extract(epoch FROM max(li."createdAt") - min(li."createdAt")) / 60) mins,
         count(*) FILTER (WHERE li."isFree") free
  FROM "LeaderInvoice" li JOIN "Booking" b ON b.id = li."bookingId"
  WHERE li."revokedAt" IS NULL GROUP BY 1, 2 HAVING count(*) > 1),
visits AS (
  SELECT "bookingId", count(*) v FROM "BookingVisit" GROUP BY 1),
inv_cls AS (
  SELECT i.*, COALESCE(v.v, 0) visits,
    CASE WHEN i.leaders = 1 AND i.amounts = 1 THEN 'أ'
         WHEN i.leaders = 1 THEN 'ب'
         ELSE 'ج' END cls
  FROM inv i LEFT JOIN visits v ON v."bookingId" = i."bookingId"),
-- ═══ ٢. «تم» المكررة (الـ11) ═══
done2 AS (
  SELECT b.code, v."bookingId", count(*) n, string_agg(v."visitNumber"::text, '/' ORDER BY v."visitNumber") nums
  FROM "BookingVisit" v JOIN "Booking" b ON b.id = v."bookingId"
  WHERE v.outcome = 'DONE' GROUP BY 1, 2 HAVING count(*) > 1),
-- ═══ ٣. نقاط ماتركس ═══
ms_dup AS (
  SELECT left("employeeId", 8) emp, source, "sourceId", rule, count(*) n
  FROM "MatrixScore" WHERE "cancelledAt" IS NULL GROUP BY 1, 2, 3, 4 HAVING count(*) > 1),
ms_done2 AS (
  SELECT d.code, left(m."employeeId", 8) emp, count(*) n
  FROM done2 d JOIN "MatrixScore" m ON m.source = 'BOOKING' AND m."sourceId" = d."bookingId" AND m."cancelledAt" IS NULL
  GROUP BY 1, 2, m.rule HAVING count(*) > 1),
top_roles AS (
  SELECT id FROM "Employee" WHERE role::text IN ('ADMIN', 'OWNER')),
-- ═══ ٤. الدوام ═══
att_open AS (
  SELECT left(a."employeeId", 8) emp, to_char(a."checkIn", 'MM-DD') d
  FROM "Attendance" a WHERE a."checkOut" IS NULL AND a."checkIn"::date < (now() AT TIME ZONE 'Asia/Baghdad')::date - 1),
att_neg AS (
  SELECT left("employeeId", 8) emp, to_char("checkIn", 'MM-DD') d FROM "Attendance" WHERE "checkOut" < "checkIn"),
att_long AS (
  SELECT left("employeeId", 8) emp, to_char("checkIn", 'MM-DD') d,
         round(extract(epoch FROM "checkOut" - "checkIn") / 3600) h
  FROM "Attendance" WHERE "checkOut" - "checkIn" > interval '16 hours'),
att_multi AS (
  SELECT left("employeeId", 8) emp, "checkIn"::date d, count(*) n
  FROM "Attendance" WHERE "checkIn" >= now() - interval '60 days' GROUP BY 1, 2 HAVING count(*) > 1),
-- ═══ ٥. الانضباط ═══
disc_noreason AS (
  SELECT left("employeeId", 8) emp, kind, delta FROM "DisciplineEvent"
  WHERE delta < 0 AND COALESCE(btrim(reason), '') = ''),
disc_dup AS (
  SELECT left("employeeId", 8) emp, "bookingId", kind, "createdAt"::date d, count(*) n
  FROM "DisciplineEvent" WHERE "bookingId" IS NOT NULL AND delta < 0 GROUP BY 1, 2, 3, 4 HAVING count(*) > 1)

SELECT '١أ فواتير مكررة: نفس الليدر ونفس المبلغ (الأرجح تكرار غلط)' AS "الفحص", count(*)::text AS "العدد",
       string_agg(code || ' (' || n || ' فواتير، ' || approved || ' معتمدة، ' || mins || ' دقيقة)', ' · ') FILTER (WHERE rn <= 10) AS "أمثلة"
  FROM (SELECT *, row_number() OVER (ORDER BY approved DESC, n DESC) rn FROM inv_cls WHERE cls = 'أ') t
UNION ALL SELECT '١ب فواتير مكررة: نفس الليدر ومبلغ مختلف', count(*)::text,
       string_agg(code || ' (' || n || ' فواتير، ' || visits || ' زيارة، ' || approved || ' معتمدة)', ' · ') FILTER (WHERE rn <= 10)
  FROM (SELECT *, row_number() OVER (ORDER BY approved DESC) rn FROM inv_cls WHERE cls = 'ب') t
UNION ALL SELECT '١ج فواتير متعددة: أكثر من ليدر (ممكن شرعية)', count(*)::text,
       string_agg(code || ' (' || n || ' فواتير، ' || leaders || ' ليدر، ' || visits || ' زيارة)', ' · ') FILTER (WHERE rn <= 10)
  FROM (SELECT *, row_number() OVER (ORDER BY n DESC) rn FROM inv_cls WHERE cls = 'ج') t
UNION ALL SELECT '١د مبلغ معتمد زايد بسبب التكرار (أ)', COALESCE(sum(extra), 0)::text, NULL FROM (
  SELECT sum(li."netTotal") - max(li."netTotal") extra FROM "LeaderInvoice" li
  JOIN inv_cls c ON c."bookingId" = li."bookingId" AND c.cls = 'أ'
  WHERE li."revokedAt" IS NULL AND li.status = 'APPROVED' GROUP BY li."bookingId") t

UNION ALL SELECT '٢ حجوزات بيها أكثر من «تم» — أرقام الزيارات', count(*)::text,
       string_agg(code || ' (' || nums || ')', ' · ') FILTER (WHERE rn <= 10)
  FROM (SELECT *, row_number() OVER (ORDER BY n DESC) rn FROM done2) t
UNION ALL SELECT '٢ب منها انحسبت نقاط ماتركس مرتين لنفس الموظف', count(*)::text,
       string_agg(code || ' ← ' || emp || ' ×' || n, ' · ') FILTER (WHERE rn <= 10)
  FROM (SELECT *, row_number() OVER () rn FROM ms_done2) t

UNION ALL SELECT '٣ نقاط ماتركس مكررة (نفس الموظف والمصدر والقاعدة)', count(*)::text,
       string_agg(emp || ' ' || source || '/' || rule || ' ×' || n, ' · ') FILTER (WHERE rn <= 10)
  FROM (SELECT *, row_number() OVER (ORDER BY n DESC) rn FROM ms_dup) t
UNION ALL SELECT '٣ب نقاط ماتركس على المدير أو المالك (ممنوع)', count(*)::text, NULL
  FROM "MatrixScore" WHERE "cancelledAt" IS NULL AND "employeeId" IN (SELECT id FROM top_roles)
UNION ALL SELECT '٣ج تقييمات KPI على المدير أو المالك (ممنوع)', count(*)::text, NULL
  FROM "KpiEvaluation" WHERE NOT cancelled AND "employeeId" IN (SELECT id FROM top_roles)
UNION ALL SELECT '٣د تقييمات الكوادر على المدير أو المالك (ممنوع)', count(*)::text, NULL
  FROM "StaffRating" WHERE "rateeId" IN (SELECT id FROM top_roles)
UNION ALL SELECT '٣هـ تقييمات KPI بيها خصم فلوس (ممنوع: غرامات فلوس)', count(*)::text,
       'المجموع ' || COALESCE(sum("deductionAmount"), 0)::text
  FROM "KpiEvaluation" WHERE NOT cancelled AND "deductionAmount" > 0

UNION ALL SELECT '٤ دوام مفتوح من أيام سابقة (ما سجل خروج)', count(*)::text,
       string_agg(emp || ' ' || d, ' · ') FILTER (WHERE rn <= 10)
  FROM (SELECT *, row_number() OVER (ORDER BY d DESC) rn FROM att_open) t
UNION ALL SELECT '٤ب خروج قبل الدخول', count(*)::text,
       string_agg(emp || ' ' || d, ' · ') FILTER (WHERE rn <= 10)
  FROM (SELECT *, row_number() OVER () rn FROM att_neg) t
UNION ALL SELECT '٤ج جلسة دوام أطول من ١٦ ساعة', count(*)::text,
       string_agg(emp || ' ' || d || ' ' || h || 'س', ' · ') FILTER (WHERE rn <= 10)
  FROM (SELECT *, row_number() OVER (ORDER BY h DESC) rn FROM att_long) t
UNION ALL SELECT '٤د أكثر من جلسة دوام باليوم (٦٠ يوم)', count(*)::text,
       string_agg(emp || ' ' || to_char(d, 'MM-DD') || ' ×' || n, ' · ') FILTER (WHERE rn <= 10)
  FROM (SELECT *, row_number() OVER (ORDER BY n DESC) rn FROM att_multi) t

UNION ALL SELECT '٥ خصم نقاط انضباط بلا سبب', count(*)::text,
       string_agg(emp || ' ' || kind || ' ' || delta, ' · ') FILTER (WHERE rn <= 10)
  FROM (SELECT *, row_number() OVER () rn FROM disc_noreason) t
UNION ALL SELECT '٥ب نفس خصم الانضباط مكرر على نفس الحجز بنفس اليوم', count(*)::text,
       string_agg(emp || ' ' || kind || ' ×' || n, ' · ') FILTER (WHERE rn <= 10)
  FROM (SELECT *, row_number() OVER (ORDER BY n DESC) rn FROM disc_dup) t
UNION ALL SELECT '٦ حجوزات خلصت بلا فاتورة (٦٠ يوم) — الرموز', count(*)::text,
       string_agg(code, ' · ') FILTER (WHERE rn <= 12)
  FROM (SELECT b.code, row_number() OVER (ORDER BY b."completedAt" DESC) rn FROM "Booking" b
        WHERE b.status = 'COMPLETED' AND b."archivedAt" IS NULL AND b."settledLegacyAt" IS NULL
          AND b."completedAt" >= (now() AT TIME ZONE 'UTC') - interval '60 days'
          AND NOT EXISTS (SELECT 1 FROM "LeaderInvoice" li WHERE li."bookingId" = b.id)) t
UNION ALL SELECT '٦ب حجوزات قيد العمل بلا أي موظف مكلّف', count(*)::text,
       string_agg(code, ' · ') FILTER (WHERE rn <= 10)
  FROM (SELECT b.code, row_number() OVER () rn FROM "Booking" b
        WHERE b.status = 'IN_PROGRESS' AND b."archivedAt" IS NULL
          AND NOT EXISTS (SELECT 1 FROM "BookingAssignment" a WHERE a."bookingId" = b.id)) t
UNION ALL SELECT '٦ج حجوزات «خلصت» بلا ولا زيارة «تم»', count(*)::text,
       string_agg(code, ' · ') FILTER (WHERE rn <= 10)
  FROM (SELECT b.code, row_number() OVER (ORDER BY b."completedAt" DESC) rn FROM "Booking" b
        WHERE b.status = 'COMPLETED' AND b."completedAt" >= (now() AT TIME ZONE 'UTC') - interval '60 days'
          AND EXISTS (SELECT 1 FROM "BookingVisit" v WHERE v."bookingId" = b.id)
          AND NOT EXISTS (SELECT 1 FROM "BookingVisit" v WHERE v."bookingId" = b.id AND v.outcome = 'DONE')) t

UNION ALL SELECT '٧ مشاريع بلا مسؤول — حسب المرحلة', COALESCE(sum(n), 0)::text,
       string_agg(stage || ' ×' || n, ' · ')
  FROM (SELECT COALESCE(p.stage, '—') stage, count(*) n FROM "Project" p
        WHERE COALESCE(p.stage, '') NOT LIKE '%مرفوض%' AND p."responsibleEmployeeId" IS NULL GROUP BY 1) t;

ROLLBACK;
