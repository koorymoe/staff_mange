import { lazy, Suspense, useEffect, useState } from 'react'
import { api, type ProgressRecentRow } from '../api'
import BookingProgressTimeline from '../components/BookingProgressTimeline'

// ═══ «📑 تقارير الحجوزات والمشاريع» — (ع) 10-16 ═══
// «وين المراقب يشوف التقارير مال الحجوزات ومال المشاريع؟» — بمكان واحد:
// تقرير العمل لكل حجز منجز، والتقرير اليومي لكل مشروع (يوم يوم) — نفس
// نموذج «تقرير تنفيذ المشروع» الي يملّيه الليدر.
const WorkReportsReview = lazy(() => import('./WorkReportsReview'))

type View = 'bookings' | 'projects'

export default function ReportsDeskPage() {
  const [view, setView] = useState<View>('bookings')
  const btn = (v: View, label: string, hint: string) => (
    <button type="button" onClick={() => setView(v)}
      className={`flex-1 rounded-2xl px-4 py-3 text-right transition ${view === v ? 'bg-[#0f2040] text-white shadow' : 'bg-white text-slate-700 ring-1 ring-slate-200 hover:bg-slate-50'}`}>
      <span className="block text-base font-extrabold">{label}</span>
      <span className={`block text-xs ${view === v ? 'text-blue-100' : 'text-slate-400'}`}>{hint}</span>
    </button>
  )
  return (
    <div className="space-y-4" dir="rtl">
      <div className="rounded-2xl bg-gradient-to-l from-[#0f2040] to-[#2c5aad] p-5 text-white shadow">
        <h1 className="text-2xl font-black">📑 تقارير الحجوزات والمشاريع</h1>
        <p className="mt-1 text-sm text-blue-100">التقرير الي يرفعه الليدر بعد كل حجز، والتقرير اليومي لكل مشروع — بنفس نموذج «تقرير تنفيذ المشروع».</p>
      </div>
      <div className="flex flex-col gap-2 sm:flex-row">
        {btn('bookings', '📋 تقارير الحجوزات', 'تقرير العمل الي يرفعه الليدر بعد ما يخلص الحجز')}
        {btn('projects', '🏗️ تقارير المشاريع اليومية', 'كل يوم شغل بالمشروع: شنو انعمل، شكد باقي، والمعرقلات')}
      </div>
      {view === 'bookings'
        ? <Suspense fallback={<p className="py-10 text-center text-slate-400">جاري التحميل…</p>}><WorkReportsReview /></Suspense>
        : <ProjectReports />}
    </div>
  )
}

function ProjectReports() {
  const [rows, setRows] = useState<ProgressRecentRow[] | null>(null)
  const [err, setErr] = useState('')
  const [open, setOpen] = useState<string | null>(null)
  useEffect(() => {
    let alive = true
    api.getRecentProgressReports().then((r) => { if (alive) setRows(r) })
      .catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذر جلب التقارير') })
    return () => { alive = false }
  }, [])
  if (err) return <p className="py-10 text-center text-red-600">{err}</p>
  if (!rows) return <p className="py-10 text-center text-slate-400">جاري التحميل…</p>
  if (rows.length === 0) return <p className="rounded-2xl bg-white py-12 text-center text-slate-500 ring-1 ring-slate-200">ماكو تقارير يومية بآخر ٣٠ يوم.</p>
  return (
    <div className="space-y-2">
      {rows.map((r) => (
        <div key={r.bookingId} className="rounded-2xl bg-white ring-1 ring-slate-200">
          <button type="button" onClick={() => setOpen(open === r.bookingId ? null : r.bookingId)}
            className="flex w-full flex-wrap items-center gap-3 px-4 py-3 text-right">
            <span className="font-extrabold text-[#0f2040]">{r.bookingCode}</span>
            <span className="text-sm text-slate-600">{r.customerName ?? '—'}</span>
            {r.leaderName && <span className="text-xs text-slate-500">👷 الليدر: {r.leaderName}</span>}
            <span className="rounded-full bg-sky-50 px-2 py-0.5 text-xs font-bold text-sky-700">{r.reports} تقرير · آخرها اليوم {r.lastDay}</span>
            <span className={`rounded-full px-2 py-0.5 text-xs font-bold ${r.lastPercent >= 100 ? 'bg-emerald-50 text-emerald-700' : 'bg-amber-50 text-amber-700'}`}>{r.lastPercent}%</span>
            <span className="ms-auto text-xs text-slate-400">{new Date(r.lastAt).toLocaleString('ar-IQ', { timeZone: 'Asia/Baghdad' })}</span>
            <span className="text-slate-400">{open === r.bookingId ? '▲' : '▼'}</span>
          </button>
          {open === r.bookingId && (
            <div className="border-t border-slate-100 p-3">
              <BookingProgressTimeline bookingId={r.bookingId} />
            </div>
          )}
        </div>
      ))}
    </div>
  )
}
