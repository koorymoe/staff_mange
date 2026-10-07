import { useEffect, useState } from 'react'
import { api, type Booking } from '../api'
import BookingCodeChip from './BookingCodeChip'

// ═══ ترحيل الحجز للتقني (قرار (ع) 10-07) ═══
// الإداري تواصل ويا الزبون والمشكلة ما يعرفون الفنيين يحلّوها: يرحّله لتقني
// أو مسؤول خدمة. من هنا الحجز برقبة التقني — يتواصل، يكشف، ويعالج.

export default function HandoverDialog({ booking, onClose, onDone }: { booking: Booking; onClose: () => void; onDone: (b: Booking) => void }) {
  const [to, setTo] = useState('')
  const [why, setWhy] = useState('')
  const [busy, setBusy] = useState(false)
  const [people, setPeople] = useState<{ id: string; name: string; position: string }[]>([])
  useEffect(() => { void api.getHandoverCandidates().then(setPeople).catch(() => {}) }, [])
  const go = async () => {
    if (!to || !why.trim()) return
    setBusy(true)
    try {
      const b = await api.handoverBooking(booking.id, to, why.trim())
      alert(`الحجز ${booking.code} انرحّل — صار برقبة التقني`)
      onDone(b)
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر الترحيل')
    } finally {
      setBusy(false)
    }
  }
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4" onClick={() => !busy && onClose()}>
      <div dir="rtl" className="w-full max-w-md rounded-2xl bg-white p-5 text-right shadow-xl" onClick={(e) => e.stopPropagation()}>
        <h3 className="text-lg font-bold text-[#0f2040]">🛠️ رحّل للتقني — حجز <BookingCodeChip code={booking.code} /></h3>
        <p className="mt-1 text-xs text-slate-500">
          الحجز يطلع من مسؤوليتك ويصير برقبة التقني: يتواصل ويا الزبون، يكتب الكشف، ويعالج.
          إنت تنحسب على سرعة تواصلك وسرعة الترحيل. {booking.confirmationContactedAt ? '' : '⚠️ لازم تضغط «تواصلت ويا الزبون» أول.'}
        </p>
        <label className="mt-4 block text-sm font-bold text-slate-700">شنو مشكلة الزبون؟</label>
        <textarea value={why} onChange={(e) => setWhy(e.target.value)} rows={3} className="mt-1 w-full rounded-lg border border-slate-300 p-2 text-sm" placeholder="مثلاً: الكاميرات تنطفي بالليل والفنيين ما لگوا السبب" />
        <label className="mt-3 block text-sm font-bold text-slate-700">التقني</label>
        <select value={to} onChange={(e) => setTo(e.target.value)} className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm">
          <option value="">— اختار تقني أو مسؤول خدمة —</option>
          {people.map((p) => <option key={p.id} value={p.id}>{p.name} ({p.position})</option>)}
        </select>
        <div className="mt-4 flex gap-2">
          <button onClick={() => void go()} disabled={!to || !why.trim() || busy} className="flex-1 rounded-lg bg-orange-600 px-4 py-2.5 text-sm font-bold text-white disabled:opacity-50">
            {busy ? 'جاري الترحيل…' : '🛠️ رحّل'}
          </button>
          <button onClick={onClose} disabled={busy} className="rounded-lg border border-slate-300 px-4 py-2.5 text-sm font-bold text-slate-600">إلغاء</button>
        </div>
      </div>
    </div>
  )
}

/** شارة «عند فلان» للحجز المرحّل. */
export function HandoverBadge({ booking }: { booking: Booking }) {
  if (!booking.handoverTo) return null
  return (
    <span className="rounded-full bg-orange-100 px-3 py-1 text-xs font-bold text-orange-800" title={booking.handoverReason ?? ''}>
      🛠️ عند {booking.handoverTo.name}{booking.techDiagnosedAt ? ' · انكشف' : booking.techContactedAt ? ' · تواصل' : ''}
    </span>
  )
}
