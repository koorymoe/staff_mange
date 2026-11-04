import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type LeaveRequest } from '../api'

// ═══ «🗓️ إجازات الموظفين» — وحدة إدارة الكوادر (ع) 10-23 ═══
// «اريد بوحدة ادارة الكوادر خانه مخصصه لاجازات الموظفين». صندوق الإجازات
// (/leaves) يبين «تنتظر قرارك» بس؛ هنا السجل كله: منو مجاز اليوم، وكل طلبات
// الشهر بحالاتها. للعرض — القرار يبقى من الصندوق.

const STATUS_STYLE: Record<string, string> = {
  APPROVED: 'bg-emerald-50 text-emerald-700 ring-emerald-200',
  PENDING: 'bg-amber-50 text-amber-800 ring-amber-200',
  PRELIMINARY: 'bg-sky-50 text-sky-700 ring-sky-200',
  REJECTED: 'bg-red-50 text-red-700 ring-red-200',
  CANCELLED: 'bg-slate-100 text-slate-500 ring-slate-200',
}
const GROUPS: { key: string; label: string; match: (s: string) => boolean }[] = [
  { key: 'open', label: '⏳ تنتظر قرار', match: (s) => s === 'PENDING' || s === 'PRELIMINARY' },
  { key: 'approved', label: '✅ معتمدة', match: (s) => s === 'APPROVED' },
  { key: 'closed', label: '✖️ مرفوضة أو ملغية', match: (s) => s === 'REJECTED' || s === 'CANCELLED' },
]
const thisMonth = () => new Date(Date.now() + 3 * 3600_000).toISOString().slice(0, 7)
const d = (iso: string) => iso.slice(0, 10)

export default function StaffLeavesPage() {
  const [month, setMonth] = useState(thisMonth)
  const [data, setData] = useState<{ onLeaveToday: LeaveRequest[]; items: LeaveRequest[] } | null>(null)
  const [error, setError] = useState('')
  useEffect(() => {
    let alive = true
    api.getLeaveRegister(month)
      .then((r) => { if (alive) { setData(r); setError('') } })
      .catch((e: Error) => { if (alive) setError(e.message || 'تعذر جلب الإجازات') })
    return () => { alive = false }
  }, [month])
  const groups = useMemo(() => GROUPS.map((g) => ({ ...g, rows: (data?.items ?? []).filter((l) => g.match(l.status)) })), [data])
  const approvedDays = (data?.items ?? []).filter((l) => l.status === 'APPROVED').reduce((n, l) => n + (l.days || 0), 0)

  const card = (l: LeaveRequest) => (
    <li key={l.id} className="rounded-xl border border-slate-200 bg-white p-3">
      <div className="flex flex-wrap items-center gap-2">
        <b className="text-slate-800">{l.employeeName}</b>
        {l.jobTitle && <span className="text-xs text-slate-500">{l.jobTitle}</span>}
        <span className={`ms-auto rounded-full px-2.5 py-0.5 text-xs font-bold ring-1 ${STATUS_STYLE[l.status] ?? STATUS_STYLE.CANCELLED}`}>{l.statusLabel}</span>
      </div>
      <p className="mt-1 text-sm text-slate-700">
        {d(l.startDate)}{d(l.endDate) !== d(l.startDate) && <> ← {d(l.endDate)}</>} · <b>{l.days}</b> يوم · {l.kindLabel}
      </p>
      {l.reason && <p className="mt-1 text-sm text-slate-500">السبب: {l.reason}</p>}
      {l.decidedByName && <p className="mt-1 text-xs text-slate-400">القرار: {l.decidedByName}{l.decisionNote ? ` — ${l.decisionNote}` : ''}</p>}
    </li>
  )

  return (
    <div dir="rtl" className="mx-auto max-w-4xl space-y-4">
      <header className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="text-2xl font-extrabold text-[#0f2040]">🗓️ إجازات الموظفين</h1>
          <p className="text-sm text-slate-500">منو مجاز اليوم، وكل طلبات الشهر. القرار من «طلبات الإجازات».</p>
        </div>
        <label className="text-sm font-bold text-slate-600">الشهر
          <input type="month" value={month} onChange={(e) => setMonth(e.target.value || thisMonth())}
            className="mt-1 block rounded-lg border border-slate-300 bg-white p-2 text-base" />
        </label>
      </header>
      {error && <p role="alert" className="rounded-xl bg-red-50 p-3 text-sm font-bold text-red-700">{error}</p>}

      <section className="rounded-2xl bg-white p-4 shadow-sm ring-1 ring-slate-200">
        <h2 className="font-extrabold text-slate-800">🏖️ مجازين اليوم ({data?.onLeaveToday.length ?? 0})</h2>
        {data && data.onLeaveToday.length === 0 && <p className="mt-2 text-sm text-slate-500">ماكو أحد مجاز اليوم.</p>}
        <ul className="mt-2 flex flex-wrap gap-2">
          {data?.onLeaveToday.map((l) => (
            <li key={l.id} className="rounded-full bg-emerald-50 px-3 py-1.5 text-sm font-bold text-emerald-800 ring-1 ring-emerald-200">
              {l.employeeName} <span className="font-normal text-emerald-700">لحد {d(l.endDate)}</span>
            </li>
          ))}
        </ul>
      </section>

      <div className="grid grid-cols-3 gap-2 text-center">
        {groups.map((g) => (
          <div key={g.key} className="rounded-xl bg-white p-3 shadow-sm ring-1 ring-slate-200">
            <p className="text-xs font-bold text-slate-500">{g.label}</p>
            <p className="text-2xl font-black text-[#0f2040] tabular-nums">{g.rows.length}</p>
          </div>
        ))}
      </div>
      <p className="text-sm text-slate-600">مجموع أيام الإجازات المعتمدة بالشهر: <b>{approvedDays}</b> يوم</p>

      {groups.map((g) => g.rows.length > 0 && (
        <section key={g.key}>
          <div className="mb-2 flex items-center justify-between">
            <h2 className="font-extrabold text-slate-800">{g.label} ({g.rows.length})</h2>
            {g.key === 'open' && <Link to="/leaves" className="rounded-lg bg-brand-600 px-3 py-2 text-sm font-bold text-white">افتح طلبات الإجازات ←</Link>}
          </div>
          <ul className="space-y-2">{g.rows.map(card)}</ul>
        </section>
      ))}
      {data && data.items.length === 0 && <p className="rounded-xl bg-white p-6 text-center text-sm text-slate-500 ring-1 ring-slate-200">ماكو طلبات إجازة بهالشهر.</p>}
    </div>
  )
}
