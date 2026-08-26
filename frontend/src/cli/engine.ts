// ═══ محرّك الـCLI ═══
//
// محرّك واحد يشتغل على أي نحو (سيسكو، هواوي VRP، TP-Link). الفرق بين
// الأجهزة **بيانات** — شجرة أوامر ولاحقة محث ونصوص أخطاء.
//
// ⚠️ أصعب جزء وأهمه: **مطابقة الاختصار**. الجهاز الحقيقي يقبل
// `int gi0/1` و`sh ru`، ويرفض `s` لأنها تطابق `show` و`shutdown`
// سوا. لو قبلنا أي اختصار غامض، الفني يتعوّد على شي ما يشتغل بالميدان
// — والمحاكي يصير يعلّم غلط. وهذا أسوأ من إنه ما يعلّم أصلاً.

import type { ArgKind, CliGrammar, CliNode, CliState } from './grammar'

export interface CliSession {
  /** مكدّس الأنماط: من الجذر للحالي. */
  modeStack: { mode: string; ctx?: string }[]
  state: CliState
  history: string[]
}

export interface CliResult {
  /** سطور الخرج — تنعرض تحت السطر المكتوب. */
  output: string[]
  /** سطر السهم `^` لأخطاء الإدخال (ينعرض بمحاذاة السطر المكتوب). */
  caretAt?: number
  session: CliSession
}

// ═══ مسارات الحالة ═══
function setPath(state: CliState, path: string, value: unknown) {
  const parts = path.split('.')
  let cur = state as Record<string, unknown>
  for (let i = 0; i < parts.length - 1; i++) {
    if (typeof cur[parts[i]] !== 'object' || cur[parts[i]] === null) cur[parts[i]] = {}
    cur = cur[parts[i]] as Record<string, unknown>
  }
  cur[parts[parts.length - 1]] = value
}

function unsetPath(state: CliState, path: string) {
  const parts = path.split('.')
  let cur = state as Record<string, unknown>
  for (let i = 0; i < parts.length - 1; i++) {
    const nxt = cur[parts[i]]
    if (typeof nxt !== 'object' || nxt === null) return
    cur = nxt as Record<string, unknown>
  }
  delete cur[parts[parts.length - 1]]
}

export function getPath(state: CliState, path: string): unknown {
  return path.split('.').reduce<unknown>(
    (acc, k) => (typeof acc === 'object' && acc !== null ? (acc as Record<string, unknown>)[k] : undefined),
    state,
  )
}

// ═══ الوسائط ═══
const IPV4 = /^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})$/

export function argMatches(kind: ArgKind, token: string): boolean {
  switch (kind) {
    case 'num':
      return /^\d+$/.test(token)
    case 'ipv4':
    case 'mask': {
      const m = IPV4.exec(token)
      return !!m && m.slice(1).every((o) => Number(o) <= 255)
    }
    case 'ifname':
      // اسم منفذ: حروف ثم أرقام بشرطات مائلة — `GigabitEthernet0/1`, `Gi0/1`, `Fa0/24`
      return /^[A-Za-z][A-Za-z-]*\d+(\/\d+)*$/.test(token)
    case 'vlanlist':
      return /^\d+([,-]\d+)*$/.test(token)
    case 'word':
    default:
      return token.length > 0
  }
}

// ═══ توسيع اسم المنفذ ═══
//
// ⚠️ الجهاز الحقيقي يقبل `gi0/1` ويخزنها **موسّعة**
// `GigabitEthernet0/1`، وكل خرج `show` يطلعها موسّعة. بدون هالتوسيع
// `show run` تطلع `interface gi0/1` — وأي فني شبكات يشوفها يعرف إنه
// مو جهاز حقيقي بثانية. تفصيلة صغيرة بس هي الي تفرّق.
// ⚠️ الترتيب مهم: `te` قبل `t`، و`ge` قبل `g`. والحروف المفردة
// (`g`, `f`, `e`) لازم تنقبل — فني هواوي يكتب `int g0/0/1` عادةً،
// وبدونها الاسم ينخزن مختصراً و`display` تطلعه مختصراً فتنكشف اللعبة.
const IF_FAMILIES: [RegExp, string][] = [
  [/^(te|ten|xge|tengig(abit)?(ethernet)?)$/i, 'TenGigabitEthernet'],
  [/^(g|gi|gig|ge|gigabit(ethernet)?)$/i, 'GigabitEthernet'],
  [/^(f|fa|fast(ethernet)?)$/i, 'FastEthernet'],
  [/^(e|et|eth|ethernet)$/i, 'Ethernet'],
  [/^(vl|vlan)$/i, 'Vlan'],
  [/^(po|port-channel|eth-trunk)$/i, 'Port-channel'],
]

