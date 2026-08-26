// ═══ محرّك الشبكة ═══
//
// «تعرف بكت تريسر شلون يطبّقون عليه شبكات؟ أريد مثله».
//
// المحرّك يجاوب السؤال الي يوگف الفني بالميدان: **ليش ما تشتغل؟**
// وأغلب الأجوبة أربعة، وكلها مفحوصة هنا:
//
// ١) الكيبل مو مربوط أو مربوط بمنفذ غلط.
// ٢) العناوين بشبكات فرعية مختلفة (القناع غلط أو العنوان غلط).
// ٣) الأجهزة بـVLAN مختلفة — فيزيائياً موصولة ومنطقياً معزولة.
//    وهذا **أخبث عطل بالشبكات**: كل الأضوية خضر والكيبل سليم.
// ٤) ميزانية PoE مالت السويچ خلصت فالكاميرات تفصل وترجع.
//
// ⚠️ درجة الدقة `F1`: وصولية منطقية بحالة مستقرة. ماكو باكيتات ولا
// ARP ولا STP ولا زمن. تكفي لتشخيص التوصيل والعنونة والعزل، وما
// تكفي لدرس بروتوكولات.
//
// ⚠️⚠️ **منطق الوصولية مصدَّر بـ`netFacts` عمداً.** أدوات الفحص
// (ping · traceroute) تحتاج **نفس** الجواب بالضبط. لو انتسخ المنطق
// بالأداة، أول تصحيح بالمحرّك ما يوصلها: `ping` ينجح والمحاكاة تفشل
// — والمتدرّب ما يعرف منو يصدّق، وهذا أسوأ من ما تكون الأداة موجودة
// أصلاً. مصدر واحد للحقيقة، والمحرّك نفسه أول زبون عنده.

import { expandIfName } from '../../cli/engine'
import { CABLE_BY_ID, checkLink } from '../cables'
import { PART_BY_ID } from '../catalog'
import type { DomainEngine, LabDoc, LabNode, SimResult } from '../types'

const num = (v: unknown, d: number) => {
  const n = typeof v === 'number' ? v : parseFloat(String(v))
  return Number.isFinite(n) ? n : d
}
const str = (v: unknown, d = '') => (v === undefined || v === null ? d : String(v))

function ipToInt(ip: string): number | null {
  const m = /^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})$/.exec(ip.trim())
  if (!m) return null
  const o = m.slice(1).map(Number)
  if (o.some((x) => x > 255)) return null
  return ((o[0] << 24) >>> 0) + (o[1] << 16) + (o[2] << 8) + o[3]
}

function intToIp(n: number): string {
  return [(n >>> 24) & 255, (n >>> 16) & 255, (n >>> 8) & 255, n & 255].join('.')
}

/** أجهزة طرفية لها عنوان — الحاسبات والكاميرات. */
const ENDPOINTS = new Set(['pc', 'ip_camera'])

/** السويچات الي تفهم VLAN — غير المدار ما يفهم. */
const MANAGED = new Set(['switch_l2', 'switch_poe', 'switch_l3'])

/** سبب فشل الاتصال — الأداة تترجمه لعَرَض، والمحرّك يشرحه. */
export type ReachCode = 'ok' | 'nopath' | 'vlan' | 'badip' | 'dupip' | 'route'

export interface ReachVerdict {
  ok: boolean
  code: ReachCode
  kind: 'warn' | 'error'
  /** الشرح الكامل بصيغة المحرّك — جاهز للعرض بالسجل. */
  text: string
}

export interface LinkCheck {
  ok: boolean
  mbps: number
  problems: string[]
  /** مقطوع بعطل محقون مو بخلل تصميم. */
  forced: boolean
  aName: string
  bName: string
  cable: string
}

