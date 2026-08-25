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

import { CABLE_BY_ID, checkLink } from '../cables'
import { PART_BY_ID } from '../catalog'
import type { DomainEngine, LabDoc, SimResult } from '../types'

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

/** أجهزة طرفية لها عنوان — الحاسبات والكاميرات. */
const ENDPOINTS = new Set(['pc', 'ip_camera'])

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

    // ═══ ١) فحص كل كيبل: النوع والطول والترانسيفر ═══
    //
    // ⚠️ هذا الفحص **قبل** أي شي: كيبل مكسور يعني الرابط ما موجود
    // أصلاً، وأي فحص عنونة فوگه يعطي تشخيصاً غلط («الشبكة تمام
    // منطقياً» وهي مقطوعة فيزيائياً).
    const deadLinks = new Set<string>()
    for (const l of doc.links) {
      const P = l.params ?? {}
      const portKind = (nodeId: string, portId: string) => {
        const nd = doc.nodes.find((n) => n.id === nodeId)
        return PART_BY_ID[nd?.partId ?? '']?.ports.find((p) => p.id === portId)?.kind
      }
      const chk = checkLink(
        String(P.cable ?? 'cat6'),
        Number(P.lengthM ?? 15),
        String(P.sfpA ?? 'none'),
        String(P.sfpB ?? 'none'),
        portKind(l.from.node, l.from.port) === 'sfp',
        portKind(l.to.node, l.to.port) === 'sfp',
      )
      const an = str(doc.nodes.find((n) => n.id === l.from.node)?.params.name
        ?? doc.nodes.find((n) => n.id === l.from.node)?.params.hostname, l.from.node)
      const bn = str(doc.nodes.find((n) => n.id === l.to.node)?.params.name
        ?? doc.nodes.find((n) => n.id === l.to.node)?.params.hostname, l.to.node)
      if (chk.ok) {
        linkState[l.id] = 'ok'
      } else {
        linkState[l.id] = 'bad'
        deadLinks.add(l.id)
        for (const pr of chk.problems) {
          messages.push({ kind: 'error', text: `كيبل ${an} ⇄ ${bn}: ${pr}` })
        }
      }
      if (chk.ok && chk.mbps < CABLE_BY_ID[String(P.cable ?? 'cat6')]?.maxMbps) {
        messages.push({
          kind: 'warn',
          text: `رابط ${an} ⇄ ${bn} يشتغل بـ${chk.mbps} ميغابت — الترانسيفر أبطأ من الكيبل.`,
        })
      }
    }

    const endpoints = doc.nodes.filter((n) => ENDPOINTS.has(n.partId))
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
            text: `«${str(sw.params.hostname, sw.id)}» ما يدعم PoE و${cams.length} كاميرا مربوطة عليه — الكاميرات ما تشتغل بلا محوّل تغذية لكل وحدة.`,
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
        const a = endpoints[i], b = endpoints[j]
        const an = str(a.params.name, a.id), bn = str(b.params.name, b.id)

        // مسار فيزيائي؟ (بحث بالعرض على الوصلات)
        const seen = new Set([a.id])
        const queue = [a.id]
        let reach = false
        while (queue.length) {
          const cur = queue.shift()!
          if (cur === b.id) { reach = true; break }
          for (const l of doc.links) {
            // ⚠️ الكيبل المكسور **ما يمرّر** — وإلا نقول «فيه مسار» على
            // رابط ميت، وهذا أسوأ تشخيص ممكن ننطيه.
            if (deadLinks.has(l.id)) continue
            const nxt = l.from.node === cur ? l.to.node : l.to.node === cur ? l.from.node : null
            if (nxt && !seen.has(nxt)) { seen.add(nxt); queue.push(nxt) }
          }
        }
        if (!reach) {
          pairsBad++
          messages.push({ kind: 'warn', text: `${an} ⇄ ${bn}: ماكو مسار — الكيبلات مو موصولة بينهم.` })
          continue
        }

        // نفس الـVLAN؟
        const va = num(a.params.vlan, 1), vb = num(b.params.vlan, 1)
        if (va !== vb) {
          pairsBad++
          messages.push({
            kind: 'error',
            text: `${an} ⇄ ${bn}: الكيبل سليم والأضوية خضر، بس ${an} بـVLAN ${va} و${bn} بـVLAN ${vb} — معزولين منطقياً. هذا أخبث عطل بالشبكات لأن كلشي يبدو تمام.`,
          })
          continue
        }

        // نفس الشبكة الفرعية؟
        const ia = ipToInt(str(a.params.ip)), ib = ipToInt(str(b.params.ip))
        const ma = ipToInt(str(a.params.mask, '255.255.255.0')), mb = ipToInt(str(b.params.mask, '255.255.255.0'))
        if (ia === null || ib === null || ma === null || mb === null) {
          pairsBad++
          messages.push({ kind: 'warn', text: `${an} ⇄ ${bn}: عنوان أو قناع مو صحيح.` })
          continue
        }
        if (ia === ib) {
          pairsBad++
          messages.push({ kind: 'error', text: `${an} و${bn} عندهم **نفس العنوان** ${str(a.params.ip)} — تعارض عناوين، الاثنان يتقطّعون.` })
          continue
        }
        if (((ia & ma) >>> 0) !== ((ib & mb) >>> 0)) {
          pairsBad++
          messages.push({
            kind: 'warn',
            text: `${an} (${str(a.params.ip)}) و${bn} (${str(b.params.ip)}): شبكتان فرعيتان مختلفتان — يحتاجون راوتر بينهم أو تصحيح العنونة.`,
          })
          continue
        }
        pairsOk++
      }
    }

    for (const e of endpoints) {
      if (linkedNodes.has(e.id)) add(e.id, str(e.params.ip, '—'), 'ok')
      add(e.id, `VLAN ${num(e.params.vlan, 1)}`)
    }

    // ═══ السويچ غير المدار يكسر العزل ═══
    for (const sw of doc.nodes.filter((n) => n.partId === 'switch_unmanaged')) {
      const attached = doc.links
        .filter((l) => l.from.node === sw.id || l.to.node === sw.id)
        .map((l) => (l.from.node === sw.id ? l.to.node : l.from.node))
        .map((id) => doc.nodes.find((n) => n.id === id))
        .filter((n) => n && ENDPOINTS.has(n.partId))
      const vlans = new Set(attached.map((n) => num(n!.params.vlan, 1)))
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
