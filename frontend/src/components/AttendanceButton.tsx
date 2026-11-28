import { useEffect, useState } from 'react'
import { api } from '../api'
import { notify } from '../ui/notify'

// (دازاي 11-28) «الموظف منين يكدر يسجل انصراف؟» — زر حضور/انصراف دائم بتبويب «دوامي» بالبروفايل،
// حتى لو يطلع قبل نهاية دوامه (جدول الدوام صار مو بالقائمة الجانبية).
export default function AttendanceButton({ onDone }: { onDone?: () => void }) {
  const [open, setOpen] = useState<boolean | null>(null)
  const [busy, setBusy] = useState(false)
  const load = () => api.getAttendanceGate().then((g) => setOpen(g.open)).catch(() => setOpen(null))
  useEffect(() => { void load() }, [])
  if (open === null) return null
  const act = async () => {
    setBusy(true)
    try {
      if (open) await api.checkOut(); else await api.checkIn()
      notify(open ? '✅ انسجّل انصرافك. الله وياك.' : '✅ انسجّل حضورك.', 'success')
      await load()
      onDone?.()
    } catch (e) { notify(e instanceof Error ? e.message : 'تعذر') } finally { setBusy(false) }
  }
  return (
    <div className="mb-3 flex flex-wrap items-center justify-between gap-2 rounded-xl bg-slate-50 p-3" data-testid="attendance-button">
      <p className="text-sm font-bold text-slate-700">{open ? '🟢 انت مسجّل حضور هسه' : '⚪ مو مسجّل حضور هسه'}</p>
      <button type="button" disabled={busy} onClick={() => void act()}
        className={`rounded-xl px-4 py-2 text-sm font-bold text-white disabled:opacity-50 ${open ? 'bg-rose-600' : 'bg-emerald-600'}`}>
        {open ? '🚪 سجّل انصرافي' : '✅ سجّل حضوري'}
      </button>
    </div>
  )
}
