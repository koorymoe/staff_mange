-- ═══ تدقيق: تسويات دوار انختمت بدوار غلط (بيانات قديمة) ═══
--
-- الخلفية: قبل الإصلاح، واجهة الموظف چانت **تخمّن** الدوار من آخر
-- عملية تسليم، فتسوية تغطي أكثر من دوار تنختم كلها بدوار واحد —
-- والفلوس ترجع لمكان غلط.
--
-- ⚠️ هذا الاستعلام **قراءة فقط**. ما يعدّل ولا صف. الغرض: نطلّع
-- الحالات المشكوك فيها حتى صاحب النظام يقرر التصحيح بنفسه — تعديل
-- فلوس بلا علمه أسوأ من الخلل نفسه.
--
-- التشغيل:
--   psql "$DATABASE_URL" -f backend-go/scripts/fund_split_audit.sql

-- ① الموظفون الي أخذوا من أكثر من دوار (هذولا وحدهم المعرَّضون للعلّة)
WITH multi AS (
    SELECT "employeeId"
    FROM "RevolvingFundTxn"
    WHERE kind = 'DISBURSE'
    GROUP BY "employeeId"
    HAVING COUNT(DISTINCT "fundId") > 1
)
-- ② رصيد كل دوار لكل واحد منهم.
--    ⚠️ **رصيد بالسالب = دليل قاطع**: رجعت لهذا الدوار فلوس أكثر
--    مما طلع منه، يعني تسوية دوار ثاني انحسبت عليه.
SELECT e.name                                   AS "الموظف",
       f.name                                   AS "الدوار",
       SUM(CASE WHEN t.kind = 'DISBURSE' THEN t.amount ELSE 0 END)      AS "استلم",
       SUM(CASE WHEN t.kind = 'SETTLEMENT' AND t.status = 'APPROVED'
                THEN t."spentAmount" + t."returnedAmount" ELSE 0 END)   AS "انتسوّى",
       SUM(CASE WHEN t.kind = 'DISBURSE' THEN t.amount
                WHEN t.kind = 'SETTLEMENT' AND t.status = 'APPROVED'
                THEN -(t."spentAmount" + t."returnedAmount")
                ELSE 0 END)                                             AS "الرصيد",
       CASE WHEN SUM(CASE WHEN t.kind = 'DISBURSE' THEN t.amount
                          WHEN t.kind = 'SETTLEMENT' AND t.status = 'APPROVED'
                          THEN -(t."spentAmount" + t."returnedAmount")
                          ELSE 0 END) < -0.001
            THEN '⚠️ سالب — تسوية دوار ثاني انحسبت هنا'
            ELSE '' END                                                 AS "ملاحظة"
FROM "RevolvingFundTxn" t
JOIN multi        ON multi."employeeId" = t."employeeId"
JOIN "Employee" e ON e.id = t."employeeId"
JOIN "RevolvingFund" f ON f.id = t."fundId"
GROUP BY e.name, f.name
ORDER BY e.name, f.name;
