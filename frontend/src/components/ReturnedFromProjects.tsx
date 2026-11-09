import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type ReturnedFromProjects as Row } from '../api'

// ═══ «↩️ رجعت من المشاريع» — (ع) 10-23 ═══
// «الحجوزات المرجوعة من ادارة المشاريع اين تذهب؟» چانت ترجع لمحطتها بلا أثر
// وتضيع. هنا تبقى بالسبب ومنو رجّعها، لحد ما المنسّق يضغط «✅ استلمته».
export default function ReturnedFromProjects({ onChanged }: { onChanged?: () => void }) {
  const [rows, setRows] = useState<Row[] | null>(null)
  const [error, setError] = useState('')
  const [reload, setReload] = useState(0)
  const [busy, setBusy] = useState<string | null>(null)
  useEffect(() => {
    let alive = true
    api.getReturnedFromProjects()
      .then((r) => { if (alive) { setRows(r); setError('') } })
      .catch((e: Error) => { if (alive) setError(e.message || 'تعذر جلب الحجوزات') })
    return () => { alive = false }
  }, [reload])
  const handled = (id: string) => {
    setBusy(id)
    api.markReturnHandled(id)
      .then(() => { setReload((n) => n + 1); onChanged?.() })
      .catch((e: Error) => setError(e.message))
      .finally(() => setBusy(null))
  }
  if (error) return <p role="alert" className="rounded-xl bg-red-50 p-3 text-sm font-bold text-red-700">{error}</p>
  if (!rows) return <p className="p-4 text-sm text-slate-500">جاري التحميل…</p>
  if (rows.length === 0) return <p className="rounded-xl bg-white p-6 text-center text-sm text-slate-500 ring-1 ring-slate-200">ماكو حجز راجع من إدارة المشاريع ينتظرك ✅</p>
  return (
    <ul className="space-y-2" dir="rtl">
      {rows.map((b) => (
        <li key={b.id} className="rounded-xl border border-amber-200 bg-amber-50/60 p-3">
          <div className="flex flex-wrap items-center gap-2">
            <b className="text-slate-900">حجز {b.code}</b>
            {b.customer && <span className="text-sm text-slate-700">{b.customer}</span>}
            {b.phone && <a href={`tel:${b.phone}`} className="text-sm text-sky-700" dir="ltr">{b.phone}</a>}
            <span className="ms-auto text-xs text-slate-500">رجع {new Date(b.returnedAt).toLocaleString('ar-IQ', { dateStyle: 'short', timeStyle: 'short' })}{b.by ? ` · ${b.by}` : ''}</span>
          </div>
          {b.note && <p className="mt-1 text-sm text-amber-900">السبب: {b.note}</p>}
          <div className="mt-3 flex gap-2">
            <Link to={`/bookings?focus=${encodeURIComponent(b.code)}`} className="flex min-h-11 flex-1 items-center justify-center rounded-xl border-2 border-sky-600 px-4 text-sm font-extrabold text-sky-700">افتح الحجز</Link>
            <button type="button" disabled={busy === b.id} onClick={() => handled(b.id)}
              className="min-h-11 flex-1 rounded-xl bg-emerald-600 px-4 text-sm font-extrabold text-white disabled:opacity-50">{busy === b.id ? '…' : '✅ استلمته'}</button>
          </div>
        </li>
      ))}
    </ul>
  )
}
