import { lazy, Suspense, useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { api, apiRequest, ensureFileToken, fileUrl, type Employee, thumbUrl } from '../api'
import PctWhy from '../ui/PctWhy'
import { SECTION_STATUS, type DesignFile, type DesignMeta, type DesignSection, type DesignSetDetail, type DesignTool } from './design/types'

const ModelViewer = lazy(() => import('./design/ModelViewer'))

// ═══ 🏗️ صيت تصاميم مشروع — (ع) 11-08 ═══
// لكل تخصص: أزرار أدوات التصميم (تنفتح بتبويب جديد)، طريقة إرجاع الشغل، الملفات
// والروابط (لصق Ctrl+V، سحب، رفع، رابط مشاركة)، والمراجعة والاعتماد.
const RETURN_STEPS = [
  'افتح الأداة من الزر (تنفتح بتبويب جديد) وصمّم.',
  'صدّر الشغل: PDF، أو صورة، أو DWG (أوتوكاد)، أو GLB (ثري دي ينفتح هنا).',
  'ارجع لهالتخصص: اسحب الملف وذبّه بالمربع، أو سوّي لقطة شاشة والصقها (Ctrl+V)، أو اضغط «اختار ملف».',
  'إذا الأداة تنطي رابط مشاركة (Share)، الصقه بخانة الرابط — يفتح التصميم الحي.',
  'لمن تخلّص كل الملفات اضغط «📤 للمراجعة».',
]

function FileCard({ f, canRemove, onRemove }: { f: DesignFile; canRemove: boolean; onRemove: () => void }) {
  const [open3d, setOpen3d] = useState(false)
  const src = f.fileKey ? fileUrl(f.fileKey) : ''
  const isImg = (f.fileType || '').startsWith('image/') && f.fileType !== 'image/vnd.dwg'
  const is3d = f.fileType === 'model/gltf-binary'
  return (
    <div className="rounded-xl border border-slate-200 bg-white p-2 text-xs">
      {isImg && <a href={src} target="_blank" rel="noreferrer"><img loading="lazy" decoding="async" src={f.fileKey ? thumbUrl(f.fileKey, 480) : src} alt={f.title} className="h-28 w-full rounded-lg object-cover" /></a>}
      {is3d && (open3d
        ? <Suspense fallback={<p className="p-4 text-center text-slate-400">…</p>}><ModelViewer src={src} alt={f.title} /></Suspense>
        : <button type="button" onClick={() => setOpen3d(true)} className="flex h-28 w-full items-center justify-center rounded-lg bg-slate-100 font-bold text-slate-600">🧊 افتح المجسّم ثري دي</button>)}
      {!isImg && !is3d && (
        <a href={f.url || src} target="_blank" rel="noreferrer" className="flex h-28 w-full items-center justify-center rounded-lg bg-slate-50 text-center font-bold text-brand-700">
          {f.url ? '🔗 افتح الرابط' : f.fileType === 'application/pdf' ? '📄 افتح PDF' : f.fileType === 'image/vnd.dwg' ? '📐 نزّل ملف DWG' : '📎 افتح'}
        </a>
      )}
      <p className="mt-1 truncate font-bold text-slate-700" title={f.title}>{f.title}</p>
      <p className="truncate text-slate-400">{f.toolName ? `${f.toolName} · ` : ''}{f.uploadedByName ?? ''}</p>
      {canRemove && <button type="button" onClick={onRemove} className="text-[11px] text-slate-400 underline">شيله</button>}
    </div>
  )
}

function SectionCard({ s, d, meta, tools, people, reload }: {
  s: DesignSection; d: DesignSetDetail; meta: DesignMeta; tools: DesignTool[]; people: Employee[]; reload: () => void
}) {
  const disc = meta.disciplines.find((x) => x.code === s.discipline)
  const [open, setOpen] = useState(false)
  const [help, setHelp] = useState(false)
  const [err, setErr] = useState('')
  const [busy, setBusy] = useState(false)
  const [link, setLink] = useState('')
  const [title, setTitle] = useState('')
  const [tool, setTool] = useState('')
  const [assign, setAssign] = useState({ who: s.assigneeId ?? '', due: s.dueDate ?? '' })
  const [note, setNote] = useState('')
  const locked = s.status === 'APPROVED'
  const st = SECTION_STATUS[s.status]
  const run = async (fn: () => Promise<unknown>) => {
    setBusy(true); setErr('')
    try { await fn(); reload() } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') } finally { setBusy(false) }
  }
  const upload = (files: FileList | File[]) => run(async () => {
    for (const file of Array.from(files)) {
      const up = await api.uploadFile(file, 'designs')
      await apiRequest(`/design-sections/${s.id}/files`, { method: 'POST', body: JSON.stringify({ title: title.trim() || file.name, fileKey: up.url, fileType: up.type, toolName: tool }) })
    }
    setTitle('')
  })
  const addLink = () => run(async () => {
    await apiRequest(`/design-sections/${s.id}/files`, { method: 'POST', body: JSON.stringify({ title: title.trim(), url: link.trim(), toolName: tool }) })
    setLink(''); setTitle('')
  })
  const late = s.dueDate && s.status !== 'APPROVED' && s.dueDate < new Date().toLocaleDateString('en-CA', { timeZone: 'Asia/Baghdad' })

  return (
    <div className={`rounded-2xl border bg-white ${s.status === 'REVIEW' ? 'border-amber-300' : late ? 'border-red-300' : 'border-slate-200'}`}>
      <button type="button" onClick={() => setOpen(!open)} className="flex w-full flex-wrap items-center justify-between gap-2 p-3 text-right">
        <span className="font-extrabold text-[#0f2040]">{disc?.icon} {disc?.name ?? s.discipline} <span className="text-xs font-normal text-slate-400">· {s.files.length} ملف</span></span>
        <span className="flex flex-wrap items-center gap-2 text-xs">
          {s.assigneeName && <span className="text-slate-500">🎨 {s.assigneeName}{s.dueDate ? ` · ${s.dueDate}` : ''}</span>}
          {late && <span className="font-bold text-red-600">⚠️ متأخر</span>}
          <span className={`rounded-full px-2 py-0.5 font-bold ${st.cls}`}>{st.label}</span>
        </span>
      </button>
      {open && (
        <div className="space-y-3 border-t border-slate-100 p-3">
          {s.reviewNote && s.status === 'RETURNED' && <p className="rounded-lg bg-red-50 p-2 text-xs font-bold text-red-700">↩️ ملاحظة المراجعة: {s.reviewNote}</p>}
          {/* أدوات التصميم */}
          <div>
            <div className="flex flex-wrap items-center gap-2">
              <span className="text-xs font-bold text-slate-600">🧰 صمّم بـ:</span>
              {tools.length === 0 && <span className="text-xs text-slate-400">ماكو أدوات لهالتخصص بعد{d.canManage ? ' — ضيفها من «🧰 إدارة الأدوات»' : ''}</span>}
              {tools.map((t) => (
                <a key={t.id} href={t.url} target="_blank" rel="noreferrer" title={t.howTo} className="rounded-lg bg-brand-50 px-3 py-1.5 text-xs font-bold text-brand-800 ring-1 ring-brand-200 hover:bg-brand-100">↗ {t.name}</a>
              ))}
              <button type="button" onClick={() => setHelp(!help)} className="text-xs text-slate-500 underline">📖 شلون أرجّع الشغل هنا؟</button>
            </div>
            {help && (
              <div className="mt-2 rounded-xl bg-slate-50 p-3 text-xs text-slate-700">
                <ol className="list-decimal space-y-1 ps-5">{RETURN_STEPS.map((x) => <li key={x}>{x}</li>)}</ol>
                {tools.filter((t) => t.howTo).map((t) => <p key={t.id} className="mt-2"><b>{t.name}:</b> {t.howTo}</p>)}
              </div>
            )}
          </div>
          {/* الملفات */}
          {s.files.length > 0 && (
            <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
              {s.files.map((f) => <FileCard key={f.id} f={f} canRemove={!locked} onRemove={() => void run(() => apiRequest(`/design-files/${f.id}`, { method: 'DELETE' }))} />)}
            </div>
          )}
          {/* الرفع */}
          {!locked && (
            <div className="space-y-2 rounded-xl border border-slate-100 p-2">
              <div className="grid gap-2 sm:grid-cols-2">
                <input value={title} onChange={(e) => setTitle(e.target.value)} placeholder="اسم الملف (مثلاً: الطابق الأرضي)" className="rounded-lg border border-slate-200 p-2 text-xs" />
                <select value={tool} onChange={(e) => setTool(e.target.value)} className="rounded-lg border border-slate-200 p-2 text-xs">
                  <option value="">سوّيته بأي أداة؟</option>{tools.map((t) => <option key={t.id} value={t.name}>{t.name}</option>)}<option value="غير ذلك">غير ذلك</option>
                </select>
              </div>
              <div
                tabIndex={0}
                onPaste={(e) => { const fs = Array.from(e.clipboardData.files); if (fs.length) { e.preventDefault(); void upload(fs) } }}
                onDragOver={(e) => e.preventDefault()}
                onDrop={(e) => { e.preventDefault(); if (e.dataTransfer.files.length) void upload(e.dataTransfer.files) }}
                className="rounded-xl border-2 border-dashed border-slate-300 p-4 text-center text-xs text-slate-500 outline-none focus:border-brand-500"
              >
                اضغط هنا وسوّي <b>Ctrl+V</b> للصق لقطة الشاشة، أو اسحب الملفات وذبّها هنا
                <label className="mt-2 block"><span className="cursor-pointer rounded-lg bg-slate-100 px-3 py-1 font-bold text-slate-700">📎 اختار ملف</span>
                  <input type="file" multiple accept="image/png,image/jpeg,image/webp,application/pdf,.glb,.dwg" className="hidden" onChange={(e) => { if (e.target.files?.length) void upload(e.target.files) }} />
                </label>
                <span className="mt-1 block text-[10px] text-slate-400">صور · PDF · DWG أوتوكاد · GLB ثري دي — لحد ٢٤ ميگا</span>
              </div>
              <div className="flex gap-2">
                <input value={link} onChange={(e) => setLink(e.target.value)} placeholder="أو الصق رابط المشاركة من الأداة (https://…)" dir="ltr" className="min-w-0 flex-1 rounded-lg border border-slate-200 p-2 text-xs" />
                <button type="button" disabled={busy || !link.trim()} onClick={() => void addLink()} className="rounded-lg bg-slate-800 px-3 text-xs font-bold text-white disabled:opacity-50">ضيف</button>
              </div>
            </div>
          )}
          {/* الإجراءات */}
          <div className="flex flex-wrap items-end gap-2">
            {!locked && s.status !== 'REVIEW' && s.files.length > 0 && (
              <button type="button" disabled={busy} onClick={() => void run(() => apiRequest(`/design-sections/${s.id}/submit`, { method: 'POST', body: '{}' }))} className="rounded-lg bg-amber-600 px-3 py-1.5 text-xs font-bold text-white disabled:opacity-50">📤 للمراجعة</button>
            )}
            {d.canReview && (s.status === 'REVIEW' || s.status === 'APPROVED') && (
              <>
                <input value={note} onChange={(e) => setNote(e.target.value)} placeholder="ملاحظة (إجبارية للإرجاع)" className="min-w-0 flex-1 rounded-lg border border-slate-200 p-1.5 text-xs" />
                {s.status === 'REVIEW' && <button type="button" disabled={busy} onClick={() => void run(() => apiRequest(`/design-sections/${s.id}/review`, { method: 'POST', body: JSON.stringify({ approve: true, note }) }))} className="rounded-lg bg-emerald-700 px-3 py-1.5 text-xs font-bold text-white disabled:opacity-50">✅ اعتمد</button>}
                <button type="button" disabled={busy} onClick={() => void run(() => apiRequest(`/design-sections/${s.id}/review`, { method: 'POST', body: JSON.stringify({ approve: false, note }) }))} className="rounded-lg bg-red-600 px-3 py-1.5 text-xs font-bold text-white disabled:opacity-50">↩️ رجّعه للتعديل</button>
              </>
            )}
          </div>
          {d.canManage && (
            <div className="flex flex-wrap items-end gap-2 rounded-xl bg-slate-50 p-2 text-xs">
              <label className="flex-1">المصمم
                <select value={assign.who} onChange={(e) => setAssign({ ...assign, who: e.target.value })} className="block w-full rounded-lg border border-slate-200 p-1.5">
                  <option value="">اختار…</option>{people.map((p) => <option key={p.id} value={p.id}>{p.name}</option>)}
                </select>
              </label>
              <label>موعد التسليم<input type="date" value={assign.due} onChange={(e) => setAssign({ ...assign, due: e.target.value })} className="block rounded-lg border border-slate-200 p-1.5" /></label>
              <button type="button" disabled={busy || !assign.who || !assign.due} onClick={() => void run(() => apiRequest(`/design-sections/${s.id}/assign`, { method: 'POST', body: JSON.stringify({ assigneeId: assign.who, dueDate: assign.due }) }))} className="rounded-lg bg-brand-700 px-3 py-1.5 font-bold text-white disabled:opacity-50">{s.assigneeId ? 'غيّر' : 'كلّف'} (تصير مهمة)</button>
            </div>
          )}
          {err && <p className="text-xs font-bold text-red-600">{err}</p>}
        </div>
      )}
    </div>
  )
}

function ToolsManager({ meta, close, saved }: { meta: DesignMeta; close: () => void; saved: () => void }) {
  const [rows, setRows] = useState(meta.tools)
  const [err, setErr] = useState('')
  const empty = { id: '', discipline: meta.disciplines[0]?.code ?? 'ARCH', name: '', url: '', howTo: '', sort: 0, active: true }
  const save = async (t: DesignTool) => {
    setErr('')
    try {
      const body = JSON.stringify({ discipline: t.discipline, name: t.name, url: t.url, howTo: t.howTo, active: t.active })
      await apiRequest(t.id ? `/design-tools/${t.id}` : '/design-tools', { method: t.id ? 'PUT' : 'POST', body })
      saved()
      if (!t.id) setRows((r) => r.filter((x) => x !== t))
    } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') }
  }
  const set = (i: number, k: keyof DesignTool, v: string | boolean) => setRows((r) => r.map((x, j) => (j === i ? { ...x, [k]: v } : x)))
  return (
    <div dir="rtl" className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={close}>
      <div className="max-h-[90vh] w-full max-w-3xl space-y-2 overflow-y-auto rounded-2xl bg-white p-4" onClick={(e) => e.stopPropagation()}>
        <div className="flex items-center justify-between"><h3 className="font-bold text-brand-900">🧰 أدوات التصميم</h3><button type="button" onClick={close}>✕</button></div>
        <p className="text-xs text-slate-500">كل أداة تطلع زر بتخصصها. «طريقة الإرجاع» تطلع للمصمم بالشرح.</p>
        {err && <p className="text-xs text-red-600">{err}</p>}
        {rows.map((t, i) => (
          <div key={t.id || `new-${i}`} className={`grid gap-1 rounded-lg border p-2 text-xs sm:grid-cols-6 ${t.active ? 'border-slate-200' : 'border-slate-100 opacity-60'}`}>
            <select value={t.discipline} onChange={(e) => set(i, 'discipline', e.target.value)} className="rounded border border-slate-200 p-1">{meta.disciplines.map((d) => <option key={d.code} value={d.code}>{d.name}</option>)}</select>
            <input value={t.name} onChange={(e) => set(i, 'name', e.target.value)} placeholder="الاسم" className="rounded border border-slate-200 p-1" />
            <input value={t.url} onChange={(e) => set(i, 'url', e.target.value)} placeholder="https://…" dir="ltr" className="rounded border border-slate-200 p-1 sm:col-span-2" />
            <label className="flex items-center gap-1"><input type="checkbox" checked={t.active} onChange={(e) => set(i, 'active', e.target.checked)} /> فعّالة</label>
            <button type="button" onClick={() => void save(t)} className="rounded bg-emerald-700 px-2 font-bold text-white">احفظ</button>
            <input value={t.howTo} onChange={(e) => set(i, 'howTo', e.target.value)} placeholder="طريقة الإرجاع للمصمم" className="rounded border border-slate-200 p-1 sm:col-span-6" />
          </div>
        ))}
        <button type="button" onClick={() => setRows((r) => [...r, { ...empty }])} className="rounded-lg bg-slate-800 px-3 py-1.5 text-xs font-bold text-white">+ أداة جديدة</button>
      </div>
    </div>
  )
}

export default function DesignSetPage() {
  const { id = '' } = useParams()
  const [meta, setMeta] = useState<DesignMeta | null>(null)
  const [d, setD] = useState<DesignSetDetail | null>(null)
  const [people, setPeople] = useState<Employee[]>([])
  const [err, setErr] = useState('')
  const [reload, setReload] = useState(0)
  const [tools, setTools] = useState(false)
  const [add, setAdd] = useState('')
  useEffect(() => {
    let alive = true
    Promise.all([apiRequest<DesignMeta>('/design-sets/meta'), apiRequest<DesignSetDetail>(`/design-sets/${encodeURIComponent(id)}`), ensureFileToken()])
      .then(([m, x]) => { if (alive) { setMeta(m); setD(x); setErr('') } })
      .catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذر') })
    return () => { alive = false }
  }, [id, reload])
  useEffect(() => {
    if (!d?.canManage) return
    let alive = true
    api.getEmployees().then((r) => { if (alive) setPeople(r.filter((e) => e.status === 'ACTIVE')) }).catch(() => {})
    return () => { alive = false }
  }, [d?.canManage])
  const refresh = () => setReload((n) => n + 1)

  if (err) return <p dir="rtl" className="text-sm text-red-600">{err}</p>
  if (!d || !meta) return <p dir="rtl" className="text-slate-400">جاري التحميل…</p>
  const missing = meta.disciplines.filter((x) => !d.sections.some((s) => s.discipline === x.code))
  const pct = d.sections.length ? Math.round((100 * d.approved) / d.sections.length) : 0

  return (
    <div dir="rtl" className="space-y-4">
      <Link to="/design-sets" className="text-xs text-slate-500 underline">← كل الصيتات</Link>
      <div className="rounded-2xl bg-white p-4 shadow-sm">
        <div className="flex flex-wrap items-start justify-between gap-2">
          <div>
            <h2 className="text-2xl font-bold text-brand-900">🏗️ {d.title}</h2>
            <p className="text-xs text-slate-500">{d.projectCode ? `مشروع ${d.projectCode} · ` : ''}{d.createdByName ? `فتحه ${d.createdByName} · ` : ''}{new Date(d.createdAt).toLocaleDateString('en-GB')}</p>
            {d.notes && <p className="mt-1 text-sm text-slate-700">{d.notes}</p>}
          </div>
          <div className="text-center"><p className="text-3xl font-black text-brand-700"><PctWhy suffix="٪" title="إنجاز الصيت" value={d.sections.length ? pct : null}
            parts={{ numLabel: 'قسم معتمد', num: d.approved, denLabel: 'قسم بالصيت', den: d.sections.length }}
            meaning={<>شكد من أقسام الصيت (كل اختصاص قسم) خلص واعتمده المراجع. القسم الي بالتصميم أو ينتظر المراجعة أو رجع للتعديل ما ينحسب لحد ما ينعتمد. تصير ١٠٠٪ لمن كل الأقسام تنعتمد.</>}
            source={<>أقسام هذا الصيت وحالة كل قسم (مثل ما بالقائمة تحت).</>}>
            <ul className="text-xs">{d.sections.map((s) => <li key={s.id}>{meta.disciplines.find((x) => x.code === s.discipline)?.name ?? s.discipline}: {SECTION_STATUS[s.status].label}</li>)}</ul>
          </PctWhy></p><p className="text-[11px] text-slate-500">{d.approved} من {d.sections.length} معتمد</p></div>
        </div>
        {d.canManage && (
          <div className="mt-3 flex flex-wrap items-center gap-2 text-xs">
            <button type="button" onClick={() => setTools(true)} className="rounded-lg border border-slate-200 px-3 py-1.5 font-bold">🧰 إدارة الأدوات</button>
            {missing.length > 0 && (
              <>
                <select value={add} onChange={(e) => setAdd(e.target.value)} className="rounded-lg border border-slate-200 p-1.5"><option value="">+ تخصص ثاني</option>{missing.map((x) => <option key={x.code} value={x.code}>{x.icon} {x.name}</option>)}</select>
                {add && <button type="button" onClick={() => void apiRequest(`/design-sets/${d.id}/sections`, { method: 'POST', body: JSON.stringify({ discipline: add }) }).then(() => { setAdd(''); refresh() })} className="rounded-lg bg-brand-700 px-3 py-1.5 font-bold text-white">ضيف</button>}
              </>
            )}
            <button type="button" onClick={() => void apiRequest(`/design-sets/${d.id}/archive`, { method: 'POST', body: JSON.stringify({ archived: !d.archivedAt }) }).then(refresh)} className="rounded-lg border border-slate-200 px-3 py-1.5">{d.archivedAt ? '↩️ رجّعه من الأرشيف' : '🗄️ أرشفة'}</button>
          </div>
        )}
      </div>
      <div className="space-y-2">
        {d.sections.map((s) => <SectionCard key={s.id} s={s} d={d} meta={meta} tools={meta.tools.filter((t) => t.discipline === s.discipline && t.active)} people={people} reload={refresh} />)}
      </div>
      {tools && <ToolsManager meta={meta} close={() => setTools(false)} saved={refresh} />}
    </div>
  )
}
