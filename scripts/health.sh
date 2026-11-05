#!/usr/bin/env bash
# قارئ صحة النظام لـ Claude (دازاي 11-05) — **قراءة بس**. كل مصدر اختياري: الي مفتاحه
# مو موجود بالبيئة ينتخطى. المفاتيح من البيئة بس — لا تنطبع ولا تنكتب بملف.
#
#   SENTRY_READ_TOKEN + SENTRY_ORG     : أخطاء Sentry غير المحلولة بآخر 24 ساعة
#   UPTIMEROBOT_API_KEY                : حالة المراقبات (read-only key)
#   OPS_BASIC (user:pass)              : Netdata عبر https://الدومين/ops/netdata/
#   BOT_BASE_URL (افتراضي staffmanage.cc)
set -uo pipefail
BASE="${BOT_BASE_URL:-https://staffmanage.cc}"

echo "== السيرفر"
code=$(curl -s -o /dev/null -w '%{http_code} %{time_total}' --max-time 15 "$BASE/api/health")
echo "   /api/health → $code (الحالة، والوقت بالثواني)"

if [ -n "${SENTRY_READ_TOKEN:-}" ] && [ -n "${SENTRY_ORG:-}" ]; then
  echo "== Sentry (غير محلولة، آخر 24 ساعة)"
  curl -fsS --max-time 20 -H "Authorization: Bearer $SENTRY_READ_TOKEN" \
    "https://sentry.io/api/0/organizations/$SENTRY_ORG/issues/?query=is:unresolved&statsPeriod=24h&sort=freq&limit=15" \
  | python3 -I -c '
import json,sys
d=json.load(sys.stdin)
print(f"   عدد المشاكل: {len(d)}")
for i in d:
    print(f"   • [{i.get(\"project\",{}).get(\"slug\",\"?\")}] {i.get(\"title\",\"\")[:120]} — مرات: {i.get(\"count\")} · مستخدمين: {i.get(\"userCount\")} · آخر: {i.get(\"lastSeen\",\"\")[:16]} · {i.get(\"permalink\",\"\")}")
' || echo "   تعذر القراءة من Sentry"
else
  echo "== Sentry: ماكو SENTRY_READ_TOKEN/SENTRY_ORG — تخطّي"
fi

if [ -n "${UPTIMEROBOT_API_KEY:-}" ]; then
  echo "== UptimeRobot"
  curl -fsS --max-time 20 -X POST "https://api.uptimerobot.com/v2/getMonitors" \
    -d "api_key=$UPTIMEROBOT_API_KEY&format=json&logs=1&logs_limit=3&custom_uptime_ratios=1-7-30" \
  | python3 -I -c '
import json,sys
d=json.load(sys.stdin); st={0:"موقوف",1:"ما انفحص",2:"شغّال ✅",8:"يبين واگف ⚠️",9:"واگف ❌"}
for m in d.get("monitors",[]):
    r=m.get("custom_uptime_ratio","?-?-?").split("-")
    print(f"   • {m.get(\"friendly_name\")}: {st.get(m.get(\"status\"),m.get(\"status\"))} · التوفر يوم/أسبوع/شهر: {\" / \".join(x+\"%\" for x in r)}")
    for l in (m.get("logs") or [])[:3]:
        if l.get("type")==1: print(f"     ↳ توقف {l.get(\"duration\",0)//60} دقيقة")
' || echo "   تعذر القراءة من UptimeRobot"
else
  echo "== UptimeRobot: ماكو UPTIMEROBOT_API_KEY — تخطّي"
fi

if [ -n "${OPS_BASIC:-}" ]; then
  echo "== Netdata (الموارد الآن)"
  n="$BASE/ops/netdata/api/v1/data?after=-60&points=1&format=json&group=average"
  for ch in system.ram system.cpu disk_space._ ; do
    curl -fsS --max-time 15 -u "$OPS_BASIC" "$n&chart=$ch" 2>/dev/null | python3 -I -c '
import json,sys
d=json.load(sys.stdin); lab=d.get("labels",[]); row=(d.get("data") or [[]])[0]
print("   •", sys.argv[1], ", ".join(f"{a}={round(b,1)}" for a,b in zip(lab[1:],row[1:]) if b is not None))
' "$ch" || echo "   تعذر $ch"
  done
  curl -fsS --max-time 15 -u "$OPS_BASIC" "$BASE/ops/netdata/api/v1/alarms" 2>/dev/null | python3 -I -c '
import json,sys
d=json.load(sys.stdin).get("alarms",{})
print(f"   • تنبيهات نشطة: {len(d)}")
for k,a in list(d.items())[:10]: print(f"     ↳ {a.get(\"status\")}: {a.get(\"name\")} = {a.get(\"value_string\")}")
' || echo "   تعذر قراءة التنبيهات"
else
  echo "== Netdata: ماكو OPS_BASIC — تخطّي"
fi
