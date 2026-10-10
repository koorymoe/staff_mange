import { useEffect, useState } from 'react'
import { api } from '../api'

// «خلص دوامك — سجّل انصرافك» — الشي الوحيد من الدوام الي يبقى بالرئيسية
// (قرار (ع) 10-06). التفاصيل كلها بـ«جدول دوامي».
export default function ShiftEndCard() {
  const [show, setShow] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)
  const [done, setDone] = useState(false)
  useEffect(() => { void api.getAttendanceGate().then((g) => { if (g.afterShift && g.open) setShow(g.endLabel) }).catch(() => {}) }, [])
  if (!show) return null
  if (done) return <p dir="rtl" className="rounded-2xl bg-emerald-50 p-3 text-sm font-bold text-emerald-800">✅ انسجّل انصرافك. الله وياك.</p>
  const out = async () => { setBusy(true); try { await api.checkOut(); setDone(true) } catch { /* يبقى الزر */ } finally { setBusy(false) } }
  return (
    <div dir="rtl" className="flex flex-wrap items-center justify-between gap-2 rounded-2xl border border-rose-200 bg-rose-50 p-3">
      <p className="text-sm font-bold text-rose-900">🤖 خلص دوامك الساعة {show} — إذا خلّصت شغلك سجّل انصرافك.</p>
      <button type="button" disabled={busy} onClick={() => void out()} className="rounded-xl bg-rose-600 px-4 py-2 text-sm font-bold text-white disabled:opacity-50">🚪 سجّل انصرافي</button>
    </div>
  )
}
