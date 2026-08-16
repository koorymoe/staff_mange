// ═══ تفرّعات «تم الإنجاز» ═══
//
// «بعد ما يوصل الحجز مرحلة التكليف راح ياخذ اتجاهين: الاتجاه الأول
// الي هو تم الإنجاز… وراها بنود نتفرّع: منجز بدون فاتورة وتقرير،
// منجز بدون تقرير، منجز بدون فاتورة، أو منجز بشكل جزئي».
//
// الحالة تنحسب بالسيرفر من وجود الفاتورة والتقرير فعلاً — ما نعتمد
// على تأشيرة يدوية تنسى أو تنكذب.
//
// ⚠️ بملف منفصل عن الشاشة حتى يضل التحديث السريع (Fast Refresh)
// شغّال بالتطوير — نفس سبب `completionStates.ts`.
export type BookingBucket = 'all' | 'pending' | 'confirmed' | 'assigned' | 'done'

export type DoneFilter = 'ALL' | 'DONE_FULL' | 'DONE_NO_INVOICE' | 'DONE_NO_REPORT' | 'DONE_NO_BOTH' | 'PARTIAL'

export const DONE_FILTERS: { key: DoneFilter; label: string }[] = [
  { key: 'ALL', label: 'الكل' },
  { key: 'DONE_FULL', label: '✔ كامل' },
  { key: 'DONE_NO_INVOICE', label: '⚠ بلا فاتورة' },
  { key: 'DONE_NO_REPORT', label: '⚠ بلا تقرير' },
  { key: 'DONE_NO_BOTH', label: '⚠ بلا فاتورة وتقرير' },
  { key: 'PARTIAL', label: '⏳ منجز جزئي' },
]
