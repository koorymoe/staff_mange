import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api, type CBooking } from '../api'

// مراحل الحجز كما يفهمها الزبون
const STAGES = [
  { key: 'PENDING', label: 'استلمنا الطلب' },
  { key: 'CONFIRMED', label: 'تثبّت الموعد' },
  { key: 'IN_PROGRESS', label: 'قيد التنفيذ' },
  { key: 'COMPLETED', label: 'منجز' },
]
const stageIdx = (s: string) => (s === 'WAITING' ? 1 : s === 'PARTIAL' ? 2 : STAGES.findIndex((x) => x.key === s))
const date = (s?: string) => (s ? new Date(s).toLocaleDateString('ar-IQ', { weekday: 'short', day: 'numeric', month: 'long', year: 'numeric' }) : '')

export default function BookingsPage() {
  const nav = useNavigate()
  const [rows, setRows] = useState<CBooking[] | null>(null)
  const [tab, setTab] = useState<'open' | 'past'>('open')
  const [err, setErr] = useState('')

  useEffect(() => { api.bookings().then(setRows).catch((e) => { setErr(e.message); setRows([]) }) }, [])

  const closed = (b: CBooking) => b.status === 'COMPLETED' || b.status === 'CANCELLED'
  const list = (rows ?? []).filter((b) => (tab === 'past' ? closed(b) : !closed(b)))

  return (
    <div className="rise px-4 pt-6">
      <h1 className="text-2xl font-black text-ink">حجوزاتي</h1>
      <p className="text-sm text-slate-500">تابع حالة طلباتك، والسابقة كلها هنا</p>

      <div className="mt-4 grid grid-cols-2 rounded-2xl bg-slate-200/70 p-1">
        {(['open', 'past'] as const).map((t) => (
          <button key={t} onClick={() => setTab(t)} className={`rounded-xl py-2 text-sm font-black transition ${tab === t ? 'bg-white text-navy shadow-sm' : 'text-slate-500'}`}>
            {t === 'open' ? 'الحالية' : 'السابقة'} ({(rows ?? []).filter((b) => (t === 'past' ? closed(b) : !closed(b))).length})
          </button>
        ))}
      </div>

      {err && <p className="mt-4 rounded-2xl bg-rose-50 p-3 text-sm text-rose-700">{err}</p>}
      <div className="mt-4 space-y-3">
        {rows === null && [0, 1].map((i) => <div key={i} className="h-32 animate-pulse rounded-3xl bg-slate-200" />)}
        {rows && list.length === 0 && (
          <div className="rounded-3xl bg-white p-8 text-center shadow-sm">
            <p className="text-4xl">🗓️</p>
            <p className="mt-2 font-black text-ink">{tab === 'open' ? 'ماكو حجز حالي' : 'ماكو حجوزات سابقة'}</p>
            <button onClick={() => nav('/matrix', { state: { prefill: 'أريد أحجز' } })} className="mt-4 rounded-2xl bg-navy px-5 py-2.5 text-sm font-black text-white">احجز ويّا ماتركس</button>
          </div>
        )}
        {list.map((b) => {
          const idx = stageIdx(b.status)
          const cancelled = b.status === 'CANCELLED'
          return (
            <article key={b.id} className="rounded-3xl bg-white p-4 shadow-sm">
              <div className="flex items-start justify-between gap-2">
                <div className="min-w-0">
                  <p className="font-black text-ink">{b.services || 'خدمة'}</p>
                  <p className="text-xs text-slate-500">طلب بتاريخ {date(b.createdAt)}</p>
                </div>
                <span className="shrink-0 rounded-lg bg-slate-100 px-2 py-1 font-mono text-xs font-bold text-slate-600">{b.code}</span>
              </div>
              <p className={`mt-3 rounded-2xl px-3 py-2 text-sm font-bold ${cancelled ? 'bg-rose-50 text-rose-700' : b.status === 'COMPLETED' ? 'bg-emerald-50 text-emerald-700' : 'bg-amber-50 text-amber-800'}`}>
                {b.statusLabel}
              </p>
              {!cancelled && (
                <ol className="mt-3 flex items-center">
                  {STAGES.map((s, i) => (
                    <li key={s.key} className="flex flex-1 flex-col items-center text-center">
                      <div className="flex w-full items-center">
                        <span className={`h-0.5 flex-1 ${i === 0 ? 'opacity-0' : i <= idx ? 'bg-emerald-500' : 'bg-slate-200'}`} />
                        <span className={`grid h-6 w-6 place-items-center rounded-full text-[11px] font-black ${i <= idx ? 'bg-emerald-500 text-white' : 'bg-slate-200 text-slate-400'}`}>{i < idx || b.status === 'COMPLETED' ? '✓' : i + 1}</span>
                        <span className={`h-0.5 flex-1 ${i === STAGES.length - 1 ? 'opacity-0' : i < idx ? 'bg-emerald-500' : 'bg-slate-200'}`} />
                      </div>
                      <span className={`mt-1 text-[10px] ${i <= idx ? 'font-bold text-emerald-700' : 'text-slate-400'}`}>{s.label}</span>
                    </li>
                  ))}
                </ol>
              )}
              <div className="mt-3 space-y-1 text-xs text-slate-500">
                {b.scheduledAt && <p>📅 الموعد: <b className="text-slate-700">{date(b.scheduledAt)}</b></p>}
                {b.completedAt && <p>🏁 انجز: {date(b.completedAt)}</p>}
                {b.address && <p className="truncate">📍 {b.address}</p>}
                {b.fromApp && <p className="text-violet-700">👁️ انحجز عن طريق ماتركس</p>}
              </div>
            </article>
          )
        })}
      </div>
    </div>
  )
}
