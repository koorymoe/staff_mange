#!/usr/bin/env python3
"""وقت الجلسة الفعلي ليوم واحد بالثانية — (ع) 11-10: «سجّل كلشي حتى أجزاء الثانية».

الاستعمال: python3 scripts/daytime.py <session.jsonl> <YYYY-MM-DD بغداد>
- فترة النشاط: رسائل متتالية بينها ≤ ٥ دقايق (شغل لوفي + كتابة دازاي).
- الفجوة > ٥ دقايق: انتظار — تنكتب بسطر وحدها وما تنحسب.
- كتابة دازاي: من آخر رد لحد رسالته الجاية (إذا ≤ ٥ دقايق).
"""
import json, sys
from datetime import datetime, timedelta, timezone

BGD = timezone(timedelta(hours=3))
GAP = 300
path, day = sys.argv[1], sys.argv[2]
ev = []
for line in open(path, encoding='utf-8'):
    try:
        d = json.loads(line)
    except ValueError:
        continue
    t = d.get('timestamp')
    if not t:
        continue
    x = datetime.fromisoformat(t.replace('Z', '+00:00')).astimezone(BGD)
    if x.strftime('%Y-%m-%d') != day:
        continue
    human = d.get('type') == 'user' and isinstance(d.get('message', {}).get('content'), str)
    ev.append((x, d.get('type'), human))
ev.sort(key=lambda e: e[0])
if not ev:
    sys.exit('ماكو نشاط بهذا اليوم')

spans, waits, typing = [], [], 0.0
start = prev = ev[0][0]
last_reply = None
for x, typ, human in ev:
    if (x - prev).total_seconds() > GAP:
        spans.append((start, prev)); waits.append((prev, x)); start = x
    if human and last_reply and 0 < (x - last_reply).total_seconds() <= GAP:
        typing += (x - last_reply).total_seconds()
    if typ == 'assistant':
        last_reply = x
    prev = x
spans.append((start, prev))
f = lambda s: f'{int(s // 3600)}:{int(s % 3600 // 60):02d}:{s % 60:06.3f}'
tot = 0.0
print('| من | لـ | المدة |')
print('|---|---|---|')
for a, b in spans:
    s = (b - a).total_seconds(); tot += s
    print(f'| {a:%H:%M:%S.%f}'[:-3] + f' | {b:%H:%M:%S.%f}'[:-3] + f' | {f(s)} |')
print(f'\nالنشاط: {f(tot)} · منه كتابة دازاي: {f(typing)}')
print(f'الانتظار (ما محسوب): {f(sum((b - a).total_seconds() for a, b in waits))} بـ{len(waits)} فجوة')
