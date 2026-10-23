-- تصحيح البيانات ٢ (الـKPI) — التنفيذ (ناروتو · 10-23). ⚠️ بس بعد المعاينة وأمر دازاي «نفّذ».
--   git show origin/avg:docs/data-fix-2-apply.sql | docker compose exec -T db psql -U staffmange -d staffmange
-- معاملة وحدة، حارس أعداد، ونسخة احتياط بجدول DataFix2Backup.
\set ON_ERROR_STOP on
\pset pager off
BEGIN;

CREATE TABLE IF NOT EXISTS "DataFix2Backup" (
  "rowId" TEXT NOT NULL, old JSONB NOT NULL, "fixedAt" TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TEMP TABLE fix_kpi ON COMMIT DROP AS
SELECT id FROM "KpiEvaluation"
WHERE (points > 0 AND reason NOT LIKE 'شكوى الزبائن%' AND reason <> 'إكمال التدريب بنجاح')
   OR (reason = 'إكمال التدريب بنجاح' AND "deductionAmount" <> 0);
DO $$ BEGIN
  IF (SELECT count(*) FROM fix_kpi) > 30 THEN
    RAISE EXCEPTION 'تقييمات أكثر من المتوقع (%) — ما تغيّر شي', (SELECT count(*) FROM fix_kpi);
  END IF;
END $$;
INSERT INTO "DataFix2Backup" ("rowId", old)
SELECT k.id, to_jsonb(k) FROM "KpiEvaluation" k WHERE k.id IN (SELECT id FROM fix_kpi);

-- ١. الخصم اليدوي سالب (المبلغ يبقى)
UPDATE "KpiEvaluation" SET points = -points
WHERE id IN (SELECT id FROM fix_kpi) AND points > 0 AND reason <> 'إكمال التدريب بنجاح';
-- ٢. «إكمال التدريب» مكافأة بلا مبلغ
UPDATE "KpiEvaluation" SET "deductionAmount" = 0
WHERE id IN (SELECT id FROM fix_kpi) AND reason = 'إكمال التدريب بنجاح';

\echo '═══ النتيجة ═══'
SELECT count(*) AS "انصلح" FROM "DataFix2Backup" WHERE "fixedAt" >= now() - interval '1 minute';
COMMIT;
