import { useEffect, useState, type ReactNode } from 'react'
import { useSearchParams } from 'react-router-dom'
import { api, type WeeklyMover, type WeeklyReport, type WeekTotals } from '../api'
import { useSession } from '../session'

/**
 * ═══ 📊 تقرير المالك الأسبوعي (ماتركس) ═══
 *
 * أسبوع ISO (الاثنين → الأحد، توقيت بغداد) مقابل الي قبله. الأسبوع
 * الحالي ينقارن بنفس المدة من الأسبوع الماضي. ينحسب لحظياً.
 * ⚠️ الحارس نفس الخادم: ADMIN/OWNER بس (المالك يوصل هنا كـ ADMIN).
 */

const fmt = (n: number) => Math.round(n).toLocaleString('en-US')

// أعلى = أحسن لهاي المؤشرات؟
const METRICS: { key: keyof WeekTotals; label: string; money?: boolean; higherIsBetter: boolean }[] = [
  { key: 'completed', label: 'حجوزات منجزة', higherIsBetter: true },
  { key: 'revenue', label: 'المحصّل', money: true, higherIsBetter: true },
  { key: 'complaints', label: 'شكاوى', higherIsBetter: false },
  { key: 'workStops', label: 'توقفات شغل', higherIsBetter: false },
  { key: 'latePaperwork', label: 'ورق متأخر (+٤٨ ساعة)', higherIsBetter: false },
]

function Delta({ cur, prev, higherIsBetter }: { cur: number; prev: number; higherIsBetter: boolean }) {
  if (cur === prev) return <span className="text-xs text-slate-400">نفسه</span>
  const up = cur > prev
  const good = up === higherIsBetter
  const pct = prev > 0 ? ` (${up ? '+' : ''}${Math.round(((cur - prev) / prev) * 100)}٪)` : ''
  return <span className={`text-xs font-bold ${good ? 'text-emerald-600' : 'text-red-600'}`}>{up ? '▲' : '▼'} {fmt(Math.abs(cur - prev))}{pct}</span>
}

function Card({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section className="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <h3 className="mb-3 font-extrabold text-[#0f2040]">{title}</h3>
      {children}
    </section>
  )
}

function Movers({ rows, empty }: { rows: WeeklyMover[]; empty: string }) {
  if (rows.length === 0) return <p className="text-sm text-slate-400">{empty}</p>
  return (
    <ol className="space-y-1.5 text-sm">
      {rows.map((m) => (
        <li key={m.employeeId} className="flex justify-between gap-2">
          <span className="text-slate-700">{m.name}</span>
          <span className="tabular-nums text-slate-500">{m.lastScore} ← {m.thisScore} <b className={m.delta > 0 ? 'text-emerald-600' : 'text-red-600'}>({m.delta > 0 ? '+' : ''}{m.delta})</b></span>
        </li>
      ))}
    </ol>
  )
}

export default function WeeklyReportPage() {
  const { employee } = useSession()
  const allowed = employee?.role === 'ADMIN'
  const [params, setParams] = useSearchParams()
  const week = params.get('week') || ''
  const [data, setData] = useState<WeeklyReport | null>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!allowed) return
    let alive = true
    api.getWeeklyReport(week || undefined)
      .then((r) => { if (alive) { setData(r); setError('') } })
      .catch((e: Error) => { if (alive) setError(e.message) })
    return () => { alive = false }
  }, [allowed, week])

  if (!allowed) return null

  return (
    <div dir="rtl" className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="text-2xl font-extrabold text-[#0f2040]">📊 التقرير الأسبوعي</h2>
          <p className="text-sm text-slate-500">هذا الأسبوع مقابل الي قبله — أرقام من النظام نفسه، بلا تقدير.</p>
        </div>
        {data && (
          <select value={data.week} onChange={(e) => setParams({ week: e.target.value })}
            className="rounded-xl border border-slate-300 bg-white px-3 py-2 text-sm">
            {(data.weeks.includes(data.week) ? data.weeks : [data.week, ...data.weeks]).map((w) => <option key={w} value={w}>{w}</option>)}
          </select>
        )}
      </div>

      {error && <p className="rounded-lg bg-red-50 p-3 text-sm text-red-600">{error}</p>}
      {!data && !error && <div className="h-24 animate-pulse rounded-2xl bg-slate-100" />}

      {data && (
        <>
          {data.partial && (
            <p className="rounded-xl bg-sky-50 px-3 py-2 text-xs text-sky-800">
              الأسبوع بعده ما خلص — المقارنة بنفس المدة من الأسبوع الماضي.
            </p>
          )}
          <div className="grid grid-cols-2 gap-3 md:grid-cols-5">
            {METRICS.map((m) => (
              <div key={m.key} className="rounded-2xl border border-slate-200 bg-white p-3">
                <p className="text-xs text-slate-500">{m.label}</p>
                <p className="mt-1 text-xl font-extrabold tabular-nums text-slate-800">{fmt(data.this[m.key])}{m.money ? ' د.ع' : ''}</p>
                <p className="mt-0.5 text-[11px] text-slate-400">قبل: {fmt(data.last[m.key])}{m.money ? ' د.ع' : ''} · <Delta cur={data.this[m.key]} prev={data.last[m.key]} higherIsBetter={m.higherIsBetter} /></p>
              </div>
            ))}
          </div>

          <div className="grid gap-4 lg:grid-cols-3">
            <Card title="📈 أكثر تحسّن">
              <Movers rows={data.improving} empty="ماكو تحسّن واضح هالأسبوع." />
            </Card>
            <Card title="📉 أكثر تراجع">
              <Movers rows={data.declining} empty="ماكو تراجع واضح هالأسبوع ✅" />
            </Card>
            <Card title="⏳ قرارات تنتظرك (هسه)">
              <ul className="space-y-1.5 text-sm text-slate-700">
                <li className="flex justify-between"><span>طلبات إجازة</span><b className="tabular-nums">{fmt(data.openDecisions.pendingLeaves)}</b></li>
                <li className="flex justify-between"><span>طلبات حذف حجوزات</span><b className="tabular-nums">{fmt(data.openDecisions.deleteRequests)}</b></li>
                <li className="flex justify-between"><span>أحكام ماتركس بلا مراجعة</span><b className="tabular-nums">{fmt(data.openDecisions.unreviewedVerdicts)}</b></li>
                {data.matrix && data.matrix.actions > 0 && (
                  <li className="border-t border-slate-100 pt-1 text-xs text-slate-600">🤖 ماتركس (٣٠ يوم): {fmt(data.matrix.actions)} تذكير · انحل {fmt(data.matrix.resolved)} · صعد {fmt(data.matrix.escalated)} · رفضت {fmt(data.matrix.rejected)}</li>
                )}
              </ul>
            </Card>
          </div>
          <p className="text-[11px] text-slate-400">
            نقاط الموظف = منجز − ٢×توقف − ٣×شكوى لكل أسبوع. توجيه للمتابعة — ماكو غرامة ولا نقاط تنطلع منها.
          </p>
        </>
      )}
    </div>
  )
}
