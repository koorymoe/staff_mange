#!/usr/bin/env python3
"""ون بيس: يبدّل mux.Handle("PAT", middleware.Chain(H, requireAuth, …)) بـ APP.Set("PAT", "ACT", H[, idem]).
الاستعمال: migrate.py main.go APPVAR mapping.json   (mapping: {"GET /x": "app.act"})"""
import json, re, sys
path, app, mp = sys.argv[1], sys.argv[2], json.load(open(sys.argv[3]))
s = open(path).read()
def call_end(s, i):  # i عند '(' بعد mux.Handle
    d = 0; q = None; j = i
    while j < len(s):
        c = s[j]
        if q:
            if c == '\\': j += 2; continue
            if c == q: q = None
        elif c in '"`': q = c
        elif c == '(': d += 1
        elif c == ')':
            d -= 1
            if d == 0: return j
        j += 1
    raise SystemExit('ما انسدّ القوس')
def split_args(t):
    out, d, q, cur = [], 0, None, ''
    for c in t:
        if q:
            cur += c
            if c == q: q = None
            continue
        if c in '"`': q = c
        if c in '([{': d += 1
        if c in ')]}': d -= 1
        if c == ',' and d == 0: out.append(cur.strip()); cur = ''; continue
        cur += c
    if cur.strip(): out.append(cur.strip())
    return out
for pat, act in mp.items():
    key = 'mux.Handle("' + pat + '",'
    i = s.find(key)
    if i < 0: raise SystemExit('ما لگيت ' + pat)
    if s.find(key, i + 1) >= 0: raise SystemExit('مكرر ' + pat)
    open_ = i + len('mux.Handle')
    end = call_end(s, open_)
    inner = s[open_ + 1:end]
    args = split_args(inner)
    chain = args[1]
    m = re.match(r'middleware\.Chain\((.*)\)$', chain, re.S)
    if not m: raise SystemExit('مو Chain: ' + pat)
    parts = split_args(m.group(1))
    h = parts[0]
    hm = re.match(r'http\.HandlerFunc\((.*)\)$', h, re.S)
    h = hm.group(1) if hm else h
    if 'requireAuth' not in parts: raise SystemExit('بلا requireAuth: ' + pat)
    extra = ', idem' if 'idem' in parts else ''
    rest = [p for p in parts[1:] if p not in ('requireAuth', 'idem')]
    s = s[:i] + f'{app}.Set("{pat}", "{act}", {h}{extra})' + s[end + 1:]
open(path, 'w').write(s)
print('انتقل', len(mp))
