// يحوّل رقم عراقي محلي (07xxxxxxxx) لصيغة دولية بلا أصفار/رموز —
// wa.me وt.me يحتاجونها هيج بالضبط.
export function toIntlPhone(phone: string): string {
  const digits = phone.replace(/\D/g, '')
  return digits.startsWith('0') ? '964' + digits.slice(1) : digits
}
