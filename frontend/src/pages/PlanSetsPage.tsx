import { useEffect, useState, type ReactNode } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { apiRequest, fileUrl } from '../api'
import { useSession } from '../session'
import {
  BlueprintLines, BuildingArt, IconArrow, IconBolt, IconCal, IconCalc, IconCheck, IconCity, IconClock, IconDoc,
  IconFolder, IconHeart, IconImage, IconLayers, IconPdf, IconPin, IconUpload, RollsArt,
} from './design/planArt'
import type { PlanMeta, PlanSetRow } from './design/planTypes'

// ═══ 📐 سيت المخططات — (ك) 11-10 بطلب دازاي ═══
// «اريد هذا بالضبط»: نفس صورة التصميم — بانر، أربع كروت، «مشاريعك»، وتحت أدوات سريعة
// وخطوات إنشاء سيت وآخر النشاطات. كل رقم ونسبة محسوبة من النظام (ماكو أرقام وهمية):
// النسبة = الي انضاف آخر ٣٠ يوم مقابل الي چان قبلها، وتختفي إذا ماكو أساس للمقارنة.
const ORANGE = 'linear-gradient(90deg,#f59e0b,#ea7a1c)'
const FAV_KEY = 'planFavProjects'

const day = (s: string) => new Date(s).toLocaleDateString('en-CA', { timeZone: 'Asia/Baghdad' }).replace(/-/g, '/')
function ago(s: string) {
  const m = Math.max(1, Math.round((Date.now() - new Date(s).getTime()) / 60000))
  if (m < 60) return `منذ ${m} دقيقة`
  const h = Math.round(m / 60)
  if (h < 24) return h === 1 ? 'منذ ساعة' : h === 2 ? 'منذ ساعتين' : `منذ ${h} ساعات`
  const d = Math.round(h / 24)
  return d === 1 ? 'منذ يوم' : d === 2 ? 'منذ يومين' : `منذ ${d} أيام`
}
function readFav(): string[] {
  try { return JSON.parse(localStorage.getItem(FAV_KEY) || '[]') as string[] } catch { return [] }
}

const Panel = ({ children, className = '' }: { children: ReactNode; className?: string }) => (
  <div className={`rounded-2xl border border-[#2a3a6e] bg-[#111c3d]/90 ${className}`}>{children}</div>
)

