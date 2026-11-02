#!/bin/bash
# يشتغل هذا السكربت على السيرفر نفسه (مو على جهازك) لرفع/تحديث النظام.
## أول مرة: انسخ المشروع، اعمل .env، وبعدين شغّل هذا السكربت.
## كل تحديث بعدين: فقط شغّل هذا السكربت من جديد بمجلد المشروع.
set -e

# 🛟 (11-02) نسخة احتياطية قبل كل رفع — إذا التحديث خرّب شي نرجع لها فوراً.
# SKIP_PREDEPLOY_BACKUP=1 للتخطي (مثلاً أول تنصيب ما بيه قاعدة بعد).
if [ -z "${SKIP_PREDEPLOY_BACKUP:-}" ] && [ -x ./backup-db.sh ] && docker compose ps --status running 2>/dev/null | grep -q db; then
  echo "==> نسخة احتياطية قبل التحديث (مع تجربة استرجاع)"
  ./backup-db.sh || { echo "❌ النسخة الاحتياطية فشلت — وقفت التحديث. صلّحها أو شغّل ويا SKIP_PREDEPLOY_BACKUP=1"; exit 1; }
fi

echo "==> سحب آخر تحديثات الكود"
git pull

echo "==> بناء وتشغيل الحاويات (قد تاخذ وكت أول مرة)"
docker compose build
docker compose up -d

echo "==> تنظيف صور Docker القديمة غير المستخدمة"
docker image prune -f

echo "==> تم! حالة الحاويات:"
docker compose ps
