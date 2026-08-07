#!/bin/bash
# ═══════════════════════════════════════════════════════════════════
# نسخة احتياطية شاملة — نظام شركة الأماني
# ═══════════════════════════════════════════════════════════════════
#
# ياخذ **كل شي يخص النظام**، مو قاعدة البيانات بس:
#
#   ١) قاعدة البيانات   — الحجوزات، الزبائن، الموظفين، الفواتير، كل شي
#   ٢) الملفات المرفوعة — صور المنتجات، الوصولات، وثائق السيارات
#   ٣) ملف .env         — كلمات السر ومفاتيح النظام
#   ٤) إعدادات النشر    — docker-compose.yml و Caddyfile
#   ٥) أدلة الاستخدام   — مجلد tutorials (المساعد الذكي يقراها)
#   ٦) شهادات HTTPS     — بيانات Caddy، حتى الموقع يشتغل فوراً بلا انتظار
#
# ليش كلها؟ لأن أي وحدة ناقصة تخلي الاسترجاع نصف استرجاع:
#   • بلا الملفات: النظام يشتغل بس كل صورة ووصل ووثيقة مكسورة.
#   • بلا .env: ما تكدر تشغّله أصلاً — كلمة سر قاعدة البيانات بيه.
#     وJWT_SECRET بيه، ولو تغيّر كل الموظفين ينطردون من جلساتهم.
#   • بلا إعدادات النشر: تقعد تعيد تركيب السيرفر بالتخمين.
#
# ⚠️ النسخة فيها كلمات سر وبيانات زبائن — تتشفّر إذا BACKUP_PASSPHRASE
#    مضبوطة بـ.env (مطلوبة قبل أي رفع لمكان خارجي).
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKUP_DIR="$SCRIPT_DIR/backups"
KEEP_DAYS="${BACKUP_KEEP_DAYS:-14}"
DB_CONTAINER="${DB_CONTAINER:-staff_mange-db-1}"
BACKEND_CONTAINER="${BACKEND_CONTAINER:-staff_mange-backend-1}"
CADDY_CONTAINER="${CADDY_CONTAINER:-staff_mange-caddy-1}"

[ -f "$SCRIPT_DIR/.env" ] && { set -a; . "$SCRIPT_DIR/.env" 2>/dev/null; set +a; }

mkdir -p "$BACKUP_DIR"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
START_EPOCH=$(date +%s)
WORK="$BACKUP_DIR/.work_$TIMESTAMP"
mkdir -p "$WORK"
trap 'rm -rf "$WORK"' EXIT

WARNINGS=()
warn() { echo "  ⚠️ $*"; WARNINGS+=("$*"); }

# ── قاعدة صارمة: كل شي يخص النسخ الاحتياطية للمالك وحده ──
# الإشعار يروح لدور OWNER بس، مو لـADMIN. المالك سلّم إدارة النظام
# وبقت مراقبة النسخ إله هو بس — فلو دزّينا الإشعار لـADMIN همّه كشفنا
# وجود النسخ وحالتها لشخص مو مفروض يشوفها.
psql_q() { docker exec "$DB_CONTAINER" psql -U staffmange -d staffmange -q -c "$1" >/dev/null 2>&1 || true; }

# sql_escape يحمي من الفواصل المفردة بنص الخطأ (اسم ملف أو رسالة فيها ')
# — بدونها أي ' تكسر جملة الـINSERT وتضيع التسجيلة بالسكوت.
sql_escape() { printf '%s' "$1" | sed "s/'/''/g"; }

# notify_owner يخلي فشل النسخة يوصل للمالك **داخل النظام نفسه**.
# بدونها، لو وقفت النسخ ما يعرف أحد إلا يوم الكارثة — وهذا بالضبط
# الي يخلي الناس تكتشف إن ماكو باكاب بأسوأ لحظة ممكنة.
notify_owner() {
  psql_q "
    INSERT INTO \"Notification\" (id, \"employeeId\", type, message)
    SELECT gen_random_uuid()::text, id, 'backup_failed', '$(sql_escape "$1")'
    FROM \"Employee\" WHERE role = 'OWNER' AND status = 'ACTIVE'
  "
}

