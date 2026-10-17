import { useEffect, useState } from 'react'
import { api, type Booking, type Service } from '../api'
import { useSession } from '../session'

// ═══ «🪚 شغل الديكور» — (ع) 10-16 ═══
// الديكور يطلعون ويا الشعبة الهندسية (حداد يسوي قاعدة الداتا شو…) أو قبلها بيوم.
// من حجز الشعبة ينسجل «حجز ديكور» مربوط بيه: كادره من الديكور، وتقريره وفاتورته إله.
const STATUS: Record<string, string> = {
  PENDING: 'معلّق', CONFIRMED: 'مثبّت', IN_PROGRESS: 'قيد التنفيذ', COMPLETED: 'خالص ✅', CANCELLED: 'ملغي', WAITING: 'منتظر',
}
const fmt = (s?: string | null) => (s ? new Date(s).toLocaleString('ar-IQ', { timeZone: 'Asia/Baghdad', weekday: 'short', day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' }) : 'بلا موعد')

export default function DecorPanel({ booking }: { booking: Booking }) {
  const { employee, permissions } = useSession()
  const canAdd = employee?.role === 'ADMIN' || employee?.role === 'HR_COORDINATOR' || employee?.role === 'MONITOR'
    || permissions.includes('coordinator') || permissions.includes('monitoring')
  const [rows, setRows] = useState<Booking[]>([])
  const [services, setServices] = useState<Service[]>([])
  const [open, setOpen] = useState(false)
  const [form, setForm] = useState({ serviceId: '', timing: 'BEFORE' as 'BEFORE' | 'WITH', at: '', notes: '' })
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  const isDecor = booking.bookingType === 'DECOR'

  useEffect(() => {
    if (isDecor) return
    let alive = true
    api.getBookingDecor(booking.id).then((x) => { if (alive) setRows(x) }).catch(() => {})
    return () => { alive = false }
  }, [booking.id, isDecor])
  useEffect(() => {
    if (!open || services.length) return
    api.getServices().then((s) => setServices(s.filter((x) => (x as { division?: string }).division === 'DECOR'))).catch(() => {})
  }, [open, services.length])

  if (isDecor) {
    return booking.parentBookingId ? (
      <p className="mt-3 rounded-lg bg-orange-50 px-3 py-2 text-xs font-bold text-orange-900">
        🪚 حجز ديكور — تابع لحجز الشعبة الهندسية {booking.parentBookingId.slice(0, 8)} · {booking.decorTiming === 'BEFORE' ? 'يطلعون قبلهم' : 'ويا الشعبة'}
      </p>
    ) : null
  }

  const add = async () => {
    if (!form.serviceId) { setErr('اختار المهنة'); return }
    setBusy(true); setErr('')
    try {
      const b = await api.createBookingDecor(booking.id, {
        serviceId: form.serviceId, timing: form.timing, notes: form.notes.trim() || undefined,
        scheduledAt: form.at ? new Date(form.at).toISOString() : undefined,
      })
      setRows((r) => [...r, b]); setOpen(false); setForm({ serviceId: '', timing: 'BEFORE', at: '', notes: '' })
    } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') } finally { setBusy(false) }
  }

  if (rows.length === 0 && !canAdd) return null
  return (
    <div className="mt-3 rounded-xl border border-orange-200 bg-orange-50/60 p-3 text-sm">
      <div className="flex items-center justify-between gap-2">
        <p className="font-extrabold text-orange-900">🪚 شغل الديكور ({rows.length})</p>
        {canAdd && !open && <button type="button" onClick={() => setOpen(true)} className="rounded-lg bg-orange-600 px-3 py-1 text-xs font-bold text-white">＋ أضف شغل ديكور</button>}
      </div>
      {rows.map((d) => {
        const notDone = d.status !== 'COMPLETED' && d.status !== 'CANCELLED'
        return (
          <p key={d.id} className="mt-1.5 text-xs text-slate-700">
            <b>{d.code}</b> · {d.decorTiming === 'BEFORE' ? 'قبل الشعبة' : 'ويا الشعبة'} · {fmt(d.scheduledAt)} · {STATUS[d.status] ?? d.status}
            {notDone && d.decorTiming === 'BEFORE' && <span className="ms-1 font-bold text-red-700">⚠️ الديكور بعدهم ما خلصوا</span>}
          </p>
        )
      })}
      {open && (
        <div className="mt-2 grid gap-2 sm:grid-cols-2">
          <select value={form.serviceId} onChange={(e) => setForm({ ...form, serviceId: e.target.value })} className="rounded-lg border border-slate-300 px-2 py-1.5">
            <option value="">المهنة (حداد، نجار…)</option>
            {services.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
          </select>
          <select value={form.timing} onChange={(e) => setForm({ ...form, timing: e.target.value as 'BEFORE' | 'WITH' })} className="rounded-lg border border-slate-300 px-2 py-1.5">
            <option value="BEFORE">يطلعون قبل الشعبة (افتراضياً قبل بيوم)</option>
            <option value="WITH">يطلعون ويا الشعبة نفس الموعد</option>
          </select>
          <input type="datetime-local" value={form.at} onChange={(e) => setForm({ ...form, at: e.target.value })} className="rounded-lg border border-slate-300 px-2 py-1.5" title="فارغ = ينحسب من موعد الحجز" />
          <input value={form.notes} onChange={(e) => setForm({ ...form, notes: e.target.value })} placeholder="شنو الشغل؟ (مثلاً قاعدة للداتا شو)" className="rounded-lg border border-slate-300 px-2 py-1.5" />
          <div className="flex gap-2 sm:col-span-2">
            <button type="button" disabled={busy} onClick={add} className="rounded-lg bg-orange-600 px-4 py-1.5 text-xs font-bold text-white disabled:opacity-50">{busy ? '…' : 'سجّل حجز الديكور'}</button>
            <button type="button" onClick={() => setOpen(false)} className="rounded-lg border border-slate-300 bg-white px-4 py-1.5 text-xs">إلغاء</button>
          </div>
        </div>
      )}
      {err && <p className="mt-1 text-xs font-bold text-red-600">{err}</p>}
    </div>
  )
}
