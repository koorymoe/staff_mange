#!/usr/bin/env python3
"""ون بيس — نقل كل mux.Handle الباقية (بحارس الدخول) للنواة دفعة وحدة.

لكل مسار: الحارس المعروف يصير قاعدة بنفس معناه بالضبط؛ أي حارس ما نعرف نترجمه
بأمان (أو حارس ثاني بنفس المسار) يبقى طبقة ورا القاعدة — فولا مسار يتوسّع.
المسارات العامة (بلا requireAuth) تبقى مثل ما هي.

الاستعمال: onepiece-bulk.py backend-go/cmd/api/main.go backend-go/internal/apps/sections/rules_gen.go
"""
import re, sys, collections

main_path, out_path = sys.argv[1], sys.argv[2]
s = open(main_path).read()

def bal_end(t, i):  # i عند '('
    d = 0; q = None; j = i
    while j < len(t):
        c = t[j]
        if q:
            if c == '\\': j += 2; continue
            if c == q: q = None
        elif c in '"`': q = c
        elif c == '/' and t[j:j+2] == '//':
            j = t.index('\n', j); continue
        elif c in '([{': d += 1
        elif c in ')]}':
            d -= 1
            if d == 0: return j
        j += 1
    raise SystemExit('قوس ما انسد')

def strip_comments(t):
    out, q, j = [], None, 0
    while j < len(t):
        c = t[j]
        if q:
            out.append(c)
            if c == '\\': out.append(t[j+1]); j += 2; continue
            if c == q: q = None
        elif c in '"`': q = c; out.append(c)
        elif t[j:j+2] == '//':
            j = t.index('\n', j) if '\n' in t[j:] else len(t); continue
        else: out.append(c)
        j += 1
    return ''.join(out)

def split_args(t):
    t = strip_comments(t)
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

# ١) تعريفات الحراس: name := expr
defs = {}
for m in re.finditer(r'\n\t+(require\w+|guardHR\w*|\w+Guard) := ', s):
    k = m.end()
    m2 = re.match(r'[\w.\[\]"]+\(', s[k:])
    if not m2: continue
    e = bal_end(s, k + m2.end() - 1)
    defs[m.group(1)] = strip_comments(s[k:e+1]).strip()

def strs(t):
    return re.findall(r'"([^"]*)"', t)

def rule_of(g):
    """يرجع (تعبير Go للقاعدة، وصف) أو None إذا ما نعرف نترجمه بأمان."""
    e = defs.get(g, g)
    e = re.sub(r'\s+', ' ', e)
    if g == 'requireAdmin':
        return 'admin(l)', 'مدير النظام'
    m = re.fullmatch(r'requireMon\["(\w+)"\]', g)
    if m:
        return f'rolePerm(l, []string{{"ADMIN", "MONITOR"}}, "monitoring", "auditing", "{m.group(1)}")', 'المراقب'
    m = re.fullmatch(r'guards\.(Permission|AnyPermission)\((.*)\)', e)
    if m and re.fullmatch(r'("[\w]+"\s*,?\s*)+', m.group(2)):
        return 'perm(l, ' + ', '.join(f'"{p}"' for p in strs(m.group(2))) + ')', 'صلاحية'
    m = re.fullmatch(r'guards\.(RoleOrPermission|RoleOrAnyPermission|LeaderOrRoleOrAnyPermission)\(\s*\[\]string\{([^}]*)\}\s*,(.*)\)', e)
    if m and re.fullmatch(r'\s*("[\w]+"\s*,?\s*)+', m.group(3)):
        roles = ', '.join(f'"{r}"' for r in strs(m.group(2)))
        perms = ', '.join(f'"{p}"' for p in strs(m.group(3)))
        if m.group(1) == 'LeaderOrRoleOrAnyPermission':
            return f'leaderRolePerm(l, []string{{{roles}}}, {perms})', 'ليدر/دور/صلاحية'
        return f'rolePerm(l, []string{{{roles}}}, {perms})', 'دور/صلاحية'
    m = re.fullmatch(r'guards\.(LeaderOrPermission|LeaderOrAnyPermission)\((.*)\)', e)
    if m and re.fullmatch(r'\s*("[\w]+"\s*,?\s*)+', m.group(2)):
        return 'leaderPerm(l, ' + ', '.join(f'"{p}"' for p in strs(m.group(2))) + ')', 'ليدر/صلاحية'
    m = re.fullmatch(r'middleware\.RequireRole\(\s*employeeRepo\s*,\s*notificationRepo\s*,(.*)\)', e)
    if m and re.fullmatch(r'\s*("[\w]+"\s*,?\s*)+', m.group(1)):
        roles = strs(m.group(1))
        lit = ', '.join(f'"{r}"' for r in roles)
        if 'ADMIN' in roles:
            return f'roles(l, {lit})', 'دور'
        return f'strictRoles(l, {lit})', 'دور (بلا المدير)'
    m = re.fullmatch(r'authz\.Guard\(\s*guards\s*,\s*authz\.(\w+)\s*\)', e)
    if m:
        return f'fromAuthz(l, authz.{m.group(1)})', 'محرك الصلاحيات'
    if re.fullmatch(r'middleware\.RequireOwner\(\)', e) or re.fullmatch(r'middleware\.RequireOwnerOnly\(.*\)', e):
        return 'owner(l)', 'المالك بس'
    return None

