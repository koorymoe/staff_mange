#!/usr/bin/env bash
# قارئ بلاغات البوت لـ Claude (دازاي 11-05) — **قراءة بس**: يجيب البلاغات وينزّل صورها.
# ما يغيّر حالة ولا يرد على أحد؛ هذا يبقى من شاشة المالك بعد موافقته.
#
#   BOT_READER_TOKEN=... scripts/botreports.sh [STATUS] [OUT_DIR]
#   STATUS: NEW (الافتراضي) | NEEDS_INFO | IN_PROGRESS | DONE | NOT_A_BUG | REJECTED | all
#   SINCE_HOURS=N (اختياري): يعرض بس البلاغات الي انخلقت بآخر N ساعة، والأقدم ينعدّ بس
#   (للفحص المجدول: كل موعد يشوف الجديد من الموعد الي قبله).
#
# ⚠️ الرمز من متغيّر البيئة بس — لا ينكتب بملف ولا ينطبع.
set -euo pipefail
: "${BOT_READER_TOKEN:?ماكو BOT_READER_TOKEN بالبيئة}"
BASE="${BOT_BASE_URL:-https://staffmanage.cc}"
STATUS="${1:-NEW}"; [ "$STATUS" = all ] && STATUS=""
OUT="${2:-./botreports-out}"
mkdir -p "$OUT"
curl -fsS -H "X-Bot-Reader-Token: $BOT_READER_TOKEN" "$BASE/api/bot-reader/reports?status=$STATUS" > "$OUT/reports.json"
python3 -I - "$OUT" "${SINCE_HOURS:-}" <<'PY'
import json, sys, os
from datetime import datetime, timedelta, timezone
out = sys.argv[1]
since = sys.argv[2] if len(sys.argv) > 2 else ''
d = json.load(open(os.path.join(out, 'reports.json'), encoding='utf-8'))
items, labels = d.get('items', []), d.get('statuses', {})
if since:
    cut = datetime.now(timezone.utc) - timedelta(hours=float(since))
    ts = lambda r: datetime.fromisoformat(r['createdAt'].replace('Z', '+00:00'))
    older = [r for r in items if ts(r) <= cut]
    items = [r for r in items if ts(r) > cut]
    print(f"جديد بآخر {since} ساعة: {len(items)} · أقدم بنفس الحالة: {len(older)}" + (f" (أرقامها: {', '.join('#'+str(r['number']) for r in older[:20])})" if older else ''))
else:
    print(f"عدد البلاغات: {len(items)}")
with open(os.path.join(out, 'photos.txt'), 'w') as f:
    for r in items:
        print(f"\n#{r['number']} · {r['employeeName']} ({r['employeeRole']}) · {labels.get(r['status'], r['status'])} · {r['source']} · {r['createdAt'][:16]}")
        if r.get('page'): print(f"   الشاشة: {r['page']}")
        print("   " + r['body'].replace('\n', '\n   '))
        for i, p in enumerate(r.get('photos') or []):
            key = p.replace('/api/files/', '', 1)
            name = f"{r['number']}-{i+1}{os.path.splitext(key)[1]}"
            f.write(f"{key} {name}\n")
            print(f"   📷 {os.path.join(out, name)}")
PY
while read -r key name; do
  curl -fsS -H "X-Bot-Reader-Token: $BOT_READER_TOKEN" "$BASE/api/bot-reader/files/$key" -o "$OUT/$name" || echo "تعذر تنزيل $name" >&2
done < "$OUT/photos.txt"
