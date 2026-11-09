#!/usr/bin/env bash
# طلبات تحديث الوكلاء لـ Claude (دازاي 11-06) — **قراءة بس**: اقتراحات ماتركس وكمال الي المالك
# وافق عليها وتحتاج برمجة. الفحص المجدول يقراها ويكتب خطة، والتنفيذ بس بعد «موافق» من دازاي.
#
#   BOT_READER_TOKEN=... scripts/agentrequests.sh
#
# ⚠️ الرمز من متغيّر البيئة بس — لا ينكتب بملف ولا ينطبع.
set -euo pipefail
: "${BOT_READER_TOKEN:?ماكو BOT_READER_TOKEN بالبيئة}"
BASE="${BOT_BASE_URL:-https://staffmanage.cc}"
OUT="$(mktemp)"; trap 'rm -f "$OUT"' EXIT
curl -fsS -H "X-Bot-Reader-Token: $BOT_READER_TOKEN" "$BASE/api/bot-reader/suggestions" > "$OUT"
python3 -I - "$OUT" <<'PY'
import json, sys
rows = json.load(open(sys.argv[1], encoding='utf-8'))
names = {'MATRIX': 'ماتركس', 'IRIS': 'كمال'}
print(f'طلبات تحديث موافق عليها وتنتظر التنفيذ: {len(rows)}')
for r in rows:
    print(f"\n[{r['id'][:8]}] {names.get(r['agent'], r['agent'])} · وافق {(r.get('decidedAt') or '')[:10]}")
    print(f"   {r['title']}")
    print(f"   {r['details']}")
PY
