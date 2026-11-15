import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type SurveyRow } from '../api'
import { loadFailed } from '../netErrors'

// ═══ 📋 كل الكشوفات (دازاي 11-15) ═══
// كشوفات المشاريع (مرحلة الكشف) + كشوفات الحجوزات (تقرير الفني) بمكان واحد.
// الإدارة ومدير المشاريع يشوفون الكل؛ غيرهم يشوف الي يخصه بس (السيرفر يفلتر).

type Kind = 'ALL' | 'PROJECT' | 'BOOKING'

export default function AllSurveys() {
  const [rows, setRows] = useState<SurveyRow[] | null>(null)
  const [kind, setKind] = useState<Kind>('ALL')
  const [q, setQ] = useState('')
  useEffect(() => { api.getSurveys().then(setRows).catch((e) => { setRows([]); loadFailed(e) }) }, [])

  const shown = useMemo(() => {
    const t = q.trim().toLowerCase()
    return (rows ?? []).filter((r) => (kind === 'ALL' || r.kind === kind) && (!t ||
      [r.code, r.title, r.customer, r.location, r.byName, r.ownerName, r.team, r.details].some((x) => x?.toLowerCase().includes(t))))
  }, [rows, kind, q])
  const count = (k: Kind) => (rows ?? []).filter((r) => k === 'ALL' || r.kind === k).length

  return (
    <div className="mt-6 rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
      <div className="flex flex-wrap items-center gap-2">
        {([['ALL', 'الكل'], ['PROJECT', '🏗️ كشوفات المشاريع'], ['BOOKING', '🔧 كشوفات الحجوزات']] as [Kind, string][]).map(([k, l]) => (
          <button key={k} type="button" onClick={() => setKind(k)}
            className={`rounded-full px-3 py-1 text-xs font-bold ring-1 ${kind === k ? 'bg-[#0f2040] text-white ring-[#0f2040]' : 'bg-white text-slate-600 ring-slate-200'}`}>
            {l} ({count(k)})
          </button>
        ))}
        <input value={q} onChange={(e) => setQ(e.target.value)} placeholder="ابحث: كود، زبون، مكان، موظف…"
          className="mr-auto min-w-[12rem] flex-1 rounded-lg border border-slate-300 px-3 py-1.5 text-sm sm:flex-none" />
      </div>
      {rows === null && <p className="mt-4 text-sm text-slate-400">جاري التحميل…</p>}
      {rows !== null && shown.length === 0 && <p className="mt-4 text-sm text-slate-400">ماكو كشوفات{q || kind !== 'ALL' ? ' بهالفلتر' : ' بعد'}.</p>}
      <div className="mt-3 flex flex-col gap-2">
        {shown.map((r) => (
          <div key={`${r.kind}-${r.id}`} className="rounded-lg border border-slate-100 p-3 text-sm">
            <div className="flex flex-wrap items-center gap-2">
              <span className={`rounded-full px-2 py-0.5 text-[11px] font-bold ${r.kind === 'PROJECT' ? 'bg-violet-50 text-violet-800' : 'bg-sky-50 text-sky-800'}`}>
                {r.kind === 'PROJECT' ? '🏗️ مشروع' : '🔧 حجز'}
              </span>
              {r.code && <span className="font-mono text-xs font-bold text-slate-700">{r.code}</span>}
              <span className="font-bold text-brand-900">{r.title || '—'}</span>
              {r.customer && <span className="text-slate-500">· {r.customer}</span>}
              {r.stage && <span className="rounded bg-slate-100 px-1.5 text-[11px] text-slate-600">{r.stage}</span>}
              {r.kind === 'PROJECT' && (
                <span className={`rounded px-1.5 text-[11px] font-bold ${r.hasAnswers ? 'bg-emerald-50 text-emerald-700' : 'bg-amber-50 text-amber-700'}`}>
                  {r.hasAnswers ? '✅ الكشف منجز' : '⏳ بعده ما انسوّى'}
                </span>
              )}
              <Link to={r.kind === 'PROJECT' ? '/projects' : `/bookings?focus=${r.refId}`} className="mr-auto text-xs font-bold text-sky-700 underline">
                افتح {r.kind === 'PROJECT' ? 'المشاريع' : 'الحجز'} ←
              </Link>
            </div>
            <div className="mt-1 flex flex-wrap gap-x-4 gap-y-0.5 text-xs text-slate-500">
              {r.byName && <span>{r.kind === 'PROJECT' ? 'منفّذ الكشف' : 'كتبه'}: {r.byName}</span>}
              {r.ownerName && <span>المسؤول: {r.ownerName}</span>}
              {r.team && <span>الفريق: {r.team}</span>}
              {r.location && <span>📍 {r.location}</span>}
              {r.when && <span>🕒 {r.when}</span>}
              {r.createdAt && <span>{new Date(r.createdAt).toLocaleDateString('ar-IQ')}</span>}
            </div>
            {r.details && <p className="mt-1 text-xs leading-5 text-slate-700">{r.details}</p>}
          </div>
        ))}
      </div>
    </div>
  )
}