export interface NetFacts {
  /** الوصلات الميتة — ما تمرّر أي شي. */
  deadLinks: Set<string>
  linkChecks: Map<string, LinkCheck>
  endpoints: LabNode[]
  nameOf: (id: string) => string
  vlan: (id: string) => number
  /** أقصر مسار فيزيائي — قائمة معرّفات القطع، فارغة لو ماكو.
   *  `includeDead` يمرّ حتى على الوصلات الميتة: يفيد لبيان **وين**
   *  انقطع المسار، مو لتقرير الوصولية. */
  pathBetween: (aId: string, bId: string, opts?: { includeDead?: boolean }) => string[]
  /** الراوترات الي يوصلها الجهاز. */
  routersFrom: (fromId: string) => string[]
  /** الحكم الكامل على الاتصال بين طرفين. */
  reach: (aId: string, bId: string) => ReachVerdict
}

/**
 * ═══ حقائق الشبكة ═══
 *
 * ⚠️ تُحسب **مرّة** من المستند، وكل من يسأل يسأل هنا: المحرّك،
 * وأداة `ping`، وأداة `traceroute`. المستند الي يوصلها مفروض يكون
 * مارّاً بـ`withFaults` — الأعطال المحقونة تدخل كخصائص عادية
 * (`forceDown`)، وهاي الدالة ما تعرف بيها ولا لازم تعرف.
 */