export function expandIfName(token: string): string {
  const m = /^([A-Za-z-]+)(\d.*)$/.exec(token)
  if (!m) return token
  for (const [re, full] of IF_FAMILIES) if (re.test(m[1])) return full + m[2]
  return token
}

/** الاسم الي يطلع بالمساعدة لعقدة وسيط. */
const ARG_LABEL: Record<ArgKind, string> = {
  word: 'WORD',
  num: '<1-4094>',
  ipv4: 'A.B.C.D',
  mask: 'A.B.C.D',
  ifname: 'IFNAME',
  vlanlist: 'VLAN-LIST',
}

export function argLabel(n: CliNode): string {
  return n.t === '<arg>' ? ARG_LABEL[n.arg || 'word'] : n.t
}

// ═══ المطابقة ═══
type Resolution =
  | { kind: 'literal'; node: CliNode }
  | { kind: 'arg'; node: CliNode }
  | { kind: 'ambiguous'; options: string[] }
  | { kind: 'none' }

/**
 * يطابق كلمة واحدة على أبناء عقدة.
 *
 * ⚠️ الترتيب مهم ويطابق الجهاز الحقيقي:
 *   ١) تطابق حرفي كامل يفوز فوراً — حتى لو چان بادئة لأمر أطول
 *      (`show` تفوز مع إن `showbla` موجود).
 *   ٢) بادئة وحدة فقط → توسّع.
 *   ٣) أكثر من بادئة → غامض (**مو** «خذ الأول»).
 *   ٤) ماكو حرفي → جرّب الوسائط.
 */
export function resolve(children: CliNode[], token: string): Resolution {
  const lits = children.filter((c) => c.t !== '<arg>')
  const exact = lits.find((c) => c.t.toLowerCase() === token.toLowerCase())
  if (exact) return { kind: 'literal', node: exact }

  const pref = lits.filter((c) => c.t.toLowerCase().startsWith(token.toLowerCase()))
  if (pref.length === 1) return { kind: 'literal', node: pref[0] }
  if (pref.length > 1) return { kind: 'ambiguous', options: pref.map((c) => c.t) }

  const arg = children.find((c) => c.t === '<arg>' && argMatches(c.arg || 'word', token))
  if (arg) return { kind: 'arg', node: arg }
  return { kind: 'none' }
}

// ═══ المحث ═══
export function prompt(grammar: CliGrammar, session: CliSession): string {
  const host = String(getPath(session.state, 'hostname') ?? 'Switch')
  const top = session.modeStack[session.modeStack.length - 1]
  const mode = grammar.modes.find((m) => m.id === top.mode)
  // القالب الكامل يسبق اللاحقة — يخدم الأنظمة الي **تحيط** الاسم.
  if (mode?.promptTemplate) {
    return mode.promptTemplate.replace(/\$host/g, host).replace(/\$ctx/g, top.ctx ?? '')
  }
  const suffix = (mode?.promptSuffix ?? '>').replace('$ctx', top.ctx ?? '')
  return host + suffix
}

function currentRoot(grammar: CliGrammar, session: CliSession): CliNode[] {
  const top = session.modeStack[session.modeStack.length - 1]
  return grammar.modes.find((m) => m.id === top.mode)?.root ?? []
}

