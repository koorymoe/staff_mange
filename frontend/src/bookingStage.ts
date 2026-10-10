import type { Booking } from './api'

// ═══ مراحل الحجز — متابعة مرحلة بعد مرحلة ═══
//
// «أريد متابعة لحالة الحجز مرحلة بعد مرحلة».
//
// الحالة (`status`) وحدها ما تگول القصة: حجز CONFIRMED ممكن يكون
// بلا كادر، وممكن كادره جاهز ومنتظر يومه، وممكن الفريق بالطريق —
// وكلهن نفس الكلمة «مثبّت».
//
// ⚠️ والأهم: **كل مرحلة اختيارية لحالها**. الحجز يتثبّت بلا كادر
// وبلا موعد، والكادر ينكلّف بعد أربع أيام. فالمراحل تنعرض «وين
// وصل» مو «شنو ناقص عليك حتى تكمّل» — ماكو خطوة تفرض الي بعدها.
export type StageKey = 'created' | 'contacted' | 'confirmed' | 'scheduled' | 'crewed' | 'started' | 'done'

export interface Stage {
  key: StageKey
  label: string
  icon: string
  /** خلصت؟ */
  done: (b: Booking) => boolean
  at: (b: Booking) => string | null | undefined
}

export const BOOKING_STAGES: Stage[] = [
  { key: 'created', label: 'انسجّل', icon: '📝', done: () => true, at: (b) => b.createdAt },
  { key: 'contacted', label: 'تواصلنا وية الزبون', icon: '📞', done: (b) => !!b.confirmationContactedAt, at: (b) => b.confirmationContactedAt },
  { key: 'confirmed', label: 'انثبّت', icon: '✅', done: (b) => !!b.confirmedAt, at: (b) => b.confirmedAt },
  { key: 'scheduled', label: 'انحدد الموعد', icon: '📅', done: (b) => !!b.scheduledAt, at: (b) => b.scheduledAt },
  { key: 'crewed', label: 'انكلّف الكادر', icon: '👥', done: (b) => (b.assignments?.length ?? 0) > 0, at: () => null },
  { key: 'started', label: 'بدا التنفيذ', icon: '🔧', done: (b) => !!b.startedAt || b.status === 'IN_PROGRESS' || b.status === 'COMPLETED', at: (b) => b.startedAt },
  { key: 'done', label: 'انجز', icon: '🏁', done: (b) => b.status === 'COMPLETED', at: (b) => b.completedAt },
]

// ═══ الحجز الي بدا التنفيذ ═══
//
// «من يتنسق ويستلمهن الليدر ويبدا التنفيذ بيهن المفروض يختفن من هاي
// الواجهة — وين يرحن؟ يرحن لحجوزات مكلفة».
//
// ⚠️ هذي الدالة هي **المصدر الوحيد** لهذا السؤال: شاشة التنسيق تخفي
// بيها، وسلّة «مكلّفة» تعرض بيها. لو كل شاشة حكمت لحالها، تجي لحظة
// يختفي بيها الحجز من التنسيق وما يطلع بالمكلّفة — وهاي هي «الهوسة»
// الي ما نريدها: حجز موجود بقاعدة البيانات وما يشوفه أحد.
export function executionStarted(b: Booking): boolean {
  return !!b.startedAt || !!b.arrivedAt || b.status === 'IN_PROGRESS' || b.status === 'COMPLETED'
}

/** أول مرحلة ما خلصت — «وين واقف الحجز هسه». */
export function currentStage(b: Booking): Stage {
  return BOOKING_STAGES.find((s) => !s.done(b)) ?? BOOKING_STAGES[BOOKING_STAGES.length - 1]
}

/** كم مرحلة خلصت من الكل. */
export function stageProgress(b: Booking): { done: number; total: number } {
  return { done: BOOKING_STAGES.filter((s) => s.done(b)).length, total: BOOKING_STAGES.length }
}
