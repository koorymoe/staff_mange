import type { Booking } from '../api'
import { COMPLETION_LABELS } from './completionStates'

// ═══ حالة الحجز بنظرة وحدة ═══
// الإداري لازم يعرف مو بس «انتهى لو لا» — لازم يعرف منو خلّص شغله كامل
// ومنو أنجز وترك الورق وراه (فاتورة التكاليف وتقرير العمل).
export default function CompletionBadge({ booking }: { booking: Pick<Booking, 'completionState' | 'settledLegacyAt'> }) {
  // ═══ التسوية الإدارية تنبيّن ═══
  // حجز انقفل «منجز بدون تفاصيل» مو نفس حجز خلص شغله وانوثّق. لو
  // انعرضوا بنفس الشارة، التقارير تگول «منجز بشكل كامل» على شغل ماكو
  // عنه ولا معلومة — والفرق ينضيع بعد شهر ومحد يتذكره.
  if (booking.settledLegacyAt) {
    return (
      <span className="inline-block shrink-0 rounded-full border border-slate-300 bg-slate-100 px-2.5 py-0.5 text-[11px] font-bold text-slate-600">
        📁 منجز (تسوية إدارية)
      </span>
    )
  }
  const l = COMPLETION_LABELS[booking.completionState]
  if (!l) return null
  return (
    <span className={`inline-block shrink-0 rounded-full border px-2.5 py-0.5 text-[11px] font-bold ${l.cls}`}>
      {l.text}
    </span>
  )
}
