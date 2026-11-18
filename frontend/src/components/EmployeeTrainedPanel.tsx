import { useEffect, useState } from 'react'
import { api, type TcNeed } from '../api'

// (ك) 11-18 زينا: «📚 دُرّب على» بملف الموظف — المفردات الي اتدرب عليها من وحدة التدريب
// (الماتشينغ ← بدي التدريب ← خلص واكتب التقرير)، ويا التاريخ ومنو درّبه والتقييم.
const dt = (s: string | null) => (s ? new Date(s).toLocaleDateString('en-GB') : '—')

export default function EmployeeTrainedPanel({ employeeId }: { employeeId: string }) {
  const [rows, setRows] = useState<TcNeed[] | null>(null)
  const [error, setError] = useState('')
  useEffect(() => {
    let alive = true
    api.tcEmployeeTrained(employeeId)
      .then((r) => { if (alive) { setRows(r); setError('') } })
      .catch(() => { if (alive) { setRows([]); setError('تعذّر جلب التدريبات') } })
    return () => { alive = false }
  }, [employeeId])

  if (rows === null) return <p className="mt-5 text-xs text-slate-400">يحمّل…</p>
  return (
    <div className="mt-5" data-testid="trained-panel">
      <div className="mb-3 flex items-center gap-2">
        <div className="h-1.5 w-1.5 rounded-full bg-[#2c5aad]" />
        <h4 className="text-sm font-bold text-[#0f2040]">دُرّب على</h4>
        <span className="mr-auto rounded-full bg-blue-50 px-2.5 py-0.5 text-[10px] font-bold text-[#2c5aad]">{rows.length} تدريب</span>
      </div>
      {error && <p className="mb-2 text-xs text-red-600">{error}</p>}
      {rows.length === 0 && !error && <p className="text-xs text-slate-400">بعده ما خلص أي تدريب من وحدة التدريب.</p>}
      <ul className="space-y-2">
        {rows.map((n) => (
          <li key={n.id} className="rounded-xl border border-slate-200 p-3 text-sm">
            <div className="flex flex-wrap items-center gap-2">
              <b className="text-[#0f2040]">{n.topic}</b>
              {n.path && <span className="text-xs text-slate-500">📚 {n.path}</span>}
              <span className="ms-auto text-amber-500" aria-label={`${n.reportScore ?? 0} من ٥`}>
                {'★'.repeat(n.reportScore ?? 0)}<span className="text-slate-300">{'★'.repeat(5 - (n.reportScore ?? 0))}</span>
              </span>
            </div>
            {n.itemTitles.length > 0 && (
              <div className="mt-2 flex flex-wrap gap-1">
                {n.itemTitles.map((t, i) => (
                  <span key={i} className="rounded-full bg-blue-50 px-2 py-0.5 text-[11px] font-bold text-[#2c5aad]">✓ {t}</span>
                ))}
              </div>
            )}
            <p className="mt-2 text-xs text-slate-500">خلص {dt(n.doneAt)}{n.reportedBy ? ` · درّبه ${n.reportedBy}` : ''}</p>
            {n.reportText && <p className="mt-1 whitespace-pre-wrap rounded-lg bg-slate-50 p-2 text-xs text-slate-700">{n.reportText}</p>}
          </li>
        ))}
      </ul>
    </div>
  )
}
