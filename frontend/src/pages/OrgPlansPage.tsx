import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { orgMonth, orgPlansOverview, type OrgPlanOverviewRow } from '../orgDesk'
import { LineIcon } from '../components/LineIcon'

// ═══ 📋 كل الخطط — (دازاي 11-08) ═══
// «المالك ومدير النظام لازم يطلعون على كل الخطط ومن حقهم يرفضوها أو يعدلون عليها».
// صف لكل إدارة ولكل وحدة بالشهر: حالتها، وإنجازها، والمتأخر. تضغط تفتح خطتها.
// (ك) 11-23 زينا: التصميم مثل الصورة — بطاقات بالأعداد، بحث وتصفية، جدول بترقيم صفحات. الألوان بـindex.css (.op-*)
const STATUS: Record<OrgPlanOverviewRow['status'], { label: string; cls: string }> = {
  NONE: { label: 'ماكو خطة', cls: 'op-pill-none' },
  DRAFT: { label: 'مسودة', cls: 'op-pill-draft' },
  SUBMITTED: { label: 'بانتظار الاعتماد', cls: 'op-pill-wait' },
  APPROVED: { label: 'معتمدة', cls: 'op-pill-ok' },
  RETURNED: { label: 'مرفوضة', cls: 'op-pill-ret' },
}
const monthLabel = (p: string) => {
  const [y, m] = p.split('-').map(Number)
  return new Date(y, (m || 1) - 1, 1).toLocaleDateString('en-US', { month: 'long', year: 'numeric' })
}

