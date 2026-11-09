#!/usr/bin/env bash
# ساعات Claude لسجل دوام دازاي (قرار 0011). يسجّل سطر واحد، والمفتاح يمنع التكرار.
#
#   REMOTE_HOURS_AGENT_TOKEN=... scripts/remotehours.sh <YYYY-MM-DD> <minutes> <key> "<note>"
#   مثال: scripts/remotehours.sh 2026-11-06 9 2026-11-06-0952-check "فحص ١٠ الصبح: #8 #9 #10"
#
# 404 = المسار بعده مطفي (الرمز مو بالسيرفر) — يطلع بصمت بكود 3.
# ⚠️ الرمز من متغيّر البيئة بس — لا ينكتب بملف ولا ينطبع.
set -euo pipefail
: "${REMOTE_HOURS_AGENT_TOKEN:?ماكو REMOTE_HOURS_AGENT_TOKEN بالبيئة}"
BASE="${BOT_BASE_URL:-https://staffmanage.cc}"
[ $# -eq 4 ] || { echo "الاستعمال: $0 <date> <minutes> <key> <note>" >&2; exit 2; }
BODY=$(python3 -I -c 'import json,sys; print(json.dumps({"date":sys.argv[1],"minutes":int(sys.argv[2]),"note":sys.argv[3]}, ensure_ascii=False))' "$1" "$2" "$4")
CODE=$(curl -sS -o /dev/null -w '%{http_code}' -X POST "$BASE/api/remote-hours/agent" \
  -H "X-Remote-Hours-Agent-Token: $REMOTE_HOURS_AGENT_TOKEN" -H "Idempotency-Key: $3" \
  -H 'Content-Type: application/json' --data "$BODY")
case "$CODE" in
  201) echo "✅ انسجّلت $2 دقيقة بتاريخ $1" ;;
  200) echo "↩️ مسجّلة قبل (نفس المفتاح)" ;;
  404) echo "المسار مطفي (الرمز مو بالسيرفر بعد)"; exit 3 ;;
  *)   echo "❌ رد $CODE"; exit 1 ;;
esac
