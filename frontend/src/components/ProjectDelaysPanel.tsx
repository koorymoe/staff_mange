import { useEffect, useState } from 'react'
import { api, type ProjectDelay } from '../api'

// ═══ ⏱️ تأخير المشاريع — مكتب المراقب (قرار (ع) 10-10) ═══
// «ليش المشرف أو التقني أو المهندس جاي يتأخر؟ وهل سوّى شي اليوم؟».
// المتجاوز بلا سبب أول، بعده المتجاوز بسبب، بعده الماشي.
export default function ProjectDelaysPanel() {
  const [rows, setRows] = useState<ProjectDelay[] | null>(null)
  const [err, setErr] = useState('')
  const [only, setOnly] = useState<'late' | 'all'>('late')
  useEffect(() => {
    let alive = true
    api.getProjectDelays().then((r) => { if (alive) setRows(r) }).catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذر') })
    return () => { alive = false }
  }, [])
  if (err) return <p className="text-sm text-red-600">{err}</p>
  if (!rows) return <p className="text-sm text-slate-400">ماتركس يحلل المشاريع…</p>
  const rank = (p: ProjectDelay) => (p.overLimit ? (p.delayReason ? 1 : 0) : 2)
  const sorted = [...rows].sort((a, b) => rank(a) - rank(b) || (b.daysInStage - b.stageLimit) - (a.daysInStage - a.stageLimit))
  const shown = only === 'late' ? sorted.filter((p) => p.overLimit) : sorted
  const late = rows.filter((p) => p.overLimit)
  const owners = new Map<string, { name: string; worked: boolean }>()
  for (const p of rows) {
    if (!p.ownerId) continue
    const o = owners.get(p.ownerId) ?? { name: p.ownerName, worked: false }
    o.worked ||= p.workedToday
    owners.set(p.ownerId, o)
  }
  const idle = [...owners.values()].filter((o) => !o.worked)
  return (
    <div dir="rtl" className="space-y-3">
      <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
        <Tile label="مشاريع شغّالة" v={rows.length} />
        <Tile label="متجاوزة الحد" v={late.length} bad={late.length > 0} />
        <Tile label="متجاوزة بلا سبب" v={late.filter((p) => !p.delayReason).length} bad={late.some((p) => !p.delayReason)} />
        <Tile label="مسؤولين ما اشتغلوا اليوم" v={idle.length} bad={idle.length > 0} />
      </div>
      {idle.length > 0 && <p className="rounded-xl bg-amber-50 px-3 py-2 text-xs text-amber-900">⛔ ما اشتغلوا اليوم على مشاريعهم: <b>{idle.map((o) => o.name).join('، ')}</b></p>}
      <div className="flex gap-2">
        {([['late', 'المتجاوزة'], ['all', 'كل المشاريع']] as const).map(([k, l]) => (
          <button key={k} type="button" onClick={() => setOnly(k)}
            className={`rounded-lg px-3 py-1.5 text-xs font-bold ${only === k ? 'bg-[#0f2040] text-white' : 'bg-white text-slate-600 ring-1 ring-slate-200'}`}>{l}</button>
        ))}
      </div>
      <div className="overflow-x-auto rounded-2xl border border-slate-200 bg-white">
        <table className="w-full min-w-[760px] text-right text-sm">
          <thead className="bg-slate-50 text-xs text-slate-500">
            <tr><th className="p-2">المشروع</th><th className="p-2">المسؤول</th><th className="p-2">المرحلة</th><th className="p-2">الأيام / الحد</th><th className="p-2">اليوم</th><th className="p-2">سبب التأخير</th></tr>
          </thead>
          <tbody className="divide-y divide-slate-100">
            {shown.length === 0 && <tr><td colSpan={6} className="p-6 text-center text-slate-400">ماكو مشاريع متجاوزة 👌</td></tr>}
            {shown.map((p) => (
              <tr key={p.id} className={p.overLimit && !p.delayReason ? 'bg-red-50/50' : ''}>
                <td className="p-2"><b>{p.name}</b> <span className="text-xs text-slate-400">{p.code}</span></td>
                <td className="p-2 text-xs">{p.ownerName || <span className="text-red-600">بلا مسؤول</span>}</td>
                <td className="p-2 text-xs">{p.stage}</td>
                <td className={`p-2 text-xs font-bold ${p.overLimit ? 'text-red-700' : 'text-slate-600'}`}>
                  {p.daysInStage} / {p.stageLimit}{p.overLimit && ` (+${p.daysInStage - p.stageLimit})`}
                </td>
                <td className="p-2 text-xs">
                  {p.workedToday ? <span className="text-emerald-700">✅ اشتغل</span> : <span className="text-red-600">⛔ ما اشتغل</span>}
                  {p.lastActivityAt && <span className="block text-[10px] text-slate-400">آخر حركة {new Date(p.lastActivityAt).toLocaleDateString('ar-IQ')}{p.lastActivityBy ? ` — ${p.lastActivityBy}` : ''}</span>}
                </td>
                <td className="p-2 text-xs">
                  {p.delayReason ? <>{p.delayReason}<span className="block text-[10px] text-slate-400">{p.delayReasonBy} · {p.delayReasonAt ? new Date(p.delayReasonAt).toLocaleDateString('ar-IQ') : ''}</span></>
                    : p.overLimit ? <span className="font-bold text-red-600">ما كتب سبب</span> : '—'}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

function Tile({ label, v, bad }: { label: string; v: number; bad?: boolean }) {
  return (
    <div className={`rounded-xl border p-3 ${bad ? 'border-red-200 bg-red-50' : 'border-slate-200 bg-white'}`}>
      <p className="text-[11px] text-slate-500">{label}</p>
      <b className={`text-2xl ${bad ? 'text-red-700' : 'text-slate-800'}`}>{v}</b>
    </div>
  )
}
