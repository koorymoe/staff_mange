import { useEffect, useState } from 'react'
import { api, type MatrixBusiness as Biz } from '../api'

// ═══ أرقام الشركة بعين ماتركس ═══
const iqd = (n: number) => `${Math.round(n).toLocaleString('en-US')} د.ع`

export default function MatrixBusiness() {
  const [b, setB] = useState<Biz | null>(null)
  const [err, setErr] = useState<string | null>(null)
  useEffect(() => {
    let alive = true
    api.getMatrixBusiness().then((r) => { if (alive) setB(r) }).catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذر') })
    return () => { alive = false }
  }, [])
  if (err) return <p className="rounded-lg bg-red-50 p-3 text-red-600">{err}</p>
  if (!b) return <p className="text-slate-400">ماتركس يحسب الأرقام…</p>
  const max = Math.max(1, ...b.months.map((m) => m.revenue))
  const delta = b.mtd.lastRevenue > 0 ? Math.round(((b.mtd.revenue - b.mtd.lastRevenue) / b.mtd.lastRevenue) * 100) : null
  const c = b.customers
  const totalCust = c.actual + c.inquiry + c.openOnly
  return (
    <section className="rounded-2xl border border-slate-200 bg-white p-4 text-sm">
      <h3 className="mb-3 font-extrabold text-[#0f2040]">💰 أرقام الشركة بعين ماتركس</h3>
      <div className="grid gap-3 md:grid-cols-3">
        <div className="rounded-xl bg-slate-50 p-3">
          <p className="text-xs text-slate-500">هالشهر لحد اليوم</p>
          <p className="text-xl font-extrabold tabular-nums">{iqd(b.mtd.revenue)}</p>
          <p className="text-xs text-slate-500">{b.mtd.bookings} حجز منجز · الشهر الماضي بنفس الفترة {iqd(b.mtd.lastRevenue)}
            {delta !== null && <b className={delta >= 0 ? ' text-emerald-700' : ' text-red-700'}> ({delta >= 0 ? '+' : ''}{delta}٪)</b>}</p>
        </div>
        <div className="rounded-xl bg-violet-50 p-3">
          <p className="text-xs text-violet-700">🔮 توقع نهاية الشهر</p>
          <p className="text-xl font-extrabold tabular-nums text-violet-900">{iqd(b.forecast.expected)}</p>
          <p className="text-[11px] text-violet-800">{b.forecast.expectedJobs} حجز · {b.forecast.basis}</p>
        </div>
        <div className="rounded-xl bg-slate-50 p-3">
          <p className="text-xs text-slate-500">الزبائن</p>
          <p className="text-sm"><b className="text-emerald-700">{c.actual}</b> فعلي (نفّذ) · <b className="text-amber-700">{c.inquiry}</b> استفسار بس · <b>{c.openOnly}</b> حجزه مفتوح</p>
          {totalCust > 0 && <p className="text-xs text-slate-500">نسبة الفعلي {Math.round((c.actual / totalCust) * 100)}٪ · <b className="text-red-700">{c.repeat}</b> مستفسر متكرر</p>}
        </div>
      </div>
      <div className="mt-4">
        <p className="mb-1 text-xs font-bold text-slate-500">الإيراد آخر ٦ أشهر (صافي فواتير الحجوزات المنجزة)</p>
        <div className="space-y-1">
          {b.months.map((m) => (
            <div key={m.month} className="flex items-center gap-2 text-xs">
              <span className="w-16 tabular-nums text-slate-500">{m.month}</span>
              <div className="h-4 flex-1 rounded bg-slate-100"><div className="h-4 rounded bg-brand-500" style={{ width: `${(m.revenue / max) * 100}%` }} /></div>
              <span className="w-40 text-left tabular-nums">{iqd(m.revenue)} · {m.bookings} حجز{m.invoiced < m.bookings ? ` (${m.bookings - m.invoiced} بلا فاتورة)` : ''}</span>
            </div>
          ))}
          {b.months.length === 0 && <p className="text-xs text-slate-400">ماكو حجوزات منجزة بآخر ٦ أشهر.</p>}
        </div>
      </div>
      {b.inquirers.length > 0 && (
        <div className="mt-4">
          <p className="mb-1 text-xs font-bold text-slate-500">مستفسرين متكررين (سجّلنالهم حجزين أو أكثر وما نفّذوا)</p>
          <div className="flex flex-wrap gap-1">
            {b.inquirers.map((q) => <span key={q.id} className="rounded-full bg-amber-50 px-2 py-0.5 text-xs text-amber-900">{q.name} · {q.archived} مرات · آخرها {q.lastAt}</span>)}
          </div>
        </div>
      )}
    </section>
  )
}
