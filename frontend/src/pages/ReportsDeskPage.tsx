import { Suspense, useEffect, useState } from 'react'
import { api, type ProgressRecentRow } from '../api'
import BookingProgressTimeline from '../components/BookingProgressTimeline'
import { lazyRetry } from '../utils/staleReload'
import PctWhy from '../ui/PctWhy'

import { notify } from '../ui/notify'
// ═══ «📑 تقارير الحجوزات والمشاريع» — (ع) 10-16 ═══
// «وين المراقب يشوف التقارير مال الحجوزات ومال المشاريع؟» — بمكان واحد:
// تقرير العمل لكل حجز منجز، والتقرير اليومي لكل مشروع (يوم يوم) — نفس
// نموذج «تقرير تنفيذ المشروع» الي يملّيه الليدر.
const WorkReportsReview = lazyRetry(() => import('./WorkReportsReview'))

type View = 'bookings' | 'projects'

export default function ReportsDeskPage() {
  const [view, setView] = useState<View>('bookings')
  const btn = (v: View, label: string, hint: string) => (
    <button type="button" onClick={() => setView(v)}
      className={`rd-tab flex-1 rounded-2xl border px-4 py-3 text-right transition ${view === v ? 'rd-tab-on' : ''}`}>
      <span className="block text-base font-extrabold">{label}</span>
      <span className="rd-hint block text-xs">{hint}</span>
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
        <div key={r.bookingId} className="rd-row rounded-2xl border">
          {/* (ك) 11-23 زينا: الصف مثل الصورة — أعمدة ثابتة بفواصل. الألوان بـindex.css (.rd-*) */}
          <button type="button" onClick={() => setOpen(open === r.bookingId ? null : r.bookingId)}
            className="flex w-full flex-wrap items-center gap-3 px-4 py-2 text-right text-sm md:grid md:min-h-12 md:grid-cols-[6rem_minmax(0,1fr)_12rem_1px_6.5rem_1px_10rem_1px_13rem_1.5rem] md:gap-4">
            <span className="rd-code justify-self-start rounded-lg px-3 py-0.5 font-extrabold" dir="ltr">{r.bookingCode}</span>
            <span className="flex min-w-0 items-center gap-2 font-bold" title={r.customerName ?? undefined}>
              <svg viewBox="0 0 24 24" className="h-5 w-5 shrink-0" aria-hidden="true"><circle cx="12" cy="8" r="4.5" fill="#f5b041" /><path d="M3.5 21a8.5 8.5 0 0 1 17 0Z" fill="#f39c12" /></svg>
              <span className="truncate">المدير {r.leaderName ?? '—'}</span>
            </span>
            <span className="rd-rep text-center font-extrabold">تقرير {r.reports}: آخرها اليوم {r.lastDay}</span>
            <span className="rd-div hidden h-6 md:block" />
            <span className="rd-pct justify-self-center rounded-full px-3 py-0.5 font-extrabold"><PctWhy title={`إنجاز ${r.bookingCode}`} value={r.lastPercent}
              formula={<>هذي مو حسبة — هي «نسبة الإنجاز» الي كتبها الفريق بآخر تقرير يومي (اليوم {r.lastDay}) لهذا الحجز. انكتب للحجز {r.reports} تقرير لحد هسه.</>}
              meaning={<>شكد خلص من الشغل حسب تقدير الفريق بالموقع. 100% يعني خلص، أقل يعني بعده شغّال. {r.unreviewed > 0 ? `بعد ${r.unreviewed} تقرير ما متدقّق، فالرقم ما مأكد من الإدارة.` : 'كل تقاريره متدقّقة.'}</>}
              source={<>جدول التقارير اليومية للحجوزات: نأخذ نسبة الإنجاز من أحدث تقرير انكتب لهذا الحجز.</>} /></span>
            <span className="rd-div hidden h-6 md:block" />
            {r.unreviewed > 0
              ? <span className="rd-wait justify-self-center whitespace-nowrap rounded-full px-3 py-0.5 font-extrabold">⏳ {r.unreviewed} تنتظر التدقيق</span>
              : <span className="rd-ok justify-self-center whitespace-nowrap rounded-full px-3 py-0.5 font-extrabold">✅ متدقّقة</span>}
            <span className="rd-div hidden h-6 md:block" />
            <span className="rd-date ms-auto text-sm font-bold md:ms-0">{new Date(r.lastAt).toLocaleString('ar-IQ', { timeZone: 'Asia/Baghdad' })}</span>
            <span className="rd-tri text-base">{open === r.bookingId ? '▲' : '▼'}</span>
          </button>
          {open === r.bookingId && (
            <div className="rd-open border-t p-3">
              {r.unreviewed > 0 && (
                <button type="button" onClick={async () => {
                  try { await api.reviewProjectReports(r.bookingId); setRows((rs) => (rs ?? []).map((x) => (x.bookingId === r.bookingId ? { ...x, unreviewed: 0 } : x))) }
                  catch (e) { notify(e instanceof Error ? e.message : 'تعذر') }
                }} className="mb-3 rounded-lg bg-emerald-600 px-4 py-1.5 text-sm font-bold text-white">✅ تم التدقيق ({r.unreviewed} تقرير)</button>
              )}
              <BookingProgressTimeline bookingId={r.bookingId} />
            </div>
          )}
        </div>
      ))}
    </div>
  )
}
