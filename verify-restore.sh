#!/bin/bash
# ═══ 🧪 تجربة استرجاع النسخة الاحتياطية (دازاي 11-02: «شبكة الأمان» — تعديلات ٦ب) ═══
#
# النسخة الي ما انجرّب استرجاعها مو نسخة — هي أمل. هذا السكربت ياخذ ملف
# database.sql، يسترجعه بقاعدة مؤقتة جوّا نفس حاوية postgres، ويقارن:
#   • عدد الجداول بالمسترجعة = عدد الجداول بالحية.
#   • عدد صفوف **كل** جدول: المسترجعة ما تقل عن الحية (ناقص هامش
#     صغير للي انضاف بين لحظة التصدير ولحظة العد).
# بعدها يمسح القاعدة المؤقتة. ما يلمس القاعدة الحقيقية أبداً.
#
# الاستعمال:   ./verify-restore.sh <database.sql>
# يطبع آخر سطر:  VERIFY_OK=<true|false> VERIFY_ROWS=<n> VERIFY_NOTE=<نص>
# بالسيرفر يشتغل جوّا حاوية DB_CONTAINER. محلياً (للاختبار): LOCAL_PSQL=1.
set -uo pipefail
SQL="${1:?ملف database.sql}"
DB_CONTAINER="${DB_CONTAINER:-staff_mange-db-1}"
DB_USER="${DB_USER:-staffmange}"
LIVE_DB="${LIVE_DB:-staffmange}"
VDB="${VERIFY_DB:-staffmange_restore_check}"

if [ -n "${LOCAL_PSQL:-}" ]; then
  P() { psql "$@"; }                       # اختبار محلي: PGHOST/PGPASSWORD من البيئة
  Pi() { psql "$@"; }
else
  P() { docker exec "$DB_CONTAINER" psql "$@"; }
  Pi() { docker exec -i "$DB_CONTAINER" psql "$@"; }
fi
q() { P -U "$DB_USER" -d "$1" -Atqc "$2" 2>/dev/null; }

result() { echo "VERIFY_OK=$1 VERIFY_ROWS=${2:-0} VERIFY_NOTE=$3"; }
cleanup() { P -U "$DB_USER" -d "$LIVE_DB" -qc "DROP DATABASE IF EXISTS \"$VDB\"" >/dev/null 2>&1; }
trap cleanup EXIT

cleanup
P -U "$DB_USER" -d "$LIVE_DB" -qc "CREATE DATABASE \"$VDB\"" >/dev/null 2>&1 \
  || { result false 0 "ما گدرت أسوي قاعدة مؤقتة للتجربة"; exit 1; }

ERRS=$(Pi -U "$DB_USER" -d "$VDB" -q -v ON_ERROR_STOP=0 < "$SQL" 2>&1 >/dev/null | grep -c "ERROR" || true)

LIVE_T=$(q "$LIVE_DB" "SELECT count(*) FROM information_schema.tables WHERE table_schema='public'")
REST_T=$(q "$VDB" "SELECT count(*) FROM information_schema.tables WHERE table_schema='public'")
if [ -z "$REST_T" ] || [ "$REST_T" -lt "${LIVE_T:-1}" ]; then
  result false 0 "الجداول المسترجعة ${REST_T:-0} من ${LIVE_T:-?}"; exit 1
fi

# عدد صفوف **كل** جدول بالطرفين (نسخة مقطوعة بالنص تخسر جداول آخر الملف)
COUNTS_SQL="SELECT table_name || '|' || (xpath('/row/c/text()', query_to_xml(format('SELECT count(*) AS c FROM public.%I', table_name), false, true, '')))[1]::text
  FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE' ORDER BY 1"
q "$LIVE_DB" "$COUNTS_SQL" > "/tmp/.vr_live.$$"
q "$VDB" "$COUNTS_SQL" > "/tmp/.vr_rest.$$"
TOTAL=0; BAD=""; NBAD=0
while IFS='|' read -r t L; do
  R=$(grep -m1 "^$t|" "/tmp/.vr_rest.$$" | cut -d'|' -f2); R=${R:-0}
  TOTAL=$((TOTAL + R))
  # هامش ١٪ + ٥ صفوف: الي انضاف بين التصدير والعد
  MIN=$(( L - L / 100 - 5 ))
  if [ "$R" -lt "$MIN" ]; then NBAD=$((NBAD + 1)); [ "$NBAD" -le 5 ] && BAD="$BAD $t($R/$L)"; fi
done < "/tmp/.vr_live.$$"
rm -f "/tmp/.vr_live.$$" "/tmp/.vr_rest.$$"
[ "$NBAD" -gt 5 ] && BAD="$BAD و$((NBAD - 5)) جدول ثاني"

if [ -n "$BAD" ]; then
  result false "$TOTAL" "صفوف ناقصة بعد الاسترجاع:$BAD"; exit 1
fi
NOTE="انسترجعت $REST_T جدول و$TOTAL صف — كل جدول طابق"
[ "$ERRS" -gt 0 ] && NOTE="$NOTE (⚠️ $ERRS رسالة خطأ أثناء الاسترجاع)"
result true "$TOTAL" "$NOTE"