/** يبدّل `$1`..`$9` بالوسائط، و`$ctx` بسياق النمط الحالي. */
function subst(tpl: string, args: string[], ctx: string | undefined): string {
  return tpl
    .replace(/\$ctx/g, ctx ?? '')
    .replace(/\$(\d)/g, (_, d: string) => args[Number(d) - 1] ?? '')
}

/** ═══ التنفيذ ═══ */
export function execute(grammar: CliGrammar, session: CliSession, line: string): CliResult {
  const raw = line.trim()
  const out: string[] = []
  const s: CliSession = {
    modeStack: [...session.modeStack],
    state: structuredClone(session.state),
    history: raw ? [...session.history, raw] : session.history,
  }
  if (!raw) return { output: [], session: s }

  let tokens = raw.split(/\s+/)

  // ⚠️ `no` تنعالج بالمحرّك مو بالشجرة: لو انكتبت بكل فرع لتصير الشجرة
  // ضعفين وأي إضافة لازم تنكتب مرتين — ومنها تجي أخطاء «الأمر موجود
  // بس `no` مالته ما يشتغل».
  let negate = false
  if (tokens.length > 1 && /^no?$/i.test(tokens[0]) && 'no'.startsWith(tokens[0].toLowerCase())) {
    const top = s.modeStack[s.modeStack.length - 1].mode
    if (top.startsWith('config')) {
      negate = true
      tokens = tokens.slice(1)
    }
  }

  let children = currentRoot(grammar, s)
  let node: CliNode | null = null
  const args: string[] = []
  // موقع الحرف الي يبدي بيه كل توكن بالسطر الأصلي — للسهم `^`.
  let cursor = raw.length - tokens.join(' ').length

  for (const tok of tokens) {
    const r = resolve(children, tok)
    if (r.kind === 'none') {
      return { output: ['% Invalid input detected at \'^\' marker.', ''], caretAt: cursor, session: s }
    }
    if (r.kind === 'ambiguous') {
      return { output: [`% Ambiguous command:  "${raw}"`, ''], session: s }
    }
    // ⚠️ يتوسّع **قبل** ما ينخزن بالحالة: لو انخزن مختصراً، `gi0/1`
    // و`GigabitEthernet0/1` يصيرون منفذين مختلفين بالحالة — والفني
    // يهيّئ منفذاً ويشوف إعداده راح.
    if (r.kind === 'arg') args.push(r.node.arg === 'ifname' ? expandIfName(tok) : tok)
    node = r.node
    children = node.children ?? []
    cursor += tok.length + 1
  }

  if (!node) return { output: [], session: s }

  // أمر ناقص: العقدة ما عدها أثر وعدها أبناء إلزاميون.
  const terminal = node.set || node.show || node.enter || node.exit || node.endTo || node.say
  if (!terminal) {
    return { output: ['% Incomplete command.', ''], session: s }
  }

  const ctx = s.modeStack[s.modeStack.length - 1].ctx

  if (node.enter) {
    s.modeStack.push({ mode: node.enter.mode, ctx: node.enter.ctx ? subst(node.enter.ctx, args, ctx) : undefined })
  }
  if (node.exit) {
    if (s.modeStack.length > 1) s.modeStack.pop()
  }
  if (node.endTo) {
    // ⚠️ الوجهة **تنسمّى بالنحو** مو تنستنتج من ترتيب الأنماط. أول
    // نسخة چانت تفترض `modes[1]` — صح بسيسكو (`end` ترجّع لـEXEC
    // المميّز) وغلط بـVRP (`return` ترجّع لنمط **المستخدم**). أي
    // نظام ثالث چان يكسرها من جديد.
    s.modeStack = [{ mode: node.endTo }]
  }
  if (node.set) {
    const path = subst(node.set, args, ctx)
    if (negate || node.unset) unsetPath(s.state, path)
    else {
      const v = node.val ? subst(node.val, args, ctx) : true
      setPath(s.state, path, v === 'true' ? true : v === 'false' ? false : v)
    }
  }
  if (node.say) out.push(subst(node.say, args, ctx), '')
  if (node.show) out.push(...renderShow(node.show, s.state, grammar), '')

  return { output: out, session: s }
}