def app_of(path):
    seg = path.split('/')[2] if path.count('/') >= 2 else 'root'
    seg = re.sub(r'[^a-z0-9]', '_', seg.lower()) or 'root'
    return seg

taken = {'system', 'people', 'org', 'me', 'requests', 'field', 'access', 'finance', 'invoices'}
verb = {'GET': 'يقرا', 'POST': 'ينفّذ', 'PUT': 'يعدّل', 'PATCH': 'يعدّل', 'DELETE': 'يشيل'}
rules = collections.defaultdict(dict)
names = set()
moved = 0; layered = 0
out = []; pos = 0
for m in re.finditer(r'mux\.Handle\("([^"]+)",', s):
    i = m.start()
    if i < pos: continue
    pat = m.group(1)
    end = bal_end(s, i + len('mux.Handle'))
    args = split_args(s[i + len('mux.Handle') + 1:end])
    cm = re.fullmatch(r'middleware\.Chain\((.*)\)', args[1], re.S)
    if not cm: continue
    parts = split_args(cm.group(1))
    if 'requireAuth' not in parts: continue  # العامة تبقى
    method, path = (pat.split(' ', 1) + [''])[:2] if ' ' in pat else ('ANY', pat)
    h = parts[0]
    hm = re.fullmatch(r'http\.HandlerFunc\((.*)\)', h, re.S)
    h = hm.group(1) if hm else f'{h}.ServeHTTP'
    guards = [p for p in parts[1:] if p != 'requireAuth']
    rule, note, mws = None, 'الدخول — الفحص بالمعالج', []
    for g in guards:
        if g == 'idem':
            mws.append(g); continue
        if rule is None:
            r = rule_of(g)
            if r:
                rule, note = r; continue
        mws.append(g)
    if rule is None:
        rule = 'any_(l)'
    if any(g != 'idem' for g in mws):
        layered += 1
        note += ' + طبقة: ' + ', '.join(re.sub(r'\s+', ' ', g)[:40] for g in mws if g != 'idem')
    app = app_of(path)
    if app in taken: app += '_x'
    slug = re.sub(r'[^a-z0-9]+', '_', '_'.join(('by_' + p[1:-1].rstrip('.')) if p.startswith('{') else p for p in path.split('/')[3:]).lower()).strip('_') or 'root'
    act = f'{app}.{method.lower()}_{slug}'
    n = 2
    while act in names:
        act = f'{app}.{method.lower()}_{slug}_{n}'; n += 1
    names.add(act)
    label = f'{verb.get(method, method)} {path} ({note})'.replace('"', "'")
    rules[app][act] = (rule, label)
    extra = ''.join(', ' + g for g in mws)
    out.append(s[pos:i])
    out.append(f'secs["{app}"].Set("{pat}", "{act}", {h}{extra})')
    pos = end + 1
    moved += 1
out.append(s[pos:])
open(main_path, 'w').write(''.join(out))

uses_authz = any('authz.' in r for a in rules.values() for r, _ in a.values())
g = ['// Code generated by scripts/onepiece-bulk.py — ون بيس 11-10. القواعد منقولة من حراس main.go بنفس المعنى.',
     '// عدّل القاعدة هنا (مو بـmain.go). الحارس الي ما انترجم بأمان بقى طبقة بـmain.go.', '',
     'package sections', '', 'import (']
if uses_authz: g.append('\t"staffmange-api/internal/authz"')
g += ['\t"staffmange-api/internal/core"', ')', '', 'func rules() map[string]map[core.Action]core.Rule {', '\treturn map[string]map[core.Action]core.Rule{']
for app in sorted(rules):
    g.append(f'\t\t"{app}": {{')
    for act in sorted(rules[app]):
        r, label = rules[app][act]
        g.append(f'\t\t\t"{act}": func(l string) core.Rule {{ return {r} }}("{label}"),')
    g.append('\t\t},')
g += ['\t}', '}', '']
open(out_path, 'w').write('\n'.join(g))
print(f'انتقل {moved} مسار بـ{len(rules)} تطبيق، منها {layered} بطبقة حارس قديم')
