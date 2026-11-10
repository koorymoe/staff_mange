import { useState } from 'react'
import { api, type BookingHistoryLine } from '../api'

// ═══ «🕓 منو سوّى شنو» — (ع) 11-10 بلاغ #30/#31 ═══
// «منو وصّلهن لخانة تم التثبيت؟ ما يطلع منو عدّلهن». كل حفظ ناجح على الحجز
// ينسجل أصلاً بسجل النشاط — هنا يطلع للمنسّق بالاسم والوقت.
// ينجلب بس لمن ينفتح، حتى ما نحمّل كل الحجوزات المفتوحة.
export default function BookingHistory({ bookingId }: { bookingId: string }) {
  const [lines, setLines] = useState<BookingHistoryLine[] | null>(null)
  const [error, setError] = useState('')
  const [open, setOpen] = useState(false)
  const toggle = () => {
    setOpen((o) => !o)
    if (lines === null) {
      api.getBookingHistory(bookingId).then(setLines).catch((e: Error) => setError(e.message || 'تعذر جلب السجل'))
    }
  }
  return (
    <div className="mt-3 rounded-xl border border-slate-200 bg-white p-3" dir="rtl">
      <button type="button" onClick={toggle} aria-expanded={open}
        className="text-[11px] font-extrabold text-[#0f2040]">🕓 منو سوّى شنو بهذا الحجز {open ? '▲' : '▼'}</button>
      {open && (error
        ? <p role="alert" className="mt-2 text-xs font-bold text-red-700">{error}</p>
        : !lines ? <p className="mt-2 text-xs text-slate-500">جاري التحميل…</p>
        : lines.length === 0 ? <p className="mt-2 text-xs text-slate-500">ماكو تعديل مسجّل (السجل يبدي من تشغيل سجل النشاط)</p>
        : (
          <ul className="mt-2 max-h-60 space-y-1 overflow-y-auto text-xs">
            {lines.map((l, i) => (
              <li key={i} className="flex flex-wrap gap-x-2 border-b border-slate-100 pb-1 last:border-0">
                <span className="tabular-nums text-slate-500">{new Date(l.at).toLocaleString('ar-IQ', { dateStyle: 'short', timeStyle: 'short' })}</span>
                <b className="text-slate-800">{l.by || 'غير معروف'}</b>
                <span className="text-slate-700">{l.text}</span>
              </li>
            ))}
          </ul>
        ))}
    </div>
  )
}
