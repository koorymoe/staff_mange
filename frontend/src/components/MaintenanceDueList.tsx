import { useEffect, useState } from 'react'
import { apiRequest } from '../api'

// 🛠️ صيانة وقائية (دازاي 11-06، فكرة ٧): زبائن فات ٦ أشهر على آخر شغلهم، أو عندهم شكاوى متكررة،
// وماكو حجز مفتوح. الجودة تتصل وتعرض صيانة — قراءة بس.
type Row = { customerId: string; name: string; phone?: string | null; lastDone: string; service: string | null; months: number; complaints: number; reason: string }

export default function MaintenanceDueList() {
  const [rows, setRows] = useState<Row[] | null>(null)
  const [err, setErr] = useState('')
  useEffect(() => {
    apiRequest<Row[]>('/maintenance-due').then(setRows).catch((e) => setErr(e.message))
  }, [])
  if (err) return <p role="alert" className="text-rose-700">{err}</p>
  if (!rows) return <p className="py-10 text-center text-slate-400">جاري التحميل…</p>
  if (rows.length === 0) return <p className="rounded-2xl bg-white p-6 text-center text-slate-500">ماكو زبون يستاهل اتصال صيانة هسه ✅</p>
  return (
    <section className="space-y-2 rounded-2xl bg-white p-4 shadow-sm" aria-label="صيانة وقائية">
      <p className="text-sm text-slate-600">
        {rows.length} زبون — اتصل واعرض صيانة أو فحص. الي عنده <b>شكاوى متكررة</b> أول.
      </p>
      <ul className="divide-y divide-slate-100 text-sm">
        {rows.map((r) => (
          <li key={r.customerId} className="flex flex-wrap items-center gap-x-3 gap-y-1 py-2">
            <b className="text-slate-800">{r.name}</b>
            {r.phone && <a href={`tel:${r.phone}`} className="text-brand-700" dir="ltr">{r.phone}</a>}
            <span className="text-slate-500">{r.service ?? '—'}</span>
            <span className="text-slate-500">آخر شغل قبل {r.months} شهر</span>
            <span className={`rounded px-2 py-0.5 text-xs font-bold ${r.complaints >= 2 ? 'bg-rose-50 text-rose-700' : 'bg-amber-50 text-amber-800'}`}>{r.reason}</span>
          </li>
        ))}
      </ul>
    </section>
  )
}
