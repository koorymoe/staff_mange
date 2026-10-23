import { useEffect, useState } from 'react'
import { api, type Booking } from '../api'
import TechHandovers from '../components/TechHandovers'
import { useSession } from '../session'

// ═══ «الحجوزات» للتقني ومسؤول الخدمة — قرار (ع) 10-07 ═══
// التقني ما عليه ليدر: الحجوزات الي يرحّلها الإداري إله تطلعله هنا، ببند
// مستقل بالقائمة اسمه «الحجوزات».
export default function MyHandoversPage() {
  const { employee } = useSession()
  const [rows, setRows] = useState<Booking[] | null>(null)
  const [err, setErr] = useState('')
  useEffect(() => {
    let alive = true
    api.getBookings({ assignedTo: 'me' })
      .then((b) => { if (alive) setRows(b) })
      .catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذر جلب الحجوزات') })
    return () => { alive = false }
  }, [])
  const mine = (rows ?? []).filter((b) => b.handoverToId === employee?.id)
  const done = mine.filter((b) => b.status === 'COMPLETED')
  return (
    <div dir="rtl" className="space-y-4">
      <h1 className="text-2xl font-bold text-brand-900">الحجوزات</h1>
      <p className="text-sm text-slate-500">الحجوزات الي رحّلها الإداري إلك. تواصل ويا الزبون، اكتب الكشف، وعالج — ماتركس يحسب وقت كل خطوة.</p>
      {err && <p className="text-sm text-red-600">{err}</p>}
      {!rows && !err && <p className="text-sm text-slate-400">…</p>}
      {rows && (
        <TechHandovers bookings={rows} me={employee?.id}
          onChange={(u) => setRows((prev) => (prev ?? []).map((x) => (x.id === u.id ? { ...x, ...u } : x)))} />
      )}
      {rows && mine.length - done.length === 0 && <p className="rounded-xl bg-slate-50 p-4 text-sm text-slate-500">ماكو حجز مفتوح مرحّل إلك هسه.</p>}
      {done.length > 0 && (
        <div className="rounded-xl border border-slate-200 bg-white p-4">
          <h3 className="mb-2 text-sm font-bold text-slate-700">✅ الي خلّصتها ({done.length})</h3>
          <ul className="space-y-1 text-sm">
            {done.slice(0, 30).map((b) => (
              <li key={b.id} className="flex flex-wrap justify-between gap-2 border-t border-slate-100 pt-1">
                <span><b>{b.code}</b> · {b.customer?.name ?? ''}</span>
                <span className="text-xs text-slate-500">{b.techDiagnosis}</span>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  )
}
