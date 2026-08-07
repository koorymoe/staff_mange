#!/bin/bash
# ═══════════════════════════════════════════════════════════════════
# استرجاع النظام من نسخة احتياطية
# ═══════════════════════════════════════════════════════════════════
#
#   ./restore-backup.sh                          → يسترجع آخر نسخة
#   ./restore-backup.sh backups/staffmange_X.tar.gz  → نسخة محددة
#
# ⚠️ يستبدل البيانات الحالية بالكامل. يسأل قبل ما ينفّذ، وياخذ نسخة
# أمان من الوضع الحالي أول — حتى لو استرجعت النسخة الغلط تكدر ترجع.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKUP_DIR="$SCRIPT_DIR/backups"
DB_CONTAINER="${DB_CONTAINER:-staff_mange-db-1}"
BACKEND_CONTAINER="${BACKEND_CONTAINER:-staff_mange-backend-1}"

fail() { echo "❌ $*" >&2; exit 1; }

ARCHIVE="${1:-}"
if [ -z "$ARCHIVE" ]; then
  ARCHIVE=$(ls -t "$BACKUP_DIR"/staffmange_*.tar.gz 2>/dev/null | head -1)
  [ -n "$ARCHIVE" ] || fail "ماكو نسخ بمجلد backups/"
fi
[ -f "$ARCHIVE" ] || fail "الملف مو موجود: $ARCHIVE"

echo "═══ استرجاع من: $(basename "$ARCHIVE") ═══"
echo "   الحجم: $(du -h "$ARCHIVE" | cut -f1)"
echo "   التاريخ: $(date -r "$ARCHIVE" '+%Y-%m-%d %H:%M')"
echo

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT
tar xzf "$ARCHIVE" -C "$WORK" || fail "الأرشيف تالف وما ينفتح"

echo "جوّا النسخة:"
[ -f "$WORK/database.sql" ] && echo "   ✓ قاعدة البيانات ($(grep -c '^CREATE TABLE' "$WORK/database.sql" || echo 0) جدول)"
[ -f "$WORK/uploads.tar" ] && echo "   ✓ الملفات المرفوعة ($(du -h "$WORK/uploads.tar" | cut -f1))"
[ -f "$WORK/env.txt" ] && echo "   ✓ ملف الإعدادات"
echo

[ -f "$WORK/database.sql" ] || fail "النسخة ماكو بيها قاعدة بيانات"

echo "⚠️  هذا راح يستبدل كل البيانات الحالية بالنسخة أعلاه."
read -rp "متأكد؟ اكتب «نعم» للمتابعة: " CONFIRM
[ "$CONFIRM" = "نعم" ] || { echo "انلغى."; exit 0; }

docker ps --format '{{.Names}}' | grep -qx "$DB_CONTAINER" \
  || fail "حاوية قاعدة البيانات مو شغالة — شغّل: docker compose up -d"

# ── نسخة أمان من الوضع الحالي ──
# لو استرجعت النسخة الغلط، هاي طوق النجاة. أهم خطوة بالسكربت كله.
SAFETY="$BACKUP_DIR/before_restore_$(date +%Y%m%d_%H%M%S).sql.gz"
echo "→ نسخة أمان من الوضع الحالي..."
docker exec "$DB_CONTAINER" pg_dump -U staffmange staffmange | gzip > "$SAFETY"
echo "  ✓ $(basename "$SAFETY")"

# ── قاعدة البيانات ──
echo "→ استرجاع قاعدة البيانات..."
# --clean --if-exists موجودة بالنسخة، فالجداول القديمة تنشال أول
docker exec -i "$DB_CONTAINER" psql -U staffmange -d staffmange -q < "$WORK/database.sql" \
  || fail "فشل الاسترجاع — بياناتك القديمة محفوظة بـ$(basename "$SAFETY")"
echo "  ✓ تمت"

# ── الملفات المرفوعة ──
if [ -f "$WORK/uploads.tar" ] && docker ps --format '{{.Names}}' | grep -qx "$BACKEND_CONTAINER"; then
  echo "→ استرجاع الملفات المرفوعة..."
  docker exec -i "$BACKEND_CONTAINER" sh -c 'mkdir -p /app/uploads && tar xf - -C /app/uploads' < "$WORK/uploads.tar" \
    && echo "  ✓ تمت" || echo "  ⚠️ فشل استرجاع الملفات"
fi

# ── الإعدادات ──
# ما ندوس على .env تلقائياً: لو السيرفر جديد وكلمات سره مختلفة، الدوس
# عليها يكسر الاتصال بقاعدة البيانات. نحطها جنبه ونخلي القرار للمستخدم.
if [ -f "$WORK/env.txt" ]; then
  cp "$WORK/env.txt" "$SCRIPT_DIR/.env.from-backup"
  echo "→ الإعدادات انحفظت بـ.env.from-backup"
  echo "  ⚠️ ما دسنا على .env الحالي. قارنهم وانسخ الي تحتاجه:"
  echo "     diff .env .env.from-backup"
fi

echo
echo "✅ خلص الاسترجاع."
echo "   أعد تشغيل النظام: docker compose restart backend"
echo "   نسخة الأمان (لو تريد ترجع): $(basename "$SAFETY")"
