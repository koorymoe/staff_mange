// ═══ أدوات فحص الشبكة ═══
//
// حقن الأعطال بلا أدوات فحص نصف درس: المتدرّب يشوف رسالة المحاكي
// **تقول له** شنو الخلل. بالميدان ماكو محاكي يحچي — اكو `ping` يرجع
// timeout وفني لازم يستنتج.
//
// ⚠️⚠️ **الأداة تعطي عَرَضاً مو تشخيصاً.** `ping` الفاشل يكتب
// `Request timed out.` — ما يكتب «الـVLAN مختلف». هذا مو تقصيراً:
// هو الدرس نفسه. أداة تكتب الجواب تلغي سبب وجودها، ونرجع لنفس
// المشكلة الي تجنّبناها بوضع التشخيص — عطل مكتوب على الشاشة مو عطل.
//
// ⚠️ والحكم يجي كله من `netFacts` بمحرّك الشبكة — **نفس** المصدر الي
// يقرا منه المحرّك. لو حسبنا هنا حساباً موازياً، أول تصحيح بالمحرّك
// ما يوصل الأداة: `ping` ينجح والمحاكاة تفشل، والمتدرّب ما يعرف منو
// يصدّق.
//
// ⚠️ درجة الدقة `F1`: ماكو نموذج تأخير حقيقي. الأزمنة **تقديرية
// ثابتة** ومكتوب هذا بمخرج كل أداة — رقم `time=` يتغيّر عشوائياً
// يوحي بقياس ما موجود، وهاي كذبة ثانية بمكان الأولى.

import { netFacts, type NetFacts, type ReachCode } from '../engines/network'
import { CABLE_BY_ID } from '../cables'
import type { LabDoc } from '../types'

const str = (v: unknown, d = '') => (v === undefined || v === null ? d : String(v))

/** زمن تقديري لكل قفزة بالمللي ثانية — ثابت عمداً. */
const HOP_MS = 1
const LAN_MS = 1

export interface ToolOutput {
  /** أسطر المخرج بشكل الجهاز الحقيقي. */
  lines: string[]
  ok: boolean
}

const ipOf = (doc: LabDoc, id: string) =>
  str(doc.nodes.find((n) => n.id === id)?.params.ip, '0.0.0.0')

const labelOf = (doc: LabDoc, id: string) => {
  const n = doc.nodes.find((x) => x.id === id)
  return str(n?.cliState?.hostname ?? n?.params.name ?? n?.params.hostname, id)
}

/**
 * ═══ ping ═══
 *
 * ⚠️ الفشل بكل أسبابه يطلع **بنفس** المخرج تقريباً — لأن هيچ هو
 * بالحقيقة. الـVLAN المختلف والبوابة الغلط والكيبل المقطوع كلهن
 * يعطون `Request timed out`، وهاي بالضبط ليش تشخيص الشبكات صعب.
 *
 * الاستثناء الوحيد: عنوان مو صالح على الجهاز نفسه — هنا الجهاز
 * الحقيقي يفشل **قبل** ما يرسل، ويكتب رسالة مختلفة.
 */
export function ping(doc: LabDoc, fromId: string, toId: string, count = 4): ToolOutput {
  const F = netFacts(doc)
  return pingWith(F, doc, fromId, toId, count)
}

