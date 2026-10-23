-- تصحيح البيانات ٢ (الـKPI) — المعاينة (ناروتو · 10-23). قراية بس، ما يغيّر شي.
-- الخصم صار سالب دائماً؛ هذا يعرض الخصومات اليدوية القديمة الي مخزونة موجب.
--   git fetch origin avg
--   git show origin/avg:docs/data-fix-2-preview.sql | docker compose exec -T db psql -U staffmange -d staffmange
\pset pager off
BEGIN READ ONLY;

\echo '═══ ١. خصومات يدوية مخزونة موجب — راح تنقلب سالب (المبلغ ما يتغير) ═══'
SELECT left(id, 8) AS "التقييم", left("employeeId", 8) AS "الموظف", points AS "النقاط هسه", -points AS "بعد التصحيح",
       "deductionAmount"::bigint AS "المبلغ", CASE WHEN cancelled THEN 'ملغي' ELSE 'فعّال' END AS "الحالة",
       to_char(baghdad_date("createdAt"), 'YYYY-MM-DD') AS "اليوم", split_part(reason, ':', 1) AS "المعيار"
FROM "KpiEvaluation"
WHERE points > 0 AND reason NOT LIKE 'شكوى الزبائن%' AND reason <> 'إكمال التدريب بنجاح'
ORDER BY "createdAt";

\echo '═══ ٢. «إكمال التدريب» بمبلغ خصم — المبلغ راح يصير صفر ═══'
SELECT count(*) AS "العدد", COALESCE(sum("deductionAmount"), 0)::bigint AS "المبلغ"
FROM "KpiEvaluation" WHERE reason = 'إكمال التدريب بنجاح' AND "deductionAmount" <> 0;

ROLLBACK;
