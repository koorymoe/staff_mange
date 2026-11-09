-- تصحيح البيانات ١ — التنفيذ (ناروتو · 10-22، بموافقة دازاي على الثلاثة).
-- ⚠️ شغّل المعاينة data-fix-1-preview.sql أول، وهذا بس بعد ما تشوفها وتگول «نفّذ».
--   git fetch origin avg
--   git show origin/avg:docs/data-fix-1-apply.sql | docker compose exec -T db psql -U staffmange -d staffmange
-- كلشي بمعاملة وحدة: إذا أي عدد ما طابق المتوقع، ينلغي كله وما يتغير شي.
-- نسخة من القيم القديمة تنحفظ بجدول DataFix1Backup حتى نكدر نرجّعها.
\set ON_ERROR_STOP on
\pset pager off
BEGIN;

CREATE TABLE IF NOT EXISTS "DataFix1Backup" (
  kind TEXT NOT NULL, "rowId" TEXT NOT NULL, old JSONB NOT NULL, "fixedAt" TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ١. سحب اعتماد الفواتير الزايدة (مو حذف — السجل يبقى، ويرجع بإعادة الاعتماد)
CREATE TEMP TABLE fix_inv ON COMMIT DROP AS
SELECT li.id FROM "LeaderInvoice" li JOIN "Booking" b ON b.id = li."bookingId"
WHERE li."revokedAt" IS NULL AND li.status = 'APPROVED'
  AND ((b.code = 'B309' AND li.id LIKE '273681e5%') OR (b.code = 'B363' AND li.id LIKE '43cbf3d9%')
    OR (b.code = 'B585' AND li.id LIKE 'b15bf94a%') OR (b.code = 'B588' AND li.id LIKE 'b88471db%'));
DO $$ BEGIN
  IF (SELECT count(*) FROM fix_inv) <> 4 THEN
    RAISE EXCEPTION 'الفواتير الزايدة لازم ٤ — لگيت %، ما تغيّر شي', (SELECT count(*) FROM fix_inv);
  END IF;
END $$;
INSERT INTO "DataFix1Backup" (kind, "rowId", old)
SELECT 'invoice', li.id, to_jsonb(li) FROM "LeaderInvoice" li WHERE li.id IN (SELECT id FROM fix_inv);
UPDATE "LeaderInvoice" SET status = 'SUBMITTED', "revokedAt" = now(),
       "revokedById" = (SELECT id FROM "Employee" WHERE role::text = 'OWNER' ORDER BY "createdAt" LIMIT 1),
       "revokeReason" = 'تكرار — نفس الليدر ونفس المبلغ على نفس الحجز (تصحيح البيانات ١، بموافقة دازاي 10-22)',
       "revokedCount" = "revokedCount" + 1, "approvedByEmployeeId" = NULL, "approvedAt" = NULL
WHERE id IN (SELECT id FROM fix_inv);

-- ٢. جلسات الدوام المفتوحة أسابيع: الخروج ينقص لنفس النهاية الي يحسبها النظام
CREATE TEMP TABLE fix_att ON COMMIT DROP AS
SELECT id, att_counted_end("employeeId", "checkIn", "checkOut") AS fixed FROM "Attendance"
WHERE "checkOut" - "checkIn" > interval '16 hours' AND NOT "approvedOut";
DO $$ BEGIN
  IF (SELECT count(*) FROM fix_att) > 250 THEN
    RAISE EXCEPTION 'جلسات أكثر من المتوقع (%) — ما تغيّر شي', (SELECT count(*) FROM fix_att);
  END IF;
END $$;
INSERT INTO "DataFix1Backup" (kind, "rowId", old)
SELECT 'attendance', a.id, to_jsonb(a) FROM "Attendance" a WHERE a.id IN (SELECT id FROM fix_att);
UPDATE "Attendance" a SET "checkOut" = f.fixed FROM fix_att f WHERE a.id = f.id AND f.fixed < a."checkOut";

\echo '═══ النتيجة ═══'
SELECT kind AS "النوع", count(*) AS "انصلح" FROM "DataFix1Backup"
WHERE "fixedAt" >= now() - interval '1 minute' GROUP BY 1;

COMMIT;
