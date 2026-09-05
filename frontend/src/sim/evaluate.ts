import type { Expect, SimAction, Step, WrongCase } from './types'

// ═══ تقييم حركة المتدرّب ═══
//
// دالة **نقية** بلا حالة ولا نداءات — تاخذ الخطوة والحركة وترجّع الحكم.
// هذا مقصود: هي قلب المحاكي، ولازم تنفحص لحالها بلا متصفح ولا سيرفر.
//
// ⚠️ والتقييم بالواجهة مو بالسيرفر: قواعد المحاكاة معقّدة وتختلف بكل
// محرّك، وإعادة كتابتها بـGo تعني نسختين تنحرفن عن بعض بأول تعديل.
// بس الدرجة النهائية **يحسبها السيرفر** من أوزان الخطوات وسجل الأحداث —
// فالتقييم هنا يقرر «صح لو غلط»، والسيرفر يقرر «شكد ياخذ».

export interface Verdict {
  result: 'PASS' | 'WRONG'
  /** تفسير الغلط — شنو صار وشنو عاقبته بالميدان. */
  say?: string
  penalty?: number
  /** ⚠️ الغلط القاتل **ما ينهي المحاولة**: يعرض العاقبة ويخلّيه يعيد.
   *  إنهاء المحاولة يعاقب، والعرض يعلّم — واحنا نريد يتعلّم. */
  fatal?: boolean
}

/** يطابق طرفين — التوصيل ما إله اتجاه، فربط أ بـب مثل ربط ب بأ. */
function sameLink(a: Expect | SimAction, b: Expect | SimAction): boolean {
  if (!a.from || !a.to || !b.from || !b.to) return false
  return (a.from === b.from && a.to === b.to) || (a.from === b.to && a.to === b.from)
}

/** يطابق حركة مع حالة غلط معرّفة بالمحتوى. */
function matchesWrong(w: WrongCase, action: SimAction): boolean {
  if (w.matchAny) return true
  if (!w.match) return false
  if (w.match.op && w.match.op !== action.op) return false
  return sameLink(w.match, action)
}

export function evaluateAction(step: Step | undefined, action: SimAction): Verdict {
  if (!step) return { result: 'WRONG', say: 'ماكو خطوة حالية.', penalty: 0 }

  if (step.expect.op === action.op && sameLink(step.expect, action)) {
    return { result: 'PASS' }
  }

  // نمشي على حالات الغلط بالترتيب: أول تطابق ياخذ رسالته.
  // ⚠️ الترتيب مهم — `matchAny` تمسك كل شي، فلازم تكون آخر وحدة
  // بالمحتوى. لو انحطت أول، ما توصل أي رسالة محدّدة أبداً.
  for (const w of step.wrong ?? []) {
    if (matchesWrong(w, action)) {
      return {
        result: 'WRONG',
        say: w.say,
        penalty: w.penalty ?? step.wrongPenalty ?? 5,
        fatal: w.fatal,
      }
    }
  }

  return {
    result: 'WRONG',
    say: 'هذا مو التوصيل المطلوب بهاي الخطوة.',
    penalty: step.wrongPenalty ?? 5,
  }
}

/** هل السلك موجود أصلاً؟ (بلا اتجاه) */
export function wireExists(wires: { from: string; to: string }[], a: string, b: string): boolean {
  return wires.some((w) => (w.from === a && w.to === b) || (w.from === b && w.to === a))
}
