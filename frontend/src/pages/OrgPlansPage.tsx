import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { orgMonth, orgPlansOverview, type OrgPlanOverviewRow } from '../orgDesk'

// ═══ 📋 كل الخطط — (دازاي 11-08) ═══
// «المالك ومدير النظام لازم يطلعون على كل الخطط ومن حقهم يرفضوها أو يعدلون عليها».
// صف لكل إدارة ولكل وحدة بالشهر: حالتها، وإنجازها، والمتأخر. تضغط تفتح خطتها.
const STATUS: Record<OrgPlanOverviewRow['status'], { label: string; cls: string }> = {
  NONE: { label: 'ماكو خطة', cls: 'bg-slate-100 text-slate-500' },
  DRAFT: { label: 'مسودة', cls: 'bg-sky-100 text-sky-800' },
  SUBMITTED: { label: '⏳ تنتظر اعتمادك', cls: 'bg-amber-100 text-amber-800' },
  APPROVED: { label: '✅ معتمدة', cls: 'bg-emerald-100 text-emerald-800' },
  RETURNED: { label: '↩️ مرجعة للتعديل', cls: 'bg-red-100 text-red-700' },
}

export default function OrgPlansPage() {
  const [period, setPeriod] = useState(orgMonth)
  const [rows, setRows] = useState<OrgPlanOverviewRow[] | null>(null)
  const [err, setErr] = useState('')
  useEffect(() => {
    let alive = true
    orgPlansOverview(period)
      .then((r) => { if (alive) { setRows(r); setErr('') } })
      .catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذر') })
    return () => { alive = false }
  }, [period])
  const waiting = rows?.filter((r) => r.status === 'SUBMITTED').length ?? 0
  const none = rows?.filter((r) => r.status === 'NONE').length ?? 0
  const link = (r: OrgPlanOverviewRow) =>
    `/org/desk/${encodeURIComponent(r.divisionId)}?tab=${r.unitId ? `units&unit=${encodeURIComponent(r.unitId)}` : 'plan'}`

  return (
    <div dir="rtl" className="space-y-4">
      <div className="flex flex-wrap items-end justify-between gap-2">
        <div>
          <h2 className="text-2xl font-bold text-brand-900">📋 خطط الإدارات والوحدات</h2>
          <p className="text-sm text-slate-500">
            كل مدير يكتب خطة شهره (وماتركس يساعده)، وبنودها تصير مهام ويتقيّم عليها. هنا تشوفها كلها، وتضغط أي وحدة حتى تعتمد أو ترجّع أو تعدّل.
          </p>
        </div>
        <label className="text-xs text-slate-600">الشهر <input type="month" value={period} onChange={(e) => { setRows(null); setPeriod(e.target.value) }} className="ms-1 rounded-lg border border-slate-200 px-2 py-1 text-sm" /></label>
      </div>
      {rows && (
        <div className="flex flex-wrap gap-2 text-xs font-bold">
          <span className="rounded-full bg-amber-100 px-3 py-1 text-amber-800">⏳ تنتظر اعتمادك: {waiting}</span>
          <span className="rounded-full bg-slate-100 px-3 py-1 text-slate-600">ماكو خطة: {none}</span>
        </div>
      )}
      {err && <p className="text-sm text-red-600">{err}</p>}
      {!rows && !err && <p className="text-slate-400">جاري التحميل…</p>}
      {rows && (
        <div className="overflow-x-auto rounded-2xl bg-white shadow-sm">
          <table className="w-full text-sm">
            <thead><tr className="border-b border-slate-100 text-xs text-slate-500"><th className="p-3 text-right">الإدارة / الوحدة</th><th className="p-3">الحالة</th><th className="p-3">البنود</th><th className="p-3">خلصت</th><th className="p-3">متأخرة</th><th className="p-3">الدرجة</th><th className="p-3" /></tr></thead>
            <tbody>
              {rows.map((r) => (
                <tr key={`${r.divisionId}:${r.unitId ?? ''}`} className={`border-b border-slate-50 ${r.unitId ? '' : 'bg-slate-50/60'}`}>
                  <td className={`p-3 ${r.unitId ? 'ps-8 text-slate-700' : 'font-extrabold text-[#0f2040]'}`}>
                    {r.unitId ? `↳ ${r.unitName}` : `🏛️ ${r.divisionName}`}
                    {r.title && <span className="block text-[11px] font-normal text-slate-400">{r.title}</span>}
                  </td>
                  <td className="p-3 text-center"><span className={`rounded-full px-2 py-0.5 text-[11px] font-bold ${STATUS[r.status].cls}`}>{STATUS[r.status].label}</span></td>
                  <td className="p-3 text-center tabular-nums">{r.items || '—'}</td>
                  <td className="p-3 text-center tabular-nums">{r.items ? r.done : '—'}</td>
                  <td className={`p-3 text-center tabular-nums ${r.late ? 'font-bold text-red-600' : ''}`}>{r.items ? r.late : '—'}</td>
                  <td className="p-3 text-center tabular-nums">{r.score ?? '—'}</td>
                  <td className="p-3 text-center"><Link to={link(r)} className="text-xs font-bold text-brand-700 underline">افتح</Link></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
