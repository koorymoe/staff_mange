// ═══ محرّك الصوت والإذاعة ═══
//
// خط الـ١٠٠ فولت يشتغل بحساب بسيط وصارم، والغلط بيه **يحرق المكبّر**
// — وأغلب الفنية يقدّرونه بالعين:
//
// ١) **مجموع تابات السماعات > قدرة المكبّر.** بخط الـ١٠٠ فولت كل
//    سماعة تسحب قدرة **مطبوعة على تابها** (٦ أو ١٢ أو ٢٤ واط)،
//    والمجموع لازم يبقى تحت قدرة المكبّر. القاعدة المتعارفة: ما
//    يتجاوز **٨٠٪** — الـ٢٠٪ الباقية هامش للإقلاع وللتشويه.
// ٢) **بوضع المقاومة المنخفضة**: السماعات على التوازي، والمقاومة
//    المكافئة تنزل كل ما تضيف وحدة. تحت حد المكبّر (٤ أو ٨ أوم)
//    يعني تيار زائد ← فصل أو احتراق.
// ٣) **خلط خط ١٠٠ فولت مع سماعات ٨ أوم** على نفس المخرج. السماعة
//    بلا محوّل على خط ١٠٠ فولت = قصر تقريباً.
// ٤) **هبوط القدرة بالخط الطويل** — نفس معادلة المقاومة الي
//    بـ`sim3d/cable.ts`، **تنعاد استعمالها مو تنكتب من جديد**.
//
// ⚠️ درجة الدقة `F1`: توازن قدرة ومقاومة بحالة مستقرة. ماكو استجابة
// ترددية ولا صوتيات غرفة — تكفي لتصميم الخط وحمايته، وما تكفي لدرس
// جودة صوت.

import { CABLE_GAUGES, cableResistance, type Cable } from '../../sim3d/cable'
import type { DomainEngine, LabDoc, SimResult } from '../types'

const num = (v: unknown, d: number) => {
  const n = typeof v === 'number' ? v : parseFloat(String(v))
  return Number.isFinite(n) ? n : d
}
const str = (v: unknown, d = '') => (v === undefined || v === null ? d : String(v))

/** هامش الأمان — ما نحمّل المكبّر أكثر من ٨٠٪. */
const HEADROOM = 0.8

const SPEAKERS = new Set(['ceiling_speaker', 'horn_speaker'])

