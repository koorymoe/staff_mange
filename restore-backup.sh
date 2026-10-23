#!/bin/bash
# ═══════════════════════════════════════════════════════════════════
# استرجاع النظام من نسخة احتياطية
# ═══════════════════════════════════════════════════════════════════
#
#   ./restore-backup.sh                  → آخر نسخة
#   ./restore-backup.sh ملف.tar.gz       → نسخة محددة
#   ./restore-backup.sh ملف.tar.gz.gpg   → نسخة مشفّرة
#
# ⚠️ يستبدل البيانات الحالية. يسأل قبل ما ينفّذ، وياخذ نسخة أمان من
# الوضع الحالي أول — حتى لو استرجعت النسخة الغلط تكدر ترجع.
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKUP_DIR="$SCRIPT_DIR/backups"
DB_CONTAINER="${DB_CONTAINER:-staff_mange-db-1}"
BACKEND_CONTAINER="${BACKEND_CONTAINER:-staff_mange-backend-1}"
CADDY_CONTAINER="${CADDY_CONTAINER:-staff_mange-caddy-1}"

[ -f "$SCRIPT_DIR/.env" ] && { set -a; . "$SCRIPT_DIR/.env" 2>/dev/null; set +a; }
fail() { echo "❌ $*" >&2; exit 1; }

ARCHIVE="${1:-}"
if [ -z "$ARCHIVE" ]; then
  ARCHIVE=$(ls -t "$BACKUP_DIR"/staffmange_*.tar.gz* 2>/dev/null | head -1)
  [ -n "$ARCHIVE" ] || fail "ماكو نسخ بمجلد backups/"
fi
[ -f "$ARCHIVE" ] || fail "الملف مو موجود: $ARCHIVE"

echo "═══ استرجاع من: $(basename "$ARCHIVE") ═══"
echo "   الحجم: $(du -h "$ARCHIVE" | cut -f1)"
echo "   التاريخ: $(date -r "$ARCHIVE" '+%Y-%m-%d %H:%M')"
echo

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

# ── فك التشفير إذا لازم ──
SRC="$ARCHIVE"
if [[ "$ARCHIVE" == *.gpg ]]; then
  command -v gpg >/dev/null 2>&1 || fail "النسخة مشفّرة وأداة gpg مو منصّبة"
  PASS="${BACKUP_PASSPHRASE:-}"
  if [ -z "$PASS" ]; then
    read -rsp "كلمة سر النسخة: " PASS; echo
  fi
  echo "→ فك التشفير..."
  gpg --batch --yes --quiet --passphrase "$PASS" -o "$WORK/archive.tar.gz" -d "$ARCHIVE" 2>/dev/null \
    || fail "كلمة السر غلط أو الملف تالف"
  SRC="$WORK/archive.tar.gz"
fi

tar xzf "$SRC" -C "$WORK" || fail "الأرشيف تالف وما ينفتح"
rm -f "$WORK/archive.tar.gz"

echo "جوّا النسخة:"
[ -f "$WORK/database.sql" ]  && echo "   ✓ قاعدة البيانات ($(grep -c '^CREATE TABLE' "$WORK/database.sql" || echo 0) جدول)"
[ -f "$WORK/uploads.tar" ]   && echo "   ✓ الملفات المرفوعة ($(du -h "$WORK/uploads.tar" | cut -f1))"
[ -f "$WORK/env.txt" ]       && echo "   ✓ الإعدادات والأسرار"
[ -f "$WORK/tutorials.tar" ] && echo "   ✓ أدلة الاستخدام"
[ -f "$WORK/caddy_data.tar" ] && echo "   ✓ شهادات HTTPS"
echo

[ -f "$WORK/database.sql" ] || fail "النسخة ماكو بيها قاعدة بيانات"

echo "⚠️  هذا راح يستبدل كل البيانات الحالية بالنسخة أعلاه."
read -rp "متأكد؟ اكتب «نعم» للمتابعة: " CONFIRM
[ "$CONFIRM" = "نعم" ] || { echo "انلغى."; exit 0; }

