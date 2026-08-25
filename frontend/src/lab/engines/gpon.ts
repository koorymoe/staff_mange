// ═══ محرّك الألياف الضوئية GPON ═══
//
// السؤال الي يوگف الفني عند زبون FTTH: **ليش الإنترنت ما يشتغل
// والليف سليم؟** والجواب أغلبه رقم واحد: **القدرة الواصلة للONT**.
//
// الشبكة الضوئية مو مثل النحاس — ما بيها «يشتغل أو ما يشتغل». بيها
// **ميزانية**: الOLT يرسل بقدرة معينة، وكل عنصر بالطريق ياكل جزءاً
// منها (السبليتر، طول الليف، كل وصلة ولحام)، والي يوصل لازم يبقى
// **بين حدّين**:
//   • أقل من الحساسية الدنيا ← الONT ما يسجّل أصلاً
//   • **أعلى** من حد الإشباع ← الONT هم ما يشتغل (وهذا الي ينساه
//     الكل: قدرة زائدة تعمي المستقبل مثل ما تعمي الشمس العين)
//
// ⚠️ درجة الدقة `F1`: ميزانية قدرة بحالة مستقرة. ماكو تشتّت لوني
// ولا انعكاسات ولا OTDR. تكفي لتصميم الشبكة وتشخيص «ما يسجّل»،
// وما تكفي لدرس قياس بالOTDR.
//
// ⚠️⚠️ الأرقام أدناه **قيم قياسية منشورة** لفئة GPON Class B+ —
// مو من كتالوگ موديل بعينه. المحتوى يبقى `verified = FALSE` لحد ما
// يجرّبه فني على شبكة حقيقية.

import type { DomainEngine, LabDoc, SimResult } from '../types'

const num = (v: unknown, d: number) => {
  const n = typeof v === 'number' ? v : parseFloat(String(v))
  return Number.isFinite(n) ? n : d
}
const str = (v: unknown, d = '') => (v === undefined || v === null ? d : String(v))

/** فقد الإدخال للسبليتر بالديسيبل — قيم قياسية منشورة. */
export const SPLIT_LOSS: Record<string, number> = {
  '2': 3.6, '4': 7.3, '8': 10.5, '16': 13.7, '32': 17.0, '64': 20.5,
}

/** توهين الليف أحادي النمط بالديسيبل/كم — النزول يستعمل ١٤٩٠ نانومتر. */
const FIBER_DB_PER_KM = 0.28
/** فقد الوصلة الواحدة (كونكتر) واللحام. */
const CONNECTOR_DB = 0.4
const SPLICE_DB = 0.1

export interface OpticalPath {
  ontId: string
  rxDbm: number
  lossDb: number
  splitters: number
}

/**
 * ═══ الميزانية الضوئية لكل ONT ═══
 *
 * ⚠️ الحساب **تراكمي على المسار** — مو رقماً ثابتاً للشبكة: ONT ورا
 * سبليترين متسلسلين يشوف قدرة أقل بكثير من ONT ورا واحد، حتى لو
 * الاثنان بنفس الشبكة. ولهذا الفني يقيس **عند الONT** مو عند الOLT.
 */
export function opticalPaths(doc: LabDoc): OpticalPath[] {
  const olt = doc.nodes.find((n) => n.partId === 'olt')
  if (!olt) return []
  const txDbm = num(olt.params.txDbm, 3)

  const out: OpticalPath[] = []
  const onts = doc.nodes.filter((n) => n.partId === 'ont')

  for (const ont of onts) {
    // بحث بالعرض من الONT للOLT، وتجميع الفقد بالطريق.
    const seen = new Set<string>([ont.id])
    const queue: { id: string; loss: number; splits: number }[] = [{ id: ont.id, loss: 0, splits: 0 }]
    let best: { loss: number; splits: number } | null = null

    while (queue.length) {
      const cur = queue.shift()!
      if (cur.id === olt.id) { best = { loss: cur.loss, splits: cur.splits }; break }
      for (const l of doc.links) {
        const nxt = l.from.node === cur.id ? l.to.node : l.to.node === cur.id ? l.from.node : null
        if (!nxt || seen.has(nxt)) continue
        const nd = doc.nodes.find((n) => n.id === nxt)
        if (!nd) continue
        seen.add(nxt)

        // ⚠️ فقد الوصلة نفسها: طول الليف + كونكترين بطرفيه.
        const km = num(l.params?.lengthKm, 0.5)
        let add = km * FIBER_DB_PER_KM + CONNECTOR_DB * 2 + num(l.params?.splices, 0) * SPLICE_DB
        let splits = cur.splits
        if (nd.partId === 'splitter') {
          add += SPLIT_LOSS[str(nd.params.ratio, '8')] ?? 10.5
          splits += 1
        }
        queue.push({ id: nxt, loss: cur.loss + add, splits })
      }
    }

    if (best) out.push({ ontId: ont.id, rxDbm: txDbm - best.loss, lossDb: best.loss, splitters: best.splits })
  }
  return out
}

