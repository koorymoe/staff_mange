import { useEffect, useState } from 'react'
import { api, type ProjectMoney } from '../api'
import MatrixNote from '../components/MatrixNote'
import ProjectPaymentsModal from '../components/ProjectPaymentsModal'

// ═══ 💰 فلوس المشاريع — قرار (ع) 10-06 ═══
// كل مشروع: قيمة العقد، المدفوع، الباقي. للمحاسب وإدارة المشاريع والمراقب.

const fmt = (v: number) => v.toLocaleString('en-US')

export default function ProjectPaymentsPage() {
  const [rows, setRows] = useState<ProjectMoney[] | null>(null)
  const [err, setErr] = useState('')
  const [open, setOpen] = useState<ProjectMoney | null>(null)
  const [only, setOnly] = useState<'all' | 'missing'>('missing')
  const load = () => { void api.getProjectMoneyOverview().then(setRows).catch((e) => setErr(e instanceof Error ? e.message : 'تعذر')) }
  useEffect(load, [])
  if (err) return <p className="text-red-600">{err}</p>
  if (!rows) return <p className="text-slate-400">…</p>
  const live = rows.filter((r) => !r.stage.includes('مرفوض'))
  const missing = live.filter((r) => (r.stage.includes('تنفيذ') || r.stage.includes('مكتمل')) && r.payments === 0)
  const shown = only === 'missing' ? missing : live
  const paid = live.reduce((a, r) => a + r.paid, 0)
  const value = live.reduce((a, r) => a + (r.contractValue ?? 0), 0)
  return (
    <div dir="rtl" className="space-y-4">
      <div>
        <h2 className="text-2xl font-bold text-brand-900">💰 فلوس المشاريع</h2>
        <p className="text-sm text-slate-500">كل دفعة تنسجّل هنا تدخل بالإيرادات. المحاسب يأكدها.</p>
      </div>
      <MatrixNote>{`المسجّل ${fmt(paid)} د.ع من قيم عقود ${fmt(value)} د.ع.${missing.length ? ` ⚠️ ${missing.length} مشروع بالتنفيذ أو مكتمل وماكو عليه ولا دفعة.` : ''}`}</MatrixNote>
      <div className="flex gap-2 text-sm">
        <button type="button" onClick={() => setOnly('missing')} className={`rounded-full px-3 py-1 ${only === 'missing' ? 'bg-red-600 text-white' : 'bg-slate-100'}`}>بلا دفعات ({missing.length})</button>
        <button type="button" onClick={() => setOnly('all')} className={`rounded-full px-3 py-1 ${only === 'all' ? 'bg-[#0f2040] text-white' : 'bg-slate-100'}`}>كل المشاريع ({live.length})</button>
      </div>
      <div className="space-y-1.5">
        {shown.map((r) => {
          const pct = r.contractValue ? Math.min(100, Math.round((r.paid / r.contractValue) * 100)) : null
          return (
            <button key={r.projectId} type="button" onClick={() => setOpen(r)} className="w-full rounded-xl border border-slate-200 bg-white p-3 text-right hover:shadow">
              <div className="flex flex-wrap items-center justify-between gap-2 text-sm">
                <span><b>{r.name}</b> <span className="text-xs text-slate-500">{r.code} · {r.stage}{r.ownerName && ` · ${r.ownerName}`}</span></span>
                <span className="text-xs"><b className="text-emerald-700">{fmt(r.paid)}</b> / {r.contractValue != null ? fmt(r.contractValue) : '—'} د.ع{r.unverified > 0 && <span className="text-amber-600"> · {fmt(r.unverified)} تنتظر تأكيد</span>}</span>
              </div>
              {pct != null && <div className="mt-1.5 h-1.5 overflow-hidden rounded-full bg-slate-100"><div className="h-full bg-emerald-500" style={{ width: `${pct}%` }} /></div>}
            </button>
          )
        })}
        {shown.length === 0 && <p className="rounded-xl bg-emerald-50 p-3 text-sm text-emerald-800">✅ ماكو شي هنا.</p>}
      </div>
      {open && <ProjectPaymentsModal projectId={open.projectId} title={open.name} onClose={() => { setOpen(null); load() }} />}
    </div>
  )
}
