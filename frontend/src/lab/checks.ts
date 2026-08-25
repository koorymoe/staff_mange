// ═══ فحوص التحديات ═══
//
// مساحة العمل بلا تقييم تبقى **لعبة بناء**. والتقييم لازم يكون على
// **نتيجة المحرّك** مو على شكل المخطط: المخطط الرئيسي (٤١) يشترط
// «كل تقييم قابل للتتبّع إلى actions ومقاييس لا إلى checkbox».
//
// ⚠️ يعني: التحدي ما يكول «حط سويچاً بالإحداثي الفلاني». يكول «خلّي
// الكاميرات الثلاث تشتغل بلا تجاوز ميزانية PoE» — والفني يوصلها
// بأي ترتيب يريده. أي طريق صحيح ينجح.
//
// ⚠️ والفحوص **نقية**: تاخذ المستند والنتيجة وترجّع نجاح/فشل مع سبب.
// بلا DOM وبلا حالة — فتنفحص بسطر أمر بلا متصفح.

import type { LabDoc, LabNode, SimResult } from './types'

export type LabCheck =
  /** المحاكاة تمشي بلا أي خطأ. */
  | { c: 'noErrors' }
  /** كل زوج أجهزة طرفية يوصل. */
  | { c: 'allPairsConnected' }
  /** عدد نسخ قطعة بعينها. */
  | { c: 'hasPart'; partId: string; min?: number; max?: number }
  /** جهازان **معزولان** منطقياً (VLAN مختلف). */
  | { c: 'isolated'; a: string; b: string }
  /** جهازان يوصلون. */
  | { c: 'connected'; a: string; b: string }
  /** ماكو كيبل مكسور بالمخطط. */
  | { c: 'allLinksOk' }
  /** قراءة على قطعة تطابق نصاً (مثل «PoE 24/185 W»). */
  | { c: 'reading'; node: string; contains: string }

export interface CheckResult {
  ok: boolean
  /** ليش فشل — يطلع للمتدرّب مثل ما تطلع رسالة الغلط بالتوصيل. */
  why?: string
}

/** يلگي عقدة باسمها المعروض (`name` أو `hostname`) — مو بمعرّفها.
 *
 *  ⚠️ التحدي ينكتب بأسماء يفهمها الإنسان («PC1»)، والمعرّفات تتولّد
 *  عشوائياً بكل وضع. ربط التحدي بالمعرّف يعني تحدياً ينكسر بأول مرة
 *  يحذف بيها المتدرّب قطعة ويعيدها. */
function byName(doc: LabDoc, name: string): LabNode | undefined {
  const wanted = name.trim().toLowerCase()
  return doc.nodes.find((n) => {
    const label = String(n.params.name ?? n.params.hostname ?? n.label ?? '')
    return label.trim().toLowerCase() === wanted
  })
}

/** يقرا حالة زوج من رسائل المحرّك.
 *
 *  ⚠️ ما نعيد حساب الوصولية هنا: المحرّك حسبها، وأي حساب ثانٍ يعني
 *  منطقين ممكن يفترقون — والمتدرّب يشوف «الاتصال ناجح» بالنتائج
 *  و«فشلت» بالتقييم. */
function pairFailed(result: SimResult, a: string, b: string): { failed: boolean; text?: string } {
  for (const m of result.messages) {
    if (m.kind === 'info') continue
    const t = m.text
    if ((t.includes(a) && t.includes(b)) && (t.includes('⇄') || t.includes('نفس العنوان'))) {
      return { failed: true, text: t }
    }
  }
  return { failed: false }
}

export function runCheck(check: LabCheck, doc: LabDoc, result: SimResult): CheckResult {
  switch (check.c) {
    case 'noErrors': {
      const errs = result.messages.filter((m) => m.kind === 'error')
      return errs.length === 0
        ? { ok: true }
        : { ok: false, why: `بعدها ${errs.length} مشكلة خطيرة: ${errs[0].text}` }
    }

    case 'allLinksOk': {
      const bad = Object.values(result.linkState).filter((s) => s === 'bad').length
      return bad === 0 ? { ok: true } : { ok: false, why: `${bad} كيبل مكسور بالمخطط.` }
    }

    case 'allPairsConnected': {
      const fail = result.messages.find((m) => m.kind !== 'info' && m.text.includes('⇄'))
      return fail ? { ok: false, why: fail.text } : { ok: true }
    }

    case 'hasPart': {
      const n = doc.nodes.filter((x) => x.partId === check.partId).length
      if (check.min !== undefined && n < check.min) {
        return { ok: false, why: `تحتاج ${check.min} على الأقل من هالقطعة، وعندك ${n}.` }
      }
      if (check.max !== undefined && n > check.max) {
        return { ok: false, why: `المسموح ${check.max} كحد أعلى من هالقطعة، وعندك ${n}.` }
      }
      return { ok: true }
    }

    case 'connected': {
      const A = byName(doc, check.a), B = byName(doc, check.b)
      if (!A || !B) return { ok: false, why: `ما لگيت «${!A ? check.a : check.b}» بالمخطط.` }
      const f = pairFailed(result, check.a, check.b)
      return f.failed ? { ok: false, why: f.text } : { ok: true }
    }

    case 'isolated': {
      const A = byName(doc, check.a), B = byName(doc, check.b)
      if (!A || !B) return { ok: false, why: `ما لگيت «${!A ? check.a : check.b}» بالمخطط.` }
      const f = pairFailed(result, check.a, check.b)
      // ⚠️ العزل المطلوب **منطقي** (VLAN) مو قطع الكيبل: قطع الكيبل
      // «يعزل» هم، بس هذا مو الي ينطلب بالمشروع — الضيف لازم يوصل
      // للإنترنت وما يوصل للشبكة الداخلية.
      if (!f.failed) return { ok: false, why: `${check.a} و${check.b} لسه يوصلون لبعض.` }
      if (!(f.text ?? '').includes('VLAN')) {
        return { ok: false, why: `انعزلوا بس مو بالـVLAN — العزل المطلوب منطقي مو بقطع الكيبل. (${f.text})` }
      }
      return { ok: true }
    }

    case 'reading': {
      const N = byName(doc, check.node)
      if (!N) return { ok: false, why: `ما لگيت «${check.node}» بالمخطط.` }
      const rs = result.nodeReadings[N.id] ?? []
      return rs.some((r) => r.text.includes(check.contains))
        ? { ok: true }
        : { ok: false, why: `قراءة «${check.node}» ما تطابق المطلوب.` }
    }

    default:
      return { ok: false, why: 'فحص مو معروف.' }
  }
}

/** يمشي على كل فحوص الخطوة — **كلها** لازم تنجح. */
export function runChecks(checks: LabCheck[], doc: LabDoc, result: SimResult): CheckResult {
  for (const c of checks) {
    const r = runCheck(c, doc, result)
    if (!r.ok) return r
  }
  return { ok: true }
}
