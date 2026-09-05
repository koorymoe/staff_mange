import type { Booking } from './api'

// ═══ حالات الحجز — تسمية وحدة للنظام كله ═══
//
// «الحجز قبل التثبيت يعتبر حجز بانتظار التثبيت».
//
// ⚠️ نفس الحالة كانت تنكتب بأربع تسميات مختلفة حسب الشاشة:
//   الزبائن      → «بانتظار التثبيت»
//   الخريطة      → «بانتظار التأكيد»
//   التدقيق اليومي → «معلّق»
//   المراقبة      → «بانتظار»
//
// والموظف الي يشوف «معلّق» بشاشة و«بانتظار التأكيد» بشاشة ثانية ما
// يدري إنهن **نفس الحجز بنفس الحالة** — فيسأل، أو يسوي الإجراء مرتين.
//
// هسه المصدر واحد: أي شاشة تستورد من هنا.
export const BOOKING_STATUS_LABELS: Record<Booking['status'], string> = {
  PENDING: 'بانتظار التثبيت',
  CONFIRMED: 'مثبّت',
  IN_PROGRESS: 'جاري التنفيذ',
  PARTIAL: 'إنجاز جزئي',
  WAITING: 'بالانتظار (ما رد الزبون)',
  COMPLETED: 'منجز',
  CANCELLED: 'ملغى',
}

export const BOOKING_STATUS_COLORS: Record<Booking['status'], string> = {
  PENDING: 'bg-amber-100 text-amber-800',
  CONFIRMED: 'bg-blue-100 text-blue-700',
  IN_PROGRESS: 'bg-orange-100 text-orange-700',
  PARTIAL: 'bg-violet-100 text-violet-700',
  WAITING: 'bg-slate-200 text-slate-600',
  COMPLETED: 'bg-emerald-100 text-emerald-700',
  CANCELLED: 'bg-red-100 text-red-700',
}

/** تسمية آمنة: حالة غريبة تنعرض كما هي بدل ما تطلع فارغة. */
export const bookingStatusLabel = (s: string): string =>
  BOOKING_STATUS_LABELS[s as Booking['status']] ?? s

export const bookingStatusColor = (s: string): string =>
  BOOKING_STATUS_COLORS[s as Booking['status']] ?? 'bg-slate-100 text-slate-600'
