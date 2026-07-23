#!/bin/bash
# نسخة احتياطية يومية لقاعدة بيانات النظام — تتخزن محلياً بمجلد backups/
# مع الاحتفاظ بآخر 14 يوم بس (تحذف الأقدم تلقائياً حتى ما تمتلئ مساحة القرص).
# يشتغل هذا السكربت تلقائياً كل يوم عن طريق cron بعد تشغيل setup-backups.sh مرة وحدة.
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKUP_DIR="$SCRIPT_DIR/backups"
mkdir -p "$BACKUP_DIR"

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
FILE="$BACKUP_DIR/staffmange_$TIMESTAMP.sql.gz"

docker exec staff_mange-db-1 pg_dump -U staffmange staffmange | gzip > "$FILE"
echo "==> تم حفظ نسخة احتياطية: $FILE ($(du -h "$FILE" | cut -f1))"

# حذف النسخ الأقدم من 14 يوم حتى ما تمتلئ مساحة القرص
find "$BACKUP_DIR" -name "staffmange_*.sql.gz" -mtime +14 -delete
