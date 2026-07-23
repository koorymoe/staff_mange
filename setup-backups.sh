#!/bin/bash
# شغّل هذا السكربت مرة وحدة بس على السيرفر حتى يفعّل النسخ الاحتياطي التلقائي
# اليومي (الساعة 3 فجراً كل يوم) عن طريق cron. آمن تشغله أكثر من مرة (ما يكرر الجدولة).
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CRON_LINE="0 3 * * * $SCRIPT_DIR/backup-db.sh >> $SCRIPT_DIR/backups/backup.log 2>&1"

mkdir -p "$SCRIPT_DIR/backups"
chmod +x "$SCRIPT_DIR/backup-db.sh"

( crontab -l 2>/dev/null | grep -v "backup-db.sh" ; echo "$CRON_LINE" ) | crontab -

echo "==> تم تفعيل النسخ الاحتياطي التلقائي — كل يوم الساعة 3:00 فجراً"
echo "==> الجدولة الحالية:"
crontab -l

echo ""
echo "==> نسوي نسخة احتياطية أولى الحين حتى نتأكد كل شي يشتغل صح:"
"$SCRIPT_DIR/backup-db.sh"