function pingWith(F: NetFacts, doc: LabDoc, fromId: string, toId: string, count: number): ToolOutput {
  const dst = ipOf(doc, toId)
  const v = F.reach(fromId, toId)
  const lines: string[] = [`${labelOf(doc, fromId)}> ping ${dst}`, '']

  if (v.code === 'badip') {
    lines.push(`PING: transmit failed. General failure.`, '')
    lines.push('⌁ تقديري — ماكو نموذج تأخير حقيقي.')
    return { lines, ok: false }
  }

  lines.push(`PING ${dst}: 32 data bytes`)
  const hops = F.pathBetween(fromId, toId).filter((id) =>
    doc.nodes.find((n) => n.id === id)?.partId === 'router').length
  const ttl = 128 - hops
  const ms = LAN_MS + hops * HOP_MS

  for (let i = 0; i < count; i++) {
    lines.push(v.ok ? `Reply from ${dst}: bytes=32 time=${ms}ms TTL=${ttl}` : 'Request timed out.')
  }
  const recv = v.ok ? count : 0
  lines.push('', `--- ${dst} ping statistics ---`)
  lines.push(`${count} packets transmitted, ${recv} received, ${v.ok ? 0 : 100}% packet loss`)
  if (v.ok) lines.push(`round-trip min/avg/max = ${ms}/${ms}/${ms} ms`)
  lines.push('', '⌁ الأزمنة تقديرية ثابتة — درجة الدقة F1، ماكو نموذج تأخير.')
  return { lines, ok: v.ok }
}

/**
 * ═══ traceroute ═══
 *
 * ⚠️ هنا القيمة الحقيقية الي ما ينطيها `ping`: **وين** انقطع. نمشي
 * على المسار حتى لو فيه وصلة ميتة (`includeDead`) حتى نبيّن آخر قفزة
 * وصلت — ping يقول «ما وصل» وtraceroute يقول «وصل لهنا وطاح».
 *
 * ⚠️ والقفزات **الراوترات بس**: السويچ مو قفزة. محاكي يعد السويچات
 * قفزات يعلّم شي يناقض أول درس بالتوجيه.
 */
export function traceroute(doc: LabDoc, fromId: string, toId: string): ToolOutput {
  const F = netFacts(doc)
  const dst = ipOf(doc, toId)
  const lines: string[] = [`${labelOf(doc, fromId)}> traceroute ${dst}`, '']
  lines.push(`traceroute to ${dst}, 30 hops max`)

  const live = F.pathBetween(fromId, toId)
  const path = live.length ? live : F.pathBetween(fromId, toId, { includeDead: true })
  if (path.length === 0) {
    for (let i = 1; i <= 3; i++) lines.push(` ${i}  * * *  Request timed out.`)
    lines.push('', '✗ ماكو مسار فيزيائي أصلاً — ولا قفزة وصلت.')
    lines.push('⌁ الأزمنة تقديرية ثابتة — درجة الدقة F1.')
    return { lines, ok: false }
  }

  // نمشي على المسار ونوگف عند أول وصلة ميتة.
  const hops: { ip: string; label: string; ms: number; isDest: boolean }[] = []
  let broke: string | null = null
  let ms = LAN_MS
  for (let i = 1; i < path.length; i++) {
    const prev = path[i - 1], cur = path[i]
    const lk = doc.links.find((l) =>
      (l.from.node === prev && l.to.node === cur) || (l.to.node === prev && l.from.node === cur))
    if (lk && F.deadLinks.has(lk.id)) { broke = prev; break }
    const nd = doc.nodes.find((n) => n.id === cur)
    ms += HOP_MS
    if (nd?.partId === 'router') {
      hops.push({ ip: routerIpToward(doc, cur, prev), label: labelOf(doc, cur), ms, isDest: false })
    } else if (cur === toId) {
      hops.push({ ip: ipOf(doc, cur), label: labelOf(doc, cur), ms, isDest: true })
    }
  }

  const v = F.reach(fromId, toId)
  // ⚠️ الوجهة تنعرض **بس إذا وصلت فعلاً**. بحالة الـVLAN المختلف
  // المسار الفيزيائي كامل والباكيت ما يوصل — عرضها كقفزة ناجحة يعني
  // أداة تكذب، والمتدرّب يستبعد الجهة الصح من التشخيص.
  const shown = v.ok || broke ? hops : hops.filter((h) => !h.isDest)
  shown.forEach((h, i) => lines.push(` ${i + 1}  ${h.ip}  ${h.ms} ms  (${h.label})`))

  if (broke) {
    lines.push(` ${shown.length + 1}  * * *  Request timed out.`)
    lines.push('', `✗ آخر قفزة وصلت: «${labelOf(doc, broke)}». الانقطاع بعدها مباشرة.`)
  } else if (!v.ok) {
    lines.push(` ${shown.length + 1}  * * *  Request timed out.`)
    lines.push('', '✗ المسار الفيزيائي كامل والباكيت ما رجع — الخلل منطقي مو بالكيبل.')
  } else {
    lines.push('', `✅ وصل بـ${shown.length} قفزة.`)
  }
  lines.push('⌁ الأزمنة تقديرية ثابتة — درجة الدقة F1.')
  return { lines, ok: v.ok && !broke }
}

