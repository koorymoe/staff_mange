#!/bin/bash
# ═══════════════════════════════════════════════════════════════════
# فحص حالة النسخ الاحتياطي — شغّله أي وقت تريد تطمّن
# ═══════════════════════════════════════════════════════════════════
#
# أخطر شي بالنسخ الاحتياطي مو إنك ما عندك نسخ — إنك تظن عندك نسخ
# وهي واقفة من شهرين. هذا السكربت يجاوب بصراحة:
#
#   • الجدولة التلقائية شغالة لو لا؟
#   • آخر نسخة متى صارت؟ وشكد حجمها؟
#   • جوّاها شنو فعلاً — قاعدة بيانات وملفات وإعدادات؟
#   • اكو نسخة خارج السيرفر لو كلها هنا؟
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKUP_DIR="$SCRIPT_DIR/backups"

echo "═══════════════════════════════════════════"
echo "   حالة النسخ الاحتياطي — نظام الأماني"
echo "═══════════════════════════════════════════"
echo

# ── الجدولة ──
echo "⏰ الجدولة التلقائية:"
if crontab -l 2>/dev/null | grep -q "backup-db.sh"; then
  echo "   ✅ مفعّلة — $(crontab -l 2>/dev/null | grep 'backup-db.sh' | awk '{print $1,$2,$3,$4,$5}')"
else
  echo "   ❌ مو مفعّلة! ماكو نسخ تلقائية تصير."
  echo "      شغّل: ./setup-backups.sh"
fi
echo

# ── آخر نسخة ──
echo "📦 النسخ المحفوظة:"
LATEST=$(ls -t "$BACKUP_DIR"/staffmange_*.tar.gz 2>/dev/null | head -1)
COUNT=$(ls "$BACKUP_DIR"/staffmange_*.tar.gz 2>/dev/null | wc -l)

if [ -z "$LATEST" ]; then
  echo "   ❌ ماكو ولا نسخة! شغّل: ./backup-db.sh"
  exit 1
fi

AGE_HOURS=$(( ( $(date +%s) - $(stat -c %Y "$LATEST") ) / 3600 ))
echo "   العدد: $COUNT"
echo "   الأحدث: $(basename "$LATEST") ($(du -h "$LATEST" | cut -f1))"

if [ "$AGE_HOURS" -lt 26 ]; then
  echo "   ✅ عمرها $AGE_HOURS ساعة — طازجة"
elif [ "$AGE_HOURS" -lt 72 ]; then
  echo "   ⚠️ عمرها $AGE_HOURS ساعة — متأخرة، تأكد من الجدولة"
else
  echo "   ❌ عمرها $((AGE_HOURS / 24)) يوم — النسخ واقفة! افحص الجدولة فوراً"
fi
echo

# ── محتوى النسخة ──
echo "🔍 محتوى آخر نسخة:"
CONTENTS=$(tar tzf "$LATEST" 2>/dev/null)
if [ -z "$CONTENTS" ]; then
  echo "   ❌ الأرشيف تالف وما ينفتح!"
  exit 1
fi

check_part() {
  if echo "$CONTENTS" | grep -q "$1"; then echo "   ✅ $2"; else echo "   ❌ $2 — ناقصة!"; fi
}
check_part "database.sql" "قاعدة البيانات (الحجوزات والزبائن والفواتير)"
check_part "uploads.tar"  "الملفات المرفوعة (الصور والوصولات والوثائق)"
check_part "env.txt"      "الإعدادات والأسرار (بدونها ما يشتغل النظام)"

# عدد الجداول جوّا النسخة — تأكيد إنها مو فاضية
TABLES=$(tar xzOf "$LATEST" ./database.sql 2>/dev/null | grep -c "^CREATE TABLE" || echo 0)
echo "   📊 عدد الجداول: $TABLES"
echo

# ── نسخة خارج السيرفر ──
echo "☁️  نسخة خارج السيرفر:"
if [ -f "$SCRIPT_DIR/.env" ]; then
  set -a; . "$SCRIPT_DIR/.env" 2>/dev/null; set +a
fi
if [ -n "${R2_BUCKET:-}" ] && command -v aws >/dev/null 2>&1; then
  REMOTE=$(AWS_ACCESS_KEY_ID="${R2_ACCESS_KEY:-}" AWS_SECRET_ACCESS_KEY="${R2_SECRET_KEY:-}" \
    aws s3 ls "s3://$R2_BUCKET/backups/" --endpoint-url "${R2_ENDPOINT:-}" 2>/dev/null | tail -3)
  if [ -n "$REMOTE" ]; then
    echo "   ✅ موجودة على R2 — آخرها:"
    echo "$REMOTE" | sed 's/^/      /'
  else
    echo "   ⚠️ R2 مضبوط بس ماكو نسخ مرفوعة"
  fi
else
  echo "   ❌ ماكو نسخة خارج السيرفر."
  echo "      كل نسخك على نفس السيرفر — لو ضاع السيرفر ضاع كل شي معاه."
  echo "      شغّل ./setup-backups.sh لتفعيلها."
fi
echo
echo "═══════════════════════════════════════════"
