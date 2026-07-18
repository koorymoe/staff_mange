#!/bin/bash
# يشتغل هذا السكربت على السيرفر نفسه (مو على جهازك) لرفع/تحديث النظام.
# أول مرة: انسخ المشروع، اعمل .env، وبعدين شغّل هذا السكربت.
# كل تحديث بعدين: فقط شغّل هذا السكربت من جديد بمجلد المشروع.
set -e

echo "==> سحب آخر تحديثات الكود"
git pull origin claude/zen-mendel-70b9j2

echo "==> بناء وتشغيل الحاويات (قد تاخذ وكت أول مرة)"
docker compose build
docker compose up -d

echo "==> تنظيف صور Docker القديمة غير المستخدمة"
docker image prune -f

echo "==> تم! حالة الحاويات:"
docker compose ps
