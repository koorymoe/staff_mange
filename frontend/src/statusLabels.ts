// ═══ (دازاي 11-15) تسميات الحالات العامة — مصدر واحد ═══
// نفس الحالة چانت تنكتب «ملغى/ملغي»، «معلّق/قيد الانتظار/بالانتظار»، «تمت الموافقة/معتمد/موافق عليه»…
// الشاشات تاخذ من هنا وتضيف بس حالاتها الخاصة (FULFILLED، RETURNED…).
// ⚠️ حالات الحجز لها مصدرها: bookingStatus.ts.
export const STATUS_LABELS = {
  PENDING: 'بالانتظار',
  CONFIRMED: 'مثبّت',
  IN_PROGRESS: 'قيد التنفيذ',
  DONE: 'منجز',
  COMPLETED: 'منجز',
  APPROVED: 'موافق عليه',
  REJECTED: 'مرفوض',
  CANCELLED: 'ملغي',
} as const

export type CommonStatus = keyof typeof STATUS_LABELS

/** تسمية آمنة: حالة غريبة ترجع كما هي. */
export const statusLabel = (s: string): string => (STATUS_LABELS as Record<string, string>)[s] ?? s
