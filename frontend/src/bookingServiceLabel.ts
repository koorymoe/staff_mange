// ═══ اسم خدمة الحجز — مكان واحد (دازاي 11-15، بلاغ موظف #42) ═══
// جانت أربع نسخ محلية (المتابعة، الحجوزات، المنسّق، المحطات) وكلها تكتب «بدون خدمة محددة»
// لحجز الصيانة حتى لو الزبون معروفة خدمته. الترتيب هسه:
//   ١. الخدمة الرئيسية (+ باقي خدمات الحجز إذا أكثر من منظومة) بـ«، »
//   ٢. حجز صيانة → «صيانة» + خدمة الزبون إذا معروفة
//   ٣. غير هيچ → «⚠️ بلا خدمة» حتى ينشاف ويتصلّح

type Named = { name?: string | null } | null | undefined

export type BookingServiceSource = {
  service?: Named
  services?: Named[] | null
  bookingType?: string | null
  customer?: { services?: string[] | null } | null
}

export const NO_SERVICE_LABEL = '⚠️ بلا خدمة'

export function bookingServiceLabel(b: BookingServiceSource | null | undefined): string {
  if (!b) return NO_SERVICE_LABEL
  const names: string[] = []
  for (const s of [b.service, ...(b.services ?? [])]) {
    const n = s?.name?.trim()
    if (n && !names.includes(n)) names.push(n)
  }
  if (names.length) return names.join('، ')
  if (b.bookingType === 'MAINTENANCE') {
    const cs = (b.customer?.services ?? []).map((x) => x.trim()).filter(Boolean)
    return cs.length ? `صيانة — ${cs.join('، ')}` : 'صيانة'
  }
  return NO_SERVICE_LABEL
}
