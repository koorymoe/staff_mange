import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api'

// ═══ طابور الإداري: ورق الأعمال داخل الشركة (قرار (ع) 10-07) ═══
// «من يكمل الحجز الي داخل الشركة الإداري لازم يسوي الفاتورة والتقرير» —
// الفاتورة يدوية: هو يكتب شنو انعمل وهو يحدد السعر.
export default function InternalPaperworkQueue() {
  const [rows, setRows] = useState<Awaited<ReturnType<typeof api.getInternalPaperwork>> | null>(null)
  useEffect(() => {
    let alive = true
    api.getInternalPaperwork().then((r) => { if (alive) setRows(r) }).catch(() => { if (alive) setRows([]) })
    return () => { alive = false }
  }, [])
  if (!rows || rows.length === 0) return null
  return (
    <div dir="rtl" className="rounded-2xl border border-indigo-200 bg-indigo-50/60 p-4">
      <p className="font-extrabold text-indigo-900">🏢 أعمال داخل الشركة تنتظر ورقك ({rows.length})</p>
      <p className="mb-3 text-xs text-indigo-800">الفني خلّص الشغل — عليك الفاتورة (تكتب شنو انعمل وتحدد السعر) والتقرير.</p>
      <div className="space-y-2">
        {rows.map((r) => (
          <div key={r.id} className="flex flex-wrap items-center justify-between gap-2 rounded-xl bg-white px-3 py-2 text-sm">
            <span><b>{r.code}</b>{r.department && <span className="text-slate-500"> · {r.department}</span>}{r.doneAt && <span className="text-xs text-slate-400"> · خلص {r.doneAt}</span>}</span>
            <span className="flex gap-2">
              {r.hasInvoice
                ? <span className="rounded-lg bg-emerald-50 px-2 py-1 text-xs font-bold text-emerald-700">✓ الفاتورة</span>
                : <Link to={`/leader-invoices/new?internal=1&bookingId=${r.id}`} className="rounded-lg bg-indigo-600 px-3 py-1.5 text-xs font-bold text-white">🧾 سوّي الفاتورة</Link>}
              {r.hasReport
                ? <span className="rounded-lg bg-emerald-50 px-2 py-1 text-xs font-bold text-emerald-700">✓ التقرير</span>
                : <Link to={`/work-reports?bookingId=${r.id}`} className="rounded-lg border border-indigo-300 bg-white px-3 py-1.5 text-xs font-bold text-indigo-800">📝 سوّي التقرير</Link>}
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}
