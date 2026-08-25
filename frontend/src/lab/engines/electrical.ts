// ═══ محرّك الدوائر الكهربائية ═══
//
// «كأنما آني فاتح ماتلاب وجاي أربط دوائر كهربائية».
//
// الي يفرّق بين محاكي حقيقي ورسمة: هنا **تنحل الدائرة فعلاً**. تحليل
// عقدي (Nodal Analysis) — نفس الي وراه SPICE وMATLAB: نبني مصفوفة
// الموصلية G ومتّجه التيارات i، ونحل G·v = i بحذف گاوس.
//
// يعني ماكو جدول جاهز ولا حالات مكتوبة سلفاً: تحط مقاومتين على
// التوازي وتشوف التيار ينقسم عليهن بالنسبة الصحيحة، لأن الرياضيات
// هي الي تقرّر مو آني.
//
// ⚠️ درجة الدقة `F2` بسلّم الـFidelity (٦): عناصر خطّية بحالة
// مستقرة (DC steady-state). ماكو مكثفات ولا ملفّات ولا زمن — وهذا
// يكفي لتدريب التوصيل والحماية وهبوط الجهد. ولمن يجي درس يحتاج
// زمناً، الطبقة تتبدّل بلا ما تتغيّر هوية التمرين.

import type { DomainEngine, LabDoc, PartDef, SimResult } from '../types'

/** موصلية «سلك» — عالية بس مو لانهاية حتى ما تنفجر المصفوفة. */
const G_WIRE = 1e6
/** موصلية «مفتوح» — واطئة بس مو صفر، للسبب نفسه. */
const G_OPEN = 1e-12
/** ⚠️ GMIN: موصلية ضئيلة من كل عقدة للأرضي. بدونها أي جزء **عائم**
 *  (مو مربوط بالأرضي بأي طريق) يخلّي المصفوفة **شاذّة** والحل ينهار
 *  بقسمة على صفر. هاي حيلة قياسية بكل محاكيات الدوائر. */
const GMIN = 1e-9

// ═══ اتحاد الشبكات ═══
//
// المنافذ المربوطة بوصلات تصير **عقدة كهربائية وحدة** (net). هذا
// union-find بسيط — بدونه كل سلك يصير مقاومة وهمية والحل يتشوّه.
class Nets {
  private parent = new Map<string, string>()
  find(x: string): string {
    const p = this.parent.get(x)
    if (p === undefined) { this.parent.set(x, x); return x }
    if (p === x) return x
    const r = this.find(p)
    this.parent.set(x, r)
    return r
  }
  union(a: string, b: string) {
    const ra = this.find(a), rb = this.find(b)
    if (ra !== rb) this.parent.set(ra, rb)
  }
  roots(): string[] {
    return [...new Set([...this.parent.keys()].map((k) => this.find(k)))]
  }
}

/** ═══ حذف گاوس-جوردان مع محور جزئي ═══
 *
 *  المحور الجزئي (اختيار أكبر عنصر بالعمود) مو رفاهية: بدونه أي
 *  دائرة بيها موصليات متباعدة جداً — مفتاح مغلق (١٠⁶) جنب مقاومة
 *  عالية (١٠⁻⁶) — تنهار عددياً وتطلع جهوداً خيالية. */
function backSolve(A: number[][], b: number[]): number[] | null {
  const n = b.length
  const M = A.map((row, i) => [...row, b[i]])
  for (let col = 0; col < n; col++) {
    let piv = col
    for (let r = col + 1; r < n; r++) if (Math.abs(M[r][col]) > Math.abs(M[piv][col])) piv = r
    if (Math.abs(M[piv][col]) < 1e-14) return null
    ;[M[col], M[piv]] = [M[piv], M[col]]
    for (let r = 0; r < n; r++) {
      if (r === col) continue
      const f = M[r][col] / M[col][col]
      if (f === 0) continue
      for (let c = col; c <= n; c++) M[r][c] -= f * M[col][c]
    }
  }
  const x = new Array<number>(n)
  for (let i = 0; i < n; i++) x[i] = M[i][n] / M[i][i]
  return x
}

const num = (v: unknown, d: number) => {
  const n = typeof v === 'number' ? v : parseFloat(String(v))
  return Number.isFinite(n) ? n : d
}

const fmt = (v: number, unit: string, digits = 2) =>
  `${Math.abs(v) < 0.001 && v !== 0 ? v.toExponential(1) : v.toFixed(digits)} ${unit}`