docker ps --format '{{.Names}}' | grep -qx "$DB_CONTAINER" \
  || fail "حاوية قاعدة البيانات مو شغالة — شغّل: docker compose up -d"

# ── نسخة أمان من الوضع الحالي ──
# أهم خطوة بالسكربت. لو استرجعت النسخة الغلط، هاي طوق النجاة.
mkdir -p "$BACKUP_DIR"
SAFETY="$BACKUP_DIR/before_restore_$(date +%Y%m%d_%H%M%S).sql.gz"
echo "→ نسخة أمان من الوضع الحالي..."
docker exec "$DB_CONTAINER" pg_dump -U staffmange staffmange | gzip > "$SAFETY"
echo "  ✓ $(basename "$SAFETY")"

# ── قاعدة البيانات ──
echo "→ استرجاع قاعدة البيانات..."
docker exec -i "$DB_CONTAINER" psql -U staffmange -d staffmange -q < "$WORK/database.sql" \
  >/dev/null 2>&1 || fail "فشل الاسترجاع — بياناتك القديمة محفوظة بـ$(basename "$SAFETY")"
echo "  ✓ تمت"

# ── الملفات المرفوعة ──
if [ -f "$WORK/uploads.tar" ] && docker ps --format '{{.Names}}' | grep -qx "$BACKEND_CONTAINER"; then
  echo "→ الملفات المرفوعة..."
  docker exec -i "$BACKEND_CONTAINER" sh -c 'mkdir -p /app/uploads && tar xf - -C /app/uploads' \
    < "$WORK/uploads.tar" 2>/dev/null && echo "  ✓ تمت" || echo "  ⚠️ فشل"
fi

# ── شهادات HTTPS ──
if [ -f "$WORK/caddy_data.tar" ] && docker ps --format '{{.Names}}' | grep -qx "$CADDY_CONTAINER"; then
  echo "→ شهادات HTTPS..."
  docker exec -i "$CADDY_CONTAINER" sh -c 'tar xf - -C /data' < "$WORK/caddy_data.tar" 2>/dev/null \
    && echo "  ✓ تمت" || echo "  ⚠️ فشل (Caddy راح يطلب شهادة جديدة، مو مشكلة)"
fi

# ── أدلة الاستخدام ──
if [ -f "$WORK/tutorials.tar" ]; then
  tar xf "$WORK/tutorials.tar" -C "$SCRIPT_DIR" 2>/dev/null && echo "→ أدلة الاستخدام ✓"
fi

# ── الإعدادات ──
# ما ندوس على .env تلقائياً: لو السيرفر جديد وكلمات سره مختلفة، الدوس
# عليها يكسر الاتصال بقاعدة البيانات. نحطها جنبه ونخلي القرار للمستخدم.
if [ -f "$WORK/env.txt" ]; then
  cp "$WORK/env.txt" "$SCRIPT_DIR/.env.from-backup"
  chmod 600 "$SCRIPT_DIR/.env.from-backup"
  echo "→ الإعدادات انحفظت بـ.env.from-backup"
  if [ -f "$SCRIPT_DIR/.env" ]; then
    echo "  ⚠️ ما دسنا على .env الحالي. قارنهم:  diff .env .env.from-backup"
  else
    cp "$SCRIPT_DIR/.env.from-backup" "$SCRIPT_DIR/.env"
    chmod 600 "$SCRIPT_DIR/.env"
    echo "  ✓ ماكو .env فصار هو الأساس"
  fi
fi

# ── إعدادات النشر ──
for f in docker-compose.yml Caddyfile; do
  [ -f "$WORK/$f" ] && [ ! -f "$SCRIPT_DIR/$f" ] && cp "$WORK/$f" "$SCRIPT_DIR/$f" && echo "→ $f ✓"
done

echo
echo "✅ خلص الاسترجاع."
echo "   أعد التشغيل:  docker compose restart backend"
echo "   نسخة الأمان:  $(basename "$SAFETY")"
