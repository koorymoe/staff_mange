import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type ReturnedFromProjects as Row } from '../api'

// ═══ «↩️ رجعت من المشاريع» — (ع) 10-23 ═══
// «الحجوزات المرجوعة من ادارة المشاريع اين تذهب؟» چانت ترجع لمحطتها بلا أثر
// وتضيع. هنا تبقى بالسبب ومنو رجّعها، لحد ما المنسّق يقرر.
// (ع) 11-10 بلاغ #29: قراران — يرجع للحجوزات، أو ينحوّل للإدارة الثانية (التقنيين).
// وما يطلع بخانة ثانية بنفس الوقت (ReturnedOpenSQL بالسيرفر).
export default function ReturnedFromProjects({ onChanged }: { onChanged?: () => void }) {
  const [rows, setRows] = useState<Row[] | null>(null)
  const [error, setError] = useState('')
  const [reload, setReload] = useState(0)
  const [busy, setBusy] = useState<string | null>(null)
  const [techFor, setTechFor] = useState<string | null>(null)
  const [cands, setCands] = useState<{ id: string; name: string; position: string }[] | null>(null)
  const [pick, setPick] = useState('')
  const openTech = (id: string) => {
    setTechFor(id); setPick('')
    if (!cands) api.getHandoverCandidates().then(setCands).catch((e: Error) => setError(e.message))
  }
  const toTech = (b: Row) => {
    if (!pick) return
    setBusy(b.id)
    api.handoverBooking(b.id, pick, `رجع من إدارة المشاريع${b.note ? ': ' + b.note : ''}`)
      .then(() => api.markReturnHandled(b.id))
      .then(() => { setTechFor(null); setReload((n) => n + 1); onChanged?.() })
      .catch((e: Error) => setError(e.message))
      .finally(() => setBusy(null))
  }
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
  if (rows.length === 0) return <p className="rounded-xl bg-white p-6 text-center text-sm text-slate-500 ring-1 ring-slate-200">ماكو حجز راجع من إدارة أخرى ينتظرك ✅</p>
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
              className="min-h-11 flex-1 rounded-xl bg-emerald-600 px-4 text-sm font-extrabold text-white disabled:opacity-50">{busy === b.id ? '…' : '↩️ يرجع للحجوزات'}</button>
            <button type="button" disabled={busy === b.id} onClick={() => (techFor === b.id ? setTechFor(null) : openTech(b.id))}
              className="min-h-11 flex-1 rounded-xl border-2 border-violet-600 px-4 text-sm font-extrabold text-violet-700 disabled:opacity-50">🛠️ لإدارة التقنيين</button>
          </div>
          {techFor === b.id && (
            <div className="mt-2 flex flex-wrap gap-2">
              <select value={pick} onChange={(e) => setPick(e.target.value)} aria-label="التقني"
                className="min-h-11 flex-1 rounded-xl border border-slate-300 bg-white px-3 text-sm">
                <option value="">{cands ? 'اختار التقني…' : 'جاري التحميل…'}</option>
                {cands?.map((c) => <option key={c.id} value={c.id}>{c.name}{c.position ? ` — ${c.position}` : ''}</option>)}
              </select>
              <button type="button" disabled={!pick || busy === b.id} onClick={() => toTech(b)}
                className="min-h-11 rounded-xl bg-violet-600 px-4 text-sm font-extrabold text-white disabled:opacity-50">حوّل</button>
            </div>
          )}
        </li>
      ))}
    </ul>
  )
}
