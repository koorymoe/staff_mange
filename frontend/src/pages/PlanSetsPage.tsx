import { useEffect, useMemo, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { apiRequest, ensureFileToken, thumbUrl } from '../api'
import { ArtCrop, IconCal, IconFolder, IconLayers, IconDoc, IconUpload } from './design/planArt'
import type { PlanMeta, PlanSetRow } from './design/planTypes'

// ═══ 📐 سيت المخططات — (ل) 11-10 بتصميم دازاي (نهاري: الصورة الثانية، ليلي: الثالثة) ═══
// بانر، أربع كروت أرقام حقيقية، «جميع سيتات التصميم» (سيتات المشاريع / نماذج جاهزة)،
// ومربع «رفع مخططات جديدة». الرفع نفسه يصير داخل السيت: قسم ← مخطط فرعي ← تصميمك.
// الأصناف العادية (bg-white/text-slate) تنقلب لحالها بالوضع الليلي (index.css).

const day = (s: string) => new Date(s).toLocaleDateString('en-CA', { timeZone: 'Asia/Baghdad' }).replace(/-/g, '/')
const isImg = (k: string | null) => !!k && /\.(png|jpe?g|webp|gif)(\?|$)/i.test(k)

function Stat({ n, label, icon }: { n: number | string; label: string; icon: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
      <div><div className="text-3xl font-extrabold tabular-nums text-slate-900">{n}</div><div className="text-sm text-slate-500">{label}</div></div>
      <div className="grid h-14 w-14 place-items-center rounded-2xl bg-indigo-50 text-indigo-600">{icon}</div>
    </div>
  )
}

export default function PlanSetsPage() {
  const nav = useNavigate()
  const [meta, setMeta] = useState<PlanMeta | null>(null)
  const [rows, setRows] = useState<PlanSetRow[] | null>(null)
  const [err, setErr] = useState('')
  const [tab, setTab] = useState<'projects' | 'templates'>('projects')
  const [showAll, setShowAll] = useState(false)
  const [create, setCreate] = useState(false)
  const [form, setForm] = useState({ title: '', projectId: '', notes: '', isTemplate: false })
  const [busy, setBusy] = useState(false)
  const [pickFor, setPickFor] = useState(false) // مربع الرفع: يختار سيت أول
  const [pickSet, setPickSet] = useState('')

  useEffect(() => {
    let alive = true
    Promise.all([apiRequest<PlanMeta>('/plan-sets/meta'), apiRequest<PlanSetRow[]>('/plan-sets'), ensureFileToken()])
      .then(([m, r]) => { if (alive) { setMeta(m); setRows(r) } })
      .catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذر') })
    return () => { alive = false }
  }, [])

  const live = useMemo(() => (rows ?? []).filter((r) => !r.archivedAt), [rows])
  const shown = useMemo(() => live.filter((r) => (tab === 'templates') === !!r.isTemplate), [live, tab])
  const typeCount = meta?.groups.reduce((n, g) => n + g.types.length, 0) ?? 0

  const submit = async () => {
    setBusy(true); setErr('')
    try {
      const r = await apiRequest<{ id: string }>('/plan-sets', { method: 'POST', body: JSON.stringify(form) })
      nav(`/plan-sets/${r.id}`)
    } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') } finally { setBusy(false) }
  }

  if (err && !meta) return <p role="alert" className="m-4 rounded-xl bg-red-50 p-3 text-sm font-bold text-red-700">{err}</p>
  if (!meta || !rows) return <p className="p-6 text-sm text-slate-500">جاري التحميل…</p>

  return (
    <div className="space-y-4 p-4" dir="rtl">
      {/* البانر */}
      <section className="relative overflow-hidden rounded-3xl bg-gradient-to-l from-[#0b1d4d] via-[#123072] to-[#1d3f8f] text-white shadow">
        <div className="absolute inset-y-0 left-0 hidden w-3/5 md:block"><ArtCrop part="banner" className="h-full w-full opacity-95" /></div>
        <div className="relative z-10 flex items-center gap-4 p-6 md:w-2/5">
          <ArtCrop part="rolls" className="hidden h-20 w-28 shrink-0 rounded-xl sm:block" />
          <div className="space-y-2">
            <h1 className="text-3xl font-extrabold">سيت المخططات</h1>
            <p className="text-sm text-white/85">إدارة وتجميع جميع المخططات في سيت واحد احترافي</p>
            {meta.canManage && (
              <button type="button" onClick={() => setCreate((v) => !v)}
                className="rounded-xl bg-gradient-to-l from-amber-500 to-orange-500 px-5 py-2 text-sm font-extrabold text-white shadow">＋ إنشاء سيت جديد</button>
            )}
          </div>
        </div>
      </section>

      <nav className="text-xs text-slate-500"><Link to="/" className="hover:underline">الرئيسية</Link> ‹ <b className="text-slate-700">سيت المخططات</b></nav>
      {err && <p role="alert" className="rounded-xl bg-red-50 p-3 text-sm font-bold text-red-700">{err}</p>}

      {create && (
        <section className="grid gap-3 rounded-2xl border border-indigo-200 bg-white p-4 sm:grid-cols-2">
          <label className="text-sm">اسم السيت
            <input id="ps-title" value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} placeholder="سيت مشروع الجامعة"
              className="mt-1 w-full rounded-xl border border-slate-300 px-3 py-2" /></label>
          <label className="text-sm">المشروع
            <select id="ps-project" value={form.projectId} onChange={(e) => setForm({ ...form, projectId: e.target.value })}
              className="mt-1 w-full rounded-xl border border-slate-300 bg-white px-3 py-2" disabled={form.isTemplate}>
              <option value="">— بلا مشروع —</option>
              {meta.projects.map((p) => <option key={p.id} value={p.id}>{p.code} · {p.name}</option>)}
            </select></label>
          <label className="flex items-center gap-2 text-sm sm:col-span-2">
            <input id="ps-template" type="checkbox" checked={form.isTemplate} onChange={(e) => setForm({ ...form, isTemplate: e.target.checked, projectId: e.target.checked ? '' : form.projectId })} />
            نموذج جاهز — للمعاينة والإعلانات، وتنسخ منه أقسام لسيتات المشاريع
          </label>
          <div className="flex gap-2 sm:col-span-2">
            <button type="button" disabled={busy} onClick={submit} className="rounded-xl bg-indigo-700 px-5 py-2 text-sm font-bold text-white disabled:opacity-50">{busy ? '…' : 'إنشاء'}</button>
            <button type="button" onClick={() => setCreate(false)} className="rounded-xl px-5 py-2 text-sm ring-1 ring-slate-300">إلغاء</button>
          </div>
        </section>
      )}

      {/* الأرقام */}
      <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
        <Stat n={typeCount} label="أنواع المخططات" icon={<IconLayers />} />
        <Stat n={meta.stats.projects} label="المشاريع" icon={<IconFolder />} />
        <Stat n={live.length} label="إجمالي السيتات" icon={<IconLayers />} />
        <Stat n={meta.stats.files} label="إجمالي التصاميم" icon={<IconDoc />} />
      </div>

      <div className="grid gap-4 lg:grid-cols-[1fr_300px]">
        {/* جميع سيتات التصميم */}
        <section className="min-w-0 rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
          <div className="mb-3 flex flex-wrap items-center gap-3">
            <h2 className="border-b-2 border-orange-500 pb-1 text-lg font-extrabold text-slate-900">جميع سيتات التصميم</h2>
            <div className="flex gap-1 rounded-xl bg-slate-100 p-1 text-xs">
              {([['projects', 'سيتات المشاريع'], ['templates', 'نماذج جاهزة']] as const).map(([k, l]) => (
                <button key={k} type="button" onClick={() => setTab(k)} aria-pressed={tab === k}
                  className={`rounded-lg px-3 py-1 font-bold ${tab === k ? 'bg-white text-indigo-700 shadow' : 'text-slate-600'}`}>{l}</button>
              ))}
            </div>
            {shown.length > 6 && <button type="button" onClick={() => setShowAll((v) => !v)} className="ms-auto text-xs font-bold text-slate-600 hover:underline">{showAll ? 'أقل' : 'عرض الكل'}</button>}
          </div>
          {shown.length === 0 ? (
            <p className="rounded-xl bg-slate-50 p-6 text-center text-sm text-slate-500">
              {tab === 'templates' ? 'ماكو نماذج جاهزة بعد — سوّي سيت وأشّر «نموذج جاهز».' : 'ماكو سيتات مشاريع بعد.'}
            </p>
          ) : (
            <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
              {(showAll ? shown : shown.slice(0, 6)).map((r) => (
                <article key={r.id} className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
                  <Link to={`/plan-sets/${r.id}`} className="relative block h-36 bg-slate-100">
                    {isImg(r.cover) ? <img loading="lazy" decoding="async" src={thumbUrl(r.cover!, 480)} alt="" className="h-full w-full object-cover" />
                      : <ArtCrop part="card0" className="h-full w-full" />}
                    <span className="absolute start-2 top-2 rounded-lg bg-slate-900/75 px-2 py-0.5 text-xs font-bold text-white">{r.files} تصاميم</span>
                    {r.isTemplate && <span className="absolute end-2 top-2 rounded-lg bg-amber-500 px-2 py-0.5 text-xs font-bold text-white">نموذج</span>}
                  </Link>
                  <div className="space-y-1 p-3">
                    <h3 className="truncate font-extrabold text-slate-900">{r.title}</h3>
                    <p className="flex items-center gap-1 text-xs text-slate-500"><IconCal /> {day(r.createdAt)}{r.projectName ? ` · ${r.projectName}` : ''}</p>
                  </div>
                  <div className="flex border-t border-slate-100 text-xs">
                    <Link to={`/plan-sets/${r.id}`} className="flex-1 py-2 text-center font-bold text-slate-700 hover:bg-slate-50">👁 معاينة</Link>
                    <Link to={`/plan-sets/${r.id}?pdf=1`} className="flex-1 border-s border-slate-100 py-2 text-center font-bold text-slate-700 hover:bg-slate-50">تحميل</Link>
                  </div>
                </article>
              ))}
            </div>
          )}
        </section>

        {/* رفع مخططات جديدة */}
        <aside className="flex flex-col items-center justify-center gap-2 rounded-2xl border-2 border-dashed border-slate-300 bg-white p-5 text-center">
          <div className="text-indigo-600"><IconUpload /></div>
          <b className="text-slate-900">رفع مخططات جديدة</b>
          <p className="text-xs text-slate-500">اختار السيت، بعدها القسم والمخطط الفرعي، وارفع تصميمك</p>
          <p className="text-xs font-bold text-orange-600" dir="ltr">PDF, DWG, JPG, PNG</p>
          {!pickFor ? (
            <button type="button" onClick={() => setPickFor(true)} disabled={live.length === 0}
              className="mt-1 rounded-xl bg-gradient-to-l from-amber-500 to-orange-500 px-5 py-2 text-sm font-bold text-white disabled:opacity-50">📁 اختيار الملفات</button>
          ) : (
            <div className="w-full space-y-2">
              <select id="ps-upload-set" value={pickSet} onChange={(e) => setPickSet(e.target.value)} className="w-full rounded-xl border border-slate-300 bg-white px-2 py-2 text-sm">
                <option value="">اختار السيت…</option>
                {live.map((r) => <option key={r.id} value={r.id}>{r.isTemplate ? '⭐ ' : ''}{r.title}</option>)}
              </select>
              <button type="button" disabled={!pickSet} onClick={() => nav(`/plan-sets/${pickSet}?upload=1`)}
                className="w-full rounded-xl bg-indigo-700 py-2 text-sm font-bold text-white disabled:opacity-50">كمّل للرفع</button>
            </div>
          )}
          {live.length === 0 && <p className="text-xs text-slate-400">سوّي سيت أول</p>}
        </aside>
      </div>
    </div>
  )
}
