// ═══ محرّك أنظمة إنذار الحريق ═══
//
// ثلاثة أعطال توگف تسليم مشاريع عدنا، وكلها **تنكشف بالحساب**:
//
// ١) **مقاومة النهاية (EOL) ناقصة أو بمحل غلط.** اللوحة تقرا «عطل:
//    دائرة مفتوحة» وتضل تصفّر. الـEOL مو قطعة زينة — اللوحة تمرّر
//    تياراً ضئيلاً بالدائرة وتقيسه: تلگاه ← الخط سليم · ما تلگاه ←
//    الخط مقطوع · يزيد فجأة ← قصر. بلا مقاومة نهاية، اللوحة **ما
//    تگدر تفرّق** بين خط سليم وخط مقطوع.
// ٢) **الصفّارات مخلوطة مع الكواشف بنفس الزون.** يشتغل بالفحص
//    السريع، وبالإنذار الحقيقي اللوحة تحسبها كاشفاً معطّلاً.
// ٣) **البطارية ما تكفي.** القاعدة الشائعة **٢٤ ساعة استعداد + ٣٠
//    دقيقة إنذار**. الفني يحط بطارية «على العين» فتنطفي اللوحة بأول
//    انقطاع طويل — وهذا بالضبط وقت ما تنحتاج.
//
// ⚠️ درجة الدقة `F1`: توبولوجي ومنطق حالة، بلا قياس تيار حقيقي
// بالدائرة. يكفي لتدريب التركيب والتشخيص، وما يكفي لدرس معايرة
// حساسية الكواشف.

import type { DomainEngine, LabDoc, SimResult } from '../types'

const num = (v: unknown, d: number) => {
  const n = typeof v === 'number' ? v : parseFloat(String(v))
  return Number.isFinite(n) ? n : d
}
const str = (v: unknown, d = '') => (v === undefined || v === null ? d : String(v))

/** أجهزة تنحط على زون كواشف. */
const DETECTORS = new Set(['smoke_detector', 'heat_detector', 'mcp'])
/** أجهزة تنحط على دائرة إنذار (صفّارات). */
const SOUNDERS = new Set(['sounder'])

/** ساعات الاستعداد ودقائق الإنذار — القاعدة الشائعة بالمواصفات. */
const STANDBY_H = 24
const ALARM_H = 0.5