export const audioEngine: DomainEngine = {
  id: 'audio',
  name: 'محرّك الصوت',

  run(doc: LabDoc): SimResult {
    const messages: SimResult['messages'] = []
    const nodeReadings: SimResult['nodeReadings'] = {}
    const linkState: SimResult['linkState'] = {}
    const add = (id: string, text: string, tone?: 'ok' | 'warn' | 'bad') => {
      ;(nodeReadings[id] ??= []).push({ text, tone })
    }
    for (const l of doc.links) linkState[l.id] = 'ok'

    const amps = doc.nodes.filter((n) => n.partId === 'amplifier')
    if (amps.length === 0) {
      messages.push({ kind: 'warn', text: 'ماكو مكبّر — ضيف واحداً وربط عليه السماعات.' })
      return { ok: false, messages, nodeReadings, linkState }
    }

    let anyBad = false

    for (const amp of amps) {
      const ampName = str(amp.params.name, 'AMP')
      const pRated = num(amp.params.pRated, 120)
      const mode = str(amp.params.mode, '100v')
      const minOhm = num(amp.params.minOhm, 4)

      // السماعات المربوطة بمخرج هالمكبّر (مباشرة أو عبر وحدة مناطق)
      const reach = new Set<string>([amp.id])
      const queue = [amp.id]
      while (queue.length) {
        const cur = queue.shift()!
        for (const l of doc.links) {
          const nxt = l.from.node === cur ? l.to.node : l.to.node === cur ? l.from.node : null
          if (!nxt || reach.has(nxt)) continue
          const nd = doc.nodes.find((n) => n.id === nxt)
          // ⚠️ ما نعبر من مكبّر لمكبّر: كل مكبّر يحسب حمله لحاله،
          // ووصل مخرجين ببعض غلط أصلاً (يحرق الاثنين).
          if (nd?.partId === 'amplifier') {
            anyBad = true
            messages.push({
              kind: 'error',
              text: `🔥 مخرجا مكبّرين مربوطين ببعض — هذا يحرق الاثنين. كل مكبّر إله خطه.`,
            })
            continue
          }
          reach.add(nxt)
          queue.push(nxt)
        }
      }

      const spk = doc.nodes.filter((n) => SPEAKERS.has(n.partId) && reach.has(n.id))
      if (spk.length === 0) {
        add(amp.id, 'ماكو سماعات', 'warn')
        messages.push({ kind: 'warn', text: `«${ampName}» ماكو عليه سماعات.` })
        continue
      }

      // ═══ خلط الأوضاع ═══
      const wrongMode = spk.filter((s) => str(s.params.mode, '100v') !== mode)
      if (wrongMode.length > 0) {
        anyBad = true
        add(amp.id, 'خلط أوضاع!', 'bad')
        messages.push({
          kind: 'error',
          text: `🔥 «${ampName}» بوضع ${mode === '100v' ? 'خط ١٠٠ فولت' : 'مقاومة منخفضة'} وعليه ${wrongMode.length} سماعة بالوضع الثاني. ` +
            (mode === '100v'
              ? 'سماعة ٨ أوم بلا محوّل على خط ١٠٠ فولت = قصر تقريباً على المكبّر.'
              : 'سماعة بمحوّل ١٠٠ فولت على مخرج مقاومة منخفضة تعطي صوتاً واطئاً جداً وحمل غلط.'),
        })
      }

      if (mode === '100v') {
        // ═══ مجموع التابات ═══
        const totalW = spk.reduce((s, x) => s + num(x.params.tapW, 6), 0)
        const pct = (totalW / pRated) * 100
        add(amp.id, `${totalW} / ${pRated} W`, totalW > pRated ? 'bad' : totalW > pRated * HEADROOM ? 'warn' : 'ok')
        add(amp.id, `${spk.length} سماعة`)

        if (totalW > pRated) {
          anyBad = true
          messages.push({
            kind: 'error',
            text: `🔥 مجموع تابات السماعات ${totalW} واط وقدرة «${ampName}» ${pRated} واط (${pct.toFixed(0)}٪). المكبّر يفصل أو يحترق — قلّل التابات أو وزّع على مكبّر ثاني.`,
          })
        } else if (totalW > pRated * HEADROOM) {
          messages.push({
            kind: 'warn',
            text: `⚠️ التحميل ${pct.toFixed(0)}٪ من قدرة «${ampName}». القاعدة المتعارفة **٨٠٪ كحد أعلى** — الباقي هامش للإقلاع وللتشويه.`,
          })
        } else {
          messages.push({
            kind: 'info',
            text: `✅ «${ampName}»: ${totalW} واط من ${pRated} (${pct.toFixed(0)}٪) — داخل الهامش الآمن.`,
          })
        }
      } else {
        // ═══ المقاومة المكافئة بالتوازي ═══
        //
        // ⚠️ 1/Rt = Σ(1/Ri). الفني الي يظن «سماعتين ٨ أوم = ١٦» يحرق
        // مكبّره — التوازي **ينزّل** المقاومة مو يرفعها.
        const inv = spk.reduce((s, x) => s + 1 / Math.max(0.5, num(x.params.ohm, 8)), 0)
        const rEq = 1 / inv
        add(amp.id, `${rEq.toFixed(1)} Ω`, rEq < minOhm ? 'bad' : 'ok')
        add(amp.id, `${spk.length} سماعة`)
        if (rEq < minOhm) {
          anyBad = true
          messages.push({
            kind: 'error',
            text: `🔥 المقاومة المكافئة ${rEq.toFixed(1)} أوم وحد «${ampName}» ${minOhm} أوم. ${spk.length} سماعة على التوازي **تنزّل** المقاومة مو ترفعها — تيار زائد وفصل أو احتراق.`,
          })
        } else {
          messages.push({
            kind: 'info',
            text: `✅ «${ampName}»: ${spk.length} سماعة على التوازي = ${rEq.toFixed(1)} أوم، فوگ الحد (${minOhm}).`,
          })
        }
      }

      // ═══ هبوط الخط — **بخط الـ١٠٠ فولت بس** ═══
      //
      // ⚠️ يعيد استعمال `cableResistance` من محرّك الأسلاك — نفس
      // المعادلة بمحل واحد. نسخها هنا يعني معادلتين تفترقان بأول
      // تصحيح.
      //
      // ⚠️⚠️ **ما ينحسب بوضع المقاومة المنخفضة**: هناك «التاب» ما إله
      // معنى — السماعة تسحب حسب مقاومتها وقدرة المكبّر، مو حسب رقم
      // مطبوع عليها. أول نسخة چانت تحسبه من التاب وتطلع «خسارة ١٧٪»
      // لخط ٣٠ متر، وهذا **رقم غلط**. ورقم غلط أسوأ من ما اكو رقم:
      // الفني يزيد المقطع بلا سبب، ويفقد ثقته بباقي القراءات.
      // وعملياً خطوط الـلو-Z قصيرة، والمشكلة بيها المقاومة مو الخسارة.
      if (mode !== '100v') continue
      for (const l of doc.links) {
        const touchesAmp = l.from.node === amp.id || l.to.node === amp.id
        if (!touchesAmp || !l.params) continue
        const lengthM = num(l.params.lengthM, 15)
        const gauge = str(l.params.gauge, 'mm15')
        const cable: Cable = {
          id: l.id, fromRef: '', fromTerminal: '', toRef: '', toTerminal: '',
          gauge: gauge as Cable['gauge'], lengthM, colorHex: '#94a3b8',
        }
        const r = cableResistance(cable)
        // خط ١٠٠ فولت: التيار = القدرة / ١٠٠ — وهذا سبب اختراع الخط
        // أصلاً (تيار واطئ ← خسارة واطئة ← أسلاك رفيعة تكفي).
        const totalW = spk.reduce((s, x) => s + num(x.params.tapW, 6), 0)
        // تيار خط الـ١٠٠ فولت = القدرة ÷ ١٠٠ — وهذا سبب اختراعه أصلاً.
        const i = totalW / 100
        const lossW = i * i * r
        const lossPct = totalW > 0 ? (lossW / totalW) * 100 : 0
        if (lossPct > 10) {
          messages.push({
            kind: 'warn',
            text: `⚠️ خسارة الخط ${lossPct.toFixed(0)}٪ (${lossW.toFixed(1)} واط) على ${lengthM} متر بمقطع ${CABLE_GAUGES.find((g) => g.id === gauge)?.label ?? gauge}. فوگ ١٠٪ يعني صوتاً واطئاً بآخر الخط — زيد المقطع.`,
          })
        }
      }
    }

    // ═══ السماعات نفسها ═══
    for (const s of doc.nodes.filter((n) => SPEAKERS.has(n.partId))) {
      const connected = doc.links.some((l) => l.from.node === s.id || l.to.node === s.id)
      if (!connected) {
        add(s.id, 'مو مربوطة', 'warn')
        continue
      }
      add(s.id, str(s.params.mode, '100v') === '100v' ? `تاب ${num(s.params.tapW, 6)} W` : `${num(s.params.ohm, 8)} Ω`)
    }

    if (anyBad) for (const l of doc.links) linkState[l.id] = 'bad'
    return { ok: !anyBad, messages, nodeReadings, linkState }
  },
}
