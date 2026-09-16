// يحوّل رقم عراقي محلي (07xxxxxxxx) لصيغة دولية بلا أصفار/رموز —
// wa.me وt.me يحتاجونها هيج بالضبط.
export function toIntlPhone(phone: string): string {
  const digits = phone.replace(/\D/g, '')
  return digits.startsWith('0') ? '964' + digits.slice(1) : digits
}

/**
 * نفس التحويل بس **يتحقّق**: يرجّع `null` للرقم المو صالح.
 *
 * ⚠️ ليش منفصلة عن `toIntlPhone`: هاي ترجّع الأرقام مثل ما هي حتى لو
 * كانت ناقصة، وعندها مستخدمون يعتمدون عليها. والزر الي يفتح
 * `wa.me/<رقم ناقص>` **يفتح ويفشل صامتاً** — الإداري يحسب إنه أرسل
 * وهو لا. فاستخدم هاي لمّا تقرّر **هل تعرض الزر أصلاً**.
 */
export function intlPhoneOrNull(raw: string | null | undefined): string | null {
  if (!raw) return null
  const d = raw.replace(/\D/g, '')
  if (/^964\d{9,10}$/.test(d)) return d
  if (/^0\d{9,10}$/.test(d)) return `964${d.slice(1)}`
  if (/^7\d{8,9}$/.test(d)) return `964${d}`
  return null
}
