import { useEffect, useState } from 'react'
import { api, type StaffPointsRow } from '../api'

// ═══ 🏅 رصيد نقاط الشهر — قرار (ع) 10-11 ═══
// «تصاعدي بدون لمت»: الكي بي اي (كل بند كامل +٥) + الحضور (بوقته +١،
// متأخر +½) + شغل الدور (بلا حد). الترتيب داخل نفس الدور بس.
const n = (v: number) => (Number.isInteger(v) ? String(v) : v.toFixed(1))

export default function StaffPointsBoard({ month, group }: { month: string; group: string }) {
  const [rows, setRows] = useState<StaffPointsRow[] | null>(null)
  const [open, setOpen] = useState<string | null>(null)
  useEffect(() => {
    let alive = true
    api.getStaffPoints(month).then((r) => { if (alive) setRows(r.staff) }).catch(() => { if (alive) setRows([]) })
    return () => { alive = false }
  }, [month])
  if (!rows) return <p className="text-slate-400">ماتركس يحسب الرصيد…</p>
  const shown = rows.filter((r) => r.group === group).sort((a, b) => b.total - a.total)
  return (
    <div className="space-y-2">
      <p className="rounded-xl bg-emerald-50 p-2 text-xs leading-relaxed text-emerald-900">
        🏅 <b>رصيد الشهر بلا حد:</b> كل بند كي بي اي ما انخصم = +٥ (الثمانية = ٤٠) · كل يوم حضور بوقته +١ (متأخر +½) · وكل شغلة بدورك تضيف — كل ما تشتغل أكثر يزيد. الترتيب ويا نفس دورك بس.
      </p>
      {shown.length === 0 && <p className="p-6 text-center text-sm text-slate-400">ماكو موظفين بهالمجموعة.</p>}
      {shown.map((r, i) => (
        <div key={r.id} className="rounded-xl border border-slate-200 bg-white">
          <button type="button" onClick={() => setOpen(open === r.id ? null : r.id)} className="flex w-full items-center justify-between gap-2 px-3 py-2 text-right">
            <span className="flex items-center gap-2">
              <span className="w-6 text-xs font-bold text-slate-400">{i + 1}</span>
              <b className="text-sm">{r.name}</b>
            </span>
            <span className="flex items-center gap-3 text-xs">
              <span className="hidden text-slate-500 sm:inline">كي بي اي {n(r.kpi)} · حضور {n(r.attendance)} · شغل {n(r.work)}</span>
              <span className="rounded-full bg-emerald-600 px-3 py-0.5 text-sm font-extrabold text-white">{n(r.total)}</span>
            </span>
          </button>
          {open === r.id && (
            <div className="border-t border-slate-100 px-3 py-2 text-xs text-slate-700">
              <p>📋 <b>الكي بي اي:</b> {n(r.kpi)} نقطة{r.kpiLost.length > 0 && <span className="text-rose-600"> — انخصم من: {r.kpiLost.join('، ')}</span>}</p>
              <p>🕗 <b>الحضور:</b> {n(r.attendance)} نقطة</p>
              <p className="mt-1 font-bold">🔧 الشغل: {n(r.work)} نقطة</p>
              {r.lines.length === 0 && <p className="text-slate-400">ماكو شغل محسوب هالشهر.</p>}
              {r.lines.map((l) => <p key={l.label} className="pr-3">• {l.label}: {n(l.count)} × {n(l.each)} = <b>{n(l.points)}</b></p>)}
            </div>
          )}
        </div>
      ))}
    </div>
  )
}