export default function PlanSetsPage() {
  const nav = useNavigate()
  const { employee, permissions } = useSession()
  const [meta, setMeta] = useState<PlanMeta | null>(null)
  const [rows, setRows] = useState<PlanSetRow[] | null>(null)
  const [err, setErr] = useState('')
  const [show, setShow] = useState(false)
  const [all, setAll] = useState(false)
  const [archived, setArchived] = useState(false)
  const [form, setForm] = useState({ title: '', projectId: '', notes: '' })
  const [busy, setBusy] = useState(false)
  const [fav, setFav] = useState<string[]>(readFav)

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
  const toggleFav = (id: string) => {
    const next = fav.includes(id) ? fav.filter((x) => x !== id) : [...fav, id]
    setFav(next)
    try { localStorage.setItem(FAV_KEY, JSON.stringify(next)) } catch { /* المتصفح مانع الحفظ — المفضلة تبقى للجلسة */ }
  }
  const openCreate = () => { setShow(true); setTimeout(() => document.getElementById('plan-create')?.scrollIntoView({ behavior: 'smooth' }), 50) }

  const ov = meta?.overview
  const stats = meta?.stats
  const latest = (rows ?? []).find((r) => !r.archivedAt)
  // عرض السعر: نفس شرط صفحة السيت — غير المسموح له ما ننطيه الرابط حتى ما تنسجل 403 عليه.
  const canQuotes = employee?.role === 'ADMIN' || employee?.actualRole === 'OWNER'
    || permissions.includes('quotation_manage_all') || permissions.includes('quotation_system')
    || permissions.includes('quotation_create') || permissions.includes('quotation_edit_own')
  const cards: [string, number | undefined, number | null | undefined, ReactNode][] = [
    ['مشروع', stats?.projects, ov?.growth.projects, <IconCity key="c" />],
    ['سيت تصميم', stats?.sets, ov?.growth.sets, <IconFolder key="f" />],
    ['مخطط', stats?.files, ov?.growth.files, <IconLayers key="l" />],
    ['ملف PDF', ov?.pdfs, ov?.growth.pdfs, <IconDoc key="d" />],
  ]
  const projects = [...(ov?.projectCards ?? [])].sort((a, b) => Number(fav.includes(b.projectId)) - Number(fav.includes(a.projectId)))
  const shown = (rows ?? []).filter((r) => !!r.archivedAt === archived)

  const tools: [string, ReactNode, () => void, boolean][] = [
    ['رفع المخططات', <IconUpload key="u" />, () => (latest ? nav(`/plan-sets/${latest.id}`) : openCreate()), true],
    ['معاينة السيت', <IconImage key="i" />, () => (latest ? nav(`/plan-sets/${latest.id}`) : setAll(true)), true],
    ['عرض السعر', <IconCalc key="q" />, () => nav('/quotations'), canQuotes],
    ['تنزيل PDF', <IconPdf key="p" />, () => latest && nav(`/plan-sets/${latest.id}?pdf=1`), !!latest],
  ]

  return (
    <div dir="rtl" className="-m-2 space-y-4 rounded-3xl bg-[#0b1430] p-3 text-white sm:p-4">
      {/* البانر */}
      <div className="relative overflow-hidden rounded-3xl border border-[#2a3a6e] bg-gradient-to-l from-[#0f1a3d] via-[#15215a] to-[#1b2b6b]">
        <div className="absolute inset-y-0 left-0 w-full sm:w-[58%]">
          <BuildingArt seed={0} className="h-full w-full" />
          <div className="absolute inset-0 bg-gradient-to-l from-[#15215a] via-transparent to-transparent" />
        </div>
        <BlueprintLines className="absolute left-[38%] top-0 hidden h-full sm:block" />
        <div className="relative flex min-h-[210px] flex-col items-start justify-center gap-2 p-6 sm:max-w-[45%]">
          <div className="flex items-center gap-3">
            <span className="text-amber-400"><IconCity size={46} /></span>
            <div>
              <h2 className="text-3xl font-extrabold">سيت المخططات</h2>
              <p className="text-sm text-white/80">إدارة مجموعات المخططات الهندسية والتصاميم</p>
            </div>
          </div>
          <p className="text-xs text-white/60">تنظيم احترافي .. لمشروعك بتفاصيل أدق</p>
          {meta?.canManage && (
            <button type="button" onClick={() => (show ? setShow(false) : openCreate())} className="mt-2 flex items-center gap-3 rounded-full px-6 py-2.5 text-sm font-bold text-white shadow-lg shadow-orange-900/40" style={{ background: ORANGE }}>
              إنشاء سيت جديد <IconArrow size={18} />
            </button>
          )}
        </div>
      </div>

      {/* الكروت الأربعة */}
      <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
        {cards.map(([label, v, g, icon]) => (
          <Panel key={label} className="flex items-center justify-between p-4">
            <div>
              <p className="text-3xl font-extrabold">{v ?? '—'}</p>
              <p className="text-sm text-white/80">{label}</p>
              {g != null && <p className="mt-1 text-xs font-bold text-emerald-400" dir="ltr">↑ +{Math.round(g)}%</p>}
            </div>
            <span className="grid h-14 w-14 place-items-center rounded-full bg-[#24336b] text-[#b9c6ff]">{icon}</span>
          </Panel>
        ))}
      </div>

      {err && <p className="rounded-xl bg-red-500/15 p-2 text-sm text-red-300">{err}</p>}

      {show && meta && (
        <Panel className="space-y-3 p-4">
          <div id="plan-create" />
          <p className="font-bold">سيت جديد</p>
          <input value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} placeholder="اسم السيت (مثلاً: سيت مشروع الجامعة)" className="w-full rounded-lg border border-[#2a3a6e] bg-[#0b1430] p-2 text-sm text-white placeholder:text-white/40" />
          <select value={form.projectId} onChange={(e) => setForm({ ...form, projectId: e.target.value })} className="w-full rounded-lg border border-[#2a3a6e] bg-[#0b1430] p-2 text-sm text-white">
            <option value="">مربوط بمشروع؟ (اختياري — منه ينجاب عرض السعر)</option>
            {meta.projects.map((p) => <option key={p.id} value={p.id}>{p.code} — {p.name}</option>)}
          </select>
          <textarea value={form.notes} onChange={(e) => setForm({ ...form, notes: e.target.value })} placeholder="ملاحظات (اختياري)" rows={2} className="w-full rounded-lg border border-[#2a3a6e] bg-[#0b1430] p-2 text-sm text-white placeholder:text-white/40" />
          <button type="button" disabled={busy || form.title.trim().length < 3} onClick={() => void create()} className="rounded-full px-6 py-2 text-sm font-bold text-white disabled:opacity-50" style={{ background: ORANGE }}>أنشئ السيت</button>
        </Panel>
      )}

      {/* مشاريعك / كل السيتات */}
      <Panel className="p-4">
        <div className="mb-3 flex items-center justify-between">
          <h3 className="flex items-center gap-2 text-lg font-extrabold"><span className="text-[#8ea2ff]"><IconCity /></span><span className="h-5 w-1 rounded bg-amber-500" />{all ? 'كل السيتات' : 'مشاريعك'}</h3>
          <div className="flex items-center gap-2">
            {all && [false, true].map((a) => (
              <button key={String(a)} type="button" onClick={() => setArchived(a)} className={`rounded-full px-3 py-1 text-xs font-bold ${archived === a ? 'bg-[#2c3d7a] text-white' : 'text-white/60'}`}>{a ? 'المؤرشفة' : 'الشغّالة'}</button>
            ))}
            <button type="button" onClick={() => setAll(!all)} className="flex items-center gap-2 rounded-full bg-[#1e2b5c] px-4 py-1.5 text-xs font-bold">{all ? 'المشاريع' : 'عرض الكل'} <IconArrow size={14} /></button>
          </div>
        </div>

        {!all && (
          projects.length === 0
            ? <p className="py-8 text-center text-sm text-white/60">ماكو مشاريع بيها سيتات بعد — افتح سيت واربطه بمشروع حتى يطلع هنا. {(rows?.length ?? 0) > 0 && <button type="button" onClick={() => setAll(true)} className="font-bold text-amber-400 underline">شوف كل السيتات</button>}</p>
            : (
              <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
                {projects.map((p, i) => {
                  const done = p.openSets === 0
                  return (
                    <div key={p.projectId} className="overflow-hidden rounded-2xl border border-[#2a3a6e] bg-[#0f1a3a]">
                      <div className="relative h-44">
                        {p.cover ? <img src={fileUrl(p.cover)} alt="" className="h-full w-full object-cover" /> : <BuildingArt seed={i + 1} className="h-full w-full" />}
                        <div className="absolute inset-0 bg-gradient-to-t from-[#0f1a3a] via-transparent to-transparent" />
                        <span className={`absolute end-3 top-3 flex items-center gap-1 rounded-full px-3 py-1 text-xs font-bold text-white ${done ? 'bg-emerald-700' : 'bg-amber-600'}`}>
                          {done ? <><IconCheck size={14} /> مكتمل</> : <><span className="h-2 w-2 rounded-full bg-white" /> قيد العمل</>}
                        </span>
                        <button type="button" onClick={() => toggleFav(p.projectId)} aria-label="مفضلة" className="absolute start-3 top-3 grid h-9 w-9 place-items-center rounded-full bg-black/35 text-white"><IconHeart on={fav.includes(p.projectId)} /></button>
                      </div>
                      <div className="space-y-2 p-4 pt-1">
                        <p className="text-lg font-extrabold">{p.name}</p>
                        <p className="flex items-center gap-1 text-xs text-white/60"><IconPin size={14} />{p.location || p.code}</p>
                        <div className="flex items-center justify-between border-t border-[#2a3a6e] pt-3 text-xs">
                          <span className="flex items-center gap-1"><IconLayers size={16} /> {p.sets} سيت</span>
                          <span className="flex items-center gap-1 text-white/70"><IconCal size={15} /> {day(p.lastAt)}</span>
                          <Link to={`/plan-sets/${p.latestSet}`} className="flex items-center gap-1 rounded-full border border-amber-500 px-3 py-1 font-bold text-amber-400">عرض التفاصيل <IconArrow size={13} /></Link>
                        </div>
                      </div>
                    </div>
                  )
                })}
              </div>
            )
        )}

        {all && (
          shown.length === 0
            ? <p className="py-8 text-center text-sm text-white/60">ماكو سيتات هنا.</p>
            : (
              <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
                {shown.map((r, i) => (
                  <div key={r.id} className="overflow-hidden rounded-2xl border border-[#2a3a6e] bg-[#0f1a3a]">
                    <div className="relative h-36">
                      {r.cover ? <img src={fileUrl(r.cover)} alt="" className="h-full w-full object-cover" /> : <BuildingArt seed={i + 2} className="h-full w-full" />}
                      <span className="absolute end-3 top-3 rounded-full bg-[#0b1430]/85 px-3 py-1 text-xs font-bold">{r.files} تصاميم</span>
                    </div>
                    <div className="space-y-1 p-3">
                      <p className="font-bold">{r.title}</p>
                      <p className="flex items-center gap-1 text-xs text-white/60"><IconCal size={14} />{day(r.createdAt)}{r.projectName ? ` · ${r.projectName}` : ''}</p>
                      <div className="flex gap-4 pt-1 text-xs font-bold">
                        <Link to={`/plan-sets/${r.id}`} className="text-[#b9c6ff]">👁️ معاينة</Link>
                        <Link to={`/plan-sets/${r.id}?pdf=1`} className="text-amber-400">⬇️ تحميل</Link>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )
        )}
      </Panel>

      {/* أدوات سريعة · خطوات إنشاء سيت · آخر النشاطات */}
      <div className="grid gap-3 lg:grid-cols-[1fr_1.15fr_1fr]">
        <Panel className="p-4">
          <p className="mb-3 flex items-center gap-2 font-bold"><span className="text-amber-400"><IconBolt size={18} /></span> أدوات سريعة</p>
          <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
            {tools.map(([label, icon, go, ok]) => (
              <button key={label} type="button" disabled={!ok} onClick={go} className="flex flex-col items-center gap-2 rounded-xl bg-[#1a2654] px-2 py-4 text-xs text-white/85 hover:bg-[#22316a] disabled:opacity-35">
                <span className="text-[#c9d3ff]">{icon}</span>{label}
              </button>
            ))}
          </div>
        </Panel>

        <Panel className="relative overflow-hidden p-4">
          <p className="mb-4 text-center font-bold">خطوات إنشاء سيت جديد</p>
          <div className="flex items-center gap-2">
            <RollsArt className="hidden w-32 shrink-0 sm:block" />
            <ol className="flex flex-1 items-start justify-between">
              {['المشروع', 'اختيار المخططات', 'رفع الملفات', 'عرض السعر'].map((s, i) => (
                <li key={s} className="relative flex flex-1 flex-col items-center gap-2 text-center text-xs">
                  {i < 3 && <span className={`absolute top-4 h-0.5 w-full -translate-x-1/2 ${i === 0 ? 'bg-amber-500' : 'bg-white/20'}`} style={{ insetInlineStart: '50%' }} />}
                  <span className={`relative z-10 grid h-8 w-8 place-items-center rounded-full border-2 font-bold ${i === 0 ? 'border-emerald-500 bg-emerald-600' : i === 1 ? 'border-amber-500 bg-amber-500' : 'border-white/40 bg-[#111c3d]'}`}>
                    {i === 0 ? <IconCheck size={16} /> : i + 1}
                  </span>
                  <span className={i === 1 ? 'font-bold text-amber-400' : 'text-white/80'}>{s}</span>
                </li>
              ))}
            </ol>
          </div>
        </Panel>

        <Panel className="p-4">
          <p className="mb-3 flex items-center gap-2 font-bold"><span className="text-amber-400"><IconClock size={18} /></span> آخر النشاطات</p>
          {(ov?.activity.length ?? 0) === 0
            ? <p className="text-xs text-white/50">ماكو نشاط بعد.</p>
            : (
              <ul className="space-y-3">
                {ov!.activity.map((a, i) => (
                  <li key={i}>
                    <Link to={`/plan-sets/${a.setId}`} className="flex items-start justify-between gap-2 text-xs">
                      <span className="flex items-start gap-2">
                        <span className="mt-0.5 text-[#8ea2ff]">{a.kind === 'set' ? <IconDoc size={18} /> : <IconUpload size={18} />}</span>
                        <span><b className="block text-sm">{a.kind === 'set' ? 'تم إنشاء سيت جديد' : 'تم رفع مخطط جديد'}</b><span className="text-white/60">{a.kind === 'set' ? (a.sub || a.title) : a.title}</span></span>
                      </span>
                      <span className="shrink-0 text-white/50">{ago(a.at)}</span>
                    </Link>
                  </li>
                ))}
              </ul>
            )}
        </Panel>
      </div>
    </div>
  )
}