export const fireEngine: DomainEngine = {
  id: 'fire',
  name: 'محرّك إنذار الحريق',

  run(doc: LabDoc): SimResult {
    const messages: SimResult['messages'] = []
    const nodeReadings: SimResult['nodeReadings'] = {}
    const linkState: SimResult['linkState'] = {}
    const add = (id: string, text: string, tone?: 'ok' | 'warn' | 'bad') => {
      ;(nodeReadings[id] ??= []).push({ text, tone })
    }
    for (const l of doc.links) linkState[l.id] = 'ok'

    const panel = doc.nodes.find((n) => n.partId === 'fire_panel')
    if (!panel) {
      messages.push({ kind: 'warn', text: 'ماكو لوحة إنذار — ضيف وحدة وابدي منها الزونات.' })
      return { ok: false, messages, nodeReadings, linkState }
    }

    /** يمشي بالسلسلة من منفذ باللوحة ويرجّع الأجهزة بالترتيب.
     *
     *  ⚠️ **يكشف التفريعات**: أول نسخة چانت تمشي بمسار واحد وتترك
     *  أي سلك ثانٍ على نفس الطرف بصمت — فني يربط صفّارة تفريعة من
     *  وسط زون الكواشف والمحاكي يكله «تمام». انكشفت بالفحص العددي.
     *  والتفريعة **غلط حقيقي**: الخط لازم يمشي سلسلة من اللوحة
     *  للمقاومة، وأي تفريعة تكسر مراقبة الخط. */
    const branches: { node: string; port: string; count: number }[] = []
    const walkCircuit = (startPort: string): { nodes: typeof doc.nodes } => {
      const chain: typeof doc.nodes = []
      let curNode = panel.id
      let curPort = startPort
      const usedLinks = new Set<string>()

      for (let guard = 0; guard < 60; guard++) {
        const here = doc.links.filter(
          (l) => !usedLinks.has(l.id) &&
            ((l.from.node === curNode && l.from.port === curPort) ||
             (l.to.node === curNode && l.to.port === curPort)),
        )
        if (here.length > 1) branches.push({ node: curNode, port: curPort, count: here.length })
        const link = here[0]
        if (!link) break
        usedLinks.add(link.id)
        const next = link.from.node === curNode && link.from.port === curPort ? link.to : link.from
        const nd = doc.nodes.find((n) => n.id === next.node)
        if (!nd) break
        chain.push(nd)
        // ⚠️ الجهاز بالسلسلة له منفذان (`in`/`out`): ندخل من واحد
        // ونطلع من الثاني. والمقاومة زيّهم — فلو انحطّت بالنص،
        // السلسلة تكمل وراها والمحرّك يشوف الأجهزة الي «خارج
        // الحماية».
        const outPort = next.port === 'in' ? 'out' : 'in'
        const hasOut = doc.links.some(
          (l) => !usedLinks.has(l.id) &&
            ((l.from.node === nd.id && l.from.port === outPort) ||
             (l.to.node === nd.id && l.to.port === outPort)),
        )
        if (!hasOut) break
        curNode = nd.id
        curPort = outPort
      }
      return { nodes: chain }
    }

    // ═══ الزونات ═══
    const zoneCount = num(panel.params.zones, 2)
    let anyAlarm = false
    let anyFault = false
    let standbyMa = num(panel.params.standbyMa, 60)
    let alarmMa = num(panel.params.alarmMa, 120)

    for (let z = 1; z <= zoneCount; z++) {
      const { nodes: chain } = walkCircuit(`z${z}`)
      if (chain.length === 0) {
        // زون فاضي مو عطل — اللوحة تنعزل زوناتها الفاضية عادةً.
        continue
      }
      const last = chain[chain.length - 1]
      const eolIdx = chain.findIndex((n) => n.partId === 'eol_resistor')
      const detectors = chain.filter((n) => DETECTORS.has(n.partId))
      const soundersHere = chain.filter((n) => SOUNDERS.has(n.partId))

      for (const d of detectors) {
        standbyMa += num(d.params.standbyMa, 0.05)
        alarmMa += num(d.params.alarmMa, 0.5)
        if (d.params.triggered) anyAlarm = true
      }

      if (eolIdx === -1) {
        anyFault = true
        add(panel.id, `زون ${z}: عطل`, 'bad')
        messages.push({
          kind: 'error',
          text: `🔴 زون ${z} بلا **مقاومة نهاية**. اللوحة تقرا «دائرة مفتوحة» وتضل تصفّر — وبدونها ما تگدر تفرّق أصلاً بين خط سليم وخط مقطوع.`,
        })
      } else if (last.partId !== 'eol_resistor') {
        anyFault = true
        add(panel.id, `زون ${z}: EOL بالنص`, 'bad')
        messages.push({
          kind: 'error',
          text: `🔴 زون ${z}: مقاومة النهاية **مو بالنهاية** — عدها ${chain.length - eolIdx - 1} جهاز وراها، وهذولا **خارج الحماية تماماً**. اللوحة ما تحس بيهم لا بالإنذار ولا بالعطل.`,
        })
      } else {
        add(panel.id, `زون ${z}: ${detectors.length} كاشف ✓`, 'ok')
      }

      if (soundersHere.length > 0) {
        anyFault = true
        messages.push({
          kind: 'error',
          text: `🔴 زون ${z} بيه ${soundersHere.length} صفّارة مخلوطة مع الكواشف. الصفّارات إلها **دائرة إنذار منفصلة** — بالخلط اللوحة تحسبها كاشفاً معطّلاً.`,
        })
      }

      // ⚠️ الفحص **بعد** فحص السلسلة: قصر بزون بلا EOL يعطي رسالتين
      // متطابقتين بالمعنى، والفني يظن عنده عطلان.
      const shorted = chain.find((n) => n.params.shorted)
      if (shorted && eolIdx !== -1) {
        anyFault = true
        messages.push({ kind: 'error', text: `🔴 زون ${z}: قصر بالدائرة — اللوحة تقرا «عطل» مو «إنذار».` })
      }
    }

    // ═══ دوائر الإنذار ═══
    const sounderCount = num(panel.params.sounderCircuits, 1)
    let sounderTotal = 0
    for (let s = 1; s <= sounderCount; s++) {
      const { nodes: chain } = walkCircuit(`s${s}`)
      if (chain.length === 0) continue
      const snd = chain.filter((n) => SOUNDERS.has(n.partId))
      const wrongHere = chain.filter((n) => DETECTORS.has(n.partId))
      const last = chain[chain.length - 1]
      for (const x of snd) {
        alarmMa += num(x.params.alarmMa, 20)
        sounderTotal += num(x.params.alarmMa, 20)
      }
      if (wrongHere.length > 0) {
        anyFault = true
        messages.push({
          kind: 'error',
          text: `🔴 دائرة الإنذار ${s} بيها ${wrongHere.length} كاشف. الكواشف على زون كواشف مو على دائرة صفّارات.`,
        })
      }
      if (last.partId !== 'eol_resistor') {
        anyFault = true
        messages.push({
          kind: 'error',
          text: `🔴 دائرة الإنذار ${s} بلا مقاومة نهاية بالنهاية — نفس قاعدة الزونات، اللوحة تراقبها بنفس الطريقة.`,
        })
      } else {
        add(panel.id, `إنذار ${s}: ${snd.length} صفّارة ✓`, 'ok')
      }
    }

    // ═══ ميزانية البطارية ═══
    //
    // ⚠️ الحساب بالأمبير·ساعة: (سحب الاستعداد × ٢٤) + (سحب الإنذار
    // × ٠٫٥). وينضاف عليه هامش ٢٥٪ لتدهور البطارية — بطارية بعمر
    // سنتين ما تعطي سعتها الاسمية، والمواصفات تفرض الهامش.
    const battery = doc.nodes.find((n) => n.partId === 'fire_battery')
    const needAh = ((standbyMa / 1000) * STANDBY_H + (alarmMa / 1000) * ALARM_H) * 1.25
    add(panel.id, `استعداد ${standbyMa.toFixed(0)} mA`)
    add(panel.id, `إنذار ${alarmMa.toFixed(0)} mA`)

    if (!battery) {
      messages.push({
        kind: 'warn',
        text: `ماكو بطارية احتياط. المطلوب تقريباً **${needAh.toFixed(1)} أمبير·ساعة** لتغطية ${STANDBY_H} ساعة استعداد و${ALARM_H * 60} دقيقة إنذار.`,
      })
    } else {
      const have = num(battery.params.ah, 7)
      add(battery.id, `${have} Ah`, have >= needAh ? 'ok' : 'bad')
      add(battery.id, `المطلوب ${needAh.toFixed(1)} Ah`, have >= needAh ? 'ok' : 'bad')
      if (have < needAh) {
        messages.push({
          kind: 'error',
          text: `🔋 البطارية ${have} أمبير·ساعة والمطلوب **${needAh.toFixed(1)}** (${STANDBY_H} ساعة استعداد + ${ALARM_H * 60} دقيقة إنذار + هامش ٢٥٪ للتدهور). تنطفي اللوحة بأول انقطاع طويل — وهذا بالضبط وقت ما تنحتاج.`,
        })
      } else {
        messages.push({
          kind: 'info',
          text: `✅ البطارية ${have} أمبير·ساعة تغطّي المطلوب (${needAh.toFixed(1)}).`,
        })
      }
    }

    if (sounderTotal > 0) messages.push({ kind: 'info', text: `سحب الصفّارات بالإنذار: ${sounderTotal.toFixed(0)} ملّي أمبير.` })

    // ═══ التفريعات ═══
    for (const br of branches) {
      anyFault = true
      const nd = doc.nodes.find((n) => n.id === br.node)
      const nm = nd ? str(nd.params.name ?? nd.params.hostname, nd.partId) : br.node
      messages.push({
        kind: 'error',
        text: `🔴 تفريعة بالخط عند «${nm}» (${br.count} أسلاك على نفس الطرف). خط الإنذار يمشي **سلسلة** من اللوحة لمقاومة النهاية — أي تفريعة تكسر مراقبة الخط، والأجهزة الي عليها ما تنراقب.`,
      })
    }

    // ═══ حالة اللوحة ═══
    const state = anyFault ? 'FAULT' : anyAlarm ? 'ALARM' : 'NORMAL'
    add(panel.id, state === 'FAULT' ? '⚠️ عطل' : state === 'ALARM' ? '🔔 إنذار' : '✅ عادي',
      state === 'FAULT' ? 'bad' : state === 'ALARM' ? 'warn' : 'ok')
    messages.push({
      kind: state === 'FAULT' ? 'warn' : 'info',
      text: `حالة اللوحة «${str(panel.params.name, 'FACP')}»: ${
        state === 'FAULT' ? 'عطل — ما تحمي المبنى بهالحالة' : state === 'ALARM' ? 'إنذار' : 'عادي'}.`,
    })

    for (const l of doc.links) if (anyFault) linkState[l.id] = 'off'
    return { ok: !anyFault, messages, nodeReadings, linkState }
  },
}
