import type { Booking } from '../api'
import { COMPLETION_LABELS } from './completionStates'

// ═══ حالة الحجز بنظرة وحدة ═══
// الإداري لازم يعرف مو بس «انتهى لو لا» — لازم يعرف منو خلّص شغله كامل
// ومنو أنجز وترك الورق وراه (فاتورة التكاليف وتقرير العمل).
export default function CompletionBadge({ booking }: { booking: Pick<Booking, 'completionState'> }) {
  const l = COMPLETION_LABELS[booking.completionState]
  if (!l) return null
  return (
    <span className={`inline-block shrink-0 rounded-full border px-2.5 py-0.5 text-[11px] font-bold ${l.cls}`}>
      {l.text}
    </span>
  )
}