// ═══ `?` المساعدة السياقية ═══
export function contextHelp(grammar: CliGrammar, session: CliSession, line: string): string[] {
  const raw = line.trimStart()
  const endsWithSpace = /\s$/.test(line) || raw === ''
  const tokens = raw.split(/\s+/).filter(Boolean)
  let children = currentRoot(grammar, session)

  const walk = endsWithSpace ? tokens : tokens.slice(0, -1)
  for (const tok of walk) {
    const r = resolve(children, tok)
    if (r.kind !== 'literal' && r.kind !== 'arg') return ['% Unrecognized command']
    children = r.node.children ?? []
  }

  const partial = endsWithSpace ? '' : (tokens[tokens.length - 1] ?? '')
  const shown = children.filter((c) => c.t === '<arg>' || c.t.toLowerCase().startsWith(partial.toLowerCase()))
  if (shown.length === 0) return ['% Unrecognized command']

  const width = Math.max(...shown.map((c) => argLabel(c).length)) + 2
  return shown.map((c) => `  ${argLabel(c).padEnd(width)}${c.help}`)
}

/** ═══ Tab ═══ يكمّل لو التطابق وحيد. */
export function complete(grammar: CliGrammar, session: CliSession, line: string): string {
  const tokens = line.trimStart().split(/\s+/)
  if (/\s$/.test(line)) return line
  const partial = tokens[tokens.length - 1] ?? ''
  let children = currentRoot(grammar, session)
  for (const tok of tokens.slice(0, -1)) {
    const r = resolve(children, tok)
    if (r.kind !== 'literal' && r.kind !== 'arg') return line
    children = r.node.children ?? []
  }
  const pref = children.filter((c) => c.t !== '<arg>' && c.t.toLowerCase().startsWith(partial.toLowerCase()))
  if (pref.length !== 1) return line
  return [...tokens.slice(0, -1), pref[0].t].join(' ') + ' '
}

// ═══ خرج `show` ═══
//
// ⚠️ ينبني **من الحالة** مو من نص محفوظ: لو انكتب نصاً جاهزاً، الفني
// يغيّر إعداداً ويشوف `show run` ما تتغيّر — وتنكشف اللعبة بثانية.
/** ⚠️ مصدَّرة حتى تبويب «الإعداد» بلوحة الكونسول يعرض **نفس** النص
 *  الي يعرضه أمر `show running-config` بالضبط. لو انكتب مولّد ثانٍ
 *  للتبويب، الاثنان يفترقان بأول تعديل — والمتدرّب يشوف إعدادين
 *  مختلفين لنفس الجهاز وما يعرف منو الصح. */