export function netFacts(doc: LabDoc): NetFacts {
  const nameOf = (id: string) => {
    const n = doc.nodes.find((x) => x.id === id)
    if (!n) return id
    return str(n.cliState?.hostname ?? n.params.name ?? n.params.hostname, id)
  }

  // ═══ فحص كل كيبل: النوع والطول والترانسيفر ═══
  //
  // ⚠️ هذا الفحص **قبل** أي شي: كيبل مكسور يعني الرابط ما موجود
  // أصلاً، وأي فحص عنونة فوگه يعطي تشخيصاً غلط («الشبكة تمام
  // منطقياً» وهي مقطوعة فيزيائياً).
  const deadLinks = new Set<string>()
  const linkChecks = new Map<string, LinkCheck>()
  const portKind = (nodeId: string, portId: string) => {
    const nd = doc.nodes.find((n) => n.id === nodeId)
    return PART_BY_ID[nd?.partId ?? '']?.ports.find((p) => p.id === portId)?.kind
  }
  for (const l of doc.links) {
    const P = l.params ?? {}
    const cable = String(P.cable ?? 'cat6')
    const chk = checkLink(
      cable,
      Number(P.lengthM ?? 15),
      String(P.sfpA ?? 'none'),
      String(P.sfpB ?? 'none'),
      portKind(l.from.node, l.from.port) === 'sfp',
      portKind(l.to.node, l.to.port) === 'sfp',
    )
    // ⚠️ `forceDown` من الأعطال المحقونة: الكيبل «مقطوع» — المحرّك
    // يعامله مثل أي رابط مكسور بلا ما يعرف إنه عطل مقصود.
    const forced = !!P.forceDown
    if (forced || !chk.ok) deadLinks.add(l.id)
    linkChecks.set(l.id, {
      ok: !forced && chk.ok,
      mbps: chk.mbps,
      problems: chk.problems,
      forced,
      aName: nameOf(l.from.node),
      bName: nameOf(l.to.node),
      cable,
    })
  }

  // ═══ الـVLAN يجي من **منفذ السويچ** ═══
  //
  // ⚠️ هذا الي يربط الكونسول باللوح. الفني يكتب بجلسة السويچ:
  //     interface gi0/2 → switchport access vlan 20
  // ومحرّك الشبكة يقراها من `cliState` مالت السويچ ويحطّ الجهاز
  // المربوط بهذاك المنفذ بـVLAN 20. ماكو خانة «اختر VLAN» بالجهاز
  // الطرفي — لأن ماكو وحدة بالميدان.
  //
  // ⚠️ اسم المنفذ ينوسّع (`gi0/2` ← `GigabitEthernet0/2`) لأن
  // الكونسول يخزنه موسّعاً. بلا التوسيع، كل تهيئة تنكتب باختصار
  // ما تنلگه — والفني يشوف أمره نُفّذ وما تغيّر شي.
  const vlanOf = (endpointId: string): number => {
    for (const l of doc.links) {
      const mine = l.from.node === endpointId ? l.from : l.to.node === endpointId ? l.to : null
      if (!mine) continue
      const other = l.from.node === endpointId ? l.to : l.from
      const sw = doc.nodes.find((n) => n.id === other.node)
      if (!sw || !MANAGED.has(sw.partId)) continue
      const ifs = (sw.cliState?.interfaces ?? {}) as Record<string, Record<string, unknown>>
      const cfg = ifs[expandIfName(other.port)]
      const v = Number(cfg?.accessVlan)
      if (Number.isFinite(v) && v > 0) return v
    }
    return 1
  }
  const vlanCache = new Map<string, number>()
  const vlan = (id: string) => {
    if (!vlanCache.has(id)) vlanCache.set(id, vlanOf(id))
    return vlanCache.get(id)!
  }

  /** ⚠️ الكيبل المكسور **ما يمرّر** — وإلا نقول «فيه مسار» على رابط
   *  ميت، وهذا أسوأ تشخيص ممكن ننطيه. */
  const pathBetween = (aId: string, bId: string, opts?: { includeDead?: boolean }): string[] => {
    if (aId === bId) return [aId]
    const parent = new Map<string, string>()
    const seen = new Set([aId])
    const queue = [aId]
    while (queue.length) {
      const cur = queue.shift()!
      if (cur === bId) {
        const path = [cur]
        let p = parent.get(cur)
        while (p) { path.unshift(p); p = parent.get(p) }
        return path
      }
      for (const l of doc.links) {
        if (!opts?.includeDead && deadLinks.has(l.id)) continue
        const nxt = l.from.node === cur ? l.to.node : l.to.node === cur ? l.from.node : null
        if (nxt && !seen.has(nxt)) { seen.add(nxt); parent.set(nxt, cur); queue.push(nxt) }
      }
    }
    return []
  }

  /** الراوترات الي **يوصلها** الجهاز (بحث بالعرض يمر بالسويچات).
   *
   *  ⚠️ مو «الي بالمسار بين الاثنين» بالضبط: الوصولية بين الطرفين
   *  متأكّدة قبل هالنداء، فأي راوتر يوصله الطرف الأول يصلح مرشّحاً —
   *  والفحص الحقيقي هو تطابق منافذه مع الشبكتين، وهو الي يصفّي. */
  const routersFrom = (fromId: string): string[] => {
    const seen = new Set([fromId])
    const queue = [fromId]
    const routers: string[] = []
    while (queue.length) {
      const cur = queue.shift()!
      const nd = doc.nodes.find((n) => n.id === cur)
      if (nd?.partId === 'router' && cur !== fromId) routers.push(cur)
      for (const l of doc.links) {
        if (deadLinks.has(l.id)) continue
        const nxt = l.from.node === cur ? l.to.node : l.to.node === cur ? l.from.node : null
        if (nxt && !seen.has(nxt)) { seen.add(nxt); queue.push(nxt) }
      }
    }
    return routers
  }

  const routeBetween = (
    aId: string, bId: string, ia: number, ma: number, ib: number, mb: number,
  ): { ok: boolean; kind: 'warn' | 'error'; why: string } => {
    const A = doc.nodes.find((n) => n.id === aId)!
    const B = doc.nodes.find((n) => n.id === bId)!
    const routers = routersFrom(aId)
    if (routers.length === 0) {
      return { ok: false, kind: 'warn', why: 'شبكتان فرعيتان مختلفتان وماكو راوتر بالمسار بينهم.' }
    }
    for (const rid of routers) {
      const R = doc.nodes.find((n) => n.id === rid)!
      const ports = PART_BY_ID[R.partId]?.ports ?? []
      // منافذ الراوتر بعناوينها
      const ifaces = ports
        .map((pt) => ({
          port: pt.id,
          ip: ipToInt(str(R.params[`ip_${pt.id}`])),
          mask: ipToInt(str(R.params[`mask_${pt.id}`], '255.255.255.0')),
        }))
        .filter((x) => x.ip !== null && x.mask !== null)

      const legA = ifaces.find((x) => ((x.ip! & ma) >>> 0) === ((ia & ma) >>> 0))
      const legB = ifaces.find((x) => ((x.ip! & mb) >>> 0) === ((ib & mb) >>> 0))
      if (!legA || !legB) continue

      const gwA = ipToInt(str(A.params.gw))
      const gwB = ipToInt(str(B.params.gw))
      const rn = str(R.params.hostname, 'الراوتر')
      if (gwA !== legA.ip) {
        return {
          ok: false, kind: 'error',
          why: `الراوتر «${rn}» موجود وعنوانه بهالشبكة ${intToIp(legA.ip!)}، بس بوابة ${str(A.params.name)} مضبوطة على ${str(A.params.gw) || '—'}. الجهاز ما يعرف وين يرسل، فيبقى داخل شبكته.`,
        }
      }
      if (gwB !== legB.ip) {
        return {
          ok: false, kind: 'error',
          why: `بوابة ${str(B.params.name)} مضبوطة على ${str(B.params.gw) || '—'} والمفروض ${intToIp(legB.ip!)} (منفذ «${rn}» بشبكته).`,
        }
      }
      return { ok: true, kind: 'warn', why: '' }
    }
    return {
      ok: false, kind: 'error',
      why: 'الراوتر بالمسار بس ماكو منفذ عنده **داخل** إحدى الشبكتين — صحّح عناوين منافذه.',
    }
  }

  /**
   * ═══ الحكم على الاتصال بين طرفين ═══
   *
   * ⚠️ **الترتيب مقصود** — من الفيزيائي للمنطقي: مسار، ثم VLAN، ثم
   * عنونة، ثم توجيه. عكسه يعطي تشخيصاً مضلّلاً: تقول «العنوان غلط»
   * على جهاز كيبله مقطوع أصلاً، والفني يروح يصحّح العنوان ساعة.
   */
  const reach = (aId: string, bId: string): ReachVerdict => {
    const a = doc.nodes.find((n) => n.id === aId)
    const b = doc.nodes.find((n) => n.id === bId)
    if (!a || !b) return { ok: false, code: 'nopath', kind: 'warn', text: 'جهاز مو موجود.' }
    const an = str(a.params.name, a.id), bn = str(b.params.name, b.id)
    const P = (t: string) => `${an} ⇄ ${bn}: ${t}`

    if (pathBetween(aId, bId).length === 0) {
      return { ok: false, code: 'nopath', kind: 'warn', text: P('ماكو مسار — الكيبلات مو موصولة بينهم.') }
    }

    const va = vlan(aId), vb = vlan(bId)
    if (va !== vb) {
      return {
        ok: false, code: 'vlan', kind: 'error',
        text: P(`الكيبل سليم والأضوية خضر، بس منفذ ${an} بـVLAN ${va} ومنفذ ${bn} بـVLAN ${vb} — معزولين منطقياً. هذا أخبث عطل بالشبكات لأن كلشي يبدو تمام.`),
      }
    }

    const ia = ipToInt(str(a.params.ip)), ib = ipToInt(str(b.params.ip))
    const ma = ipToInt(str(a.params.mask, '255.255.255.0')), mb = ipToInt(str(b.params.mask, '255.255.255.0'))
    if (ia === null || ib === null || ma === null || mb === null) {
      return { ok: false, code: 'badip', kind: 'warn', text: P('عنوان أو قناع مو صحيح.') }
    }
    if (ia === ib) {
      return {
        ok: false, code: 'dupip', kind: 'error',
        text: `${an} و${bn} عندهم **نفس العنوان** ${str(a.params.ip)} — تعارض عناوين، الاثنان يتقطّعون.`,
      }
    }
    if (((ia & ma) >>> 0) !== ((ib & mb) >>> 0)) {
      // ═══ شبكتان مختلفتان: لازم راوتر بينهم ═══
      //
      // ⚠️ ماكو «راوتر موجود = تمام». التوجيه يشتغل بثلاثة شروط
      // **كلها**، وأي واحد ناقص يعطي نفس العرض بالضبط (كلشي أخضر
      // وما يوصل):
      //   ١) الراوتر بالمسار بين الاثنين.
      //   ٢) عنده منفذ **داخل** كل شبكة من الشبكتين.
      //   ٣) بوابة كل جهاز تؤشّر على **عنوان منفذ الراوتر بشبكته**.
      // والثالث هو الي ينساه الفني أكثر شي.
      const r = routeBetween(aId, bId, ia, ma, ib, mb)
      if (r.ok) return { ok: true, code: 'ok', kind: 'warn', text: '' }
      return { ok: false, code: 'route', kind: r.kind, text: P(r.why) }
    }
    return { ok: true, code: 'ok', kind: 'warn', text: '' }
  }

  return {
    deadLinks,
    linkChecks,
    endpoints: doc.nodes.filter((n) => ENDPOINTS.has(n.partId)),
    nameOf,
    vlan,
    pathBetween,
    routersFrom,
    reach,
  }
}

