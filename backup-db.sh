#!/bin/bash
# ═══════════════════════════════════════════════════════════════════
# نسخة احتياطية كاملة لنظام شركة الأماني
# ═══════════════════════════════════════════════════════════════════
#
# ياخذ **ثلاث** أشياء، مو قاعدة البيانات بس:
#
#   ١) قاعدة البيانات   — كل الحجوزات والزبائن والموظفين والفواتير
#   ٢) الملفات المرفوعة — صور المنتجات، الوصولات، وثائق السيارات
#   ٣) ملف .env         — كلمات السر ومفاتيح النظام
#
# ليش الثلاثة سوه؟ لأن أي وحدة ناقصة تخلي الاسترجاع نصف استرجاع:
#   • بلا الملفات: النظام يشتغل بس كل صورة ووصل ووثيقة تطلع مكسورة.
#   • بلا .env: ما تكدر تشغّل النظام أصلاً — كلمة سر قاعدة البيانات
#     بيه، ولو ضاعت ما تفتح النسخة الي خزنتها. وJWT_SECRET بيه، ولو
#     تغيّر كل الموظفين ينطردون من جلساتهم.
#
# النسخة تنحفظ محلياً، وتنرفع لـCloudflare R2 إذا كان مضبوط —
# ⚠️ نسخة محفوظة على نفس السيرفر ما تحميك من ضياع السيرفر نفسه.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKUP_DIR="$SCRIPT_DIR/backups"
KEEP_DAYS=14
DB_CONTAINER="${DB_CONTAINER:-staff_mange-db-1}"
BACKEND_CONTAINER="${BACKEND_CONTAINER:-staff_mange-backend-1}"

mkdir -p "$BACKUP_DIR"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
WORK="$BACKUP_DIR/.work_$TIMESTAMP"
mkdir -p "$WORK"
# لو وكع السكربت بالنص، ما نخلي مجلد شغل نص-كامل يتراكم
trap 'rm -rf "$WORK"' EXIT

fail() { echo "❌ $*" >&2; exit 1; }

echo "═══ نسخة احتياطية: $TIMESTAMP ═══"

# ── ١) قاعدة البيانات ──
docker ps --format '{{.Names}}' | grep -qx "$DB_CONTAINER" \
  || fail "حاوية قاعدة البيانات «$DB_CONTAINER» مو شغالة — ما نكدر ناخذ نسخة"

echo "→ قاعدة البيانات..."
docker exec "$DB_CONTAINER" pg_dump -U staffmange --clean --if-exists staffmange \
  > "$WORK/database.sql" || fail "فشل تصدير قاعدة البيانات"

# فحص سلامة: pg_dump ينهي الملف بهذا السطر. بدون الفحص ممكن ننام
# مرتاحين على نسخة مقطوعة بالنص ونكتشفها بس يوم الكارثة.
tail -5 "$WORK/database.sql" | grep -q "PostgreSQL database dump complete" \
  || fail "ملف قاعدة البيانات ناقص أو مقطوع — النسخة مرفوضة"

TABLES=$(grep -c "^CREATE TABLE" "$WORK/database.sql" || true)
[ "$TABLES" -ge 20 ] || fail "عدد الجداول بالنسخة $TABLES — قليل جداً، شكله في خطأ"
echo "  ✓ $TABLES جدول"

# ── ٢) الملفات المرفوعة ──
if docker ps --format '{{.Names}}' | grep -qx "$BACKEND_CONTAINER"; then
  echo "→ الملفات المرفوعة..."
  docker exec "$BACKEND_CONTAINER" sh -c 'cd /app/uploads 2>/dev/null && tar cf - . 2>/dev/null' \
    > "$WORK/uploads.tar" 2>/dev/null || true
  if [ -s "$WORK/uploads.tar" ]; then
    echo "  ✓ $(du -h "$WORK/uploads.tar" | cut -f1)"
  else
    rm -f "$WORK/uploads.tar"
    echo "  — ماكو ملفات مرفوعة (طبيعي إذا التخزين على R2)"
  fi
else
  echo "  ⚠️ حاوية الباك إند مو شغالة — الملفات المرفوعة ما انأخذت"
fi

# ── ٣) الإعدادات والأسرار ──
if [ -f "$SCRIPT_DIR/.env" ]; then
  cp "$SCRIPT_DIR/.env" "$WORK/env.txt"
  echo "→ ملف الإعدادات ✓"
else
  echo "  ⚠️ ماكو ملف .env — بدونه ما تكدر تشغّل النظام من النسخة!"
fi

# ── التغليف ──
ARCHIVE="$BACKUP_DIR/staffmange_$TIMESTAMP.tar.gz"
tar czf "$ARCHIVE" -C "$WORK" .
SIZE=$(du -h "$ARCHIVE" | cut -f1)

# فحص أخير: نتأكد الأرشيف ينفتح فعلاً قبل ما نعتبره نسخة
tar tzf "$ARCHIVE" >/dev/null 2>&1 || fail "الأرشيف تالف — النسخة مرفوضة"

echo "✅ النسخة جاهزة: $(basename "$ARCHIVE") ($SIZE)"

# ── رفع خارج السيرفر ──
# النسخة على نفس السيرفر ما تحميك من ضياع السيرفر. إذا R2 مضبوط
# نرفعها عليه — تخزين خارجي بحساب ثاني عند كلاودفلير.
if [ -f "$SCRIPT_DIR/.env" ]; then
  set -a; . "$SCRIPT_DIR/.env"; set +a
fi
if [ -n "${R2_BUCKET:-}" ] && [ -n "${R2_ACCESS_KEY:-}" ] && command -v aws >/dev/null 2>&1; then
  echo "→ الرفع لـR2..."
  if AWS_ACCESS_KEY_ID="$R2_ACCESS_KEY" AWS_SECRET_ACCESS_KEY="$R2_SECRET_KEY" \
     aws s3 cp "$ARCHIVE" "s3://$R2_BUCKET/backups/$(basename "$ARCHIVE")" \
     --endpoint-url "$R2_ENDPOINT" >/dev/null 2>&1; then
    echo "  ✓ انرفعت خارج السيرفر"
  else
    echo "  ⚠️ فشل الرفع لـR2 — النسخة محفوظة محلياً بس"
  fi
else
  echo "  ⚠️ ماكو نسخة خارج السيرفر — شغّل ./setup-backups.sh حتى تشوف الخيارات"
fi

# ── تنظيف القديم ──
find "$BACKUP_DIR" -name "staffmange_*.tar.gz" -mtime +$KEEP_DAYS -delete
find "$BACKUP_DIR" -name "staffmange_*.sql.gz" -mtime +$KEEP_DAYS -delete  # الصيغة القديمة
COUNT=$(find "$BACKUP_DIR" -name "staffmange_*.tar.gz" | wc -l)
echo "📦 عدد النسخ المحفوظة: $COUNT (نحتفظ بـ$KEEP_DAYS يوم)"
