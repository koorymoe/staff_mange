import type { Booking } from '../api'

// ═══ ورق الحجز على مسؤول الخدمة؟ (دازاي 11-21) ═══
// «فواتير الجي بي اس والداش كام… تترحل كبل لمسؤول الخدمة هوه يسوي فاتورة وتقرير».
// الخدمة ممكن تجي رئيسية (`service`) أو ضمن خدمات الحجز (`services`) — فالفحص على الكل:
// إذا كل خدمات الحجز ورقها على مسؤولها ← ما تطلع للفني/الليدر.
export function paperworkOnManager(b: Booking): boolean {
  const all = [b.service, ...(b.services ?? [])].filter((s): s is NonNullable<typeof s> => !!s)
  return all.length > 0 && all.every((s) => !!s.managerHandlesPaperwork || s.serviceKind === 'GPS' || s.serviceKind === 'DASHCAM')
}