export const networkEngine: DomainEngine = {
  id: 'network',
  name: 'محرّك الشبكة',

  run(doc: LabDoc): SimResult {
    const messages: SimResult['messages'] = []
    const nodeReadings: SimResult['nodeReadings'] = {}
    const linkState: SimResult['linkState'] = {}
    const add = (id: string, text: string, tone?: 'ok' | 'warn' | 'bad') => {
      ;(nodeReadings[id] ??= []).push({ text, tone })
    }

    // ⚠️ **نفس** الحقائق الي تشوفها أدوات الفحص — مو حساباً موازياً.
    const F = netFacts(doc)

    // ═══ ١) حالة كل كيبل ═══
    for (const l of doc.links) {
      const c = F.linkChecks.get(l.id)!
      linkState[l.id] = c.ok ? 'ok' : 'bad'
      if (c.forced) {
        messages.push({ kind: 'error', text: `كيبل ${c.aName} ⇄ ${c.bName}: الرابط مقطوع — ماكو إشارة.` })
      } else if (!c.ok) {
        for (const pr of c.problems) {
          messages.push({ kind: 'error', text: `كيبل ${c.aName} ⇄ ${c.bName}: ${pr}` })
        }
      } else if (c.mbps < CABLE_BY_ID[c.cable]?.maxMbps) {
        messages.push({
          kind: 'warn',
          text: `رابط ${c.aName} ⇄ ${c.bName} يشتغل بـ${c.mbps} ميغابت — الترانسيفر أبطأ من الكيبل.`,
        })
      }
    }

    const endpoints = F.endpoints
    const switches = doc.nodes.filter((n) => n.partId.startsWith('switch_'))

    // كل جهاز طرفي: مربوط لو لا
    const linkedNodes = new Set<string>()
    for (const l of doc.links) { linkedNodes.add(l.from.node); linkedNodes.add(l.to.node) }
    for (const e of endpoints) {
      if (!linkedNodes.has(e.id)) {
        add(e.id, 'مو مربوط', 'bad')
        messages.push({ kind: 'warn', text: `«${str(e.params.name, e.id)}» ما عليه كيبل.` })
      }
    }

    // ═══ ٢) ميزانية PoE ═══
    /** اسم السويچ — من الكونسول إذا انهيّأ، وإلا من الخصائص. */
    const swName = (sw: { params: Record<string, unknown>; cliState?: Record<string, unknown> }) =>
      str(sw.cliState?.hostname ?? sw.params.hostname, 'SW')

    for (const sw of switches) {
      const cams = doc.links
        .filter((l) => l.from.node === sw.id || l.to.node === sw.id)
        .map((l) => (l.from.node === sw.id ? l.to.node : l.from.node))
        .map((id) => doc.nodes.find((n) => n.id === id))
        .filter((n) => n?.partId === 'ip_camera')
      const draw = cams.reduce((s, c) => s + num(c!.params.poeW, 0), 0)
      // ⚠️ الميزانية من **خصائص السويچ** مو رقماً ثابتاً: موديل بلا PoE
      // ميزانيته صفر، والكاميرا عليه تحتاج تغذية مستقلة.
      const budget = num(sw.params.poeBudget, 0)
      if (cams.length) {
        if (budget === 0) {
          add(sw.id, 'بلا PoE', 'bad')
          messages.push({
            kind: 'error',
            text: `«${swName(sw)}» ما يدعم PoE و${cams.length} كاميرا مربوطة عليه — الكاميرات ما تشتغل بلا محوّل تغذية لكل وحدة.`,
          })
          continue
        }
        add(sw.id, `PoE ${draw}/${budget} W`, draw > budget ? 'bad' : draw > budget * 0.85 ? 'warn' : 'ok')
        if (draw > budget) {
          messages.push({
            kind: 'error',
            text: `سحب PoE ${draw} واط وميزانية السويچ ${budget}. الكاميرات راح تفصل وترجع بالتناوب — وهذا عطل يضيّع أيام لأن كل كاميرا لحالها تشتغل تمام.`,
          })
        }
      }
    }

    // ═══ ٣) الوصولية بين الأجهزة الطرفية ═══
    //
    // ⚠️ نفحص **كل زوج** مو الزوج الأول: الشبكة ممكن تكون شغّالة بين
    // جهازين ومكسورة بين غيرهم، وتقرير «الشبكة تمام» يخفي هذا.
    let pairsOk = 0, pairsBad = 0
    for (let i = 0; i < endpoints.length; i++) {
      for (let j = i + 1; j < endpoints.length; j++) {
        const r = F.reach(endpoints[i].id, endpoints[j].id)
        if (r.ok) { pairsOk++; continue }
        pairsBad++
        messages.push({ kind: r.kind, text: r.text })
      }
    }

    for (const e of endpoints) {
      if (linkedNodes.has(e.id)) add(e.id, str(e.params.ip, '—'), 'ok')
      add(e.id, `VLAN ${F.vlan(e.id)}`)
    }

    // ═══ السويچ غير المدار يكسر العزل ═══
    for (const sw of doc.nodes.filter((n) => n.partId === 'switch_unmanaged')) {
      const attached = doc.links
        .filter((l) => l.from.node === sw.id || l.to.node === sw.id)
        .map((l) => (l.from.node === sw.id ? l.to.node : l.from.node))
        .map((id) => doc.nodes.find((n) => n.id === id))
        .filter((n) => n && ENDPOINTS.has(n.partId))
      const vlans = new Set(attached.map((n) => F.vlan(n!.id)))
      if (vlans.size > 1) {
        add(sw.id, 'يكسر عزل الـVLAN', 'bad')
        messages.push({
          kind: 'error',
          text: `«${str(sw.params.hostname, sw.id)}» سويچ **غير مدار** وعليه أجهزة بـVLANات مختلفة (${[...vlans].join('، ')}). ما يفهم VLAN — يخلط الشبكات كلها ببعض والعزل الي بنيته ينهار من هنا.`,
        })
      }
    }

    if (endpoints.length < 2) {
      messages.push({ kind: 'info', text: 'حط جهازين على الأقل حتى نفحص الاتصال بينهم.' })
    } else if (pairsBad === 0) {
      messages.push({ kind: 'info', text: `✅ كل الأزواج توصل: ${pairsOk} اتصال ناجح.` })
    } else {
      messages.push({ kind: 'info', text: `${pairsOk} اتصال ناجح و${pairsBad} فاشل.` })
    }

    return { ok: pairsBad === 0, messages, nodeReadings, linkState }
  },
}