export const electricalEngine: DomainEngine = {
  id: 'electrical',
  name: 'محرّك الدوائر',

  run(doc: LabDoc, catalog: Record<string, PartDef>): SimResult {
    const messages: SimResult['messages'] = []
    const nodeReadings: SimResult['nodeReadings'] = {}
    const linkState: SimResult['linkState'] = {}

    // ═══ ١) بناء الشبكات ═══
    const nets = new Nets()
    const key = (n: string, p: string) => `${n}:${p}`
    for (const n of doc.nodes) {
      const part = catalog[n.partId]
      for (const p of part?.ports ?? []) nets.find(key(n.id, p.id))
    }
    for (const l of doc.links) {
      nets.union(key(l.from.node, l.from.port), key(l.to.node, l.to.port))
      linkState[l.id] = 'off'
    }

    // ═══ ٢) اختيار الأرضي ═══
    // سالب أول مصدر هو المرجع. بلا مرجع ماكو معنى لكلمة «جهد».
    const source = doc.nodes.find((n) => n.partId === 'dc_source')
    if (!source) {
      messages.push({ kind: 'warn', text: 'ماكو مصدر تغذية بالدائرة — ما تنحل. ضيف مصدراً وربطه.' })
      return { ok: false, messages, nodeReadings, linkState }
    }
    const gnd = nets.find(key(source.id, 'neg'))

    const roots = nets.roots().filter((r) => r !== gnd)
    const idx = new Map(roots.map((r, i) => [r, i]))
    const n = roots.length
    if (n === 0) {
      messages.push({ kind: 'warn', text: 'الدائرة مو مكتملة — ماكو ولا مسار.' })
      return { ok: false, messages, nodeReadings, linkState }
    }

    const G = Array.from({ length: n }, () => new Array<number>(n).fill(0))
    const I = new Array<number>(n).fill(0)

    const at = (net: string) => (net === gnd ? -1 : (idx.get(net) ?? -1))
    /** يضيف موصلية بين شبكتين — نمط الوسم القياسي بالتحليل العقدي. */
    const stampG = (na: string, nb: string, g: number) => {
      const a = at(na), b = at(nb)
      if (a >= 0) G[a][a] += g
      if (b >= 0) G[b][b] += g
      if (a >= 0 && b >= 0) { G[a][b] -= g; G[b][a] -= g }
    }
    /** يضيف مصدر تيار من nb إلى na. */
    const stampI = (na: string, nb: string, i: number) => {
      const a = at(na), b = at(nb)
      if (a >= 0) I[a] += i
      if (b >= 0) I[b] -= i
    }

    for (let i = 0; i < n; i++) G[i][i] += GMIN

    // ═══ ٣) وسم العناصر ═══
    //
    // ⚠️ المصدر ينوسم بمكافئ **نورتون** (مصدر تيار موازي موصلية) مو
    // بمصدر جهد مباشر: مصدر الجهد يحتاج معادلة إضافية (MNA) وصف
    // ومحور بالمصفوفة. المقاومة الداخلية موجودة أصلاً بالواقع، فنستغلها
    // — وهذا هم يخلّي «هبوط الجهد تحت الحمل» يطلع لحاله.
    const comps: { id: string; a: string; b: string; g: number; kind: string; part: PartDef }[] = []

    for (const nd of doc.nodes) {
      const part = catalog[nd.partId]
      if (!part) continue
      const P = (id: string) => nets.find(key(nd.id, id))

      if (nd.partId === 'dc_source') {
        const v = num(nd.params.v, 12)
        const rInt = Math.max(1e-4, num(nd.params.rInt, 0.05))
        stampG(P('pos'), P('neg'), 1 / rInt)
        stampI(P('pos'), P('neg'), v / rInt)
        continue
      }
      if (nd.partId === 'resistor') {
        const r = Math.max(1e-6, num(nd.params.r, 100))
        stampG(P('a'), P('b'), 1 / r)
        comps.push({ id: nd.id, a: P('a'), b: P('b'), g: 1 / r, kind: 'resistor', part })
        continue
      }
      if (nd.partId === 'lamp') {
        // مقاومة اللمبة من بياناتها الاسمية: R = V²/P
        const vN = num(nd.params.vNom, 12), pN = Math.max(0.01, num(nd.params.pNom, 5))
        const r = (vN * vN) / pN
        stampG(P('a'), P('b'), 1 / r)
        comps.push({ id: nd.id, a: P('a'), b: P('b'), g: 1 / r, kind: 'lamp', part })
        continue
      }
      if (nd.partId === 'motor') {
        const vN = num(nd.params.vNom, 12), iN = Math.max(0.01, num(nd.params.iNom, 1))
        const r = vN / iN
        stampG(P('a'), P('b'), 1 / r)
        comps.push({ id: nd.id, a: P('a'), b: P('b'), g: 1 / r, kind: 'motor', part })
        continue
      }
      if (nd.partId === 'switch') {
        const g = nd.params.closed ? G_WIRE : G_OPEN
        stampG(P('a'), P('b'), g)
        comps.push({ id: nd.id, a: P('a'), b: P('b'), g, kind: 'switch', part })
        continue
      }
      if (nd.partId === 'fuse') {
        stampG(P('a'), P('b'), G_WIRE)
        comps.push({ id: nd.id, a: P('a'), b: P('b'), g: G_WIRE, kind: 'fuse', part })
        continue
      }
    }

    // ═══ ٤) الحل ═══
    const v = backSolve(G, I)
    if (!v) {
      messages.push({ kind: 'error', text: 'ما گدرت أحل الدائرة — يمكن بيها جزء عائم أو ربط متضارب.' })
      return { ok: false, messages, nodeReadings, linkState }
    }
    const V = (net: string) => (net === gnd ? 0 : v[idx.get(net) ?? -1] ?? 0)

    const netVoltages: Record<string, number> = {}
    for (const r of roots) netVoltages[r] = V(r)
    netVoltages[gnd] = 0

    // ═══ ٥) القراءات والتشخيص ═══
    const add = (id: string, text: string, tone?: 'ok' | 'warn' | 'bad') => {
      ;(nodeReadings[id] ??= []).push({ text, tone })
    }

    // تيار المصدر = مجموع التيارات الخارجة
    const srcV = V(nets.find(key(source.id, 'pos'))) - V(nets.find(key(source.id, 'neg')))
    add(source.id, fmt(srcV, 'V'), Math.abs(srcV - num(source.params.v, 12)) > num(source.params.v, 12) * 0.1 ? 'warn' : 'ok')

    let anyCurrent = false
    for (const c of comps) {
      const dv = V(c.a) - V(c.b)
      const i = dv * c.g
      if (Math.abs(i) > 1e-4) anyCurrent = true

      if (c.kind === 'switch') {
        const closed = doc.nodes.find((x) => x.id === c.id)?.params.closed
        add(c.id, closed ? 'مغلق' : 'مفتوح', closed ? 'ok' : 'warn')
        if (closed) add(c.id, fmt(Math.abs(i), 'A'))
        continue
      }
      if (c.kind === 'fuse') {
        const iMax = num(doc.nodes.find((x) => x.id === c.id)?.params.iMax, 5)
        const blown = Math.abs(i) > iMax
        add(c.id, fmt(Math.abs(i), 'A'), blown ? 'bad' : 'ok')
        if (blown) {
          add(c.id, 'منقطع!', 'bad')
          messages.push({
            kind: 'error',
            text: `الفيوز انقطع: مرّ بيه ${Math.abs(i).toFixed(2)} أمبير وحدّه ${iMax}. بالميدان هذا يعني حملاً زائداً أو قصراً — دوّر السبب قبل ما تكبّر الفيوز.`,
          })
        }
        continue
      }

      const p = Math.abs(dv * i)
      add(c.id, fmt(Math.abs(dv), 'V'))
      add(c.id, fmt(Math.abs(i), 'A', 3))

      if (c.kind === 'lamp') {
        const nd = doc.nodes.find((x) => x.id === c.id)!
        const vN = num(nd.params.vNom, 12), pN = num(nd.params.pNom, 5)
        const ratio = Math.abs(dv) / vN
        if (ratio > 1.15) {
          add(c.id, 'محترقة!', 'bad')
          messages.push({ kind: 'error', text: `اللمبة عليها ${Math.abs(dv).toFixed(1)} فولت وهي ${vN} فولت — تحترق.` })
        } else if (ratio > 0.85) add(c.id, `تضوّي ${Math.round(ratio * 100)}٪`, 'ok')
        else if (ratio > 0.3) add(c.id, `خافتة ${Math.round(ratio * 100)}٪`, 'warn')
        else add(c.id, 'مطفية', 'warn')
        if (p > pN * 1.15) messages.push({ kind: 'warn', text: `اللمبة تسحب ${p.toFixed(1)} واط وقدرتها ${pN}.` })
      }
      if (c.kind === 'motor') {
        const nd = doc.nodes.find((x) => x.id === c.id)!
        const vN = num(nd.params.vNom, 12)
        if (dv < 0) add(c.id, 'يدور بالعكس — القطبية مقلوبة', 'bad')
        else if (Math.abs(dv) < vN * 0.6) add(c.id, 'ما يدور — الجهد ناقص', 'warn')
        else add(c.id, 'يدور', 'ok')
      }
    }

    // حالة الوصلات: الوصلة «شغّالة» إذا شبكتها مو صفراً أو يمر تيار
    for (const l of doc.links) {
      const net = nets.find(key(l.from.node, l.from.port))
      linkState[l.id] = anyCurrent && Math.abs(V(net)) > 1e-6 ? 'ok' : 'off'
    }

    if (!anyCurrent) {
      messages.push({ kind: 'warn', text: 'ماكو تيار يمر — الدائرة مفتوحة. افحص المفاتيح والوصلات.' })
    } else {
      messages.push({ kind: 'info', text: `الدائرة انحلّت: ${roots.length + 1} عقدة كهربائية، ${comps.length} عنصر.` })
    }

    return { ok: true, messages, nodeReadings, linkState, netVoltages }
  },
}
