import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { apiRequest, fileUrl } from '../api'
import type { PlanMeta, PlanSetRow } from './design/planTypes'

// ═══ 📐 سيت المخططات — (ك) 11-10 بطلب دازاي ═══
// «إدارة وتجميع جميع المخططات في سيت واحد احترافي»: كل مشروع سيت، والمخططات بأنواعها
// (الأقسام العشرة)، وتنزيل PDF بهوية الشركة ويّا عرض السعر.
export default function PlanSetsPage() {
  const nav = useNavigate()
  const [meta, setMeta] = useState<PlanMeta | null>(null)
  const [rows, setRows] = useState<PlanSetRow[] | null>(null)
  const [err, setErr] = useState('')
  const [show, setShow] = useState(false)
  const [archived, setArchived] = useState(false)
  const [form, setForm] = useState({ title: '', projectId: '', notes: '' })
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    let alive = true
    Promise.all([apiRequest<PlanMeta>('/plan-sets/meta'), apiRequest<PlanSetRow[]>('/plan-sets')])
      .then(([m, r]) => { if (alive) { setMeta(m); setRows(r) } })
      .catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذر') })
    return () => { alive = false }
  }, [])

  const create = async () => {
    setBusy(true); setErr('')
    try {
      const r = await apiRequest<{ id: string }>('/plan-sets', { method: 'POST', body: JSON.stringify(form) })
      nav(`/plan-sets/${r.id}`)
    } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') } finally { setBusy(false) }
  }
  const shown = (rows ?? []).filter((r) => !!r.archivedAt === archived)
  const stats = meta?.stats
  const cards: [string, number | undefined, string][] = [
    ['أنواع المخططات', stats?.types, '🧊'],
    ['المشاريع', stats?.projects, '📁'],
    ['إجمالي السيتات', stats?.sets, '🗂️'],
    ['إجمالي التصاميم', stats?.files, '📄'],
  ]

  return (
    <div dir="rtl" className="space-y-4">
      <div className="relative overflow-hidden rounded-2xl bg-[#0f2040] px-6 py-6 text-white"
        style={{ backgroundImage: 'linear-gradient(rgba(255,255,255,.06) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,.06) 1px, transparent 1px)', backgroundSize: '22px 22px' }}>
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 className="text-2xl font-extrabold">📐 سيت المخططات</h2>
            <p className="text-sm text-white/70">إدارة وتجميع جميع المخططات في سيت واحد احترافي</p>
          </div>
          {meta?.canManage && <button type="button" onClick={() => setShow(!show)} className="rounded-xl bg-amber-500 px-4 py-2 text-sm font-bold text-white">+ سيت جديد</button>}
        </div>
      </div>

      <div className="grid grid-cols-2 gap-3 md:grid-cols-4">
        {cards.map(([label, v, icon]) => (
          <div key={label} className="flex items-center justify-between rounded-2xl border border-slate-200 bg-white p-4">
            <div><p className="text-2xl font-extrabold text-brand-900">{v ?? '—'}</p><p className="text-xs text-slate-500">{label}</p></div>
            <span className="text-2xl">{icon}</span>
          </div>
        ))}
      </div>

      {err && <p className="text-sm text-red-600">{err}</p>}
      {show && meta && (
        <div className="space-y-3 rounded-2xl border border-slate-200 bg-white p-4">
          <input value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} placeholder="اسم السيت (مثلاً: سيت مشروع الجامعة)" className="w-full rounded-lg border border-slate-200 p-2 text-sm" />
          <select value={form.projectId} onChange={(e) => setForm({ ...form, projectId: e.target.value })} className="w-full rounded-lg border border-slate-200 p-2 text-sm">
            <option value="">مربوط بمشروع؟ (اختياري — منه ينجاب عرض السعر)</option>
            {meta.projects.map((p) => <option key={p.id} value={p.id}>{p.code} — {p.name}</option>)}
          </select>
          <textarea value={form.notes} onChange={(e) => setForm({ ...form, notes: e.target.value })} placeholder="ملاحظات (اختياري)" rows={2} className="w-full rounded-lg border border-slate-200 p-2 text-sm" />
          <button type="button" disabled={busy || form.title.trim().length < 3} onClick={() => void create()} className="rounded-lg bg-emerald-700 px-4 py-2 text-sm font-bold text-white disabled:opacity-50">أنشئ السيت</button>
        </div>
      )}

      <div className="flex items-center justify-between">
        <h3 className="text-lg font-bold text-brand-900">جميع سيتات التصميم</h3>
        <div className="flex gap-2">
          {[false, true].map((a) => (
            <button key={String(a)} type="button" onClick={() => setArchived(a)} className={`rounded-full px-3 py-1 text-xs font-bold ${archived === a ? 'bg-brand-700 text-white' : 'bg-slate-100 text-slate-500'}`}>{a ? 'المؤرشفة' : 'الشغّالة'}</button>
          ))}
        </div>
      </div>
      {rows && shown.length === 0 && <p className="rounded-2xl border border-slate-200 bg-white p-6 text-center text-sm text-slate-500">ماكو سيتات بعد.</p>}
      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
        {shown.map((r) => (
          <div key={r.id} className="overflow-hidden rounded-2xl border border-slate-200 bg-white">
            <div className="relative h-36 bg-slate-100">
              {r.cover ? <img src={fileUrl(r.cover)} alt="" className="h-full w-full object-cover" /> : <div className="grid h-full place-items-center text-4xl">📐</div>}
              <span className="absolute end-2 top-2 rounded-full bg-[#0f2040] px-2 py-0.5 text-xs font-bold text-white">{r.files} تصاميم</span>
            </div>
            <div className="space-y-1 p-3">
              <p className="font-bold text-slate-800">{r.title}</p>
              <p className="text-xs text-slate-500">{r.projectCode ? `${r.projectCode} · ` : ''}{new Date(r.createdAt).toLocaleDateString('en-CA', { timeZone: 'Asia/Baghdad' })}</p>
              <div className="flex gap-3 pt-1 text-xs font-bold">
                <Link to={`/plan-sets/${r.id}`} className="text-brand-700">👁️ معاينة</Link>
                <Link to={`/plan-sets/${r.id}?pdf=1`} className="text-amber-600">⬇️ تحميل</Link>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
