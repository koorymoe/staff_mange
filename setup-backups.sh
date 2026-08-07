#!/bin/bash
# ═══════════════════════════════════════════════════════════════════
# تفعيل النسخ الاحتياطي التلقائي — شغّله مرة وحدة بس
# ═══════════════════════════════════════════════════════════════════
# آمن تشغّله أكثر من مرة (ما يكرر الجدولة).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CRON_LINE="0 3 * * * $SCRIPT_DIR/backup-db.sh >> $SCRIPT_DIR/backups/backup.log 2>&1"

mkdir -p "$SCRIPT_DIR/backups"
chmod +x "$SCRIPT_DIR"/backup-db.sh "$SCRIPT_DIR"/backup-status.sh "$SCRIPT_DIR"/restore-backup.sh 2>/dev/null || true

( crontab -l 2>/dev/null | grep -v "backup-db.sh" ; echo "$CRON_LINE" ) | crontab -

echo "✅ النسخ الاحتياطي التلقائي مفعّل — كل يوم الساعة 3:00 فجراً"
echo
crontab -l | grep backup-db.sh
echo

# ═══ النسخة خارج السيرفر ═══
# هاي أهم فقرة بالملف كله. النسخ الي بمجلد backups/ موجودة على نفس
# السيرفر — لو تلف السيرفر أو انحذف أو ضاع الوصول له، النسخ تروح معاه
# بنفس اللحظة. يعني عملياً ما عندك نسخة احتياطية، عندك نسخة ثانية من
# نفس الشي بنفس المكان.
echo "═══════════════════════════════════════════"
echo "⚠️  مهم جداً: النسخة خارج السيرفر"
echo "═══════════════════════════════════════════"
if [ -f "$SCRIPT_DIR/.env" ]; then
  set -a; . "$SCRIPT_DIR/.env" 2>/dev/null; set +a
fi

if [ -n "${R2_BUCKET:-}" ] && [ -n "${R2_ACCESS_KEY:-}" ]; then
  if command -v aws >/dev/null 2>&1; then
    echo "✅ R2 مضبوط وأداة aws موجودة — النسخ راح تنرفع تلقائياً."
  else
    echo "⚠️ R2 مضبوط بس أداة aws مو منصّبة. نصّبها بأمر واحد:"
    echo "   apt-get install -y awscli"
  fi
else
  cat <<'GUIDE'
❌ ماكو نسخة خارج السيرفر حالياً.

كل نسخك محفوظة على نفس السيرفر. لو ضاع السيرفر — تلف، انحذف، أو ما
عاد عندك وصول له — تروح كل النسخ معاه بنفس اللحظة.

عندك خياران، اختار واحد على الأقل:

  ── الخيار ١: Cloudflare R2 (تلقائي، مجاني لحد ١٠ جيجا) ──
  ١. من لوحة Cloudflare: R2 → Create bucket
  ٢. Manage API Tokens → أنشئ توكن بصلاحية قراءة وكتابة
  ٣. حط القيم بملف .env على السيرفر:
       R2_BUCKET=اسم-الباكت
       R2_ACCESS_KEY=...
       R2_SECRET_KEY=...
       R2_ENDPOINT=https://<account-id>.r2.cloudflarestorage.com
  ٤. نصّب الأداة:  apt-get install -y awscli
  ٥. جرّب:        ./backup-db.sh
  بعدها كل نسخة يومية تنرفع تلقائياً برّا السيرفر.

  ── الخيار ٢: Hetzner Backups (صورة كاملة للسيرفر) ──
  من لوحة Hetzner: Server → Options → Backups → Enable
  تاخذ صورة كاملة للسيرفر أسبوعياً. تكلف ٢٠٪ من سعر السيرفر.
  ملاحظة: هاي تحميك من ضياع السيرفر، بس أسبوعية — يعني ممكن تخسر
  شغل أسبوع كامل. الأفضل تفعّل الاثنين سوه: R2 يومي + Hetzner أسبوعي.

GUIDE
fi

echo "═══════════════════════════════════════════"
echo "→ نسوي نسخة أولى الحين للتأكد:"
echo
"$SCRIPT_DIR/backup-db.sh"
echo
echo "لفحص الحالة بأي وقت:  ./backup-status.sh"
echo "للاسترجاع:            ./restore-backup.sh"