# ── تسجيل نتيجة كل تشغيل بجدول BackupRun ──
# ليش؟ حاوية الباك إند ما تشوف مجلد backups/ على السيرفر. بدون هذا
# الصف ما اكو طريقة النظام يعرض للمالك حالة النسخ إلا بفتح SSH كل يوم.
# الجدول يقراه مسار /api/owner/backups المحمي بـRequireOwner.
STATUS_OK=false
STATUS_FILE=""
STATUS_SIZE=0
STATUS_TABLES=0
STATUS_ENCRYPTED=false
STATUS_OFFSITE=false
STATUS_TARGETS=""
STATUS_UPLOADS=false
STATUS_ENV=false
STATUS_KEPT=0
STATUS_ERROR=""

record_run() {
  local warns=""
  if [ ${#WARNINGS[@]} -gt 0 ]; then
    warns=$(printf '%s؛ ' "${WARNINGS[@]}")
  fi
  psql_q "
    INSERT INTO \"BackupRun\" (
      id, \"startedAt\", \"finishedAt\", ok, \"fileName\", \"sizeBytes\",
      \"tableCount\", encrypted, offsite, \"offsiteTarget\", \"hasUploads\",
      \"hasEnv\", warnings, error, \"keptCount\"
    ) VALUES (
      gen_random_uuid()::text, NOW() - make_interval(secs => $(( $(date +%s) - START_EPOCH ))), NOW(),
      $STATUS_OK, NULLIF('$(sql_escape "$STATUS_FILE")',''), $STATUS_SIZE,
      $STATUS_TABLES, $STATUS_ENCRYPTED, $STATUS_OFFSITE,
      NULLIF('$(sql_escape "$STATUS_TARGETS")',''), $STATUS_UPLOADS,
      $STATUS_ENV, NULLIF('$(sql_escape "$warns")',''),
      NULLIF('$(sql_escape "$STATUS_ERROR")',''), $STATUS_KEPT
    )
  "
}

fail() {
  echo "❌ $*" >&2
  STATUS_ERROR="$*"
  record_run
  notify_owner "🔴 فشلت النسخة الاحتياطية اليوم: $* — راجع السيرفر فوراً"
  exit 1
}

echo "═══ نسخة احتياطية شاملة: $TIMESTAMP ═══"

# ── ١) قاعدة البيانات ──
docker ps --format '{{.Names}}' | grep -qx "$DB_CONTAINER" \
  || fail "حاوية قاعدة البيانات «$DB_CONTAINER» مو شغالة"

echo "→ قاعدة البيانات..."
docker exec "$DB_CONTAINER" pg_dump -U staffmange --clean --if-exists staffmange \
  > "$WORK/database.sql" 2>/dev/null || fail "فشل تصدير قاعدة البيانات"

# pg_dump ينهي ملفه بسطر معروف. بدون هذا الفحص ممكن ننام مرتاحين على
# نسخة مقطوعة بالنص ونكتشفها بس يوم الكارثة.
tail -5 "$WORK/database.sql" | grep -q "PostgreSQL database dump complete" \
  || fail "ملف قاعدة البيانات مقطوع — النسخة مرفوضة"
TABLES=$(grep -c "^CREATE TABLE" "$WORK/database.sql" || true)
[ "$TABLES" -ge 20 ] || fail "عدد الجداول $TABLES — قليل جداً، في خطأ"
echo "  ✓ $TABLES جدول ($(du -h "$WORK/database.sql" | cut -f1))"
STATUS_TABLES=$TABLES

# ── ٢) الملفات المرفوعة ──
echo "→ الملفات المرفوعة..."
if docker ps --format '{{.Names}}' | grep -qx "$BACKEND_CONTAINER"; then
  docker exec "$BACKEND_CONTAINER" sh -c 'cd /app/uploads 2>/dev/null && tar cf - . 2>/dev/null' \
    > "$WORK/uploads.tar" 2>/dev/null
  if [ -s "$WORK/uploads.tar" ]; then
    echo "  ✓ $(du -h "$WORK/uploads.tar" | cut -f1)"
    STATUS_UPLOADS=true
  else
    rm -f "$WORK/uploads.tar"
    echo "  — فاضي (طبيعي إذا التخزين على R2)"
  fi
else
  warn "حاوية الباك إند مو شغالة — الملفات المرفوعة ما انأخذت"
fi

# ── ٣) الأسرار ── ٤) إعدادات النشر ── ٥) الأدلة ──
echo "→ الإعدادات والأسرار..."
if [ -f "$SCRIPT_DIR/.env" ]; then
  cp "$SCRIPT_DIR/.env" "$WORK/env.txt"; echo "  ✓ .env"; STATUS_ENV=true
else
  warn "ماكو .env — بدونه ما تكدر تشغّل النظام من النسخة!"
fi
for f in docker-compose.yml Caddyfile; do
  [ -f "$SCRIPT_DIR/$f" ] && cp "$SCRIPT_DIR/$f" "$WORK/$f" && echo "  ✓ $f"
done
if [ -d "$SCRIPT_DIR/tutorials" ]; then
  tar cf "$WORK/tutorials.tar" -C "$SCRIPT_DIR" tutorials 2>/dev/null && echo "  ✓ أدلة الاستخدام"
fi

# ── ٦) شهادات HTTPS ──
# بدونها الموقع يشتغل، بس Caddy يطلب شهادة جديدة — وإذا وصلت حد
# الطلبات عند Let's Encrypt يقعد الموقع بلا HTTPS ساعات.
if docker ps --format '{{.Names}}' | grep -qx "$CADDY_CONTAINER"; then
  docker exec "$CADDY_CONTAINER" sh -c 'cd /data 2>/dev/null && tar cf - . 2>/dev/null' \
    > "$WORK/caddy_data.tar" 2>/dev/null
  [ -s "$WORK/caddy_data.tar" ] && echo "  ✓ شهادات HTTPS" || rm -f "$WORK/caddy_data.tar"
fi

# ── ورقة تعليمات جوّا النسخة ──
# لو فتحتها بعد سنة أو فتحها شخص ثاني، لازم يعرف شنو يسوي بيها.
cat > "$WORK/README.txt" <<INNER
نسخة احتياطية — نظام شركة الأماني
التاريخ: $(date '+%Y-%m-%d %H:%M')
الجداول: $TABLES

المحتويات:
  database.sql    قاعدة البيانات كاملة
  uploads.tar     الملفات المرفوعة (صور، وصولات، وثائق)
  env.txt         كلمات السر والمفاتيح  ← بدونه النظام ما يشتغل
  docker-compose.yml, Caddyfile   إعدادات النشر
  tutorials.tar   أدلة المساعد الذكي
  caddy_data.tar  شهادات HTTPS

للاسترجاع:
  cd ~/staff_mange && ./restore-backup.sh <هذا-الملف>

للاسترجاع على سيرفر جديد: اقرأ EMERGENCY.md بالمستودع.
INNER

# ── التغليف ──
ARCHIVE="$BACKUP_DIR/staffmange_$TIMESTAMP.tar.gz"
tar czf "$ARCHIVE" -C "$WORK" . || fail "فشل ضغط النسخة"
tar tzf "$ARCHIVE" >/dev/null 2>&1 || fail "الأرشيف تالف — النسخة مرفوضة"

# ── التشفير ──
# النسخة فيها كلمات سر وأرقام هواتف زبائن. رفعها لمكان خارجي بلا
# تشفير معناه إن أي واحد يوصل لذاك المكان يوصل لكل شي.
FINAL="$ARCHIVE"
if [ -n "${BACKUP_PASSPHRASE:-}" ] && command -v gpg >/dev/null 2>&1; then
  echo "→ التشفير..."
  if gpg --batch --yes --symmetric --cipher-algo AES256 \
       --passphrase "$BACKUP_PASSPHRASE" -o "$ARCHIVE.gpg" "$ARCHIVE" 2>/dev/null; then
    rm -f "$ARCHIVE"; FINAL="$ARCHIVE.gpg"
    STATUS_ENCRYPTED=true
    echo "  ✓ مشفّرة"
  else
    warn "فشل التشفير — النسخة غير مشفّرة"
  fi
fi

echo "✅ النسخة جاهزة: $(basename "$FINAL") ($(du -h "$FINAL" | cut -f1))"
STATUS_OK=true
STATUS_FILE="$(basename "$FINAL")"
STATUS_SIZE=$(stat -c %s "$FINAL" 2>/dev/null || echo 0)

# ═══ النسخة خارج السيرفر ═══
# نسخة على نفس السيرفر ما تحميك من ضياع السيرفر. ندعم ثلاث وجهات،
# ويكفي وحدة تنجح.
OFFSITE_OK=0

# وجهة ١: R2 / أي تخزين متوافق مع S3
if [ -n "${R2_BUCKET:-}" ] && [ -n "${R2_ACCESS_KEY:-}" ]; then
  if command -v aws >/dev/null 2>&1; then
    echo "→ الرفع لـR2..."
    if AWS_ACCESS_KEY_ID="$R2_ACCESS_KEY" AWS_SECRET_ACCESS_KEY="$R2_SECRET_KEY" \
       aws s3 cp "$FINAL" "s3://$R2_BUCKET/backups/$(basename "$FINAL")" \
       --endpoint-url "$R2_ENDPOINT" >/dev/null 2>&1; then
      echo "  ✓ انرفعت لـR2"; OFFSITE_OK=1; STATUS_OFFSITE=true; STATUS_TARGETS="${STATUS_TARGETS}R2 "
    else
      warn "فشل الرفع لـR2"
    fi
  else
    warn "R2 مضبوط بس أداة aws مو منصّبة (apt-get install -y awscli)"
  fi
fi

# وجهة ٢: سيرفر ثاني عبر SSH — مفيدة إذا عندك سيرفر أو جهاز بالمكتب
if [ -n "${BACKUP_SSH_TARGET:-}" ]; then
  echo "→ النسخ لسيرفر ثاني..."
  if scp -o StrictHostKeyChecking=accept-new -q "$FINAL" "$BACKUP_SSH_TARGET" 2>/dev/null; then
    echo "  ✓ انرفعت لـ$BACKUP_SSH_TARGET"; OFFSITE_OK=1; STATUS_OFFSITE=true; STATUS_TARGETS="${STATUS_TARGETS}$BACKUP_SSH_TARGET "
  else
    warn "فشل النسخ لـ$BACKUP_SSH_TARGET"
  fi
fi

# وجهة ٣: قرص/مسار ثاني على نفس الجهاز — أضعف حماية بس أحسن من ولا شي
if [ -n "${BACKUP_SECOND_DIR:-}" ] && [ -d "${BACKUP_SECOND_DIR}" ]; then
  cp "$FINAL" "$BACKUP_SECOND_DIR/" 2>/dev/null \
    && { echo "  ✓ انسخت لـ$BACKUP_SECOND_DIR"; OFFSITE_OK=1; STATUS_OFFSITE=true; STATUS_TARGETS="${STATUS_TARGETS}$BACKUP_SECOND_DIR "; } \
    || warn "فشل النسخ لـ$BACKUP_SECOND_DIR"
fi

if [ "$OFFSITE_OK" -eq 0 ]; then
  warn "ماكو نسخة خارج السيرفر! لو ضاع السيرفر تروح كل النسخ معاه"
  notify_owner "⚠️ النسخة الاحتياطية اليوم محفوظة على السيرفر بس — ماكو نسخة بمكان ثاني. شغّل ./setup-backups.sh"
fi

# ── تنظيف القديم ──
find "$BACKUP_DIR" -name "staffmange_*.tar.gz*" -mtime +"$KEEP_DAYS" -delete 2>/dev/null
find "$BACKUP_DIR" -name "staffmange_*.sql.gz" -mtime +"$KEEP_DAYS" -delete 2>/dev/null
COUNT=$(find "$BACKUP_DIR" -name "staffmange_*.tar.gz*" | wc -l)
STATUS_KEPT=$COUNT

echo "📦 النسخ المحفوظة: $COUNT (نحتفظ بـ$KEEP_DAYS يوم)"
if [ ${#WARNINGS[@]} -gt 0 ]; then
  echo
  echo "⚠️ تحذيرات (${#WARNINGS[@]}):"
  printf '   • %s\n' "${WARNINGS[@]}"
fi

# ── تسجيل النتيجة بالنظام (يقراها المالك وحده) ──
record_run
