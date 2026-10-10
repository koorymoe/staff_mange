import type { MatrixProposal } from '../api'

// أدلة اقتراحات ماتركس بكلام مفهوم — مشتركة بين صندوق القرارات ومركز القيادة.

export const EV_LABEL: Record<string, string> = {
  done: 'المنجز', left: 'الباقي', hoursElapsed: 'ساعات مضت', hoursRemaining: 'ساعات باقية', pace: 'الوتيرة/ساعة',
  expected: 'المتوقع ينجز', total: 'كل التذكيرات', escalated: 'صعدت', resolved: 'انحلت', rejected: 'رفضتها',
  kind: 'النوع', escalations30d: 'تصعيدات بشهر', group: 'المجموعة', hits: 'مرات الانطباق', lastHitAt: 'آخر انطباق',
  ratio: 'معدّل الوقت هالشهر (الفعلي ÷ المتوقع)', prevRatio: 'معدّل الشهر الي قبله', timed: 'حجوزات موقوتة هالشهر', prevTimed: 'حجوزات موقوتة قبله',
}

// ── الأهمية والأدلة — محسوبة من أرقام الدليل بس، بلا تخمين ──
export type Sev = 'URGENT' | 'MEDIUM' | 'NORMAL'
export const SEV: Record<Sev, { t: string; c: string; icon: string }> = {
  URGENT: { t: 'عاجل', c: 'bg-red-50 text-red-700 ring-red-200', icon: '⚠️' },
  MEDIUM: { t: 'متوسط', c: 'bg-amber-50 text-amber-700 ring-amber-200', icon: '⏱️' },
  NORMAL: { t: 'عادي', c: 'bg-sky-50 text-sky-700 ring-sky-200', icon: 'ℹ️' },
}
export const num = (v: unknown) => (typeof v === 'number' ? v : Number(v) || 0)
export const isDecline = (p: MatrixProposal) => p.evidence != null && 'prevRatio' in p.evidence
export function severity(p: MatrixProposal): Sev {
  const e = p.evidence ?? {}
  if (isDecline(p)) {
    const r = num(e.ratio), pr = num(e.prevRatio)
    return pr > 0 && (r - pr) / pr >= 0.5 ? 'URGENT' : 'MEDIUM'
  }
  if (p.kind === 'PREDICTION') {
    const left = num(e.left), exp = Math.max(1, num(e.expected))
    return left >= exp * 3 ? 'URGENT' : left >= exp * 2 ? 'MEDIUM' : 'NORMAL'
  }
  return 'NORMAL'
}
export function category(p: MatrixProposal) {
  if (isDecline(p)) return 'الأداء والحجوزات'
  if (p.kind === 'PREDICTION') return 'اعتماد الدوام'
  return 'تعليمات ماتركس'
}
export function evidenceCards(p: MatrixProposal): { label: string; value: string; icon: string }[] {
  const e = p.evidence ?? {}
  if (isDecline(p)) {
    return [
      { icon: '📈', label: 'معدّل الوقت هالشهر', value: ratioText(num(e.ratio), num(e.timed)) },
      { icon: '🕒', label: 'الشهر الي قبله', value: ratioText(num(e.prevRatio), num(e.prevTimed)) },
    ]
  }
  if (p.kind === 'PREDICTION') {
    const left = num(e.left), pace = num(e.pace)
    const hrs = pace > 0 ? Math.ceil(left / pace) : null
    return [
      { icon: '📋', label: 'المهام المعلّقة', value: `${left} مهمة` },
      { icon: '⏰', label: 'موعد الإنجاز المتوقع', value: hrs == null ? 'غير معروف (ما بدأ)' : hrs > 24 ? `بعد ${Math.ceil(hrs / 24)} أيام` : `خلال ${hrs} ساعات` },
    ]
  }
  return Object.entries(e).slice(0, 2).map(([k, v]) => ({ icon: '📊', label: EV_LABEL[k] ?? k, value: typeof v === 'object' ? JSON.stringify(v) : String(v) }))
}

// نسبة فوگ ×10 = تسجيل أوقات غلط، مو بطء حقيقي — ما نعرضها كرقم.
export function ratioText(r: number, n: number) {
  if (r > 10) return `بيانات غير منطقية — تحتاج مراجعة (${n} حجز)`
  return `×${r.toFixed(2)} (${n} حجز)`
}
