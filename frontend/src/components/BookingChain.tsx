import { useEffect, useState } from 'react'
import { api, type BookingChain } from '../api'
import MatrixNote from './MatrixNote'
import { CHAIN_TONE } from './chainTone'

// ═══ ماتركس ٢٠٥٠ — «سلسلة الحجز» ═══
// ١٤ محطة من تسجيل الحجز لحكم المراقب: منو المسؤول، شكد أخذت، والحد
// (وسيط الشركة) معلن ويا كل حكم. عرض بس — ماكو غرامة ولا نقاط.


const dt = (s: string | null) => (s ? new Date(s).toLocaleString('en-GB', { day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' }) : '')

export default function BookingChainView({ bookingId }: { bookingId: string }) {
  const [loaded, setLoaded] = useState<{ id: string; ch: BookingChain | null } | null>(null)
  const loading = loaded?.id !== bookingId
  const ch = loading ? null : loaded.ch

  useEffect(() => {
    let alive = true
    void api.getBookingChain(bookingId).catch(() => null).then((d) => { if (alive) setLoaded({ id: bookingId, ch: d }) })
    return () => { alive = false }
  }, [bookingId])

  if (loading) return <p className="text-xs text-slate-400">ماتركس يبني سلسلة الحجز…</p>
  if (!ch) return null
  const s = ch.score

  return (
    <div dir="rtl" className="mt-3 rounded-xl border border-slate-200 bg-white p-4">
      <div className="mb-2 flex flex-wrap items-center justify-between gap-2">
        <h4 className="text-sm font-bold text-[#0f2040]">🔗 سلسلة الحجز {ch.code}{ch.service && <span className="font-normal text-slate-500"> · {ch.service}{ch.solo ? ' (فني وحده)' : ''}</span>}{ch.project && <span className="mr-1 rounded bg-lime-100 px-1.5 text-[11px] font-bold text-lime-800">🏗️ حجز مشروع</span>}</h4>
        <div className="flex flex-wrap gap-1 text-[11px]">
          <span className="rounded-full bg-emerald-100 px-2 py-0.5 text-emerald-800">✅ {s.ok}</span>
          {s.issue > 0 && <span className="rounded-full bg-amber-100 px-2 py-0.5 text-amber-800">⚠️ {s.issue}</span>}
          {s.late > 0 && <span className="rounded-full bg-orange-100 px-2 py-0.5 text-orange-800">🟠 {s.late}</span>}
          {s.missed > 0 && <span className="rounded-full bg-red-100 px-2 py-0.5 text-red-800">🔴 {s.missed}</span>}
          {s.waiting > 0 && <span className="rounded-full bg-sky-100 px-2 py-0.5 text-sky-800">⏳ {s.waiting}</span>}
        </div>
      </div>
      <MatrixNote className="mb-3">{ch.summary}</MatrixNote>
      <ol className="space-y-2">
        {ch.stations.map((st) => {
          const t = CHAIN_TONE[st.status]
          return (
            <li key={st.key} className={`rounded-lg border p-2.5 ${t.cls}`}>
              <div className="flex flex-wrap items-center justify-between gap-2">
                <p className="text-sm font-bold text-slate-800">{t.icon} {st.no}. {st.title}
                  <span className="ms-2 text-[11px] font-normal text-slate-500">عين {st.roleTitle}</span></p>
                <span className="text-[11px] text-slate-600">
                  {st.owners.length > 0 ? st.owners.map((o) => `${o.name}${o.status && o.status !== st.status ? ' ' + CHAIN_TONE[o.status].icon : ''}`).join('، ') : 'بلا مسؤول مسجّل'}
                </span>
              </div>
              <p className="mt-1 text-xs text-slate-700">{st.verdict}</p>
              {(st.startAt || st.endAt) && (
                <p className="text-[11px] text-slate-400">{dt(st.startAt)}{st.endAt && st.endAt !== st.startAt ? ` ← ${dt(st.endAt)}` : ''}</p>
              )}
              {st.facts.length > 0 && <p className="mt-0.5 text-[11px] text-slate-600">• {st.facts.join(' • ')}</p>}
              {st.issues.length > 0 && <p className="mt-0.5 text-[11px] font-bold text-amber-800">⚠️ {st.issues.join(' ')}</p>}
            </li>
          )
        })}
      </ol>
      {ch.ratings && ch.ratings.length > 0 && (
        <div className="mt-3 rounded-lg border border-slate-200 p-2 text-xs">
          <b>⭐ تقييم الليدر للفنيين:</b>
          {ch.ratings.map((r) => <p key={r.technicianId} className="text-slate-600">{'★'.repeat(r.score)}{'☆'.repeat(5 - r.score)} {r.note ?? ''}</p>)}
        </div>
      )}
    </div>
  )
}
