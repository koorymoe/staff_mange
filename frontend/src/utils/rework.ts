// (ل) 11-24: مساعدات «خطأ فني من الكادر السابق» — شوف components/ReworkFlag.tsx
import { apiRequest } from '../api'
import { notify } from '../ui/notify'

export type ReworkState = { on: boolean; orig: string; note: string }
export const emptyRework: ReworkState = { on: false, orig: '', note: '' }

export function reworkProblem(r: ReworkState): string {
  if (!r.on) return ''
  if (!r.orig.trim()) return 'اكتب رقم الحجز الأصلي الي بيه الخطأ الفني'
  if (r.note.trim().length < 5) return 'اكتب شنو چان الخطأ الفني وشنو صلّحتوا'
  return ''
}

/** يسجّل التصحيح بعد حفظ الفاتورة — فشله ما يلغي الفاتورة، بس ينبّه. */
export async function fileRework(r: ReworkState, fixCode: string | undefined) {
  if (!r.on) return
  if (!fixCode) { notify('الفاتورة انحفظت، بس تسجيل الخطأ الفني يحتاج حجز — سجّله من «الخطأ الفني والتصحيح»'); return }
  try {
    await apiRequest('/rework', { method: 'POST', body: JSON.stringify({ fixCode, origCode: r.orig.trim(), note: r.note.trim() }) })
    notify('🔧 انسجّل خطأ فني من الكادر السابق — ينتظر موافقة المراقب', 'success')
  } catch (e) {
    notify(`الفاتورة انحفظت، بس تسجيل الخطأ الفني ما صار: ${e instanceof Error ? e.message : ''} — سجّله من «الخطأ الفني والتصحيح»`)
  }
}

