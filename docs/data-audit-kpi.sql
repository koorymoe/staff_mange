-- فحص نقاط الـKPI (ناروتو · 10-23) — قراية بس (READ ONLY + ROLLBACK)، بلا أسماء.
--   git fetch origin avg
--   git show origin/avg:docs/data-audit-kpi.sql | docker compose exec -T db psql -U staffmange -d staffmange
-- السؤال: الخصم ينخزن بإشارة وحدة؟ شاشة الـKPI تدزّ الخصم موجب (١..٨)، وخصم الجودة
-- ينخزن سالب، و«إكمال التدريب» +١ (مكافأة) بس وياه مبلغ خصم ١٠ آلاف.
\pset pager off
BEGIN READ ONLY;

\echo '═══ ١. التقييمات (غير الملغية) حسب المصدر والإشارة ═══'
SELECT CASE WHEN reason LIKE 'شكوى الزبائن%' THEN 'جودة (تلقائي)'
            WHEN reason = 'إكمال التدريب بنجاح' THEN 'إكمال تدريب'
            ELSE 'يدوي (شاشة KPI)' END AS "المصدر",
       CASE WHEN points > 0 THEN 'موجب' WHEN points < 0 THEN 'سالب' ELSE 'صفر' END AS "الإشارة",
       count(*) AS "العدد", sum(points) AS "مجموع النقاط", sum("deductionAmount")::bigint AS "مجموع مبلغ الخصم"
FROM "KpiEvaluation" WHERE NOT cancelled
GROUP BY 1, 2 ORDER BY 1, 2;

\echo '═══ ٢. المبلغ ما يطابق النقاط (لازم |النقاط| × ١٠ آلاف) ═══'
SELECT count(*) AS "العدد", string_agg(left(id, 8) || ' ' || points || '→' || "deductionAmount"::bigint, ' · ') FILTER (WHERE rn <= 10) AS "أمثلة"
FROM (SELECT *, row_number() OVER () rn FROM "KpiEvaluation"
      WHERE NOT cancelled AND "deductionAmount" <> abs(points) * 10000) t;

\echo '═══ ٣. إلغاء ناقص (cancelled و cancelledAt ما يتطابقون) ═══'
SELECT count(*) AS "العدد" FROM "KpiEvaluation"
WHERE cancelled <> ("cancelledAt" IS NOT NULL);

\echo '═══ ٤. موظفين ينحسبلهم الخصم «مكافأة»: عندهم خصم يدوي موجب (الشهر الحالي والماضي) ═══'
SELECT to_char(baghdad_date("createdAt"), 'YYYY-MM') AS "الشهر", count(DISTINCT "employeeId") AS "موظفين",
       count(*) AS "تقييمات", sum(points) AS "نقاط تنحسب موجبة وهي خصم"
FROM "KpiEvaluation"
WHERE NOT cancelled AND points > 0 AND reason <> 'إكمال التدريب بنجاح'
  AND "createdAt" >= date_trunc('month', now()) - interval '1 month'
GROUP BY 1 ORDER BY 1;

ROLLBACK;
