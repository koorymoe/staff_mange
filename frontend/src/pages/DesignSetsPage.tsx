import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { apiRequest } from '../api'
import PctWhy from '../ui/PctWhy'
import type { DesignMeta, DesignSetRow } from './design/types'

// ═══ 🏗️ صيت التصاميم — (ع) 11-08 ═══
// «كل مشروع إله مجموعة تصاميم — من البناء والهيكل لآخر شي الطاقة الشمسية — كلشي بمكان واحد».
export default function DesignSetsPage() {
  const nav = useNavigate()
  const [meta, setMeta] = useState<DesignMeta | null>(null)
  const [rows, setRows] = useState<DesignSetRow[] | null>(null)
  const [err, setErr] = useState('')
  const [show, setShow] = useState(false)
  const [archived, setArchived] = useState(false)
  const [form, setForm] = useState({ title: '', projectId: '', notes: '', disciplines: [] as string[] })
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    let alive = true
    Promise.all([apiRequest<DesignMeta>('/design-sets/meta'), apiRequest<DesignSetRow[]>('/design-sets')])
      .then(([m, r]) => { if (alive) { setMeta(m); setRows(r); setForm((f) => ({ ...f, disciplines: m.disciplines.map((d) => d.code) })) } })
      .catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذر') })
    return () => { alive = false }
  }, [])

  const create = async () => {
    setBusy(true); setErr('')
    try {
      const r = await apiRequest<{ id: string }>('/design-sets', { method: 'POST', body: JSON.stringify(form) })
      nav(`/design-sets/${r.id}`)
    } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') } finally { setBusy(false) }
  }
  const toggle = (c: string) => setForm((f) => ({ ...f, disciplines: f.disciplines.includes(c) ? f.disciplines.filter((x) => x !== c) : [...f.disciplines, c] }))
  const shown = (rows ?? []).filter((r) => !!r.archivedAt === archived)

  return (
    <div dir="rtl" className="space-y-4">
      <div className="flex flex-wrap items-end justify-between gap-2">
        <div>
          <h2 className="text-2xl font-bold text-brand-900">🏗️ صيت التصاميم</h2>
          <p className="text-sm text-slate-500">كل مشروع وتصاميمه بمكان واحد: معماري، إنشائي، ثري دي، ديكور، كاميرات، شبكات، صوتيات، إنذار حريق، طاقة شمسية، مقاولات… لكل تخصص مصمم وموعد ومراجعة.</p>
        </div>
        {meta?.canManage && <button type="button" onClick={() => setShow(!show)} className="rounded-xl bg-[#0f2040] px-4 py-2 text-sm font-bold text-white">+ صيت جديد</button>}
      </div>
      {err && <p className="text-sm text-red-600">{err}</p>}
      {show && meta && (
        <div className="space-y-3 rounded-2xl border border-slate-200 bg-white p-4">
          <input value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} placeholder="اسم الصيت (مثلاً: فيلا أبو علي — الجادرية)" className="w-full rounded-lg border border-slate-200 p-2 text-sm" />
          <select value={form.projectId} onChange={(e) => setForm({ ...form, projectId: e.target.value })} className="w-full rounded-lg border border-slate-200 p-2 text-sm">
            <option value="">مربوط بمشروع؟ (اختياري)</option>
            {meta.projects.map((p) => <option key={p.id} value={p.id}>{p.code} — {p.name}</option>)}
          </select>
          <textarea value={form.notes} onChange={(e) => setForm({ ...form, notes: e.target.value })} rows={2} placeholder="ملاحظات (المساحة، طلبات الزبون…)" className="w-full rounded-lg border border-slate-200 p-2 text-sm" />
          <div>
            <p className="mb-1 text-xs font-bold text-slate-600">التخصصات المطلوبة ({form.disciplines.length})</p>
            <div className="flex flex-wrap gap-1.5">
              {meta.disciplines.map((d) => (
                <button key={d.code} type="button" onClick={() => toggle(d.code)}
                  className={`rounded-full px-3 py-1 text-xs font-bold ${form.disciplines.includes(d.code) ? 'bg-brand-700 text-white' : 'bg-slate-100 text-slate-500'}`}>{d.icon} {d.name}</button>
              ))}
            </div>
          </div>
          <button type="button" disabled={busy || form.title.trim().length < 3 || !form.disciplines.length} onClick={() => void create()} className="rounded-lg bg-emerald-700 px-4 py-2 text-sm font-bold text-white disabled:opacity-50">أنشئ الصيت</button>
        </div>
      )}
      <div className="flex gap-2 text-xs">
        <button type="button" onClick={() => setArchived(false)} className={`rounded-full px-3 py-1 font-bold ${!archived ? 'bg-brand-700 text-white' : 'bg-slate-100 text-slate-600'}`}>الشغّالة</button>
        <button type="button" onClick={() => setArchived(true)} className={`rounded-full px-3 py-1 font-bold ${archived ? 'bg-brand-700 text-white' : 'bg-slate-100 text-slate-600'}`}>المؤرشفة</button>
      </div>
      {!rows && !err && <p className="text-slate-400">جاري التحميل…</p>}
      {rows && shown.length === 0 && <p className="rounded-xl bg-white p-8 text-center text-slate-400">ماكو صيتات {archived ? 'مؤرشفة' : 'بعد'}.</p>}
      <div className="grid gap-3 md:grid-cols-2">
        {shown.map((r) => {
          const pct = r.sections ? Math.round((100 * r.approved) / r.sections) : 0
          return (
            <Link key={r.id} to={`/design-sets/${r.id}`} className="block rounded-2xl border border-slate-200 bg-white p-4 transition hover:-translate-y-0.5 hover:shadow-md">
              <div className="flex items-center justify-between gap-2">
                <p className="font-extrabold text-[#0f2040]">{r.title}</p>
                <PctWhy title={`اعتماد صيت «${r.title}»`} suffix="٪" className="text-sm font-black tabular-nums text-brand-700"
                  parts={{ numLabel: 'قسم معتمد', num: r.approved, denLabel: 'قسم بالصيت', den: r.sections }}
                  meaning={<>شكد من أقسام هالصيت (بناء، هيكل… لحد الطاقة) صار تصميمه معتمد. ١٠٠٪ يعني كل الأقسام اعتمدت؛ القسم الي بعده بالمراجعة{r.inReview > 0 ? ` (هسه ${r.inReview})` : ''} أو ما انرفع ما ينحسب لحد ما يعتمد.</>}
                  source={<>أقسام صيت التصاميم المسجلة لهالمشروع بالنظام، والمعتمد منها هو الي حالته «معتمد».</>} />
              </div>
              <p className="text-xs text-slate-500">{r.projectCode ? `مشروع ${r.projectCode} · ` : ''}{new Date(r.createdAt).toLocaleDateString('en-GB')}</p>
              <div className="mt-2 h-2 overflow-hidden rounded-full bg-slate-100"><div className="h-full bg-emerald-500" style={{ width: `${pct}%` }} /></div>
              <p className="mt-2 flex flex-wrap gap-2 text-[11px]">
                <span className="text-slate-600">{r.approved} من {r.sections} معتمد</span>
                {r.inReview > 0 && <span className="font-bold text-amber-700">⏳ {r.inReview} ينتظر المراجعة</span>}
                {r.late > 0 && <span className="font-bold text-red-600">⚠️ {r.late} متأخر</span>}
              </p>
            </Link>
          )
        })}
      </div>
    </div>
  )
}
