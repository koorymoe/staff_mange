import type { CompletionState } from '../api'

// نصوص وألوان حالات الاكتمال — بملف منفصل عن المكوّن حتى يضل
// التحديث السريع (Fast Refresh) شغّال بالتطوير.
export const COMPLETION_LABELS: Record<CompletionState, { text: string; cls: string }> = {
  CONFIRMED_ONLY:  { text: '📌 مثبت (بلا كادر)',        cls: 'bg-violet-50 text-violet-700 border-violet-200' },
  ASSIGNED:        { text: '👷 في حالة التكليف',        cls: 'bg-sky-50 text-sky-700 border-sky-200' },
  STOPPED:         { text: '⏸ توقف العمل',              cls: 'bg-slate-100 text-slate-600 border-slate-300' },
  DONE_NO_BOTH:    { text: '⚠ منجز بدون فاتورة وتقرير', cls: 'bg-red-50 text-red-700 border-red-200' },
  DONE_NO_INVOICE: { text: '⚠ منجز بدون فاتورة',        cls: 'bg-amber-50 text-amber-800 border-amber-200' },
  DONE_NO_REPORT:  { text: '⚠ منجز بدون تقرير',         cls: 'bg-amber-50 text-amber-800 border-amber-200' },
  DONE_FULL:       { text: '✔ منجز بشكل كامل',          cls: 'bg-emerald-50 text-emerald-700 border-emerald-200' },
}

export const completionLabel = (s: CompletionState) => COMPLETION_LABELS[s]?.text ?? s

// ترتيب الحالات بالفلتر — الأهم (الي محتاج متابعة) أول
export const COMPLETION_ORDER: CompletionState[] = [
  'DONE_NO_BOTH', 'DONE_NO_INVOICE', 'DONE_NO_REPORT', 'DONE_FULL',
  'CONFIRMED_ONLY', 'ASSIGNED', 'STOPPED',
]