export default function OrgPlansPage() {
  const [period, setPeriod] = useState(orgMonth)
  const [rows, setRows] = useState<OrgPlanOverviewRow[] | null>(null)
  const [err, setErr] = useState('')
  const [q, setQ] = useState('')
  const [st, setSt] = useState('')
  const [per, setPer] = useState(10)
  const [page, setPage] = useState(1)
  useEffect(() => {
    let alive = true
    orgPlansOverview(period)
      .then((r) => { if (alive) { setRows(r); setErr('') } })
      .catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذر') })
    return () => { alive = false }
  }, [period])
  const link = (r: OrgPlanOverviewRow) =>
    `/org/desk/${encodeURIComponent(r.divisionId)}?tab=${r.unitId ? `units&unit=${encodeURIComponent(r.unitId)}` : 'plan'}`
  const count = (s: OrgPlanOverviewRow['status']) => rows?.filter((r) => r.status === s).length ?? 0
  const all = rows ?? []
  const filtered = all.filter((r) => (!q.trim() || `${r.divisionName} ${r.unitName ?? ''} ${r.title ?? ''}`.includes(q.trim())) && (!st || r.status === st))
  const pages = Math.max(1, Math.ceil(filtered.length / per))
  const pg = Math.min(page, pages)
  const shown = filtered.slice((pg - 1) * per, pg * per)
  // أول إدارة بلا خطة — «إضافة خطة» يفتحها حتى المدير يكتب خطتها
  const firstNone = all.find((r) => r.status === 'NONE')
  const cards: [string, number | string, string, string][] = [
    ['مسودة', count('DRAFT'), 'file', 'op-c-blue'],
    ['مرفوضة', count('RETURNED'), 'xCircle', 'op-c-red'],
    ['بانتظار الاعتماد', count('SUBMITTED'), 'clock', 'op-c-amber'],
    ['معتمدة', count('APPROVED'), 'check', 'op-c-green'],
    ['إجمالي الخطط', all.length, 'list', 'op-c-list'],
  ]
  const heads: [string, string][] = [['', '#'], ['bank', 'الإدارة / الوحدة'], ['file', 'الخطة'], ['layers', 'الحالة'], ['list', 'البنود'], ['checkCircle', 'خلصت'], ['clock', 'متأخرة'], ['award', 'الدرجة'], ['', 'الإجراءات']]
  const dash = <span className="opacity-70">—</span>

  return (
    <div dir="rtl" className="space-y-5">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h2 className="text-3xl font-extrabold text-brand-900">خطط الإدارات والوحدات</h2>
          <p className="mt-1 text-[15px] text-slate-500">
            كل مدير يكتب خطة شهره (وماتركس يساعده)، وبنودها تصير مهام ويتقيّم عليها. هنا تشوفها كلها، وتضغط أي وحدة حتى تعتمد أو ترجّع أو تعدّل.
          </p>
        </div>
        <label className="flex items-center gap-3 text-[15px] text-slate-600">الشهر
          <span className="op-field relative flex items-center gap-3 rounded-xl border px-4 py-2">
            <LineIcon n="cal" className="h-5 w-5" /><span className="min-w-[8.5rem] text-center font-bold" dir="ltr">{monthLabel(period)}</span>
            <input type="month" value={period} onChange={(e) => { if (e.target.value) { setRows(null); setPage(1); setPeriod(e.target.value) } }} className="absolute inset-0 cursor-pointer opacity-0" aria-label="الشهر" />
          </span>
        </label>
      </div>

      {rows && (
        <div className="grid grid-cols-2 gap-4 md:grid-cols-3 xl:grid-cols-5">
          {cards.map(([label, n, ic, c]) => (
            <div key={label} className="op-card flex items-center gap-4 rounded-2xl border px-5 py-4">
              <span className={`op-ic ${c} flex h-12 w-12 shrink-0 items-center justify-center rounded-xl`}><LineIcon n={ic} className="h-7 w-7" /></span>
              <div className="min-w-0 flex-1 text-left">
                <p className="text-right text-base font-bold">{label}</p>
                <p className={`op-n ${c} text-3xl font-extrabold tabular-nums`}>{n}</p>
              </div>
            </div>
          ))}
        </div>
      )}
      {err && <p className="text-sm text-red-600">{err}</p>}
      {!rows && !err && <p className="text-slate-400">جاري التحميل…</p>}
      {rows && (
        <div className="op-panel space-y-4 rounded-3xl border p-5">
          <div className="flex flex-wrap items-center gap-3">
            <label className="op-field flex min-w-[260px] max-w-md flex-1 items-center gap-2 rounded-xl border px-4 py-2.5">
              <LineIcon n="search" className="h-5 w-5 shrink-0 opacity-80" />
              <input value={q} onChange={(e) => { setQ(e.target.value); setPage(1) }} placeholder="ابحث عن إدارة أو وحدة ..." className="w-full bg-transparent text-[15px] outline-none" />
            </label>
            <label className="op-field ms-auto flex items-center gap-2 rounded-xl border px-4 py-2.5 text-[15px] font-bold">
              <LineIcon n="filter" className="h-5 w-5" />
              <select value={st} onChange={(e) => { setSt(e.target.value); setPage(1) }} className="bg-transparent outline-none" aria-label="تصفية">
                <option value="">تصفية</option>
                {Object.entries(STATUS).map(([k, v]) => <option key={k} value={k}>{v.label}</option>)}
              </select>
            </label>
            {firstNone
              ? <Link to={link(firstNone)} title={`أول إدارة بلا خطة: ${firstNone.unitName ?? firstNone.divisionName}`} className="op-add flex items-center gap-2 rounded-xl px-5 py-2.5 text-[15px] font-bold"><LineIcon n="plus" className="h-5 w-5" />إضافة خطة</Link>
              : <span className="op-add flex cursor-default items-center gap-2 rounded-xl px-5 py-2.5 text-[15px] font-bold opacity-50" title="كل الإدارات عدها خطة بهالشهر"><LineIcon n="plus" className="h-5 w-5" />إضافة خطة</span>}
          </div>

          <div className="op-table overflow-x-auto rounded-2xl border" data-testid="plans-table">
            <table className="w-full min-w-[1000px] table-fixed text-[15px]">
              <colgroup>{['5%', '19%', '16%', '11%', '9%', '9%', '9%', '9%', '13%'].map((w, i) => <col key={i} style={{ width: w }} />)}</colgroup>
              <thead className="font-bold">
                <tr>
                  {heads.map(([ic, h]) => (
                    <th key={h} className="p-0 text-center">
                      <span className="op-th flex h-12 items-center justify-center gap-2 px-1.5">{ic && <LineIcon n={ic} className="h-5 w-5" />}<span>{h}</span></span>
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {shown.map((r, i) => {
                  const s = STATUS[r.status]
                  return (
                    <tr key={`${r.divisionId}:${r.unitId ?? ''}`} className="op-row h-12 align-middle">
                      <td className="px-2 text-center font-bold">{(pg - 1) * per + i + 1}</td>
                      <td className="px-3 py-1.5">
                        <div className={`flex items-center gap-2 font-extrabold ${r.unitId ? 'ps-5' : ''}`}><LineIcon n="bank" className="h-5 w-5 shrink-0 opacity-70" />{r.unitId ? `↳ ${r.unitName}` : r.divisionName}</div>
                        {r.title && <div className="text-xs opacity-70">{r.title}</div>}
                      </td>
                      <td className="px-2 text-center">{r.title ? <span className="inline-flex items-center gap-1.5 text-sm"><LineIcon n="file" className="h-4 w-4 shrink-0" />{r.title}</span> : dash}</td>
                      <td className="px-2 text-center"><span className={`op-pill ${s.cls} inline-flex min-w-[6.5rem] items-center justify-center gap-1.5 rounded-full border px-3 py-0.5 text-sm font-bold`}>{r.status !== 'NONE' && <span className="op-dot h-2 w-2 rounded-full" />}{s.label}</span></td>
                      <td className="px-2 text-center tabular-nums">{r.items || dash}</td>
                      <td className="px-2 text-center tabular-nums">{r.items ? r.done : dash}</td>
                      <td className={`px-2 text-center tabular-nums ${r.late ? 'font-bold text-red-500' : ''}`}>{r.items ? r.late : dash}</td>
                      <td className="px-2 text-center tabular-nums">{r.score ?? dash}</td>
                      <td className="px-2 text-center"><Link to={link(r)} className="op-open inline-flex items-center gap-1.5 rounded-lg border px-5 py-1 text-sm font-bold"><LineIcon n="ext" className="h-4 w-4" />افتح</Link></td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
            {filtered.length === 0 && <p className="py-6 text-center text-sm opacity-70">ماكو إدارة أو وحدة تطابق البحث.</p>}
          </div>

          <div className="op-foot flex flex-wrap items-center justify-between gap-3 rounded-2xl border px-4 py-3">
            <label className="flex items-center gap-3 text-[15px]">عرض
              <select value={per} onChange={(e) => { setPer(Number(e.target.value)); setPage(1) }} className="op-field rounded-xl border px-4 py-1.5 font-bold" aria-label="لكل صفحة">
                {[10, 25, 50].map((n) => <option key={n} value={n}>{n}</option>)}
              </select>لكل صفحة</label>
            <div className="flex items-center gap-2">
              <button type="button" disabled={pg <= 1} onClick={() => setPage(pg - 1)} className="op-pg flex h-11 w-12 items-center justify-center rounded-xl border disabled:opacity-40" aria-label="السابقة"><LineIcon n="right" className="h-5 w-5" /></button>
              {Array.from({ length: pages }, (_, k) => k + 1).map((n) => (
                <button key={n} type="button" onClick={() => setPage(n)} className={`op-pg flex h-11 min-w-[3rem] items-center justify-center rounded-xl border px-3 font-bold ${n === pg ? 'op-pg-on' : ''}`}>{n}</button>
              ))}
              <button type="button" disabled={pg >= pages} onClick={() => setPage(pg + 1)} className="op-pg flex h-11 w-12 items-center justify-center rounded-xl border disabled:opacity-40" aria-label="التالية"><LineIcon n="left" className="h-5 w-5" /></button>
            </div>
            <span className="op-field flex items-center gap-3 rounded-xl border px-5 py-2 text-[15px] font-bold"><LineIcon n="list" className="h-5 w-5" />إجمالي الخطط: <b className="tabular-nums">{filtered.length}</b></span>
          </div>
        </div>
      )}
    </div>
  )
}
