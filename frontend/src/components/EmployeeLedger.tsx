import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type EmployeeLedger as Ledger } from '../api'

// ═══ «📒 الحصيلة» — شنو سوّى هذا الموظف اليوم/هذا الشهر بالتفصيل (ع 10-16) ═══
// كل قسم من الجداول نفسها: أي حجز طلع وشنو خدمته ومراحله، جرد، حضور، تقييم،
// فواتير، تقارير… والمراقب شنو دقّق. القسم الفارغ ما يطلع.
const fmt = (s: string) => new Date(s.endsWith('Z') ? s : s + 'Z').toLocaleString('ar-IQ', { timeZone: 'Asia/Baghdad', day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' })

export default function EmployeeLedger({ employeeId, day }: { employeeId: string; day: string }) {
  const [period, setPeriod] = useState<'day' | 'month'>('day')
  const [d, setD] = useState<Ledger | null>(null)
  const [err, setErr] = useState('')
  const [open, setOpen] = useState<string | null>(null)
  useEffect(() => {
    let alive = true
    api.getEmployeeLedger(employeeId, period, day).then((x) => { if (alive) { setD(x); setErr('') } })
      .catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذر') })
    return () => { alive = false }
  }, [employeeId, period, day])
  return (
    <section className="rounded-2xl border-2 border-violet-200 bg-white p-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h3 className="text-lg font-extrabold text-[#0f2040]">📒 الحصيلة — شنو سوّى {period === 'day' ? 'بهاليوم' : 'بهذا الشهر'}{d ? ` (${d.total})` : ''}</h3>
        <div className="flex gap-1">
          {(['day', 'month'] as const).map((p) => (
            <button key={p} type="button" onClick={() => setPeriod(p)}
              className={`rounded-lg px-3 py-1.5 text-sm font-bold ${period === p ? 'bg-violet-700 text-white' : 'border border-slate-300 text-slate-600'}`}>
              {p === 'day' ? 'اليوم' : 'الشهر'}
            </button>
          ))}
        </div>
      </div>
      {err && <p className="mt-2 text-sm text-red-600">{err}</p>}
      {!d && !err && <p className="mt-2 text-sm text-slate-400">جاري الحساب…</p>}
      {d && d.sections.length === 0 && <p className="mt-3 text-sm text-slate-500">ماكو أي شغل مسجّل إله {period === 'day' ? 'بهاليوم' : 'بهذا الشهر'}.</p>}
      {d && d.sections.length > 0 && (
        <>
          <div className="mt-3 flex flex-wrap gap-2">
            {d.sections.map((s) => (
              <button key={s.key} type="button" onClick={() => setOpen(open === s.key ? null : s.key)}
                className={`rounded-xl px-3 py-2 text-sm font-bold ${open === s.key ? 'bg-violet-700 text-white' : 'bg-violet-50 text-violet-900 hover:bg-violet-100'}`}>
                {s.title} <span className="tabular-nums">({s.count})</span>
              </button>
            ))}
          </div>
          <div className="mt-3 space-y-3">
            {d.sections.filter((s) => !open || s.key === open).map((s) => (
              <div key={s.key}>
                <p className="mb-1 text-sm font-extrabold text-slate-700">{s.title} — {s.count}</p>
                <ul className="divide-y divide-slate-100 rounded-lg border border-slate-200 text-sm">
                  {s.items.map((it, i) => (
                    <li key={i} className="flex flex-wrap items-center gap-2 px-3 py-1.5">
                      <span className="w-28 shrink-0 text-xs tabular-nums text-slate-400">{fmt(it.at)}</span>
                      <span className="flex-1 text-slate-800">{it.text}</span>
                      {it.link && <Link to={it.link} className="text-xs font-bold text-brand-700">افتح ←</Link>}
                    </li>
                  ))}
                </ul>
              </div>
            ))}
          </div>
        </>
      )}
    </section>
  )
}
