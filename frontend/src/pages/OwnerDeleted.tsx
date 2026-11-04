import { useState } from 'react'
import { apiRequest } from '../api'
import { useApi } from '../utils/useApi'
import { card } from '../ui/classes'
import { useSession } from '../session'
import { dateTimeAr } from '../utils/format'

// ═══ 🗑️ المحذوفات (دازاي 11-04 — «ماكو حذف»، تعديلات ٦ج) ═══
// أي سجل شغل انحذف (مشروع، عرض سعر، ساعات، مركبة، فاتورة...) محفوظ كامل بسجل
// التدقيق لحظة حذفه. المالك يشوفه هنا ويرجّعه كما كان.

type Row = { id: number; at: string; entity: string; label: string; entityId: string | null; actorName: string | null; before: Record<string, unknown>; restored: boolean }
type Resp = { rows: Row[]; entities: Record<string, string> }

// أول حقل مفهوم يوصف السجل (اسم، كود، عنوان، مبلغ...)
const describe = (b: Record<string, unknown>) => {
  for (const k of ['name', 'title', 'code', 'quotationNumber', 'customerName', 'plateNumber', 'note', 'notes', 'description']) {
    const v = b?.[k]
    if (typeof v === 'string' && v.trim()) return v.length > 70 ? v.slice(0, 70) + '…' : v
  }
  for (const k of ['amount', 'seconds', 'netTotal']) if (typeof b?.[k] === 'number') return `${k}: ${(b[k] as number).toLocaleString('en-US')}`
  return '—'
}

export default function OwnerDeleted() {
  const { employee } = useSession()
  const [entity, setEntity] = useState('')
  const [reload, setReload] = useState(0)
  const [msg, setMsg] = useState<{ ok: boolean; t: string } | null>(null)
  const [open, setOpen] = useState<number | null>(null)
  const isOwner = employee?.actualRole === 'OWNER' || employee?.role === 'OWNER'
  const { data, error } = useApi(() => apiRequest<Resp>(`/owner/deleted?entity=${entity}`), [entity, reload], { enabled: isOwner })
  if (!isOwner) return <p className="p-6 text-center text-slate-500">هاي الصفحة للمالك بس.</p>
  const restore = async (r: Row) => {
    setMsg(null)
    try {
      await apiRequest(`/owner/deleted/${r.id}/restore`, { method: 'POST' })
      setMsg({ ok: true, t: `✅ رجع ${r.label}: ${describe(r.before)}` })
      setReload((n) => n + 1)
    } catch (e) { setMsg({ ok: false, t: e instanceof Error ? e.message : 'تعذر' }) }
  }
  return (
    <div dir="rtl" className="space-y-4">
      <div>
        <h1 className="text-2xl font-extrabold text-[#0f2040]">🗑️ المحذوفات</h1>
        <p className="text-sm text-slate-500">أي سجل شغل ينحذف ينحفظ كامل لحظة حذفه. تكدر ترجّعه كما كان.</p>
      </div>
      <div className={`${card} space-y-3`}>
        <div className="flex flex-wrap items-center gap-2">
          <select value={entity} onChange={(e) => setEntity(e.target.value)} className="rounded-lg border border-slate-300 px-2 py-1.5 text-sm">
            <option value="">كل الأنواع</option>
            {data && Object.entries(data.entities).sort((a, b) => a[1].localeCompare(b[1], 'ar')).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
          </select>
          {data && <span className="text-sm text-slate-500">{data.rows.length} سجل</span>}
        </div>
        {msg && <p className={`rounded-lg p-2 text-sm ${msg.ok ? 'bg-emerald-50 text-emerald-800' : 'bg-red-50 text-red-700'}`}>{msg.t}</p>}
        {error && <p className="text-sm text-red-600">{error}</p>}
        <div className="divide-y divide-slate-100">
          {data?.rows.map((r) => (
            <div key={r.id} className="py-2.5 text-sm">
              <div className="flex flex-wrap items-center gap-2">
                <span className="w-36 shrink-0 text-xs text-slate-500">{dateTimeAr(r.at)}</span>
                <span className="rounded-full bg-slate-100 px-2 py-0.5 text-[11px] font-bold text-slate-600">{r.label}</span>
                <button onClick={() => setOpen(open === r.id ? null : r.id)} className="flex-1 text-right font-bold text-[#0f2040] hover:underline">{describe(r.before)}</button>
                <span className="text-xs text-slate-400">{r.actorName ? `حذفه/آخر من لمسه: ${r.actorName}` : ''}</span>
                {r.restored
                  ? <span className="rounded-full bg-emerald-50 px-2 py-0.5 text-[11px] font-bold text-emerald-700 ring-1 ring-emerald-200">✓ راجع</span>
                  : <button onClick={() => void restore(r)} className="rounded-lg bg-[#0f2040] px-3 py-1 text-xs font-bold text-white">↩️ رجّعه</button>}
              </div>
              {open === r.id && <pre dir="ltr" className="mt-2 max-h-64 overflow-auto rounded-lg bg-slate-50 p-2 text-[11px] text-slate-600">{JSON.stringify(r.before, null, 2)}</pre>}
            </div>
          ))}
          {data && data.rows.length === 0 && <p className="py-6 text-center text-slate-400">ماكو محذوفات.</p>}
        </div>
      </div>
    </div>
  )
}