/** عنوان منفذ الراوتر الي دخلنا منه — مو أول عنوان عنده.
 *  ⚠️ traceroute الحقيقي يعرض **المنفذ الداخل**، والفرق مهم لمن
 *  يكون الراوتر بشبكتين ويحتار الفني بأي جهة الخلل. */
function routerIpToward(doc: LabDoc, routerId: string, fromId: string): string {
  const R = doc.nodes.find((n) => n.id === routerId)
  if (!R) return '—'
  const link = doc.links.find((l) =>
    (l.from.node === routerId && l.to.node === fromId) || (l.to.node === routerId && l.from.node === fromId))
  const port = link ? (link.from.node === routerId ? link.from.port : link.to.port) : null
  if (port) {
    const ip = str(R.params[`ip_${port}`])
    if (ip) return ip
  }
  const any = Object.entries(R.params).find(([k, v]) => k.startsWith('ip_') && str(v))
  return any ? str(any[1]) : '—'
}

/**
 * ═══ فحص الوصلة ═══
 *
 * ⚠️ السعة الفعلية **مو** المكتوبة على الكيبل: Cat6 على ١٤٠ متر
 * يطلع من المواصفة، وSFP نحاسي ١ غيغا على كيبل ١٠ غيغا يحدّد السرعة
 * بواحد. الأداة تبيّن **الفرق** بين المتوقّع والفعلي — وهذا الي
 * يشوفه الفني بجهاز الفحص.
 */
export function linkTest(doc: LabDoc, linkId: string): ToolOutput {
  const F = netFacts(doc)
  const l = doc.links.find((x) => x.id === linkId)
  const c = F.linkChecks.get(linkId)
  if (!l || !c) return { lines: ['ماكو وصلة بهذا المعرّف.'], ok: false }

  const cable = CABLE_BY_ID[c.cable]
  const lengthM = Number(l.params?.lengthM ?? 15)
  const lines = [
    `فحص الوصلة: ${c.aName} ⇄ ${c.bName}`,
    '',
    `الكيبل           ${cable?.name ?? c.cable}`,
    `الطول            ${lengthM} متر (الحد ${cable?.maxM ?? '—'} متر)`,
    `السعة الاسمية    ${cable?.maxMbps ?? '—'} ميغابت`,
    `السعة الفعلية    ${c.ok ? `${c.mbps} ميغابت` : '— (الرابط ميت)'}`,
    `الحالة           ${c.forced ? '✗ مقطوع' : c.ok ? '✅ سليم' : '✗ فاشل'}`,
  ]
  if (c.ok && cable && c.mbps < cable.maxMbps) {
    lines.push('', `⚠️ يشتغل بـ${c.mbps} من ${cable.maxMbps} — الترانسيفر يحدّد السرعة مو الكيبل.`)
  }
  for (const p of c.problems) lines.push('', `✗ ${p}`)
  return { lines, ok: c.ok }
}

/** الأجهزة الي تنفع طرفاً لـping — الي عندها عنوان. */
export function pingableNodes(doc: LabDoc): { id: string; label: string; ip: string }[] {
  return netFacts(doc).endpoints.map((n) => ({
    id: n.id,
    label: labelOf(doc, n.id),
    ip: ipOf(doc, n.id),
  }))
}

export type { ReachCode }
