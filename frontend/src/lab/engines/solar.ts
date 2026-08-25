// ═══ محرّك الطاقة الشمسية ═══
//
// ثلاثة أشياء تحرق منظومات بالميدان عدنا، والمحرّك يفحصهن كلهن:
//
// ١) **PV بمدخل البطارية.** المداخل متجاورة وتشبه بعضها بأغلب
//    الموديلات، وستring بـ٣٠٠ فولت على مدخل بطارية ٤٨ فولت يعني
//    إنفرتر محروق بثانية وبلا ضمان.
// ٢) **Voc بالبرد.** الفني يحسب الستring على Vmp ويطلع تمام
//    بالصيف — وبأول صباح شتوي بارد الجهد يرتفع فوگ حد الإنفرتر.
//    الحساب لازم يكون على **Voc** مو على Vmp.
// ٣) **بنك بطاريات مو مطابق.** ٢٤ فولت على إنفرتر مضبوط ٤٨ يعني
//    ما يشحن ولا يشتغل، والزبون يظن الألواح خربانة.
//
// ⚠️ درجة الدقة `F1`: توازن قدرة بحالة مستقرة. ماكو منحنى IV ولا
// MPPT يتتبّع لحظياً ولا حرارة ديناميكية — يكفي لفحص التصميم
// والتوصيل، وما يكفي لدرس كفاءة MPPT.

import type { DomainEngine, LabDoc, SimResult } from '../types'

const num = (v: unknown, d: number) => {
  const n = typeof v === 'number' ? v : parseFloat(String(v))
  return Number.isFinite(n) ? n : d
}

/** معامل ارتفاع Voc بالبرد — تقريب شائع: ٠٫٣٪ لكل درجة تحت ٢٥. */
const COLD_C = 0
const VOC_TEMP_COEF = 0.0033

/** ═══ جهد كل منفذ بالمنظومة ═══
 *
 *  ⚠️ هذا الي يخلّي **الأڤوميتر يشتغل**: بلا جهد لكل منفذ، الأداة
 *  تصير رسمة. والقيم محسوبة من نفس منطق المحرّك — مو جدولاً ثانياً
 *  يفترق عنه.
 *
 *  ⚠️ والقياس **بين طرفين** مثل الميدان: الأڤوميتر ما يعطي «جهد
 *  نقطة» — يعطي **فرق** بين مسبارين. لهذا نرجّع جهد كل منفذ نسبةً
 *  لمرجع مشترك، والفرق ينحسب بالطرح.
 */
export function portVoltages(doc: LabDoc): Record<string, number> {
  const v: Record<string, number> = {}
  const inv = doc.nodes.find((n) => n.partId === 'inverter')

  for (const n of doc.nodes) {
    if (n.partId === 'pv_panel') {
      // ⚠️ الجهد المقاس على أطراف اللوح = Vmp لمن يشتغل تحت حمل،
      // و**Voc** لمن يكون مفصولاً. والفرق بينهما هو الي يخدع الفني:
      // يقيس ستring مفصولاً فيشوف رقماً أعلى من الي يتوقعه.
      const count = Math.max(1, num(n.params.count, 1))
      const connected = doc.links.some((l) =>
        (l.from.node === n.id && l.to.node === inv?.id) || (l.to.node === n.id && l.from.node === inv?.id))
      const val = connected ? num(n.params.vmp, 41.5) * count : num(n.params.voc, 49.8) * count
      v[`${n.id}:pos`] = val
      v[`${n.id}:neg`] = 0
    } else if (n.partId === 'battery') {
      const bv = num(n.params.v, 48)
      const soc = num(n.params.soc, 80) / 100
      // جهد البنك يتبع حالة الشحن — تقريب خطّي شائع ±١٠٪.
      v[`${n.id}:pos`] = bv * (0.9 + 0.2 * soc)
      v[`${n.id}:neg`] = 0
    } else if (n.partId === 'inverter') {
      const bat = doc.nodes.find((x) => x.partId === 'battery')
      const pv = doc.nodes.find((x) => x.partId === 'pv_panel')
      v[`${n.id}:pv_pos`] = pv ? (v[`${pv.id}:pos`] ?? 0) : 0
      v[`${n.id}:pv_neg`] = 0
      v[`${n.id}:bat_pos`] = bat ? (v[`${bat.id}:pos`] ?? 0) : 0
      v[`${n.id}:bat_neg`] = 0
      // مخرج التيار المتناوب — قيمة فعّالة، ما تنقاس بمسبار DC عادي.
      v[`${n.id}:ac_out`] = 230
    } else if (n.partId === 'load') {
      v[`${n.id}:ac_in`] = 230
    }
  }
  return v
}