export function renderShow(kind: string, state: CliState, grammar: CliGrammar): string[] {
  const host = String(getPath(state, 'hostname') ?? 'Switch')
  const vlans = (getPath(state, 'vlans') as Record<string, { name?: string }>) ?? {}
  const ifs = (getPath(state, 'interfaces') as Record<string, Record<string, unknown>>) ?? {}

  if (kind === 'running-config') {
    // ⚠️ نفس الحالة بالضبط تنعرض بلغتين — الجهاز واحد والعرض يختلف.
    if (grammar.showStyle === 'vrp') {
      const L = ['#', `sysname ${host}`, '#']
      for (const [id, v] of Object.entries(vlans)) {
        L.push(`vlan ${id}`)
        if (v?.name) L.push(` description ${v.name}`)
        L.push('#')
      }
      for (const [name, cfg] of Object.entries(ifs)) {
        L.push(`interface ${name}`)
        if (cfg.description) L.push(` description ${String(cfg.description)}`)
        if (cfg.mode) L.push(` port link-type ${String(cfg.mode)}`)
        if (cfg.accessVlan) L.push(` port default vlan ${String(cfg.accessVlan)}`)
        if (cfg.trunkVlans) L.push(` port trunk allow-pass vlan ${String(cfg.trunkVlans)}`)
        if (cfg.shutdown) L.push(' shutdown')
        L.push('#')
      }
      L.push('return')
      return L
    }
    const L = ['Building configuration...', '', `Current configuration : ${400 + Object.keys(ifs).length * 60} bytes`, '!', `hostname ${host}`, '!']
    for (const [id, v] of Object.entries(vlans)) {
      L.push(`vlan ${id}`)
      if (v?.name) L.push(` name ${v.name}`)
      L.push('!')
    }
    for (const [name, cfg] of Object.entries(ifs)) {
      L.push(`interface ${name}`)
      if (cfg.description) L.push(` description ${String(cfg.description)}`)
      if (cfg.mode) L.push(` switchport mode ${String(cfg.mode)}`)
      if (cfg.accessVlan) L.push(` switchport access vlan ${String(cfg.accessVlan)}`)
      if (cfg.trunkVlans) L.push(` switchport trunk allowed vlan ${String(cfg.trunkVlans)}`)
      if (cfg.shutdown) L.push(' shutdown')
      L.push('!')
    }
    L.push('end')
    return L
  }

  if (kind === 'vlan-brief' && grammar.showStyle === 'vrp') {
    const L = ['VID  Type    Ports']
    const all: Record<string, string> = { 1: 'common' }
    for (const id of Object.keys(vlans)) all[id] = 'common'
    for (const id of Object.keys(all).sort((a, b) => Number(a) - Number(b))) {
      const ports = Object.entries(ifs)
        .filter(([, c]) => String(c.accessVlan ?? '1') === id)
        .map(([n]) => shortIf(n))
      L.push(`${id.padEnd(5)}${all[id].padEnd(8)}${ports.join(' ')}`)
    }
    return L
  }

  if (kind === 'vlan-brief') {
    const L = ['VLAN Name                             Status    Ports', '---- -------------------------------- --------- -------------------------------']
    const all: Record<string, string> = { 1: 'default' }
    for (const [id, v] of Object.entries(vlans)) all[id] = v?.name || `VLAN${String(id).padStart(4, '0')}`
    for (const id of Object.keys(all).sort((a, b) => Number(a) - Number(b))) {
      const ports = Object.entries(ifs)
        .filter(([, c]) => String(c.accessVlan ?? '1') === id)
        .map(([n]) => shortIf(n))
      L.push(`${id.padEnd(5)}${all[id].padEnd(33)}active    ${ports.join(', ')}`)
    }
    return L
  }

  if (kind === 'interfaces-status') {
    const L = ['Port      Name               Status       Vlan       Duplex  Speed Type']
    for (const [name, cfg] of Object.entries(ifs)) {
      const st = cfg.shutdown ? 'disabled' : 'connected'
      L.push(
        `${shortIf(name).padEnd(10)}${String(cfg.description ?? '').slice(0, 18).padEnd(19)}${st.padEnd(13)}${String(cfg.accessVlan ?? '1').padEnd(11)}a-full  a-100 10/100BaseTX`,
      )
    }
    return L.length > 1 ? L : [...L, '% No interfaces configured']
  }

  if (kind === 'version') {
    return [
      `${grammar.name} Software, Version ${grammar.os}`,
      'Compiled for training simulation — NOT a vendor image.',
      '',
      `${host} uptime is 4 minutes`,
      'System returned to ROM by power-on',
    ]
  }

  return [`% Unsupported show: ${kind}`]
}

/** `GigabitEthernet0/1` → `Gi0/1` — مثل ما يختصرها الجهاز بجداوله. */
function shortIf(name: string): string {
  const m = /^([A-Za-z])[A-Za-z]*([A-Za-z])?[A-Za-z]*(\d.*)$/.exec(name)
  if (!m) return name
  const two = /^(Gigabit|Fast|Ten)/i.exec(name)
  if (two) return (two[1][0].toUpperCase() + (two[1][1] ?? '').toLowerCase()) + m[3]
  return name
}

export function newSession(grammar: CliGrammar, initialState: CliState = {}): CliSession {
  return { modeStack: [{ mode: grammar.startMode }], state: structuredClone(initialState), history: [] }
}
