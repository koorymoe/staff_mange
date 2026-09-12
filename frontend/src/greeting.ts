// ═══ التحية حسب وقت اليوم ═══
//
// «مرحباً» وحدة لكل الناس بكل الأوقات — تنقرا مرة وبعدها تصير خلفية
// ما ينتبهلها أحد.
//
// الي يشتغل الصبح لازم يوصله «صباح الخير»، والي يفتح النظام العصر
// «مساء الخير». التحية الصحيحة بوقتها تخلي الشاشة تحس إنها **تعرف
// منو فتحها ومتى** — مو لوحة ثابتة.
//
// السلوك: أول ما يفتح تطلع التحية الزمنية، وبعد ثواني تتحوّل لـ
// «مرحباً» بحركة ناعمة. سبب التحوّل: التحية الزمنية جميلة أول لحظة،
// بس تبقى على الشاشة كل اليوم تصير غلط — الموظف يفتح النظام الساعة
// ١١ ليلاً ويشوف «مساء الخير» وهو داخل يخلّص شغلة متأخرة.

/** حدود الفترات بتوقيت بغداد المحلي للجهاز. */
export type DayPart = 'dawn' | 'morning' | 'noon' | 'evening' | 'night'

export function dayPartAt(d: Date = new Date()): DayPart {
  const h = d.getHours()
  if (h >= 3 && h < 6) return 'dawn'
  if (h >= 6 && h < 12) return 'morning'
  if (h >= 12 && h < 16) return 'noon'
  if (h >= 16 && h < 21) return 'evening'
  return 'night'
}

/** التحية الزمنية — الي تطلع أول ما يفتح النظام. */
export function timeGreeting(d: Date = new Date()): { text: string; icon: string } {
  switch (dayPartAt(d)) {
    // ⚠️ الفجر منفصل عن الصباح: مو مجاملة — أكو كوادر تطلع للشد قبل
    // الشمس، و«صباح الخير» الساعة ٤ فجراً تحس مثل النظام ما يعرف
    // إنهم صاحين من الليل.
    case 'dawn':    return { text: 'صبّحك الله بالخير', icon: '🌅' }
    case 'morning': return { text: 'صباح الخير',        icon: '☀️' }
    case 'noon':    return { text: 'نهارك سعيد',        icon: '🌤️' }
    case 'evening': return { text: 'مساء الخير',        icon: '🌇' }
    case 'night':   return { text: 'مساء الخير',        icon: '🌙' }
  }
}

/** التحية الثابتة الي تستقر عليها الشاشة بعد الترحيب. */
export function restingGreeting(): string {
  return 'مرحباً'
}

/** كم يقعد الترحيب قبل ما يتحوّل — بالمللي ثانية. */
export const GREETING_HOLD_MS = 4200
