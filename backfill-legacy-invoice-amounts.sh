#!/bin/bash
# ══════════════════════════════════════════════════════════════════════════
# تصحيح رجعي: مزامنة حجوزات الفواتير التاريخية المستوردة سابقاً
# ══════════════════════════════════════════════════════════════════════════
#
# 🔴 السبب: import-legacy-invoices.sh (قبل هذا التصحيح) كان يكتب فاتورة
# بجدول LeaderInvoice بس، وما يلمس حجزها. شاشة «تدقيق الحسابات» ما تعرف
# شي عن LeaderInvoice أبداً — تحكم بس من Booking.amountCollected/
# amountVerified. فالفاتورة تدخل، والحجز يبقى يقول «ماكو مبلغ، ماكو
# تدقيق»، فيطلع بطابور «بانتظار التدقيق» بخانة فاضية، والمحاسب مطالب
# يعيد كتابة مبلغ موجود أصلاً. وإجمالي الإيرادات ما يعكس هالمبالغ.
#
# هذا السكربت يصحّح كل الفواتير الي استوردناها **قبل** هذا التصحيح —
# مرة وحدة، رجعياً. الاستيرادات الجاية تتصحح تلقائياً (نفس المنطق
# انضاف لـ import-legacy-invoices.sh نفسه).
#
# ⚠️ دفتر تدقيق الحسابات (الإكسل) هو نفسه مصدر التدقيق — طلب صاحب
# النظام صراحة: «أي فلوس مربوطة بحجز بزبون اعتبرها مدققة»، بغض النظر
# عن وجود رقم فاتورة خارجي (APPROVED أو SUBMITTED سوا).
#
# ⚠️ ما نلمس حجزاً دقّقه إنسان مسبقاً بمبلغ حقيقي (amountCollected غير
# صفري) — رقمه هو الصحيح ونحترمه.
#
# ⚠️ الترشيح **بعلامة الاستيراد** (manualPriceNote) حتى ما نلمس أي
# فاتورة مو من سكربتنا — لو محاسب سوّى فاتورة يدوية بمبلغ صفري لسبب
# مختلف، هذا السكربت ما يمسها.
#
# الاستخدام:
#   ./backfill-legacy-invoice-amounts.sh          # تجربة (ما يحفظ)
#   ./backfill-legacy-invoice-amounts.sh --apply  # حفظ فعلي
# ══════════════════════════════════════════════════════════════════════════
set -euo pipefail

MODE="${1:-}"
DB_CONTAINER="${DB_CONTAINER:-staff_mange-db-1}"

if [ "$MODE" = "--apply" ]; then
  FINISH="COMMIT;"
  echo "═══ وضع الحفظ الفعلي (--apply) — التغييرات راح تنحفظ ═══"
else
  FINISH="ROLLBACK;"
  echo "═══ وضع التجربة — ولا تغيير راح ينحفظ (زيد --apply للحفظ) ═══"
fi
echo

docker exec -i "$DB_CONTAINER" psql -U staffmange -d staffmange -v ON_ERROR_STOP=1 <<SQL
BEGIN;

\echo ''
\echo '───────────── قبل التصحيح ─────────────'
\echo ''
\echo '» فواتير تاريخية مستوردة، وكم منها بعده حجزها بلا مبلغ/تدقيق:'
SELECT
  count(*)                                                      AS "فواتير تاريخية مستوردة",
  count(*) FILTER (WHERE NOT b."amountVerified")                AS "حجزها بعده بانتظار التدقيق",
  COALESCE(sum(li."netTotal") FILTER (WHERE NOT b."amountVerified"), 0)
                                                                 AS "مبلغها (راح يضاف للإيرادات)"
FROM "LeaderInvoice" li
JOIN "Booking" b ON b.id = li."bookingId"
WHERE li."manualPriceNote" = 'تسوية محاسبية لشغل قبل النظام — المصدر: دفتر تدقيق الحسابات';

-- ═══ التصحيح ═══
-- ⚠️ RETURNING يمسك **بالضبط** الصفوف الي هذا التشغيل لمسها — تقرير
-- دقيق بدل ما نحسب على كل الفواتير التاريخية (بعضها ممكن يكون
-- مدقق من زمان بمبلغ حقيقي، وحسابه ضمن «تزامن هذا التشغيل» يكذب).
CREATE TEMP TABLE fixed_rows AS
WITH upd AS (
  UPDATE "Booking" b
  SET "amountCollected" = li."netTotal",
      "amountVerified"  = true
  FROM "LeaderInvoice" li
  WHERE b.id = li."bookingId"
    AND li."manualPriceNote" = 'تسوية محاسبية لشغل قبل النظام — المصدر: دفتر تدقيق الحسابات'
    AND (b."amountCollected" IS NULL OR b."amountCollected" = 0)
    AND b."amountVerified" = false
  RETURNING b.id, b."amountCollected"
)
SELECT * FROM upd;

\echo ''
\echo '───────────── بعد التصحيح ─────────────'
\echo ''
\echo '» شكد حجز صار مدقق تلقائياً هالتشغيلة، وشكد زاد إجمالي الإيرادات:'
SELECT count(*) AS "حجوزات انصلحت الآن", COALESCE(sum("amountCollected"), 0) AS "زيادة إجمالي الإيرادات"
  FROM fixed_rows;

\echo ''
\echo '» نفس فحص «قبل التصحيح» — لازم يطلع صفر «بانتظار التدقيق»:'
SELECT
  count(*)                                       AS "فواتير تاريخية مستوردة",
  count(*) FILTER (WHERE NOT b."amountVerified") AS "حجزها بعده بانتظار التدقيق"
FROM "LeaderInvoice" li
JOIN "Booking" b ON b.id = li."bookingId"
WHERE li."manualPriceNote" = 'تسوية محاسبية لشغل قبل النظام — المصدر: دفتر تدقيق الحسابات';

$FINISH
SQL

echo
if [ "$MODE" = "--apply" ]; then
  echo "═══ انحفظ ═══"
else
  echo "═══ تجربة بس — ولا شي انحفظ. للحفظ: $0 --apply ═══"
fi
