import { useEffect, useState } from 'react'
import { api, type StaffScoreDetail } from '../api'
import { ScoreBreakdown } from '../pages/StaffScorePage'

// «🏅 تقييمي هالشهر» — الموظف يشوف رقمه وليش، وشلون يرفعه. ما يشوف غيره.
export default function MyScoreCard() {
  const [d, setD] = useState<StaffScoreDetail | null>(null)
  const [open, setOpen] = useState(false)
  useEffect(() => { void api.getMyScore().then((r) => { if (r?.score && r.score.max + r.score.humanCount > 0) setD(r.score) }).catch(() => {}) }, [])
  if (!d) return null
  const v = d.final ?? 0
  const color = v >= 80 ? '#059669' : v >= 60 ? '#d97706' : '#dc2626'
  const tip = d.topLosses[0]
  return (
    <div dir="rtl" className="rounded-2xl border border-slate-200 bg-white p-3 shadow-sm">
      <button type="button" onClick={() => setOpen(!open)} className="flex w-full items-center justify-between gap-3 text-right">
        <span>
          <b className="text-sm text-slate-800">🏅 تقييمي هالشهر</b>
          <span className="block text-[11px] text-slate-500">ماتركس ٦٠٪ + تقييم مسؤوليك ٤٠٪{tip && <> · أكثر شي نزّله: <b>{tip.title}</b></>}</span>
        </span>
        <b className="text-2xl" style={{ color }}>{d.final == null ? '—' : `${Math.round(d.final)}%`}</b>
      </button>
      {tip?.advice && !open && <p className="mt-1 text-xs text-violet-800">💡 {tip.advice}</p>}
      {open && <div className="mt-3"><ScoreBreakdown d={d} /></div>}
    </div>
  )
}