export const gponEngine: DomainEngine = {
  id: 'gpon',
  name: 'محرّك الألياف الضوئية',

  run(doc: LabDoc): SimResult {
    const messages: SimResult['messages'] = []
    const nodeReadings: SimResult['nodeReadings'] = {}
    const linkState: SimResult['linkState'] = {}
    const add = (id: string, text: string, tone?: 'ok' | 'warn' | 'bad') => {
      ;(nodeReadings[id] ??= []).push({ text, tone })
    }
    for (const l of doc.links) linkState[l.id] = 'ok'

    const olt = doc.nodes.find((n) => n.partId === 'olt')
    if (!olt) {
      messages.push({ kind: 'warn', text: 'ماكو OLT — ضيف واحداً وابدي منه الشبكة الضوئية.' })
      return { ok: false, messages, nodeReadings, linkState }
    }

    const txDbm = num(olt.params.txDbm, 3)
    const capacity = num(olt.params.ponCapacity, 64)
    add(olt.id, `TX ${txDbm.toFixed(1)} dBm`)

    const onts = doc.nodes.filter((n) => n.partId === 'ont')
    if (onts.length === 0) {
      messages.push({ kind: 'warn', text: 'ماكو ONT مربوط — ضيف وحدة مشترك.' })
      return { ok: false, messages, nodeReadings, linkState }
    }
    add(olt.id, `${onts.length}/${capacity} مشترك`, onts.length > capacity ? 'bad' : 'ok')

    // ═══ سعة منفذ PON ═══
    if (onts.length > capacity) {
      messages.push({
        kind: 'error',
        text: `🔴 ${onts.length} مشترك على منفذ PON سعته ${capacity}. الزيادة ما تنسجّل أصلاً — والمشتركون الجدد يبقون بلا خدمة والشبكة تبدو «شغّالة».`,
      })
    }

    // ═══ الميزانية الضوئية ═══
    const paths = opticalPaths(doc)
    let anyBad = false

    for (const ont of onts) {
      const path = paths.find((p) => p.ontId === ont.id)
      const nm = str(ont.params.name, 'ONT')
      if (!path) {
        add(ont.id, 'ماكو مسار ضوئي', 'bad')
        anyBad = true
        messages.push({ kind: 'warn', text: `«${nm}» ما عليه ليف موصول للOLT.` })
        continue
      }

      const minRx = num(ont.params.rxMin, -27)
      const maxRx = num(ont.params.rxMax, -8)
      add(ont.id, `RX ${path.rxDbm.toFixed(1)} dBm`,
        path.rxDbm < minRx || path.rxDbm > maxRx ? 'bad' : path.rxDbm < minRx + 3 ? 'warn' : 'ok')
      add(ont.id, `فقد ${path.lossDb.toFixed(1)} dB`)

      if (path.rxDbm < minRx) {
        anyBad = true
        add(ont.id, 'ما يسجّل — إشارة ضعيفة', 'bad')
        messages.push({
          kind: 'error',
          text: `🔴 «${nm}»: القدرة الواصلة ${path.rxDbm.toFixed(1)} dBm وحساسية الONT ${minRx} dBm. **ما يسجّل بالOLT أصلاً** — الضوء ما يكفي. الفقد الكلي ${path.lossDb.toFixed(1)} dB عبر ${path.splitters} سبليتر. قلّل نسبة التقسيم أو قصّر المسافة أو افحص اللحامات.`,
        })
      } else if (path.rxDbm > maxRx) {
        anyBad = true
        add(ont.id, 'إشباع — قدرة زائدة', 'bad')
        messages.push({
          kind: 'error',
          text: `🔴 «${nm}»: القدرة الواصلة ${path.rxDbm.toFixed(1)} dBm **أعلى** من حد الإشباع (${maxRx} dBm). قدرة زائدة تعمي المستقبل مثل ما تعمي الشمس العين — الONT ما يشتغل. هذا يصير بالمسافات القصيرة جداً بلا سبليتر، والحل مخمّد (attenuator).`,
        })
      } else if (path.rxDbm < minRx + 3) {
        messages.push({
          kind: 'warn',
          text: `⚠️ «${nm}»: ${path.rxDbm.toFixed(1)} dBm — ضمن الحد بس بلا هامش (أقل من ٣ dB فوق الحساسية). أي لحام إضافي أو اتساخ كونكتر يطيّح الخدمة، والعطل يطلع «متقطّع» ويصعب تشخيصه.`,
        })
      }

      // ═══ تهيئة الWAN ═══
      //
      // ⚠️ الONT يسجّل ضوئياً وما ينطي إنترنت: التسجيل شي والخدمة شي.
      // فني يشوف الضوء أخضر ويظن خلص — والزبون بلا إنترنت.
      const mode = str(ont.params.wanMode, 'pppoe')
      if (mode === 'pppoe' && !str(ont.params.pppoeUser).trim()) {
        anyBad = true
        add(ont.id, 'PPPoE بلا مستخدم', 'bad')
        messages.push({
          kind: 'error',
          text: `🔴 «${nm}»: وضع WAN على PPPoE بلا اسم مستخدم. الضوء يصير أخضر والONT يسجّل — **بس ماكو إنترنت**. التسجيل الضوئي شي والخدمة شي.`,
        })
      }
      const vlan = num(ont.params.wanVlan, 0)
      const oltVlan = num(olt.params.serviceVlan, 35)
      if (vlan !== oltVlan) {
        anyBad = true
        add(ont.id, `VLAN ${vlan} ≠ ${oltVlan}`, 'bad')
        messages.push({
          kind: 'error',
          text: `🔴 «${nm}»: معرّف VLAN بالONT ${vlan} وخدمة الOLT على ${oltVlan}. الONT مسجّل والضوء أخضر وماكو إنترنت — وهذا أكثر بلاغ «الليف سليم والنت ما يجي».`,
        })
      }
    }

    // ═══ السبليترات ═══
    for (const sp of doc.nodes.filter((n) => n.partId === 'splitter')) {
      const ratio = str(sp.params.ratio, '8')
      const loss = SPLIT_LOSS[ratio] ?? 10.5
      add(sp.id, `1:${ratio}`)
      add(sp.id, `−${loss.toFixed(1)} dB`)
      const outs = doc.links.filter((l) => l.from.node === sp.id || l.to.node === sp.id).length - 1
      if (outs > Number(ratio)) {
        anyBad = true
        add(sp.id, 'مخارج أكثر من النسبة', 'bad')
        messages.push({
          kind: 'error',
          text: `🔴 سبليتر 1:${ratio} عليه ${outs} مخرج — أكثر من مخارجه الفعلية.`,
        })
      }
    }

    if (!anyBad) {
      const worst = paths.reduce((w, p) => (p.rxDbm < w.rxDbm ? p : w), paths[0])
      messages.push({
        kind: 'info',
        text: `✅ كل الوحدات ضمن الميزانية الضوئية. أضعف مسار: ${worst.rxDbm.toFixed(1)} dBm بفقد ${worst.lossDb.toFixed(1)} dB.`,
      })
    }
    for (const l of doc.links) if (anyBad) linkState[l.id] = 'bad'

    return { ok: !anyBad, messages, nodeReadings, linkState }
  },
}