export const solarEngine: DomainEngine = {
  id: 'solar',
  name: 'محرّك الطاقة الشمسية',

  run(doc: LabDoc): SimResult {
    const messages: SimResult['messages'] = []
    const nodeReadings: SimResult['nodeReadings'] = {}
    const linkState: SimResult['linkState'] = {}
    const add = (id: string, text: string, tone?: 'ok' | 'warn' | 'bad') => {
      ;(nodeReadings[id] ??= []).push({ text, tone })
    }
    for (const l of doc.links) linkState[l.id] = 'off'

    const inv = doc.nodes.find((n) => n.partId === 'inverter')
    if (!inv) {
      messages.push({ kind: 'warn', text: 'ماكو إنفرتر — ضيف واحداً وربط عليه الألواح والبطارية والأحمال.' })
      return { ok: false, messages, nodeReadings, linkState }
    }

    /** منو مربوط بمنفذ الإنفرتر الفلاني؟ */
    const peers = (port: string) =>
      doc.links
        .filter((l) => (l.from.node === inv.id && l.from.port === port) || (l.to.node === inv.id && l.to.port === port))
        .map((l) => (l.from.node === inv.id ? l.to : l.from))

    const mpptMin = num(inv.params.mpptMin, 120)
    const mpptMax = num(inv.params.mpptMax, 450)
    const vdcMax = num(inv.params.vdcMax, 500)
    const pRated = num(inv.params.pRated, 5000)
    const batV = num(inv.params.batV, 48)

    // ═══ ١) الغلط القاتل: PV بمدخل البطارية ═══
    for (const port of ['bat_pos', 'bat_neg']) {
      for (const p of peers(port)) {
        const peer = doc.nodes.find((n) => n.id === p.node)
        if (peer?.partId === 'pv_panel') {
          add(inv.id, 'PV بمدخل البطارية!', 'bad')
          messages.push({
            kind: 'error',
            text: '🔥 الألواح مربوطة بمدخل **البطارية**. بالميدان هذا يحرق الإنفرتر فوراً وما ينفع الضمان — المداخل متجاورة وتشبه بعضها، اقرا الطبع الي فوگ الطرف قبل ما تسنّب.',
          })
        }
      }
    }
    for (const port of ['pv_pos', 'pv_neg']) {
      for (const p of peers(port)) {
        const peer = doc.nodes.find((n) => n.id === p.node)
        if (peer?.partId === 'battery') {
          add(inv.id, 'بطارية بمدخل PV!', 'bad')
          messages.push({ kind: 'error', text: '🔥 البطارية مربوطة بمدخل الألواح — غلط قاتل بنفس الخطورة.' })
        }
      }
    }

    // ═══ ٢) الستring ═══
    const pvNodes = doc.nodes.filter((n) => n.partId === 'pv_panel')
    const pvConnected = pvNodes.filter((n) =>
      doc.links.some((l) =>
        (l.from.node === n.id && l.to.node === inv.id && l.to.port.startsWith('pv')) ||
        (l.to.node === n.id && l.from.node === inv.id && l.from.port.startsWith('pv'))))

    let pvPower = 0
    for (const pv of pvConnected) {
      const count = Math.max(1, num(pv.params.count, 1))
      const vmp = num(pv.params.vmp, 41.5) * count
      const voc = num(pv.params.voc, 49.8) * count
      // ⚠️ الحساب على **Voc بالبرد** مو على Vmp — هذا الفرق الي يحرق.
      const vocCold = voc * (1 + VOC_TEMP_COEF * (25 - COLD_C))
      pvPower += num(pv.params.pmax, 550) * count

      add(pv.id, `Vmp ${vmp.toFixed(0)} V`)
      add(pv.id, `Voc البرد ${vocCold.toFixed(0)} V`, vocCold > vdcMax ? 'bad' : 'ok')

      if (vocCold > vdcMax) {
        messages.push({
          kind: 'error',
          text: `🔥 الستring ${count} ألواح يعطي Voc ${vocCold.toFixed(0)} فولت ببرد الشتاء، وحد الإنفرتر ${vdcMax}. يحرقه بأول صباح بارد — قلّل الألواح بالسلسلة.`,
        })
      } else if (vmp > mpptMax) {
        messages.push({ kind: 'warn', text: `Vmp ${vmp.toFixed(0)} فوگ نافذة MPPT (${mpptMax}) — الإنفرتر ما يشتغل بأعلى كفاءة.` })
      } else if (vmp < mpptMin) {
        messages.push({
          kind: 'warn',
          text: `Vmp ${vmp.toFixed(0)} تحت أدنى نافذة MPPT (${mpptMin}) — الإنفرتر ما يبدي بالغيم أو الصبح. زيد الألواح بالسلسلة.`,
        })
      } else {
        messages.push({ kind: 'info', text: `الستring داخل نافذة MPPT (${mpptMin}–${mpptMax}) و Voc البرد ${vocCold.toFixed(0)} تحت الحد ${vdcMax}. ✅` })
      }
    }
    if (pvConnected.length === 0) messages.push({ kind: 'warn', text: 'ماكو ألواح مربوطة بمدخل PV.' })

    // ═══ ٣) البطارية ═══
    const batNodes = doc.nodes.filter((n) => n.partId === 'battery')
    const batConnected = batNodes.filter((n) =>
      doc.links.some((l) =>
        (l.from.node === n.id && l.to.node === inv.id && l.to.port.startsWith('bat')) ||
        (l.to.node === n.id && l.from.node === inv.id && l.from.port.startsWith('bat'))))
    let usableWh = 0
    for (const b of batConnected) {
      const bv = num(b.params.v, 48)
      const ah = num(b.params.ah, 100)
      const soc = num(b.params.soc, 80) / 100
      const dod = num(b.params.dod, 80) / 100
      usableWh += bv * ah * Math.min(soc, dod)
      add(b.id, `${bv} V · ${ah} Ah`)
      if (bv !== batV) {
        add(b.id, 'جهد مو مطابق!', 'bad')
        messages.push({
          kind: 'error',
          text: `بنك البطاريات ${bv} فولت والإنفرتر مضبوط على ${batV}. ما يشحن ولا يشتغل — والزبون يظن الألواح خربانة.`,
        })
      }
    }
    if (batConnected.length === 0) messages.push({ kind: 'warn', text: 'ماكو بطارية مربوطة — ماكو احتياطي لمن تنقطع الكهرباء.' })

    // ═══ ٤) الأحمال ═══
    const loads = doc.nodes.filter((n) => n.partId === 'load')
    const loadConnected = loads.filter((n) =>
      doc.links.some((l) =>
        (l.from.node === n.id && l.to.node === inv.id && l.to.port === 'ac_out') ||
        (l.to.node === n.id && l.from.node === inv.id && l.from.port === 'ac_out')))
    let loadW = 0, loadWh = 0
    for (const ld of loadConnected) {
      const p = num(ld.params.p, 0)
      loadW += p
      loadWh += p * num(ld.params.hours, 0)
      add(ld.id, `${p} W`)
    }

    add(inv.id, `PV ${pvPower} W`)
    add(inv.id, `حمل ${loadW} W`, loadW > pRated ? 'bad' : 'ok')

    if (loadW > pRated) {
      messages.push({ kind: 'error', text: `الأحمال ${loadW} واط وقدرة الإنفرتر ${pRated} — يفصل على حمل زائد.` })
    } else if (loadW > pRated * 0.8) {
      messages.push({ kind: 'warn', text: `الأحمال ${loadW} واط يعني ${Math.round((loadW / pRated) * 100)}٪ من قدرة الإنفرتر — ماكو هامش للإقلاع.` })
    }

    if (usableWh > 0 && loadW > 0) {
      const hours = usableWh / loadW
      add(inv.id, `احتياطي ${hours.toFixed(1)} س`, hours < 2 ? 'warn' : 'ok')
      messages.push({
        kind: hours < 2 ? 'warn' : 'info',
        text: `البطارية تغطّي الأحمال ${hours.toFixed(1)} ساعة (${Math.round(usableWh)} واط·ساعة قابلة للاستعمال).`,
      })
    }
    if (loadWh > 0 && pvPower > 0) {
      // إنتاج يومي تقريبي: القدرة × ٤٫٥ ساعة شمس مكافئة (تقريب للعراق) × ٠٫٨ كفاءة
      const dailyWh = pvPower * 4.5 * 0.8
      messages.push({
        kind: dailyWh < loadWh ? 'warn' : 'info',
        text: `الإنتاج اليومي التقريبي ${Math.round(dailyWh)} واط·ساعة والاستهلاك ${Math.round(loadWh)} — ${dailyWh < loadWh ? 'ما يكفي، تحتاج ألواحاً أكثر.' : 'يكفي بهامش.'}`,
      })
    }

    for (const l of doc.links) linkState[l.id] = 'ok'
    const bad = messages.some((m) => m.kind === 'error')
    return { ok: !bad, messages, nodeReadings, linkState }
  },
}
