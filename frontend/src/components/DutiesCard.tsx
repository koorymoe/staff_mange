import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type DutiesReport } from '../api'

// ═══ «🎯 واجباتك اليوم» — (ع) 10-17 ═══
// كل منصب اله شغل يومي بالنظام. الكارت يبين شنو سوّى اليوم وشنو بعده ينتظر،
// والحكم: خوش / سوّى جزء وعاف جزء / ما سوّى واجباته. بلا غرامة — تنبيه بس.
const VERDICT: Record<DutiesReport['verdict'], string> = {
  GOOD: 'bg-emerald-50 text-emerald-700 ring-emerald-200',
  PARTIAL: 'bg-amber-50 text-amber-800 ring-amber-200',
  BAD: 'bg-red-50 text-red-700 ring-red-200',
}
const ICON = { DONE: '✅', PARTIAL: '🟡', MISSED: '⭕', INFO: 'ℹ️' } as const

export default function DutiesCard({ employeeId }: { employeeId?: string }) {
  const [r, setR] = useState<DutiesReport | null>(null)
  useEffect(() => {
    let alive = true
    ;(employeeId ? api.getEmployeeDuties(employeeId) : api.getMyDuties())
      .then((x) => { if (alive) setR(x) }).catch(() => {})
    return () => { alive = false }
  }, [employeeId])
  if (!r || !r.duties || r.duties.length === 0) return null
  return (
    <div className="rounded-2xl bg-white p-4 shadow-sm ring-1 ring-slate-200" dir="rtl">
      <div className="mb-3 flex flex-wrap items-center gap-2">
        <h2 className="text-lg font-extrabold text-[#0f2040]">🎯 {employeeId ? 'واجباته اليوم' : 'واجباتك اليوم'}</h2>
        <span className="text-xs text-slate-500">{r.doneCount} من {r.total}</span>
        <span className={`ms-auto rounded-full px-3 py-1 text-xs font-bold ring-1 ${VERDICT[r.verdict]}`}>{r.verdictAr}</span>
      </div>
      <ul className="space-y-1.5">
        {r.duties.map((d) => {
          const body = (
            <>
              <span>{ICON[d.status]}</span>
              <span className="font-bold text-slate-800">{d.label}</span>
              <span className="text-xs text-slate-500">سوّى {d.done}</span>
              {d.pending != null && d.pending > 0 && (
                <span className="rounded-full bg-amber-100 px-2 py-0.5 text-xs font-bold text-amber-800">⏳ {d.pending} ينتظر</span>
              )}
              {d.hint && <span className="hidden text-xs text-slate-400 sm:inline">— {d.hint}</span>}
            </>
          )
          const cls = 'flex flex-wrap items-center gap-2 rounded-xl px-3 py-2 text-sm hover:bg-slate-50'
          return (
            <li key={d.key}>
              {d.link && !employeeId ? <Link to={d.link} className={cls}>{body}</Link> : <div className={cls}>{body}</div>}
            </li>
          )
        })}
      </ul>
    </div>
  )
}
